/**
 * Media proxy utilities for handling anti-hotlinking and caching
 */

// Cache for media cache enabled setting to avoid repeated API calls
let mediaCacheEnabledCache: boolean | null = null;
let mediaCachePromise: Promise<boolean> | null = null;

/**
 * Convert a media URL to use the proxy endpoint
 * @param url Original media URL
 * @param referer Optional referer URL for anti-hotlinking and resolving relative URLs
 * @param forceCache Force caching even if globally disabled (e.g., for image mode feeds)
 * @returns Proxied URL
 */
export function getProxiedMediaUrl(url: string, referer?: string, forceCache?: boolean): string {
  if (!url) return '';

  // Don't proxy data URLs or blob URLs
  if (url.startsWith('data:') || url.startsWith('blob:')) {
    return url;
  }

  // Don't proxy localhost URLs (these are local development URLs)
  if (url.startsWith('http://localhost') || url.startsWith('http://127.0.0.1')) {
    return url;
  }

  // Resolve relative URLs using the referer
  // Relative URLs need to be converted to absolute URLs before proxying
  let urlToProxy = url;
  if (referer && !url.startsWith('http://') && !url.startsWith('https://')) {
    // This is a relative URL, resolve it against the referer
    try {
      const baseUrl = new URL(referer);
      urlToProxy = new URL(url, baseUrl).href;
    } catch (e) {
      // If URL resolution fails, use the original URL
      console.warn('Failed to resolve relative URL:', url, 'against referer:', referer, e);
      urlToProxy = url;
    }
  }

  // CRITICAL FIX: Use base64 encoding to avoid all URL encoding issues
  // This prevents double-encoding problems with special characters, Chinese characters, etc.
  // Base64 encoding is safe for URLs and doesn't interfere with query parameter parsing
  const urlB64 = btoa(urlToProxy);

  // Build proxy URL with base64-encoded parameters
  let proxyUrl = `/api/media/proxy?url_b64=${urlB64}`;

  // Add referer if provided (also base64-encoded)
  if (referer) {
    const refererB64 = btoa(referer);
    proxyUrl += `&referer_b64=${refererB64}`;
  }

  // Add force_cache parameter if specified
  if (forceCache) {
    proxyUrl += `&force_cache=true`;
  }

  return proxyUrl;
}

/**
 * Check if media caching is enabled (with caching to avoid repeated API calls)
 * @returns Promise<boolean>
 */
export async function isMediaCacheEnabled(): Promise<boolean> {
  // Return cached value if available
  if (mediaCacheEnabledCache !== null) {
    return mediaCacheEnabledCache;
  }

  // If a request is already in flight, wait for it
  if (mediaCachePromise) {
    return mediaCachePromise;
  }

  // Start a new request
  mediaCachePromise = (async () => {
    try {
      const response = await fetch('/api/settings');
      if (response.ok) {
        const settings = await response.json();
        mediaCacheEnabledCache =
          settings.media_cache_enabled === 'true' || settings.media_cache_enabled === true;
        return mediaCacheEnabledCache;
      }
    } catch (error) {
      console.error('Failed to check media cache status:', error);
    }
    mediaCacheEnabledCache = false;
    return false;
  })();

  const result = await mediaCachePromise;
  mediaCachePromise = null; // Clear the promise after completion
  return result;
}

/**
 * Clear the media cache enabled cache (call this when settings change)
 */
export function clearMediaCacheEnabledCache(): void {
  mediaCacheEnabledCache = null;
}

/**
 * Process HTML content to proxy image URLs
 * @param html HTML content
 * @param referer Optional referer URL
 * @returns HTML with proxied image URLs
 * @note Unquoted src attributes are supported but must not contain spaces (per HTML spec)
 */
export function proxyImagesInHtml(html: string, referer?: string): string {
  if (!html) return html;

  // First, convert lazy-loaded images to normal images
  // This ensures images load immediately without waiting for lazy loading scripts
  let processed = convertLazyImages(html);

  // Then proxy the src attributes
  processed = proxyImgAttribute(processed, 'src', referer);

  return processed;
}

/**
 * Convert lazy-loaded images to normal images
 * For images with data-original or data-src attributes, move those URLs to src
 * This prevents lazy loading and ensures immediate display
 */
function convertLazyImages(html: string): string {
  // Match img tags with lazy loading attributes
  // Pattern: <img ... src="placeholder" data-original="real-image" ...>
  const lazyImgRegex =
    /<img([^>]*?)\s+(data-original|data-src)\s*=\s*(['"]?)([^"'\s>]+)\3([^>]*?)>/gi;

  return html.replace(lazyImgRegex, (match, _beforeAttr, lazyAttr, quote, lazySrc, _afterAttr) => {
    // Extract the current src attribute if it exists
    const srcMatch = match.match(/\ssrc\s*=\s*(['"]?)([^"'\s>]+)\1/i);

    if (srcMatch) {
      // Image already has src, replace it with the lazy src
      const newSrc = `src=${quote}${lazySrc}${quote}`;

      // Replace the src attribute with the lazy-loaded URL
      let newMatch = match.replace(/\ssrc\s*=\s*(['"]?)([^"'\s>]+)\1/i, ' ' + newSrc);

      // Remove lazy loading class if present
      newMatch = newMatch.replace(
        /\sclass\s*=\s*(['"]?)([^"'\s>]*\blazy\b[^"'\s>]*)\1/i,
        (_classMatch, classQuote, classValue) => {
          const newClassValue = classValue.replace(/\blazy\b/g, '').trim();
          if (newClassValue) {
            return ` class=${classQuote}${newClassValue}${classQuote}`;
          }
          return '';
        }
      );

      // Remove the data-original/data-src attribute since we've moved it to src
      // Build a simple regex that matches the attribute
      // If quote is empty, match unquoted value; otherwise match quoted value
      let removeRegex: RegExp;
      if (quote) {
        // Match quoted value: data-original="value" or data-original='value'
        const quoteChar = quote === '"' ? '"' : "'";
        removeRegex = new RegExp(`${lazyAttr}=${quoteChar}[^${quoteChar}]+${quoteChar}`, 'gi');
      } else {
        // Match unquoted value: data-original=value
        removeRegex = new RegExp(`${lazyAttr}=[^\\s>]+`, 'gi');
      }
      newMatch = newMatch.replace(removeRegex, '');

      return newMatch;
    }

    // No src attribute, add one with the lazy src (this shouldn't happen with valid HTML)
    return match;
  });
}

/**
 * Proxy a specific img attribute
 * @param html HTML content
 * @param attrName Attribute name to proxy (e.g., 'src', 'data-original', 'data-src')
 * @param referer Optional referer URL
 * @returns HTML with proxied attribute
 */
function proxyImgAttribute(html: string, attrName: string, referer?: string): string {
  // Enhanced regex to handle img attributes with better pattern matching
  // Handles double quotes, single quotes, and unquoted values
  // Note: Unquoted values cannot contain spaces per HTML specification
  const imgRegex = new RegExp(`<img([^>]+)${attrName}\\s*=\\s*(['"]?)([^"'\\s>]+)\\2`, 'gi');

  return html.replace(imgRegex, (match, _attrs, quote, src) => {
    // CRITICAL FIX: Decode HTML entities before processing the URL
    // HTML attributes contain &amp; which should be decoded to & before URL encoding
    // For example: &amp; becomes &, then gets properly URL-encoded as %26
    const decodedSrc = decodeHTMLEntities(src);
    const proxiedUrl = getProxiedMediaUrl(decodedSrc, referer);

    // If proxying failed or returned the same URL, keep original
    if (!proxiedUrl || proxiedUrl === decodedSrc) {
      return match;
    }

    // Replace the attribute, preserving the original quote style
    const newAttr = `${attrName}=${quote}${proxiedUrl}${quote}`;
    const attrRegex = new RegExp(
      `${attrName}\\s*=\\s*${quote}${src.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}${quote}`,
      'i'
    );

    return match.replace(attrRegex, newAttr);
  });
}

/**
 * Decode HTML entities in a string
 * Handles common entities like &amp;, &lt;, &gt;, &quot;, &#39;, etc.
 */
function decodeHTMLEntities(text: string): string {
  const textarea = document.createElement('textarea');
  textarea.innerHTML = text;
  return textarea.value;
}

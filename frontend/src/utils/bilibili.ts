/**
 * Bilibili utility functions for video URL handling
 */

/**
 * Regular expression to match Bilibili video URLs
 * IMPORTANT: Must match the full URL including protocol
 */
const BILIBILI_URL_PATTERNS = [
  /https?:\/\/[\w.]*bilibili\.com\/video\/(BV[\w]+)/i,
  /https?:\/\/[\w.]*bilibili\.com\/blackboard\/html5mobileplayer\.html/i,
];

/**
 * Regular expression to extract BVID from Bilibili URLs
 */
const BVID_EXTRACT_PATTERNS = [/bilibili\.com\/video\/(BV[\w]+)/i, /bvid=([\w]+)/i];

/**
 * Check if a URL is a Bilibili video URL
 * @param url - The URL to check
 * @returns True if the URL is a Bilibili video URL
 */
export function isBilibiliUrl(url: string | undefined): boolean {
  if (!url) return false;

  return BILIBILI_URL_PATTERNS.some((pattern) => pattern.test(url));
}

/**
 * Check if an article has a Bilibili video
 * @param article - The article to check
 * @returns True if the article has a Bilibili video
 */
export function isBilibiliArticle(article: { video_url?: string } | undefined): boolean {
  return article ? isBilibiliUrl(article.video_url) : false;
}

/**
 * Extract BVID from various Bilibili URL formats
 * @param url - The Bilibili URL
 * @returns The BVID or empty string if not found
 */
export function extractBilibiliBVID(url: string): string {
  if (!url) return '';

  for (const pattern of BVID_EXTRACT_PATTERNS) {
    const match = url.match(pattern);
    if (match && match[1]) {
      return match[1];
    }
  }

  return '';
}

/**
 * Get Bilibili embed URL from a video URL
 * @param url - The Bilibili video URL or iframe URL
 * @returns The embed URL or original URL if not a Bilibili URL
 */
export function getBilibiliEmbedUrl(url: string): string {
  if (!url) return '';

  // If it's already an html5mobileplayer URL, return as-is
  if (url.includes('bilibili.com/blackboard/html5mobileplayer.html')) {
    return url;
  }

  // Extract BVID and convert to embed format
  const bvid = extractBilibiliBVID(url);
  if (bvid) {
    return `https://www.bilibili.com/blackboard/html5mobileplayer.html?bvid=${bvid}&autoplay=0`;
  }

  return url;
}

/**
 * Get Bilibili thumbnail URL from a BVID
 * Note: Bilibili thumbnails are extracted from RSS content, not from BVID alone
 * This function is a placeholder for future implementation
 * @param bvid - The Bilibili video BVID
 * @returns Empty string (thumbnails should be extracted from RSS content)
 */
export function getBilibiliThumbnailUrl(bvid: string): string {
  // Bilibili doesn't provide a simple thumbnail URL pattern like YouTube
  // Thumbnails should be extracted from the RSS feed's image content
  return '';
}

/**
 * Extract Bilibili video URL from RSS content/description
 * @param content - The RSS content or description HTML
 * @returns The Bilibili iframe URL or empty string if not found
 */
export function extractBilibiliVideoFromHTML(content: string): string {
  if (!content) return '';

  // Look for Bilibili iframe in the content
  const iframeRegex =
    /<iframe[^>]+src=["']([^"']*bilibili\.com\/blackboard\/html5mobileplayer\.html[^"']*)["'][^>]*>/i;
  const match = content.match(iframeRegex);

  if (match && match[1]) {
    // Unescape HTML entities
    const unescapedUrl = match[1].replace(/&amp;/g, '&');
    return unescapedUrl;
  }

  return '';
}

/**
 * Get the Bilibili watch page URL from a BVID
 * @param bvid - The Bilibili video BVID
 * @returns The watch page URL
 */
export function getBilibiliWatchUrl(bvid: string): string {
  if (!bvid) return '';
  return `https://www.bilibili.com/video/${bvid}`;
}

/**
 * Adds native lazy-loading hints to images inside prepared article HTML.
 *
 * Article HTML arrives already sanitized from the backend; this helper only
 * adds attributes and never alters structure or content. Parsing happens in
 * an inert DOMParser document, so no scripts run and nothing is fetched.
 *
 * `loading="lazy"` only defers offscreen images; images in the initial
 * viewport still load immediately. Offscreen article images are a major
 * source of decoded-bitmap memory in the web view.
 */
export function withLazyImages(html: string): string {
  if (!html || !html.includes('<img')) return html;

  try {
    const doc = new DOMParser().parseFromString(html, 'text/html');
    const images = doc.querySelectorAll('img');
    if (images.length === 0) return html;

    images.forEach((img) => {
      if (!img.hasAttribute('loading')) img.setAttribute('loading', 'lazy');
      if (!img.hasAttribute('decoding')) img.setAttribute('decoding', 'async');
    });

    return doc.body.innerHTML;
  } catch {
    // On any parsing problem, fall back to the original HTML.
    return html;
  }
}

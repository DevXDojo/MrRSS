import { sanitizeArticleHTML } from './articleHtml';

/**
 * Adds native lazy-loading hints to images inside prepared article HTML.
 *
 * Sanitize before parsing into a template: old cached content and parser
 * differences must not turn inert text into active markup.
 *
 * `loading="lazy"` only defers offscreen images; images in the initial
 * viewport still load immediately. Offscreen article images are a major
 * source of decoded-bitmap memory in the web view.
 */
export function withLazyImages(html: string): string {
  html = sanitizeArticleHTML(html);
  if (!html || !/<img\b/i.test(html)) return html;

  try {
    const template = document.createElement('template');
    template.innerHTML = html;
    const images = template.content.querySelectorAll('img');
    if (images.length === 0) return html;

    images.forEach((img) => {
      if (!img.hasAttribute('loading')) img.setAttribute('loading', 'lazy');
      if (!img.hasAttribute('decoding')) img.setAttribute('decoding', 'async');
    });

    return template.innerHTML;
  } catch {
    // Fail closed if the DOM cannot be prepared.
    return '';
  }
}

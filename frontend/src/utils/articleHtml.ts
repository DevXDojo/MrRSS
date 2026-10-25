import DOMPurify from 'dompurify';

const embedHosts = ['youtube.com', 'youtube-nocookie.com', 'player.vimeo.com', 'player.bilibili.com'];

/** Sanitize at the rendering boundary, including content cached by older versions. */
export function sanitizeArticleHTML(html: string | null | undefined): string {
  if (!html) return '';
  const fragment = DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true, mathMl: true },
    ADD_TAGS: ['iframe'],
    ADD_ATTR: ['sandbox', 'allowfullscreen', 'data-math', 'loading', 'decoding', 'referrerpolicy'],
    FORBID_TAGS: ['noscript', 'template', 'style', 'svg', 'annotation-xml'],
    FORBID_ATTR: ['style', 'srcdoc'],
    ALLOW_DATA_ATTR: false,
    RETURN_DOM_FRAGMENT: true,
  });
  fragment.querySelectorAll('iframe').forEach((frame) => {
    try {
      const url = new URL(frame.getAttribute('src') || '');
      if (
        !['http:', 'https:'].includes(url.protocol) ||
        url.username || url.password ||
        !embedHosts.some((host) => url.hostname === host || url.hostname.endsWith(`.${host}`))
      ) {
        frame.remove();
        return;
      }
      frame.setAttribute('sandbox', 'allow-scripts allow-same-origin allow-presentation');
      frame.setAttribute('allowfullscreen', '');
    } catch {
      frame.remove();
    }
  });
  fragment.querySelectorAll('img').forEach((img) => img.setAttribute('referrerpolicy', 'no-referrer'));
  const template = document.createElement('template');
  template.content.append(fragment);
  return template.innerHTML;
}

import { getProxiedMediaUrl } from './mediaProxy';

// Chromium's DownloadURL drag format offers a file to desktop/file-manager
// drop targets. Other WebViews retain their native image drag fallback.
export function setImageDragData(
  event: DragEvent,
  image: HTMLImageElement,
  referer?: string
): void {
  if (!event.dataTransfer) return;
  try {
    const source = new URL(image.currentSrc || image.src, window.location.href);
    if (!['http:', 'https:'].includes(source.protocol)) return;
    const isProxy =
      source.origin === window.location.origin && source.pathname === '/api/media/proxy';
    let original = source;
    if (isProxy) {
      const encoded = source.searchParams.get('url_b64');
      const raw = encoded ? atob(encoded) : source.searchParams.get('url');
      if (raw) original = new URL(raw);
    }
    let name = decodeURIComponent(original.pathname.split('/').pop() || 'image');
    name = Array.from(name, (char) => (char.charCodeAt(0) < 32 ? '_' : char))
      .join('')
      .replace(/[<>:"/\\|?*]/g, '_')
      .replace(/[. ]+$/g, '')
      .slice(0, 120);
    if (!name || /^(con|prn|aux|nul|com[1-9]|lpt[1-9])(?:\.|$)/i.test(name)) name = 'image';
    const download = new URL(
      isProxy ? source.href : getProxiedMediaUrl(source.href, referer),
      window.location.href
    ).href;
    event.dataTransfer.effectAllowed = 'copy';
    event.dataTransfer.setData('DownloadURL', `application/octet-stream:${name}:${download}`);
    event.dataTransfer.setData('text/uri-list', download);
    event.dataTransfer.setData('text/plain', download);
  } catch {
    // Keep native image dragging available if a source cannot be represented.
  }
}

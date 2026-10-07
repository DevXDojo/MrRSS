import { describe, expect, it, vi } from 'vitest';
import { setImageDragData } from './imageDrag';

describe('article image file dragging', () => {
  it('offers a proxy download with a safe original filename instead of a link shortcut', () => {
    const image = document.createElement('img');
    image.src = 'https://example.org/pictures/photo.jpg?token=secret';
    const dataTransfer = { effectAllowed: '', setData: vi.fn() };
    setImageDragData(
      { dataTransfer } as unknown as DragEvent,
      image,
      'https://example.org/article'
    );
    expect(dataTransfer.effectAllowed).toBe('copy');
    const payload = dataTransfer.setData.mock.calls.find(([type]) => type === 'DownloadURL')?.[1];
    expect(payload).toContain('application/octet-stream:photo.jpg:');
    expect(payload).toContain('/api/media/proxy?url_b64=');
  });
  it('does not nest proxies and strips reserved filenames', () => {
    const image = document.createElement('img');
    image.src = `/api/media/proxy?url_b64=${btoa('https://example.org/CON.jpg')}`;
    const dataTransfer = { effectAllowed: '', setData: vi.fn() };
    setImageDragData({ dataTransfer } as unknown as DragEvent, image);
    expect(dataTransfer.setData).toHaveBeenCalledWith(
      'DownloadURL',
      `application/octet-stream:image:${image.src}`
    );
  });
  it('does not offer non-HTTP resources as downloads', () => {
    const image = document.createElement('img');
    image.src = 'file:///private.png';
    const dataTransfer = { effectAllowed: '', setData: vi.fn() };
    setImageDragData({ dataTransfer } as unknown as DragEvent, image);
    expect(dataTransfer.setData).not.toHaveBeenCalled();
  });
});

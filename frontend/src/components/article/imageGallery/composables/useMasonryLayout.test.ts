import { describe, it, expect, vi, afterEach } from 'vitest';
import { ref, nextTick } from 'vue';
import type { Article } from '@/types/models';
import { useMasonryLayout } from './useMasonryLayout';
afterEach(() => vi.unstubAllGlobals());
describe('image masonry', () => {
  it('balances decoded image ratios and deduplicates appended articles', async () => {
    vi.stubGlobal(
      'requestAnimationFrame',
      vi.fn(() => 1)
    );
    vi.stubGlobal('cancelAnimationFrame', vi.fn());
    const articles = ref(
      [1, 2, 3, 4].map((id) => ({ id, published_at: `2026-09-0${5 - id}` }) as Article)
    );
    const layout = useMasonryLayout(articles);
    layout.columnCount.value = 2;
    layout.setImageSize(1, 100, 500);
    layout.setImageSize(2, 500, 100);
    articles.value.push(articles.value[0]);
    layout.arrangeColumns();
    await nextTick();
    expect(layout.columns.value.map((c) => c.map((a) => a.id))).toEqual([[1], [2, 3, 4]]);
    layout.cleanupResizeObserver();
  });
  it('supports narrow and already-mounted containers', () => {
    vi.stubGlobal(
      'requestAnimationFrame',
      vi.fn(() => 1)
    );
    vi.stubGlobal('cancelAnimationFrame', vi.fn());
    const disconnect = vi.fn();
    vi.stubGlobal(
      'ResizeObserver',
      class {
        observe() {}
        disconnect = disconnect;
      }
    );
    const layout = useMasonryLayout(ref([]));
    const container = document.createElement('div');
    Object.defineProperty(container, 'clientWidth', { value: 220 });
    layout.containerRef.value = container;
    expect(() => layout.setupResizeObserver()).not.toThrow();
    expect(layout.columnCount.value).toBe(1);
    layout.cleanupResizeObserver();
    expect(disconnect).toHaveBeenCalled();
  });
});

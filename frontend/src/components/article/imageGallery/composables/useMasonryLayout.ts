import { ref, nextTick, watch } from 'vue';
import type { Article } from '@/types/models';
import type { MasonryLayoutReturn } from '../types';

export function useMasonryLayout(articles: { value: Article[] }): MasonryLayoutReturn {
  const columns = ref<Article[][]>([]);
  const columnCount = ref(1);
  const containerRef = ref<HTMLElement | null>(null);
  const imageDimensions = ref(new Map<number, { width: number; height: number }>());
  let resizeObserver: ResizeObserver | null = null;
  let stopWatch: (() => void) | null = null;
  let frame: number | null = null;
  let lastWidth = 0;

  function arrangeColumns() {
    const width = Math.max(1, ((containerRef.value?.clientWidth || 1000) - 32 - (columnCount.value - 1) * 16) / columnCount.value);
    const cols: Article[][] = Array.from({ length: columnCount.value }, () => []);
    const heights = Array(columnCount.value).fill(0);
    const sorted = [...new Map(articles.value.map((article) => [article.id, article])).values()]
      .sort((a, b) => new Date(b.published_at).getTime() - new Date(a.published_at).getTime() || b.id - a.id);
    for (const article of sorted) {
      const index = heights.indexOf(Math.min(...heights));
      cols[index].push(article);
      const size = imageDimensions.value.get(article.id);
      heights[index] += width * (size ? size.height / size.width : 0.75) + 80;
    }
    columns.value = sorted.length ? cols : [];
    if (!sorted.length) imageDimensions.value.clear();
  }

  // Batch image decodes into a single layout update and keep the visible card anchored.
  function scheduleLayout() {
    if (frame !== null) return;
    frame = requestAnimationFrame(async () => {
      frame = null;
      const container = containerRef.value;
      const top = container?.getBoundingClientRect().top || 0;
      const anchor = container && [...container.querySelectorAll<HTMLElement>('[data-gallery-article]')]
        .find((card) => card.getBoundingClientRect().bottom > top);
      const id = anchor?.dataset.galleryArticle;
      const before = anchor?.getBoundingClientRect().top;
      const scrollTop = container?.scrollTop || 0;
      arrangeColumns();
      await nextTick();
      if (!container || containerRef.value !== container || !id || before === undefined || scrollTop === 0) return;
      const replacement = container.querySelector<HTMLElement>(`[data-gallery-article="${id}"]`);
      if (replacement) container.scrollTop += replacement.getBoundingClientRect().top - before;
    });
  }

  function setImageSize(id: number, width: number, height: number) {
    if (!Number.isFinite(width) || !Number.isFinite(height) || width <= 0 || height <= 0) return;
    const previous = imageDimensions.value.get(id);
    if (previous?.width === width && previous?.height === height) return;
    imageDimensions.value.set(id, { width, height });
    scheduleLayout();
  }

  function calculateColumns() {
    const width = containerRef.value?.clientWidth || 0;
    if (!width || width === lastWidth) return;
    lastWidth = width;
    columnCount.value = Math.max(1, Math.floor((width - 16) / 266));
    scheduleLayout();
  }

  function setupResizeObserver() {
    if (stopWatch) return;
    stopWatch = watch(containerRef, (element) => {
      resizeObserver?.disconnect();
      if (!element) return;
      lastWidth = 0;
      resizeObserver = new ResizeObserver(calculateColumns);
      resizeObserver.observe(element);
      calculateColumns();
    }, { immediate: true });
  }

  function cleanupResizeObserver() {
    stopWatch?.(); stopWatch = null;
    resizeObserver?.disconnect(); resizeObserver = null;
    if (frame !== null) cancelAnimationFrame(frame);
    frame = null;
    containerRef.value = null;
  }

  return { columns, columnCount, containerRef, imageDimensions, setImageSize,
    calculateColumns, arrangeColumns, setupResizeObserver, cleanupResizeObserver };
}

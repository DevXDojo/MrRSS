import { computed, nextTick, onMounted, ref, watch, type ComputedRef, type Ref } from 'vue';
import type { Article } from '@/types/models';

/**
 * Windowed rendering for the article list.
 *
 * The list used to render every loaded article, so a long scroll left thousands
 * of rows mounted and the web view grew by roughly 0.5 MB per row (measured:
 * 683 rows => +359 MB). This composable keeps only the rows around the viewport
 * mounted, which bounds that growth.
 *
 * Row heights are measured from the DOM instead of assumed: the rendered block
 * of an item may include a group header, a summary line or a thumbnail, so the
 * composable derives each item's height from the distance to the next rendered
 * item and falls back to a running average for the rows it has not measured yet.
 */
export function useArticleListWindow(
  items: ComputedRef<Article[]> | Ref<Article[]>,
  scrollEl: Ref<HTMLElement | null>,
  options: { itemSelector?: string } = {}
) {
  const itemSelector = options.itemSelector ?? '[data-article-id]';

  // Lists shorter than this are rendered in full: windowing them would add
  // complexity without a meaningful memory win.
  const VIRTUALIZE_FROM = 240;
  const MIN_RENDERED = 60;
  const MAX_RENDERED = 160;
  const OVERSCAN_ROWS = 12;
  const FALLBACK_ROW_HEIGHT = 96;

  const heights = new Map<number, number>();
  const rangeStart = ref(0);
  const rangeEnd = ref(MAX_RENDERED);
  const topSpacerHeight = ref(0);
  const bottomSpacerHeight = ref(0);
  const isVirtualized = computed(() => items.value.length >= VIRTUALIZE_FROM);

  let cachedHeight = 0;
  let cachedOffsets: number[] | null = null;

  function averageHeight(): number {
    if (heights.size === 0) return FALLBACK_ROW_HEIGHT;
    return cachedHeight / heights.size;
  }

  function invalidateOffsets(): void {
    cachedOffsets = null;
    cachedHeight = 0;
    for (const height of heights.values()) cachedHeight += height;
  }

  /** Cumulative offsets; offsets[i] is the top of item i, offsets[n] the end. */
  function offsets(): number[] {
    if (cachedOffsets && cachedOffsets.length === items.value.length + 1) return cachedOffsets;
    const list = items.value;
    const fallback = averageHeight();
    const result = new Array<number>(list.length + 1);
    let total = 0;
    for (let i = 0; i < list.length; i += 1) {
      result[i] = total;
      total += heights.get(list[i].id) ?? fallback;
    }
    result[list.length] = total;
    cachedOffsets = result;
    return result;
  }

  function indexAt(offset: number): number {
    const list = offsets();
    let low = 0;
    let high = list.length - 1;
    while (low < high) {
      const mid = (low + high + 1) >> 1;
      if (list[mid] <= offset) low = mid;
      else high = mid - 1;
    }
    return low;
  }

  function applyRange(start: number, end: number): void {
    const list = offsets();
    // A caller may hand over a range the current list is too short for (for
    // example right after the list shrank). Clamping keeps the spacers finite
    // and stops an inverted slice from rendering an empty list.
    const count = items.value.length;
    const safeStart = Math.min(Math.max(start, 0), count);
    const safeEnd = Math.min(Math.max(end, safeStart), count);
    // offsets has n + 1 entries, so the total sits at the last index.
    const total = list[list.length - 1] ?? 0;
    rangeStart.value = safeStart;
    rangeEnd.value = safeEnd;
    topSpacerHeight.value = safeStart > 0 ? list[safeStart] : 0;
    bottomSpacerHeight.value = safeEnd < count ? total - list[safeEnd] : 0;
  }

  const windowItems = computed(() => {
    if (!isVirtualized.value) return items.value;
    return items.value.slice(rangeStart.value, rangeEnd.value);
  });

  function measureMountedRows(): void {
    const container = scrollEl.value;
    if (!container || !isVirtualized.value) return;
    const rendered = Array.from(container.querySelectorAll<HTMLElement>(itemSelector));
    let changed = false;
    let index = 0;
    while (index < rendered.length) {
      const rowTop = rendered[index].getBoundingClientRect().top;
      // Multi-column layouts (card mode) mount several items at the same top;
      // they share one row, so the row height is split between them.
      let rowEnd = index + 1;
      while (
        rowEnd < rendered.length &&
        Math.abs(rendered[rowEnd].getBoundingClientRect().top - rowTop) < 1
      ) {
        rowEnd += 1;
      }
      const rowSize = rowEnd - index;
      const nextRowTop =
        rowEnd < rendered.length ? rendered[rowEnd].getBoundingClientRect().top : null;
      const rowHeight =
        nextRowTop !== null
          ? nextRowTop - rowTop
          : Math.max(...rendered.slice(index, rowEnd).map((element) => element.offsetHeight));
      const perItem = rowHeight > 0 ? rowHeight / rowSize : 0;
      for (let item = index; item < rowEnd; item += 1) {
        const id = Number(rendered[item].dataset.articleId);
        if (!Number.isFinite(id)) continue;
        const height = perItem > 0 ? perItem : (heights.get(id) ?? 0);
        if (height > 0 && heights.get(id) !== height) {
          heights.set(id, height);
          changed = true;
        }
      }
      index = rowEnd;
    }
    if (changed) invalidateOffsets();
  }

  /**
   * Recomputes the window from the current scroll position. Sliding the window
   * keeps the retained rows where they are - the leading spacer takes the place
   * of the rows that were unmounted - so the scroll offset is corrected only by
   * how far the anchor row actually moved, which happens when the
   * measured-height model changes.
   */
  function updateFromScroll(): void {
    const container = scrollEl.value;
    if (!container || !isVirtualized.value) {
      if (topSpacerHeight.value !== 0 || bottomSpacerHeight.value !== 0)
        applyRange(0, items.value.length);
      return;
    }
    const { scrollTop, clientHeight } = container;
    const startIndex = Math.max(0, indexAt(scrollTop) - OVERSCAN_ROWS);
    const visibleEnd = indexAt(scrollTop + clientHeight);
    let endIndex = Math.min(items.value.length, visibleEnd + OVERSCAN_ROWS + 1);
    const size = endIndex - startIndex;
    if (size < MIN_RENDERED) endIndex = Math.min(items.value.length, startIndex + MIN_RENDERED);
    if (endIndex - startIndex > MAX_RENDERED)
      endIndex = Math.min(items.value.length, startIndex + MAX_RENDERED);
    if (startIndex === rangeStart.value && endIndex === rangeEnd.value) {
      measureMountedRows();
      return;
    }

    const anchor = container.querySelector<HTMLElement>(itemSelector);
    const anchorId = anchor ? Number(anchor.dataset.articleId) : NaN;
    const anchorTop = anchor ? anchor.getBoundingClientRect().top : null;

    applyRange(startIndex, endIndex);
    void nextTick(() => {
      measureMountedRows();
      if (anchorTop === null || !Number.isFinite(anchorId)) return;
      const moved = container.querySelector<HTMLElement>(
        `${itemSelector}[data-article-id="${anchorId}"]`
      );
      if (!moved) return;
      const drift = moved.getBoundingClientRect().top - anchorTop;
      if (drift !== 0) container.scrollTop += drift;
    });
  }

  /** Brings an item into the rendered window before scrolling it into view. */
  async function ensureArticleVisible(articleId: number): Promise<void> {
    if (!isVirtualized.value) return;
    const index = items.value.findIndex((item) => item.id === articleId);
    if (index === -1) return;
    if (index >= rangeStart.value && index < rangeEnd.value) return;
    const half = Math.floor(MAX_RENDERED / 2);
    const start = Math.max(0, Math.min(index - half, items.value.length - MAX_RENDERED));
    applyRange(start, Math.min(items.value.length, start + MAX_RENDERED));
    await nextTick();
    measureMountedRows();
  }

  function resetWindow(): void {
    heights.clear();
    invalidateOffsets();
    applyRange(0, isVirtualized.value ? MAX_RENDERED : items.value.length);
  }

  watch(
    () => items.value.length,
    () => {
      invalidateOffsets();
      // applyRange clamps, which covers a shorter list as well as a longer one.
      applyRange(rangeStart.value, Math.min(rangeEnd.value, items.value.length));
      updateFromScroll();
    }
  );

  onMounted(() => {
    resetWindow();
  });

  watch(isVirtualized, (virtualized) => {
    if (!virtualized) {
      heights.clear();
      invalidateOffsets();
      applyRange(0, items.value.length);
    } else {
      updateFromScroll();
    }
  });

  return {
    windowItems,
    isVirtualized,
    topSpacerHeight,
    bottomSpacerHeight,
    rangeStart,
    rangeEnd,
    updateFromScroll,
    ensureArticleVisible,
    resetWindow,
    measureMountedRows,
  };
}

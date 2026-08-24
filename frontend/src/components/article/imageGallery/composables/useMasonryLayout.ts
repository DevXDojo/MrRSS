import { ref, nextTick, watch } from 'vue';
import type { Article } from '@/types/models';
import type { MasonryLayoutReturn } from '../types';

const TARGET_COLUMN_WIDTH = 250; // Target column width: 250px for optimal image viewing
const MIN_COLUMNS = 2; // Minimum 2 columns, no maximum

/**
 * Composable for managing masonry layout calculations
 * @param articles - The articles to arrange in masonry layout
 * @returns Masonry layout state and methods
 */
export function useMasonryLayout(articles: { value: Article[] }): MasonryLayoutReturn {
  const columns = ref<Article[][]>([]);
  const columnCount = ref(4);
  const containerRef = ref<HTMLElement | null>(null);
  let resizeObserver: ResizeObserver | null = null;
  let isObserverSetup = false;

  /**
   * Calculate number of columns based on container width dynamically
   */
  function calculateColumns(): void {
    if (!containerRef.value) return;
    const width = containerRef.value.offsetWidth;

    // Calculate columns based on target width
    const calculatedColumns = Math.floor(width / TARGET_COLUMN_WIDTH);

    // Ensure at least MIN_COLUMNS columns
    columnCount.value = Math.max(MIN_COLUMNS, calculatedColumns);

    // Rearrange columns after calculating new count
    arrangeColumns();
  }

  /**
   * Setup resize observer on container
   */
  function setupResizeObserver(): void {
    if (isObserverSetup) return; // Already set up

    // Watch for containerRef to become available
    const stopWatch = watch(
      containerRef,
      (el) => {
        if (el && !isObserverSetup) {
          // Set up the observer
          resizeObserver = new ResizeObserver(() => {
            calculateColumns();
          });
          resizeObserver.observe(el);
          isObserverSetup = true;

          // Calculate columns immediately after setting up observer
          nextTick(() => {
            calculateColumns();
          });

          // Stop watching once observer is set up
          stopWatch();
        }
      },
      { immediate: true }
    );
  }

  /**
   * Arrange articles into columns by time, balancing heights
   */
  function arrangeColumns(): void {
    if (articles.value.length === 0) {
      columns.value = [];
      return;
    }

    // Initialize columns
    const cols: Article[][] = Array.from({ length: columnCount.value }, () => []);
    const colHeights: number[] = Array(columnCount.value).fill(0);

    // Sort articles by published date (newest first)
    const sortedArticles = [...articles.value].sort((a, b) => {
      return new Date(b.published_at).getTime() - new Date(a.published_at).getTime();
    });

    // Place each article in the shortest column
    sortedArticles.forEach((article) => {
      const shortestColIndex = colHeights.indexOf(Math.min(...colHeights));
      cols[shortestColIndex].push(article);
      // Estimate height: 200px for image + 80px for info
      colHeights[shortestColIndex] += 280;
    });

    columns.value = cols;
  }

  /**
   * Cleanup resize observer
   */
  function cleanupResizeObserver(): void {
    if (resizeObserver) {
      if (containerRef.value) {
        resizeObserver.unobserve(containerRef.value);
      }
      resizeObserver.disconnect();
      resizeObserver = null;
    }
    isObserverSetup = false;
  }

  return {
    columns,
    columnCount,
    containerRef,
    calculateColumns,
    arrangeColumns,
    setupResizeObserver,
    cleanupResizeObserver,
  };
}

/**
 * Image Gallery Composable
 *
 * Core logic for the image gallery feature including:
 * - Fetching and paginating gallery articles
 * - Managing article columns for masonry layout
 * - Image count caching
 * - Article actions (favorite, read status)
 */

import { ref, computed, watch, nextTick } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import type { Article } from '@/types/models';

// Constants
export const ITEMS_PER_PAGE = 30;
export const SCROLL_THRESHOLD_PX = 500;

export function useImageGallery() {
  const store = useAppStore();
  const { t } = useI18n();

  // State
  const articles = ref<Article[]>([]);
  const isLoading = ref(false);
  const page = ref(1);
  const hasMore = ref(true);
  const columns = ref<Article[][]>([]);
  const columnCount = ref(4);
  const imageCountCache = ref<Map<number, number>>(new Map());

  // Computed
  const feedId = computed(() => store.currentFeedId);
  const category = computed(() => store.currentCategory);

  /**
   * Fetch image gallery articles from API
   */
  async function fetchImages(loadMore = false) {
    if (isLoading.value) return;

    isLoading.value = true;
    try {
      let url = `/api/articles/images?page=${page.value}&limit=${ITEMS_PER_PAGE}`;
      if (feedId.value) {
        url += `&feed_id=${feedId.value}`;
      } else if (category.value !== null) {
        url += `&category=${encodeURIComponent(category.value)}`;
      }

      const res = await fetch(url);
      if (res.ok) {
        const data = await res.json();

        // Validate that data is an array
        if (!Array.isArray(data)) {
          console.error('API response is not an array:', data);
          return;
        }

        const newArticles = data;

        if (loadMore) {
          articles.value = [...articles.value, ...newArticles];
        } else {
          articles.value = newArticles;
        }

        hasMore.value = newArticles.length >= ITEMS_PER_PAGE;

        // Preload image counts for new articles
        newArticles.forEach((article: Article) => {
          if (!imageCountCache.value.has(article.id)) {
            fetchImageCount(article.id);
          }
        });
      }
    } catch (e) {
      console.error('Failed to load images:', e);
    } finally {
      isLoading.value = false;
    }
  }

  /**
   * Fetch image count for an article
   */
  async function fetchImageCount(articleId: number) {
    try {
      const res = await fetch(`/api/articles/extract-images?id=${articleId}`);
      if (res.ok) {
        const data = await res.json();
        if (data.images && Array.isArray(data.images)) {
          imageCountCache.value.set(articleId, data.images.length);
        }
      }
    } catch (e) {
      console.error('Failed to fetch image count:', e);
    }
  }

  /**
   * Get cached image count for an article
   */
  function getImageCount(article: Article): number {
    return imageCountCache.value.get(article.id) || 1;
  }

  /**
   * Arrange articles into columns for masonry layout
   */
  function arrangeColumns() {
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
   * Calculate number of columns based on container width
   */
  function calculateColumns(containerWidth: number) {
    // Target column width: 250px for optimal image viewing
    // Minimum 2 columns, no maximum
    const targetColumnWidth = 250;
    const calculatedColumns = Math.floor(containerWidth / targetColumnWidth);

    // Ensure at least 2 columns
    columnCount.value = Math.max(2, calculatedColumns);

    // Rearrange columns after calculating new count
    arrangeColumns();
  }

  /**
   * Handle scroll for infinite loading
   */
  function handleScroll(containerEl: HTMLElement | null) {
    if (!containerEl) return;

    const scrollTop = containerEl.scrollTop;
    const containerHeight = containerEl.clientHeight;
    const scrollHeight = containerEl.scrollHeight;

    if (
      scrollTop + containerHeight >= scrollHeight - SCROLL_THRESHOLD_PX &&
      !isLoading.value &&
      hasMore.value
    ) {
      // Increment page before fetching
      const nextPage = page.value + 1;
      page.value = nextPage;
      fetchImages(true);
    }
  }

  /**
   * Toggle article favorite status
   */
  async function toggleFavorite(article: Article, event?: Event) {
    if (event) {
      event.stopPropagation();
    }
    try {
      const res = await fetch(`/api/articles/favorite?id=${article.id}`, {
        method: 'POST',
      });
      if (res.ok) {
        article.is_favorite = !article.is_favorite;
        // Update filter counts after toggling favorite status
        await store.fetchFilterCounts();
      }
    } catch (e) {
      console.error('Failed to toggle favorite:', e);
    }
  }

  /**
   * Mark article as read
   */
  async function markAsRead(article: Article) {
    try {
      const res = await fetch(`/api/articles/read?id=${article.id}&read=true`, {
        method: 'POST',
      });
      if (res.ok) {
        article.is_read = true;
        // Update unread counts after marking as read
        await store.fetchUnreadCounts();
        await store.fetchFilterCounts();
      }
    } catch (e) {
      console.error('Failed to mark as read:', e);
    }
  }

  /**
   * Toggle article read status
   */
  async function toggleReadStatus(article: Article) {
    const newState = !article.is_read;
    article.is_read = newState;
    try {
      await fetch(`/api/articles/read?id=${article.id}&read=${newState}`, {
        method: 'POST',
      });
      // Update unread counts after toggling read status
      await store.fetchUnreadCounts();
      await store.fetchFilterCounts();
    } catch (e) {
      console.error('Error toggling read status:', e);
      // Revert the state change on error
      article.is_read = !newState;
      window.showToast(t('common.errors.savingSettings'), 'error');
    }
  }

  /**
   * Reset gallery state and refetch
   */
  async function resetAndFetch() {
    page.value = 1;
    articles.value = [];
    hasMore.value = true;
    await fetchImages();
    await nextTick();
  }

  // Watch for articles changes and rearrange
  watch(articles, () => {
    nextTick(() => {
      arrangeColumns();
    });
  });

  return {
    // State
    articles,
    isLoading,
    page,
    hasMore,
    columns,
    columnCount,
    imageCountCache,

    // Computed
    feedId,
    category,

    // Methods
    fetchImages,
    fetchImageCount,
    getImageCount,
    arrangeColumns,
    calculateColumns,
    handleScroll,
    toggleFavorite,
    markAsRead,
    toggleReadStatus,
    resetAndFetch,
  };
}

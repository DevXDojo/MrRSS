import { ref, watch } from 'vue';
import type { Article } from '@/types/models';
import type { ImageGalleryDataReturn } from '../types';

const ITEMS_PER_PAGE = 30;

/**
 * Composable for managing image gallery data fetching and state
 * @returns Image gallery data state and methods
 */
export function useImageGalleryData(): ImageGalleryDataReturn {
  const articles = ref<Article[]>([]);
  const isLoading = ref(false);
  const page = ref(1);
  const hasMore = ref(true);
  const imageCountCache = ref<Map<number, number>>(new Map());

  // Load showOnlyUnread preference from localStorage
  const showOnlyUnread = ref<boolean>(
    localStorage.getItem('imageGalleryShowOnlyUnread') === 'true'
  );

  // Watch for changes and save to localStorage
  watch(showOnlyUnread, (newValue) => {
    localStorage.setItem('imageGalleryShowOnlyUnread', String(newValue));
  });

  /**
   * Fetch image gallery articles
   * @param loadMore - Whether to append to existing articles or replace them
   */
  async function fetchImages(loadMore = false): Promise<void> {
    if (isLoading.value) return;

    isLoading.value = true;
    try {
      // Build URL with query parameters
      let url = `/api/articles/images?page=${page.value}&limit=${ITEMS_PER_PAGE}`;

      // Add only_unread filter if enabled
      if (showOnlyUnread.value) {
        url += '&only_unread=true';
      }

      // Add feed_id filter if viewing a specific feed
      const feedId = (window as any).store?.currentFeedId;
      if (feedId) {
        url += `&feed_id=${feedId}`;
      } else {
        // Add category filter if viewing a specific category
        const category = (window as any).store?.currentCategory;
        if (category !== null) {
          url += `&category=${encodeURIComponent(category)}`;
        }
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
   * @param articleId - The article ID to fetch image count for
   */
  async function fetchImageCount(articleId: number): Promise<void> {
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
   * Get image count for an article from cache
   * @param article - The article to get image count for
   * @returns The number of images in the article
   */
  function getImageCount(article: Article): number {
    return imageCountCache.value.get(article.id) || 1;
  }

  /**
   * Refresh the image gallery by resetting and fetching from page 1
   */
  async function refresh(): Promise<void> {
    page.value = 1;
    articles.value = [];
    hasMore.value = true;
    await fetchImages();
  }

  /**
   * Toggle the show only unread filter
   */
  function toggleShowOnlyUnread(): void {
    showOnlyUnread.value = !showOnlyUnread.value;
  }

  return {
    articles,
    isLoading,
    page,
    hasMore,
    imageCountCache,
    showOnlyUnread,
    fetchImages,
    fetchImageCount,
    getImageCount,
    refresh,
    toggleShowOnlyUnread,
  };
}

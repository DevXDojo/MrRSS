/**
 * Image Viewer Composable
 *
 * Core logic for the image viewer modal including:
 * - Zoom and pan functionality
 * - Image navigation (prev/next)
 * - Keyboard shortcuts
 * - Thumbnail strip management
 */

import { ref, computed, watch, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import type { Article } from '@/types/models';

// Constants
export const MIN_SCALE = 0.5;
export const MAX_SCALE = 5;
export const SCALE_STEP = 0.25;

export function useImageViewer() {
  const { t } = useI18n();

  // State
  const showImageViewer = ref(false);
  const selectedArticle = ref<Article | null>(null);
  const allImages = ref<string[]>([]);
  const currentImageIndex = ref(0);
  const currentImageLoading = ref(false);

  // Zoom and pan state
  const scale = ref(1);
  const position = ref<{ x: number; y: number }>({ x: 0, y: 0 });
  const isDragging = ref(false);
  const dragStart = ref<{ x: number; y: number }>({ x: 0, y: 0 });

  // Thumbnail strip state
  const showThumbnailStrip = ref(true);
  const thumbnailStripRef = ref<HTMLElement | null>(null);
  const thumbnailStripWidth = ref(0);

  // Load preferences from localStorage
  const savedShowThumbnailStrip = localStorage.getItem('imageGalleryShowThumbnailStrip');
  if (savedShowThumbnailStrip !== null) {
    showThumbnailStrip.value = savedShowThumbnailStrip === 'true';
  }

  // Watch for changes and save to localStorage
  watch(showThumbnailStrip, (newValue) => {
    localStorage.setItem('imageGalleryShowThumbnailStrip', String(newValue));
  });

  // Computed
  const currentImageUrl = computed(() => {
    if (allImages.value.length > 0 && currentImageIndex.value < allImages.value.length) {
      return allImages.value[currentImageIndex.value];
    }
    return selectedArticle.value?.image_url || '';
  });

  const imageStyle = computed(() => ({
    transform: `translate(${position.value.x}px, ${position.value.y}px) scale(${scale.value})`,
  }));

  const shouldCenterThumbnails = computed(() => {
    if (allImages.value.length === 0) return false;
    // Each thumbnail is 64px (w-16) + 8px (gap-2) = 72px
    const thumbnailWidth = 72;
    const totalThumbnailsWidth = allImages.value.length * thumbnailWidth;
    return totalThumbnailsWidth < thumbnailStripWidth.value;
  });

  /**
   * Fetch all images from article content
   */
  async function fetchArticleImages(article: Article) {
    try {
      const res = await fetch(`/api/articles/extract-images?id=${article.id}`);
      if (res.ok) {
        const data = await res.json();
        if (data.images && Array.isArray(data.images) && data.images.length > 0) {
          allImages.value = data.images;
          // Find the index of the article's main image
          currentImageIndex.value = data.images.findIndex(
            (img: string) => img === article.image_url
          );
          if (currentImageIndex.value < 0) {
            currentImageIndex.value = 0;
          }
        } else {
          // Fallback to just the article's main image
          allImages.value = [article.image_url || ''];
          currentImageIndex.value = 0;
        }
      } else {
        // Fallback on error
        allImages.value = [article.image_url || ''];
        currentImageIndex.value = 0;
      }
    } catch (e) {
      console.error('Failed to fetch article images:', e);
      // Fallback on error
      allImages.value = [article.image_url || ''];
      currentImageIndex.value = 0;
    }
  }

  /**
   * Open image viewer for an article
   */
  async function openImage(article: Article) {
    selectedArticle.value = article;
    showImageViewer.value = true;
    currentImageLoading.value = true;
    // Reset zoom and position
    scale.value = 1;
    position.value = { x: 0, y: 0 };

    // Fetch all images from the article
    await fetchArticleImages(article);
  }

  /**
   * Close image viewer
   */
  function closeImageViewer() {
    showImageViewer.value = false;
    selectedArticle.value = null;
    allImages.value = [];
    currentImageIndex.value = 0;
    scale.value = 1;
    position.value = { x: 0, y: 0 };
  }

  /**
   * Reset view (zoom and position)
   */
  function resetView() {
    scale.value = 1;
    position.value = { x: 0, y: 0 };
  }

  // Zoom functions
  function zoomIn() {
    if (scale.value < MAX_SCALE) {
      scale.value = Math.min(scale.value + SCALE_STEP, MAX_SCALE);
    }
  }

  function zoomOut() {
    if (scale.value > MIN_SCALE) {
      scale.value = Math.max(scale.value - SCALE_STEP, MIN_SCALE);
      // Reset position if zooming out to 1 or less
      if (scale.value <= 1) {
        position.value = { x: 0, y: 0 };
      }
    }
  }

  // Drag functions
  function startDrag(e: MouseEvent) {
    isDragging.value = true;
    dragStart.value = {
      x: e.clientX - position.value.x,
      y: e.clientY - position.value.y,
    };
  }

  function onDrag(e: MouseEvent) {
    if (isDragging.value) {
      position.value = {
        x: e.clientX - dragStart.value.x,
        y: e.clientY - dragStart.value.y,
      };
    }
  }

  function stopDrag() {
    isDragging.value = false;
  }

  /**
   * Handle image load event
   */
  function handleImageLoad() {
    currentImageLoading.value = false;
  }

  /**
   * Handle image error event
   */
  function handleImageError() {
    currentImageLoading.value = false;
  }

  /**
   * Navigate to a specific image index
   */
  function goToImage(index: number) {
    currentImageIndex.value = index;
    currentImageLoading.value = true;
    resetView();
  }

  /**
   * Handle mouse wheel on thumbnail strip for horizontal scrolling
   */
  function handleThumbnailWheel(e: globalThis.WheelEvent) {
    if (!thumbnailStripRef.value) return;

    // Prevent vertical scrolling
    e.preventDefault();

    // Scroll horizontally with smooth behavior
    thumbnailStripRef.value.scrollBy({
      left: e.deltaY,
      behavior: 'smooth',
    });
  }

  // Update thumbnail strip width when the ref is available
  watch(thumbnailStripRef, () => {
    if (thumbnailStripRef.value) {
      thumbnailStripWidth.value = thumbnailStripRef.value.offsetWidth;
    }
  });

  // Watch for allImages changes and update thumbnail strip width
  watch(allImages, async () => {
    await nextTick();
    if (thumbnailStripRef.value) {
      thumbnailStripWidth.value = thumbnailStripRef.value.offsetWidth;
    }
  });

  return {
    // State
    showImageViewer,
    selectedArticle,
    allImages,
    currentImageIndex,
    currentImageLoading,
    scale,
    position,
    isDragging,
    dragStart,
    showThumbnailStrip,
    thumbnailStripRef,
    thumbnailStripWidth,

    // Constants
    MIN_SCALE,
    MAX_SCALE,
    SCALE_STEP,

    // Computed
    currentImageUrl,
    imageStyle,
    shouldCenterThumbnails,

    // Methods
    fetchArticleImages,
    openImage,
    closeImageViewer,
    resetView,
    zoomIn,
    zoomOut,
    startDrag,
    onDrag,
    stopDrag,
    handleImageLoad,
    handleImageError,
    goToImage,
    handleThumbnailWheel,

    // Translation
    t,
  };
}

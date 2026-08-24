<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick, watch } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';
import type { Article } from '@/types/models';

// Import composables
import { useImageGalleryData } from './composables/useImageGalleryData';
import { useMasonryLayout } from './composables/useMasonryLayout';
import { useImageActions } from './composables/useImageActions';
import { useGalleryKeyboard } from './composables/useGalleryKeyboard';

// Import components
import ImageGalleryHeader from './components/ImageGalleryHeader.vue';
import ImageGalleryGrid from './components/ImageGalleryGrid.vue';
import ImageViewerModal from './components/ImageViewerModal.vue';

const store = useAppStore();
const { t } = useI18n();

interface Props {
  isSidebarOpen?: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  toggleSidebar: [];
}>();

// Constants
const SCROLL_THRESHOLD_PX = 500; // Start loading more items when user is 500px from bottom

// Global state for the store (for composables to access)
(window as any).store = store;

// Use composables
const galleryData = useImageGalleryData();
const masonryLayout = useMasonryLayout(galleryData.articles);
const imageActions = useImageActions();

// UI state
const showTextOverlay = ref(true);
const showThumbnailStrip = ref(true);

// Image viewer state
const showImageViewer = ref(false);
const selectedArticle = ref<Article | null>(null);
const allImages = ref<string[]>([]);
const currentImageIndex = ref(0);

// Load showTextOverlay preference from localStorage
const savedShowTextOverlay = localStorage.getItem('imageGalleryShowTextOverlay');
if (savedShowTextOverlay !== null) {
  showTextOverlay.value = savedShowTextOverlay === 'true';
}

// Load showThumbnailStrip preference from localStorage
const savedShowThumbnailStrip = localStorage.getItem('imageGalleryShowThumbnailStrip');
if (savedShowThumbnailStrip !== null) {
  showThumbnailStrip.value = savedShowThumbnailStrip === 'true';
}

// Watch for changes and save to localStorage
watch(showTextOverlay, (newValue) => {
  localStorage.setItem('imageGalleryShowTextOverlay', String(newValue));
});

watch(showThumbnailStrip, (newValue) => {
  localStorage.setItem('imageGalleryShowThumbnailStrip', String(newValue));
});

// Compute which feed ID to fetch (if viewing a specific feed)
const feedId = computed(() => store.currentFeedId);

// Compute which category to fetch (if viewing a specific category)
const category = computed(() => store.currentCategory);

// Find current article index in articles array
const currentArticleIndex = computed(() => {
  if (!selectedArticle.value) return -1;
  return galleryData.articles.value.findIndex((a) => a.id === selectedArticle.value!.id);
});

// Setup keyboard shortcuts
const keyboard = useGalleryKeyboard({
  onClose: closeImageViewer,
  onPrevious: () => handleNavigate('prev'),
  onNext: () => handleNavigate('next'),
  onZoomIn: () => {
    // Handle zoom in via modal
  },
  onZoomOut: () => {
    // Handle zoom out via modal
  },
  isImageViewerOpen: () => showImageViewer.value,
});

/**
 * Handle scroll for infinite loading
 */
function handleScroll(): void {
  if (!masonryLayout.containerRef.value) return;

  const scrollTop = masonryLayout.containerRef.value.scrollTop;
  const containerHeight = masonryLayout.containerRef.value.clientHeight;
  const scrollHeight = masonryLayout.containerRef.value.scrollHeight;

  if (
    scrollTop + containerHeight >= scrollHeight - SCROLL_THRESHOLD_PX &&
    !galleryData.isLoading.value &&
    galleryData.hasMore.value
  ) {
    // Increment page before fetching
    const nextPage = galleryData.page.value + 1;
    galleryData.page.value = nextPage;
    galleryData.fetchImages(true);
  }
}

/**
 * Ensure the container is filled with enough content so users can scroll.
 * If the initial content fits without scrolling, load additional pages
 * until either the container becomes scrollable (beyond threshold) or
 * there is no more content.
 */
async function ensureFill(): Promise<void> {
  const container = masonryLayout.containerRef.value;
  if (!container) return;

  // Keep loading while there are more pages and the container isn't tall enough
  while (
    galleryData.hasMore.value &&
    !galleryData.isLoading.value &&
    container.scrollHeight - container.clientHeight <= SCROLL_THRESHOLD_PX
  ) {
    // Load next page
    galleryData.page.value = galleryData.page.value + 1;
    await galleryData.fetchImages(true);

    // Allow DOM to update and recalculate layout
    await nextTick();
    masonryLayout.arrangeColumns();

    // If container ref was cleared during fetch (unmount), stop
    if (!masonryLayout.containerRef.value) break;
  }
}

/**
 * Open image viewer
 */
async function openImage(article: Article): Promise<void> {
  selectedArticle.value = article;
  showImageViewer.value = true;
  currentImageIndex.value = 0;

  // Fetch all images from the article
  await fetchArticleImages(article);

  // Mark as read
  if (!article.is_read) {
    await imageActions.markAsRead(article);
  }
}

/**
 * Fetch all images from article content
 */
async function fetchArticleImages(article: Article): Promise<void> {
  try {
    const res = await fetch(`/api/articles/extract-images?id=${article.id}`);
    if (res.ok) {
      const data = await res.json();
      if (data.images && Array.isArray(data.images) && data.images.length > 0) {
        allImages.value = data.images;
        // Find the index of the article's main image
        currentImageIndex.value = data.images.findIndex((img: string) => img === article.image_url);
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
 * Handle navigation (prev/next) with cross-article support
 */
async function handleNavigate(direction: 'prev' | 'next'): Promise<void> {
  if (!selectedArticle.value) return;

  if (direction === 'prev') {
    // Check if we can navigate backward
    if (currentImageIndex.value > 0) {
      currentImageIndex.value--;
    } else {
      // At first image of current article, go to previous article
      const prevArticle = galleryData.articles.value[currentArticleIndex.value - 1];
      if (prevArticle) {
        // Update selected article without closing viewer
        const wasFavorite = selectedArticle.value?.is_favorite;
        selectedArticle.value = prevArticle;
        if (wasFavorite !== undefined) {
          selectedArticle.value.is_favorite = wasFavorite;
        }

        await fetchArticleImages(prevArticle);
        // Move to last image
        currentImageIndex.value = allImages.value.length - 1;

        if (!prevArticle.is_read) {
          await imageActions.markAsRead(prevArticle);
        }
      }
    }
  } else {
    // direction === 'next'
    // Check if we can navigate forward
    if (currentImageIndex.value < allImages.value.length - 1) {
      currentImageIndex.value++;
    } else {
      // At last image of current article, go to next article
      const nextArticle = galleryData.articles.value[currentArticleIndex.value + 1];

      // Try to load more articles if available
      if (!nextArticle && galleryData.hasMore.value) {
        await galleryData.fetchImages(true);
      }

      const articleToUse = galleryData.articles.value[currentArticleIndex.value + 1];
      if (articleToUse) {
        // Update selected article without closing viewer
        const wasFavorite = selectedArticle.value?.is_favorite;
        selectedArticle.value = articleToUse;
        if (wasFavorite !== undefined) {
          selectedArticle.value.is_favorite = wasFavorite;
        }

        await fetchArticleImages(articleToUse);
        // Start at first image
        currentImageIndex.value = 0;

        if (!articleToUse.is_read) {
          await imageActions.markAsRead(articleToUse);
        }
      }
    }
  }
}

/**
 * Close image viewer
 */
function closeImageViewer(): void {
  showImageViewer.value = false;
  selectedArticle.value = null;
  allImages.value = [];
  currentImageIndex.value = 0;
}

/**
 * Handle context menu action
 */
async function handleImageAction(action: string, article: Article): Promise<void> {
  if (action === 'toggleRead') {
    await imageActions.toggleReadStatus(article);
  } else if (action === 'toggleFavorite') {
    await imageActions.toggleFavorite(article);
    // Update selected article if it's the same
    if (selectedArticle.value && selectedArticle.value.id === article.id) {
      selectedArticle.value.is_favorite = article.is_favorite;
    }
  } else if (action === 'copyTitle') {
    await imageActions.copyArticleTitle(article);
  } else if (action === 'copyLink') {
    await imageActions.copyArticleLink(article);
  } else if (action === 'downloadImage') {
    await imageActions.downloadImage(article.image_url || '');
  } else if (action === 'openBrowser' || action === 'openOriginal') {
    imageActions.openOriginal(article);
  } else if (action === 'copyImage' && selectedArticle.value) {
    await imageActions.copyImage(selectedArticle.value.image_url || '');
  } else if (action === 'openArticleDetail') {
    openArticleDetail();
  }
}

/**
 * Handle right-click context menu
 */
function handleContextMenu(event: MouseEvent, article: Article): void {
  event.preventDefault();
  event.stopPropagation();

  // Use global context menu system to avoid conflicts with sidebar/context menu
  const menuItems = [
    {
      label: article.is_read ? t('article.action.markAsUnread') : t('article.action.markAsRead'),
      action: 'toggleRead',
      icon: article.is_read ? 'ph-envelope' : 'ph-envelope-open',
    },
    {
      label: article.is_favorite
        ? t('article.action.removeFromFavorites')
        : t('article.imageGallery.addToFavorite'),
      action: 'toggleFavorite',
      icon: 'ph-star',
      iconWeight: article.is_favorite ? 'fill' : 'regular',
      iconColor: article.is_favorite ? 'text-yellow-500' : '',
    },
    { separator: true },
    {
      label: t('common.contextMenu.copyTitle'),
      action: 'copyTitle',
      icon: 'ph-text-t',
    },
    {
      label: t('common.contextMenu.copyLink'),
      action: 'copyLink',
      icon: 'ph-link-simple',
    },
    { separator: true },
    {
      label: t('common.contextMenu.downloadImage'),
      action: 'downloadImage',
      icon: 'PhDownloadSimple',
    },
    {
      label: t('article.action.openInBrowser'),
      action: 'openBrowser',
      icon: 'ph-arrow-square-out',
    },
  ];

  window.dispatchEvent(
    new CustomEvent('open-context-menu', {
      detail: {
        x: event.clientX,
        y: event.clientY,
        items: menuItems,
        data: article,
        callback: handleImageAction,
      },
    })
  );
}

/**
 * Open article in detail view
 */
function openArticleDetail(): void {
  if (!selectedArticle.value) return;

  const article = selectedArticle.value;

  store.selectFeedInArticleList(article.feed_id, article.id);

  // Close the image viewer
  closeImageViewer();
}

/**
 * Handle image index update from thumbnail strip
 */
function handleImageIndexUpdate(index: number): void {
  currentImageIndex.value = index;
}

// Watch for container ref to be set up and add scroll listener
watch(
  () => masonryLayout.containerRef.value,
  (container) => {
    if (container) {
      // Remove any existing listener to avoid duplicates
      container.removeEventListener('scroll', handleScroll);
      // Add the scroll listener
      container.addEventListener('scroll', handleScroll);

      // If the container was just mounted, try to fill it with enough items
      // so that users can scroll even on large viewports.
      // Fire-and-forget: don't block mounting.
      void ensureFill();
    }
  },
  { immediate: true }
);

// Watch for articles changes and rearrange columns
watch(
  () => galleryData.articles.value,
  async () => {
    await nextTick();
    masonryLayout.arrangeColumns();
    // Ensure enough content is loaded so the container can scroll
    await ensureFill();
  }
);

// Watch for feed ID changes and refetch
watch(
  feedId,
  async () => {
    // Close image viewer when switching feeds
    closeImageViewer();

    await galleryData.refresh();
    // Recalculate columns after fetching new articles
    await nextTick();
    masonryLayout.calculateColumns();
  },
  { flush: 'post' }
);

// Watch for category changes and refetch
watch(
  category,
  async () => {
    // Close image viewer when switching categories
    closeImageViewer();

    await galleryData.refresh();
    // Recalculate columns after fetching new articles
    await nextTick();
    masonryLayout.calculateColumns();
  },
  { flush: 'post' }
);

// Watch for showOnlyUnread changes and refetch
watch(
  () => galleryData.showOnlyUnread.value,
  async () => {
    await galleryData.refresh();
    // Recalculate columns after fetching new articles
    await nextTick();
    masonryLayout.calculateColumns();
  }
);

onMounted(async () => {
  // Initial fetch and ensure container is filled
  await galleryData.fetchImages();

  // Recalculate columns after initial fetch
  await nextTick();
  masonryLayout.calculateColumns();

  // Try to load more if initial content doesn't allow scrolling
  await ensureFill();

  // Setup resize observer
  masonryLayout.setupResizeObserver();

  // Enable keyboard shortcuts
  keyboard.enable();
});

onUnmounted(() => {
  if (masonryLayout.containerRef.value) {
    masonryLayout.containerRef.value.removeEventListener('scroll', handleScroll);
  }

  // Cleanup resize observer
  masonryLayout.cleanupResizeObserver();

  // Disable keyboard shortcuts
  keyboard.disable();
});
</script>

<template>
  <div class="flex flex-col flex-1 h-full bg-bg-primary">
    <!-- Header -->
    <ImageGalleryHeader
      :show-text-overlay="showTextOverlay"
      :show-only-unread="galleryData.showOnlyUnread.value"
      @toggle-sidebar="emit('toggleSidebar')"
      @toggle-text-overlay="showTextOverlay = !showTextOverlay"
      @toggle-show-only-unread="galleryData.toggleShowOnlyUnread()"
    />

    <!-- Grid View -->
    <ImageGalleryGrid
      :columns="masonryLayout.columns.value"
      :is-loading="galleryData.isLoading.value"
      :show-text-overlay="showTextOverlay"
      :image-count-cache="galleryData.imageCountCache.value"
      @open-image="openImage"
      @context-menu="handleContextMenu"
      @toggle-favorite="imageActions.toggleFavorite"
      @container-mounted="
        (el) => {
          masonryLayout.containerRef.value = el;
        }
      "
    />

    <!-- Image Viewer Modal -->
    <ImageViewerModal
      v-if="showImageViewer && selectedArticle"
      :article="selectedArticle"
      :all-images="allImages"
      :current-image-index="currentImageIndex"
      :show-thumbnail-strip="showThumbnailStrip"
      :articles="galleryData.articles.value"
      @close="closeImageViewer"
      @navigate="handleNavigate"
      @action="handleImageAction"
      @update-index="handleImageIndexUpdate"
      @toggle-thumbnail-strip="showThumbnailStrip = !showThumbnailStrip"
    />
  </div>
</template>

<style scoped>
/* Prose content styling (kept for compatibility) */
.prose-content {
  line-height: 1.6;
}

.prose-content :deep(img) {
  max-width: 100%;
  height: auto;
}

.prose-content :deep(p) {
  margin-bottom: 0.75rem;
}

.prose-content :deep(a) {
  color: #4daafc;
  text-decoration: underline;
}

.dark-mode .prose-content :deep(a) {
  color: #4daafc;
}
</style>

<script setup lang="ts">
/**
 * Image Gallery View
 *
 * A masonry-style image gallery for browsing RSS articles with images.
 * Uses composables and sub-components for better code organization.
 */

import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { useImageGallery } from '@/composables/article/useImageGallery';
import type { Article } from '@/types/models';
import { PhImage, PhList, PhTextT, PhTextTSlash } from '@phosphor-icons/vue';
import { openInBrowser } from '@/utils/browser';

// Sub-components
import GalleryGrid from './GalleryGrid.vue';
import ImageViewer from './ImageViewer.vue';
const { t } = useI18n();

// Props
interface Props {
  isSidebarOpen?: boolean;
}

defineProps<Props>();

// Emits
const emit = defineEmits<{
  toggleSidebar: [];
}>();

// Use image gallery composable
const {
  articles,
  isLoading,
  hasMore,
  columns,
  imageCountCache,
  feedId,
  category,
  fetchImages,
  calculateColumns,
  handleScroll,
  toggleFavorite,
  markAsRead,
  toggleReadStatus,
  resetAndFetch,
} = useImageGallery();

// Container ref
const containerRef = ref<HTMLElement | null>(null);
// eslint-disable-next-line no-undef
let resizeObserver: ResizeObserver | null = null;

// Text overlay preference
const showTextOverlay = ref(true);

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

// Watch for changes and save to localStorage
watch(showTextOverlay, (newValue) => {
  localStorage.setItem('imageGalleryShowTextOverlay', String(newValue));
});

/**
 * Handle container scroll for infinite loading
 */
function onContainerScroll() {
  handleScroll(containerRef.value);
}

/**
 * Open image viewer for an article
 */
async function openImage(article: Article) {
  selectedArticle.value = article;
  showImageViewer.value = true;

  // Fetch all images from the article
  await fetchArticleImages(article);

  // Mark as read
  if (!article.is_read) {
    markAsRead(article);
  }
}

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
        currentImageIndex.value = data.images.findIndex((img: string) => img === article.image_url);
        if (currentImageIndex.value < 0) {
          currentImageIndex.value = 0;
        }
      } else {
        allImages.value = [article.image_url || ''];
        currentImageIndex.value = 0;
      }
    } else {
      allImages.value = [article.image_url || ''];
      currentImageIndex.value = 0;
    }
  } catch (e) {
    console.error('Failed to fetch article images:', e);
    allImages.value = [article.image_url || ''];
    currentImageIndex.value = 0;
  }
}

/**
 * Close image viewer
 */
function closeImageViewer() {
  showImageViewer.value = false;
  selectedArticle.value = null;
  allImages.value = [];
  currentImageIndex.value = 0;
}

/**
 * Update current image index
 */
function updateCurrentIndex(index: number) {
  currentImageIndex.value = index;
}

/**
 * Update selected article (when navigating between articles)
 */
async function updateArticle(article: Article) {
  selectedArticle.value = article;
  await fetchArticleImages(article);
}

/**
 * Handle toggle favorite with image viewer sync
 */
async function handleToggleFavorite(article: Article, event?: Event) {
  await toggleFavorite(article, event);
  // Sync favorite state if this is the selected article
  if (selectedArticle.value && selectedArticle.value.id === article.id) {
    selectedArticle.value.is_favorite = article.is_favorite;
  }
}

/**
 * Handle context menu
 */
function handleContextMenu(event: MouseEvent, article: Article) {
  event.preventDefault();
  event.stopPropagation();

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
      icon: 'ph-link',
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
      icon: 'ph-globe',
    },
  ];

  window.dispatchEvent(
    new CustomEvent('open-context-menu', {
      detail: {
        x: event.clientX,
        y: event.clientY,
        items: menuItems,
        data: article,
        callback: handleContextAction,
      },
    })
  );
}

/**
 * Handle context menu actions
 */
async function handleContextAction(action: string, article: Article): Promise<void> {
  if (action === 'toggleRead') {
    await toggleReadStatus(article);
  } else if (action === 'toggleFavorite') {
    await toggleFavorite(article);
  } else if (action === 'copyTitle') {
    await copyToClipboard(article.title);
  } else if (action === 'copyLink') {
    await copyToClipboard(article.url);
  } else if (action === 'downloadImage') {
    await downloadImage(article.image_url || '');
  } else if (action === 'openBrowser') {
    openInBrowser(article.url);
  }
}

/**
 * Copy text to clipboard
 */
async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text);
    window.showToast(t('common.toast.copiedToClipboard'), 'success');
  } catch (error) {
    console.error('Failed to copy:', error);
    window.showToast(t('common.errors.failedToCopy'), 'error');
  }
}

/**
 * Download image
 */
async function downloadImage(src: string) {
  try {
    const response = await fetch(src);
    const blob = await response.blob();

    let filename = 'image';
    try {
      const url = new URL(src);
      const pathname = url.pathname;
      const pathSegments = pathname.split('/').filter((segment) => segment.length > 0);
      if (pathSegments.length > 0) {
        const lastSegment = pathSegments[pathSegments.length - 1];
        filename = lastSegment.split('?')[0].replace(/[^a-zA-Z0-9._-]/g, '_') || 'image';
      }
    } catch {
      filename = 'image';
    }

    if (!filename.match(/\.(jpg|jpeg|png|gif|webp|svg|bmp)$/i)) {
      const mimeType = blob.type;
      const ext = mimeType.split('/')[1]?.replace('jpeg', 'jpg') || 'png';
      filename = `${filename}.${ext}`;
    }

    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    window.URL.revokeObjectURL(url);
  } catch (e) {
    console.error('Failed to download image:', e);
    window.showToast(t('common.toast.downloadFailed'), 'error');
  }
}

/**
 * Load more articles
 */
async function loadMore() {
  await fetchImages(true);
}

// Watch for feed ID changes and refetch
watch(feedId, async () => {
  closeImageViewer();
  await resetAndFetch();
  await nextTick();
  if (containerRef.value) {
    calculateColumns(containerRef.value.offsetWidth);
  }
});

// Watch for category changes and refetch
watch(category, async () => {
  closeImageViewer();
  await resetAndFetch();
  await nextTick();
  if (containerRef.value) {
    calculateColumns(containerRef.value.offsetWidth);
  }
});

onMounted(() => {
  fetchImages();
  if (containerRef.value) {
    containerRef.value.addEventListener('scroll', onContainerScroll);

    // Set up ResizeObserver
    // eslint-disable-next-line no-undef
    resizeObserver = new ResizeObserver(() => {
      if (containerRef.value) {
        calculateColumns(containerRef.value.offsetWidth);
      }
    });
    resizeObserver.observe(containerRef.value);
  }
});

onUnmounted(() => {
  if (containerRef.value) {
    containerRef.value.removeEventListener('scroll', onContainerScroll);
  }
  if (resizeObserver && containerRef.value) {
    resizeObserver.unobserve(containerRef.value);
    resizeObserver.disconnect();
    resizeObserver = null;
  }
});
</script>

<template>
  <div class="flex flex-col flex-1 h-full bg-bg-primary">
    <!-- Header -->
    <div
      class="flex-shrink-0 bg-bg-primary border-b border-border p-2 sm:p-4 flex items-center gap-3"
    >
      <button
        class="p-2 rounded-lg hover:bg-bg-tertiary text-text-primary transition-colors md:hidden"
        :title="t('shortcut.toggle.sidebar')"
        @click="emit('toggleSidebar')"
      >
        <PhList :size="24" />
      </button>
      <div class="flex items-center gap-2 sm:gap-2 flex-1">
        <h1 class="text-base sm:text-lg font-bold text-text-primary line-height-fixed-32">
          {{ t('sidebar.activity.imageGallery') }}
        </h1>
      </div>
      <button
        class="p-1 sm:p-1.5 rounded hover:bg-bg-tertiary text-text-primary transition-colors"
        :title="showTextOverlay ? t('setting.reading.hideText') : t('setting.reading.showText')"
        @click="showTextOverlay = !showTextOverlay"
      >
        <PhTextTSlash v-if="showTextOverlay" :size="20" />
        <PhTextT v-else :size="20" />
      </button>
    </div>

    <!-- Scrollable content area -->
    <div ref="containerRef" class="flex-1 overflow-y-scroll scroll-smooth">
      <!-- Masonry Grid -->
      <GalleryGrid
        v-if="articles.length > 0"
        :columns="columns"
        :show-text-overlay="showTextOverlay"
        :image-count-cache="imageCountCache"
        @open-image="openImage"
        @toggle-favorite="handleToggleFavorite"
        @context-menu="handleContextMenu"
      />

      <!-- Empty State -->
      <div
        v-else-if="!isLoading"
        class="flex flex-col items-center justify-center h-full w-full gap-4"
      >
        <PhImage :size="64" class="text-text-secondary opacity-50" />
        <p class="text-text-secondary">{{ t('article.content.noArticles') }}</p>
      </div>

      <!-- Loading Indicator -->
      <div v-if="isLoading" class="flex justify-center py-8">
        <div
          class="w-8 h-8 border-4 border-accent border-t-transparent rounded-full animate-spin"
        ></div>
      </div>
    </div>

    <!-- Image Viewer Modal -->
    <ImageViewer
      v-if="showImageViewer && selectedArticle"
      :article="selectedArticle"
      :all-images="allImages"
      :current-index="currentImageIndex"
      :articles="articles"
      :has-more="hasMore"
      @close="closeImageViewer"
      @update:current-index="updateCurrentIndex"
      @update:article="updateArticle"
      @toggle-favorite="handleToggleFavorite"
      @mark-as-read="markAsRead"
      @load-more="loadMore"
    />
  </div>
</template>

<style scoped>
/* Define keyframes for spinner animation */
@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>

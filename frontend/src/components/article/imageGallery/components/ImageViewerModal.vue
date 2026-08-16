<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhX,
  PhCopy,
  PhDownloadSimple,
  PhStar,
  PhMagnifyingGlassPlus,
  PhMagnifyingGlassMinus,
} from '@phosphor-icons/vue';
import type { Article } from '@/types/models';
import { useImageViewer } from '../composables/useImageViewer';
import ThumbnailStrip from './ThumbnailStrip.vue';
import { getProxiedMediaUrl, isMediaCacheEnabled } from '@/utils/mediaProxy';
import { isYouTubeArticle, extractYouTubeVideoId } from '@/utils/youtube';
import { isBilibiliArticle } from '@/utils/bilibili';

interface Props {
  article: Article | null;
  allImages: string[];
  currentImageIndex: number;
  showThumbnailStrip: boolean;
  articles: Article[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
  navigate: [direction: 'prev' | 'next'];
  action: [action: string, data?: any];
  updateIndex: [index: number];
  toggleThumbnailStrip: [];
}>();

const { t } = useI18n();

// Track media cache setting
const mediaCacheEnabled = ref(false);

// Check if current article is a YouTube video
const isYouTube = computed(() => (props.article ? isYouTubeArticle(props.article) : false));

// Check if current article is a Bilibili video
const isBilibili = computed(() => (props.article ? isBilibiliArticle(props.article) : false));

// Check if current article is any video (YouTube or Bilibili)
const isVideo = computed(() => isYouTube.value || isBilibili.value);

// Get YouTube embed URL if applicable
const youtubeEmbedUrl = computed(() => {
  if (isYouTube.value && props.article?.video_url) {
    const videoId = extractYouTubeVideoId(props.article.video_url);
    if (videoId) {
      return `https://www.youtube.com/embed/${videoId}?autoplay=1`;
    }
  }
  return '';
});

// Get Bilibili video URL (already in embed format)
const bilibiliVideoUrl = computed(() => {
  if (isBilibili.value && props.article?.video_url) {
    return props.article.video_url;
  }
  return '';
});

// Check media cache setting on mount
onMounted(async () => {
  mediaCacheEnabled.value = await isMediaCacheEnabled();
});

// Local state for image index (managed separately for navigation)
const localImageIndex = ref(props.currentImageIndex);

// Sync with prop changes
watch(
  () => props.currentImageIndex,
  (newIndex) => {
    localImageIndex.value = newIndex;
  }
);

// Use image viewer composable
// Pass getter functions to ensure we always get the latest values
const viewer = useImageViewer(
  () => props.allImages,
  localImageIndex,
  () => props.articles,
  () => props.article
);

// Get current image URL
// Cover images (article.image_url) are always cached
// Other images (from article content) are cached only if global media cache is enabled
const currentImageUrl = computed(() => {
  let originalUrl: string;
  let isCoverImage: boolean;

  if (props.allImages.length > 0 && localImageIndex.value < props.allImages.length) {
    originalUrl = props.allImages[localImageIndex.value];
    // Check if this is the cover image
    isCoverImage = originalUrl === props.article?.image_url;
  } else {
    originalUrl = props.article?.image_url || '';
    isCoverImage = true;
  }

  // Cover images are always cached to ensure they display correctly
  if (isCoverImage) {
    return getProxiedMediaUrl(originalUrl, undefined, true);
  }

  // Non-cover images: only cache if global media cache is enabled
  if (mediaCacheEnabled.value) {
    return getProxiedMediaUrl(originalUrl, undefined, true);
  }

  // If cache is disabled for non-cover images, use original URL directly
  return originalUrl;
});

/**
 * Navigate to previous image (with cross-article support)
 */
async function previousImage(): void {
  if (!viewer.canNavigatePrevious.value) return;

  // Reset zoom and position when navigating
  viewer.resetView();

  if (localImageIndex.value > 0) {
    localImageIndex.value--;
    // Reset loading state
    viewer.currentImageLoading.value = true;
    emit('updateIndex', localImageIndex.value);
  } else {
    // At first image of current article, signal parent to navigate to previous article
    emit('navigate', 'prev');
  }
}

/**
 * Navigate to next image (with cross-article support)
 */
async function nextImage(): void {
  if (!viewer.canNavigateNext.value) return;

  // Reset zoom and position when navigating
  viewer.resetView();

  if (localImageIndex.value < props.allImages.length - 1) {
    localImageIndex.value++;
    // Reset loading state
    viewer.currentImageLoading.value = true;
    emit('updateIndex', localImageIndex.value);
  } else {
    // At last image of current article, signal parent to navigate to next article
    emit('navigate', 'next');
  }
}

/**
 * Handle thumbnail selection
 */
function handleThumbnailSelect(index: number): void {
  localImageIndex.value = index;
  viewer.currentImageLoading.value = true;
  viewer.resetView();
  emit('updateIndex', index);
}

/**
 * Handle viewer action (copy, download, favorite, etc.)
 */
function handleViewerAction(action: string): void {
  emit('action', action, props.article);
  // Close any open context menu when clicking action buttons
  document.dispatchEvent(new MouseEvent('click', { bubbles: true }));
}

/**
 * Handle wheel navigation
 */
function handleWheel(e: WheelEvent): void {
  // Determine direction
  const isNavigatingForward = e.deltaY > 0 || e.deltaX > 0;
  const isNavigatingBackward = e.deltaY < 0 || e.deltaX < 0;

  // Check if navigation is possible
  if (isNavigatingForward && !viewer.canNavigateNext.value) return;
  if (isNavigatingBackward && !viewer.canNavigatePrevious.value) return;

  // Prevent default scrolling only if we can navigate
  e.preventDefault();

  // Navigate
  if (isNavigatingForward) {
    nextImage();
  } else if (isNavigatingBackward) {
    previousImage();
  }
}

// Listen for custom wheel navigation event from composable
window.addEventListener('image-wheel-navigate', ((e: CustomEvent) => {
  if (e.detail.direction === 'prev') {
    previousImage();
  } else if (e.detail.direction === 'next') {
    nextImage();
  }
}) as EventListener);
</script>

<template>
  <div
    class="fixed inset-0 z-50 bg-black/90 flex flex-col p-4"
    data-image-viewer="true"
    @click="emit('close')"
  >
    <!-- Top bar: Close button, Image counter, Zoom controls, Action buttons -->
    <div class="relative shrink-0 mb-2" @click.stop>
      <!-- Left: Image counter -->
      <div class="absolute left-0 top-0 flex items-center gap-2">
        <div
          v-if="!isVideo && allImages.length > 1"
          class="px-2 py-1 rounded bg-black/50 text-white text-sm font-medium min-w-[60px] text-center backdrop-blur-sm"
        >
          {{ localImageIndex + 1 }} / {{ allImages.length }}
        </div>
        <div
          v-else-if="isYouTube"
          class="px-2 py-1 rounded bg-red-600/80 text-white text-sm font-medium backdrop-blur-sm flex items-center gap-1"
        >
          YouTube
        </div>
        <div
          v-else-if="isBilibili"
          class="px-2 py-1 rounded bg-pink-600/80 text-white text-sm font-medium backdrop-blur-sm flex items-center gap-1"
        >
          Bilibili
        </div>
      </div>

      <!-- Center: Zoom controls and Action buttons -->
      <div class="flex items-center justify-center gap-2">
        <!-- Zoom controls (only for images) -->
        <template v-if="!isVideo">
          <button
            class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
            :disabled="viewer.scale.value <= 0.5"
            :title="t('common.imageViewer.zoomOut')"
            @click="viewer.zoomOut"
          >
            <PhMagnifyingGlassMinus :size="20" />
          </button>
          <span
            class="px-2 py-1.5 rounded bg-black/50 text-white text-sm font-medium min-w-[60px] text-center"
          >
            {{ Math.round(viewer.scale.value * 100) }}%
          </span>
          <button
            class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
            :disabled="viewer.scale.value >= 5"
            :title="t('common.imageViewer.zoomIn')"
            @click="viewer.zoomIn"
          >
            <PhMagnifyingGlassPlus :size="20" />
          </button>
        </template>

        <!-- Action buttons -->
        <button
          v-if="!isVideo"
          class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
          :title="t('common.contextMenu.copyImage')"
          @click="handleViewerAction('copyImage')"
        >
          <PhCopy :size="20" />
        </button>
        <button
          v-if="!isVideo"
          class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
          :title="t('common.contextMenu.downloadImage')"
          @click="handleViewerAction('downloadImage')"
        >
          <PhDownloadSimple :size="20" />
        </button>
        <button
          class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
          :title="
            article?.is_favorite
              ? t('article.imageGallery.actionUnfavorite')
              : t('article.imageGallery.actionFavorite')
          "
          @click="handleViewerAction('toggleFavorite')"
        >
          <PhStar
            :size="20"
            :weight="article?.is_favorite ? 'fill' : 'regular'"
            :class="article?.is_favorite ? 'text-yellow-500' : 'text-white'"
          />
        </button>
      </div>

      <!-- Right: Close button -->
      <div class="absolute right-0 top-0">
        <button
          class="w-8 h-8 bg-black/50 hover:bg-black/70 rounded-full text-white flex items-center justify-center transition-colors"
          @click="emit('close')"
        >
          <PhX :size="20" />
        </button>
      </div>
    </div>

    <!-- Navigation buttons (available for both images and videos) -->
    <template v-if="viewer.canNavigatePrevious">
      <button
        class="absolute top-[calc(50%-64px-8px)] left-4 -translate-y-1/2 w-12 h-12 rounded text-white text-4xl flex items-center justify-center transition-all duration-200 hover:scale-110 active:scale-95 z-10"
        style="
          text-shadow:
            0 1px 3px rgba(0, 0, 0, 0.8),
            0 1px 2px rgba(0, 0, 0, 0.6);
        "
        @click.stop="previousImage"
      >
        ‹
      </button>
    </template>
    <template v-if="viewer.canNavigateNext">
      <button
        class="absolute top-[calc(50%-64px-8px)] right-4 -translate-y-1/2 w-12 h-12 rounded text-white text-4xl flex items-center justify-center transition-all duration-200 hover:scale-110 active:scale-95 z-10"
        style="
          text-shadow:
            0 1px 3px rgba(0, 0, 0, 0.8),
            0 1px 2px rgba(0, 0, 0, 0.6);
        "
        @click.stop="nextImage"
      >
        ›
      </button>
    </template>

    <!-- Main image area -->
    <div class="flex-1 flex flex-col items-center justify-center min-h-0 relative" @click.stop>
      <!-- YouTube video player -->
      <div
        v-if="isYouTube && youtubeEmbedUrl"
        class="flex-1 flex items-center justify-center w-full min-h-0"
      >
        <div class="relative w-full h-full max-w-5xl max-h-[80vh]">
          <iframe
            :src="youtubeEmbedUrl"
            :title="article?.title || ''"
            class="w-full h-full border-none rounded-lg"
            allow="
              accelerometer;
              autoplay;
              clipboard-write;
              encrypted-media;
              gyroscope;
              picture-in-picture;
              web-share;
            "
            allowfullscreen
          />
        </div>
      </div>

      <!-- Bilibili video player -->
      <div
        v-else-if="isBilibili && bilibiliVideoUrl"
        class="flex-1 flex items-center justify-center w-full min-h-0"
      >
        <div class="relative w-full h-full max-w-5xl max-h-[80vh]">
          <iframe
            :src="bilibiliVideoUrl"
            :title="article?.title || ''"
            class="w-full h-full border-none rounded-lg"
            allow="
              accelerometer;
              autoplay;
              clipboard-write;
              encrypted-media;
              gyroscope;
              picture-in-picture;
              web-share;
            "
            allowfullscreen
            sandbox="allow-forms allow-scripts allow-same-origin allow-presentation"
          />
        </div>
      </div>

      <!-- Image viewer -->
      <div
        v-else
        class="flex-1 flex items-center justify-center w-full min-h-0 overflow-hidden"
        :class="{
          'cursor-grab': !viewer.isDragging.value,
          'cursor-grabbing': viewer.isDragging.value,
        }"
        @wheel="handleWheel"
        @mousedown="viewer.startDrag"
        @mousemove="viewer.onDrag"
        @mouseup="viewer.stopDrag"
        @mouseleave="viewer.stopDrag"
      >
        <!-- Loading placeholder -->
        <div
          v-if="viewer.currentImageLoading.value"
          class="absolute inset-0 flex items-center justify-center z-10"
        >
          <div
            class="w-12 h-12 border-4 border-white/20 border-t-white rounded-full animate-spin"
          ></div>
        </div>

        <img
          :src="currentImageUrl"
          :alt="article?.title || ''"
          class="max-w-full max-h-full object-contain select-none"
          :class="[
            viewer.isDragging.value ? '' : 'transition-transform duration-150',
            { 'opacity-0': viewer.currentImageLoading.value },
          ]"
          :style="viewer.imageStyle.value"
          @load="viewer.handleImageLoad"
          @error="viewer.handleImageError"
          @dragstart.prevent
        />
      </div>

      <!-- Thumbnail strip (only show for images, not videos) -->
      <ThumbnailStrip
        v-if="!isVideo"
        :images="allImages"
        :current-index="localImageIndex"
        :show="showThumbnailStrip"
        :cover-image-u-r-l="article?.image_url"
        @select="handleThumbnailSelect"
        @toggle="emit('toggleThumbnailStrip')"
      />
    </div>

    <!-- Info bar -->
    <div class="mt-2 px-3 py-3 rounded-lg bg-black/60 backdrop-blur-sm shrink-0" @click.stop>
      <!-- Basic info -->
      <div class="flex items-center justify-between gap-4 mb-2">
        <h2 class="text-base font-bold text-white flex-1 line-clamp-2">
          {{ article?.title }}
        </h2>
        <div class="flex items-center gap-2 shrink-0">
          <button
            class="px-3 py-1.5 bg-accent hover:bg-accent-hover text-white rounded-md text-sm whitespace-nowrap transition-colors duration-200"
            @click="handleViewerAction('openOriginal')"
          >
            {{ t('article.action.viewOriginal') }}
          </button>
          <button
            v-if="!isVideo"
            class="px-3 py-1.5 bg-white/10 hover:bg-white/20 text-white rounded-md text-sm whitespace-nowrap transition-all duration-200"
            :title="t('article.action.viewArticle')"
            @click="handleViewerAction('openArticleDetail')"
          >
            {{ t('article.action.viewArticle') }}
          </button>
        </div>
      </div>
      <div class="flex items-center gap-4 text-sm text-white/80">
        <span class="truncate flex-1">{{ article?.feed_title }}</span>
        <span class="shrink-0">{{
          article?.published_at ? new Date(article.published_at).toLocaleDateString() : ''
        }}</span>
      </div>
    </div>
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

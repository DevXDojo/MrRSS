<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAppStore } from '@/stores/app';
import type { Article } from '@/types/models';
import {
  PhHeart,
  PhCopy,
  PhDownloadSimple,
  PhX,
  PhMagnifyingGlassPlus,
  PhMagnifyingGlassMinus,
} from '@phosphor-icons/vue';

const { t } = useI18n();
const store = useAppStore();

// Props
interface Props {
  article: Article;
  allImages: string[];
  currentIndex: number;
  articles: Article[];
  hasMore: boolean;
}

const props = defineProps<Props>();

// Emits
const emit = defineEmits<{
  close: [];
  'update:currentIndex': [index: number];
  'update:article': [article: Article];
  toggleFavorite: [article: Article];
  markAsRead: [article: Article];
  loadMore: [];
}>();

// Constants
const MIN_SCALE = 0.5;
const MAX_SCALE = 5;
const SCALE_STEP = 0.25;

// Zoom and pan state
const scale = ref(1);
const position = ref<{ x: number; y: number }>({ x: 0, y: 0 });
const isDragging = ref(false);
const dragStart = ref<{ x: number; y: number }>({ x: 0, y: 0 });
const currentImageLoading = ref(true);

// Thumbnail strip state
const showThumbnailStrip = ref(true);
const thumbnailStripRef = ref<HTMLElement | null>(null);
const thumbnailStripWidth = ref(0);

// Image container ref
const imageContainerRef = ref<HTMLElement | null>(null);

// Load preferences from localStorage
const savedShowThumbnailStrip = localStorage.getItem('imageGalleryShowThumbnailStrip');
if (savedShowThumbnailStrip !== null) {
  showThumbnailStrip.value = savedShowThumbnailStrip === 'true';
}

// Computed
const currentImageUrl = computed(() => {
  if (props.allImages.length > 0 && props.currentIndex < props.allImages.length) {
    return props.allImages[props.currentIndex];
  }
  return props.article?.image_url || '';
});

const imageStyle = computed(() => ({
  transform: `translate(${position.value.x}px, ${position.value.y}px) scale(${scale.value})`,
}));

const currentArticleIndex = computed(() => {
  if (!props.article) return -1;
  return props.articles.findIndex((a) => a.id === props.article.id);
});

const canNavigatePrevious = computed(() => {
  if (props.currentIndex > 0) return true;
  if (currentArticleIndex.value > 0) return true;
  return false;
});

const canNavigateNext = computed(() => {
  if (props.currentIndex < props.allImages.length - 1) return true;
  if (currentArticleIndex.value >= 0 && currentArticleIndex.value < props.articles.length - 1)
    return true;
  return false;
});

const shouldCenterThumbnails = computed(() => {
  if (props.allImages.length === 0) return false;
  const thumbnailWidth = 72;
  const totalThumbnailsWidth = props.allImages.length * thumbnailWidth;
  return totalThumbnailsWidth < thumbnailStripWidth.value;
});

// Reset view
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

// Image load handlers
function handleImageLoad() {
  currentImageLoading.value = false;
}

function handleImageError() {
  currentImageLoading.value = false;
}

// Navigate to previous image
async function previousImage() {
  if (!canNavigatePrevious.value) return;

  resetView();

  if (props.currentIndex > 0) {
    emit('update:currentIndex', props.currentIndex - 1);
    currentImageLoading.value = true;
  } else {
    // Go to previous article
    const prevArticle = props.articles[currentArticleIndex.value - 1];
    emit('update:article', prevArticle);
    if (!prevArticle.is_read) {
      emit('markAsRead', prevArticle);
    }
  }
}

// Navigate to next image
async function nextImage() {
  if (!canNavigateNext.value) {
    if (props.hasMore) {
      emit('loadMore');
    }
    return;
  }

  resetView();

  if (props.currentIndex < props.allImages.length - 1) {
    emit('update:currentIndex', props.currentIndex + 1);
    currentImageLoading.value = true;
  } else {
    // Go to next article
    const nextArticle = props.articles[currentArticleIndex.value + 1];
    emit('update:article', nextArticle);
    if (!nextArticle.is_read) {
      emit('markAsRead', nextArticle);
    }
  }
}

// Go to specific thumbnail
function goToThumbnail(index: number) {
  emit('update:currentIndex', index);
  currentImageLoading.value = true;
  resetView();
}

// Handle keyboard shortcuts
function handleKeyDown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    e.stopImmediatePropagation();
    emit('close');
    return;
  }

  if (e.key === 'ArrowLeft') {
    e.preventDefault();
    previousImage();
  } else if (e.key === 'ArrowRight') {
    e.preventDefault();
    nextImage();
  } else if (e.key === '+' || e.key === '=') {
    e.preventDefault();
    zoomIn();
  } else if (e.key === '-' || e.key === '_') {
    e.preventDefault();
    zoomOut();
  }
}

// Handle mouse wheel on thumbnail strip
function handleThumbnailWheel(e: WheelEvent) {
  if (!thumbnailStripRef.value) return;
  e.preventDefault();
  thumbnailStripRef.value.scrollBy({
    left: e.deltaY,
    behavior: 'smooth',
  });
}

// Handle mouse wheel on main image
function handleImageWheel(e: WheelEvent) {
  const isNavigatingForward = e.deltaY > 0 || e.deltaX > 0;
  const isNavigatingBackward = e.deltaY < 0 || e.deltaX < 0;

  if (isNavigatingForward && !canNavigateNext.value) return;
  if (isNavigatingBackward && !canNavigatePrevious.value) return;

  e.preventDefault();

  if (isNavigatingForward) {
    nextImage();
  } else if (isNavigatingBackward) {
    previousImage();
  }
}

// Download image
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

// Copy image
async function copyImage(src: string) {
  try {
    const response = await fetch(src);
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    const blob = await response.blob();

    const pngBlob = await new Promise<Blob>((resolve, reject) => {
      const img = new Image();
      img.crossOrigin = 'anonymous';

      img.onload = () => {
        const canvas = document.createElement('canvas');
        canvas.width = img.width;
        canvas.height = img.height;
        const ctx = canvas.getContext('2d');
        if (ctx) {
          ctx.drawImage(img, 0, 0);
          canvas.toBlob((convertedBlob) => {
            if (convertedBlob) {
              resolve(convertedBlob);
            } else {
              reject(new Error('Failed to convert image to PNG'));
            }
          }, 'image/png');
        } else {
          reject(new Error('Failed to get canvas context'));
        }
      };

      img.onerror = () => {
        reject(new Error('Failed to load image for conversion'));
      };

      img.src = URL.createObjectURL(blob);
    });

    await navigator.clipboard.write([
      new ClipboardItem({
        'image/png': pngBlob,
      }),
    ]);

    window.showToast(t('common.toast.copiedToClipboard'), 'success');
  } catch (error) {
    console.error('Failed to copy image:', error);
    window.showToast(t('common.errors.failedToCopy'), 'error');
  }
}

// Open article in detail view
function openArticleDetail() {
  if (!props.article) return;
  store.currentArticleId = props.article.id;
  store.setFilter('all');
  emit('close');
}

// Format date
function formatDate(dateString: string): string {
  const date = new Date(dateString);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const days = Math.floor(diff / (1000 * 60 * 60 * 24));

  if (days === 0) {
    const hours = Math.floor(diff / (1000 * 60 * 60));
    if (hours === 0) {
      const minutes = Math.floor(diff / (1000 * 60));
      return minutes <= 0
        ? t('common.time.justNow')
        : t('common.time.minutesAgo', { count: minutes });
    }
    return t('common.time.hoursAgo', { count: hours });
  } else if (days < 7) {
    return t('common.time.daysAgo', { count: days });
  }
  return date.toLocaleDateString();
}

// Save thumbnail strip preference
function toggleThumbnailStrip() {
  showThumbnailStrip.value = !showThumbnailStrip.value;
  localStorage.setItem('imageGalleryShowThumbnailStrip', String(showThumbnailStrip.value));
}

// Update thumbnail strip width
function updateThumbnailStripWidth() {
  if (thumbnailStripRef.value) {
    thumbnailStripWidth.value = thumbnailStripRef.value.offsetWidth;
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown, { capture: true });
  updateThumbnailStripWidth();
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown, { capture: true });
});
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
          v-if="allImages.length > 1"
          class="px-2 py-1 rounded bg-black/50 text-white text-sm font-medium min-w-[60px] text-center backdrop-blur-sm"
        >
          {{ currentIndex + 1 }} / {{ allImages.length }}
        </div>
      </div>

      <!-- Center: Zoom controls and Action buttons -->
      <div class="flex items-center justify-center gap-2">
        <button
          class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
          :disabled="scale <= MIN_SCALE"
          :title="t('common.imageViewer.zoomOut')"
          @click="zoomOut"
        >
          <PhMagnifyingGlassMinus :size="20" />
        </button>
        <span
          class="px-2 py-1.5 rounded bg-black/50 text-white text-sm font-medium min-w-[60px] text-center"
        >
          {{ Math.round(scale * 100) }}%
        </span>
        <button
          class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
          :disabled="scale >= MAX_SCALE"
          :title="t('common.imageViewer.zoomIn')"
          @click="zoomIn"
        >
          <PhMagnifyingGlassPlus :size="20" />
        </button>

        <button
          class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
          :title="t('common.contextMenu.copyImage')"
          @click="copyImage(currentImageUrl)"
        >
          <PhCopy :size="20" />
        </button>
        <button
          class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
          :title="t('common.contextMenu.downloadImage')"
          @click="downloadImage(currentImageUrl)"
        >
          <PhDownloadSimple :size="20" />
        </button>
        <button
          class="px-2 py-1.5 rounded bg-black/50 hover:bg-black/70 text-white transition-all duration-200 hover:scale-105 active:scale-95"
          :title="
            article.is_favorite
              ? t('article.imageGallery.actionUnfavorite')
              : t('article.imageGallery.actionFavorite')
          "
          @click="emit('toggleFavorite', article)"
        >
          <PhHeart
            :size="20"
            :weight="article.is_favorite ? 'fill' : 'regular'"
            :class="article.is_favorite ? 'text-red-500' : 'text-white'"
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

    <!-- Navigation buttons -->
    <template v-if="canNavigatePrevious">
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
    <template v-if="canNavigateNext">
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

    <div class="flex-1 flex flex-col items-center justify-center min-h-0 relative" @click.stop>
      <div
        ref="imageContainerRef"
        class="flex-1 flex items-center justify-center w-full min-h-0 overflow-hidden"
        :class="{
          'cursor-grab': !isDragging,
          'cursor-grabbing': isDragging,
        }"
        @wheel="handleImageWheel"
        @mousedown="startDrag"
        @mousemove="onDrag"
        @mouseup="stopDrag"
        @mouseleave="stopDrag"
      >
        <!-- Loading placeholder -->
        <div
          v-if="currentImageLoading"
          class="absolute inset-0 flex items-center justify-center z-10"
        >
          <div
            class="w-12 h-12 border-4 border-white/20 border-t-white rounded-full animate-spin"
          ></div>
        </div>

        <img
          :src="currentImageUrl"
          :alt="article.title"
          class="max-w-full max-h-full object-contain select-none"
          :class="[
            isDragging ? '' : 'transition-transform duration-150',
            { 'opacity-0': currentImageLoading },
          ]"
          :style="imageStyle"
          @load="handleImageLoad"
          @error="handleImageError"
          @dragstart.prevent
        />
      </div>

      <!-- Thumbnail strip -->
      <div v-if="allImages.length > 1" class="w-full shrink-0" @click.stop>
        <!-- Collapsed state -->
        <div
          v-if="!showThumbnailStrip"
          class="relative w-full py-3 flex items-center justify-center"
        >
          <div
            class="h-1 w-12 bg-white/30 rounded-full cursor-pointer hover:bg-white/50 hover:w-16 transition-all duration-300"
            @click="toggleThumbnailStrip"
          ></div>
        </div>

        <!-- Expanded state -->
        <template v-else>
          <div class="relative w-full py-2 flex items-center justify-center">
            <div
              class="h-1 w-16 bg-white/20 rounded-full cursor-pointer hover:bg-white/40 hover:w-20 transition-all duration-300"
              @click="toggleThumbnailStrip"
            ></div>
          </div>

          <div class="w-full px-2" @click.stop>
            <div
              ref="thumbnailStripRef"
              class="flex gap-2 overflow-x-auto pb-2 scrollbar-hide scroll-smooth"
              :class="shouldCenterThumbnails ? 'justify-center' : 'justify-start'"
              @wheel="handleThumbnailWheel"
            >
              <button
                v-for="(image, index) in allImages"
                :key="index"
                class="relative shrink-0 w-16 h-16 rounded overflow-hidden border-2 transition-all duration-200"
                :class="
                  index === currentIndex
                    ? 'border-accent shadow-lg shadow-accent/30'
                    : 'border-white/30 hover:border-white/60'
                "
                @click="goToThumbnail(index)"
              >
                <img
                  :src="image"
                  :alt="`${t('common.text.image')} ${index + 1}`"
                  class="w-full h-full object-cover"
                  loading="lazy"
                />
                <div
                  v-if="index === currentIndex"
                  class="absolute inset-0 bg-accent/20 pointer-events-none"
                ></div>
              </button>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- Info bar -->
    <div class="mt-2 px-3 py-3 rounded-lg bg-black/60 backdrop-blur-sm shrink-0" @click.stop>
      <div class="flex items-center justify-between gap-4 mb-2">
        <h2 class="text-base font-bold text-white flex-1 line-clamp-2">
          {{ article.title }}
        </h2>
        <div class="flex items-center gap-2 shrink-0">
          <a
            :href="article.url"
            target="_blank"
            rel="noopener noreferrer"
            class="px-3 py-1.5 bg-accent hover:bg-accent-hover text-white rounded-md text-sm whitespace-nowrap transition-colors duration-200"
          >
            {{ t('article.action.viewOriginal') }}
          </a>
          <button
            class="px-3 py-1.5 bg-white/10 hover:bg-white/20 text-white rounded-md text-sm whitespace-nowrap transition-all duration-200"
            :title="t('article.action.viewArticle')"
            @click="openArticleDetail"
          >
            {{ t('article.action.viewArticle') }}
          </button>
        </div>
      </div>
      <div class="flex items-center gap-4 text-sm text-white/80">
        <span class="truncate flex-1">{{ article.feed_title }}</span>
        <span class="shrink-0">{{ formatDate(article.published_at) }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.scrollbar-hide {
  -ms-overflow-style: none;
  scrollbar-width: none;
}

.scrollbar-hide::-webkit-scrollbar {
  display: none;
}
</style>

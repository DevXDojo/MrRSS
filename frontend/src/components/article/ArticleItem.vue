<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhEyeSlash, PhStar, PhClockCountdown } from '@phosphor-icons/vue';
import type { Article } from '@/types/models';
import { formatDate as formatDateUtil } from '@/utils/date';
import { getProxiedMediaUrl, isMediaCacheEnabled } from '@/utils/mediaProxy';
import { useShowPreviewImages } from '@/composables/ui/useShowPreviewImages';
import { useAppStore } from '@/stores/app';
import { useSettings } from '@/composables/core/useSettings';
import { imageCache } from '@/utils/imageCache';

interface Props {
  article: Article;
  isActive: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  click: [];
  contextmenu: [event: MouseEvent];
  observeElement: [element: Element | null];
  hoverMarkAsRead: [articleId: number];
}>();

const { t, locale } = useI18n();
const { showPreviewImages } = useShowPreviewImages();
const { settings } = useSettings();
const store = useAppStore();

// Compact mode setting (layout_mode === 'compact')
const compactMode = computed(() => {
  return settings.value.layout_mode === 'compact';
});

// Listen for layout mode changes and initial settings load
let handleLayoutModeChange: (() => void) | null = null;

// Function to load layout mode settings
function loadLayoutModeSettings() {
  fetch('/api/settings')
    .then((res) => res.json())
    .then((data) => {
      settings.value = {
        ...settings.value,
        layout_mode: data.layout_mode || 'normal',
      };
    })
    .catch((err) => console.error('Error loading settings in ArticleItem:', err));
}

onMounted(() => {
  // Load settings immediately when component mounts
  loadLayoutModeSettings();

  // Listen for layout mode changes
  handleLayoutModeChange = () => {
    loadLayoutModeSettings();
  };

  window.addEventListener('layout-mode-changed', handleLayoutModeChange);
});

onUnmounted(() => {
  if (handleLayoutModeChange) {
    window.removeEventListener('layout-mode-changed', handleLayoutModeChange);
  }
});

// Check if article is from RSSHub feed - O(1) lookup using feedMap
const isRSSHubArticle = computed(() => {
  // Early return if no feed_title
  if (!props.article.feed_title) return false;

  // Use feedMap for O(1) lookup instead of O(n) find/some
  const feed = store.feedMap.get(props.article.feed_id);
  return feed?.url.startsWith('rsshub://') || false;
});

// Translation function wrapper for formatDate
const formatDateWithI18n = (dateStr: string): string => {
  return formatDateUtil(dateStr, locale.value, t);
};

const mediaCacheEnabled = ref(false);
const hoverMarkAsRead = ref(false);
let hoverTimeout: ReturnType<typeof setTimeout> | null = null;

const imageUrl = computed(() => {
  if (!props.article.image_url) return '';

  const originalUrl = props.article.image_url;
  const finalUrl = mediaCacheEnabled.value
    ? getProxiedMediaUrl(props.article.image_url, props.article.url)
    : originalUrl;

  // Use global cache manager to get the appropriate URL
  return imageCache.getImageUrl(finalUrl);
});

const shouldShowImage = computed(() => {
  return showPreviewImages.value && props.article.image_url;
});

// Track if image has failed to load - use a ref to avoid recomputation
const imageFailed = ref(false);
const imageLoading = ref(true);
// Track if image is in viewport for lazy loading
const imageInViewport = ref(false);
const imageContainerRef = ref<HTMLDivElement | null>(null);

// Shared intersection observer for all ArticleItem instances
let sharedObserver: IntersectionObserver | null = null;
const observerTargets = new WeakMap<Element, () => void>();

onMounted(() => {
  // Use IntersectionObserver to load images only when near viewport
  if ('IntersectionObserver' in window && imageContainerRef.value) {
    // Create or get shared observer
    if (!sharedObserver) {
      sharedObserver = new IntersectionObserver(
        (entries) => {
          entries.forEach((entry) => {
            const callback = observerTargets.get(entry.target);
            if (callback && entry.isIntersecting) {
              callback();
            }
          });
        },
        {
          // Start loading when image is 200px away from viewport
          rootMargin: '200px',
          // Trigger as soon as any part is visible
          threshold: 0,
        }
      );
    }

    // Setup callback for this image
    const callback = () => {
      imageInViewport.value = true;
      // Once loaded, stop observing this specific target
      if (sharedObserver && imageContainerRef.value) {
        sharedObserver.unobserve(imageContainerRef.value);
        observerTargets.delete(imageContainerRef.value);
      }
    };

    observerTargets.set(imageContainerRef.value, callback);
    sharedObserver.observe(imageContainerRef.value);
  } else {
    // Fallback: always load if IntersectionObserver not available
    imageInViewport.value = true;
  }

  // Check media cache setting
  isMediaCacheEnabled().then((enabled) => {
    mediaCacheEnabled.value = enabled;
  });
});

onBeforeUnmount(() => {
  if (sharedObserver && imageContainerRef.value) {
    sharedObserver.unobserve(imageContainerRef.value);
    observerTargets.delete(imageContainerRef.value);
  }
});

function handleImageLoad(event: Event) {
  const target = event.target as HTMLImageElement;
  const url = target.src;

  // Mark as loaded in global cache
  imageCache.markAsLoaded(url);
  imageLoading.value = false;
  imageFailed.value = false;

  // Add fade-in animation
  target.style.opacity = '1';
}

function handleImageError(event: Event) {
  const target = event.target as HTMLImageElement;
  const url = target.src;

  // Mark as failed and stop retrying
  imageLoading.value = false;
  imageFailed.value = true;

  // Update cache to mark as permanently failed
  imageCache.handleLoadError(url);
}

// Hover mark as read functionality
function handleMouseEnter() {
  // Don't mark as read if:
  // - Setting is disabled
  // - Article is already read
  // - Article is in "Read Later" list (user explicitly wants to read it later)
  if (!hoverMarkAsRead.value || props.article.is_read || props.article.is_read_later) {
    return;
  }

  // Use a small delay to avoid marking as read when quickly scrolling through the list
  hoverTimeout = setTimeout(() => {
    markAsRead();
  }, 300);
}

function handleMouseLeave() {
  if (hoverTimeout) {
    clearTimeout(hoverTimeout);
    hoverTimeout = null;
  }
}

async function markAsRead() {
  if (props.article.is_read) return;

  try {
    await fetch(`/api/articles/read?id=${props.article.id}&read=true`, {
      method: 'POST',
    });
    // Emit event to parent to update article state
    emit('hoverMarkAsRead', props.article.id);
    await store.fetchUnreadCounts();
    await store.fetchFilterCounts();
  } catch (e) {
    console.error('Error marking as read on hover:', e);
  }
}

async function loadSettings() {
  try {
    const res = await fetch('/api/settings');
    const data = await res.json();
    hoverMarkAsRead.value = data.hover_mark_as_read === 'true';
  } catch (e) {
    console.error('Error loading hover mark as read setting:', e);
  }
}

onMounted(async () => {
  mediaCacheEnabled.value = await isMediaCacheEnabled();
  await loadSettings();
});

onUnmounted(() => {
  if (hoverTimeout) {
    clearTimeout(hoverTimeout);
  }
});
</script>

<template>
  <div
    :ref="(el) => emit('observeElement', el as Element | null)"
    :data-article-id="article.id"
    :class="[
      'article-card',
      article.is_read ? 'read' : '',
      article.is_favorite ? 'favorite' : '',
      article.is_hidden ? 'hidden' : '',
      article.is_read_later ? 'read-later' : '',
      isActive ? 'active' : '',
      compactMode ? 'compact' : '',
    ]"
    @click="emit('click')"
    @contextmenu="emit('contextmenu', $event)"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave"
  >
    <!-- Image placeholder with lazy loading - hidden completely on error -->
    <div
      v-if="shouldShowImage && !imageFailed"
      ref="imageContainerRef"
      class="article-thumbnail-placeholder"
      :class="{ 'compact-thumbnail': compactMode }"
    >
      <img
        v-if="imageInViewport && imageUrl"
        :src="imageUrl"
        :alt="article.title"
        class="article-thumbnail"
        :class="{ 'image-loaded': !imageLoading }"
        decoding="async"
        @load="handleImageLoad"
        @error="handleImageError"
      />
      <!-- Loading placeholder - only shown while loading -->
      <div
        v-if="imageLoading && imageInViewport"
        class="article-thumbnail article-thumbnail-loading"
      />
    </div>

    <div class="flex-1 min-w-0">
      <div class="flex items-start gap-1.5 sm:gap-2">
        <!-- Normal mode or no translation: single line title -->
        <h4
          v-if="
            !article.translated_title || article.translated_title === article.title || compactMode
          "
          class="flex-1 m-0 text-base sm:text-base font-semibold leading-snug text-text-primary article-title"
          :class="{
            'mb-0.5 sm:mb-1': !compactMode,
            'mb-0': compactMode,
            'compact-title': compactMode,
            'read-title': article.is_read && compactMode,
          }"
        >
          <span
            v-if="article.translated_title && article.translated_title !== article.title"
            :class="{ 'read-translated-title': article.is_read && compactMode }"
          >
            {{ article.translated_title }}
          </span>
          <span v-else>{{ article.title }}</span>
          <span
            v-if="
              compactMode && article.translated_title && article.translated_title !== article.title
            "
            class="original-title-inline"
          >
            {{ article.title }}
          </span>
        </h4>
        <!-- Non-compact mode with translation: separate lines -->
        <div v-else class="flex-1">
          <h4
            class="m-0 mb-0 sm:mb-0.5 text-base sm:text-base font-semibold leading-snug text-text-primary article-title"
          >
            {{ article.translated_title }}
          </h4>
          <div
            class="text-[9px] sm:text-xs text-text-secondary italic mb-0.5 sm:mb-1 article-title"
          >
            {{ article.title }}
          </div>
        </div>
        <PhEyeSlash
          v-if="article.is_hidden"
          :size="18"
          class="text-text-secondary flex-shrink-0 sm:w-5 sm:h-5"
          :title="t('article.action.hideArticle')"
        />
        <!-- Compact mode icons on the right -->
        <div
          v-if="compactMode"
          class="flex items-center gap-1.5 sm:gap-2 shrink-0 ml-1 self-center"
        >
          <PhClockCountdown
            v-if="article.is_read_later"
            :size="16"
            class="text-blue-500"
            weight="fill"
          />
          <PhStar v-if="article.is_favorite" :size="16" class="text-yellow-500" weight="fill" />
          <!-- FreshRSS indicator -->
          <img
            v-if="article.freshrss_item_id"
            src="/assets/plugin_icons/freshrss.svg"
            class="w-4 h-4 shrink-0"
            :title="t('setting.freshrss.syncedFeed')"
            alt="FreshRSS"
          />
          <!-- RSSHub indicator -->
          <img
            v-if="isRSSHubArticle"
            src="/assets/plugin_icons/rsshub.svg"
            class="w-4 h-4 shrink-0"
            :title="t('setting.rsshub.feed')"
            alt="RSSHub"
          />
        </div>
      </div>

      <!-- Feed source name and time - shown in both normal and compact mode -->
      <div
        class="flex justify-between items-center text-[11px] sm:text-xs text-text-secondary"
        :class="{
          'mt-0 sm:mt-1': !compactMode,
          'mt-0': compactMode,
        }"
      >
        <span class="flex items-center gap-1.5 truncate flex-1 min-w-0 mr-2">
          <span class="font-medium text-accent">{{ article.feed_title }}</span>
          <template v-if="article.author && article.author !== article.feed_title">
            <span
              class="text-[11px] sm:text-[11px] text-text-secondary opacity-75 truncate max-w-[120px]"
              >{{ article.author }}</span
            >
          </template>
        </span>
        <div class="flex items-center gap-1 sm:gap-2 shrink-0 min-h-[14px] sm:min-h-[18px]">
          <!-- Icons only shown in normal mode -->
          <template v-if="!compactMode">
            <PhClockCountdown
              v-if="article.is_read_later"
              :size="14"
              class="text-blue-500 sm:w-[18px] sm:h-[18px]"
              weight="fill"
            />
            <PhStar
              v-if="article.is_favorite"
              :size="14"
              class="text-yellow-500 sm:w-[18px] sm:h-[18px]"
              weight="fill"
            />
            <!-- FreshRSS indicator -->
            <img
              v-if="article.freshrss_item_id"
              src="/assets/plugin_icons/freshrss.svg"
              class="w-3.5 h-3.5 shrink-0 sm:w-4 sm:h-4"
              :title="t('setting.freshrss.syncedFeed')"
              alt="FreshRSS"
            />
            <!-- RSSHub indicator -->
            <img
              v-if="isRSSHubArticle"
              src="/assets/plugin_icons/rsshub.svg"
              class="w-3.5 h-3.5 shrink-0 sm:w-4 sm:h-4"
              :title="t('setting.rsshub.feed')"
              alt="RSSHub"
            />
          </template>
          <span class="whitespace-nowrap">{{ formatDateWithI18n(article.published_at) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../style.css";

.article-card {
  @apply py-2 px-1.5 sm:p-3 border-b border-border cursor-pointer transition-colors flex gap-2 sm:gap-3 relative border-l-2 sm:border-l-[3px] border-l-transparent;
}

/* Compact mode: reduce padding */
.article-card.compact {
  @apply py-1 px-2;
}

.article-card:hover {
  @apply bg-bg-tertiary;
}

.article-card.active {
  @apply bg-bg-tertiary border-l-accent;
}

.article-card.read h4 {
  @apply text-text-secondary font-normal;
}

.article-card.read .text-sm {
  @apply text-text-secondary opacity-80;
}

.article-card.favorite {
  background-color: rgba(255, 215, 0, 0.05);
}

.article-card.read-later {
  background-color: rgba(59, 130, 246, 0.05);
}

.article-card.hidden {
  @apply opacity-60 bg-gray-100 dark:bg-gray-800;
}

.article-card.hidden:hover {
  @apply opacity-80;
}

.article-title {
  word-break: break-word;
  overflow-wrap: anywhere;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  display: -webkit-box;
  overflow: hidden;
}

.article-title.compact-title {
  -webkit-line-clamp: 1;
  font-size: 0.875rem; /* 14px, smaller than normal */
}

/* Compact mode: read article title styling */
.article-card.read .read-title {
  @apply text-text-secondary opacity-70;
}

/* Compact mode: translated title when article is read */
.article-card.read .read-translated-title {
  @apply text-text-secondary opacity-75 font-normal;
}

/* Compact mode: original title displayed inline */
.original-title-inline {
  @apply text-text-secondary text-xs font-normal opacity-50 italic ml-2;
}

/* Compact mode: original title when article is read */
.article-card.read .original-title-inline {
  @apply opacity-40;
}

.article-thumbnail {
  @apply w-16 h-12 sm:w-20 sm:h-[60px] object-cover rounded bg-bg-tertiary shrink-0 border border-border;
  /* Performance optimizations */
  contain: layout style paint;
  will-change: auto;
  opacity: 0;
  transition: opacity 0.2s ease-in-out;
}

/* Compact mode: smaller thumbnails */
.article-card.compact .article-thumbnail {
  @apply w-12 h-9 sm:w-14 sm:h-[42px];
}

.article-thumbnail.image-loaded {
  opacity: 1;
}

.article-thumbnail-placeholder {
  @apply w-16 h-12 sm:w-20 sm:h-[60px] shrink-0 border border-border rounded overflow-hidden bg-bg-tertiary;
  /* Prevent layout shift and optimize rendering */
  contain: layout style;
  flex-shrink: 0;
}

/* Compact mode: smaller thumbnails */
.article-thumbnail-placeholder.compact-thumbnail {
  @apply w-12 h-9 sm:w-14 sm:h-[42px];
}

.article-thumbnail-loading {
  @apply w-full h-full bg-bg-tertiary animate-pulse;
  /* Minimal styling for loading state */
  contain: layout style;
}

/* Responsive optimizations for medium screens */
@media (max-width: 1400px) {
  /* Smaller thumbnails */
  .article-thumbnail,
  .article-thumbnail-placeholder {
    width: 56px !important;
    height: 42px !important;
  }

  /* Smaller title font */
  .article-card .article-title {
    font-size: 0.938rem !important; /* 15px, increased from 14px */
  }

  /* Smaller metadata font */
  .article-card .text-xs {
    font-size: 0.688rem !important; /* 11px */
  }

  .article-card .text-\[11px\] {
    font-size: 0.688rem !important; /* 11px */
  }

  /* Tighter spacing */
  .article-card {
    padding: 0.5rem 0.5rem !important; /* py-2 px-1.5 -> py-2 px-1.5 (8px 6px) */
    gap: 0.5rem !important; /* gap-2 sm:gap-3 -> gap-2 */
  }

  /* Smaller icon sizes */
  .article-card .text-sm {
    font-size: 0.75rem !important; /* 12px */
  }

  .article-card .text-\[9px\] {
    font-size: 0.563rem !important; /* 9px */
  }

  .article-card .text-\[11px\] {
    font-size: 0.688rem !important; /* 11px */
  }
}
</style>

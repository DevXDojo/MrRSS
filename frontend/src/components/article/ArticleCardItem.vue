<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhEyeSlash, PhStar, PhClockCountdown } from '@phosphor-icons/vue';
import type { Article } from '@/types/models';
import { formatDate as formatDateUtil } from '@/utils/date';
import { getProxiedMediaUrl, isMediaCacheEnabled } from '@/utils/mediaProxy';
import { useAppStore } from '@/stores/app';
import { imageCache } from '@/utils/imageCache';

interface Props {
  article: Article;
  isActive: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  click: [];
  contextmenu: [event: MouseEvent];
}>();

const { t, locale } = useI18n();
const store = useAppStore();

// Check if article is from RSSHub feed - O(1) lookup using feedMap
const isRSSHubArticle = computed(() => {
  if (!props.article.feed_title) return false;
  const feed = store.feedMap.get(props.article.feed_id);
  return feed?.url.startsWith('rsshub://') || false;
});

// Translation function wrapper for formatDate
const formatDateWithI18n = (dateStr: string): string => {
  return formatDateUtil(dateStr, locale.value, t);
};

const mediaCacheEnabled = ref(false);

const imageUrl = computed(() => {
  if (!props.article.image_url) return '';
  const originalUrl = props.article.image_url;
  const finalUrl = mediaCacheEnabled.value
    ? getProxiedMediaUrl(props.article.image_url, props.article.url)
    : originalUrl;
  return imageCache.getImageUrl(finalUrl);
});

const shouldShowImage = computed(() => {
  return props.article.image_url && !imageFailed.value;
});

// Track if image has failed to load
const imageFailed = ref(false);
const imageLoading = ref(true);
const imageInViewport = ref(false);
const imageContainerRef = ref<HTMLDivElement | null>(null);

// Shared intersection observer for lazy loading
let sharedObserver: IntersectionObserver | null = null;
const observerTargets = new WeakMap<Element, () => void>();

onMounted(() => {
  if ('IntersectionObserver' in window && imageContainerRef.value) {
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
        { rootMargin: '200px', threshold: 0 }
      );
    }

    const callback = () => {
      imageInViewport.value = true;
      if (sharedObserver && imageContainerRef.value) {
        sharedObserver.unobserve(imageContainerRef.value);
        observerTargets.delete(imageContainerRef.value);
      }
    };

    observerTargets.set(imageContainerRef.value, callback);
    sharedObserver.observe(imageContainerRef.value);
  } else {
    imageInViewport.value = true;
  }

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
  imageCache.markAsLoaded(target.src);
  imageLoading.value = false;
  imageFailed.value = false;
  target.style.opacity = '1';
}

function handleImageError(event: Event) {
  const target = event.target as HTMLImageElement;
  imageLoading.value = false;
  imageFailed.value = true;
  imageCache.handleLoadError(target.src);
}
</script>

<template>
  <div
    :data-article-id="article.id"
    :class="[
      'article-card-item',
      article.is_read ? 'read' : '',
      article.is_favorite ? 'favorite' : '',
      article.is_hidden ? 'hidden-article' : '',
      article.is_read_later ? 'read-later' : '',
      isActive ? 'active' : '',
      shouldShowImage ? 'has-thumbnail' : 'no-thumbnail',
    ]"
    @click="emit('click')"
    @contextmenu="emit('contextmenu', $event)"
  >
    <!-- Thumbnail area - only shown when image exists -->
    <div v-if="shouldShowImage" ref="imageContainerRef" class="card-thumbnail">
      <img
        v-if="imageInViewport && imageUrl"
        :src="imageUrl"
        :alt="article.title"
        class="thumbnail-image"
        :class="{ 'image-loaded': !imageLoading }"
        decoding="async"
        @load="handleImageLoad"
        @error="handleImageError"
      />
      <div v-if="imageLoading && imageInViewport" class="thumbnail-loading" />
    </div>

    <!-- Content area -->
    <div class="card-content">
      <!-- Title row with indicators on the right -->
      <div class="card-title-row">
        <div class="card-title-wrapper">
          <!-- Title -->
          <h4 class="card-title" :class="{ 'read-title': article.is_read }">
            <span v-if="article.translated_title && article.translated_title !== article.title">
              {{ article.translated_title }}
            </span>
            <span v-else>{{ article.title }}</span>
          </h4>
          <!-- Original title when translated -->
          <div
            v-if="article.translated_title && article.translated_title !== article.title"
            class="card-original-title"
          >
            {{ article.title }}
          </div>
        </div>
        <!-- Status indicators on the right -->
        <div class="card-indicators">
          <PhEyeSlash
            v-if="article.is_hidden"
            :size="14"
            class="text-text-secondary"
            :title="t('article.action.hideArticle')"
          />
          <PhClockCountdown
            v-if="article.is_read_later"
            :size="14"
            class="text-blue-500"
            weight="fill"
          />
          <PhStar v-if="article.is_favorite" :size="14" class="text-yellow-500" weight="fill" />
          <img
            v-if="article.freshrss_item_id"
            src="/assets/plugin_icons/freshrss.svg"
            class="w-3.5 h-3.5"
            :title="t('setting.freshrss.syncedFeed')"
            alt="FreshRSS"
          />
          <img
            v-if="isRSSHubArticle"
            src="/assets/plugin_icons/rsshub.svg"
            class="w-3.5 h-3.5"
            :title="t('setting.rsshub.feed')"
            alt="RSSHub"
          />
        </div>
      </div>

      <!-- Meta info -->
      <div class="card-meta">
        <span class="feed-name">{{ article.feed_title }}</span>
        <span class="publish-date">{{ formatDateWithI18n(article.published_at) }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../style.css";

.article-card-item {
  @apply bg-bg-primary border border-border rounded-lg overflow-hidden cursor-pointer;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
}

/* All cards have the same fixed height for consistency */
.article-card-item {
  height: 260px;
}

.article-card-item:hover {
  @apply shadow-md;
  border-color: var(--accent-color);
}

.article-card-item.active {
  @apply shadow-md;
  border-color: var(--accent-color);
  box-shadow: 0 0 0 2px rgba(var(--accent-rgb, 59, 130, 246), 0.2);
}

.article-card-item.read {
  @apply opacity-70;
}

.article-card-item.read:hover {
  @apply opacity-100;
}

.article-card-item.favorite {
  background-color: rgba(234, 179, 8, 0.05);
}

.article-card-item.read-later {
  background-color: rgba(59, 130, 246, 0.05);
}

.article-card-item.hidden-article {
  @apply opacity-50;
}

/* Thumbnail area - fixed height */
.card-thumbnail {
  @apply w-full bg-bg-secondary relative overflow-hidden flex-shrink-0;
  height: 130px;
}

.thumbnail-image {
  @apply w-full h-full object-cover;
  opacity: 0;
  transition: opacity 0.2s ease-in-out;
}

.thumbnail-image.image-loaded {
  opacity: 1;
}

.thumbnail-loading {
  @apply w-full h-full bg-bg-tertiary animate-pulse;
}

/* Content area */
.card-content {
  @apply flex flex-col flex-1 min-h-0;
  padding: 1rem;
}

.card-title-row {
  @apply flex items-start gap-2 mb-1.5;
}

.card-title-wrapper {
  @apply flex-1 min-w-0;
}

.card-title {
  @apply text-base font-semibold text-text-primary leading-snug;
  display: -webkit-box;
  -webkit-line-clamp: 4;
  line-clamp: 4;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

/* Allow more lines for cards without thumbnails */
.no-thumbnail .card-title {
  -webkit-line-clamp: 10;
  line-clamp: 10;
}

.card-title.read-title {
  @apply text-text-secondary font-normal;
}

.card-original-title {
  @apply text-xs text-text-secondary italic mt-0.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  word-break: break-word;
}

/* Allow more lines for original title in cards without thumbnails */
.no-thumbnail .card-original-title {
  -webkit-line-clamp: 6;
  line-clamp: 6;
}

.card-indicators {
  @apply flex items-center gap-1 shrink-0 mt-0.5;
}

.card-meta {
  @apply flex items-center justify-between text-xs text-text-secondary gap-2;
  margin-top: auto;
}

.feed-name {
  @apply font-medium text-accent truncate flex-1;
}

.publish-date {
  @apply whitespace-nowrap flex-shrink-0;
}
</style>

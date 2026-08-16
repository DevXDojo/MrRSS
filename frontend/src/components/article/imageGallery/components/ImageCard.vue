<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhImage, PhStar, PhPlay } from '@phosphor-icons/vue';
import { computed } from 'vue';
import type { Article } from '@/types/models';
import { getProxiedMediaUrl } from '@/utils/mediaProxy';
import { isYouTubeArticle, extractYouTubeVideoId, getYouTubeThumbnailUrl } from '@/utils/youtube';
import { isBilibiliArticle } from '@/utils/bilibili';

interface Props {
  article: Article;
  imageCount: number;
  showTextOverlay: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  click: [];
  favorite: [event: Event];
  contextMenu: [event: MouseEvent];
}>();

/**
 * Check if this article has a YouTube video
 */
const isYouTube = computed(() => isYouTubeArticle(props.article));

/**
 * Check if this article has a Bilibili video
 */
const isBilibili = computed(() => isBilibiliArticle(props.article));

/**
 * Check if this article has any video (YouTube or Bilibili)
 */
const isVideo = computed(() => isYouTube.value || isBilibili.value);

/**
 * Get platform badge icon path
 */
const platformBadge = computed(() => {
  if (isYouTube.value) {
    return {
      label: 'YouTube',
      iconPath: '/assets/video_icons/youtube.svg',
    };
  }
  if (isBilibili.value) {
    return {
      label: 'Bilibili',
      iconPath: '/assets/video_icons/bilibili.svg',
    };
  }
  return null;
});

/**
 * Get the display URL (image, YouTube thumbnail, or Bilibili thumbnail)
 */
const displayUrl = computed(() => {
  if (isYouTube.value && props.article.video_url) {
    const videoId = extractYouTubeVideoId(props.article.video_url);
    if (videoId) {
      return getYouTubeThumbnailUrl(videoId, 'high');
    }
  }
  // For Bilibili, use the article's image_url which contains the thumbnail from RSS
  if (isBilibili.value && props.article.image_url) {
    return getProxiedMediaUrl(props.article.image_url, undefined, true);
  }
  // Use proxy with force_cache=true for cover images
  return getProxiedMediaUrl(props.article.image_url || '', undefined, true);
});

/**
 * Handle favorite button click
 * Emits the favorite event and closes any open context menu
 */
function handleFavoriteClick(event: Event): void {
  emit('favorite', event);
  // Close any open context menu by dispatching a click event to document
  // The ContextMenu component's handleClickOutside will catch this
  document.dispatchEvent(new MouseEvent('click', { bubbles: true }));
}

/**
 * Format date for display
 * @param dateString - ISO date string
 * @returns Formatted date string
 */
function formatDate(dateString: string): string {
  const { t } = useI18n();
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
</script>

<template>
  <div
    class="cursor-pointer group"
    @click="emit('click')"
    @contextmenu="emit('contextMenu', $event)"
  >
    <!-- Image container -->
    <div
      class="relative overflow-hidden rounded-lg bg-bg-secondary transition-transform duration-200 hover:scale-[1.02] group/image-container"
    >
      <!-- Hover overlay (gray mask on hover) -->
      <div
        class="absolute inset-0 bg-black/0 group-hover/image-container:bg-black/30 pointer-events-none transition-all duration-200 z-0"
      ></div>

      <img
        :src="displayUrl"
        :alt="article.title"
        class="w-full h-auto block relative z-0"
        loading="lazy"
      />

      <!-- Platform badge (top-left) -->
      <div
        v-if="platformBadge"
        class="absolute top-2 left-2 px-2 py-1 rounded-md bg-white/90 dark:bg-gray-800/90 text-gray-900 dark:text-white text-xs font-semibold shadow-lg z-10 flex items-center gap-1.5 backdrop-blur-sm pointer-events-auto"
      >
        <img :src="platformBadge.iconPath" class="w-4 h-4" alt="" />
        <span>{{ platformBadge.label }}</span>
      </div>

      <!-- Unified play button (center) -->
      <div
        v-if="isVideo"
        class="absolute inset-0 flex items-center justify-center pointer-events-none transition-opacity duration-200 group-hover:opacity-0 z-10"
      >
        <div class="bg-gray-800/70 rounded-full p-4 shadow-lg backdrop-blur-sm">
          <PhPlay :size="32" weight="fill" class="text-white" />
        </div>
      </div>

      <!-- Image count indicator -->
      <div
        v-if="imageCount > 1 && !isVideo"
        class="absolute bottom-2 left-2 px-2 py-1 rounded-full bg-black/60 text-white text-xs font-semibold backdrop-blur-sm z-10 flex items-center gap-1 transition-all duration-200 pointer-events-auto"
        :class="{ 'group-hover:bottom-20': !showTextOverlay }"
      >
        <PhImage :size="14" />
        <span class="ml-1">{{ imageCount }}</span>
      </div>

      <!-- Favorite button -->
      <button
        class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 bg-black/50 rounded-full p-1.5 hover:bg-black/70 z-10 pointer-events-auto"
        @click="handleFavoriteClick($event)"
      >
        <PhStar
          :size="20"
          :weight="article.is_favorite ? 'fill' : 'regular'"
          :class="article.is_favorite ? 'text-yellow-500' : 'text-white'"
        />
      </button>

      <!-- Hover overlay when text is hidden -->
      <div
        v-if="!showTextOverlay"
        class="absolute inset-x-0 bottom-0 p-3 bg-gradient-to-t from-black/80 via-black/50 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-200"
      >
        <p class="text-sm font-medium text-white line-clamp-2 mb-1">
          {{ article.title }}
        </p>
        <div class="flex items-center justify-between text-xs text-white/80">
          <span class="truncate flex-1">{{ article.feed_title }}</span>
          <span class="ml-2 shrink-0">{{ formatDate(article.published_at) }}</span>
        </div>
      </div>
    </div>

    <!-- Text info (always shown when showTextOverlay is true) -->
    <div v-if="showTextOverlay" class="p-2">
      <p class="text-sm font-medium text-text-primary line-clamp-2 mb-1">
        {{ article.title }}
      </p>
      <div class="flex items-center justify-between text-xs text-text-secondary">
        <span class="truncate flex-1">{{ article.feed_title }}</span>
        <span class="ml-2 shrink-0">{{ formatDate(article.published_at) }}</span>
      </div>
    </div>
  </div>
</template>

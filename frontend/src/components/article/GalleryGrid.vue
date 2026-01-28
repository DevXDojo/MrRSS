<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { Article } from '@/types/models';
import { PhImage, PhHeart } from '@phosphor-icons/vue';

const { t } = useI18n();

// Props
interface Props {
  columns: Article[][];
  showTextOverlay: boolean;
  imageCountCache: Map<number, number>;
}

const props = defineProps<Props>();

// Emits
const emit = defineEmits<{
  openImage: [article: Article];
  toggleFavorite: [article: Article, event?: Event];
  contextMenu: [event: MouseEvent, article: Article];
}>();

// Get image count for an article
function getImageCount(article: Article): number {
  return props.imageCountCache.get(article.id) || 1;
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

function handleToggleFavorite(article: Article, event: Event) {
  event.stopPropagation();
  emit('toggleFavorite', article, event);
}
</script>

<template>
  <div class="p-4 flex gap-4">
    <div v-for="(column, colIndex) in columns" :key="colIndex" class="flex-1 flex flex-col gap-4">
      <div
        v-for="article in column"
        :key="article.id"
        class="cursor-pointer group"
        @click="emit('openImage', article)"
        @contextmenu="emit('contextMenu', $event, article)"
      >
        <div
          class="relative overflow-hidden rounded-lg bg-bg-secondary transition-transform duration-200 hover:scale-[1.02]"
        >
          <img
            :src="article.image_url"
            :alt="article.title"
            class="w-full h-auto block"
            loading="lazy"
          />
          <!-- Image count indicator -->
          <div
            v-if="getImageCount(article) > 1"
            class="absolute bottom-2 left-2 px-2 py-1 rounded-full bg-black/60 text-white text-xs font-semibold backdrop-blur-sm z-10 flex items-center gap-1"
          >
            <PhImage :size="14" />
            <span class="ml-1">{{ getImageCount(article) }}</span>
          </div>
          <div
            class="absolute inset-0 bg-black/0 hover:bg-black/30 transition-all duration-200 flex items-start justify-end p-2"
          >
            <button
              class="opacity-0 group-hover:opacity-100 transition-opacity duration-200 bg-black/50 rounded-full p-1.5 hover:bg-black/70"
              @click="handleToggleFavorite(article, $event)"
            >
              <PhHeart
                :size="20"
                :weight="article.is_favorite ? 'fill' : 'regular'"
                :class="article.is_favorite ? 'text-red-500' : 'text-white'"
              />
            </button>
          </div>
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
    </div>
  </div>
</template>

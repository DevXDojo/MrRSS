<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ref, watch } from 'vue';
import { PhCheckCircle, PhImage } from '@phosphor-icons/vue';
import type { Article } from '@/types/models';
import ImageCard from './ImageCard.vue';

interface Props {
  columns: Article[][];
  imageDimensions: Map<number, { width: number; height: number }>;
  isLoading: boolean;
  showOnlyUnread: boolean;
  showTextOverlay: boolean;
  imageCountCache: Map<number, number>;
  showMarkAllRead: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  imageSize: [id: number, width: number, height: number];
  openImage: [article: Article];
  contextMenu: [event: MouseEvent, article: Article];
  toggleFavorite: [article: Article, event: Event];
  containerMounted: [element: HTMLElement];
  markAllRead: [];
}>();

const { t } = useI18n();

// Local ref for the container element
const localContainerRef = ref<HTMLElement | null>(null);

// Emit event when container is mounted so parent can set up its ref
watch(
  localContainerRef,
  (newVal) => {
    if (newVal) {
      emit('containerMounted', newVal);
    }
  },
  { immediate: true }
);

/**
 * Get image count for an article
 */
function getImageCount(article: Article): number {
  return props.imageCountCache.get(article.id) || 1;
}
</script>

<template>
  <div
    ref="localContainerRef"
    class="flex-1 min-h-0 min-w-0 overflow-y-auto"
    data-testid="gallery-scroll"
  >
    <!-- Masonry Grid -->
    <div v-if="columns.length > 0 && columns.some((col) => col.length > 0)" class="p-4 flex gap-4">
      <div
        v-for="(column, colIndex) in columns"
        :key="colIndex"
        class="flex-1 min-w-0 flex flex-col gap-4"
      >
        <ImageCard
          v-for="article in column"
          :key="article.id"
          :article="article"
          :image-size="imageDimensions.get(article.id)"
          :image-count="getImageCount(article)"
          :show-text-overlay="showTextOverlay"
          @image-size="(width, height) => emit('imageSize', article.id, width, height)"
          @click="emit('openImage', article)"
          @context-menu="emit('contextMenu', $event, article)"
          @favorite="emit('toggleFavorite', article, $event)"
        />
      </div>
    </div>

    <div v-if="showMarkAllRead" class="px-4 pb-6 text-center">
      <button
        type="button"
        class="inline-flex cursor-pointer items-center gap-2 rounded-lg bg-accent px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-accent-hover"
        @click="emit('markAllRead')"
      >
        <PhCheckCircle :size="18" />
        {{ t('article.imageGallery.markAllRead') }}
      </button>
    </div>

    <!-- Empty State -->
    <div
      v-else-if="!isLoading"
      class="flex min-h-full w-full flex-col items-center justify-center p-6 text-center text-text-secondary"
      data-testid="gallery-empty"
    >
      <template v-if="showOnlyUnread">
        <PhCheckCircle :size="40" weight="duotone" class="mb-3 text-green-500" />
        <div class="text-base font-medium text-text-primary">
          {{ t('article.list.allCaughtUp') }}
        </div>
        <div class="mt-1 text-sm">{{ t('article.list.noUnreadArticles') }}</div>
      </template>
      <template v-else>
        <PhImage :size="64" class="mb-4 opacity-50" />
        <p>{{ t('article.content.noArticles') }}</p>
      </template>
    </div>

    <!-- Loading Indicator -->
    <div v-if="isLoading" class="flex justify-center py-8">
      <div
        class="w-8 h-8 border-4 border-accent border-t-transparent rounded-full animate-spin"
      ></div>
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

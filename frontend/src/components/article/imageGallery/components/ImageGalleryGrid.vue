<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ref, watch } from 'vue';
import { PhImage } from '@phosphor-icons/vue';
import type { Article } from '@/types/models';
import ImageCard from './ImageCard.vue';

interface Props {
  columns: Article[][];
  isLoading: boolean;
  showTextOverlay: boolean;
  imageCountCache: Map<number, number>;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  openImage: [article: Article];
  contextMenu: [event: MouseEvent, article: Article];
  toggleFavorite: [article: Article, event: Event];
  containerMounted: [element: HTMLElement];
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
  <div ref="localContainerRef" class="flex-1 overflow-y-auto scroll-smooth" style="min-height: 0">
    <!-- Masonry Grid -->
    <div v-if="columns.length > 0 && columns.some((col) => col.length > 0)" class="p-4 flex gap-4">
      <div v-for="(column, colIndex) in columns" :key="colIndex" class="flex-1 flex flex-col gap-4">
        <ImageCard
          v-for="article in column"
          :key="article.id"
          :article="article"
          :image-count="getImageCount(article)"
          :show-text-overlay="showTextOverlay"
          @click="emit('openImage', article)"
          @context-menu="emit('contextMenu', $event, article)"
          @favorite="emit('toggleFavorite', article, $event)"
        />
      </div>
    </div>

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

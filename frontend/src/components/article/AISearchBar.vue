<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhMagnifyingGlass, PhX, PhSparkle, PhSpinner } from '@phosphor-icons/vue';
import type { Article } from '@/types/models';

const { t } = useI18n();

const emit = defineEmits<{
  search: [articles: Article[]];
  clear: [];
}>();

// State
const searchQuery = ref('');
const isSearching = ref(false);
const hasResults = ref(false);
const errorMessage = ref('');

// Computed
const canSearch = computed(() => searchQuery.value.trim().length > 0 && !isSearching.value);

// Methods
async function performAISearch() {
  if (!canSearch.value) return;

  isSearching.value = true;
  errorMessage.value = '';

  try {
    const response = await fetch('/api/ai/search', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ query: searchQuery.value.trim() }),
    });

    const data = await response.json();

    if (!data.success) {
      errorMessage.value = data.error || t('aiSearch.searchFailed');
      window.showToast(errorMessage.value, 'error');
      return;
    }

    // Convert response to Article format
    const articles: Article[] = data.articles.map((item: Record<string, unknown>) => ({
      id: item.id as number,
      feed_id: item.feed_id as number,
      title: item.title as string,
      url: item.url as string,
      image_url: item.image_url as string,
      audio_url: item.audio_url as string,
      video_url: item.video_url as string,
      published_at: item.published_at as string,
      is_read: item.is_read as boolean,
      is_favorite: item.is_favorite as boolean,
      is_hidden: item.is_hidden as boolean,
      is_read_later: item.is_read_later as boolean,
      feed_title: item.feed_title as string,
      author: item.author as string,
      translated_title: item.translated_title as string,
      summary: item.summary as string,
    }));

    hasResults.value = true;
    emit('search', articles);

    if (articles.length === 0) {
      window.showToast(t('aiSearch.noResults'), 'info');
    } else {
      window.showToast(t('aiSearch.foundResults', { count: articles.length }), 'success');
    }
  } catch (error) {
    console.error('AI Search error:', error);
    errorMessage.value = t('aiSearch.searchFailed');
    window.showToast(errorMessage.value, 'error');
  } finally {
    isSearching.value = false;
  }
}

function clearSearch() {
  searchQuery.value = '';
  hasResults.value = false;
  errorMessage.value = '';
  emit('clear');
}

function handleKeyDown(event: KeyboardEvent) {
  if (event.key === 'Enter' && canSearch.value) {
    performAISearch();
  } else if (event.key === 'Escape') {
    clearSearch();
  }
}
</script>

<template>
  <div class="ai-search-bar border-b border-border bg-bg-primary">
    <div class="flex items-center">
      <!-- Search Input Container -->
      <div class="relative flex-1">
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="t('aiSearch.placeholder')"
          class="w-full bg-bg-tertiary px-3 py-2 pl-8 text-sm focus:outline-none transition-colors"
          :disabled="isSearching"
          @keydown="handleKeyDown"
        />
        <PhMagnifyingGlass
          :size="14"
          class="absolute left-2.5 top-1/2 -translate-y-1/2 text-text-secondary"
        />
        <!-- Clear button -->
        <button
          v-if="searchQuery || hasResults"
          class="absolute right-2 top-1/2 -translate-y-1/2 p-1 text-text-secondary hover:text-text-primary transition-colors"
          :title="t('common.clear')"
          @click="clearSearch"
        >
          <PhX :size="12" />
        </button>
      </div>

      <!-- AI Search Button -->
      <button
        class="ai-search-button flex items-center gap-1 px-2.5 py-2 text-sm transition-colors flex-shrink-0"
        :class="[
          canSearch ? 'text-accent hover:bg-accent/10' : 'text-text-tertiary cursor-not-allowed',
        ]"
        :disabled="!canSearch"
        :title="t('aiSearch.buttonTitle')"
        @click="performAISearch"
      >
        <PhSpinner v-if="isSearching" :size="16" class="animate-spin" />
        <PhSparkle v-else :size="16" />
        <span class="hidden sm:inline">{{ t('aiSearch.button') }}</span>
      </button>
    </div>

    <!-- Results indicator -->
    <div
      v-if="hasResults"
      class="px-3 py-1.5 text-xs text-accent flex items-center gap-1 border-t border-border/50"
    >
      <PhSparkle :size="12" />
      <span>{{ t('aiSearch.showingResults') }}</span>
      <button class="ml-1 underline hover:text-accent/80" @click="clearSearch">
        {{ t('aiSearch.clearResults') }}
      </button>
    </div>
  </div>
</template>

<style scoped>
@reference "../../style.css";

.ai-search-bar {
  flex-shrink: 0;
}

.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>

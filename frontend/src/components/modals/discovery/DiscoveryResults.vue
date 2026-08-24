<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { DiscoveredFeed } from '@/composables/discovery/useDiscoverAllFeeds';
import DiscoveredFeedItem from './DiscoveredFeedItem.vue';

const { t } = useI18n();

interface Props {
  discoveredFeeds: DiscoveredFeed[];
  selectedFeeds: Set<number>;
  allSelected: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  toggleFeedSelection: [index: number];
  selectAll: [];
}>();

function toggleFeedSelection(index: number) {
  emit('toggleFeedSelection', index);
}

function selectAll() {
  emit('selectAll');
}
</script>

<template>
  <div class="mb-4 flex items-center justify-between bg-bg-secondary rounded-lg p-3">
    <p class="text-sm font-medium text-text-primary">
      {{ t('modal.discovery.foundFeeds', { count: discoveredFeeds.length }) }}
    </p>
    <button
      class="text-sm text-accent hover:text-accent-hover font-medium px-3 py-1 rounded hover:bg-accent/10 transition-colors"
      @click="selectAll"
    >
      {{ allSelected ? t('common.action.deselectAll') : t('common.search.selectAll') }}
    </button>
  </div>

  <div class="space-y-3">
    <DiscoveredFeedItem
      v-for="(feed, index) in discoveredFeeds"
      :key="index"
      :feed="feed"
      :is-selected="selectedFeeds.has(index)"
      @toggle="toggleFeedSelection(index)"
    />
  </div>
</template>

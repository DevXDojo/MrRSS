<script setup lang="ts">
import { computed } from 'vue';
import { watch, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhX } from '@phosphor-icons/vue';
import type { Feed } from '@/types/models';
import DiscoveredFeedItem from './DiscoveredFeedItem.vue';
import DiscoveryProgress from './DiscoveryProgress.vue';
import BaseModal from '@/components/common/BaseModal.vue';
import ModalFooter from '@/components/common/ModalFooter.vue';
import { useFeedDiscovery } from '@/composables/discovery/useFeedDiscovery';
import { useFeedSubscription } from '@/composables/discovery/useFeedSubscription';

const { t } = useI18n();

interface Props {
  feed: Feed;
  show: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
}>();

// Use discovery composable
const {
  isDiscovering,
  discoveredFeeds,
  errorMessage,
  progressMessage,
  progressDetail,
  progressCounts,
  startDiscovery,
  cleanup: cleanupDiscovery,
} = useFeedDiscovery(props.feed);

// Use subscription composable
const {
  selectedFeeds,
  isSubscribing,
  hasSelection,
  allSelected,
  toggleFeedSelection,
  selectAll,
  subscribeSelected,
} = useFeedSubscription(props.feed, discoveredFeeds);

function close() {
  // Clear polling interval if active
  cleanupDiscovery();
  emit('close');
}

// Computed subscribe button text
const subscribeButtonText = computed(() => {
  if (isSubscribing.value) {
    return t('modal.feed.subscribing');
  }
  return t('modal.feed.subscribeSelected');
});

// Computed subscribe button label with count
const subscribeButtonLabel = computed(() => {
  const baseText = subscribeButtonText.value;
  if (hasSelection.value && !isSubscribing.value) {
    return `${baseText} (${selectedFeeds.size})`;
  }
  return baseText;
});

// Auto-start discovery when component is mounted
onMounted(() => {
  if (props.show) {
    startDiscovery();
  }
});

// Watch for modal opening and trigger discovery (for when modal is reused)
watch(
  () => props.show,
  (newShow, oldShow) => {
    if (newShow && !oldShow) {
      startDiscovery();
    }
  }
);

// Cleanup on unmount
onUnmounted(() => {
  cleanupDiscovery();
});
</script>

<template>
  <BaseModal v-if="show" size="4xl" :z-index="50" @close="close">
    <!-- Custom Header with gradient background -->
    <template #header>
      <div
        class="flex justify-between items-center bg-gradient-to-r from-accent/5 to-transparent -m-3 sm:-m-5 p-3 sm:p-5 mb-3 sm:mb-0"
      >
        <div class="min-w-0 flex-1">
          <h2 class="text-base sm:text-xl font-bold text-text-primary">
            {{ t('modal.discovery.discoverFeeds') }}
          </h2>
          <p class="text-xs sm:text-sm text-text-secondary mt-1 truncate">
            {{ t('modal.filter.fromFeed') }}: {{ feed.title }}
          </p>
        </div>
        <button
          class="p-1.5 sm:p-2 hover:bg-bg-tertiary rounded-lg transition-colors shrink-0 ml-2"
          @click="close"
        >
          <PhX :size="20" class="sm:w-6 sm:h-6 text-text-secondary" />
        </button>
      </div>
    </template>

    <!-- Content -->
    <div class="p-4 sm:p-6">
      <!-- Loading State -->
      <DiscoveryProgress
        v-if="isDiscovering"
        :progress-message="progressMessage"
        :progress-detail="progressDetail"
        :progress-counts="progressCounts"
      />

      <!-- Error State -->
      <div
        v-else-if="errorMessage"
        class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-3 sm:p-4 text-red-600 dark:text-red-400 text-sm sm:text-base"
      >
        {{ errorMessage }}
      </div>

      <!-- Results -->
      <div v-else-if="discoveredFeeds.length > 0">
        <div
          class="mb-3 sm:mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 bg-bg-secondary rounded-lg p-2 sm:p-3"
        >
          <p class="text-xs sm:text-sm font-medium text-text-primary">
            {{ t('modal.discovery.foundFeeds', { count: discoveredFeeds.length }) }}
          </p>
          <button
            class="text-xs sm:text-sm text-accent hover:text-accent-hover font-medium px-2 sm:px-3 py-1 rounded hover:bg-accent/10 transition-colors"
            @click="selectAll"
          >
            {{ allSelected ? t('common.action.deselectAll') : t('common.search.selectAll') }}
          </button>
        </div>

        <div class="space-y-2 sm:space-y-3">
          <DiscoveredFeedItem
            v-for="(discoveredFeed, index) in discoveredFeeds"
            :key="index"
            :feed="discoveredFeed"
            :is-selected="selectedFeeds.has(index)"
            @toggle="toggleFeedSelection(index)"
          />
        </div>
      </div>

      <!-- Initial State (should not be visible as discovery auto-starts) -->
      <div v-else class="text-center py-12 sm:py-16">
        <div
          class="w-12 h-12 sm:w-16 sm:h-16 border-4 border-accent border-t-transparent rounded-full animate-spin mx-auto mb-3 sm:mb-4"
        ></div>
        <p class="text-text-secondary text-base sm:text-lg">
          {{ t('common.pagination.preparing') }}...
        </p>
      </div>
    </div>

    <!-- Footer -->
    <template #footer>
      <ModalFooter
        align="space-between"
        :secondary-button="{
          label: t('common.cancel'),
          disabled: isSubscribing,
          onClick: close,
        }"
        :primary-button="{
          label: subscribeButtonLabel,
          disabled: !hasSelection || isSubscribing,
          loading: isSubscribing,
          onClick: subscribeSelected,
        }"
      >
        <template #right>
          <button
            :disabled="!hasSelection || isSubscribing"
            :class="[
              'btn-primary flex items-center justify-center gap-2 text-sm sm:text-base',
              (!hasSelection || isSubscribing) && 'opacity-50 cursor-not-allowed',
            ]"
            @click="subscribeSelected"
          >
            <div
              v-if="isSubscribing"
              class="w-3 h-3 sm:w-4 sm:h-4 border-2 border-white border-t-transparent rounded-full animate-spin"
            ></div>
            {{ subscribeButtonText }}
            <span
              v-if="hasSelection && !isSubscribing"
              class="bg-white/20 px-1.5 sm:px-2 py-0.5 rounded-full text-xs sm:text-sm"
              >({{ selectedFeeds.size }})</span
            >
          </button>
        </template>
      </ModalFooter>
    </template>
  </BaseModal>
</template>

<style scoped>
@reference "../../../style.css";

.btn-primary {
  @apply px-4 sm:px-6 py-2 sm:py-2.5 bg-accent text-white rounded-lg hover:bg-accent-hover transition-all font-medium shadow-sm hover:shadow-md;
}
</style>

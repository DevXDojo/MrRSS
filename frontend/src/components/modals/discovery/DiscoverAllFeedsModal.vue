<script setup lang="ts">
import { computed } from 'vue';
import { watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhCircleNotch } from '@phosphor-icons/vue';
import { useDiscoverAllFeeds } from '@/composables/discovery/useDiscoverAllFeeds';
import DiscoveryProgress from './DiscoveryProgress.vue';
import DiscoveryResults from './DiscoveryResults.vue';
import BaseModal from '@/components/common/BaseModal.vue';
import ModalFooter from '@/components/common/ModalFooter.vue';

const { t } = useI18n();

interface Props {
  show: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
}>();

const {
  isDiscovering,
  discoveredFeeds,
  selectedFeeds,
  errorMessage,
  progressMessage,
  progressDetail,
  progressCounts,
  isSubscribing,
  hasSelection,
  allSelected,
  startDiscovery,
  toggleFeedSelection,
  selectAll,
  subscribeSelected,
  cleanup,
} = useDiscoverAllFeeds();

function close() {
  cleanup();
  emit('close');
}

// Computed subscribe button label with count
const subscribeButtonLabel = computed(() => {
  if (isSubscribing.value) {
    return t('modal.feed.subscribing');
  }
  const baseText = t('modal.feed.subscribeSelected');
  if (hasSelection.value) {
    return `${baseText} (${selectedFeeds.value.size})`;
  }
  return baseText;
});

// Auto-start discovery when component is mounted and shown
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
</script>

<template>
  <BaseModal v-if="show" size="4xl" :z-index="70" @close="close">
    <!-- Custom Header with gradient background -->
    <template #header>
      <div class="bg-gradient-to-r from-accent/5 to-transparent -m-3 sm:-m-5 p-3 sm:p-5 mb-3 sm:mb-0">
        <h2 class="text-base sm:text-xl font-bold text-text-primary">
          {{ t('modal.discovery.discoverAllFeeds') }}
        </h2>
        <p class="text-xs sm:text-sm text-text-secondary mt-1">
          {{ t('modal.discovery.discoverAllFeedsDesc') }}
        </p>
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
      <DiscoveryResults
        v-else-if="discoveredFeeds.length > 0"
        :discovered-feeds="discoveredFeeds"
        :selected-feeds="selectedFeeds"
        :all-selected="allSelected"
        @toggle-feed-selection="toggleFeedSelection"
        @select-all="selectAll"
      />

      <!-- Empty / fallback (normally skipped because discovery auto-starts) -->
      <div v-else class="text-center py-12 sm:py-16">
        <PhCircleNotch
          :size="48"
          class="sm:w-16 sm:h-16 text-accent mx-auto mb-3 sm:mb-4 animate-spin"
        />
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
      />
    </template>
  </BaseModal>
</template>

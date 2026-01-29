<script setup lang="ts">
import { computed } from 'vue';
import { watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhX, PhCircleNotch } from '@phosphor-icons/vue';
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

// Computed subscribe button text
const subscribeButtonText = computed(() => {
  if (isSubscribing) {
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
      <div
        class="flex justify-between items-center bg-gradient-to-r from-accent/5 to-transparent -m-3 sm:-m-5 p-3 sm:p-5 mb-3 sm:mb-0"
      >
        <div class="min-w-0 flex-1">
          <h2 class="text-base sm:text-xl font-bold text-text-primary">
            {{ t('modal.discovery.discoverAllFeeds') }}
          </h2>
          <p class="text-xs sm:text-sm text-text-secondary mt-1">
            {{ t('modal.discovery.discoverAllFeedsDesc') }}
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
      <DiscoveryResults
        v-if="discoveredFeeds.length > 0"
        :discovered-feeds="discoveredFeeds"
        :selected-feeds="selectedFeeds"
        :all-selected="allSelected"
        @toggle-feed-selection="toggleFeedSelection"
        @select-all="selectAll"
      />

      <!-- Initial State (should not be visible as discovery auto-starts) -->
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
            <PhCircleNotch v-if="isSubscribing" :size="16" class="animate-spin" />
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

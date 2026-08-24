<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhCircleNotch } from '@phosphor-icons/vue';

const { t } = useI18n();

interface ProgressCounts {
  current: number;
  total: number;
  found: number;
}

interface Props {
  progressMessage: string;
  progressDetail: string;
  progressCounts: ProgressCounts;
}

defineProps<Props>();
</script>

<template>
  <div class="flex flex-col items-center justify-center py-12">
    <PhCircleNotch :size="48" class="text-accent animate-spin mb-4" />
    <p class="text-text-primary font-medium mb-2">{{ t('modal.discovery.discovering') }}</p>
    <p v-if="progressMessage" class="text-sm text-text-secondary">{{ progressMessage }}</p>
    <p v-if="progressDetail" class="text-xs text-text-tertiary mt-1 font-mono">
      {{ progressDetail }}
    </p>

    <div v-if="progressCounts.total > 0" class="mt-4 w-full max-w-md">
      <div class="w-full bg-bg-tertiary rounded-full h-2 overflow-hidden mb-2">
        <div
          class="bg-accent h-full transition-all duration-300"
          :style="{ width: (progressCounts.current / progressCounts.total) * 100 + '%' }"
        ></div>
      </div>
      <div class="flex justify-between text-xs text-text-tertiary">
        <span>{{ progressCounts.current }}/{{ progressCounts.total }}</span>
        <span v-if="progressCounts.found > 0">
          {{ t('modal.discovery.foundSoFar', { count: progressCounts.found }) }}
        </span>
      </div>
    </div>
  </div>
</template>

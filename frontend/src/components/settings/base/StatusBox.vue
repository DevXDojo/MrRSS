<script setup lang="ts">
import { computed } from 'vue';
import type { Component } from 'vue';
import { PhCheckCircle, PhWarningCircle } from '@phosphor-icons/vue';

export interface Status {
  type?: 'success' | 'error' | 'neutral' | 'warning';
  label: string;
  value: string | number;
  unit?: string;
  icon?: Component;
}

interface Props {
  status: Status;
  showIcon?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  showIcon: true,
});

// Export for StatusBoxGroup import
defineOptions({
  name: 'StatusBox',
});

const statusClass = computed(() => {
  switch (props.status.type) {
    case 'success':
      return 'border-green-500/30 text-green-500';
    case 'error':
      return 'border-red-500/30 text-red-500';
    case 'warning':
      return 'border-yellow-500/30 text-yellow-500';
    default:
      return 'border-border';
  }
});

const displayIcon = computed(() => {
  if (props.status.type === 'success') return PhCheckCircle;
  if (props.status.type === 'error' || props.status.type === 'warning') return PhWarningCircle;
  return props.status.icon;
});
</script>

<template>
  <div
    class="status-box flex flex-col gap-2 p-3 rounded-lg bg-bg-primary border w-full sm:min-w-[120px]"
    :class="statusClass"
  >
    <span class="text-sm text-text-secondary text-left">{{ status.label }}</span>
    <div class="flex items-center gap-2">
      <component :is="displayIcon" v-if="showIcon && displayIcon" :size="20" class="shrink-0" />
      <div class="flex items-baseline gap-1">
        <span class="text-xl sm:text-2xl font-bold text-text-primary">{{ status.value }}</span>
        <span v-if="status.unit" class="text-sm text-text-secondary">{{ status.unit }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.status-box {
  transition:
    border-color 0.2s,
    color 0.2s;
}
</style>

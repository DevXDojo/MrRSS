<script setup lang="ts">
import type { Component } from 'vue';
import StatusBox from './StatusBox.vue';
import type { Status } from './StatusBox.vue';

interface ActionButton {
  label: string;
  icon?: Component;
  loading?: boolean;
  disabled?: boolean;
  onClick: () => void;
}

interface StatusInfo {
  label: string;
  time?: string;
}

interface Props {
  statuses: Status[];
  actionButton?: ActionButton;
  statusInfo?: StatusInfo;
}

defineProps<Props>();

// Export for component usage
defineOptions({
  name: 'StatusBoxGroup',
});
</script>

<template>
  <div
    class="status-box-group flex flex-col sm:flex-row sm:items-stretch sm:justify-between gap-3 sm:gap-4 p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border"
  >
    <!-- Status Boxes -->
    <div class="flex flex-col sm:flex-row items-center gap-3 sm:gap-4">
      <StatusBox v-for="(status, index) in statuses" :key="index" :status="status" />
    </div>

    <!-- Action Area: Button and Status Info -->
    <div v-if="actionButton" class="flex flex-col sm:justify-between flex-1 gap-2">
      <div class="flex justify-center sm:justify-end">
        <button
          class="btn-secondary"
          :disabled="actionButton.disabled || actionButton.loading"
          @click="actionButton.onClick"
        >
          <component
            :is="actionButton.icon"
            v-if="actionButton.icon"
            :size="16"
            class="sm:w-5 sm:h-5"
            :class="{ 'animate-spin': actionButton.loading }"
          />
          <span>{{ actionButton.loading ? '...' : actionButton.label }}</span>
        </button>
      </div>

      <!-- Status Info (e.g., last test time) -->
      <div v-if="statusInfo" class="flex items-center justify-center sm:justify-end gap-2">
        <span class="text-xs text-text-secondary">{{ statusInfo.label }}</span>
        <span v-if="statusInfo.time" class="text-xs text-accent font-medium">{{
          statusInfo.time
        }}</span>
      </div>
    </div>
    <!-- Status Info only (no action button) -->
    <div v-else-if="statusInfo" class="flex items-center justify-center sm:justify-end gap-2">
      <span class="text-xs text-text-secondary">{{ statusInfo.label }}</span>
      <span v-if="statusInfo.time" class="text-xs text-accent font-medium">{{
        statusInfo.time
      }}</span>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../style.css";

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.animate-spin {
  animation: spin 1s linear infinite;
}
</style>

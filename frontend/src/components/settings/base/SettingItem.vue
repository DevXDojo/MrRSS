<script setup lang="ts">
import type { Component } from 'vue';

interface Props {
  icon?: Component;
  title: string;
  description?: string;
  required?: boolean;
  layout?: 'row' | 'column';
}

withDefaults(defineProps<Props>(), {
  icon: undefined,
  description: '',
  required: false,
  layout: 'row',
});
</script>

<template>
  <div class="setting-item" :class="{ 'setting-item-column': layout === 'column' }">
    <div
      class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0"
      :class="{ 'w-full': layout === 'column' }"
    >
      <component
        :is="icon"
        v-if="icon"
        :size="20"
        class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6"
      />
      <div class="flex-1 min-w-0">
        <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
          {{ title }} <span v-if="required" class="text-red-500">*</span>
        </div>
        <slot name="description">
          <div v-if="description" class="text-xs text-text-secondary hidden sm:block">
            {{ description }}
          </div>
        </slot>
      </div>
    </div>
    <div v-if="layout === 'row'" class="setting-item-action">
      <slot name="action">
        <slot />
      </slot>
    </div>
    <div v-else class="setting-item-action setting-item-action-column">
      <slot name="action">
        <slot />
      </slot>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../style.css";
.setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-3 bg-bg-secondary border border-border;
  border-radius: var(--ui-radius-surface);
  transition:
    border-color var(--ui-transition-fast),
    background-color var(--ui-transition-fast);
}

.setting-item-action {
  @apply shrink-0;
}

.setting-item-column {
  @apply flex-col items-stretch;
}

.setting-item-action-column {
  @apply w-full;
}
</style>

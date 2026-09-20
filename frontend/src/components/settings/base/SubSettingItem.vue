<script setup lang="ts">
import type { Component } from 'vue';

interface Props {
  icon?: Component;
  title: string;
  description?: string;
  required?: boolean;
  customClass?: string;
  layout?: 'row' | 'column';
}

withDefaults(defineProps<Props>(), {
  icon: undefined,
  description: '',
  required: false,
  customClass: '',
  layout: 'row',
});
</script>

<template>
  <div
    class="sub-setting-item"
    :class="[customClass, { 'sub-setting-item-column': layout === 'column' }]"
  >
    <div
      class="flex items-center sm:items-start justify-between gap-2 sm:gap-4 w-full"
      :class="{ 'flex-col': layout === 'column' }"
    >
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
          <div class="font-medium mb-0 sm:mb-1 text-xs sm:text-sm">
            {{ title }} <span v-if="required" class="text-red-500">*</span>
          </div>
          <div
            v-if="description"
            class="text-[10px] sm:text-xs text-text-secondary hidden sm:block"
          >
            {{ description }}
          </div>
          <slot name="extraInfo" />
        </div>
      </div>
      <div :class="layout === 'column' ? 'w-full' : 'shrink-0'">
        <slot />
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../style.css";
.sub-setting-item {
  @apply flex items-center sm:items-start gap-2 sm:gap-4 p-2 sm:p-2.5 bg-bg-tertiary;
  border-radius: var(--ui-radius-surface);
}

.sub-setting-item-column {
  @apply items-stretch;
}
</style>

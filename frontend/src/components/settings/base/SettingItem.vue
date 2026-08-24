<script setup lang="ts">
import type { Component } from 'vue';

interface Props {
  icon?: Component;
  title: string;
  description?: string;
  required?: boolean;
}

withDefaults(defineProps<Props>(), {
  icon: undefined,
  description: '',
  required: false,
});
</script>

<template>
  <div class="setting-item">
    <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
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
    <div class="setting-item-action">
      <slot name="action">
        <slot />
      </slot>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../style.css";
.setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border;
}

.setting-item-action {
  @apply shrink-0;
}
</style>

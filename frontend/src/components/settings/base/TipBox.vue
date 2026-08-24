<script setup lang="ts">
import { computed } from 'vue';
import type { Component } from 'vue';
import { PhInfo, PhQuestion, PhLightbulb } from '@phosphor-icons/vue';

interface Props {
  type?: 'info' | 'help' | 'tip';
  icon?: Component;
  title?: string;
}

const props = withDefaults(defineProps<Props>(), {
  type: 'tip',
  icon: undefined,
  title: undefined,
});

const defaultIcon = computed(() => {
  switch (props.type) {
    case 'info':
      return PhInfo;
    case 'help':
      return PhQuestion;
    default:
      return PhLightbulb;
  }
});

const displayIcon = computed(() => props.icon || defaultIcon.value);

const boxClass = computed(() => {
  switch (props.type) {
    case 'info':
      return 'tip-box-info';
    case 'help':
      return 'tip-box-help';
    default:
      return 'tip-box-tip';
  }
});
</script>

<template>
  <div class="tip-box" :class="boxClass">
    <div v-if="title || displayIcon" class="tip-box-header">
      <component :is="displayIcon" v-if="displayIcon" :size="18" class="tip-box-icon" />
      <span v-if="title" class="tip-box-title">{{ title }}</span>
    </div>
    <div v-if="$slots.default" class="tip-box-content">
      <slot />
    </div>
  </div>
</template>

<style scoped>
@reference "../../../style.css";
.tip-box {
  @apply py-2.5 px-3 rounded-lg border;
}

.tip-box-header {
  @apply flex items-start gap-2;
}

/* Add margin-bottom only when there's content after */
.tip-box-header:not(:last-child) {
  @apply mb-2;
}

.tip-box-icon {
  @apply shrink-0 mt-px;
}

.tip-box-title {
  @apply text-sm font-medium;
}

.tip-box-content {
  @apply text-xs sm:text-sm text-text-secondary ps-6;
}

.tip-box-content :deep(p) {
  @apply mb-2 last:mb-0;
}

.tip-box-content :deep(ol),
.tip-box-content :deep(ul) {
  @apply list-inside space-y-1 my-2;
}

.tip-box-content :deep(ol) {
  @apply list-decimal;
}

.tip-box-content :deep(ul) {
  @apply list-disc;
}

.tip-box-content :deep(a) {
  @apply text-accent underline underline-offset-2 hover:no-underline;
}

/* Info type - blue */
.tip-box-info {
  background-color: rgba(59, 130, 246, 0.05);
  border-color: rgba(59, 130, 246, 0.3);
}

.tip-box-info .tip-box-icon,
.tip-box-info .tip-box-title {
  @apply text-blue-500;
}

/* Help type - purple */
.tip-box-help {
  background-color: rgba(139, 92, 246, 0.05);
  border-color: rgba(139, 92, 246, 0.3);
}

.tip-box-help .tip-box-icon,
.tip-box-help .tip-box-title {
  @apply text-purple-500;
}

/* Tip type - green (default) */
.tip-box-tip {
  background-color: rgba(34, 197, 94, 0.05);
  border-color: rgba(34, 197, 94, 0.4);
}

.tip-box-tip .tip-box-icon,
.tip-box-tip .tip-box-title {
  @apply text-green-600;
}
</style>

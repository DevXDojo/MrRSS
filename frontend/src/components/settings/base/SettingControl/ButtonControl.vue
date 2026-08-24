<script setup lang="ts">
import type { Component } from 'vue';

interface Props {
  label?: string;
  icon?: Component;
  type?: 'primary' | 'secondary' | 'danger';
  disabled?: boolean;
  loading?: boolean;
  onClick?: () => void | Promise<void>;
}

defineProps<Props>();

const emit = defineEmits<{
  click: [];
}>();

async function handleClick() {
  emit('click');
}
</script>

<template>
  <button class="btn" :class="[`btn-${type}`]" :disabled="disabled || loading" @click="handleClick">
    <component
      :is="icon"
      v-if="icon"
      :size="16"
      class="sm:w-5 sm:h-5"
      :class="{ 'animate-spin': loading }"
    />
    <span v-if="label">{{ label }}</span>
    <span v-if="loading">{{ '...' }}</span>
  </button>
</template>

<style scoped>
@reference "../../../../style.css";
.btn {
  @apply px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium transition-colors shrink-0;
}

.btn:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.btn-primary {
  @apply bg-accent text-white border-none hover:bg-accent-hover;
}

.btn-secondary {
  @apply bg-bg-tertiary border border-border text-text-primary hover:bg-bg-secondary;
}

.btn-danger {
  @apply bg-red-50 dark:bg-red-900/20 border border-red-300 dark:border-red-400 text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/30;
}

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

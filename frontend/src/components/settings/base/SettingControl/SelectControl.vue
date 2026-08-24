<script setup lang="ts">
import BaseSelect from '@/components/common/BaseSelect.vue';

interface Option {
  value: string | number;
  label: string;
  disabled?: boolean;
}

interface Props {
  modelValue: string | number;
  options: Array<Option>;
  disabled?: boolean;
  width?: string;
}

defineProps<Props>();

const emit = defineEmits<{
  'update:modelValue': [value: string | number];
}>();

const widthClass = (width?: string) => {
  switch (width) {
    case 'sm':
      return 'w-20 sm:w-24';
    case 'md':
      return 'w-32 sm:w-48';
    case 'lg':
      return 'w-48 sm:w-64';
    default:
      return width || 'w-24 sm:w-48';
  }
};
</script>

<template>
  <BaseSelect
    :model-value="modelValue"
    :options="options"
    :disabled="disabled"
    :width="widthClass(width)"
    :class="{ 'opacity-50 cursor-not-allowed': disabled }"
    bg-mode="secondary"
    @update:model-value="emit('update:modelValue', $event)"
  />
</template>

<style scoped>
/* Styles are now handled by BaseSelect and select.css */
</style>

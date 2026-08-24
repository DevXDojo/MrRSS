<script setup lang="ts">
interface Props {
  modelValue: boolean;
  disabled?: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
}>();

function handleChange(event: Event) {
  const target = event.target as HTMLInputElement;
  emit('update:modelValue', target.checked);
}
</script>

<template>
  <input
    type="checkbox"
    class="toggle"
    :checked="modelValue"
    :disabled="disabled"
    @change="handleChange"
  />
</template>

<style scoped>
@reference "../../../../style.css";
.toggle {
  @apply w-10 h-5 appearance-none bg-bg-tertiary rounded-full relative cursor-pointer border border-border transition-colors checked:bg-accent checked:border-accent shrink-0;
}

.toggle:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.toggle::after {
  content: '';
  @apply absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full shadow-sm transition-transform;
}

.toggle:checked::after {
  transform: translateX(20px);
}
</style>

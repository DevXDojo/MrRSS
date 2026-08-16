<script setup lang="ts">
interface Props {
  modelValue: string;
  placeholder?: string;
  disabled?: boolean;
  error?: boolean;
  rows?: number;
  resize?: boolean;
  fontMono?: boolean;
}

withDefaults(defineProps<Props>(), {
  placeholder: '',
  rows: 3,
  resize: false,
  fontMono: false,
});

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

function handleInput(event: Event) {
  const target = event.target as HTMLTextAreaElement;
  emit('update:modelValue', target.value);
}
</script>

<template>
  <textarea
    class="input-field textarea"
    :class="[
      { 'border-red-500': error, 'opacity-50 cursor-not-allowed': disabled },
      { 'resize-none': !resize },
      { 'font-mono': fontMono },
    ]"
    :value="modelValue"
    :placeholder="placeholder"
    :disabled="disabled"
    :rows="rows"
    @input="handleInput"
  />
</template>

<style scoped>
@reference "../../../../style.css";
.input-field {
  @apply p-1.5 sm:p-2.5 border border-border rounded-md bg-bg-secondary text-text-primary focus:border-accent focus:outline-none transition-colors text-xs sm:text-sm;
}

.input-field:disabled {
  @apply cursor-not-allowed;
}

.textarea {
  @apply w-full;
}

.textarea::placeholder {
  @apply text-text-secondary;
}
</style>

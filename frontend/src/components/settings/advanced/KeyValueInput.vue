<script setup lang="ts">
import { ref } from 'vue';

interface Props {
  modelValue: string;
  placeholder?: string;
  asciiOnly?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  placeholder: '',
  asciiOnly: false,
});

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

// Local ref to prevent parent re-render
const localValue = ref(props.modelValue);

function filterNonAscii(str: string): string {
  // eslint-disable-next-line no-control-regex
  return str.replace(/[^\x00-\x7F]/g, '');
}

function onInput(event: Event) {
  const target = event.target as HTMLInputElement;
  let value = target.value;

  if (props.asciiOnly) {
    const filtered = filterNonAscii(value);
    if (filtered !== value) {
      // Directly update the input element value
      value = filtered;
      target.value = filtered;

      // Maintain cursor position
      const cursorPosition = target.selectionStart;
      if (cursorPosition !== null) {
        target.setSelectionRange(cursorPosition, cursorPosition);
      }
    }
  }

  localValue.value = value;
}

function onChange(event: Event) {
  const target = event.target as HTMLInputElement;
  localValue.value = target.value;
  emit('update:modelValue', target.value);
}
</script>

<template>
  <input
    :value="localValue"
    type="text"
    :placeholder="placeholder"
    class="input-field"
    @input="onInput"
    @change="onChange"
  />
</template>

<style scoped>
@reference "../../../style.css";
.input-field {
  @apply p-1.5 sm:p-2.5 border border-border rounded-md bg-bg-secondary text-text-primary focus:border-accent focus:outline-none transition-colors;
  @apply text-xs sm:text-sm flex-1;
}
</style>

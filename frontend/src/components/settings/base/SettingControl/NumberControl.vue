<script setup lang="ts">
import { ref, watch } from 'vue';

interface Props {
  modelValue: number;
  min?: number;
  max?: number;
  step?: number;
  suffix?: string;
  placeholder?: string;
  disabled?: boolean;
  error?: boolean;
  width?: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:modelValue': [value: number];
}>();

// Local state for the input
const inputValue = ref(props.modelValue);

// Watch for prop changes
watch(
  () => props.modelValue,
  (newValue) => {
    inputValue.value = newValue;
  },
  { immediate: true }
);

function handleInput(event: Event) {
  const target = event.target as HTMLInputElement;
  const value = parseFloat(target.value) || 0;
  inputValue.value = value;
  emit('update:modelValue', value);
}

const widthClass = (width?: string) => {
  switch (width) {
    case 'sm':
      return 'w-14 sm:w-16';
    case 'md':
      return 'w-20 sm:w-24';
    case 'lg':
      return 'w-28 sm:w-32';
    default:
      return width || 'w-20 sm:w-24';
  }
};
</script>

<template>
  <div class="flex items-center gap-1 sm:gap-2 shrink-0">
    <input
      class="input-field number-input"
      :class="[
        widthClass(width),
        { 'border-red-500': error, 'opacity-50 cursor-not-allowed': disabled },
      ]"
      type="number"
      :value="inputValue"
      :min="min"
      :max="max"
      :step="step"
      :placeholder="placeholder"
      :disabled="disabled"
      @input="
        (e) => {
          inputValue = parseFloat((e.target as HTMLInputElement).value) || 0;
        }
      "
      @change="handleInput"
    />
    <span v-if="suffix" class="text-xs sm:text-sm text-text-secondary">
      {{ suffix }}
    </span>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";
.input-field {
  @apply p-1.5 sm:p-2.5 border border-border rounded-md bg-bg-secondary text-text-primary focus:border-accent focus:outline-none transition-colors text-xs sm:text-sm;
}

.input-field:disabled {
  @apply cursor-not-allowed;
}

.number-input {
  text-align: center;
}

/* Remove number input spinners */
.number-input::-webkit-outer-spin-button,
.number-input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.number-input[type='number'] {
  -moz-appearance: textfield;
}
</style>

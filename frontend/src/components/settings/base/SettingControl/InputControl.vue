<script setup lang="ts">
interface Props {
  modelValue: string;
  type?: 'text' | 'password' | 'email' | 'url';
  placeholder?: string;
  disabled?: boolean;
  error?: boolean | string;
  required?: boolean;
  width?: string;
}

defineProps<Props>();

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

function handleInput(event: Event) {
  const target = event.target as HTMLInputElement;
  emit('update:modelValue', target.value);
}

const widthClass = (width?: string) => {
  switch (width) {
    case 'sm':
      return 'w-20 sm:w-24';
    case 'md':
      return 'w-36 sm:w-48';
    case 'lg':
      return 'w-48 sm:w-64';
    default:
      return width || 'w-32 sm:w-48';
  }
};
</script>

<template>
  <input
    class="input-field"
    :class="[
      widthClass(width),
      {
        'border-red-500': typeof error === 'boolean' ? error : error,
        'opacity-50 cursor-not-allowed': disabled,
      },
    ]"
    :type="type"
    :value="modelValue"
    :placeholder="placeholder"
    :disabled="disabled"
    :required="required"
    @input="handleInput"
  />
  <div v-if="typeof error === 'string' && error" class="text-red-500 text-xs mt-1">
    {{ error }}
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

.input-field::placeholder {
  @apply text-text-secondary;
}
</style>

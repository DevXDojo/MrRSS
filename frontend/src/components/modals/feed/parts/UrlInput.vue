<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

interface Props {
  modelValue: string;
  mode: 'add' | 'edit';
  isInvalid?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  isInvalid: false,
});

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

const { t } = useI18n();

// Local input value for handling
const localValue = ref(props.modelValue);

// Sync with props.modelValue
watch(
  () => props.modelValue,
  (newValue) => {
    localValue.value = newValue;
  }
);

// Dynamic placeholder
const inputPlaceholder = computed(() => {
  return t('setting.rsshub.urlPlaceholder');
});

// Handle input event - just update local value
function handleInput(event: Event) {
  const target = event.target as HTMLInputElement;
  localValue.value = target.value;
  emit('update:modelValue', target.value);
}
</script>

<template>
  <div class="mb-3 sm:mb-4">
    <label class="block mb-1 sm:mb-1.5 font-semibold text-xs sm:text-sm text-text-secondary"
      >{{ t('modal.feed.rssUrl') }}
      <span v-if="props.mode === 'add'" class="text-red-500">*</span></label
    >
    <input
      v-model="localValue"
      type="text"
      :placeholder="inputPlaceholder"
      :class="['input-field', props.mode === 'add' && props.isInvalid ? 'border-red-500' : '']"
      @input="handleInput"
    />
  </div>
</template>

<style scoped>
@reference "../../../../style.css";
.input-field {
  @apply w-full p-2 sm:p-2.5 border border-border rounded-md bg-bg-tertiary text-text-primary text-xs sm:text-sm focus:border-accent focus:outline-none transition-colors;
}
</style>

<script setup lang="ts">
import { ref, onMounted, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import BaseModal from '@/components/common/BaseModal.vue';
import ModalFooter from '@/components/common/ModalFooter.vue';

const { t } = useI18n();

interface Props {
  title?: string;
  message?: string;
  placeholder?: string;
  defaultValue?: string;
  confirmText?: string;
  cancelText?: string;
  suggestions?: string[];
}

const props = withDefaults(defineProps<Props>(), {
  title: 'Input',
  message: '',
  placeholder: '',
  defaultValue: '',
  confirmText: undefined,
  cancelText: undefined,
  suggestions: () => [],
});

// Use i18n translations if not provided
const getConfirmText = (customText?: string) => customText || t('common.confirm');
const getCancelText = (customText?: string) => customText || t('common.cancel');

const emit = defineEmits<{
  confirm: [value: string];
  cancel: [];
  close: [];
}>();

const inputValue = ref(props.defaultValue);
const inputRef: Ref<HTMLInputElement | null> = ref(null);

onMounted(() => {
  // Focus the input when dialog opens
  if (inputRef.value) {
    inputRef.value.focus();
    inputRef.value.select();
  }
});

function handleConfirm() {
  emit('confirm', inputValue.value);
  emit('close');
}

function handleCancel() {
  emit('cancel');
  emit('close');
}

function handleClose() {
  emit('cancel');
  emit('close');
}

function handleKeyDown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    e.preventDefault();
    handleConfirm();
  }
}
</script>

<template>
  <BaseModal :title="title" size="md" :closable="false" @close="handleClose">
    <!-- Body -->
    <div class="p-3 sm:p-5">
      <p v-if="message" class="m-0 mb-2 sm:mb-3 text-text-primary text-sm sm:text-base">
        {{ message }}
      </p>
      <input
        ref="inputRef"
        v-model="inputValue"
        type="text"
        :placeholder="placeholder"
        :list="suggestions.length > 0 ? 'input-suggestions' : undefined"
        class="input-field w-full text-sm sm:text-base"
        @keydown="handleKeyDown"
      />
      <datalist v-if="suggestions.length > 0" id="input-suggestions">
        <option v-for="suggestion in suggestions" :key="suggestion" :value="suggestion" />
      </datalist>
    </div>

    <!-- Footer -->
    <template #footer>
      <ModalFooter
        :secondary-button="{
          label: getCancelText(cancelText),
          onClick: handleCancel,
        }"
        :primary-button="{
          label: getConfirmText(confirmText),
          onClick: handleConfirm,
        }"
      />
    </template>
  </BaseModal>
</template>

<style scoped>
@reference "../../../style.css";

.input-field {
  @apply px-3 py-2 rounded-lg border border-border bg-bg-secondary text-text-primary;
  @apply focus:outline-none focus:ring-2 focus:ring-accent;
}
</style>

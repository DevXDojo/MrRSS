<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import BaseModal from '@/components/common/BaseModal.vue';
import ModalFooter from '@/components/common/ModalFooter.vue';

interface Props {
  title?: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  isDanger?: boolean;
}

withDefaults(defineProps<Props>(), {
  title: 'Confirm',
  confirmText: undefined,
  cancelText: undefined,
  isDanger: false,
});

const { t } = useI18n();

const emit = defineEmits<{
  confirm: [];
  cancel: [];
  close: [];
}>();

// Use i18n translations if not provided
const getConfirmText = (customText?: string) => customText || t('common.confirm');
const getCancelText = (customText?: string) => customText || t('common.cancel');

function handleConfirm() {
  emit('confirm');
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
</script>

<template>
  <BaseModal :title="title" :closable="false" size="md" @close="handleClose">
    <!-- Body -->
    <div class="p-3 sm:p-5">
      <p class="m-0 text-text-primary text-sm sm:text-base">{{ message }}</p>
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
          type: isDanger ? 'danger' : 'primary',
          onClick: handleConfirm,
        }"
      />
    </template>
  </BaseModal>
</template>

<style scoped>
@reference "../../../style.css";
</style>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhCheckSquare, PhSquare } from '@phosphor-icons/vue';
import BaseModal from '@/components/common/BaseModal.vue';
import ModalFooter from '@/components/common/ModalFooter.vue';
import type { MultiSelectOption } from '@/types/global';

const { t } = useI18n();

interface Props {
  title?: string;
  message?: string;
  options?: MultiSelectOption[];
  confirmText?: string;
  cancelText?: string;
  searchable?: boolean;
}

// Use i18n translations if not provided
const getConfirmText = (customText?: string) => customText || t('common.confirm');
const getCancelText = (customText?: string) => customText || t('common.cancel');

const emit = defineEmits<{
  confirm: [values: string[]];
  cancel: [];
  close: [];
}>();

const props = withDefaults(defineProps<Props>(), {
  title: 'Select',
  message: '',
  options: () => [],
  confirmText: undefined,
  cancelText: undefined,
  searchable: false,
});

const selectedValues = ref<string[]>([]);
const searchQuery = ref('');

// Filter options based on search query
const filteredOptions = computed(() => {
  if (!props.searchable || !searchQuery.value.trim()) {
    return props.options;
  }

  const query = searchQuery.value.toLowerCase().trim();
  return props.options.filter((option: MultiSelectOption) =>
    option.label.toLowerCase().includes(query)
  );
});

function handleConfirm() {
  emit('confirm', selectedValues.value);
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
  } else if (e.key === 'Escape') {
    e.preventDefault();
    handleCancel();
  }
}

onMounted(() => {
  // Initialize with no selection
  selectedValues.value = [];
  window.addEventListener('keydown', handleKeyDown);
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown);
});

function toggleOption(value: string) {
  const index = selectedValues.value.indexOf(value);
  if (index > -1) {
    selectedValues.value.splice(index, 1);
  } else {
    selectedValues.value.push(value);
  }
}

function isOptionSelected(value: string): boolean {
  return selectedValues.value.includes(value);
}
</script>

<template>
  <BaseModal :title="title" size="md" :closable="false" :z-index="150" @close="handleClose">
    <!-- Body -->
    <div class="p-3 sm:p-5">
      <p v-if="message" class="m-0 mb-3 text-text-primary text-sm sm:text-base">
        {{ message }}
      </p>

      <!-- Search Input -->
      <div v-if="searchable" class="mb-2">
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="t('common.select.searchPlaceholder')"
          class="w-full px-2 py-1 bg-bg-secondary text-text-primary text-xs focus:outline-none"
          @keydown.enter.stop
        />
      </div>

      <!-- Options List -->
      <div class="max-h-64 overflow-y-auto border border-border rounded-lg bg-bg-secondary">
        <div
          v-for="option in filteredOptions"
          :key="option.value"
          class="flex items-center gap-2 px-3 py-2 hover:bg-bg-tertiary cursor-pointer transition-colors border-b border-border last:border-0"
          @click="toggleOption(option.value)"
        >
          <!-- Checkbox Icon -->
          <component
            :is="isOptionSelected(option.value) ? PhCheckSquare : PhSquare"
            :size="20"
            :class="[
              'shrink-0',
              isOptionSelected(option.value) ? 'text-accent' : 'text-text-secondary',
            ]"
          />

          <!-- Option Label -->
          <span class="flex-1 text-sm sm:text-base text-text-primary">
            {{ option.label }}
          </span>

          <!-- Color Indicator -->
          <span
            v-if="option.color"
            class="w-4 h-4 rounded-full border border-border"
            :style="{ backgroundColor: option.color }"
          />
        </div>

        <!-- Empty State -->
        <div
          v-if="filteredOptions.length === 0"
          class="px-3 py-8 text-center text-text-secondary text-sm"
        >
          {{ searchable && searchQuery ? t('common.select.noResults') : t('modal.tag.noTags') }}
        </div>
      </div>

      <!-- Selection Count -->
      <div v-if="selectedValues.length > 0" class="mt-2 text-xs text-text-secondary">
        {{
          t('common.search.totalAndSelected', {
            total: filteredOptions.length,
            selected: selectedValues.length,
          })
        }}
      </div>
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

<style scoped></style>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import BaseSelect from '@/components/common/BaseSelect.vue';
import type { SelectOption } from '@/types/select';

interface Props {
  category: string;
  categorySelection: string;
  showCustomCategory: boolean;
  existingCategories: string[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:category': [value: string];
  'update:categorySelection': [value: string];
  'update:showCustomCategory': [value: boolean];
  'handle-category-change': [value: string];
}>();

const { t } = useI18n();

function handleCategoryChange(value: string) {
  emit('handle-category-change', value);
}

// Build options for BaseSelect
const categoryOptions = computed<SelectOption[]>(() => {
  const options: SelectOption[] = [{ value: '', label: t('sidebar.feedList.uncategorized') }];

  // Add existing categories
  props.existingCategories.forEach((cat) => {
    options.push({ value: cat, label: cat });
  });

  // If current category is not in the list and not empty, add it
  // This handles newly added categories that haven't been saved yet
  if (
    props.categorySelection &&
    props.categorySelection !== '' &&
    !props.existingCategories.includes(props.categorySelection)
  ) {
    options.push({ value: props.categorySelection, label: props.categorySelection });
  }

  return options;
});

// Handle select value change
function handleSelectChange(value: string | number) {
  handleCategoryChange(String(value));
}

// Handle add new category from dropdown
function handleAddCategory(categoryName: string) {
  // Update the category values
  emit('update:category', categoryName);
  emit('update:categorySelection', categoryName);
  // Also trigger handleCategoryChange to ensure proper state management
  emit('handle-category-change', categoryName);
}
</script>

<template>
  <div class="mb-3 sm:mb-4">
    <label class="block mb-1 sm:mb-1.5 font-semibold text-xs sm:text-sm text-text-secondary">{{
      t('common.form.category')
    }}</label>

    <!-- Using BaseSelect with add new option support -->
    <BaseSelect
      v-if="!props.showCustomCategory"
      :model-value="props.categorySelection"
      :options="categoryOptions"
      allow-add
      :add-placeholder="t('modal.feed.enterCategoryName')"
      :position="'auto'"
      @update:model-value="handleSelectChange"
      @add="handleAddCategory"
    />

    <!-- Custom category input (legacy mode for backward compatibility) -->
    <div v-else class="flex gap-2">
      <input
        :value="props.category"
        type="text"
        :placeholder="t('modal.feed.enterCategoryName')"
        class="input-field flex-1"
        autofocus
        @input="emit('update:category', ($event.target as HTMLInputElement).value)"
      />
      <button
        type="button"
        class="px-3 py-2 text-xs sm:text-sm text-text-secondary hover:text-text-primary border border-border rounded-md hover:bg-bg-tertiary transition-colors"
        @click="
          emit('update:showCustomCategory', false);
          emit('update:categorySelection', '');
        "
      >
        {{ t('common.cancel') }}
      </button>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";
.input-field {
  @apply w-full p-2 sm:p-2.5 border border-border rounded-md bg-bg-tertiary text-text-primary text-xs sm:text-sm focus:border-accent focus:outline-none transition-colors;
  box-sizing: border-box;
}
</style>

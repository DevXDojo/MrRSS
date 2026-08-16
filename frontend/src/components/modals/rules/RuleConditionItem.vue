<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhProhibit, PhTrash } from '@phosphor-icons/vue';
import BaseSelect from '@/components/common/BaseSelect.vue';
import BaseMultiSelect from '@/components/common/BaseMultiSelect.vue';
import {
  useRuleOptions,
  type Condition,
  isDateField,
  isBooleanField,
  isNumberField,
  needsOperator,
} from '@/composables/rules/useRuleOptions';
import type { SelectOption } from '@/types/select';

const { t } = useI18n();

// Helper function to translate feed type code to display text
function getFeedTypeLabel(typeCode: string): string {
  const mapping: Record<string, string> = {
    regular: t('modal.feed.typeRegular'),
    freshrss: t('modal.feed.typeFreshRSS'),
    rsshub: t('modal.feed.typeRSSHub'),
    script: t('modal.feed.typeCustomScript'),
    xpath: t('modal.feed.typeXPath'),
    email: t('modal.feed.typeEmail'),
  };
  return mapping[typeCode] || typeCode;
}

const {
  fieldOptions,
  textOperatorOptions,
  booleanOptions,
  feedNames,
  feedCategories,
  feedTypes,
  feedTags,
} = useRuleOptions();

interface Props {
  condition: Condition;
  index: number;
}

defineProps<Props>();

const emit = defineEmits<{
  'update:field': [value: string];
  'update:operator': [value: string];
  'update:value': [value: string];
  'update:values': [values: string[]];
  'update:negate': [];
  remove: [];
}>();

// Build options for BaseSelect
const fieldSelectOptions = computed<SelectOption[]>(() => {
  return fieldOptions.map((opt) => ({
    value: opt.value,
    label: t(opt.labelKey),
  }));
});

const operatorSelectOptions = computed<SelectOption[]>(() => {
  return textOperatorOptions.map((opt) => ({
    value: opt.value,
    label: t(opt.labelKey),
  }));
});

const booleanSelectOptions = computed<SelectOption[]>(() => {
  return booleanOptions.map((opt) => ({
    value: opt.value,
    label: t(opt.labelKey),
  }));
});

const updateStatusOptions = computed<SelectOption[]>(() => {
  return [
    { value: '', label: t('modal.filter.updateSuccess') },
    { value: 'success', label: t('modal.filter.updateSuccess') },
    { value: 'failed', label: t('modal.filter.updateFailed') },
  ];
});

function handleValueChange(event: Event): void {
  const target = event.target as HTMLInputElement;
  emit('update:value', target.value);
}

function handleMultiSelectUpdate(values: (string | number)[]): void {
  emit('update:values', values.map(String));
}
</script>

<template>
  <div class="condition-row bg-bg-secondary border border-border rounded-lg p-2 sm:p-3">
    <div class="flex flex-wrap gap-2 items-end">
      <!-- NOT toggle button -->
      <div class="flex-shrink-0">
        <label class="block text-[10px] sm:text-xs text-text-secondary mb-1">&nbsp;</label>
        <button
          :class="['not-btn', condition.negate ? 'active' : '']"
          :title="t('modal.filter.not')"
          @click="emit('update:negate')"
        >
          <PhProhibit :size="14" class="sm:w-4 sm:h-4" />
          <span class="text-[10px] sm:text-xs font-medium">{{ t('modal.filter.not') }}</span>
        </button>
      </div>

      <!-- Field selector -->
      <div class="flex-1 min-w-[100px] sm:min-w-[130px]">
        <label class="block text-[10px] sm:text-xs text-text-secondary mb-1">{{
          t('modal.filter.filterField')
        }}</label>
        <BaseSelect
          :model-value="condition.field"
          :options="fieldSelectOptions"
          @update:model-value="emit('update:field', String($event))"
        />
      </div>

      <!-- Operator selector (only for article_title) -->
      <div v-if="needsOperator(condition.field)" class="w-24 sm:w-28">
        <label class="block text-[10px] sm:text-xs text-text-secondary mb-1">{{
          t('modal.filter.filterOperator')
        }}</label>
        <BaseSelect
          :model-value="condition.operator"
          :options="operatorSelectOptions"
          @update:model-value="emit('update:operator', String($event))"
        />
      </div>

      <!-- Value input -->
      <div class="flex-1 min-w-[100px] sm:min-w-[140px]">
        <label class="block text-[10px] sm:text-xs text-text-secondary mb-1">{{
          t('modal.filter.filterValue')
        }}</label>

        <!-- Date input -->
        <input
          v-if="isDateField(condition.field)"
          type="date"
          :value="condition.value"
          class="date-field w-full text-xs sm:text-sm"
          @input="handleValueChange"
        />

        <!-- Boolean select -->
        <BaseSelect
          v-else-if="isBooleanField(condition.field)"
          :model-value="condition.value"
          :options="booleanSelectOptions"
          @update:model-value="emit('update:value', String($event))"
        />

        <!-- Number input -->
        <input
          v-else-if="isNumberField(condition.field)"
          type="number"
          :value="condition.value"
          class="input-field w-full text-xs sm:text-sm"
          :placeholder="t('modal.filter.filterValue')"
          @input="handleValueChange"
        />

        <!-- Special dropdown for feed_last_update_status -->
        <BaseSelect
          v-else-if="condition.field === 'feed_last_update_status'"
          :model-value="condition.value"
          :options="updateStatusOptions"
          @update:model-value="emit('update:value', String($event))"
        />

        <!-- Multi-select dropdown for feed name -->
        <BaseMultiSelect
          v-else-if="condition.field === 'feed_name'"
          :model-value="condition.values || []"
          :options="feedNames.map((name) => ({ value: name, label: name }))"
          :placeholder="t('common.search.selectItems')"
          :searchable="true"
          @update:model-value="handleMultiSelectUpdate"
        />

        <!-- Multi-select dropdown for category -->
        <BaseMultiSelect
          v-else-if="condition.field === 'feed_category'"
          :model-value="condition.values || []"
          :options="feedCategories.map((cat) => ({ value: cat, label: cat }))"
          :placeholder="t('common.search.selectItems')"
          :searchable="true"
          @update:model-value="handleMultiSelectUpdate"
        />

        <!-- Multi-select dropdown for feed type -->
        <BaseMultiSelect
          v-else-if="condition.field === 'feed_type'"
          :model-value="condition.values || []"
          :options="feedTypes.map((type) => ({ value: type, label: getFeedTypeLabel(type) }))"
          :placeholder="t('common.search.selectItems')"
          @update:model-value="handleMultiSelectUpdate"
        />

        <!-- Multi-select dropdown for feed tags -->
        <BaseMultiSelect
          v-else-if="condition.field === 'feed_tags'"
          :model-value="condition.values || []"
          :options="feedTags.map((tag) => ({ value: tag, label: tag }))"
          :placeholder="t('common.search.selectItems')"
          :searchable="true"
          @update:model-value="handleMultiSelectUpdate"
        />

        <!-- Regular text input -->
        <input
          v-else
          type="text"
          :value="condition.value"
          class="input-field w-full text-xs sm:text-sm"
          :placeholder="t('modal.filter.filterValue')"
          @input="handleValueChange"
        />
      </div>

      <!-- Remove button -->
      <div class="flex-shrink-0">
        <label class="block text-[10px] sm:text-xs text-text-secondary mb-1">&nbsp;</label>
        <button
          class="btn-danger-icon"
          :title="t('setting.rule.removeCondition')"
          @click="emit('remove')"
        >
          <PhTrash :size="16" class="sm:w-[18px] sm:h-[18px]" />
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../style.css";
.input-field {
  @apply p-1.5 sm:p-2 border border-border rounded-md bg-bg-primary text-text-primary focus:border-accent focus:outline-none transition-colors;
  height: 38px;
}

.date-field {
  @apply p-1.5 sm:p-2 border border-border rounded-md bg-bg-primary text-text-primary focus:border-accent focus:outline-none transition-colors cursor-pointer;
  color-scheme: light dark;
  height: 38px;
}

.btn-danger-icon {
  @apply p-1.5 sm:p-2 rounded-lg text-red-500 hover:bg-red-500/10 transition-colors cursor-pointer;
  height: 38px;
  width: 38px;
}

/* NOT button styling */
.not-btn {
  @apply flex items-center gap-1 px-1.5 sm:px-2 rounded-md border transition-all cursor-pointer;
  @apply text-text-secondary bg-bg-primary border-border;
  height: 38px;
}
.not-btn:hover {
  @apply border-red-400 text-red-500;
}
.not-btn.active {
  @apply bg-red-500/10 border-red-500 text-red-500;
}
</style>

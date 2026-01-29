<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhProhibit, PhTrash } from '@phosphor-icons/vue';
import {
  useRuleOptions,
  type Condition,
  isDateField,
  isBooleanField,
  needsOperator,
} from '@/composables/rules/useRuleOptions';

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
  isDropdownOpen: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:field': [value: string];
  'update:operator': [value: string];
  'update:value': [value: string];
  'update:values': [values: string[]];
  'update:negate': [];
  'toggle-dropdown': [];
  remove: [];
}>();

function handleFieldChange(event: Event): void {
  const target = event.target as HTMLSelectElement;
  emit('update:field', target.value);
}

function handleOperatorChange(event: Event): void {
  const target = event.target as HTMLSelectElement;
  emit('update:operator', target.value);
}

function handleValueChange(event: Event): void {
  const target = event.target as HTMLInputElement;
  emit('update:value', target.value);
}

function handleToggleMultiSelectValue(value: string): void {
  const currentValues = props.condition.values || [];
  const newValues = currentValues.includes(value)
    ? currentValues.filter((v) => v !== value)
    : [...currentValues, value];
  emit('update:values', newValues);
}

function getMultiSelectDisplayText(): string {
  const values = props.condition.values || [];
  if (values.length === 0) return t('common.search.selectItems');

  // For feed_type, translate the type codes to display text
  if (props.condition.field === 'feed_type') {
    if (values.length === 1) return getFeedTypeLabel(values[0]);
    return t('common.search.itemsSelected', { count: values.length });
  }

  // For other fields, use the values directly
  if (values.length === 1) return values[0];
  return t('common.search.itemsSelected', { count: values.length });
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
        <select
          :value="condition.field"
          class="select-field w-full text-xs sm:text-sm"
          @change="handleFieldChange"
        >
          <option v-for="opt in fieldOptions" :key="opt.value" :value="opt.value">
            {{ t(opt.labelKey) }}
          </option>
        </select>
      </div>

      <!-- Operator selector (only for article_title) -->
      <div v-if="needsOperator(condition.field)" class="w-24 sm:w-28">
        <label class="block text-[10px] sm:text-xs text-text-secondary mb-1">{{
          t('modal.filter.filterOperator')
        }}</label>
        <select
          :value="condition.operator"
          class="select-field w-full text-xs sm:text-sm"
          @change="handleOperatorChange"
        >
          <option v-for="opt in textOperatorOptions" :key="opt.value" :value="opt.value">
            {{ t(opt.labelKey) }}
          </option>
        </select>
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
        <select
          v-else-if="isBooleanField(condition.field)"
          :value="condition.value"
          class="select-field w-full text-xs sm:text-sm"
          @change="handleValueChange"
        >
          <option v-for="opt in booleanOptions" :key="opt.value" :value="opt.value">
            {{ t(opt.labelKey) }}
          </option>
        </select>

        <!-- Multi-select dropdown for feed name -->
        <div v-else-if="condition.field === 'feed_name'" class="dropdown-container">
          <button
            type="button"
            class="dropdown-trigger text-xs sm:text-sm"
            @click="emit('toggle-dropdown')"
          >
            <span class="dropdown-text truncate">{{ getMultiSelectDisplayText() }}</span>
            <span class="dropdown-arrow">▼</span>
          </button>
          <div v-if="isDropdownOpen" class="dropdown-menu dropdown-down">
            <div
              v-for="name in feedNames"
              :key="name"
              :class="[
                'dropdown-option text-xs sm:text-sm',
                condition.values.includes(name) ? 'selected' : '',
              ]"
              @click.stop="handleToggleMultiSelectValue(name)"
            >
              <input
                type="checkbox"
                :checked="condition.values.includes(name)"
                class="checkbox-input"
                tabindex="-1"
              />
              <span class="truncate">{{ name }}</span>
            </div>
            <div v-if="feedNames.length === 0" class="text-text-secondary text-xs sm:text-sm p-2">
              {{ t('article.content.noArticles') }}
            </div>
          </div>
        </div>

        <!-- Multi-select dropdown for category -->
        <div v-else-if="condition.field === 'feed_category'" class="dropdown-container">
          <button
            type="button"
            class="dropdown-trigger text-xs sm:text-sm"
            @click="emit('toggle-dropdown')"
          >
            <span class="dropdown-text truncate">{{ getMultiSelectDisplayText() }}</span>
            <span class="dropdown-arrow">▼</span>
          </button>
          <div v-if="isDropdownOpen" class="dropdown-menu dropdown-down">
            <div
              v-for="cat in feedCategories"
              :key="cat"
              :class="[
                'dropdown-option text-xs sm:text-sm',
                condition.values.includes(cat as string) ? 'selected' : '',
              ]"
              @click.stop="handleToggleMultiSelectValue(cat as string)"
            >
              <input
                type="checkbox"
                :checked="condition.values.includes(cat as string)"
                class="checkbox-input"
                tabindex="-1"
              />
              <span class="truncate">{{ cat }}</span>
            </div>
            <div
              v-if="feedCategories.length === 0"
              class="text-text-secondary text-xs sm:text-sm p-2"
            >
              {{ t('article.content.noArticles') }}
            </div>
          </div>
        </div>

        <!-- Multi-select dropdown for feed type -->
        <div v-else-if="condition.field === 'feed_type'" class="dropdown-container">
          <button
            type="button"
            class="dropdown-trigger text-xs sm:text-sm"
            @click="emit('toggle-dropdown')"
          >
            <span class="dropdown-text truncate">{{ getMultiSelectDisplayText() }}</span>
            <span class="dropdown-arrow">▼</span>
          </button>
          <div v-if="isDropdownOpen" class="dropdown-menu dropdown-down">
            <div
              v-for="type in feedTypes"
              :key="type"
              :class="[
                'dropdown-option text-xs sm:text-sm',
                condition.values.includes(type as string) ? 'selected' : '',
              ]"
              @click.stop="handleToggleMultiSelectValue(type as string)"
            >
              <input
                type="checkbox"
                :checked="condition.values.includes(type as string)"
                class="checkbox-input"
                tabindex="-1"
              />
              <span class="truncate">{{ getFeedTypeLabel(type as string) }}</span>
            </div>
            <div v-if="feedTypes.length === 0" class="text-text-secondary text-xs sm:text-sm p-2">
              {{ t('article.content.noArticles') }}
            </div>
          </div>
        </div>

        <!-- Multi-select dropdown for feed tags -->
        <div v-else-if="condition.field === 'feed_tags'" class="dropdown-container">
          <button
            type="button"
            class="dropdown-trigger text-xs sm:text-sm"
            @click="emit('toggle-dropdown')"
          >
            <span class="dropdown-text truncate">{{ getMultiSelectDisplayText() }}</span>
            <span class="dropdown-arrow">▼</span>
          </button>
          <div v-if="isDropdownOpen" class="dropdown-menu dropdown-down">
            <div
              v-for="tag in feedTags"
              :key="tag"
              :class="[
                'dropdown-option text-xs sm:text-sm',
                condition.values.includes(tag as string) ? 'selected' : '',
              ]"
              @click.stop="handleToggleMultiSelectValue(tag as string)"
            >
              <input
                type="checkbox"
                :checked="condition.values.includes(tag as string)"
                class="checkbox-input"
                tabindex="-1"
              />
              <span class="truncate">{{ tag }}</span>
            </div>
            <div v-if="feedTags.length === 0" class="text-text-secondary text-xs sm:text-sm p-2">
              {{ t('article.content.noArticles') }}
            </div>
          </div>
        </div>

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
.select-field {
  @apply p-1.5 sm:p-2 border border-border rounded-md bg-bg-primary text-text-primary focus:border-accent focus:outline-none transition-colors cursor-pointer;
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

/* Dropdown multi-select styling */
.dropdown-container {
  @apply relative;
}
.dropdown-trigger {
  @apply w-full p-1.5 sm:p-2 border border-border rounded-md bg-bg-primary text-text-primary;
  @apply flex items-center justify-between cursor-pointer hover:border-accent transition-colors;
  height: 38px;
}
.dropdown-text {
  @apply flex-1 text-left;
}
.dropdown-arrow {
  @apply text-text-secondary text-[10px] sm:text-xs ml-2;
}
.dropdown-menu {
  @apply absolute left-0 right-0 border border-border rounded-md bg-bg-primary;
  @apply max-h-40 overflow-y-auto z-50 shadow-lg scroll-smooth;
}
.dropdown-menu.dropdown-down {
  @apply top-full mt-1;
}
.dropdown-option {
  @apply flex items-center gap-2 px-2 sm:px-3 py-1.5 sm:py-2 cursor-pointer text-text-primary hover:bg-bg-tertiary;
  min-height: 36px;
}
.dropdown-option.selected {
  background-color: rgba(59, 130, 246, 0.1);
}
.checkbox-input {
  @apply w-4 h-4 accent-accent cursor-pointer flex-shrink-0;
}
</style>

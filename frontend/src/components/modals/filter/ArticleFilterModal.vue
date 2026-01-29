<script setup lang="ts">
import { watch, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhPlus, PhFunnel } from '@phosphor-icons/vue';
import type { FilterCondition } from '@/types/filter';
import { useFilterFields } from '@/composables/filter/useFilterFields';
import { useFilterConditions } from '@/composables/filter/useFilterConditions';
import RuleConditionItem from '../rules/RuleConditionItem.vue';
import BaseModal from '@/components/common/BaseModal.vue';
import ModalFooter from '@/components/common/ModalFooter.vue';

const { t } = useI18n();

interface Props {
  show?: boolean;
  currentFilters?: FilterCondition[];
}

const props = withDefaults(defineProps<Props>(), {
  show: false,
  currentFilters: () => [],
});

const emit = defineEmits<{
  close: [];
  apply: [filters: FilterCondition[]];
}>();

// Use composables
const { logicOptions, onFieldChange: handleFieldChange } = useFilterFields();

const {
  conditions,
  openDropdownIndex,
  initializeConditions,
  addCondition,
  removeCondition,
  toggleNegate,
  toggleDropdown,
  clearConditions,
  getValidConditions,
} = useFilterConditions();

// Watch for modal show changes to reload filters
watch(
  () => props.show,
  (newVal) => {
    if (newVal && props.currentFilters && props.currentFilters.length > 0) {
      initializeConditions(props.currentFilters);
    }
  }
);

onMounted(() => {
  // Load existing filters if provided
  if (props.currentFilters && props.currentFilters.length > 0) {
    initializeConditions(props.currentFilters);
  }
});

function onFieldChange(index: number): void {
  handleFieldChange(conditions.value[index]);
}

function clearFilters(): void {
  clearConditions();
  // Auto-apply when clearing filters
  emit('apply', []);
  emit('close');
}

function applyFilters(): void {
  const validConditions = getValidConditions();
  emit('apply', validConditions);
  emit('close');
}

function close() {
  emit('close');
}
</script>

<template>
  <BaseModal v-if="show" size="2xl" :z-index="100" @close="close">
    <!-- Custom Header -->
    <template #header>
      <h3 class="text-lg font-semibold m-0 flex items-center gap-2 text-text-primary">
        <PhFunnel :size="20" />
        {{ t('modal.filter.filterArticles') }}
      </h3>
    </template>

    <!-- Content -->
    <div class="p-4 sm:p-6">
      <!-- Empty state -->
      <div v-if="conditions.length === 0" class="text-center text-text-secondary py-8">
        <PhFunnel :size="48" class="mx-auto mb-3 opacity-50" />
        <p>{{ t('modal.filter.noFiltersApplied') }}</p>
      </div>

      <!-- Condition list -->
      <div v-else class="space-y-3">
        <div v-for="(condition, index) in conditions" :key="condition.id">
          <!-- Logic connector (AND/OR) between conditions - styled distinctly -->
          <div v-if="index > 0" class="flex items-center justify-center my-3">
            <div class="flex-1 h-px bg-border"></div>
            <div class="logic-connector mx-3">
              <button
                v-for="opt in logicOptions"
                :key="opt.value"
                :class="['logic-btn', condition.logic === opt.value ? 'active' : '']"
                @click="(condition.logic as 'and' | 'or' | null) = opt.value"
              >
                {{ t(opt.labelKey) }}
              </button>
            </div>
            <div class="flex-1 h-px bg-border"></div>
          </div>

          <!-- Condition card -->
          <RuleConditionItem
            :condition="condition"
            :index="index"
            :is-dropdown-open="openDropdownIndex === index"
            @update:field="
              (value) => {
                condition.field = value;
                onFieldChange(index);
              }
            "
            @update:operator="(value) => (condition.operator = value)"
            @update:value="(value) => (condition.value = value)"
            @update:values="(values) => (condition.values = values)"
            @update:negate="toggleNegate(index)"
            @toggle-dropdown="toggleDropdown(index)"
            @remove="removeCondition(index)"
          />
        </div>
      </div>

      <!-- Add condition button -->
      <button
        class="btn-secondary w-full mt-4 flex items-center justify-center gap-2"
        @click="addCondition"
      >
        <PhPlus :size="18" />
        {{ t('modal.filter.addCondition') }}
      </button>
    </div>

    <!-- Footer -->
    <template #footer>
      <ModalFooter
        align="right"
        :secondary-button="{
          label: t('modal.filter.clearFilters'),
          disabled: conditions.length === 0,
          onClick: clearFilters,
        }"
        :primary-button="{
          label: t('modal.filter.applyFilters'),
          onClick: applyFilters,
        }"
      />
    </template>
  </BaseModal>
</template>

<style scoped>
@reference "../../../style.css";

.btn-secondary {
  @apply bg-bg-tertiary text-text-primary border border-border px-4 py-2.5 rounded-lg cursor-pointer font-medium hover:bg-bg-secondary transition-colors disabled:opacity-50 disabled:cursor-not-allowed;
}

/* Logic connector styling - distinct visual appearance */
.logic-connector {
  @apply flex items-center gap-1 bg-bg-tertiary rounded-full p-1;
}
.logic-btn {
  @apply px-3 py-1 text-xs font-bold rounded-full transition-all cursor-pointer;
  @apply text-text-secondary bg-transparent;
}
.logic-btn:hover {
  @apply text-text-primary bg-bg-secondary;
}
.logic-btn.active {
  @apply text-white bg-accent;
}
</style>

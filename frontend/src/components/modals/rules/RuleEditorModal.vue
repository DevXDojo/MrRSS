<script setup lang="ts">
import { ref, computed, watch, type Ref, type ComputedRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhLightning, PhPlus, PhFunnel, PhListChecks } from '@phosphor-icons/vue';
import RuleLogicConnector from './RuleLogicConnector.vue';
import RuleAction from './RuleAction.vue';
import RuleConditionItem from './RuleConditionItem.vue';
import {
  useRuleOptions,
  type Condition,
  isMultiSelectField,
} from '@/composables/rules/useRuleOptions';
import { useRuleConditions } from '@/composables/rules/useRuleConditions';
import { useRuleActions } from '@/composables/rules/useRuleActions';
import BaseModal from '@/components/common/BaseModal.vue';
import ModalFooter from '@/components/common/ModalFooter.vue';

const { t } = useI18n();

// Use composables
const { actionOptions } = useRuleOptions();

const {
  openDropdownIndex,
  addCondition: addConditionHelper,
  removeCondition: removeConditionHelper,
  onFieldChange,
  toggleNegate,
  toggleDropdown,
} = useRuleConditions();

const {
  addAction: addActionHelper,
  removeAction: removeActionHelper,
  updateAction: updateActionHelper,
} = useRuleActions(actionOptions);

interface Rule {
  id: number;
  name: string;
  enabled: boolean;
  conditions: Condition[];
  actions: string[];
  position?: number;
}

interface Props {
  show?: boolean;
  rule?: Rule | null;
}

const props = withDefaults(defineProps<Props>(), {
  show: false,
  rule: null,
});

const emit = defineEmits<{
  close: [];
  save: [rule: Rule];
}>();

// Form data
const ruleName = ref('');
const conditions: Ref<Condition[]> = ref([]);
const actions: Ref<string[]> = ref([]);

// Store initial state for unsaved changes detection
const initialState = ref<{
  ruleName: string;
  conditions: Condition[];
  actions: string[];
}>({
  ruleName: '',
  conditions: [],
  actions: [],
});

// Check if there are unsaved changes
const hasUnsavedChanges = computed(() => {
  // If it's a new rule and has any content, consider it unsaved
  if (!props.rule) {
    return ruleName.value !== '' || conditions.value.length > 0 || actions.value.length > 0;
  }

  // For existing rules, compare with initial state
  const currentConditions = JSON.stringify(conditions.value);
  const initialConditions = JSON.stringify(initialState.value.conditions);
  const currentActions = JSON.stringify(actions.value);
  const initialActions = JSON.stringify(initialState.value.actions);

  return (
    ruleName.value !== initialState.value.ruleName ||
    currentConditions !== initialConditions ||
    currentActions !== initialActions
  );
});

// Computed title
const modalTitle = computed(() => {
  return props.rule ? t('modal.rule.editRule') : t('modal.rule.addRule');
});

// Initialize form when rule changes
watch(
  () => props.rule,
  (newRule) => {
    if (newRule) {
      ruleName.value = newRule.name || '';
      conditions.value = newRule.conditions ? JSON.parse(JSON.stringify(newRule.conditions)) : [];
      actions.value = newRule.actions ? [...newRule.actions] : [];
    } else {
      ruleName.value = '';
      conditions.value = [];
      actions.value = [];
    }

    // Store initial state for unsaved changes detection
    initialState.value = {
      ruleName: ruleName.value,
      conditions: JSON.parse(JSON.stringify(conditions.value)),
      actions: [...actions.value],
    };
  },
  { immediate: true }
);

// Condition helpers
function addCondition(): void {
  addConditionHelper(conditions.value);
}

function removeCondition(index: number): void {
  removeConditionHelper(conditions.value, index);
}

function handleFieldChange(index: number): void {
  onFieldChange(conditions.value[index]);
}

function handleToggleNegate(index: number): void {
  toggleNegate(conditions.value[index]);
}

// Action helpers
function addAction(): void {
  addActionHelper(actions);
}

function removeAction(index: number): void {
  removeActionHelper(actions, index);
}

function updateAction(index: number, value: string): void {
  updateActionHelper(actions, index, value);
}

// Form validation
const isValid: ComputedRef<boolean> = computed(() => {
  return actions.value.length > 0;
});

// Save handler
function handleSave(): void {
  if (!isValid.value) {
    window.showToast(t('setting.rule.noActionsSelected'), 'warning');
    return;
  }

  const rule: Rule = {
    id: props.rule ? props.rule.id : Date.now(),
    name: ruleName.value || t('modal.rule.rules'),
    enabled: props.rule ? props.rule.enabled : true,
    conditions: conditions.value.filter((c) => {
      if (isMultiSelectField(c.field)) {
        return c.values && c.values.length > 0;
      }
      return c.value !== '';
    }),
    actions: [...actions.value],
  };

  emit('save', rule);
}

async function handleClose(checkUnsaved = false): Promise<void> {
  // Check for unsaved changes if requested
  if (checkUnsaved && hasUnsavedChanges.value) {
    const confirmed = await window.showConfirm({
      title: t('modal.common.unsavedChangesTitle'),
      message: t('modal.common.unsavedChangesMessage'),
      confirmText: t('common.action.discard'),
      cancelText: t('common.cancel'),
      isDanger: true,
    });

    if (!confirmed) {
      return;
    }
  }

  openDropdownIndex.value = null;
  emit('close');
}
</script>

<template>
  <BaseModal v-if="show" size="2xl" :z-index="70" @close="handleClose(true)">
    <!-- Custom Header -->
    <template #header>
      <h3 class="text-lg font-semibold m-0 flex items-center gap-2 text-text-primary">
        <PhLightning :size="20" />
        {{ modalTitle }}
      </h3>
    </template>

    <!-- Content -->
    <div class="p-4 sm:p-6 space-y-6">
      <!-- Rule Name -->
      <div class="space-y-2">
        <label class="block text-sm font-medium">{{ t('modal.rule.name') }}</label>
        <input
          v-model="ruleName"
          type="text"
          :placeholder="t('modal.rule.namePlaceholder')"
          class="input-field w-full"
        />
      </div>

      <!-- Conditions Section -->
      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <label class="flex items-center gap-2 text-sm font-medium">
            <PhFunnel :size="16" />
            {{ t('modal.rule.condition') }}
          </label>
        </div>

        <!-- Empty state -->
        <div
          v-if="conditions.length === 0"
          class="text-center text-text-secondary py-4 bg-bg-secondary rounded-lg border border-border"
        >
          <p class="text-sm">{{ t('modal.filter.conditionAlways') }}</p>
        </div>

        <!-- Condition list -->
        <div v-else class="space-y-3">
          <div v-for="(condition, index) in conditions" :key="condition.id">
            <!-- Logic connector -->
            <RuleLogicConnector
              v-if="index > 0"
              :logic="condition.logic || 'and'"
              @update="(logic) => (condition.logic = logic)"
            />

            <!-- Condition card -->
            <RuleConditionItem
              :condition="condition"
              :index="index"
              :is-dropdown-open="openDropdownIndex === index"
              @update:field="
                (value) => {
                  condition.field = value;
                  handleFieldChange(index);
                }
              "
              @update:operator="(value) => (condition.operator = value)"
              @update:value="(value) => (condition.value = value)"
              @update:values="(values) => (condition.values = values)"
              @update:negate="handleToggleNegate(index)"
              @toggle-dropdown="toggleDropdown(index)"
              @remove="removeCondition(index)"
            />
          </div>
        </div>

        <!-- Add condition button -->
        <button
          class="btn-secondary w-full flex items-center justify-center gap-2"
          @click="addCondition"
        >
          <PhPlus :size="16" />
          {{ t('modal.rule.addCondition') }}
        </button>
      </div>

      <!-- Actions Section -->
      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <label class="flex items-center gap-2 text-sm font-medium">
            <PhListChecks :size="16" />
            {{ t('modal.rule.actions') }}
          </label>
        </div>

        <!-- Empty state -->
        <div
          v-if="actions.length === 0"
          class="text-center text-text-secondary py-4 bg-bg-secondary rounded-lg border border-border"
        >
          <p class="text-sm">{{ t('setting.rule.noActionsSelected') }}</p>
        </div>

        <!-- Action list -->
        <div v-else class="space-y-2">
          <RuleAction
            v-for="(action, index) in actions"
            :key="index"
            :action="action"
            :index="index"
            :selected-actions="actions"
            :all-action-options="actionOptions"
            @update="(value) => updateAction(index, value)"
            @remove="removeAction(index)"
          />
        </div>

        <!-- Add action button -->
        <button
          class="btn-secondary w-full flex items-center justify-center gap-2"
          :disabled="actions.length >= actionOptions.length"
          @click="addAction"
        >
          <PhPlus :size="16" />
          {{ t('modal.rule.addAction') }}
        </button>
      </div>
    </div>

    <!-- Footer -->
    <template #footer>
      <ModalFooter
        align="right"
        :secondary-button="{
          label: t('common.cancel'),
          onClick: () => handleClose(true),
        }"
        :primary-button="{
          label: t('common.action.saveChanges'),
          disabled: !isValid,
          onClick: handleSave,
        }"
      />
    </template>
  </BaseModal>
</template>

<style scoped>
@reference "../../../style.css";

.input-field {
  @apply p-2 border border-border rounded-md bg-bg-primary text-text-primary text-sm focus:border-accent focus:outline-none transition-colors;
  height: 38px;
}
.btn-secondary {
  @apply bg-bg-tertiary text-text-primary border border-border px-4 py-2.5 rounded-lg cursor-pointer font-medium hover:bg-bg-secondary transition-colors disabled:opacity-50 disabled:cursor-not-allowed;
}
</style>

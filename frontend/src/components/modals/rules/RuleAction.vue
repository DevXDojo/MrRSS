<script setup lang="ts">
import { computed, type ComputedRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhTrash } from '@phosphor-icons/vue';
import BaseSelect from '@/components/common/BaseSelect.vue';
import type { SelectOption } from '@/types/select';

interface ActionOption {
  value: string;
  labelKey: string;
}

interface Props {
  action: string;
  index: number;
  selectedActions: string[];
  allActionOptions: ActionOption[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
  update: [value: string];
  remove: [];
}>();

const { t } = useI18n();

// Get available actions (exclude already selected ones, except current)
const availableActions: ComputedRef<ActionOption[]> = computed(() => {
  const selectedSet = new Set(props.selectedActions);
  return props.allActionOptions.filter(
    (opt) => !selectedSet.has(opt.value) || opt.value === props.action
  );
});

// Build options for BaseSelect
const actionSelectOptions = computed<SelectOption[]>(() => {
  return availableActions.value.map((opt) => ({
    value: opt.value,
    label: t(opt.labelKey),
  }));
});

function handleUpdate(value: string | number): void {
  emit('update', String(value));
}
</script>

<template>
  <div class="action-row">
    <span class="text-xs text-text-secondary">{{ index + 1 }}.</span>
    <BaseSelect
      :model-value="action"
      :options="actionSelectOptions"
      :searchable="true"
      @update:model-value="handleUpdate"
    />
    <button class="btn-danger-icon" :title="t('setting.rule.removeAction')" @click="emit('remove')">
      <PhTrash :size="16" />
    </button>
  </div>
</template>

<style scoped>
@reference "../../../style.css";
.action-row {
  @apply flex items-center gap-2 p-2 bg-bg-secondary border border-border rounded-lg;
}

.btn-danger-icon {
  @apply p-2 rounded-lg text-red-500 hover:bg-red-500/10 transition-colors cursor-pointer;
}
</style>

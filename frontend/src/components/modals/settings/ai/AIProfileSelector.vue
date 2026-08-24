<script setup lang="ts">
import { computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAIProfiles } from '@/composables/ai/useAIProfiles';
import BaseSelect from '@/components/common/BaseSelect.vue';

const { t } = useI18n();
const { profiles, fetchProfiles } = useAIProfiles();

interface Props {
  modelValue: string | null;
  disabled?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
});

const emit = defineEmits<{
  'update:modelValue': [value: string | null];
}>();

// Load profiles on mount
onMounted(() => {
  if (profiles.value.length === 0) {
    fetchProfiles();
  }
});

// Computed selected value (keep as string)
// If no value is set, show the first profile but don't emit update
const selectedValue = computed(() => {
  if (props.modelValue === null || props.modelValue === '') {
    return profiles.value.length > 0 ? String(profiles.value[0].id) : '';
  }
  return String(props.modelValue);
});

// Build options for BaseSelect
const profileOptions = computed(() => {
  if (profiles.value.length === 0) {
    return [{ value: '', label: t('setting.ai.noProfiles'), disabled: true }];
  }
  return profiles.value.map((profile) => ({
    value: String(profile.id),
    label: profile.name,
  }));
});

// Handle value change
function handleChange(value: string | number) {
  const stringValue = String(value);
  if (stringValue === '') {
    emit('update:modelValue', null);
  } else {
    emit('update:modelValue', stringValue);
  }
}
</script>

<template>
  <div class="ai-profile-selector">
    <BaseSelect
      :model-value="selectedValue"
      :options="profileOptions"
      :disabled="disabled || profiles.length === 0"
      :placeholder="profiles.length === 0 ? t('setting.ai.noProfiles') : ''"
      :searchable="true"
      @update:model-value="handleChange"
    />

    <!-- No profiles warning -->
    <div v-if="profiles.length === 0" class="text-xs text-text-tertiary mt-1">
      {{ t('setting.ai.noProfilesHint') }}
    </div>
  </div>
</template>

<style scoped>
/* Styles are now handled by BaseSelect and select.css */
</style>

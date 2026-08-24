<script setup lang="ts">
import { computed } from 'vue';
import type { SettingsData } from '@/types/settings';
import { useSettingsAutoSave } from '@/composables/core/useSettingsAutoSave';
import { useSettingsValidation } from '@/composables/core/useSettingsValidation';
import { useI18n } from 'vue-i18n';
import { PhWarning } from '@phosphor-icons/vue';
import TranslationSettings from './TranslationSettings.vue';
import SummarySettings from './SummarySettings.vue';

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();
const { t } = useI18n();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

// Create a computed ref that returns the settings object
// This ensures reactivity while allowing modifications
const settingsRef = computed(() => props.settings);

// Use composable for auto-save with reactivity
useSettingsAutoSave(settingsRef);

// Use validation composable
const { isValid, isTranslationValid, isSummaryValid } = useSettingsValidation(settingsRef);

// Handler for settings updates from child components
function handleUpdateSettings(updatedSettings: SettingsData) {
  // Emit the updated settings to parent
  emit('update:settings', updatedSettings);
}
</script>

<template>
  <div class="space-y-4 sm:space-y-6">
    <!-- Validation Warning -->
    <div
      v-if="!isValid"
      class="p-3 sm:p-4 rounded-lg border-2 border-red-500 bg-red-500/10 flex items-start gap-3"
    >
      <PhWarning :size="20" class="text-red-500 shrink-0 mt-0.5" :weight="'fill'" />
      <div class="flex-1">
        <div class="font-semibold text-red-500 text-sm sm:text-base mb-1">
          {{ t('common.form.requiredField') }}
        </div>
        <div class="text-xs sm:text-sm text-text-secondary">
          <span v-if="!isTranslationValid">
            {{ t('setting.content.translationCredentialsRequired') }}
          </span>
          <span v-if="!isTranslationValid && !isSummaryValid"> • </span>
          <span v-if="!isSummaryValid">
            {{ t('setting.content.summaryCredentialsRequired') }}
          </span>
        </div>
      </div>
    </div>

    <TranslationSettings :settings="settings" @update:settings="handleUpdateSettings" />

    <SummarySettings :settings="settings" @update:settings="handleUpdateSettings" />
  </div>
</template>

<style scoped></style>

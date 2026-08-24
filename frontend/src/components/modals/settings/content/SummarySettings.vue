<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhTextAlignLeft,
  PhTextT,
  PhPackage,
  PhRobot,
  PhTrash,
  PhBroom,
} from '@phosphor-icons/vue';
import {
  SettingGroup,
  SettingWithToggle,
  NestedSettingsContainer,
  SubSettingItem,
  TextAreaControl,
} from '@/components/settings';
import BaseSelect from '@/components/common/BaseSelect.vue';
import AIProfileSelector from '@/components/modals/settings/ai/AIProfileSelector.vue';
import '@/components/settings/styles.css';
import type { SettingsData } from '@/types/settings';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

function updateSetting(key: keyof SettingsData, value: any) {
  emit('update:settings', {
    ...props.settings,
    [key]: value,
  });
}

async function clearSummaryCache() {
  const confirmed = await window.showConfirm({
    title: t('setting.content.clearSummaryCache'),
    message: t('setting.content.clearSummaryCacheConfirm'),
    isDanger: true,
  });
  if (!confirmed) return;

  isClearingCache.value = true;
  try {
    const response = await fetch('/api/articles/clear-summaries', {
      method: 'DELETE',
    });

    if (response.ok) {
      window.showToast(t('setting.content.clearSummaryCacheSuccess'), 'success');
      // Refresh article list to show updated summaries
      window.dispatchEvent(new CustomEvent('refresh-articles'));
    } else {
      console.error('Server error:', response.status);
      window.showToast(t('setting.content.clearSummaryCacheFailed'), 'error');
    }
  } catch (error) {
    console.error('Failed to clear summary cache:', error);
    window.showToast(t('setting.content.clearSummaryCacheFailed'), 'error');
  } finally {
    isClearingCache.value = false;
  }
}

const isClearingCache = ref(false);
</script>

<template>
  <SettingGroup :icon="PhTextAlignLeft" :title="t('setting.content.summary')">
    <SettingWithToggle
      :icon="PhTextT"
      :title="t('setting.content.enableSummary')"
      :description="t('setting.content.enableSummaryDesc')"
      :model-value="settings.summary_enabled"
      @update:model-value="updateSetting('summary_enabled', $event)"
    />

    <NestedSettingsContainer v-if="settings.summary_enabled">
      <SubSettingItem
        :icon="PhPackage"
        :title="t('setting.content.summaryProvider')"
        :description="t('setting.content.summaryProviderDesc')"
      >
        <BaseSelect
          :model-value="settings.summary_provider"
          :options="[
            { value: 'rss', label: t('setting.content.rssSummary') },
            { value: 'local', label: t('setting.content.localAlgorithm') },
            { value: 'ai', label: t('setting.content.aiSummary') },
          ]"
          width="w-32 sm:w-48"
          @update:model-value="updateSetting('summary_provider', $event)"
        />
      </SubSettingItem>

      <!-- AI Summary Prompt -->
      <template v-if="settings.summary_provider === 'ai'">
        <!-- AI Profile Selection -->
        <SubSettingItem
          :icon="PhRobot"
          :title="t('setting.ai.selectProfile')"
          :description="t('setting.ai.selectProfileForSummary')"
        >
          <AIProfileSelector
            :model-value="settings.ai_summary_profile_id"
            @update:model-value="updateSetting('ai_summary_profile_id', $event)"
          />
        </SubSettingItem>

        <div class="sub-setting-item-col">
          <div class="flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
            <PhRobot :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
            <div class="flex-1 min-w-0">
              <div class="font-medium mb-0 sm:mb-1 text-xs sm:text-sm">
                {{ t('setting.content.aiSummaryPrompt') }}
              </div>
              <div class="text-[10px] sm:text-xs text-text-secondary hidden sm:block">
                {{ t('setting.content.aiSummaryPromptDesc') }}
              </div>
            </div>
          </div>
          <TextAreaControl
            :model-value="settings.ai_summary_prompt"
            :placeholder="t('setting.content.aiSummaryPromptPlaceholder')"
            :rows="3"
            @update:model-value="updateSetting('ai_summary_prompt', $event)"
          />
        </div>

        <SubSettingItem
          :icon="PhRobot"
          :title="t('setting.content.summaryTriggerMode')"
          :description="t('setting.content.summaryTriggerModeDesc')"
        >
          <BaseSelect
            :model-value="settings.summary_trigger_mode"
            :options="[
              { value: 'auto', label: t('setting.content.summaryTriggerModeAuto') },
              { value: 'manual', label: t('setting.content.summaryTriggerModeManual') },
            ]"
            width="w-32 sm:w-48"
            @update:model-value="updateSetting('summary_trigger_mode', $event)"
          />
        </SubSettingItem>
      </template>

      <SubSettingItem
        :icon="PhTextAlignLeft"
        :title="t('setting.content.summaryLength')"
        :description="t('setting.content.summaryLengthDesc')"
      >
        <BaseSelect
          :model-value="settings.summary_length"
          :options="[
            { value: 'short', label: t('setting.content.summaryLengthShort') },
            { value: 'medium', label: t('setting.content.summaryLengthMedium') },
            { value: 'long', label: t('setting.content.summaryLengthLong') },
          ]"
          width="w-32 sm:w-48"
          @update:model-value="updateSetting('summary_length', $event)"
        />
      </SubSettingItem>

      <!-- Cache Management -->
      <SubSettingItem
        :icon="PhTrash"
        :title="t('setting.content.clearSummaryCache')"
        :description="t('setting.content.clearSummaryCacheDesc')"
      >
        <button
          type="button"
          :disabled="isClearingCache"
          class="btn-secondary"
          @click="clearSummaryCache"
        >
          <PhBroom :size="16" class="sm:w-5 sm:h-5" />
          {{
            isClearingCache
              ? t('setting.database.cleaning')
              : t('setting.content.clearSummaryCacheButton')
          }}
        </button>
      </SubSettingItem>
    </NestedSettingsContainer>
  </SettingGroup>
</template>

<style scoped>
/* Styles are now handled by BaseSelect and select.css */
</style>

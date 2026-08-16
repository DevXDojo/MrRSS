<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhKey, PhFile } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';
import {
  NestedSettingsContainer,
  SubSettingItem,
  InputControl,
  TipBox,
} from '@/components/settings';

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
</script>

<template>
  <!-- Enable Notion Integration -->
  <div class="setting-item">
    <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
      <img
        src="/assets/plugin_icons/notion.svg"
        alt="Notion"
        class="w-5 h-5 sm:w-6 sm:h-6 mt-0.5 shrink-0"
      />
      <div class="flex-1 min-w-0">
        <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
          {{ t('setting.plugins.notion.integration') }}
        </div>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('setting.plugins.notion.integrationDescription') }}
        </div>
      </div>
    </div>
    <input
      type="checkbox"
      :checked="props.settings.notion_enabled"
      class="toggle"
      @change="updateSetting('notion_enabled', ($event.target as HTMLInputElement).checked)"
    />
  </div>

  <NestedSettingsContainer v-if="props.settings.notion_enabled">
    <!-- Help text -->
    <TipBox type="help" :title="t('setting.plugins.notion.setupInstructions')">
      <ol>
        <li>{{ t('setting.plugins.notion.step1') }}</li>
        <li>{{ t('setting.plugins.notion.step2') }}</li>
        <li>{{ t('setting.plugins.notion.step3') }}</li>
        <li>{{ t('setting.plugins.notion.step4') }}</li>
      </ol>
    </TipBox>

    <!-- API Key -->
    <SubSettingItem
      :icon="PhKey"
      :title="t('setting.plugins.notion.apiKey')"
      :description="t('setting.plugins.notion.apiKeyDesc')"
      required
    >
      <InputControl
        :model-value="props.settings.notion_api_key"
        :placeholder="t('setting.plugins.notion.apiKeyPlaceholder')"
        type="password"
        width="lg"
        @update:model-value="updateSetting('notion_api_key', $event)"
      />
    </SubSettingItem>

    <!-- Page ID -->
    <SubSettingItem
      :icon="PhFile"
      :title="t('setting.plugins.notion.pageId')"
      :description="t('setting.plugins.notion.pageIdDesc')"
      required
    >
      <InputControl
        :model-value="props.settings.notion_page_id"
        :placeholder="t('setting.plugins.notion.pageIdPlaceholder')"
        width="lg"
        @update:model-value="updateSetting('notion_page_id', $event)"
      />
    </SubSettingItem>
  </NestedSettingsContainer>
</template>

<style scoped>
@reference "../../../../style.css";
.toggle {
  @apply w-10 h-5 appearance-none bg-bg-tertiary rounded-full relative cursor-pointer border border-border transition-colors checked:bg-accent checked:border-accent shrink-0;
}
.toggle::after {
  content: '';
  @apply absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full shadow-sm transition-transform;
}
.toggle:checked::after {
  transform: translateX(20px);
}

.setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border;
}
</style>

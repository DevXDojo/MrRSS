<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhKey, PhIdentificationCard } from '@phosphor-icons/vue';
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
  <!-- Enable Zotero Integration -->
  <div class="setting-item">
    <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
      <img
        src="/assets/plugin_icons/zotero.png"
        alt="Zotero"
        class="w-5 h-5 sm:w-6 sm:h-6 mt-0.5 shrink-0"
      />
      <div class="flex-1 min-w-0">
        <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
          {{ t('setting.plugins.zotero.integration') }}
        </div>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('setting.plugins.zotero.integrationDescription') }}
        </div>
      </div>
    </div>
    <input
      type="checkbox"
      :checked="props.settings.zotero_enabled"
      class="toggle"
      @change="updateSetting('zotero_enabled', ($event.target as HTMLInputElement).checked)"
    />
  </div>

  <NestedSettingsContainer v-if="props.settings.zotero_enabled">
    <!-- Help text -->
    <TipBox type="help" :title="t('setting.plugins.zotero.setupInstructions')">
      <ol>
        <li>{{ t('setting.plugins.zotero.step1') }}</li>
        <li>{{ t('setting.plugins.zotero.step2') }}</li>
        <li>{{ t('setting.plugins.zotero.step3') }}</li>
        <li>{{ t('setting.plugins.zotero.step4') }}</li>
      </ol>
    </TipBox>

    <!-- User ID -->
    <SubSettingItem
      :icon="PhIdentificationCard"
      :title="t('setting.plugins.zotero.userId')"
      :description="t('setting.plugins.zotero.userIdDesc')"
      required
    >
      <InputControl
        :model-value="props.settings.zotero_user_id"
        :placeholder="t('setting.plugins.zotero.userIdPlaceholder')"
        width="lg"
        @update:model-value="updateSetting('zotero_user_id', $event)"
      />
    </SubSettingItem>

    <!-- API Key -->
    <SubSettingItem
      :icon="PhKey"
      :title="t('setting.plugins.zotero.apiKey')"
      :description="t('setting.plugins.zotero.apiKeyDesc')"
      required
    >
      <InputControl
        :model-value="props.settings.zotero_api_key"
        :placeholder="t('setting.plugins.zotero.apiKeyPlaceholder')"
        type="password"
        width="lg"
        @update:model-value="updateSetting('zotero_api_key', $event)"
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

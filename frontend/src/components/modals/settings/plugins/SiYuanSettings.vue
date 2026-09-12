<script setup lang="ts">
import SiYuanIcon from '@/components/common/SiYuanIcon.vue';
import { useI18n } from 'vue-i18n';
import { PhNotebook, PhKey, PhGlobe, PhFolder } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';
import {
  SettingWithToggle,
  NestedSettingsContainer,
  SubSettingItem,
  InputControl,
  TipBox,
} from '@/components/settings';

const { t } = useI18n();
const props = defineProps<{ settings: SettingsData }>();
const emit = defineEmits<{ 'update:settings': [settings: SettingsData] }>();
function updateSetting(key: keyof SettingsData, value: string | boolean) {
  emit('update:settings', { ...props.settings, [key]: value });
}
</script>

<template>
  <div>
    <SettingWithToggle
      :icon="SiYuanIcon"
      :title="t('setting.plugins.siyuan.integration')"
      :description="t('setting.plugins.siyuan.description')"
      :model-value="settings.siyuan_enabled"
      @update:model-value="updateSetting('siyuan_enabled', $event)"
    />
    <NestedSettingsContainer v-if="settings.siyuan_enabled">
      <TipBox type="help" :title="t('setting.plugins.siyuan.setup')">
        <p>{{ t('setting.plugins.siyuan.instructions') }}</p>
      </TipBox>
      <SubSettingItem
        :icon="PhGlobe"
        :title="t('setting.plugins.siyuan.endpoint')"
        :description="t('setting.plugins.siyuan.endpointDesc')"
        required
      >
        <InputControl
          :model-value="settings.siyuan_endpoint"
          type="url"
          width="lg"
          placeholder="http://127.0.0.1:6806"
          @update:model-value="updateSetting('siyuan_endpoint', $event)"
        />
      </SubSettingItem>
      <SubSettingItem :icon="PhKey" :title="t('setting.plugins.siyuan.token')">
        <InputControl
          :model-value="settings.siyuan_api_token"
          type="password"
          width="lg"
          @update:model-value="updateSetting('siyuan_api_token', $event)"
        />
      </SubSettingItem>
      <SubSettingItem
        :icon="PhNotebook"
        :title="t('setting.plugins.siyuan.notebook')"
        :description="t('setting.plugins.siyuan.notebookDesc')"
        required
      >
        <InputControl
          :model-value="settings.siyuan_notebook_id"
          width="lg"
          @update:model-value="updateSetting('siyuan_notebook_id', $event)"
        />
      </SubSettingItem>
      <SubSettingItem
        :icon="PhFolder"
        :title="t('setting.plugins.siyuan.folder')"
        :description="t('setting.plugins.siyuan.folderDesc')"
      >
        <InputControl
          :model-value="settings.siyuan_folder"
          width="lg"
          placeholder="/MrRSS"
          @update:model-value="updateSetting('siyuan_folder', $event)"
        />
      </SubSettingItem>
    </NestedSettingsContainer>
  </div>
</template>

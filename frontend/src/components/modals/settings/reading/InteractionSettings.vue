<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhCursorClick, PhEyeSlash } from '@phosphor-icons/vue';
import { SettingGroup, SettingWithToggle } from '@/components/settings';
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
</script>

<template>
  <SettingGroup :icon="PhCursorClick" :title="t('setting.tab.interactionSettings')">
    <SettingWithToggle
      :icon="PhCursorClick"
      :title="t('setting.reading.hoverMarkAsRead')"
      :description="t('setting.reading.hoverMarkAsReadDesc')"
      :model-value="settings.hover_mark_as_read"
      @update:model-value="updateSetting('hover_mark_as_read', $event)"
    />

    <SettingWithToggle
      :icon="PhEyeSlash"
      :title="t('setting.reading.showHiddenArticles')"
      :description="t('setting.reading.showHiddenArticlesDesc')"
      :model-value="settings.show_hidden_articles"
      @update:model-value="updateSetting('show_hidden_articles', $event)"
    />
  </SettingGroup>
</template>

<style scoped></style>

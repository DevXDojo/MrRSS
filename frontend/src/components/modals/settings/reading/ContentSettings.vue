<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhArticleNyTimes, PhImages, PhPlayCircle } from '@phosphor-icons/vue';
import {
  SettingGroup,
  SettingWithToggle,
  SubSettingItem,
  NestedSettingsContainer,
  ToggleControl,
} from '@/components/settings';
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
  <SettingGroup :icon="PhArticleNyTimes" :title="t('setting.tab.contentSettings')">
    <SettingWithToggle
      :icon="PhArticleNyTimes"
      :title="t('setting.feed.enableFullTextFetch')"
      :description="t('setting.feed.enableFullTextFetchDesc')"
      :model-value="settings.full_text_fetch_enabled"
      @update:model-value="updateSetting('full_text_fetch_enabled', $event)"
    />

    <NestedSettingsContainer v-if="settings.full_text_fetch_enabled">
      <SubSettingItem
        :icon="PhPlayCircle"
        :title="t('setting.reading.autoShowAllContent')"
        :description="t('setting.reading.autoShowAllContentDesc')"
      >
        <ToggleControl
          :model-value="settings.auto_show_all_content"
          @update:model-value="updateSetting('auto_show_all_content', $event)"
        />
      </SubSettingItem>
    </NestedSettingsContainer>

    <SettingWithToggle
      :icon="PhImages"
      :title="t('setting.reading.imageGalleryEnabled')"
      :description="t('setting.reading.imageGalleryEnabledDesc')"
      :model-value="settings.image_gallery_enabled"
      @update:model-value="updateSetting('image_gallery_enabled', $event)"
    />
  </SettingGroup>
</template>

<style scoped></style>

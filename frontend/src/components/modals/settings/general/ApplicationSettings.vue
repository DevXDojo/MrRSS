<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhPalette,
  PhMoon,
  PhTranslate,
  PhPower,
  PhArchiveTray,
  PhTextT,
  PhTextAa,
} from '@phosphor-icons/vue';
import {
  SettingGroup,
  SettingItem,
  SettingWithToggle,
  SettingWithSelect,
  NumberControl,
} from '@/components/settings';
import FontFamilySelect from '@/components/settings/FontFamilySelect.vue';
import type { SettingsData } from '@/types/settings';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

const displayUiSize = computed(() => {
  return parseInt(props.settings.ui_font_size as any) || 16;
});

function updateSetting(key: keyof SettingsData, value: any) {
  emit('update:settings', {
    ...props.settings,
    [key]: value,
  });
}

function updateUiFontSize(value: number) {
  const nextValue = Number.isFinite(value) ? Math.min(20, Math.max(12, value)) : 16;
  updateSetting('ui_font_size', nextValue);
}
</script>

<template>
  <SettingGroup :icon="PhPalette" :title="t('setting.general.application')">
    <SettingWithToggle
      :icon="PhPower"
      :title="t('setting.general.startupOnBoot')"
      :description="t('setting.general.startupOnBootDesc')"
      :model-value="settings.startup_on_boot"
      @update:model-value="updateSetting('startup_on_boot', $event)"
    />

    <SettingWithToggle
      :icon="PhArchiveTray"
      :title="t('setting.general.closeToTray')"
      :description="t('setting.general.closeToTrayDesc')"
      :model-value="settings.close_to_tray"
      @update:model-value="updateSetting('close_to_tray', $event)"
    />

    <SettingWithSelect
      :icon="PhMoon"
      :title="t('setting.general.theme')"
      :description="t('setting.general.themeDesc')"
      :model-value="settings.theme"
      :options="[
        { value: 'light', label: t('setting.general.light') },
        { value: 'dark', label: t('setting.general.dark') },
        { value: 'auto', label: t('setting.general.auto') },
      ]"
      width="md"
      @update:model-value="updateSetting('theme', $event)"
    />

    <SettingWithSelect
      :icon="PhTranslate"
      :title="t('setting.general.language')"
      :description="t('setting.general.languageDesc')"
      :model-value="settings.language"
      :options="[
        { value: 'en-US', label: t('common.language.english') },
        { value: 'zh-CN', label: t('common.language.chinese') },
      ]"
      width="md"
      @update:model-value="updateSetting('language', $event)"
    />

    <SettingItem :icon="PhTextT" :title="t('setting.general.uiFontFamily')">
      <template #description>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('setting.general.uiFontFamilyDesc') }}
        </div>
      </template>
      <FontFamilySelect
        :model-value="settings.ui_font_family"
        @update:model-value="updateSetting('ui_font_family', $event)"
      />
    </SettingItem>

    <SettingItem :icon="PhTextAa" :title="t('setting.general.uiFontSize')">
      <template #description>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('setting.general.uiFontSizeDesc') }}
        </div>
      </template>
      <NumberControl
        :model-value="displayUiSize"
        :min="12"
        :max="20"
        suffix="px"
        @update:model-value="updateUiFontSize"
      />
    </SettingItem>
  </SettingGroup>
</template>

<style scoped></style>

<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhTextT, PhTextIndent, PhTextAa } from '@phosphor-icons/vue';
import { SettingGroup, SettingItem, NumberControl } from '@/components/settings';
import FontFamilySelect from '@/components/settings/FontFamilySelect.vue';
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

// Computed values for display (handle string/number conversion)
const displayContentSize = computed(() => {
  return parseInt(props.settings.content_font_size as any) || 16;
});
const displayLineHeight = computed(() => {
  return parseFloat(props.settings.content_line_height as any) || 1.6;
});

function updateSetting(key: keyof SettingsData, value: any) {
  emit('update:settings', {
    ...props.settings,
    [key]: value,
  });
}
</script>

<template>
  <SettingGroup :icon="PhTextT" :title="t('setting.tab.typography')">
    <!-- Content Font Family -->
    <SettingItem :icon="PhTextT" :title="t('setting.typography.contentFontFamily')">
      <template #description>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('setting.typography.contentFontFamilyDesc') }}
        </div>
      </template>
      <FontFamilySelect
        :model-value="settings.content_font_family"
        @update:model-value="updateSetting('content_font_family', $event)"
      />
    </SettingItem>

    <!-- Content Font Size -->
    <SettingItem :icon="PhTextAa" :title="t('setting.typography.contentFontSize')">
      <template #description>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('setting.typography.contentFontSizeDesc') }}
        </div>
      </template>
      <NumberControl
        :model-value="displayContentSize"
        :min="10"
        :max="24"
        suffix="px"
        @update:model-value="(v) => updateSetting('content_font_size', isNaN(v) ? 16 : v)"
      />
    </SettingItem>

    <!-- Content Line Height -->
    <SettingItem :icon="PhTextIndent" :title="t('setting.typography.contentLineHeight')">
      <template #description>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('setting.typography.contentLineHeightDesc') }}
        </div>
      </template>
      <NumberControl
        :model-value="displayLineHeight"
        :min="1"
        :max="3"
        :step="0.1"
        @update:model-value="
          (v) => updateSetting('content_line_height', isNaN(v) ? '1.6' : v.toString())
        "
      />
    </SettingItem>
  </SettingGroup>
</template>

<style scoped>
/* Styles are now handled by BaseSelect and select.css */
</style>

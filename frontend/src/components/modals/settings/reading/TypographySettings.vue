<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhTextT, PhTextIndent, PhTextAa } from '@phosphor-icons/vue';
import { SettingGroup, SettingItem, NumberControl } from '@/components/settings';
import BaseSelect from '@/components/common/BaseSelect.vue';
import type { SelectOption, SelectOptionGroup } from '@/types/select';
import '@/components/settings/styles.css';
import type { SettingsData } from '@/types/settings';
import { getRecommendedFonts } from '@/utils/fontDetector';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

// Font categories
const availableFonts = ref<{
  serif: string[];
  sansSerif: string[];
  monospace: string[];
}>({
  serif: [],
  sansSerif: [],
  monospace: [],
});

// Computed values for display (handle string/number conversion)
const displayContentSize = computed(() => {
  return parseInt(props.settings.content_font_size as any) || 16;
});
const displayUiSize = computed(() => {
  return parseInt(props.settings.ui_font_size as any) || 16;
});
const displayLineHeight = computed(() => {
  return parseFloat(props.settings.content_line_height as any) || 1.6;
});

// Build font options with groups
const fontOptions = computed<SelectOptionGroup[]>(() => {
  const groups: SelectOptionGroup[] = [];

  // System fonts
  groups.push({
    label: t('setting.typography.fontSystem'),
    options: [
      {
        value: 'system',
        label: t('setting.typography.fontSystemDefault'),
      },
    ],
  });

  // Serif fonts
  if (availableFonts.value.serif.length > 0) {
    const serifOptions: SelectOption[] = [
      {
        value: 'serif',
        label: t('setting.typography.fontSerifDefault'),
      },
    ];
    for (const font of availableFonts.value.serif) {
      serifOptions.push({
        value: font,
        label: font,
        style: { fontFamily: font + ', serif' }, // Custom style for font preview
      });
    }
    groups.push({
      label: t('setting.typography.fontSerif'),
      options: serifOptions,
    });
  }

  // Sans-serif fonts
  if (availableFonts.value.sansSerif.length > 0) {
    const sansSerifOptions: SelectOption[] = [
      {
        value: 'sans-serif',
        label: t('setting.typography.fontSansSerifDefault'),
      },
    ];
    for (const font of availableFonts.value.sansSerif) {
      sansSerifOptions.push({
        value: font,
        label: font,
        style: { fontFamily: font + ', sans-serif' }, // Custom style for font preview
      });
    }
    groups.push({
      label: t('setting.typography.fontSansSerif'),
      options: sansSerifOptions,
    });
  }

  // Monospace fonts
  if (availableFonts.value.monospace.length > 0) {
    const monospaceOptions: SelectOption[] = [
      {
        value: 'monospace',
        label: t('setting.typography.fontMonospaceDefault'),
      },
    ];
    for (const font of availableFonts.value.monospace) {
      monospaceOptions.push({
        value: font,
        label: font,
        style: { fontFamily: font + ', monospace' }, // Custom style for font preview
      });
    }
    groups.push({
      label: t('setting.typography.fontMonospace'),
      options: monospaceOptions,
    });
  }

  return groups;
});

// Load system fonts on mount
onMounted(() => {
  try {
    availableFonts.value = getRecommendedFonts();
  } catch (error) {
    console.error('Failed to detect system fonts:', error);
  }
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
  <SettingGroup :icon="PhTextT" :title="t('setting.tab.typography')">
    <!-- Interface Font Family -->
    <SettingItem :icon="PhTextT" :title="t('setting.typography.uiFontFamily')">
      <template #description>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('setting.typography.uiFontFamilyDesc') }}
        </div>
      </template>
      <BaseSelect
        :model-value="settings.ui_font_family"
        :options="fontOptions"
        :searchable="true"
        width="w-36 sm:w-48"
        max-height="max-h-60"
        @update:model-value="updateSetting('ui_font_family', $event)"
      >
        <template #option="{ option }">
          <span :style="option.style">{{ option.label }}</span>
        </template>
      </BaseSelect>
    </SettingItem>

    <!-- Interface Font Size -->
    <SettingItem :icon="PhTextAa" :title="t('setting.typography.uiFontSize')">
      <template #description>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('setting.typography.uiFontSizeDesc') }}
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

    <!-- Content Font Family -->
    <SettingItem :icon="PhTextT" :title="t('setting.typography.contentFontFamily')">
      <template #description>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('setting.typography.contentFontFamilyDesc') }}
        </div>
      </template>
      <BaseSelect
        :model-value="settings.content_font_family"
        :options="fontOptions"
        :searchable="true"
        width="w-36 sm:w-48"
        max-height="max-h-60"
        @update:model-value="updateSetting('content_font_family', $event)"
      >
        <template #option="{ option }">
          <span :style="option.style">{{ option.label }}</span>
        </template>
      </BaseSelect>
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

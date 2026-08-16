<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhTextT } from '@phosphor-icons/vue';
import { SettingGroup, SettingItem } from '@/components/settings';
import {
  READER_FONT_SIZE,
  READER_LINE_HEIGHT,
  READER_MAX_WIDTH,
  clampReaderSetting,
} from '@/constants/reader';
import type { SettingsData } from '@/types/settings';

const props = defineProps<{ settings: SettingsData }>();
const emit = defineEmits<{ 'update:settings': [settings: SettingsData] }>();
const { t } = useI18n();

const fontOptions = computed(() => [
  { value: 'serif', label: t('setting.typography.fontSerif') },
  { value: 'sans-serif', label: t('setting.typography.fontSansSerif') },
  { value: 'hyperlegible', label: t('setting.typography.fontHyperlegible') },
]);
const fontSize = computed(() =>
  clampReaderSetting(
    props.settings.content_font_size,
    READER_FONT_SIZE.min,
    READER_FONT_SIZE.max,
    READER_FONT_SIZE.default
  )
);
const leadingPercent = computed(() =>
  clampReaderSetting(
    Math.round(Number(props.settings.content_line_height) * 100),
    READER_LINE_HEIGHT.min,
    READER_LINE_HEIGHT.max,
    READER_LINE_HEIGHT.default
  )
);
const readerWidth = computed(() =>
  clampReaderSetting(
    props.settings.reader_max_width,
    READER_MAX_WIDTH.min,
    READER_MAX_WIDTH.max,
    READER_MAX_WIDTH.default
  )
);

function updateSetting<K extends keyof SettingsData>(key: K, value: SettingsData[K]) {
  emit('update:settings', { ...props.settings, [key]: value });
}
</script>

<template>
  <SettingGroup :icon="PhTextT" :title="t('setting.tab.typography')">
    <SettingItem :title="t('setting.typography.contentFontFamily')">
      <template #description>{{ t('setting.typography.contentFontFamilyDesc') }}</template>
      <div
        class="reader-segments"
        role="group"
        :aria-label="t('setting.typography.contentFontFamily')"
      >
        <button
          v-for="option in fontOptions"
          :key="option.value"
          type="button"
          :class="{ selected: settings.content_font_family === option.value }"
          :aria-pressed="settings.content_font_family === option.value"
          @click="updateSetting('content_font_family', option.value)"
        >
          {{ option.label }}
        </button>
      </div>
    </SettingItem>

    <SettingItem :title="t('setting.typography.contentFontSize')">
      <template #description>{{ t('setting.typography.contentFontSizeDesc') }}</template>
      <label class="reader-slider">
        <input
          type="range"
          :min="READER_FONT_SIZE.min"
          :max="READER_FONT_SIZE.max"
          :step="READER_FONT_SIZE.step"
          :value="fontSize"
          :aria-label="t('setting.typography.contentFontSize')"
          :aria-valuemin="READER_FONT_SIZE.min"
          :aria-valuemax="READER_FONT_SIZE.max"
          :aria-valuenow="fontSize"
          :aria-valuetext="`${fontSize}px`"
          @input="
            updateSetting('content_font_size', Number(($event.target as HTMLInputElement).value))
          "
        />
        <output>{{ fontSize }}px</output>
      </label>
    </SettingItem>

    <SettingItem :title="t('setting.typography.contentLineHeight')">
      <template #description>{{ t('setting.typography.contentLineHeightDesc') }}</template>
      <label class="reader-slider">
        <input
          type="range"
          :min="READER_LINE_HEIGHT.min"
          :max="READER_LINE_HEIGHT.max"
          :step="READER_LINE_HEIGHT.step"
          :value="leadingPercent"
          :aria-label="t('setting.typography.contentLineHeight')"
          :aria-valuemin="READER_LINE_HEIGHT.min"
          :aria-valuemax="READER_LINE_HEIGHT.max"
          :aria-valuenow="leadingPercent"
          :aria-valuetext="`${leadingPercent}%`"
          @input="
            updateSetting(
              'content_line_height',
              (Number(($event.target as HTMLInputElement).value) / 100).toFixed(2)
            )
          "
        />
        <output>{{ leadingPercent }}%</output>
      </label>
    </SettingItem>

    <SettingItem :title="t('setting.typography.readerMaxWidth')">
      <template #description>{{ t('setting.typography.readerMaxWidthDesc') }}</template>
      <label class="reader-slider">
        <input
          type="range"
          :min="READER_MAX_WIDTH.min"
          :max="READER_MAX_WIDTH.max"
          :step="READER_MAX_WIDTH.step"
          :value="readerWidth"
          :aria-label="t('setting.typography.readerMaxWidth')"
          :aria-valuemin="READER_MAX_WIDTH.min"
          :aria-valuemax="READER_MAX_WIDTH.max"
          :aria-valuenow="readerWidth"
          :aria-valuetext="`${readerWidth}px`"
          @input="
            updateSetting('reader_max_width', Number(($event.target as HTMLInputElement).value))
          "
        />
        <output>{{ readerWidth }}px</output>
      </label>
    </SettingItem>
  </SettingGroup>
</template>

<style scoped>
.reader-segments {
  display: flex;
  padding: 3px;
  border: 1px solid var(--border-subtle);
  border-radius: 10px;
  background: var(--surface-muted);
}
.reader-segments button {
  padding: 7px 12px;
  border-radius: 7px;
  color: var(--text-secondary);
  font-size: 0.8125rem;
}

.reader-segments button.selected {
  color: var(--text-primary);
  background: var(--surface-bg);
  box-shadow: 0 1px 3px rgb(50 35 20 / 10%);
}

.reader-segments button:focus-visible,
.reader-slider input:focus-visible {
  outline: 2px solid var(--accent-color) !important;
  outline-offset: 2px;
}

.reader-slider {
  display: grid;
  grid-template-columns: minmax(150px, 245px) 52px;
  align-items: center;
  gap: 14px;
}

.reader-slider input {
  width: 100%;
  accent-color: var(--accent-color);
}

.reader-slider output {
  color: var(--text-secondary);
  text-align: right;
  font-variant-numeric: tabular-nums;
}

@media (max-width: 640px) {
  .reader-slider {
    grid-template-columns: minmax(110px, 1fr) 48px;
  }
}
</style>

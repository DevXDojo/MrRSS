<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhBookmarkSimple, PhCalendarCheck, PhCheckCircle, PhCursorClick, PhEyeSlash } from '@phosphor-icons/vue';
import { NestedSettingsContainer, NumberControl, SettingGroup, SettingWithSelect, SettingWithToggle, SubSettingItem } from '@/components/settings';
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

const autoReadPresetDays = computed({
  get: () => ([1, 3, 7, 30, 90].includes(props.settings.auto_mark_read_days) ? props.settings.auto_mark_read_days : 'custom'),
  set: (value: string | number) => {
    if (value !== 'custom') updateSetting('auto_mark_read_days', Number(value));
  },
});

const autoReadOptions = computed(() => [
  ...[1, 3, 7, 30, 90].map((days) => ({
    value: days,
    label: t('setting.reading.autoMarkReadDaysOption', { count: days }),
  })),
  { value: 'custom', label: t('setting.reading.autoMarkReadCustom') },
]);
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
      :icon="PhCheckCircle"
      :title="t('setting.reading.confirmMarkAsRead')"
      :description="t('setting.reading.confirmMarkAsReadDesc')"
      :model-value="settings.confirm_mark_as_read"
      @update:model-value="updateSetting('confirm_mark_as_read', $event)"
    />

    <SettingWithToggle
      :icon="PhCursorClick"
      :title="t('setting.reading.scrollMarkAsRead')"
      :description="t('setting.reading.scrollMarkAsReadDesc')"
      :model-value="settings.scroll_mark_as_read"
      @update:model-value="updateSetting('scroll_mark_as_read', $event)"
    />

    <SettingWithToggle
      :icon="PhCalendarCheck"
      :title="t('setting.reading.autoMarkRead')"
      :description="t('setting.reading.autoMarkReadDesc')"
      :model-value="settings.auto_mark_read_enabled"
      @update:model-value="updateSetting('auto_mark_read_enabled', $event)"
    />
    <NestedSettingsContainer v-if="settings.auto_mark_read_enabled">
      <SettingWithSelect
        :icon="PhCalendarCheck"
        :title="t('setting.reading.autoMarkReadAfter')"
        :model-value="autoReadPresetDays"
        :options="autoReadOptions"
        @update:model-value="autoReadPresetDays = $event"
      />
      <SubSettingItem
        v-if="autoReadPresetDays === 'custom'"
        :title="t('setting.reading.autoMarkReadCustomDays')"
        :description="t('setting.reading.autoMarkReadCustomDaysDesc')"
      >
        <NumberControl
          :model-value="settings.auto_mark_read_days"
          :min="1"
          :max="3650"
          :suffix="t('common.time.days')"
          @update:model-value="updateSetting('auto_mark_read_days', Math.max(1, $event))"
        />
      </SubSettingItem>
    </NestedSettingsContainer>

    <SettingWithToggle
      :icon="PhBookmarkSimple"
      :title="t('setting.reading.rememberArticlePosition')"
      :description="t('setting.reading.rememberArticlePositionDesc')"
      :model-value="settings.remember_article_position"
      @update:model-value="updateSetting('remember_article_position', $event)"
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

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';
import { formatExactDateTime } from '@/utils/date';
import { PhArticle, PhEyeSlash, PhImage, PhListNumbers, PhSquaresFour } from '@phosphor-icons/vue';
import { SettingGroup, SettingWithToggle, SettingWithSelect } from '@/components/settings';
import '@/components/settings/styles.css';
import type { SettingsData } from '@/types/settings';
import {
  articleTableColumns,
  parseArticleTableColumns,
  type ArticleTableColumn,
} from '@/utils/articleTable';

const { t, locale } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();
const datePreview = computed(() =>
  formatExactDateTime('2026-12-31T15:04:00', locale.value, props.settings)
);
const visibleTableColumns = computed(() =>
  parseArticleTableColumns(props.settings.article_table_columns)
);
function toggleTableColumn(column: ArticleTableColumn, enabled: boolean): void {
  const columns = new Set(visibleTableColumns.value);
  if (enabled) columns.add(column);
  else columns.delete(column);
  updateSetting(
    'article_table_columns',
    JSON.stringify(articleTableColumns.filter((item) => item === 'title' || columns.has(item)))
  );
}

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

function updateSetting(key: keyof SettingsData, value: string | number | boolean) {
  emit('update:settings', {
    ...props.settings,
    [key]: value,
  });
}
</script>

<template>
  <SettingGroup :icon="PhArticle" :title="t('setting.tab.articleDisplay')">
    <SettingWithSelect
      :icon="PhArticle"
      :title="t('setting.reading.dateFormat')"
      :description="t('setting.reading.dateFormatDesc')"
      :model-value="settings.date_format"
      :options="[
        { value: 'locale', label: t('setting.reading.localeFormat') },
        { value: 'yyyy-mm-dd', label: '2026-12-31' },
        { value: 'mm/dd/yyyy', label: '12/31/2026' },
        { value: 'dd/mm/yyyy', label: '31/12/2026' },
        { value: 'dd.mm.yyyy', label: '31.12.2026' },
      ]"
      width="md"
      @update:model-value="updateSetting('date_format', $event)"
    />
    <SettingWithSelect
      :icon="PhArticle"
      :title="t('setting.reading.timeFormat')"
      :description="t('setting.reading.dateTimePreview', { value: datePreview })"
      :model-value="settings.time_format"
      :options="[
        { value: 'locale', label: t('setting.reading.localeFormat') },
        { value: '12h', label: t('setting.reading.timeFormat12') },
        { value: '24h', label: t('setting.reading.timeFormat24') },
      ]"
      width="md"
      @update:model-value="updateSetting('time_format', $event)"
    />
    <SettingWithToggle
      :icon="PhArticle"
      :title="t('setting.reading.relativeTime')"
      :description="t('setting.reading.relativeTimeDesc')"
      :model-value="settings.relative_time"
      @update:model-value="updateSetting('relative_time', $event)"
    />
    <SettingWithSelect
      :icon="PhArticle"
      :title="t('setting.reading.defaultViewMode')"
      :description="t('setting.reading.defaultViewModeDesc')"
      :model-value="settings.default_view_mode"
      :options="[
        { value: 'original', label: t('article.action.viewModeOriginal') },
        { value: 'rendered', label: t('article.action.viewModeRendered') },
        { value: 'external', label: t('article.action.viewModeExternal') },
      ]"
      width="md"
      @update:model-value="updateSetting('default_view_mode', $event)"
    />

    <SettingWithToggle
      :icon="PhImage"
      :title="t('setting.reading.showArticlePreviewImages')"
      :description="t('setting.reading.showArticlePreviewImagesDesc')"
      :model-value="settings.show_article_preview_images"
      @update:model-value="updateSetting('show_article_preview_images', $event)"
    />

    <SettingWithToggle
      :icon="PhListNumbers"
      :title="t('setting.reading.showFloatingToc')"
      :description="t('setting.reading.showFloatingTocDesc')"
      :model-value="settings.show_floating_toc"
      @update:model-value="updateSetting('show_floating_toc', $event)"
    />

    <SettingWithToggle
      :icon="PhEyeSlash"
      :title="t('setting.reading.showUnreadCounts')"
      :description="t('setting.reading.showUnreadCountsDesc')"
      :model-value="settings.show_unread_counts"
      @update:model-value="updateSetting('show_unread_counts', $event)"
    />

    <SettingWithSelect
      :icon="PhSquaresFour"
      :title="t('setting.typography.layoutMode')"
      :description="t('setting.typography.layoutModeDesc')"
      :model-value="settings.layout_mode"
      :options="[
        { value: 'normal', label: t('setting.typography.layoutModeNormal') },
        { value: 'compact', label: t('setting.typography.layoutModeCompact') },
        { value: 'card', label: t('setting.typography.layoutModeCard') },
        { value: 'table', label: t('setting.typography.layoutModeTable') },
      ]"
      width="md"
      @update:model-value="updateSetting('layout_mode', $event)"
    />
    <template v-if="settings.layout_mode === 'table'">
      <SettingWithToggle
        v-for="column in articleTableColumns.filter((item) => item !== 'title')"
        :key="column"
        :icon="PhListNumbers"
        :title="t('article.table.showColumn', { name: t(`article.table.${column}`) })"
        :model-value="visibleTableColumns.includes(column)"
        @update:model-value="toggleTableColumn(column, $event)"
      />
    </template>
  </SettingGroup>
</template>

<style scoped></style>

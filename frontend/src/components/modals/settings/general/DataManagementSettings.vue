<script setup lang="ts">
import { ref, onMounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhDatabase,
  PhBroom,
  PhHardDrive,
  PhCalendarX,
  PhImage,
  PhTrash,
} from '@phosphor-icons/vue';
import {
  SettingGroup,
  SettingWithToggle,
  SubSettingItem,
  NumberControl,
  NestedSettingsContainer,
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

const mediaCacheSize = ref<number>(0);
const articleCacheCount = ref<number>(0);
const isCleaningCache = ref(false);
const isCleaningArticleCache = ref(false);

function updateSetting(key: keyof SettingsData, value: any) {
  emit('update:settings', {
    ...props.settings,
    [key]: value,
  });
}

// Fetch current media cache size
async function fetchMediaCacheSize() {
  try {
    const response = await fetch('/api/media/info');
    if (response.ok) {
      const data = await response.json();
      mediaCacheSize.value = data.cache_size_mb || 0;
    }
  } catch (error) {
    console.error('Failed to fetch media cache size:', error);
  }
}

// Fetch article content cache count
async function fetchArticleCacheCount() {
  try {
    const response = await fetch('/api/articles/content-cache-info');
    if (response.ok) {
      const data = await response.json();
      articleCacheCount.value = data.cached_articles || 0;
    }
  } catch (error) {
    console.error('Failed to fetch article cache count:', error);
  }
}

// Clean media cache
async function cleanMediaCache() {
  const confirmed = await window.showConfirm({
    title: t('setting.database.mediaCacheCleanup'),
    message: t('setting.database.clearMediaCacheConfirm'),
    isDanger: true,
  });
  if (!confirmed) return;

  isCleaningCache.value = true;
  try {
    const response = await fetch('/api/media/cleanup?all=true', { method: 'POST' });
    if (response.ok) {
      const data = await response.json();
      window.showToast(
        `${t('setting.database.mediaCacheCleanup')}: ${t('modal.feed.filesRemoved', { count: data.files_cleaned })}`,
        'success'
      );
      // Immediately update cache size
      await fetchMediaCacheSize();
    } else {
      window.showToast(t('common.errors.cleaningDatabase'), 'error');
    }
  } catch (error) {
    console.error('Failed to clean media cache:', error);
    window.showToast(t('common.errors.cleaningDatabase'), 'error');
  } finally {
    isCleaningCache.value = false;
  }
}

// Clean article content cache
async function cleanArticleContentCache() {
  const confirmed = await window.showConfirm({
    title: t('setting.database.articleContentCacheCleanup'),
    message: t('setting.database.clearArticleContentCacheConfirm'),
    isDanger: true,
  });
  if (!confirmed) return;

  isCleaningArticleCache.value = true;
  try {
    const response = await fetch('/api/articles/cleanup-content', { method: 'POST' });
    if (response.ok) {
      const data = await response.json();
      window.showToast(
        `${t('setting.database.articleContentCacheCleanup')}: ${t('modal.feed.articlesRemoved', { count: data.entries_cleaned })}`,
        'success'
      );
      // Immediately update cache count
      await fetchArticleCacheCount();
    } else {
      window.showToast(t('common.errors.cleaningDatabase'), 'error');
    }
  } catch (error) {
    console.error('Failed to clean article content cache:', error);
    window.showToast(t('common.errors.cleaningDatabase'), 'error');
  } finally {
    isCleaningArticleCache.value = false;
  }
}

// Fetch all cache data
async function fetchAllCacheData() {
  if (props.settings.media_cache_enabled) {
    await fetchMediaCacheSize();
  }
  await fetchArticleCacheCount();
}

onMounted(() => {
  // Fetch cache sizes when component mounts
  fetchAllCacheData();
});

// Watch for settings changes to refetch media cache info
watch(
  () => props.settings.media_cache_enabled,
  () => {
    fetchAllCacheData();
  }
);
</script>

<template>
  <SettingGroup :icon="PhDatabase" :title="t('setting.database.dataManagement')">
    <!-- Article Cleanup -->
    <SettingWithToggle
      :icon="PhBroom"
      :title="t('setting.database.autoCleanup')"
      :description="t('setting.database.autoCleanupDesc')"
      :model-value="settings.auto_cleanup_enabled"
      @update:model-value="updateSetting('auto_cleanup_enabled', $event)"
    />

    <NestedSettingsContainer v-if="settings.auto_cleanup_enabled">
      <SubSettingItem
        :icon="PhHardDrive"
        :title="t('setting.database.maxCacheSize')"
        :description="t('setting.database.maxCacheSizeDesc')"
      >
        <NumberControl
          :model-value="settings.max_cache_size_mb"
          :min="1"
          :max="1000"
          suffix="MB"
          @update:model-value="updateSetting('max_cache_size_mb', $event)"
        />
      </SubSettingItem>

      <SubSettingItem
        :icon="PhCalendarX"
        :title="t('setting.database.maxArticleAge')"
        :description="t('setting.database.maxArticleAgeDesc')"
      >
        <NumberControl
          :model-value="settings.max_article_age_days"
          :min="1"
          :max="365"
          :suffix="t('common.time.days')"
          @update:model-value="updateSetting('max_article_age_days', $event)"
        />
      </SubSettingItem>

      <SubSettingItem
        :icon="PhTrash"
        :title="t('setting.database.articleContentCacheCleanup')"
        :description="t('setting.database.articleContentCacheCleanupDesc')"
      >
        <template #extraInfo>
          <div class="text-xs text-text-secondary mt-1">
            {{ t('setting.database.currentCachedArticles') }}:
            <span class="theme-number">{{ articleCacheCount }}</span>
          </div>
        </template>
        <button
          :disabled="isCleaningArticleCache"
          class="btn-secondary"
          @click="cleanArticleContentCache"
        >
          <PhBroom :size="16" class="sm:w-5 sm:h-5" />
          {{
            isCleaningArticleCache
              ? t('setting.database.cleaning')
              : t('setting.database.cleanupArticleContentCache')
          }}
        </button>
      </SubSettingItem>
    </NestedSettingsContainer>

    <!-- Media Cache -->
    <SettingWithToggle
      :icon="PhImage"
      :title="t('setting.database.mediaCacheEnabled')"
      :description="t('setting.database.mediaCacheEnabledDesc')"
      :model-value="settings.media_cache_enabled"
      @update:model-value="updateSetting('media_cache_enabled', $event)"
    />

    <NestedSettingsContainer v-if="settings.media_cache_enabled">
      <SubSettingItem
        :icon="PhHardDrive"
        :title="t('setting.database.mediaCacheMaxSize')"
        :description="t('setting.database.mediaCacheMaxSizeDesc')"
      >
        <NumberControl
          :model-value="settings.media_cache_max_size_mb"
          :min="10"
          :max="1000"
          suffix="MB"
          @update:model-value="updateSetting('media_cache_max_size_mb', $event)"
        />
      </SubSettingItem>

      <SubSettingItem
        :icon="PhCalendarX"
        :title="t('setting.database.mediaCacheMaxAge')"
        :description="t('setting.database.mediaCacheMaxAgeDesc')"
      >
        <NumberControl
          :model-value="settings.media_cache_max_age_days"
          :min="1"
          :max="90"
          :suffix="t('setting.database.days')"
          @update:model-value="updateSetting('media_cache_max_age_days', $event)"
        />
      </SubSettingItem>

      <SubSettingItem
        :icon="PhTrash"
        :title="t('setting.database.mediaCacheCleanup')"
        :description="t('setting.database.mediaCacheCleanupDesc')"
      >
        <template #extraInfo>
          <div class="text-xs text-text-secondary mt-1">
            {{ t('setting.database.currentCacheSize') }}:
            <span class="theme-number">{{ mediaCacheSize.toFixed(2) }} MB</span>
          </div>
        </template>
        <button :disabled="isCleaningCache" class="btn-secondary" @click="cleanMediaCache">
          <PhBroom :size="16" class="sm:w-5 sm:h-5" />
          {{
            isCleaningCache
              ? t('setting.database.cleaning')
              : t('setting.database.cleanupMediaCache')
          }}
        </button>
      </SubSettingItem>
    </NestedSettingsContainer>
  </SettingGroup>
</template>

<style scoped></style>

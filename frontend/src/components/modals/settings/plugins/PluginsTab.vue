<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount } from 'vue';
import { useAppStore } from '@/stores/app';
import type { SettingsData } from '@/types/settings';
import { useSettingsAutoSave } from '@/composables/core/useSettingsAutoSave';
import { useI18n } from 'vue-i18n';
import { TipBox } from '@/components/settings';
import ObsidianSettings from './ObsidianSettings.vue';
import NotionSettings from './NotionSettings.vue';
import SiYuanSettings from './SiYuanSettings.vue';
import ZoteroSettings from './ZoteroSettings.vue';
import ReaderIntegrationSettings from './ReaderIntegrationSettings.vue';
import RSSHubSettings from './RSSHubSettings.vue';

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();
const { t } = useI18n();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

// Create a computed ref that returns the settings object
// This ensures reactivity while allowing modifications
const settingsRef = computed(() => props.settings);

// Use composable for auto-save with reactivity
useSettingsAutoSave(settingsRef);

// Handler for settings updates from child components
function handleUpdateSettings(updatedSettings: SettingsData) {
  // Emit the updated settings to parent
  emit('update:settings', updatedSettings);
}
const store = useAppStore();
function readerState() {
  return JSON.stringify(
    ['freshrss', 'miniflux'].map((provider) =>
      Object.entries(props.settings).filter(([key]) => key.startsWith(`${provider}_`))
    )
  );
}
let previousReaderState = readerState();
async function handleSavedSettings() {
  const current = readerState();
  if (current === previousReaderState) return;
  previousReaderState = current;
  await store.startFreshRSSStatusPolling();
  await Promise.all([store.fetchFeeds(), store.fetchArticles(), store.fetchUnreadCounts()]);
}
onMounted(() => window.addEventListener('settings-updated', handleSavedSettings));
onBeforeUnmount(() => window.removeEventListener('settings-updated', handleSavedSettings));
</script>

<template>
  <div class="space-y-4 sm:space-y-6">
    <TipBox type="info" :title="t('common.warning.isInDevelopment')" />

    <ObsidianSettings :settings="settings" @update:settings="handleUpdateSettings" />

    <NotionSettings :settings="settings" @update:settings="handleUpdateSettings" />

    <SiYuanSettings :settings="settings" @update:settings="handleUpdateSettings" />

    <ZoteroSettings :settings="settings" @update:settings="handleUpdateSettings" />

    <ReaderIntegrationSettings
      provider="freshrss"
      :settings="settings"
      @update:settings="handleUpdateSettings"
    />

    <ReaderIntegrationSettings
      provider="miniflux"
      :settings="settings"
      @update:settings="handleUpdateSettings"
    />

    <RSSHubSettings :settings="settings" @update:settings="handleUpdateSettings" />
  </div>
</template>

<style scoped></style>

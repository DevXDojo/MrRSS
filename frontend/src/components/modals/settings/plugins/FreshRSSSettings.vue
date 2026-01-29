<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhLink, PhUser, PhKey, PhArrowClockwise, PhCloudCheck } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';
import { useAppStore } from '@/stores/app';
import { NestedSettingsContainer, SubSettingItem, InputControl } from '@/components/settings';

const { t } = useI18n();
const appStore = useAppStore();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
  'settings-changed': [];
}>();

function updateSetting(key: keyof SettingsData, value: any) {
  emit('update:settings', {
    ...props.settings,
    [key]: value,
  });
}

const isSyncing = ref(false);
const syncStatus = ref<{
  pending_changes: number;
  failed_items: number;
  last_sync_time: string | null;
}>({
  pending_changes: 0,
  failed_items: 0,
  last_sync_time: null,
});

let statusPollInterval: ReturnType<typeof setInterval> | null = null;

// Fetch sync status
async function fetchSyncStatus() {
  try {
    const response = await fetch('/api/freshrss/status');
    if (response.ok) {
      const data = await response.json();
      syncStatus.value = data;
    }
  } catch (error) {
    console.error('Failed to fetch sync status:', error);
  }
}

// Start polling for status updates (only for UI display in settings)
function startStatusPolling() {
  fetchSyncStatus();
  statusPollInterval = setInterval(fetchSyncStatus, 5000); // Poll every 5 seconds
}

// Stop polling
function stopStatusPolling() {
  if (statusPollInterval) {
    clearInterval(statusPollInterval);
    statusPollInterval = null;
  }
}

onMounted(() => {
  if (props.settings.freshrss_enabled) {
    startStatusPolling();
  }
});

onUnmounted(() => {
  stopStatusPolling();
});

// Sync with FreshRSS server
async function syncNow() {
  isSyncing.value = true;

  try {
    const response = await fetch('/api/freshrss/sync', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (response.ok) {
      window.showToast(t('setting.freshrss.syncStarted'), 'success');
      // Sync status polling will detect completion and refresh data automatically
    } else {
      throw new Error(t('setting.freshrss.syncFailed'));
    }
  } catch (error) {
    window.showToast(
      error instanceof Error ? error.message : t('setting.freshrss.syncFailed'),
      'error'
    );
  } finally {
    isSyncing.value = false;
  }
}

// Handle FreshRSS enabled toggle with confirmation
async function handleFreshRSSToggle(event: Event) {
  const target = event.target as HTMLInputElement;
  const newEnabled = target.checked;

  // If disabling, show confirmation dialog
  if (!newEnabled && props.settings.freshrss_enabled) {
    const confirmed = await window.showConfirm({
      title: t('setting.freshrss.enabled'),
      message: t('setting.freshrss.disableConfirm'),
      isDanger: true,
    });
    if (!confirmed) {
      // Revert the checkbox
      target.checked = true;
      return;
    }
  }

  // Emit the change
  updateSetting('freshrss_enabled', newEnabled);
}

// Watch for FreshRSS enabled changes and refresh data accordingly
watch(
  () => props.settings.freshrss_enabled,
  async (newVal, oldVal) => {
    // Use boolean values directly (settings system ensures these are booleans)
    const oldBool = oldVal;
    const newBool = newVal;

    // When FreshRSS is disabled, refresh feeds and unread counts
    if (oldBool && !newBool) {
      // FreshRSS was just disabled, cleanup will happen on backend
      // Stop global polling
      appStore.stopFreshRSSStatusPolling();
      // Wait a bit for cleanup to complete, then refresh
      setTimeout(async () => {
        await appStore.fetchFeeds();
        await appStore.fetchArticles();
        await appStore.fetchUnreadCounts();
        stopStatusPolling();
      }, 1000);
    } else if (!oldBool && newBool) {
      // FreshRSS was just enabled, start global polling
      await appStore.startFreshRSSStatusPolling();
      startStatusPolling(); // Also start local polling for UI display
      emit('settings-changed');
    }
  }
);

// Format sync time
function formatSyncTime(timeStr: string | null): string {
  if (!timeStr) return t('setting.freshrss.never');
  const date = new Date(timeStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);

  if (diffMins < 1) return t('setting.freshrss.justNow');
  if (diffMins < 60) return t('setting.freshrss.minsAgo', { n: diffMins });
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return t('setting.freshrss.hoursAgo', { n: diffHours });
  const diffDays = Math.floor(diffHours / 24);
  return t('setting.freshrss.daysAgo', { n: diffDays });
}
</script>

<template>
  <!-- Enable FreshRSS Sync -->
  <div class="setting-item">
    <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
      <img
        src="/assets/plugin_icons/freshrss.svg"
        alt="FreshRSS"
        class="w-5 h-5 sm:w-6 sm:h-6 mt-0.5 shrink-0"
      />
      <div class="flex-1 min-w-0">
        <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
          {{ t('setting.freshrss.enabled') }}
        </div>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('setting.freshrss.enabledDesc') }}
        </div>
      </div>
    </div>
    <input
      type="checkbox"
      :checked="props.settings.freshrss_enabled"
      class="toggle"
      @change="handleFreshRSSToggle"
    />
  </div>
  <NestedSettingsContainer v-if="props.settings.freshrss_enabled">
    <!-- Server URL -->
    <SubSettingItem
      :icon="PhLink"
      :title="t('setting.freshrss.serverUrl')"
      :description="t('setting.freshrss.serverUrlDesc')"
      required
    >
      <InputControl
        type="url"
        :model-value="props.settings.freshrss_server_url"
        :placeholder="t('setting.freshrss.serverUrlPlaceholder')"
        width="md"
        @update:model-value="updateSetting('freshrss_server_url', $event)"
      />
    </SubSettingItem>

    <!-- Username -->
    <SubSettingItem
      :icon="PhUser"
      :title="t('setting.freshrss.username')"
      :description="t('setting.freshrss.usernameDesc')"
      required
    >
      <InputControl
        :model-value="props.settings.freshrss_username"
        :placeholder="t('setting.freshrss.usernamePlaceholder')"
        width="md"
        @update:model-value="updateSetting('freshrss_username', $event)"
      />
    </SubSettingItem>

    <!-- API Password -->
    <SubSettingItem
      :icon="PhKey"
      :title="t('setting.freshrss.apiPassword')"
      :description="t('setting.freshrss.apiPasswordDesc')"
    >
      <InputControl
        type="password"
        :model-value="props.settings.freshrss_api_password"
        :placeholder="t('setting.freshrss.apiPasswordPlaceholder')"
        width="md"
        @update:model-value="updateSetting('freshrss_api_password', $event)"
      />
    </SubSettingItem>

    <!-- Sync Button -->
    <SubSettingItem
      :icon="PhCloudCheck"
      :title="t('setting.freshrss.syncNow')"
      :description="t('setting.freshrss.syncNowDesc')"
    >
      <template #description>
        <div>
          {{ t('setting.freshrss.syncNowDesc') }}
          <div class="text-xs text-text-secondary mt-1">
            {{ t('setting.freshrss.lastSync') }}:
            <span class="theme-number">{{ formatSyncTime(syncStatus.last_sync_time) }}</span>
          </div>
        </div>
      </template>
      <button class="btn-secondary" :disabled="isSyncing" @click="syncNow">
        <PhArrowClockwise :size="16" class="sm:w-5 sm:h-5" :class="{ 'animate-spin': isSyncing }" />
        {{ isSyncing ? t('setting.freshrss.syncing') : t('setting.freshrss.sync') }}
      </button>
    </SubSettingItem>
  </NestedSettingsContainer>
</template>

<style scoped>
@reference "../../../../style.css";

.toggle {
  @apply w-10 h-5 appearance-none bg-bg-tertiary rounded-full relative cursor-pointer border border-border transition-colors checked:bg-accent checked:border-accent shrink-0;
}
.toggle::after {
  content: '';
  @apply absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full shadow-sm transition-transform;
}
.toggle:checked::after {
  transform: translateX(20px);
}

.setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border;
}

.btn-secondary {
  @apply bg-bg-tertiary border border-border text-text-primary px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium hover:bg-bg-secondary transition-colors;
}
.btn-secondary:disabled {
  @apply cursor-not-allowed opacity-50;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(360deg);
  }
}

.animate-spin {
  animation: spin 1s linear infinite;
}

.theme-number {
  @apply text-accent font-semibold;
}
</style>

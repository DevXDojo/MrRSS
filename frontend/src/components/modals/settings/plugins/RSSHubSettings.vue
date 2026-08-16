<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhLink, PhKey, PhTestTube } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';
import { useAppStore } from '@/stores/app';
import {
  NestedSettingsContainer,
  SubSettingItem,
  InputControl,
  TipBox,
} from '@/components/settings';

const { t } = useI18n();
const store = useAppStore();

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

const isTesting = ref(false);

// Check if there are any RSSHub feeds
const hasRSSHubFeeds = computed(() => {
  return store.feeds && store.feeds.some((f) => f.url.startsWith('rsshub://'));
});

// Handle RSSHub toggle - prevent disabling if there are RSSHub feeds
function handleToggleRSSHub(e: Event) {
  const target = e.target as HTMLInputElement;
  const newValue = target.checked;

  // Prevent disabling if there are RSSHub feeds
  if (!newValue && hasRSSHubFeeds.value) {
    window.showToast(t('setting.rsshub.cannotDisableWithFeeds'), 'error');
    // Reset checkbox to enabled
    target.checked = true;
    return;
  }

  updateSetting('rsshub_enabled', newValue);
}

// Test RSSHub connection
async function testConnection() {
  isTesting.value = true;

  try {
    const response = await fetch('/api/rsshub/test-connection', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        endpoint: props.settings.rsshub_endpoint,
        api_key: props.settings.rsshub_api_key,
      }),
    });

    const result = await response.json();

    if (result.success) {
      window.showToast(t('setting.rsshub.connectionSuccessful'), 'success');
    } else {
      window.showToast(result.error || t('setting.rsshub.connectionFailed'), 'error');
    }
  } catch (error) {
    window.showToast(
      error instanceof Error ? error.message : t('setting.rsshub.connectionFailed'),
      'error'
    );
  } finally {
    isTesting.value = false;
  }
}
</script>

<template>
  <!-- Enable RSSHub -->
  <div class="setting-item">
    <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
      <img
        src="/assets/plugin_icons/rsshub.svg"
        alt="RSSHub"
        class="w-5 h-5 sm:w-6 sm:h-6 mt-0.5 shrink-0"
      />
      <div class="flex-1 min-w-0">
        <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
          {{ t('setting.rsshub.enabled') }}
        </div>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('setting.rsshub.enabledDesc') }}
        </div>
      </div>
    </div>
    <input
      type="checkbox"
      :checked="props.settings.rsshub_enabled"
      class="toggle"
      @change="handleToggleRSSHub"
    />
  </div>

  <NestedSettingsContainer v-if="props.settings.rsshub_enabled">
    <TipBox type="tip" :title="t('setting.rsshub.notSuggestOfficial')" />

    <!-- Endpoint -->
    <SubSettingItem
      :icon="PhLink"
      :title="t('setting.rsshub.endpoint')"
      :description="t('setting.rsshub.endpointDesc')"
      required
    >
      <InputControl
        :model-value="props.settings.rsshub_endpoint"
        placeholder="https://rsshub.app"
        width="md"
        @update:model-value="updateSetting('rsshub_endpoint', $event)"
      />
    </SubSettingItem>

    <!-- API Key -->
    <SubSettingItem
      :icon="PhKey"
      :title="t('setting.rsshub.apiKey')"
      :description="t('setting.rsshub.apiKeyDesc')"
    >
      <InputControl
        type="password"
        :model-value="props.settings.rsshub_api_key"
        :placeholder="t('setting.rsshub.optional')"
        width="md"
        @update:model-value="updateSetting('rsshub_api_key', $event)"
      />
    </SubSettingItem>

    <!-- Test Connection -->
    <SubSettingItem
      :icon="PhTestTube"
      :title="t('setting.rsshub.testConnection')"
      :description="t('setting.rsshub.testConnectionDesc')"
    >
      <button class="btn-secondary" :disabled="isTesting" @click="testConnection">
        {{ isTesting ? t('setting.rsshub.testing') : t('setting.rsshub.testConnection') }}
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
</style>

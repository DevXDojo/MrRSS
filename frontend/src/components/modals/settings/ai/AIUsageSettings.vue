<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhChartLine, PhArrowCounterClockwise } from '@phosphor-icons/vue';
import { SettingGroup, SettingItem, StatusBoxGroup } from '@/components/settings';
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

const usageRefreshIntervalMs = 15_000;
let usageRefreshTimer: ReturnType<typeof setInterval> | null = null;
let usageRequestInFlight: Promise<void> | null = null;
let usageRefreshQueued = false;
let usageMounted = false;
let usageFetchErrorShown = false;

// AI usage tracking
const aiUsage = ref<{
  usage: number;
  limit: number;
  limit_reached: boolean;
}>({
  usage: 0,
  limit: 0,
  limit_reached: false,
});

async function fetchAIUsage(): Promise<void> {
  if (usageRequestInFlight) {
    usageRefreshQueued = true;
    return usageRequestInFlight;
  }

  usageRequestInFlight = (async () => {
    try {
      const response = await fetch('/api/ai-usage');
      if (!response.ok) {
        throw new Error(`Failed to fetch AI usage: HTTP ${response.status}`);
      }
      const result = await response.json();
      aiUsage.value = {
        usage: Number.isFinite(Number(result.usage)) ? Number(result.usage) : 0,
        limit: Number.isFinite(Number(result.limit)) ? Number(result.limit) : 0,
        limit_reached: Boolean(result.limit_reached),
      };
      usageFetchErrorShown = false;
    } catch (error) {
      console.error('Failed to fetch AI usage:', error);
      if (!usageFetchErrorShown) {
        usageFetchErrorShown = true;
        window.showToast(t('common.errors.loadingSettings'), 'error');
      }
    } finally {
      usageRequestInFlight = null;
      if (usageRefreshQueued && usageMounted) {
        usageRefreshQueued = false;
        queueMicrotask(() => void fetchAIUsage());
      }
    }
  })();

  return usageRequestInFlight;
}

async function resetAIUsage() {
  const confirmed = await window.showConfirm({
    title: t('common.confirm'),
    message: t('setting.ai.aiUsageResetConfirm'),
    isDanger: true,
  });
  if (!confirmed) return;

  try {
    const response = await fetch('/api/ai-usage/reset', { method: 'POST' });
    if (!response.ok) {
      throw new Error(`Failed to reset AI usage: HTTP ${response.status}`);
    }

    aiUsage.value.usage = 0;
    await fetchAIUsage();
    // Reset the local settings value as well
    emit('update:settings', {
      ...props.settings,
      ai_usage_tokens: '0',
    });
    window.showToast(t('setting.ai.aiUsageResetSuccess'), 'success');
  } catch (error) {
    console.error('Failed to reset AI usage:', error);
    window.showToast(t('setting.ai.aiUsageResetError'), 'error');
  }
}

const currentUsageLimit = computed(() => {
  const parsed = Number(props.settings.ai_usage_limit.trim());
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : 0;
});

const usageLimitReached = computed(
  () => currentUsageLimit.value > 0 && aiUsage.value.usage >= currentUsageLimit.value
);

// Calculate usage percentage from the current input value so the card updates
// immediately, before the debounced settings request finishes.
const usagePercentage = computed(() => {
  if (currentUsageLimit.value === 0) return 0;
  return Math.min(100, (aiUsage.value.usage / currentUsageLimit.value) * 100);
});

// Status box type based on usage
const statusType = computed(() => {
  if (currentUsageLimit.value === 0) return 'neutral';
  if (usageLimitReached.value) return 'error';
  if (usagePercentage.value > 80) return 'warning';
  return 'success';
});

// Token display value
const tokenDisplay = computed(() => {
  if (currentUsageLimit.value > 0) {
    return `${aiUsage.value.usage.toLocaleString()} / ${currentUsageLimit.value.toLocaleString()}`;
  }
  return `${aiUsage.value.usage.toLocaleString()} / ∞`;
});

function handleSettingsUpdated() {
  void fetchAIUsage();
}

onMounted(() => {
  usageMounted = true;
  void fetchAIUsage();
  window.addEventListener('settings-updated', handleSettingsUpdated);
  usageRefreshTimer = setInterval(() => void fetchAIUsage(), usageRefreshIntervalMs);
});

onUnmounted(() => {
  usageMounted = false;
  usageRefreshQueued = false;
  window.removeEventListener('settings-updated', handleSettingsUpdated);
  if (usageRefreshTimer) {
    clearInterval(usageRefreshTimer);
    usageRefreshTimer = null;
  }
});
</script>

<template>
  <SettingGroup :icon="PhChartLine" :title="t('setting.ai.aiUsage')">
    <!-- AI Usage Display -->
    <StatusBoxGroup
      data-testid="ai-usage-status"
      class="ai-usage-status-group"
      :statuses="[
        {
          label: t('setting.ai.aiUsageTokens'),
          value: tokenDisplay,
          unit: currentUsageLimit > 0 ? t('setting.ai.tokens') : '',
          type: statusType,
        },
      ]"
      :action-button="{
        label: t('setting.ai.aiUsageReset'),
        icon: PhArrowCounterClockwise,
        onClick: resetAIUsage,
      }"
      :status-info="
        currentUsageLimit > 0
          ? {
              label: t('common.text.progress'),
              time: usagePercentage.toFixed(2) + '%',
            }
          : undefined
      "
    />

    <!-- Set AI Usage Limit -->
    <SettingItem
      :icon="PhChartLine"
      :title="t('setting.ai.setUsageLimit')"
      :description="t('setting.ai.setUsageLimitDesc')"
    >
      <input
        data-testid="ai-usage-limit-input"
        :value="props.settings.ai_usage_limit"
        type="number"
        min="0"
        :placeholder="t('setting.ai.aiUsageLimitPlaceholder')"
        class="input-field w-32 sm:w-48 text-xs sm:text-sm"
        @input="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              ai_usage_limit: (e.target as HTMLInputElement).value,
            })
        "
      />
    </SettingItem>
  </SettingGroup>
</template>

<style scoped>
@reference "../../../../style.css";
.input-field {
  @apply p-1.5 sm:p-2.5 border border-border rounded-md bg-bg-secondary text-text-primary focus:border-accent focus:outline-none transition-colors;
}

.ai-usage-status-group :deep(.status-box) {
  @apply sm:min-w-[180px];
}
</style>

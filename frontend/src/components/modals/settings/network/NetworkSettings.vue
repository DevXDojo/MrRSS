<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhNetwork, PhArrowClockwise } from '@phosphor-icons/vue';
import { SettingGroup, StatusBoxGroup, TipBox } from '@/components/settings';
import '@/components/settings/styles.css';
import type { NetworkInfo } from '@/types/settings';

const { t } = useI18n();

const networkInfo = ref<NetworkInfo>({
  speed_level: 'medium',
  bandwidth_mbps: 0,
  latency_ms: 0,
  max_concurrency: 5,
  detection_time: '',
  detection_success: false,
});

const isDetecting = ref(false);
const errorMessage = ref('');

async function loadNetworkInfo() {
  try {
    const response = await fetch('/api/network/info');
    if (response.ok) {
      const data = await response.json();
      networkInfo.value = data;
    }
  } catch (error) {
    console.error('Failed to load network info:', error);
  }
}

async function detectNetwork() {
  isDetecting.value = true;
  errorMessage.value = '';

  try {
    const response = await fetch('/api/network/detect', {
      method: 'POST',
    });

    if (response.ok) {
      const data = await response.json();
      networkInfo.value = data;

      if (!data.detection_success) {
        errorMessage.value = t('setting.network.detectionFailed');
      } else {
        window.showToast(t('setting.network.detectionComplete'), 'success');
      }
    } else {
      errorMessage.value = t('setting.network.detectionFailed');
    }
  } catch (error) {
    console.error('Network detection error:', error);
    errorMessage.value = t('setting.network.detectionFailed');
  } finally {
    isDetecting.value = false;
  }
}

function formatTime(timeStr: string): string {
  if (!timeStr) return '';

  const date = new Date(timeStr);

  // Check if the date is invalid or is the Unix epoch (zero time)
  if (isNaN(date.getTime()) || date.getTime() === 0) {
    return '';
  }

  const now = new Date();
  const diff = now.getTime() - date.getTime();

  // If the date is in the future or too far in the past (more than 10 years),
  // it's likely an invalid/uninitialized timestamp
  const maxReasonableDiff = 10 * 365 * 24 * 60 * 60 * 1000; // 10 years in milliseconds
  if (diff < 0 || diff > maxReasonableDiff) {
    return '';
  }

  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) {
    return t('common.time.daysAgo', { count: days });
  } else if (hours > 0) {
    return t('common.time.hoursAgo', { count: hours });
  } else if (minutes > 0) {
    return t('common.time.minutesAgo', { count: minutes });
  } else {
    return t('common.time.justNow');
  }
}

onMounted(() => {
  loadNetworkInfo();
});
</script>

<template>
  <SettingGroup :icon="PhNetwork" :title="t('setting.network.networkSettings')">
    <div class="text-xs sm:text-sm text-text-secondary mb-3 sm:mb-4">
      {{ t('setting.network.networkSettingsDescription') }}
    </div>

    <TipBox type="tip" :title="t('setting.network.tunModeInfoTitle')">
      {{ t('setting.network.tunModeInfo') }}
    </TipBox>

    <!-- Network Status Display -->
    <StatusBoxGroup
      :statuses="[
        {
          label: t('setting.network.bandwidthLabel'),
          value: networkInfo.bandwidth_mbps.toFixed(1),
          unit: t('setting.network.bandwidthMbps'),
        },
        {
          label: t('setting.network.latencyLabel'),
          value: networkInfo.latency_ms,
          unit: t('setting.network.latencyMs'),
        },
      ]"
      :action-button="{
        label: isDetecting ? t('modal.discovery.detecting') : t('setting.network.reDetectNetwork'),
        icon: PhArrowClockwise,
        loading: isDetecting,
        onClick: detectNetwork,
      }"
      :status-info="
        networkInfo.detection_time
          ? {
              label: t('setting.network.lastDetection'),
              time: formatTime(networkInfo.detection_time),
            }
          : undefined
      "
    />

    <!-- Error Message -->
    <div
      v-if="errorMessage"
      class="bg-red-500/10 border border-red-500/30 rounded-lg p-2 sm:p-3 text-xs sm:text-sm text-red-500 mt-3"
    >
      {{ errorMessage }}
    </div>
  </SettingGroup>
</template>

<style scoped></style>

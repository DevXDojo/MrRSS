<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhArrowCircleUp, PhDownloadSimple, PhCircleNotch, PhGear } from '@phosphor-icons/vue';
import BaseModal from '@/components/common/BaseModal.vue';
import ModalFooter from '@/components/common/ModalFooter.vue';
import { openInBrowser } from '@/utils/browser';

interface UpdateInfo {
  has_update: boolean;
  current_version: string;
  latest_version: string;
  download_url?: string;
  error?: string;
}

interface Props {
  updateInfo: UpdateInfo;
  downloadingUpdate?: boolean;
  installingUpdate?: boolean;
  downloadProgress?: number;
  downloadProgressKnown?: boolean;
  downloadBytesWritten?: number;
  downloadTotalBytes?: number;
  downloadErrorCode?: string;
}

const props = withDefaults(defineProps<Props>(), {
  downloadingUpdate: false,
  installingUpdate: false,
  downloadProgress: 0,
  downloadProgressKnown: false,
  downloadBytesWritten: 0,
  downloadTotalBytes: 0,
  downloadErrorCode: '',
});

const emit = defineEmits<{
  close: [];
  update: [];
}>();

const { t } = useI18n();

function handleClose() {
  emit('close');
}

function handleUpdate() {
  emit('update');
}

// Computed button text
const updateButtonText = computed(() => {
  if (props.downloadingUpdate) {
    return t('common.action.downloading');
  } else if (props.installingUpdate) {
    return t('setting.update.installingUpdate');
  } else {
    return t('setting.update.updateNow');
  }
});

const normalizedDownloadProgress = computed(() =>
  Math.min(100, Math.max(0, props.downloadProgress))
);

function formatBytes(bytes: number): string {
  if (bytes <= 0) return '0 MB';
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

const downloadSizeLabel = computed(() => {
  if (props.downloadProgressKnown && props.downloadTotalBytes > 0) {
    return `${formatBytes(props.downloadBytesWritten)} / ${formatBytes(props.downloadTotalBytes)}`;
  }
  return formatBytes(props.downloadBytesWritten);
});

const downloadErrorMessage = computed(() => {
  switch (props.downloadErrorCode) {
    case 'download_proxy_error':
      return t('setting.update.downloadProxyError');
    case 'download_timeout':
      return t('setting.update.downloadTimeout');
    case 'download_network_error':
      return t('setting.update.downloadNetworkError');
    case 'download_server_error':
      return t('setting.update.downloadServerError');
    default:
      return t('setting.update.downloadFailedHelp');
  }
});

function openReleasePage() {
  openInBrowser('https://github.com/DevXDojo/MrRSS/releases/latest');
}

function handleKeyDown(e: KeyboardEvent) {
  if (e.key === 'Enter') {
    e.preventDefault();
    // Only trigger update if not downloading/installing and download URL is available
    if (!props.downloadingUpdate && !props.installingUpdate && props.updateInfo.download_url) {
      handleUpdate();
    }
  } else if (e.key === 'Escape') {
    e.preventDefault();
    // Only allow closing if not downloading/installing
    if (!props.downloadingUpdate && !props.installingUpdate) {
      handleClose();
    }
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown);
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown);
});
</script>

<template>
  <BaseModal size="md" :closable="true" @close="handleClose">
    <!-- Custom Header -->
    <template #header>
      <div class="flex items-center gap-3">
        <div class="bg-green-500/20 rounded-full p-2">
          <PhArrowCircleUp :size="28" class="text-green-500" />
        </div>
        <h3 class="text-lg sm:text-xl font-bold">{{ t('setting.update.updateAvailable') }}</h3>
      </div>
    </template>

    <!-- Content -->
    <div class="p-4 sm:p-6">
      <p class="text-text-secondary text-sm mb-4">
        {{ t('modal.update.newVersionAvailable', { version: updateInfo.latest_version }) }}
      </p>

      <div class="bg-bg-secondary rounded-lg p-3 sm:p-4 space-y-2 text-sm">
        <div class="flex justify-between items-center">
          <span class="text-text-secondary">{{ t('setting.update.currentVersion') }}:</span>
          <span class="font-mono font-medium">{{ updateInfo.current_version }}</span>
        </div>
        <div class="flex justify-between items-center">
          <span class="text-text-secondary">{{ t('setting.update.latestVersion') }}:</span>
          <span class="font-mono font-medium text-green-500">{{ updateInfo.latest_version }}</span>
        </div>
      </div>

      <div v-if="props.downloadingUpdate" class="mt-4" data-testid="update-download-progress">
        <div class="mb-1 flex items-center justify-between text-xs text-text-secondary">
          <span>{{ t('common.action.downloading') }}</span>
          <span>
            {{
              props.downloadProgressKnown
                ? `${Math.round(normalizedDownloadProgress)}%`
                : downloadSizeLabel
            }}
          </span>
        </div>
        <progress
          class="download-progress block h-2 w-full overflow-hidden rounded-full"
          :value="props.downloadProgressKnown ? normalizedDownloadProgress : undefined"
          max="100"
        ></progress>
        <div v-if="props.downloadProgressKnown" class="mt-1 text-right text-xs text-text-secondary">
          {{ downloadSizeLabel }}
        </div>
      </div>

      <div
        v-if="props.downloadErrorCode"
        class="mt-4 rounded-lg border border-red-500/30 bg-red-500/10 p-3 text-xs text-text-secondary"
      >
        <p>{{ downloadErrorMessage }}</p>
        <button type="button" class="mt-2 text-accent hover:underline" @click="openReleasePage">
          {{ t('setting.update.downloadManually') }}
        </button>
      </div>
      <p v-if="!updateInfo.download_url" class="text-text-secondary text-xs mt-4">
        {{ t('setting.update.noInstallerAvailable') }}
        <a
          href="https://github.com/DevXDojo/MrRSS/releases/latest"
          target="_blank"
          class="text-accent hover:underline"
        >
          {{ t('setting.about.viewOnGitHub') }}
        </a>
      </p>
    </div>

    <!-- Footer -->
    <template #footer>
      <ModalFooter
        align="right"
        :secondary-button="{
          label: t('setting.update.notNow'),
          disabled: props.downloadingUpdate || props.installingUpdate,
          onClick: handleClose,
        }"
      >
        <template v-if="props.updateInfo.download_url" #right>
          <button
            class="btn-primary"
            :disabled="props.downloadingUpdate || props.installingUpdate"
            @click="handleUpdate"
          >
            <PhCircleNotch v-if="props.downloadingUpdate" :size="20" class="animate-spin" />
            <PhGear v-else-if="props.installingUpdate" :size="20" class="animate-spin" />
            <PhDownloadSimple v-else :size="20" />
            <span>{{ updateButtonText }}</span>
          </button>
        </template>
      </ModalFooter>
    </template>
  </BaseModal>
</template>

<style scoped>
@reference "../../../style.css";
.btn-primary {
  @apply bg-accent text-white border-none px-5 py-2.5 rounded-lg cursor-pointer font-semibold hover:bg-accent-hover transition-colors flex items-center gap-2;
}
.btn-primary:disabled {
  @apply opacity-50 cursor-not-allowed;
}

.download-progress {
  appearance: none;
  border: 0;
  background: var(--bg-tertiary);
}

.download-progress::-webkit-progress-bar {
  background: var(--bg-tertiary);
}

.download-progress::-webkit-progress-value,
.download-progress::-moz-progress-bar {
  background: var(--accent-color);
}

.download-progress:indeterminate {
  animation: progress-pulse 1.2s ease-in-out infinite;
}

.animate-spin {
  animation: spin 1s linear infinite;
}
@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
@keyframes progress-pulse {
  0%,
  100% {
    opacity: 0.45;
  }
  50% {
    opacity: 1;
  }
}
</style>

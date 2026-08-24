<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ref, onMounted, type Ref } from 'vue';
import {
  PhArrowsClockwise,
  PhArrowCircleUp,
  PhCheckCircle,
  PhCircleNotch,
  PhGear,
  PhGithubLogo,
  PhDownloadSimple,
} from '@phosphor-icons/vue';
import { openInBrowser } from '@/utils/browser';

const { t } = useI18n();

interface UpdateInfo {
  has_update: boolean;
  current_version: string;
  latest_version: string;
  download_url?: string;
  error?: string;
}

interface Props {
  updateInfo?: UpdateInfo | null;
  checkingUpdates?: boolean;
  downloadingUpdate?: boolean;
  installingUpdate?: boolean;
  downloadProgress?: number;
}

withDefaults(defineProps<Props>(), {
  updateInfo: null,
  checkingUpdates: false,
  downloadingUpdate: false,
  installingUpdate: false,
  downloadProgress: 0,
});

const emit = defineEmits<{
  'check-updates': [];
  'download-install-update': [];
}>();

const appVersion: Ref<string> = ref('');
const isLoadingVersion: Ref<boolean> = ref(true);

onMounted(async () => {
  // Fetch current version from API
  try {
    const versionRes = await fetch('/api/version');
    if (versionRes.ok) {
      const versionData = await versionRes.json();
      appVersion.value = versionData.version;
    } else {
      console.error('Failed to fetch version:', versionRes.status, versionRes.statusText);
      appVersion.value = 'Unknown';
    }
  } catch (e) {
    console.error('Error fetching version:', e);
    appVersion.value = 'Unknown';
  } finally {
    isLoadingVersion.value = false;
  }
});

function handleCheckUpdates() {
  emit('check-updates');
}

function handleDownloadInstall() {
  emit('download-install-update');
}

function openGitHubRepo() {
  openInBrowser('https://github.com/DevXDojo/MrRSS');
}

function openGitHubRelease() {
  openInBrowser('https://github.com/DevXDojo/MrRSS/releases/latest');
}
</script>

<template>
  <div class="text-center py-6 sm:py-10 px-2">
    <img src="/assets/logo.svg" alt="Logo" class="h-12 sm:h-16 w-auto mb-3 sm:mb-4 mx-auto" />
    <h3 class="text-lg sm:text-xl font-bold mb-2">{{ t('appName') }}</h3>
    <p class="text-text-secondary text-xs sm:text-sm">
      {{ t('setting.about.version') }}
      <span v-if="isLoadingVersion" class="inline-block animate-pulse">Loading...</span>
      <span v-else>{{ appVersion }}</span>
    </p>

    <div class="mt-4 sm:mt-6 mb-4 sm:mb-6 flex justify-center">
      <button
        :disabled="checkingUpdates"
        class="btn btn-secondary justify-center text-sm sm:text-base"
        @click="handleCheckUpdates"
      >
        <PhArrowsClockwise
          :size="18"
          class="sm:w-5 sm:h-5"
          :class="{ 'animate-spin': checkingUpdates }"
        />
        {{ checkingUpdates ? t('common.checking') : t('setting.update.checkForUpdates') }}
      </button>
    </div>

    <div
      v-if="updateInfo && !updateInfo.error"
      class="mt-3 sm:mt-4 mx-auto max-w-md text-left bg-bg-secondary p-3 sm:p-4 rounded-lg border border-border"
    >
      <div class="flex items-start gap-2 sm:gap-3">
        <PhArrowCircleUp
          v-if="updateInfo.has_update"
          :size="28"
          class="text-green-500 mt-0.5 shrink-0 sm:w-8 sm:h-8"
        />
        <PhCheckCircle v-else :size="28" class="text-accent mt-0.5 shrink-0 sm:w-8 sm:h-8" />
        <div class="flex-1 min-w-0">
          <h4 class="font-semibold mb-1 text-sm sm:text-base">
            {{
              updateInfo.has_update
                ? t('setting.update.updateAvailable')
                : t('setting.update.upToDate')
            }}
          </h4>
          <div class="text-xs sm:text-sm text-text-secondary space-y-1">
            <div class="truncate">
              {{ t('setting.update.currentVersion') }}: {{ updateInfo.current_version }}
            </div>
            <div v-if="updateInfo.has_update" class="truncate">
              {{ t('setting.update.latestVersion') }}: {{ updateInfo.latest_version }}
            </div>
          </div>

          <!-- Download and Install Button -->
          <div v-if="updateInfo.has_update && updateInfo.download_url" class="mt-2 sm:mt-3">
            <button
              :disabled="downloadingUpdate || installingUpdate"
              class="btn btn-primary w-full justify-center text-sm sm:text-base"
              @click="handleDownloadInstall"
            >
              <PhCircleNotch
                v-if="downloadingUpdate"
                :size="18"
                class="animate-spin sm:w-5 sm:h-5"
              />
              <PhGear v-else-if="installingUpdate" :size="18" class="animate-spin sm:w-5 sm:h-5" />
              <PhDownloadSimple v-else :size="18" class="sm:w-5 sm:h-5" />
              <span v-if="downloadingUpdate">{{ t('common.action.downloading') }}</span>
              <span v-else-if="installingUpdate">{{ t('setting.update.installingUpdate') }}</span>
              <span v-else>{{ t('modal.update.downloadUpdate') }}</span>
            </button>
          </div>

          <!-- Fallback to GitHub if no download URL -->
          <div
            v-else-if="updateInfo.has_update && !updateInfo.download_url"
            class="mt-2 sm:mt-3 text-xs text-text-secondary"
          >
            <p class="mb-2">No installer available for your platform. Please download manually:</p>
            <button
              type="button"
              class="text-accent hover:underline break-all"
              @click="openGitHubRelease"
            >
              View on GitHub
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="mt-4 sm:mt-6">
      <button
        type="button"
        class="inline-flex items-center gap-1.5 sm:gap-2 text-accent hover:text-accent-hover transition-colors text-xs sm:text-sm font-medium"
        @click="openGitHubRepo"
      >
        <PhGithubLogo :size="20" class="sm:w-6 sm:h-6" />
        {{ t('setting.about.viewOnGitHub') }}
      </button>
    </div>

    <!-- Copyright information at the bottom -->
    <div class="mt-auto pt-4 text-center">
      <p class="text-text-secondary text-xs">© 2026 MrRSS. All rights reserved.</p>
      <p class="text-text-secondary text-xs">Open source and available under GPL-3.0 License.</p>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";
.btn {
  @apply px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium transition-colors;
}
.btn:disabled {
  @apply opacity-50 cursor-not-allowed;
}
.btn-primary {
  @apply bg-accent text-white border-none hover:bg-accent-hover;
}
.btn-secondary {
  @apply bg-bg-tertiary border border-border text-text-primary hover:bg-bg-secondary;
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
</style>

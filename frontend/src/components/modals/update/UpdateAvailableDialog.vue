<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhArrowCircleUp, PhDownloadSimple, PhCircleNotch, PhGear } from '@phosphor-icons/vue';
import BaseModal from '@/components/common/BaseModal.vue';
import ModalFooter from '@/components/common/ModalFooter.vue';

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
}

const props = withDefaults(defineProps<Props>(), {
  downloadingUpdate: false,
  installingUpdate: false,
  downloadProgress: 0,
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
    return `${t('common.action.downloading')} ${props.downloadProgress}%`;
  } else if (props.installingUpdate) {
    return t('setting.update.installingUpdate');
  } else {
    return t('setting.update.updateNow');
  }
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

      <p v-if="!updateInfo.download_url" class="text-text-secondary text-xs mt-4">
        {{ t('setting.update.noInstallerAvailable') }}
        <a
          href="https://github.com/WCY-dt/MrRSS/releases/latest"
          target="_blank"
          class="text-accent hover:underline"
        >
          {{ t('setting.about.viewOnGitHub') }}
        </a>
      </p>

      <!-- Progress bar -->
      <div v-if="props.downloadingUpdate" class="mt-4">
        <div class="w-full bg-bg-tertiary rounded-full h-2 overflow-hidden">
          <div
            class="bg-accent h-full transition-all duration-300"
            :style="{ width: props.downloadProgress + '%' }"
          ></div>
        </div>
      </div>
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
        :primary-button="
          props.updateInfo.download_url
            ? {
                label: updateButtonText,
                disabled: props.downloadingUpdate || props.installingUpdate,
                onClick: handleUpdate,
              }
            : undefined
        "
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

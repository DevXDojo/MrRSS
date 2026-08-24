<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useSettings } from '@/composables/core/useSettings';
import { PhPalette, PhUpload, PhTrash, PhCheck, PhBookOpen } from '@phosphor-icons/vue';
import { SettingGroup, SettingItem } from '@/components/settings';
import '@/components/settings/styles.css';
import type { SettingsData } from '@/types/settings';
import { openInBrowser } from '@/utils/browser';

const { t, locale } = useI18n();
const { fetchSettings } = useSettings();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

const uploading = ref(false);
const deleteLoading = ref(false);

// Use props settings for real-time updates (passed from parent)
const hasCustomCSS = computed(() => !!props.settings.custom_css_file);

function openDocumentation() {
  const docUrl = locale.value.startsWith('zh')
    ? 'https://github.com/DevXDojo/MrRSS/blob/main/docs/CUSTOM_CSS.zh.md'
    : 'https://github.com/DevXDojo/MrRSS/blob/main/docs/CUSTOM_CSS.md';
  openInBrowser(docUrl);
}

const handleFileUpload = async () => {
  uploading.value = true;

  try {
    const response = await fetch('/api/custom-css/upload-dialog', {
      method: 'POST',
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || 'Upload failed');
    }

    const result = await response.json();

    if (result.status === 'cancelled') {
      console.log('CSS upload cancelled by user');
      return;
    }

    if (result.status === 'success') {
      console.log('CSS upload successful:', result);
      window.showToast(t('setting.customization.cssUploaded'), 'success');

      // Reload settings from backend to update composable
      try {
        const updatedSettings = await fetchSettings();

        // Emit the updated settings to parent
        emit('update:settings', updatedSettings);
        console.log('Settings updated with custom_css_file:', updatedSettings.custom_css_file);
      } catch (settingsError) {
        console.error('Failed to reload settings:', settingsError);
        // Don't show error toast for this, since upload succeeded
      }

      // Notify ArticleBody components to reload CSS
      window.dispatchEvent(new CustomEvent('custom-css-changed'));
    } else {
      console.error('CSS upload failed:', result);
      window.showToast(result.message || t('setting.customization.cssUploadFailed'), 'error');
    }
  } catch (error) {
    console.error('CSS upload error:', error);
    window.showToast(t('setting.customization.cssUploadFailed'), 'error');
  } finally {
    uploading.value = false;
  }
};

const handleDeleteCSS = async () => {
  deleteLoading.value = true;

  try {
    console.log('Deleting custom CSS...');

    const response = await fetch('/api/custom-css/delete', {
      method: 'POST',
    });

    if (!response.ok) {
      console.error('Delete failed with status:', response.status);
      throw new Error('Delete failed');
    }

    const result = await response.json();
    console.log('Delete response:', result);

    window.showToast(t('setting.customization.cssDeleted'), 'success');

    // Reload settings from backend to update composable
    try {
      const updatedSettings = await fetchSettings();

      // Emit the updated settings to parent
      emit('update:settings', updatedSettings);
      console.log('Settings updated with custom_css_file:', updatedSettings.custom_css_file);
    } catch (settingsError) {
      console.error('Failed to reload settings:', settingsError);
    }

    // Notify ArticleBody components to reload CSS
    window.dispatchEvent(new CustomEvent('custom-css-changed'));
  } catch (error) {
    console.error('Failed to delete CSS file:', error);
    window.showToast(t('setting.customization.cssDeleteFailed'), 'error');
  } finally {
    deleteLoading.value = false;
  }
};
</script>

<template>
  <SettingGroup :icon="PhPalette" :title="t('setting.tab.customization')">
    <!-- Custom CSS Setting -->
    <SettingItem :icon="PhPalette" :title="t('setting.customization.css')">
      <template #description>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('setting.customization.cssDesc') }}
        </div>
        <div v-if="hasCustomCSS" class="flex items-center gap-1 mt-1">
          <PhCheck :size="14" class="text-green-500" />
          <span class="text-xs text-text-secondary">{{
            t('setting.customization.cssApplied')
          }}</span>
        </div>
        <!-- Documentation Link -->
        <button
          type="button"
          class="text-xs text-accent hover:underline flex items-center gap-1 mt-1"
          @click="openDocumentation"
        >
          <PhBookOpen :size="12" />
          {{ t('setting.customization.cssGuide') }}
        </button>
      </template>

      <template #action>
        <div class="flex items-center gap-2">
          <button
            v-if="!hasCustomCSS"
            class="btn-secondary"
            :disabled="uploading"
            @click="handleFileUpload"
          >
            <PhUpload v-if="!uploading" :size="16" class="sm:w-5 sm:h-5" />
            <span class="hidden sm:inline">{{
              uploading ? t('common.pagination.uploading') : t('setting.customization.cssUpload')
            }}</span>
          </button>
          <button
            v-if="hasCustomCSS"
            class="btn-danger"
            :disabled="deleteLoading"
            @click="handleDeleteCSS"
          >
            <PhTrash v-if="!deleteLoading" :size="16" class="sm:w-5 sm:h-5" />
            <span class="hidden sm:inline">{{
              deleteLoading ? t('common.pagination.deleting') : t('setting.customization.deleteCSS')
            }}</span>
          </button>
        </div>
      </template>
    </SettingItem>
  </SettingGroup>
</template>

<style scoped>
@reference "../../../../style.css";
.btn-danger {
  @apply bg-bg-tertiary border border-border text-red-500 px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium hover:bg-bg-secondary transition-colors disabled:opacity-50 disabled:cursor-not-allowed;
}
</style>

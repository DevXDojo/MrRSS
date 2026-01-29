/**
 * Composable for app update checking and installation
 */
import { ref, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { UpdateInfo, DownloadResponse, InstallResponse } from '@/types/settings';

export function useAppUpdates() {
  const { t } = useI18n();

  const updateInfo: Ref<UpdateInfo | null> = ref(null);
  const checkingUpdates = ref(false);
  const downloadingUpdate = ref(false);
  const installingUpdate = ref(false);
  const downloadProgress = ref(0);

  /**
   * Check for available updates
   * @param silent - If true, don't show toast when up to date (for startup checks)
   */
  async function checkForUpdates(silent = false) {
    checkingUpdates.value = true;
    updateInfo.value = null;

    try {
      const res = await fetch('/api/check-updates');
      if (res.ok) {
        const data = await res.json();
        updateInfo.value = data;

        if (data.error) {
          // Handle different error types with specific messages
          if (data.error === 'network_error') {
            window.showToast(t('common.errors.networkErrorCheckingUpdates'), 'error');
          } else {
            window.showToast(t('common.errors.errorCheckingUpdates'), 'error');
          }
        } else if (data.has_update) {
          window.showToast(t('setting.update.updateAvailable'), 'info');
        } else if (!silent) {
          // Only show "up to date" toast if not in silent mode
          window.showToast(t('setting.update.upToDate'), 'success');
        }
      } else {
        window.showToast(t('common.errors.errorCheckingUpdates'), 'error');
      }
    } catch (e) {
      console.error('Error checking updates:', e);
      window.showToast(t('common.errors.errorCheckingUpdates'), 'error');
    } finally {
      checkingUpdates.value = false;
    }
  }

  /**
   * Download and install update
   */
  async function downloadAndInstallUpdate() {
    if (!updateInfo.value || !updateInfo.value.download_url) {
      window.showToast(t('common.errors.errorCheckingUpdates'), 'error');
      return;
    }

    downloadingUpdate.value = true;
    downloadProgress.value = 0;

    // Simulate progress while downloading
    const progressInterval = setInterval(() => {
      if (downloadProgress.value < 90) {
        downloadProgress.value += 10;
      }
    }, 500);

    try {
      // Download the update
      const downloadRes = await fetch('/api/download-update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          download_url: updateInfo.value.download_url,
          asset_name: updateInfo.value.asset_name,
        }),
      });

      clearInterval(progressInterval);

      if (!downloadRes.ok) {
        const errorText = await downloadRes.text();
        console.error('Download error:', errorText);
        throw new Error('DOWNLOAD_ERROR: ' + errorText);
      }

      const downloadData = (await downloadRes.json()) as DownloadResponse;
      if (!downloadData.success || !downloadData.file_path) {
        throw new Error('DOWNLOAD_ERROR: Invalid response from server');
      }

      downloadingUpdate.value = false;
      downloadProgress.value = 100;

      // Show notification
      window.showToast(t('common.toast.downloadComplete'), 'success');

      // Wait a moment to ensure file is fully written
      await new Promise((resolve) => setTimeout(resolve, 500));

      // Install the update
      installingUpdate.value = true;
      window.showToast(t('setting.update.installingUpdate'), 'info');

      const installRes = await fetch('/api/install-update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          file_path: downloadData.file_path,
        }),
      });

      if (!installRes.ok) {
        const errorText = await installRes.text();
        console.error('Install error:', errorText);
        throw new Error('INSTALL_ERROR: ' + errorText);
      }

      const installData = (await installRes.json()) as InstallResponse;
      if (!installData.success) {
        throw new Error('INSTALL_ERROR: Installation failed');
      }

      // Show final message - app will close automatically from backend
      window.showToast(t('setting.update.updateWillRestart'), 'info');
    } catch (e) {
      console.error('Update error:', e);
      clearInterval(progressInterval);
      downloadingUpdate.value = false;
      installingUpdate.value = false;

      // Use error codes for more reliable error classification
      const errorMessage = (e as Error).message || '';
      if (errorMessage.includes('DOWNLOAD_ERROR')) {
        window.showToast(t('common.toast.downloadFailed'), 'error');
      } else if (errorMessage.includes('INSTALL_ERROR')) {
        window.showToast(t('setting.update.installFailed'), 'error');
      } else {
        window.showToast(t('common.errors.errorCheckingUpdates'), 'error');
      }
    }
  }

  return {
    updateInfo,
    checkingUpdates,
    downloadingUpdate,
    installingUpdate,
    downloadProgress,
    checkForUpdates,
    downloadAndInstallUpdate,
  };
}

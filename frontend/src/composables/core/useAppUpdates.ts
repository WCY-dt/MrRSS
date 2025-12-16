/**
 * Composable for app update checking and installation
 */
import { ref, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { UpdateInfo, DownloadResponse, InstallResponse } from '@/types/settings';

// TEMPORARY WORKAROUND: Helper to get optimized interval for macOS Sequoia
let platformInfoCache: { needsUIThrottling: boolean } | null = null;
async function getPlatformInfo() {
  if (platformInfoCache) {
    return platformInfoCache;
  }
  try {
    const response = await fetch('/api/platform/info');
    if (response.ok) {
      const data = await response.json();
      platformInfoCache = { needsUIThrottling: data.needs_ui_throttling || false };
      return platformInfoCache;
    }
  } catch (error) {
    console.warn('Failed to fetch platform info:', error);
  }
  platformInfoCache = { needsUIThrottling: false };
  return platformInfoCache;
}

function getOptimizedInterval(baseInterval: number): number {
  if (platformInfoCache?.needsUIThrottling) {
    return Math.max(baseInterval * 2, 1000);
  }
  return baseInterval;
}

export function useAppUpdates() {
  const { t } = useI18n();

  const updateInfo: Ref<UpdateInfo | null> = ref(null);
  const checkingUpdates = ref(false);
  const downloadingUpdate = ref(false);
  const installingUpdate = ref(false);
  const downloadProgress = ref(0);

  /**
   * Check for available updates
   */
  async function checkForUpdates() {
    checkingUpdates.value = true;
    updateInfo.value = null;

    try {
      const res = await fetch('/api/check-updates');
      if (res.ok) {
        const data = await res.json();
        updateInfo.value = data;

        if (data.error) {
          window.showToast(t('errorCheckingUpdates'), 'error');
        } else if (data.has_update) {
          window.showToast(t('updateAvailable'), 'info');
        } else {
          window.showToast(t('upToDate'), 'success');
        }
      } else {
        window.showToast(t('errorCheckingUpdates'), 'error');
      }
    } catch (e) {
      console.error('Error checking updates:', e);
      window.showToast(t('errorCheckingUpdates'), 'error');
    } finally {
      checkingUpdates.value = false;
    }
  }

  /**
   * Download and install update
   */
  async function downloadAndInstallUpdate() {
    if (!updateInfo.value || !updateInfo.value.download_url) {
      window.showToast(t('errorCheckingUpdates'), 'error');
      return;
    }

    downloadingUpdate.value = true;
    downloadProgress.value = 0;

    // TEMPORARY WORKAROUND: Initialize platform info for macOS Sequoia throttling
    await getPlatformInfo();

    // Simulate progress while downloading
    // TEMPORARY WORKAROUND: Use optimized interval for macOS Sequoia
    const progressIntervalMs = getOptimizedInterval(500);
    const progressInterval = setInterval(() => {
      if (downloadProgress.value < 90) {
        downloadProgress.value += 10;
      }
    }, progressIntervalMs);

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
      window.showToast(t('downloadComplete'), 'success');

      // Wait a moment to ensure file is fully written
      await new Promise((resolve) => setTimeout(resolve, 500));

      // Install the update
      installingUpdate.value = true;
      window.showToast(t('installingUpdate'), 'info');

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
      window.showToast(t('updateWillRestart'), 'info');
    } catch (e) {
      console.error('Update error:', e);
      clearInterval(progressInterval);
      downloadingUpdate.value = false;
      installingUpdate.value = false;

      // Use error codes for more reliable error classification
      const errorMessage = (e as Error).message || '';
      if (errorMessage.includes('DOWNLOAD_ERROR')) {
        window.showToast(t('downloadFailed'), 'error');
      } else if (errorMessage.includes('INSTALL_ERROR')) {
        window.showToast(t('installFailed'), 'error');
      } else {
        window.showToast(t('errorCheckingUpdates'), 'error');
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

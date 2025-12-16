import { ref, onMounted } from 'vue';
import { Environment } from '@/wailsjs/wailsjs/runtime/runtime';

const isMacOS = ref(false);
const isWindows = ref(false);
const isLinux = ref(false);
const platformDetected = ref(false);
const isMacOSSequoia = ref(false);
const needsUIThrottling = ref(false);
const macOSVersion = ref('');

export function usePlatform() {
  onMounted(async () => {
    if (platformDetected.value) {
      return; // Already detected
    }

    try {
      // Get basic platform info from Wails
      const env = await Environment();
      isMacOS.value = env.platform === 'darwin';
      isWindows.value = env.platform === 'windows';
      isLinux.value = env.platform === 'linux';

      // Get detailed platform info from our API
      try {
        const response = await fetch('/api/platform/info');
        if (response.ok) {
          const data = await response.json();
          isMacOSSequoia.value = data.is_macos_sequoia || false;
          needsUIThrottling.value = data.needs_ui_throttling || false;
          macOSVersion.value = data.macos_version || '';

          if (needsUIThrottling.value) {
            console.log(
              `Detected macOS ${macOSVersion.value} - UI throttling enabled for WKWebView compatibility`
            );
          }
        }
      } catch (error) {
        console.warn('Failed to fetch detailed platform info:', error);
      }

      platformDetected.value = true;
    } catch (error) {
      console.error('Failed to detect platform:', error);
      // Fallback to user agent detection
      const ua = navigator.userAgent.toLowerCase();
      isMacOS.value = ua.includes('mac');
      isWindows.value = ua.includes('win');
      isLinux.value = ua.includes('linux');
      platformDetected.value = true;
    }
  });

  return {
    isMacOS,
    isWindows,
    isLinux,
    platformDetected,
    isMacOSSequoia,
    needsUIThrottling,
    macOSVersion,
  };
}

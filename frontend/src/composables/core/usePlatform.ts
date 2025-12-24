import { ref, onMounted } from 'vue';
import { System } from '@wailsio/runtime';

const isMacOS = ref(false);
const isWindows = ref(false);
const isLinux = ref(false);
const platformDetected = ref(false);

export function usePlatform() {
  onMounted(async () => {
    if (platformDetected.value) {
      return; // Already detected
    }

    // Wait for Wails runtime to be available
    const checkWailsRuntime = (): Promise<void> => {
      return new Promise((resolve) => {
        // Check if Wails runtime is available
        if (typeof window !== 'undefined' && (window as any)._wails) {
          resolve();
          return;
        }

        // Wait a bit and check again
        let attempts = 0;
        const maxAttempts = 50; // 5 seconds max wait
        const checkInterval = setInterval(() => {
          attempts++;
          if (typeof window !== 'undefined' && (window as any)._wails) {
            clearInterval(checkInterval);
            resolve();
          } else if (attempts >= maxAttempts) {
            clearInterval(checkInterval);
            resolve(); // Resolve anyway to avoid blocking
          }
        }, 100);
      });
    };

    try {
      // Wait for Wails runtime before using System API
      await checkWailsRuntime();

      // Check if System is available
      if (System && System.Environment) {
        const env = await System.Environment();
        isMacOS.value = env.OS === 'darwin';
        isWindows.value = env.OS === 'windows';
        isLinux.value = env.OS === 'linux';
        platformDetected.value = true;
      } else {
        throw new Error('System API not available');
      }
    } catch (error) {
      console.warn('Failed to detect platform via Wails API, using fallback:', error);
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
  };
}

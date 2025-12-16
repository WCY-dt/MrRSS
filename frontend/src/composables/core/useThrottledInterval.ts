import { onUnmounted } from 'vue';
import { usePlatform } from './usePlatform';

/**
 * Helper composable to provide platform-aware interval timing
 * On macOS 15+ (Sequoia), intervals are throttled to prevent WKWebView crashes
 * caused by excessive UI updates.
 *
 * This is a TEMPORARY WORKAROUND for macOS Sequoia WKWebView issues.
 * Once Apple fixes the WKWebView refresh rate issue, this can be removed.
 *
 * NOTE: This utility exists for consistency but is NOT currently used.
 * The throttling logic is intentionally duplicated inline in each file
 * (app.ts, useWindowState.ts, useFeedDiscovery.ts, etc.) to make it
 * easier to locate and remove when the workaround is no longer needed.
 * All duplicated code is marked with "TEMPORARY WORKAROUND" comments.
 *
 * If this workaround becomes permanent, consider refactoring to use this
 * centralized utility instead of the duplicated code.
 */
export function useThrottledInterval() {
  const { needsUIThrottling } = usePlatform();

  /**
   * Get an optimized interval duration for polling operations
   * @param baseInterval - The desired interval in milliseconds
   * @returns Adjusted interval based on platform
   */
  function getOptimizedInterval(baseInterval: number): number {
    // TEMPORARY WORKAROUND: On macOS 15+, increase polling intervals
    // to reduce UI refresh rate and prevent WKWebView crashes
    if (needsUIThrottling.value) {
      // Increase intervals by 2x for macOS Sequoia
      return Math.max(baseInterval * 2, 1000); // Minimum 1 second
    }
    return baseInterval;
  }

  /**
   * Create a throttled interval that automatically clears on unmount
   * @param callback - Function to call on each interval
   * @param baseInterval - Base interval in milliseconds
   * @returns Cleanup function
   */
  function createThrottledInterval(callback: () => void, baseInterval: number) {
    const interval = getOptimizedInterval(baseInterval);
    const intervalId = setInterval(callback, interval);

    // Auto-cleanup on unmount
    onUnmounted(() => {
      clearInterval(intervalId);
    });

    return () => clearInterval(intervalId);
  }

  return {
    getOptimizedInterval,
    createThrottledInterval,
    needsUIThrottling,
  };
}

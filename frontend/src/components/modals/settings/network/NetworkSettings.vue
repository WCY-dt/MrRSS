<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhNetwork, PhArrowClockwise } from '@phosphor-icons/vue';
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
        errorMessage.value = t('networkDetectionFailed');
      } else {
        window.showToast(t('networkDetectionComplete'), 'success');
      }
    } else {
      errorMessage.value = t('networkDetectionFailed');
    }
  } catch (error) {
    console.error('Network detection error:', error);
    errorMessage.value = t('networkDetectionFailed');
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
    return t('daysAgo', { count: days });
  } else if (hours > 0) {
    return t('hoursAgo', { count: hours });
  } else if (minutes > 0) {
    return t('minutesAgo', { count: minutes });
  } else {
    return t('justNow');
  }
}

onMounted(() => {
  loadNetworkInfo();
});
</script>

<template>
  <div class="setting-group">
    <label
      class="font-semibold mb-2 sm:mb-3 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2"
    >
      <PhNetwork :size="14" class="sm:w-4 sm:h-4" />
      {{ t('networkSettings') }}
    </label>

    <div class="text-xs sm:text-sm text-text-secondary mb-3 sm:mb-4">
      {{ t('networkSettingsDescription') }}
    </div>

    <div class="tip-box">
      <PhInfo :size="16" class="text-accent shrink-0 sm:w-5 sm:h-5" />
      <span class="text-xs sm:text-sm">{{ t('tunModeInfo') }}</span>
    </div>

    <!-- Network Status Display -->
    <div
      class="flex flex-col sm:flex-row sm:items-stretch sm:justify-between gap-3 sm:gap-4 p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border mt-2 sm:mt-3"
    >
      <!-- Top/Bottom: Speed and Latency Boxes -->
      <div class="flex flex-col sm:flex-row items-center gap-3 sm:gap-4">
        <!-- Bandwidth Box -->
        <div
          class="flex flex-col gap-2 p-3 rounded-lg bg-bg-primary border border-border w-full sm:min-w-[120px]"
        >
          <span class="text-sm text-text-secondary text-left">{{ t('bandwidthLabel') }}</span>
          <div class="flex items-baseline gap-1">
            <span class="text-xl sm:text-2xl font-bold text-text-primary">{{
              networkInfo.bandwidth_mbps.toFixed(1)
            }}</span>
            <span class="text-sm text-text-secondary">{{ t('bandwidthMbps') }}</span>
          </div>
        </div>

        <!-- Latency Box -->
        <div
          class="flex flex-col gap-2 p-3 rounded-lg bg-bg-primary border border-border w-full sm:min-w-[120px]"
        >
          <span class="text-sm text-text-secondary text-left">{{ t('latencyLabel') }}</span>
          <div class="flex items-baseline gap-1">
            <span class="text-xl sm:text-2xl font-bold text-text-primary">{{
              networkInfo.latency_ms
            }}</span>
            <span class="text-sm text-text-secondary">{{ t('latencyMs') }}</span>
          </div>
        </div>
      </div>

      <!-- Bottom/Right: Button and Detection Time -->
      <div class="flex flex-col sm:justify-between flex-1 gap-2 sm:gap-0">
        <div class="flex justify-center sm:justify-end">
          <button class="btn-secondary" :disabled="isDetecting" @click="detectNetwork">
            <PhArrowClockwise
              :size="16"
              :class="{ 'animate-spin': isDetecting, 'sm:w-5 sm:h-5': true }"
            />
            <span>{{ isDetecting ? t('detecting') : t('reDetectNetwork') }}</span>
          </button>
        </div>

        <div
          v-if="networkInfo.detection_time"
          class="flex items-center justify-center sm:justify-end gap-2"
        >
          <span class="text-xs text-text-secondary">{{ t('lastDetection') }}:</span>
          <span class="text-xs text-accent font-medium">{{
            formatTime(networkInfo.detection_time)
          }}</span>
        </div>
      </div>
    </div>

    <!-- Error Message -->
    <div
      v-if="errorMessage"
      class="bg-red-500/10 border border-red-500/30 rounded-lg p-2 sm:p-3 text-xs sm:text-sm text-red-500 mt-3"
    >
      {{ errorMessage }}
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";

.btn-secondary {
  @apply bg-bg-tertiary border border-border text-text-primary px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium hover:bg-bg-secondary transition-colors;
}

.btn-secondary:disabled {
  @apply cursor-not-allowed opacity-50;
}

.setting-group {
  @apply mb-4 sm:mb-6;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.animate-spin {
  animation: spin 1s linear infinite;
}

.tip-box {
  @apply flex items-center gap-2 sm:gap-3 py-2 sm:py-2.5 px-2.5 sm:px-3 rounded-lg;
  background-color: rgba(59, 130, 246, 0.05);
  border: 1px solid rgba(59, 130, 246, 0.3);
}
</style>

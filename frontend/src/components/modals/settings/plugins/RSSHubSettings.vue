<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PhInfo,
  PhGlobe,
  PhKey,
  PhRewind,
  PhCheckCircle,
  PhArrowClockwise,
  PhWarningCircle,
  PhBookOpen as PhGuide,
} from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

interface RSSHubTestInfo {
  connection_success: boolean;
  response_time_ms: number;
  status_code: number;
  test_time: string;
  error?: string;
}

const testInfo = ref<RSSHubTestInfo>({
  connection_success: false,
  response_time_ms: 0,
  status_code: 0,
  test_time: '',
});

const isTesting = ref(false);
const errorMessage = ref('');

// Test RSSHub connection
async function testConnection() {
  isTesting.value = true;
  errorMessage.value = '';

  try {
    const response = await fetch('/api/rsshub/config/test', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        instance_url: props.settings.rsshub_instance_url,
        api_key: props.settings.rsshub_api_key,
      }),
    });

    const data = await response.json();

    // Update test info
    testInfo.value = {
      connection_success: data.success || false,
      response_time_ms: data.response_time_ms || 0,
      status_code: data.status_code || 0,
      test_time: new Date().toISOString(),
    };

    if (response.ok && data.success) {
      window.showToast?.(t('rsshubTestSuccess'), 'success');
    } else {
      errorMessage.value = data.error || t('rsshubTestFailed');
      window.showToast?.(errorMessage.value, 'error');
    }
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : t('rsshubTestFailed');
    window.showToast?.(errorMessage.value, 'error');
  } finally {
    isTesting.value = false;
  }
}

function formatTime(timeStr: string): string {
  if (!timeStr) return '';
  const date = new Date(timeStr);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
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
</script>

<template>
  <!-- Enable RSSHub Integration -->
  <div class="setting-item">
    <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
      <img
        src="/assets/plugin_icons/rsshub.svg"
        alt="RSSHub"
        class="w-5 h-5 sm:w-6 sm:h-6 mt-0.5 shrink-0"
        @error="($event.target as HTMLImageElement).style.display = 'none'"
      />
      <div class="flex-1 min-w-0">
        <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
          {{ t('rsshubEnabled') }}
        </div>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('rsshubEnabledDesc') }}
        </div>
      </div>
    </div>
    <input
      type="checkbox"
      :checked="props.settings.rsshub_enabled"
      class="toggle"
      @change="
        (e) =>
          emit('update:settings', {
            ...props.settings,
            rsshub_enabled: (e.target as HTMLInputElement).checked,
          })
      "
    />
  </div>

  <div
    v-if="props.settings.rsshub_enabled"
    class="ml-2 sm:ml-4 space-y-2 sm:space-y-3 border-l-2 border-border pl-2 sm:pl-4"
  >
    <!-- Usage Hint -->
    <div class="tip-box">
      <PhInfo :size="16" class="text-accent shrink-0 sm:w-5 sm:h-5" />
      <span class="text-xs sm:text-sm">{{ t('rsshubUsageHint') }}</span>
    </div>
    <!-- Instance URL -->
    <div class="sub-setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhGlobe :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('rsshubInstanceUrl') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('rsshubInstanceUrlDesc') }}
          </div>
        </div>
      </div>
      <input
        type="url"
        :value="props.settings.rsshub_instance_url"
        :placeholder="t('rsshubInstanceUrlPlaceholder')"
        class="input-field w-32 sm:w-48 text-xs sm:text-sm"
        @input="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              rsshub_instance_url: (e.target as HTMLInputElement).value,
            })
        "
      />
    </div>

    <!-- API Key -->
    <div class="sub-setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhKey :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('rsshubApiKey') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('rsshubApiKeyDesc') }}
          </div>
        </div>
      </div>
      <input
        type="password"
        :value="props.settings.rsshub_api_key"
        :placeholder="t('rsshubApiKeyPlaceholder')"
        class="input-field w-32 sm:w-48 text-xs sm:text-sm"
        @input="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              rsshub_api_key: (e.target as HTMLInputElement).value,
            })
        "
      />
    </div>

    <!-- Fallback Enabled -->
    <div class="sub-setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhRewind :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('rsshubFallbackEnabled') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('rsshubFallbackEnabledDesc') }}
          </div>
        </div>
      </div>
      <input
        type="checkbox"
        :checked="props.settings.rsshub_fallback_enabled"
        class="toggle"
        @change="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              rsshub_fallback_enabled: (e.target as HTMLInputElement).checked,
            })
        "
      />
    </div>

    <!-- RSSHub Test Status Display -->
    <div
      class="flex flex-col sm:flex-row sm:items-stretch sm:justify-between gap-3 sm:gap-4 p-2 sm:p-2.5 rounded-md bg-bg-tertiary"
    >
      <!-- Status Indicators -->
      <div class="flex flex-col sm:flex-row items-center gap-3 sm:gap-4">
        <!-- Connection Success Box -->
        <div
          class="flex flex-col gap-2 p-3 rounded-lg bg-bg-primary border border-border min-w-[100px] sm:min-w-[120px]"
          :class="{
            'border-green-500/30': testInfo.connection_success,
            'border-red-500/30': testInfo.test_time && !testInfo.connection_success,
          }"
        >
          <span class="text-xs sm:text-sm text-text-secondary text-left">{{
            t('connectionSuccess')
          }}</span>
          <div class="flex items-center gap-2">
            <PhCheckCircle
              v-if="testInfo.connection_success"
              :size="20"
              class="text-green-500 shrink-0"
            />
            <PhWarningCircle
              v-else-if="testInfo.test_time"
              :size="20"
              class="text-red-500 shrink-0"
            />
            <span
              class="text-lg sm:text-2xl font-bold truncate"
              :class="{
                'text-green-500': testInfo.connection_success,
                'text-red-500': testInfo.test_time && !testInfo.connection_success,
                'text-text-primary': !testInfo.test_time,
              }"
            >
              {{ testInfo.test_time ? (testInfo.connection_success ? t('yes') : t('no')) : '-' }}
            </span>
          </div>
        </div>

        <!-- Response Time Box -->
        <div
          class="flex flex-col gap-2 p-3 rounded-lg bg-bg-primary border border-border min-w-[100px] sm:min-w-[120px]"
        >
          <span class="text-xs sm:text-sm text-text-secondary text-left">{{
            t('responseTime')
          }}</span>
          <div class="flex items-baseline gap-1">
            <span class="text-lg sm:text-2xl font-bold text-text-primary truncate">{{
              testInfo.response_time_ms > 0 ? testInfo.response_time_ms : '-'
            }}</span>
            <span class="text-xs sm:text-sm text-text-secondary shrink-0">{{
              t('latencyMs')
            }}</span>
          </div>
        </div>

        <!-- Status Code Box -->
        <div
          class="flex flex-col gap-2 p-3 rounded-lg bg-bg-primary border border-border min-w-[100px] sm:min-w-[120px]"
        >
          <span class="text-xs sm:text-sm text-text-secondary text-left">{{
            t('statusCode')
          }}</span>
          <div class="flex items-baseline gap-1">
            <span class="text-lg sm:text-2xl font-bold text-text-primary truncate">{{
              testInfo.status_code > 0 ? testInfo.status_code : '-'
            }}</span>
          </div>
        </div>
      </div>

      <!-- Right: Test Button and Test Time -->
      <div class="flex flex-col sm:justify-between flex-1 gap-2 sm:gap-0">
        <div class="flex justify-center sm:justify-end">
          <button class="btn-primary" :disabled="isTesting" @click="testConnection">
            <PhArrowClockwise
              :size="16"
              :class="{ 'animate-spin': isTesting, 'sm:w-5 sm:h-5': true }"
            />
            <span>{{ isTesting ? t('testing') : t('testConnection') }}</span>
          </button>
        </div>

        <div
          v-if="testInfo.test_time"
          class="flex items-center justify-center sm:justify-end gap-2"
        >
          <span class="text-xs text-text-secondary">{{ t('lastTest') }}:</span>
          <span class="text-xs text-accent font-medium">{{ formatTime(testInfo.test_time) }}</span>
        </div>
      </div>
    </div>

    <!-- Error Message -->
    <div
      v-if="errorMessage"
      class="bg-red-500/10 border border-red-500/30 rounded-lg p-2 sm:p-3 text-xs sm:text-sm text-red-500"
    >
      {{ errorMessage }}
    </div>

    <!-- Success Message -->
    <div
      v-if="testInfo.connection_success && !errorMessage"
      class="bg-green-500/10 border border-green-500/30 rounded-lg p-2 sm:p-3 text-xs sm:text-sm text-green-500"
    >
      {{ t('rsshubTestSuccess') }}
    </div>

    <!-- Documentation Link -->
    <div class="mt-3">
      <a
        href="https://docs.rsshub.app"
        target="_blank"
        rel="noopener noreferrer"
        class="text-xs sm:text-sm text-accent hover:underline flex items-center gap-1"
      >
        <PhGuide :size="14" />
        {{ t('rsshubDocumentation') }}
      </a>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";

.input-field {
  @apply p-1.5 sm:p-2.5 border border-border rounded-md bg-bg-secondary text-text-primary focus:border-accent focus:outline-none transition-colors;
}
.toggle {
  @apply w-10 h-5 appearance-none bg-bg-tertiary rounded-full relative cursor-pointer border border-border transition-colors checked:bg-accent checked:border-accent shrink-0;
}
.toggle::after {
  content: '';
  @apply absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full shadow-sm transition-transform;
}
.toggle:checked::after {
  transform: translateX(20px);
}
.setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border;
}
.sub-setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-2.5 rounded-md bg-bg-tertiary;
}
.btn-primary {
  @apply bg-accent text-white border-none px-3 py-2 sm:px-4 sm:py-2.5 rounded-lg cursor-pointer flex items-center gap-1 sm:gap-2 font-medium hover:bg-accent-hover transition-colors text-sm sm:text-base;
}
.btn-primary:disabled {
  @apply opacity-50 cursor-not-allowed;
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

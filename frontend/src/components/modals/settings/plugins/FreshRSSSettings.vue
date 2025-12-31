<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhLink, PhUser, PhKey, PhArrowClockwise, PhCloudCheck } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';
import { useAppStore } from '@/stores/app';

const { t } = useI18n();
const appStore = useAppStore();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

const isSyncing = ref(false);
const syncStatus = ref<{
  pending_changes: number;
  failed_items: number;
  last_sync_time: string | null;
}>({
  pending_changes: 0,
  failed_items: 0,
  last_sync_time: null,
});

let statusPollInterval: ReturnType<typeof setInterval> | null = null;

// Fetch sync status
async function fetchSyncStatus() {
  try {
    const response = await fetch('/api/freshrss/status');
    if (response.ok) {
      const data = await response.json();
      syncStatus.value = data;
    }
  } catch (error) {
    console.error('Failed to fetch sync status:', error);
  }
}

// Start polling for status updates
function startStatusPolling() {
  fetchSyncStatus();
  statusPollInterval = setInterval(fetchSyncStatus, 5000); // Poll every 5 seconds
}

// Stop polling
function stopStatusPolling() {
  if (statusPollInterval) {
    clearInterval(statusPollInterval);
    statusPollInterval = null;
  }
}

onMounted(() => {
  if (props.settings.freshrss_enabled) {
    startStatusPolling();
  }
});

onUnmounted(() => {
  stopStatusPolling();
});

// Sync with FreshRSS server
async function syncNow() {
  isSyncing.value = true;

  try {
    const response = await fetch('/api/freshrss/sync', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (response.ok) {
      window.showToast(t('freshrssSyncCompleted'), 'success');

      // Refresh feeds and articles after sync completes
      // Wait a bit for the backend sync to complete (it runs in background)
      setTimeout(async () => {
        await appStore.fetchFeeds();
        await appStore.fetchArticles();
        await appStore.fetchUnreadCounts();
        fetchSyncStatus();
      }, 2000); // Wait 2 seconds for sync to complete
    } else {
      throw new Error(t('freshrssSyncFailed'));
    }
  } catch (error) {
    window.showToast(error instanceof Error ? error.message : t('freshrssSyncFailed'), 'error');
  } finally {
    isSyncing.value = false;
  }
}

// Format sync time
function formatSyncTime(timeStr: string | null): string {
  if (!timeStr) return t('freshrssNever');
  const date = new Date(timeStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);

  if (diffMins < 1) return t('freshrssJustNow');
  if (diffMins < 60) return t('freshrssMinsAgo', { n: diffMins });
  const diffHours = Math.floor(diffMins / 60);
  if (diffHours < 24) return t('freshrssHoursAgo', { n: diffHours });
  const diffDays = Math.floor(diffHours / 24);
  return t('freshrssDaysAgo', { n: diffDays });
}
</script>

<template>
  <!-- Enable FreshRSS Sync -->
  <div class="setting-item">
    <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
      <img
        src="/assets/plugin_icons/freshrss.svg"
        alt="FreshRSS"
        class="w-5 h-5 sm:w-6 sm:h-6 mt-0.5 shrink-0"
      />
      <div class="flex-1 min-w-0">
        <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
          {{ t('freshrssEnabled') }}
        </div>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('freshrssEnabledDesc') }}
        </div>
      </div>
    </div>
    <input
      type="checkbox"
      :checked="props.settings.freshrss_enabled"
      class="toggle"
      @change="
        (e) =>
          emit('update:settings', {
            ...props.settings,
            freshrss_enabled: (e.target as HTMLInputElement).checked,
          })
      "
    />
  </div>
  <div
    v-if="props.settings.freshrss_enabled"
    class="ml-2 sm:ml-4 space-y-2 sm:space-y-3 border-l-2 border-border pl-2 sm:pl-4"
  >
    <!-- Server URL -->
    <div class="sub-setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhLink :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('freshrssServerUrl') }} <span class="text-red-500">*</span>
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('freshrssServerUrlDesc') }}
          </div>
        </div>
      </div>
      <input
        type="url"
        :value="props.settings.freshrss_server_url"
        :placeholder="t('freshrssServerUrlPlaceholder')"
        class="input-field w-32 sm:w-48 text-xs sm:text-sm"
        @input="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              freshrss_server_url: (e.target as HTMLInputElement).value,
            })
        "
      />
    </div>

    <!-- Username -->
    <div class="sub-setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhUser :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('freshrssUsername') }} <span class="text-red-500">*</span>
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('freshrssUsernameDesc') }}
          </div>
        </div>
      </div>
      <input
        type="text"
        :value="props.settings.freshrss_username"
        :placeholder="t('freshrssUsernamePlaceholder')"
        class="input-field w-32 sm:w-48 text-xs sm:text-sm"
        @input="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              freshrss_username: (e.target as HTMLInputElement).value,
            })
        "
      />
    </div>

    <!-- API Password -->
    <div class="sub-setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhKey :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('freshrssApiPassword') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('freshrssApiPasswordDesc') }}
          </div>
        </div>
      </div>
      <input
        type="password"
        :value="props.settings.freshrss_api_password"
        :placeholder="t('freshrssApiPasswordPlaceholder')"
        class="input-field w-32 sm:w-48 text-xs sm:text-sm"
        @input="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              freshrss_api_password: (e.target as HTMLInputElement).value,
            })
        "
      />
    </div>

    <!-- Sync Button -->
    <div class="sub-setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhCloudCheck :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('freshrssSyncNow') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('freshrssSyncNowDesc') }}
          </div>
          <div class="text-xs text-text-secondary mt-1">
            {{ t('freshrssLastSync') }}:
            <span class="theme-number">{{ formatSyncTime(syncStatus.last_sync_time) }}</span>
          </div>
        </div>
      </div>
      <button class="btn-secondary" :disabled="isSyncing" @click="syncNow">
        <PhArrowClockwise :size="16" class="sm:w-5 sm:h-5" :class="{ 'animate-spin': isSyncing }" />
        {{ isSyncing ? t('freshrssSyncing') : t('freshrssSync') }}
      </button>
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

.btn-secondary {
  @apply bg-bg-tertiary border border-border text-text-primary px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium hover:bg-bg-secondary transition-colors;
}

.btn-secondary:disabled {
  @apply cursor-not-allowed opacity-50;
}

.setting-group {
  @apply space-y-2 sm:space-y-3;
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

.theme-number {
  @apply text-accent font-semibold;
}
</style>

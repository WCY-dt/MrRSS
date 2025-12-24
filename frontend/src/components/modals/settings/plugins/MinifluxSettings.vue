<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhGlobe, PhKey, PhWifiHigh, PhArrowClockwise } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

const isSyncing = ref(false);
const isTesting = ref(false);

// Test connection to Miniflux server
async function testConnection() {
  if (!props.settings.miniflux_server_url || !props.settings.miniflux_api_key) {
    window.showToast(t('minifluxMissingCredentials'), 'error');
    return;
  }

  isTesting.value = true;

  try {
    const response = await fetch('/api/miniflux/test-connection', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        server_url: props.settings.miniflux_server_url,
        api_key: props.settings.miniflux_api_key,
      }),
    });

    const data = await response.json();

    if (data.success) {
      const feedCount = data.feedCount || 0;
      window.showToast(
        `${t('connectionSuccessful')} - ${feedCount} ${t('feeds')}`,
        'success'
      );
    } else {
      throw new Error(data.message || t('connectionFailed'));
    }
  } catch (error) {
    window.showToast(error instanceof Error ? error.message : t('connectionFailed'), 'error');
  } finally {
    isTesting.value = false;
  }
}

// Sync with Miniflux server
async function syncNow() {
  if (!props.settings.miniflux_server_url || !props.settings.miniflux_api_key) {
    window.showToast(t('minifluxMissingCredentials'), 'error');
    return;
  }

  isSyncing.value = true;

  try {
    const response = await fetch('/api/miniflux/sync', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    const data = await response.json();

    if (data.success) {
      window.showToast(t('minifluxSyncSuccess'), 'success');
    } else {
      throw new Error(data.message || t('minifluxSyncFailed'));
    }
  } catch (error) {
    window.showToast(error instanceof Error ? error.message : t('minifluxSyncFailed'), 'error');
  } finally {
    isSyncing.value = false;
  }
}
</script>

<template>
  <!-- Miniflux Sync Settings -->
  <div class="setting-item">
    <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
      <img
        src="/assets/plugin_icons/miniflux.svg"
        alt="Miniflux"
        class="w-5 h-5 sm:w-6 sm:h-6 mt-0.5 shrink-0"
      />
      <div class="flex-1 min-w-0">
        <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
          {{ t('minifluxSync') }}
        </div>
        <div class="text-xs text-text-secondary hidden sm:block">
          {{ t('minifluxSyncDesc') }}
        </div>
      </div>
    </div>
  </div>
  <div class="ml-2 sm:ml-4 space-y-2 sm:space-y-3 border-l-2 border-border pl-2 sm:pl-4">
    <!-- Server URL -->
    <div class="sub-setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhGlobe :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('minifluxServerUrl') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('minifluxServerUrlDesc') }}
          </div>
        </div>
      </div>
      <input
        type="url"
        :value="props.settings.miniflux_server_url"
        :placeholder="t('minifluxServerUrlPlaceholder')"
        class="input-field w-32 sm:w-48 text-xs sm:text-sm"
        @input="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              miniflux_server_url: (e.target as HTMLInputElement).value,
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
            {{ t('minifluxApiKey') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('minifluxApiKeyDesc') }}
          </div>
        </div>
      </div>
      <input
        type="password"
        :value="props.settings.miniflux_api_key"
        :placeholder="t('minifluxApiKeyPlaceholder')"
        class="input-field w-32 sm:w-48 text-xs sm:text-sm"
        @input="
          (e) =>
            emit('update:settings', {
              ...props.settings,
              miniflux_api_key: (e.target as HTMLInputElement).value,
            })
        "
      />
    </div>

    <!-- Connection Test and Sync Buttons -->
    <div class="sub-setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhWifiHigh :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('actions') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('testConnectionDesc') }}
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2 shrink-0">
        <button
          :disabled="isTesting"
          class="btn-secondary text-xs sm:text-sm px-2 sm:px-3 py-1 sm:py-1.5"
          @click="testConnection"
        >
          <PhWifiHigh
            :size="16"
            :class="{ 'animate-pulse': isTesting, 'sm:w-5 sm:h-5': true }"
          />
          {{ isTesting ? t('testing') : t('testConnection') }}
        </button>
        <button
          :disabled="isSyncing"
          class="btn-primary text-xs sm:text-sm px-2 sm:px-3 py-1 sm:py-1.5"
          @click="syncNow"
        >
          <PhArrowClockwise
            :size="16"
            :class="{ 'animate-spin': isSyncing, 'sm:w-5 sm:h-5': true }"
          />
          {{ isSyncing ? t('syncing') : t('syncNow') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";

.input-field {
  @apply p-1.5 sm:p-2.5 border border-border rounded-md bg-bg-secondary text-text-primary focus:border-accent focus:outline-none transition-colors;
}
.setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border;
}
.sub-setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-2.5 rounded-md bg-bg-tertiary;
}
.btn-primary {
  @apply bg-accent text-white border-none px-2 py-1 sm:px-3 sm:py-1.5 rounded-lg cursor-pointer flex items-center gap-1 sm:gap-2 font-medium hover:bg-accent-hover transition-colors text-sm sm:text-base;
}
.btn-secondary {
  @apply bg-bg-secondary text-text-primary border border-border px-2 py-1 sm:px-3 sm:py-1.5 rounded-lg cursor-pointer flex items-center gap-1 sm:gap-2 font-medium hover:bg-bg-primary transition-colors text-sm sm:text-base;
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

@keyframes pulse {
  0%,
  100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.animate-pulse {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}
</style>

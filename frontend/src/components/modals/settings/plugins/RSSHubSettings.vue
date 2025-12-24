<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhBookOpen, PhGlobe, PhKey, PhToggleLeft, PhWifiHigh } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';
import { openInBrowser } from '@/utils/browser';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

const isTesting = ref(false);
const testResult = ref<{ success: boolean; message: string } | null>(null);

// Test RSSHub connection
async function testConnection() {
  isTesting.value = true;
  testResult.value = null;

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
    if (response.ok && data.success) {
      testResult.value = {
        success: true,
        message: t('rsshubTestSuccess'),
      };
      window.showToast?.(t('rsshubTestSuccess'), 'success');
    } else {
      throw new Error(data.error || t('rsshubTestFailed'));
    }
  } catch (error) {
    testResult.value = {
      success: false,
      message: error instanceof Error ? error.message : t('rsshubTestFailed'),
    };
    window.showToast?.(testResult.value.message, 'error');
  } finally {
    isTesting.value = false;
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
    <div class="sub-setting-item">
      <div class="flex-1 flex items-start gap-2 sm:gap-3 min-w-0">
        <PhBookOpen :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="text-xs text-text-secondary leading-relaxed">
            <span class="font-medium text-text-primary">{{ t('rsshubUsageHint') }}</span>
            <span
              class="text-accent hover:underline font-semibold ml-1 cursor-pointer"
              @click="openInBrowser('https://docs.rsshub.app')"
            >
              {{ t('rsshubUsageHintLink') }} →
            </span>
          </div>
        </div>
      </div>
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
        <PhToggleLeft :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
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

    <!-- Connection Test -->
    <div class="sub-setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhWifiHigh :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('testConnection') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('rsshubTestConnectionDesc') }}
          </div>
        </div>
      </div>
      <div class="flex items-center gap-2 shrink-0">
        <button
          :disabled="isTesting"
          class="btn-primary text-xs sm:text-sm px-2 sm:px-3 py-1 sm:py-1.5"
          @click="testConnection"
        >
          <PhWifiHigh :size="16" :class="{ 'animate-spin': isTesting, 'sm:w-5 sm:h-5': true }" />
          {{ isTesting ? t('testing') : t('testConnection') }}
        </button>
      </div>
    </div>

    <!-- Test Result -->
    <div v-if="testResult" class="sub-setting-item">
      <div :class="['text-xs sm:text-sm', testResult.success ? 'text-green-500' : 'text-red-500']">
        {{ testResult.message }}
      </div>
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
  @apply bg-accent text-white border-none px-2 py-1 sm:px-3 sm:py-1.5 rounded-lg cursor-pointer flex items-center gap-1 sm:gap-2 font-medium hover:bg-accent-hover transition-colors text-sm sm:text-base disabled:opacity-50 disabled:cursor-not-allowed;
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
</style>

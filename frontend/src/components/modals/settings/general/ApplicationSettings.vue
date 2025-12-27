<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import {
  PhPalette,
  PhMoon,
  PhTranslate,
  PhPower,
  PhArchiveTray,
  PhGitPullRequest,
} from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();
</script>

<template>
  <div class="setting-group">
    <label
      class="font-semibold mb-2 sm:mb-3 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2"
    >
      <PhPalette :size="14" class="sm:w-4 sm:h-4" />
      {{ t('application') }}
    </label>

    <div class="setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhPower :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('startupOnBoot') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('startupOnBootDesc') }}
          </div>
        </div>
      </div>
      <input
        :checked="settings.startup_on_boot"
        type="checkbox"
        class="toggle"
        @change="
          (e) =>
            emit('update:settings', {
              ...settings,
              startup_on_boot: (e.target as HTMLInputElement).checked,
            })
        "
      />
    </div>

    <div class="setting-item mt-2 sm:mt-3">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhGitPullRequest :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('autoUpdateApp') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('autoUpdateAppDesc') }}
          </div>
        </div>
      </div>
      <input
        :checked="settings.auto_update"
        type="checkbox"
        class="toggle"
        @change="
          (e) =>
            emit('update:settings', {
              ...settings,
              auto_update: (e.target as HTMLInputElement).checked,
            })
        "
      />
    </div>

    <div class="setting-item mt-2 sm:mt-3">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhArchiveTray :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('closeToTray') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('closeToTrayDesc') }}
          </div>
        </div>
      </div>
      <input
        :checked="settings.close_to_tray"
        type="checkbox"
        class="toggle"
        @change="
          (e) =>
            emit('update:settings', {
              ...settings,
              close_to_tray: (e.target as HTMLInputElement).checked,
            })
        "
      />
    </div>

    <div class="setting-item mt-2 sm:mt-3">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhMoon :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">{{ t('theme') }}</div>
          <div class="text-xs text-text-secondary hidden sm:block">{{ t('themeDesc') }}</div>
        </div>
      </div>
      <select
        :value="settings.theme"
        data-testid="theme-selector"
        class="input-field w-24 sm:w-48 text-xs sm:text-sm"
        @change="
          (e) =>
            emit('update:settings', { ...settings, theme: (e.target as HTMLSelectElement).value })
        "
      >
        <option value="light">{{ t('light') }}</option>
        <option value="dark">{{ t('dark') }}</option>
        <option value="auto">{{ t('auto') }}</option>
      </select>
    </div>

    <div class="setting-item mt-2 sm:mt-3">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhTranslate :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">{{ t('language') }}</div>
          <div class="text-xs text-text-secondary hidden sm:block">{{ t('languageDesc') }}</div>
        </div>
      </div>
      <select
        :value="settings.language"
        data-testid="language-selector"
        class="input-field w-24 sm:w-48 text-xs sm:text-sm"
        @change="
          (e) =>
            emit('update:settings', {
              ...settings,
              language: (e.target as HTMLSelectElement).value,
            })
        "
      >
        <option value="en-US">{{ t('english') }}</option>
        <option value="zh-CN">{{ t('chinese') }}</option>
      </select>
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
</style>

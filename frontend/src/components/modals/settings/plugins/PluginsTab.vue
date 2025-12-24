<script setup lang="ts">
import { computed } from 'vue';
import { PhInfo } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';
import { useSettingsAutoSave } from '@/composables/core/useSettingsAutoSave';
import { useI18n } from 'vue-i18n';
import ObsidianSettings from './ObsidianSettings.vue';
import FreshRSSSettings from './FreshRSSSettings.vue';
import MinifluxSettings from './MinifluxSettings.vue';

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();
const { t } = useI18n();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

// Create a computed ref that returns the settings object
// This ensures reactivity while allowing modifications
const settingsRef = computed(() => props.settings);

// Use composable for auto-save with reactivity
useSettingsAutoSave(settingsRef);

// Handler for settings updates from child components
function handleUpdateSettings(updatedSettings: SettingsData) {
  // Emit the updated settings to parent
  emit('update:settings', updatedSettings);
}
</script>

<template>
  <div class="space-y-4 sm:space-y-6">
    <div class="tip-box">
      <PhInfo :size="16" class="text-accent shrink-0 sm:w-5 sm:h-5" />
      <span class="text-xs sm:text-sm">{{ t('isInDevelopment') }}</span>
    </div>

    <ObsidianSettings :settings="settings" @update:settings="handleUpdateSettings" />

    <FreshRSSSettings :settings="settings" @update:settings="handleUpdateSettings" />

    <MinifluxSettings :settings="settings" @update:settings="handleUpdateSettings" />
  </div>
</template>

<style scoped>
@reference "../../../../style.css";

.tip-box {
  @apply flex items-center gap-2 sm:gap-3 py-2 sm:py-2.5 px-2.5 sm:px-3 rounded-lg;
  background-color: rgba(59, 130, 246, 0.05);
  border: 1px solid rgba(59, 130, 246, 0.3);
}
</style>

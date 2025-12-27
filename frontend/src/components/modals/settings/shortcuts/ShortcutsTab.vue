<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ref, onMounted, onUnmounted, computed, watch, type Ref, type Component } from 'vue';
import {
  PhKeyboard,
  PhArrowDown,
  PhArrowUp,
  PhArrowRight,
  PhX,
  PhBookOpen,
  PhStar,
  PhClockCountdown,
  PhArrowSquareOut,
  PhArticle,
  PhArrowClockwise,
  PhCheckCircle,
  PhGear,
  PhPlus,
  PhMagnifyingGlass,
  PhListDashes,
  PhCircle,
  PhHeart,
  PhArrowCounterClockwise,
  PhInfo,
  PhFunnel,
} from '@phosphor-icons/vue';
import ShortcutItem from './ShortcutItem.vue';
import type { SettingsData } from '@/types/settings';
import { useSettingsAutoSave } from '@/composables/core/useSettingsAutoSave';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:settings': [settings: SettingsData];
}>();

// Create a computed ref that returns the settings object for auto-save
const settingsRef = computed(() => props.settings);

// Use composable for auto-save with reactivity
useSettingsAutoSave(settingsRef);

// Handle shortcuts enabled toggle change
function handleToggleChange(event: Event) {
  const target = event.target as HTMLInputElement;
  emit('update:settings', {
    ...props.settings,
    shortcuts_enabled: target.checked,
  });
  // Dispatch event to notify keyboard shortcuts system
  if (typeof window !== 'undefined') {
    window.dispatchEvent(
      new CustomEvent('shortcuts-enabled-changed', {
        detail: { enabled: target.checked },
      })
    );
  }
}

interface Shortcuts {
  nextArticle: string;
  previousArticle: string;
  openArticle: string;
  closeArticle: string;
  toggleReadStatus: string;
  toggleFavoriteStatus: string;
  toggleReadLaterStatus: string;
  openInBrowser: string;
  toggleContentView: string;
  refreshFeeds: string;
  markAllRead: string;
  openSettings: string;
  addFeed: string;
  focusSearch: string;
  toggleFilter: string;
  goToAllArticles: string;
  goToUnread: string;
  goToFavorites: string;
  goToReadLater: string;
}

interface ShortcutItemData {
  key: keyof Shortcuts;
  label: string;
  icon: Component;
}

// Default shortcuts configuration
const defaultShortcuts: Shortcuts = {
  nextArticle: 'j',
  previousArticle: 'k',
  openArticle: 'Enter',
  closeArticle: 'Escape',
  toggleReadStatus: 'r',
  toggleFavoriteStatus: 's',
  toggleReadLaterStatus: 'l',
  openInBrowser: 'o',
  toggleContentView: 'v',
  refreshFeeds: 'Shift+r',
  markAllRead: 'Shift+a',
  openSettings: ',',
  addFeed: 'a',
  focusSearch: '/',
  toggleFilter: 'f',
  goToAllArticles: '1',
  goToUnread: '2',
  goToFavorites: '3',
  goToReadLater: '4',
};

// Current shortcuts (loaded from settings or use defaults)
const shortcuts: Ref<Shortcuts> = ref({ ...defaultShortcuts });

// Track which shortcut is being edited
const editingShortcut: Ref<keyof Shortcuts | null> = ref(null);
const recordedKey = ref('');

// Shortcut groups for display
const shortcutGroups = computed<Array<{ label: string; items: ShortcutItemData[] }>>(() => [
  {
    label: t('shortcutNavigation'),
    items: [
      { key: 'nextArticle', label: t('nextArticle'), icon: PhArrowDown },
      { key: 'previousArticle', label: t('previousArticle'), icon: PhArrowUp },
      { key: 'openArticle', label: t('openArticle'), icon: PhArrowRight },
      { key: 'closeArticle', label: t('closeArticle'), icon: PhX },
      { key: 'goToAllArticles', label: t('goToAllArticles'), icon: PhListDashes },
      { key: 'goToUnread', label: t('goToUnread'), icon: PhCircle },
      { key: 'goToFavorites', label: t('goToFavorites'), icon: PhHeart },
      { key: 'goToReadLater', label: t('goToReadLater'), icon: PhClockCountdown },
    ],
  },
  {
    label: t('shortcutArticles'),
    items: [
      { key: 'toggleReadStatus', label: t('toggleReadStatus'), icon: PhBookOpen },
      { key: 'toggleFavoriteStatus', label: t('toggleFavoriteStatus'), icon: PhStar },
      { key: 'toggleReadLaterStatus', label: t('toggleReadLaterStatus'), icon: PhClockCountdown },
      { key: 'openInBrowser', label: t('openInBrowserShortcut'), icon: PhArrowSquareOut },
      { key: 'toggleContentView', label: t('toggleContentView'), icon: PhArticle },
    ],
  },
  {
    label: t('shortcutOther'),
    items: [
      { key: 'refreshFeeds', label: t('refreshFeedsShortcut'), icon: PhArrowClockwise },
      { key: 'markAllRead', label: t('markAllReadShortcut'), icon: PhCheckCircle },
      { key: 'openSettings', label: t('openSettingsShortcut'), icon: PhGear },
      { key: 'addFeed', label: t('addFeedShortcut'), icon: PhPlus },
      { key: 'focusSearch', label: t('focusFeedSearch'), icon: PhMagnifyingGlass },
      { key: 'toggleFilter', label: t('toggleFilter'), icon: PhFunnel },
    ],
  },
]);

// Load shortcuts from settings
onMounted(() => {
  if (props.settings.shortcuts) {
    try {
      const parsed =
        typeof props.settings.shortcuts === 'string'
          ? JSON.parse(props.settings.shortcuts)
          : props.settings.shortcuts;
      shortcuts.value = { ...defaultShortcuts, ...parsed };
    } catch (e) {
      console.error('Error parsing shortcuts:', e);
      shortcuts.value = { ...defaultShortcuts };
    }
  }

  // Add global keyboard listener for recording
  window.addEventListener('keydown', handleKeyRecord, true);
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyRecord, true);
});

// Start editing a shortcut
function startEditing(shortcutKey: keyof Shortcuts) {
  editingShortcut.value = shortcutKey;
  recordedKey.value = '';
}

// Stop editing
function stopEditing() {
  editingShortcut.value = null;
  recordedKey.value = '';
}

// Handle key recording
function handleKeyRecord(e: KeyboardEvent) {
  if (!editingShortcut.value) return;

  e.preventDefault();
  e.stopPropagation();

  // Handle Escape to clear the shortcut
  if (e.key === 'Escape' && !e.shiftKey && !e.ctrlKey && !e.altKey && !e.metaKey) {
    // Clear the shortcut
    shortcuts.value[editingShortcut.value] = '';
    saveShortcuts();
    window.showToast(t('shortcutCleared'), 'info');
    stopEditing();
    return;
  }

  // Build key combination
  let key = '';
  if (e.ctrlKey) key += 'Ctrl+';
  if (e.altKey) key += 'Alt+';
  if (e.shiftKey) key += 'Shift+';
  if (e.metaKey) key += 'Meta+';

  // Get the actual key
  let actualKey = e.key;

  // Skip modifier keys alone
  if (['Control', 'Alt', 'Shift', 'Meta'].includes(actualKey)) {
    return;
  }

  // Normalize key names
  if (actualKey === ' ') actualKey = 'Space';
  else if (actualKey.length === 1) actualKey = actualKey.toLowerCase();

  key += actualKey;

  // Check for conflicts
  const conflictKey = Object.entries(shortcuts.value).find(
    ([k, v]) => v === key && k !== editingShortcut.value
  );

  if (conflictKey) {
    window.showToast(t('shortcutConflict'), 'warning');
    stopEditing();
    return;
  }

  // Update the shortcut
  shortcuts.value[editingShortcut.value] = key;
  saveShortcuts();
  window.showToast(t('shortcutUpdated'), 'success');
  stopEditing();
}

// Save shortcuts to settings
async function saveShortcuts() {
  try {
    // Update props.settings.shortcuts
    const updatedSettings = { ...props.settings, shortcuts: JSON.stringify(shortcuts.value) };
    emit('update:settings', updatedSettings);

    // The parent component will handle auto-save via the watcher
    // But we also dispatch an event to notify the app
    window.dispatchEvent(
      new CustomEvent('shortcuts-changed', {
        detail: { shortcuts: shortcuts.value },
      })
    );
  } catch (e) {
    console.error('Error saving shortcuts:', e);
  }
}

// Reset all shortcuts to defaults
function resetToDefaults() {
  shortcuts.value = { ...defaultShortcuts };
  saveShortcuts();
  window.showToast(t('shortcutUpdated'), 'success');
}

// Watch for settings changes from parent
watch(
  () => props.settings.shortcuts,
  (newVal) => {
    if (newVal) {
      try {
        const parsed = typeof newVal === 'string' ? JSON.parse(newVal) : newVal;
        shortcuts.value = { ...defaultShortcuts, ...parsed };
      } catch (e) {
        console.error('Error parsing shortcuts:', e);
      }
    }
  },
  { immediate: true }
);
</script>

<template>
  <div class="space-y-4 sm:space-y-6">
    <div class="flex items-center justify-between mb-3">
      <div class="flex items-center gap-2 sm:gap-3">
        <PhKeyboard :size="20" class="text-text-secondary sm:w-6 sm:h-6" />
        <div>
          <h3 class="font-semibold text-sm sm:text-base">{{ t('shortcuts') }}</h3>
          <p class="text-xs text-text-secondary hidden sm:block">{{ t('shortcutsDesc') }}</p>
        </div>
      </div>
      <button class="btn-secondary" @click="resetToDefaults">
        <PhArrowCounterClockwise :size="16" class="sm:w-5 sm:h-5" />
        {{ t('resetToDefault') }}
      </button>
    </div>

    <!-- Enable/Disable Shortcuts Toggle -->
    <div class="setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhKeyboard :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('shortcutsEnabled') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('shortcutsEnabledDesc') }}
          </div>
        </div>
      </div>
      <input
        :checked="settings.shortcuts_enabled === true"
        type="checkbox"
        class="toggle"
        @change="handleToggleChange"
      />
    </div>

    <!-- Tip moved to top with improved styling -->
    <div class="tip-box">
      <PhInfo :size="16" class="text-accent shrink-0 sm:w-5 sm:h-5" />
      <span class="text-xs sm:text-sm">{{ t('escToClear') }}</span>
    </div>

    <div v-for="group in shortcutGroups" :key="group.label" class="setting-group">
      <label
        class="font-semibold mb-2 sm:mb-3 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2"
      >
        {{ group.label }}
      </label>

      <div class="space-y-2">
        <ShortcutItem
          v-for="item in group.items"
          :key="item.key"
          :item="item"
          :shortcut-value="shortcuts[item.key as keyof Shortcuts]"
          :is-editing="editingShortcut === item.key"
          @edit="startEditing(item.key as keyof Shortcuts)"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../../style.css";

.btn-secondary {
  @apply bg-bg-tertiary border border-border text-text-primary px-3 sm:px-4 py-1.5 sm:py-2 rounded-md cursor-pointer flex items-center gap-1.5 sm:gap-2 font-medium hover:bg-bg-secondary transition-colors;
}

.setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border;
}

.tip-box {
  @apply flex items-center gap-2 sm:gap-3 py-2 sm:py-2.5 px-2.5 sm:px-3 rounded-lg;
  background-color: rgba(59, 130, 246, 0.05);
  border: 1px solid rgba(59, 130, 246, 0.3);
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
</style>

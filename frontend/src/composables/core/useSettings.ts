/**
 * Composable for settings management
 */
import { ref, type Ref, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import type { SettingsData } from '@/types/settings';
import type { ThemePreference } from '@/stores/app';
import { generateInitialSettings, parseSettingsData } from './useSettings.generated';

export function useSettings() {
  const { locale } = useI18n();

  // Use generated helper for initial settings (alphabetically sorted)
  const settings: Ref<SettingsData> = ref(generateInitialSettings());

  /**
   * Fetch settings from backend
   */
  async function fetchSettings(): Promise<SettingsData> {
    try {
      const res = await fetch('/api/settings');
      const data = await res.json();

      // Use generated helper to parse settings (alphabetically sorted)
      settings.value = parseSettingsData(data);

      return settings.value;
    } catch (e) {
      console.error('Error fetching settings:', e);
      throw e;
    }
  }

  /**
   * Apply fetched settings to the app
   */

  function applySettings(data: SettingsData, setTheme: (preference: ThemePreference) => void) {
    // Apply the saved language
    if (data.language) {
      locale.value = data.language;
    }

    // Apply the saved theme
    if (data.theme) {
      setTheme(data.theme as ThemePreference);
    }

    // Initialize shortcuts in store
    if (data.shortcuts) {
      try {
        const parsed = JSON.parse(data.shortcuts);
        window.dispatchEvent(
          new CustomEvent('shortcuts-changed', {
            detail: { shortcuts: parsed },
          })
        );
      } catch (e) {
        console.error('Error parsing shortcuts:', e);
      }
    }
  }

  /**
   * Handle settings-updated event
   * Re-fetches settings when backend updates them (e.g., after feed refresh)
   */
  function handleSettingsUpdated() {
    fetchSettings().catch((e) => {
      console.error('Error re-fetching settings after update:', e);
    });
  }

  // Listen for settings-updated events
  onMounted(() => {
    window.addEventListener('settings-updated', handleSettingsUpdated);
  });

  onUnmounted(() => {
    window.removeEventListener('settings-updated', handleSettingsUpdated);
  });

  return {
    settings,
    fetchSettings,
    applySettings,
  };
}

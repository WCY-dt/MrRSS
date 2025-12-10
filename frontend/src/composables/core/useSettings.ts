/**
 * Composable for settings management
 */
import { ref, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import type { SettingsData } from '@/types/settings';
import type { ThemePreference } from '@/stores/app';
import { settingsDefaults } from '@/config/defaults';

export function useSettings() {
  const { locale } = useI18n();

  const settings: Ref<SettingsData> = ref({
    update_interval: settingsDefaults.update_interval,
    translation_enabled: settingsDefaults.translation_enabled,
    target_language: settingsDefaults.target_language,
    translation_provider: settingsDefaults.translation_provider,
    deepl_api_key: settingsDefaults.deepl_api_key,
    auto_cleanup_enabled: settingsDefaults.auto_cleanup_enabled,
    max_cache_size_mb: settingsDefaults.max_cache_size_mb,
    max_article_age_days: settingsDefaults.max_article_age_days,
    language: locale.value || settingsDefaults.language,
    theme: settingsDefaults.theme,
    last_article_update: settingsDefaults.last_article_update,
    show_hidden_articles: settingsDefaults.show_hidden_articles,
    default_view_mode: settingsDefaults.default_view_mode,
    startup_on_boot: settingsDefaults.startup_on_boot,
    shortcuts: settingsDefaults.shortcuts,
    rules: settingsDefaults.rules,
    summary_enabled: settingsDefaults.summary_enabled,
    summary_length: settingsDefaults.summary_length,
    summary_provider: settingsDefaults.summary_provider,
    summary_ai_api_key: settingsDefaults.summary_ai_api_key,
    summary_ai_endpoint: settingsDefaults.summary_ai_endpoint,
    summary_ai_model: settingsDefaults.summary_ai_model,
    summary_ai_system_prompt: settingsDefaults.summary_ai_system_prompt,
    label_enabled: settingsDefaults.label_enabled,
    label_provider: settingsDefaults.label_provider,
    label_ai_api_key: settingsDefaults.label_ai_api_key,
    label_ai_endpoint: settingsDefaults.label_ai_endpoint,
    label_ai_model: settingsDefaults.label_ai_model,
    label_ai_system_prompt: settingsDefaults.label_ai_system_prompt,
    label_show_in_list: settingsDefaults.label_show_in_list,
    label_max_count: settingsDefaults.label_max_count,
    baidu_app_id: settingsDefaults.baidu_app_id,
    baidu_secret_key: settingsDefaults.baidu_secret_key,
    ai_api_key: settingsDefaults.ai_api_key,
    ai_endpoint: settingsDefaults.ai_endpoint,
    ai_model: settingsDefaults.ai_model,
    ai_system_prompt: settingsDefaults.ai_system_prompt,
  });

  /**
   * Fetch settings from backend
   */
  async function fetchSettings() {
    try {
      const res = await fetch('/api/settings');
      const data = await res.json();

      settings.value = {
        update_interval: data.update_interval || settingsDefaults.update_interval,
        translation_enabled: data.translation_enabled === 'true',
        target_language: data.target_language || settingsDefaults.target_language,
        translation_provider: data.translation_provider || settingsDefaults.translation_provider,
        deepl_api_key: data.deepl_api_key || settingsDefaults.deepl_api_key,
        auto_cleanup_enabled: data.auto_cleanup_enabled === 'true',
        max_cache_size_mb: parseInt(data.max_cache_size_mb) || settingsDefaults.max_cache_size_mb,
        max_article_age_days:
          parseInt(data.max_article_age_days) || settingsDefaults.max_article_age_days,
        language: data.language || locale.value || settingsDefaults.language,
        theme: data.theme || settingsDefaults.theme,
        last_article_update: data.last_article_update || settingsDefaults.last_article_update,
        show_hidden_articles: data.show_hidden_articles === 'true',
        default_view_mode: data.default_view_mode || settingsDefaults.default_view_mode,
        startup_on_boot: data.startup_on_boot === 'true',
        shortcuts: data.shortcuts || settingsDefaults.shortcuts,
        rules: data.rules || settingsDefaults.rules,
        summary_enabled: data.summary_enabled === 'true',
        summary_length: data.summary_length || settingsDefaults.summary_length,
        summary_provider: data.summary_provider || settingsDefaults.summary_provider,
        summary_ai_api_key: data.summary_ai_api_key || settingsDefaults.summary_ai_api_key,
        summary_ai_endpoint: data.summary_ai_endpoint || settingsDefaults.summary_ai_endpoint,
        summary_ai_model: data.summary_ai_model || settingsDefaults.summary_ai_model,
        summary_ai_system_prompt:
          data.summary_ai_system_prompt || settingsDefaults.summary_ai_system_prompt,
        label_enabled: data.label_enabled === 'true',
        label_provider: data.label_provider || settingsDefaults.label_provider,
        label_ai_api_key: data.label_ai_api_key || settingsDefaults.label_ai_api_key,
        label_ai_endpoint: data.label_ai_endpoint || settingsDefaults.label_ai_endpoint,
        label_ai_model: data.label_ai_model || settingsDefaults.label_ai_model,
        label_ai_system_prompt:
          data.label_ai_system_prompt || settingsDefaults.label_ai_system_prompt,
        label_show_in_list: data.label_show_in_list === 'true',
        label_max_count: data.label_max_count || settingsDefaults.label_max_count,
        baidu_app_id: data.baidu_app_id || settingsDefaults.baidu_app_id,
        baidu_secret_key: data.baidu_secret_key || settingsDefaults.baidu_secret_key,
        ai_api_key: data.ai_api_key || settingsDefaults.ai_api_key,
        ai_endpoint: data.ai_endpoint || settingsDefaults.ai_endpoint,
        ai_model: data.ai_model || settingsDefaults.ai_model,
        ai_system_prompt: data.ai_system_prompt || settingsDefaults.ai_system_prompt,
      };

      return settings.value;
    } catch (e) {
      console.error('Error fetching settings:', e);
      throw e;
    }
  }

  /**
   * Apply fetched settings to the app
   */
  function applySettings(data: SettingsData, setTheme: (theme: ThemePreference) => void) {
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

  return {
    settings,
    fetchSettings,
    applySettings,
  };
}

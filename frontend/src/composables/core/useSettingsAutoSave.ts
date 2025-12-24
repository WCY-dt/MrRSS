/**
 * Composable for auto-saving settings with debouncing
 */
import { ref, watch, onMounted, onUnmounted, type Ref, toRef, computed, isRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAppStore } from '@/stores/app';
import type { SettingsData } from '@/types/settings';
import { settingsDefaults } from '@/config/defaults';
import { useSettingsValidation } from './useSettingsValidation';

export function useSettingsAutoSave(settings: Ref<SettingsData> | (() => SettingsData)) {
  const { locale } = useI18n();
  const store = useAppStore();

  let saveTimeout: ReturnType<typeof setTimeout> | null = null;
  let isInitialLoad = true;

  // Convert to ref if it's a getter function
  const settingsRef = isRef(settings) ? settings : computed(settings);

  // Use validation composable
  const { isValid } = useSettingsValidation(settingsRef);

  // Track previous translation settings
  const prevTranslationSettings: Ref<{
    enabled: boolean;
    targetLang: string;
    provider: string;
  }> = ref({
    enabled: settingsDefaults.translation_enabled,
    targetLang: settingsDefaults.target_language,
    provider: settingsDefaults.translation_provider,
  });

  /**
   * Initialize translation tracking
   */
  onMounted(() => {
    setTimeout(() => {
      prevTranslationSettings.value = {
        enabled: settingsRef.value.translation_enabled,
        targetLang: settingsRef.value.target_language,
        provider: settingsRef.value.translation_provider,
      };
      isInitialLoad = false;
    }, 100);
  });

  /**
   * Save settings to backend
   */
  async function autoSave() {
    try {
      // Skip translation clearing on initial load
      if (isInitialLoad) {
        return;
      }

      // Check if translation settings changed
      const translationChanged =
        prevTranslationSettings.value.enabled !== settingsRef.value.translation_enabled ||
        prevTranslationSettings.value.provider !== settingsRef.value.translation_provider ||
        (settingsRef.value.translation_enabled &&
          prevTranslationSettings.value.targetLang !== settingsRef.value.target_language);

      // Always apply basic settings immediately (theme, language, etc.)
      // even if validation fails - these don't require API keys
      locale.value = settingsRef.value.language;
      store.setTheme(settingsRef.value.theme as 'light' | 'dark' | 'auto');
      store.startAutoRefresh(settingsRef.value.update_interval);

      // Notify components about default view mode change
      window.dispatchEvent(
        new CustomEvent('default-view-mode-changed', {
          detail: {
            mode: settingsRef.value.default_view_mode,
          },
        })
      );

      // Note: Validation is used for UI feedback only (showing red borders on invalid fields).
      // We do NOT block saving settings to the backend based on validation.
      // This allows users to save their preferences immediately, even if some fields are invalid.
      // The backend saves all provided values to the database without validation.
      // Features that require valid settings (e.g., translation with API keys) will check
      // for valid values at runtime and fail gracefully if settings are incomplete/invalid.

      // Save to backend
      await fetch('/api/settings', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          update_interval: (
            settingsRef.value.update_interval ?? settingsDefaults.update_interval
          ).toString(),
          refresh_mode: settingsRef.value.refresh_mode ?? settingsDefaults.refresh_mode,
          translation_enabled: (
            settingsRef.value.translation_enabled ?? settingsDefaults.translation_enabled
          ).toString(),
          target_language: settingsRef.value.target_language ?? settingsDefaults.target_language,
          translation_provider:
            settingsRef.value.translation_provider ?? settingsDefaults.translation_provider,
          deepl_api_key: settingsRef.value.deepl_api_key ?? settingsDefaults.deepl_api_key,
          deepl_endpoint: settingsRef.value.deepl_endpoint ?? settingsDefaults.deepl_endpoint,
          baidu_app_id: settingsRef.value.baidu_app_id ?? settingsDefaults.baidu_app_id,
          baidu_secret_key: settingsRef.value.baidu_secret_key ?? settingsDefaults.baidu_secret_key,
          ai_api_key: settingsRef.value.ai_api_key ?? settingsDefaults.ai_api_key,
          ai_endpoint: settingsRef.value.ai_endpoint ?? settingsDefaults.ai_endpoint,
          ai_model: settingsRef.value.ai_model ?? settingsDefaults.ai_model,
          ai_translation_prompt:
            settingsRef.value.ai_translation_prompt ?? settingsDefaults.ai_translation_prompt,
          ai_summary_prompt:
            settingsRef.value.ai_summary_prompt ?? settingsDefaults.ai_summary_prompt,
          ai_usage_tokens: settingsRef.value.ai_usage_tokens ?? settingsDefaults.ai_usage_tokens,
          ai_usage_limit: settingsRef.value.ai_usage_limit ?? settingsDefaults.ai_usage_limit,
          ai_chat_enabled: (
            settingsRef.value.ai_chat_enabled ?? settingsDefaults.ai_chat_enabled
          ).toString(),
          auto_cleanup_enabled: (
            settingsRef.value.auto_cleanup_enabled ?? settingsDefaults.auto_cleanup_enabled
          ).toString(),
          max_cache_size_mb: (
            settingsRef.value.max_cache_size_mb ?? settingsDefaults.max_cache_size_mb
          ).toString(),
          max_article_age_days: (
            settingsRef.value.max_article_age_days ?? settingsDefaults.max_article_age_days
          ).toString(),
          language: settingsRef.value.language ?? settingsDefaults.language,
          theme: settingsRef.value.theme ?? settingsDefaults.theme,
          show_hidden_articles: (
            settingsRef.value.show_hidden_articles ?? settingsDefaults.show_hidden_articles
          ).toString(),
          hover_mark_as_read: (
            settingsRef.value.hover_mark_as_read ?? settingsDefaults.hover_mark_as_read
          ).toString(),
          default_view_mode:
            settingsRef.value.default_view_mode ?? settingsDefaults.default_view_mode,
          media_cache_enabled: (
            settingsRef.value.media_cache_enabled ?? settingsDefaults.media_cache_enabled
          ).toString(),
          media_cache_max_size_mb: (
            settingsRef.value.media_cache_max_size_mb ?? settingsDefaults.media_cache_max_size_mb
          ).toString(),
          media_cache_max_age_days: (
            settingsRef.value.media_cache_max_age_days ?? settingsDefaults.media_cache_max_age_days
          ).toString(),
          startup_on_boot: (
            settingsRef.value.startup_on_boot ?? settingsDefaults.startup_on_boot
          ).toString(),
          close_to_tray: (
            settingsRef.value.close_to_tray ?? settingsDefaults.close_to_tray
          ).toString(),
          shortcuts: settingsRef.value.shortcuts ?? settingsDefaults.shortcuts,
          summary_enabled: (
            settingsRef.value.summary_enabled ?? settingsDefaults.summary_enabled
          ).toString(),
          summary_length: settingsRef.value.summary_length ?? settingsDefaults.summary_length,
          summary_provider: settingsRef.value.summary_provider ?? settingsDefaults.summary_provider,
          summary_trigger_mode:
            settingsRef.value.summary_trigger_mode ?? settingsDefaults.summary_trigger_mode,
          proxy_enabled: (
            settingsRef.value.proxy_enabled ?? settingsDefaults.proxy_enabled
          ).toString(),
          proxy_type: settingsRef.value.proxy_type ?? settingsDefaults.proxy_type,
          proxy_host: settingsRef.value.proxy_host ?? settingsDefaults.proxy_host,
          proxy_port: settingsRef.value.proxy_port ?? settingsDefaults.proxy_port,
          proxy_username: settingsRef.value.proxy_username ?? settingsDefaults.proxy_username,
          proxy_password: settingsRef.value.proxy_password ?? settingsDefaults.proxy_password,
          google_translate_endpoint:
            settingsRef.value.google_translate_endpoint ??
            settingsDefaults.google_translate_endpoint,
          show_article_preview_images: (
            settingsRef.value.show_article_preview_images ??
            settingsDefaults.show_article_preview_images
          ).toString(),
          obsidian_enabled: (
            settingsRef.value.obsidian_enabled ?? settingsDefaults.obsidian_enabled
          ).toString(),
          obsidian_vault: settingsRef.value.obsidian_vault ?? settingsDefaults.obsidian_vault,
          obsidian_vault_path:
            settingsRef.value.obsidian_vault_path ?? settingsDefaults.obsidian_vault_path,
          network_speed: settingsRef.value.network_speed ?? settingsDefaults.network_speed,
          network_bandwidth_mbps:
            settingsRef.value.network_bandwidth_mbps ?? settingsDefaults.network_bandwidth_mbps,
          network_latency_ms:
            settingsRef.value.network_latency_ms ?? settingsDefaults.network_latency_ms,
          max_concurrent_refreshes:
            settingsRef.value.max_concurrent_refreshes ?? settingsDefaults.max_concurrent_refreshes,
          last_network_test:
            settingsRef.value.last_network_test ?? settingsDefaults.last_network_test,
          image_gallery_enabled: (
            settingsRef.value.image_gallery_enabled ?? settingsDefaults.image_gallery_enabled
          ).toString(),
          freshrss_enabled: (
            settingsRef.value.freshrss_enabled ?? settingsDefaults.freshrss_enabled
          ).toString(),
          freshrss_server_url:
            settingsRef.value.freshrss_server_url ?? settingsDefaults.freshrss_server_url,
          freshrss_username:
            settingsRef.value.freshrss_username ?? settingsDefaults.freshrss_username,
          freshrss_api_password:
            settingsRef.value.freshrss_api_password ?? settingsDefaults.freshrss_api_password,
          full_text_fetch_enabled: (
            settingsRef.value.full_text_fetch_enabled ?? settingsDefaults.full_text_fetch_enabled
          ).toString(),
          rsshub_enabled: (
            settingsRef.value.rsshub_enabled ?? settingsDefaults.rsshub_enabled
          ).toString(),
          rsshub_instance_url:
            settingsRef.value.rsshub_instance_url ?? settingsDefaults.rsshub_instance_url,
          rsshub_fallback_enabled: (
            settingsRef.value.rsshub_fallback_enabled ?? settingsDefaults.rsshub_fallback_enabled
          ).toString(),
          rsshub_api_key: settingsRef.value.rsshub_api_key ?? settingsDefaults.rsshub_api_key,
        }),
      });

      // Clear and re-translate if translation settings changed
      if (translationChanged) {
        await fetch('/api/articles/clear-translations', { method: 'POST' });
        // Update tracking
        prevTranslationSettings.value = {
          enabled: settingsRef.value.translation_enabled,
          targetLang: settingsRef.value.target_language,
          provider: settingsRef.value.translation_provider,
        };
        // Notify ArticleList about translation settings change
        window.dispatchEvent(
          new CustomEvent('translation-settings-changed', {
            detail: {
              enabled: settingsRef.value.translation_enabled,
              targetLang: settingsRef.value.target_language,
            },
          })
        );
        // Refresh articles to show without translations, then re-translate if enabled
        store.fetchArticles();
      }

      // Refresh articles if show_hidden_articles changed
      if (settingsRef.value.show_hidden_articles !== undefined) {
        store.fetchArticles();
      }

      // Notify about show_article_preview_images change
      window.dispatchEvent(
        new CustomEvent('show-preview-images-changed', {
          detail: {
            value: settingsRef.value.show_article_preview_images,
          },
        })
      );

      // Notify about image_gallery_enabled change
      window.dispatchEvent(
        new CustomEvent('image-gallery-setting-changed', {
          detail: {
            enabled: settingsRef.value.image_gallery_enabled,
          },
        })
      );
    } catch (e) {
      console.error('Error auto-saving settings:', e);
    }
  }

  /**
   * Debounced auto-save function
   */
  function debouncedAutoSave() {
    if (saveTimeout) {
      clearTimeout(saveTimeout);
    }
    saveTimeout = setTimeout(autoSave, 500); // Wait 500ms after last change
  }

  // Watch the entire settings object for changes
  watch(() => settingsRef.value, debouncedAutoSave, { deep: true });

  // Clean up timeout on unmount to prevent memory leaks
  onUnmounted(() => {
    if (saveTimeout) {
      clearTimeout(saveTimeout);
      saveTimeout = null;
    }
  });

  return {
    autoSave,
    debouncedAutoSave,
  };
}

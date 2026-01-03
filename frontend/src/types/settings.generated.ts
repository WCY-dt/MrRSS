// Copyright 2026 Ch3nyang & MrRSS Team. All rights reserved.
//
// Auto-generated settings types
// CODE GENERATED - DO NOT EDIT MANUALLY
// To add new settings, edit internal/config/settings_schema.json and run: go run tools/settings-generator/main.go

export interface SettingsData {
  ai_api_key: string;
  ai_chat_enabled: boolean;
  ai_custom_headers: string;
  ai_endpoint: string;
  ai_model: string;
  ai_summary_prompt: string;
  ai_translation_prompt: string;
  ai_usage_limit: string;
  ai_usage_tokens: string;
  auto_cleanup_enabled: boolean;
  auto_show_all_content: boolean;
  auto_update: boolean;
  baidu_app_id: string;
  baidu_secret_key: string;
  close_to_tray: boolean;
  custom_css_file: string;
  deepl_api_key: string;
  deepl_endpoint: string;
  default_view_mode: string;
  freshrss_api_password: string;
  freshrss_auto_sync_interval: number;
  freshrss_enabled: boolean;
  freshrss_last_sync_time: string;
  freshrss_server_url: string;
  freshrss_sync_on_startup: boolean;
  freshrss_username: string;
  full_text_fetch_enabled: boolean;
  google_translate_endpoint: string;
  hover_mark_as_read: boolean;
  image_gallery_enabled: boolean;
  language: string;
  last_global_refresh: string;
  last_network_test: string;
  max_article_age_days: number;
  max_cache_size_mb: number;
  max_concurrent_refreshes: string;
  media_cache_enabled: boolean;
  media_cache_max_age_days: number;
  media_cache_max_size_mb: number;
  media_proxy_fallback: boolean;
  network_bandwidth_mbps: string;
  network_latency_ms: string;
  network_speed: string;
  obsidian_enabled: boolean;
  obsidian_vault: string;
  obsidian_vault_path: string;
  proxy_enabled: boolean;
  proxy_host: string;
  proxy_password: string;
  proxy_port: string;
  proxy_type: string;
  proxy_username: string;
  refresh_mode: string;
  retry_timeout_seconds: number;
  rsshub_api_key: string;
  rsshub_endpoint: string;
  rules: string;
  shortcuts: string;
  shortcuts_enabled: boolean;
  show_article_preview_images: boolean;
  show_hidden_articles: boolean;
  startup_on_boot: boolean;
  summary_enabled: boolean;
  summary_length: string;
  summary_provider: string;
  summary_trigger_mode: string;
  target_language: string;
  theme: string;
  translation_enabled: boolean;
  translation_provider: string;
  update_interval: number;
  window_height: string;
  window_maximized: string;
  window_width: string;
  window_x: string;
  window_y: string;
  [key: string]: unknown; // Allow additional properties
}

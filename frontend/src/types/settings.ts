/**
 * Settings types for SettingsModal and related components
 */

export interface SettingsData {
  update_interval: number;
  refresh_mode: string;
  translation_enabled: boolean;
  target_language: string;
  translation_provider: string;
  deepl_api_key: string;
  deepl_endpoint: string;
  baidu_app_id: string;
  baidu_secret_key: string;
  ai_api_key: string;
  ai_endpoint: string;
  ai_model: string;
  ai_translation_prompt: string;
  ai_summary_prompt: string;
  ai_usage_tokens: string;
  ai_usage_limit: string;
  ai_chat_enabled: boolean;
  auto_cleanup_enabled: boolean;
  max_cache_size_mb: number;
  max_article_age_days: number;
  media_cache_enabled: boolean;
  media_cache_max_size_mb: number;
  media_cache_max_age_days: number;
  language: string;
  theme: string;
  last_article_update: string;
  show_hidden_articles: boolean;
  hover_mark_as_read: boolean;
  default_view_mode: string;
  startup_on_boot: boolean;
  close_to_tray: boolean;
  shortcuts: string;
  rules: string;
  summary_enabled: boolean;
  summary_length: string;
  summary_provider: string;
  summary_trigger_mode: string;
  proxy_enabled: boolean;
  proxy_type: string;
  proxy_host: string;
  proxy_port: string;
  proxy_username: string;
  proxy_password: string;
  google_translate_endpoint: string;
  show_article_preview_images: boolean;
  obsidian_enabled: boolean;
  obsidian_vault: string;
  obsidian_vault_path: string;
  network_speed: string;
  network_bandwidth_mbps: string;
  network_latency_ms: string;
  max_concurrent_refreshes: string;
  last_network_test: string;
  image_gallery_enabled: boolean;
  freshrss_enabled: boolean;
  freshrss_server_url: string;
  freshrss_username: string;
  freshrss_api_password: string;
  full_text_fetch_enabled: boolean;
  rsshub_enabled: boolean;
  rsshub_instance_url: string;
  rsshub_fallback_enabled: boolean;
  rsshub_api_key: string;
  [key: string]: unknown; // Allow additional properties
}

export interface NetworkInfo {
  speed_level: 'slow' | 'medium' | 'fast';
  bandwidth_mbps: number;
  latency_ms: number;
  max_concurrency: number;
  detection_time: string;
  detection_success: boolean;
  error_message?: string;
}

export interface AITestInfo {
  config_valid: boolean;
  connection_success: boolean;
  model_available: boolean;
  response_time_ms: number;
  test_time: string;
  error_message?: string;
}

export interface UpdateInfo {
  has_update: boolean;
  current_version: string;
  latest_version: string;
  download_url: string;
  asset_name: string;
  is_portable: boolean;
  error?: string;
}

export interface DownloadResponse {
  success: boolean;
  file_path: string;
}

export interface InstallResponse {
  success: boolean;
}

export type TabName =
  | 'general'
  | 'feeds'
  | 'ai'
  | 'rules'
  | 'network'
  | 'plugins'
  | 'shortcuts'
  | 'about';

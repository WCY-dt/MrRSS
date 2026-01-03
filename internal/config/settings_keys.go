// Copyright 2026 Ch3nyang & MrRSS Team. All rights reserved.
//
// Package config provides settings keys for database initialization
// CODE GENERATED - DO NOT EDIT MANUALLY
// To add new settings, edit internal/config/settings_schema.json and run: go run tools/settings-generator/main.go
package config

// SettingsKeys returns all valid setting keys
func SettingsKeys() []string {
	return []string{"ai_api_key", "ai_chat_enabled", "ai_custom_headers", "ai_endpoint", "ai_model", "ai_summary_prompt", "ai_translation_prompt", "ai_usage_limit", "ai_usage_tokens", "auto_cleanup_enabled", "auto_show_all_content", "auto_update", "baidu_app_id", "baidu_secret_key", "close_to_tray", "custom_css_file", "deepl_api_key", "deepl_endpoint", "default_view_mode", "freshrss_api_password", "freshrss_auto_sync_interval", "freshrss_enabled", "freshrss_last_sync_time", "freshrss_server_url", "freshrss_sync_on_startup", "freshrss_username", "full_text_fetch_enabled", "google_translate_endpoint", "hover_mark_as_read", "image_gallery_enabled", "language", "last_global_refresh", "last_network_test", "max_article_age_days", "max_cache_size_mb", "max_concurrent_refreshes", "media_cache_enabled", "media_cache_max_age_days", "media_cache_max_size_mb", "media_proxy_fallback", "network_bandwidth_mbps", "network_latency_ms", "network_speed", "obsidian_enabled", "obsidian_vault", "obsidian_vault_path", "proxy_enabled", "proxy_host", "proxy_password", "proxy_port", "proxy_type", "proxy_username", "refresh_mode", "retry_timeout_seconds", "rsshub_api_key", "rsshub_endpoint", "rules", "shortcuts", "shortcuts_enabled", "show_article_preview_images", "show_hidden_articles", "startup_on_boot", "summary_enabled", "summary_length", "summary_provider", "summary_trigger_mode", "target_language", "theme", "translation_enabled", "translation_provider", "update_interval", "window_height", "window_maximized", "window_width", "window_x", "window_y"}
}

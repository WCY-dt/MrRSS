package settings

import (
	"encoding/json"
	"log"
	"net/http"

	"MrRSS/internal/handlers/core"
)

// HandleSettings handles GET and POST requests for application settings.
// @Summary      Get application settings
// @Description  Get all application settings (sensitive values like passwords are cleared)
// @Tags         settings
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string  "Application settings (key-value pairs)"
// @Router       /settings [get]
// CODE GENERATED - DO NOT EDIT MANUALLY
// To add new settings, edit internal/config/settings_schema.json and run: go run tools/settings-generator/main.go
func HandleSettings(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		aiApiKey, _ := h.DB.GetEncryptedSetting("ai_api_key")
		aiChatEnabled, _ := h.DB.GetSetting("ai_chat_enabled")
		aiCustomHeaders, _ := h.DB.GetSetting("ai_custom_headers")
		aiEndpoint, _ := h.DB.GetSetting("ai_endpoint")
		aiModel, _ := h.DB.GetSetting("ai_model")
		aiSummaryPrompt, _ := h.DB.GetSetting("ai_summary_prompt")
		aiTranslationPrompt, _ := h.DB.GetSetting("ai_translation_prompt")
		aiUsageLimit, _ := h.DB.GetSetting("ai_usage_limit")
		aiUsageTokens, _ := h.DB.GetSetting("ai_usage_tokens")
		autoCleanupEnabled, _ := h.DB.GetSetting("auto_cleanup_enabled")
		autoShowAllContent, _ := h.DB.GetSetting("auto_show_all_content")
		autoUpdate, _ := h.DB.GetSetting("auto_update")
		baiduAppId, _ := h.DB.GetSetting("baidu_app_id")
		baiduSecretKey, _ := h.DB.GetEncryptedSetting("baidu_secret_key")
		closeToTray, _ := h.DB.GetSetting("close_to_tray")
		customCssFile, _ := h.DB.GetSetting("custom_css_file")
		deeplApiKey, _ := h.DB.GetEncryptedSetting("deepl_api_key")
		deeplEndpoint, _ := h.DB.GetSetting("deepl_endpoint")
		defaultViewMode, _ := h.DB.GetSetting("default_view_mode")
		freshrssApiPassword, _ := h.DB.GetEncryptedSetting("freshrss_api_password")
		freshrssAutoSyncInterval, _ := h.DB.GetSetting("freshrss_auto_sync_interval")
		freshrssEnabled, _ := h.DB.GetSetting("freshrss_enabled")
		freshrssLastSyncTime, _ := h.DB.GetSetting("freshrss_last_sync_time")
		freshrssServerUrl, _ := h.DB.GetSetting("freshrss_server_url")
		freshrssSyncOnStartup, _ := h.DB.GetSetting("freshrss_sync_on_startup")
		freshrssUsername, _ := h.DB.GetSetting("freshrss_username")
		fullTextFetchEnabled, _ := h.DB.GetSetting("full_text_fetch_enabled")
		googleTranslateEndpoint, _ := h.DB.GetSetting("google_translate_endpoint")
		hoverMarkAsRead, _ := h.DB.GetSetting("hover_mark_as_read")
		imageGalleryEnabled, _ := h.DB.GetSetting("image_gallery_enabled")
		language, _ := h.DB.GetSetting("language")
		lastGlobalRefresh, _ := h.DB.GetSetting("last_global_refresh")
		lastNetworkTest, _ := h.DB.GetSetting("last_network_test")
		maxArticleAgeDays, _ := h.DB.GetSetting("max_article_age_days")
		maxCacheSizeMb, _ := h.DB.GetSetting("max_cache_size_mb")
		maxConcurrentRefreshes, _ := h.DB.GetSetting("max_concurrent_refreshes")
		mediaCacheEnabled, _ := h.DB.GetSetting("media_cache_enabled")
		mediaCacheMaxAgeDays, _ := h.DB.GetSetting("media_cache_max_age_days")
		mediaCacheMaxSizeMb, _ := h.DB.GetSetting("media_cache_max_size_mb")
		mediaProxyFallback, _ := h.DB.GetSetting("media_proxy_fallback")
		networkBandwidthMbps, _ := h.DB.GetSetting("network_bandwidth_mbps")
		networkLatencyMs, _ := h.DB.GetSetting("network_latency_ms")
		networkSpeed, _ := h.DB.GetSetting("network_speed")
		obsidianEnabled, _ := h.DB.GetSetting("obsidian_enabled")
		obsidianVault, _ := h.DB.GetSetting("obsidian_vault")
		obsidianVaultPath, _ := h.DB.GetSetting("obsidian_vault_path")
		proxyEnabled, _ := h.DB.GetSetting("proxy_enabled")
		proxyHost, _ := h.DB.GetSetting("proxy_host")
		proxyPassword, _ := h.DB.GetEncryptedSetting("proxy_password")
		proxyPort, _ := h.DB.GetSetting("proxy_port")
		proxyType, _ := h.DB.GetSetting("proxy_type")
		proxyUsername, _ := h.DB.GetEncryptedSetting("proxy_username")
		refreshMode, _ := h.DB.GetSetting("refresh_mode")
		retryTimeoutSeconds, _ := h.DB.GetSetting("retry_timeout_seconds")
		rules, _ := h.DB.GetSetting("rules")
		shortcuts, _ := h.DB.GetSetting("shortcuts")
		shortcutsEnabled, _ := h.DB.GetSetting("shortcuts_enabled")
		showArticlePreviewImages, _ := h.DB.GetSetting("show_article_preview_images")
		showHiddenArticles, _ := h.DB.GetSetting("show_hidden_articles")
		startupOnBoot, _ := h.DB.GetSetting("startup_on_boot")
		summaryEnabled, _ := h.DB.GetSetting("summary_enabled")
		summaryLength, _ := h.DB.GetSetting("summary_length")
		summaryProvider, _ := h.DB.GetSetting("summary_provider")
		summaryTriggerMode, _ := h.DB.GetSetting("summary_trigger_mode")
		targetLanguage, _ := h.DB.GetSetting("target_language")
		theme, _ := h.DB.GetSetting("theme")
		translationEnabled, _ := h.DB.GetSetting("translation_enabled")
		translationProvider, _ := h.DB.GetSetting("translation_provider")
		updateInterval, _ := h.DB.GetSetting("update_interval")
		windowHeight, _ := h.DB.GetSetting("window_height")
		windowMaximized, _ := h.DB.GetSetting("window_maximized")
		windowWidth, _ := h.DB.GetSetting("window_width")
		windowX, _ := h.DB.GetSetting("window_x")
		windowY, _ := h.DB.GetSetting("window_y")
		json.NewEncoder(w).Encode(map[string]string{
			"ai_api_key":                  aiApiKey,
			"ai_chat_enabled":             aiChatEnabled,
			"ai_custom_headers":           aiCustomHeaders,
			"ai_endpoint":                 aiEndpoint,
			"ai_model":                    aiModel,
			"ai_summary_prompt":           aiSummaryPrompt,
			"ai_translation_prompt":       aiTranslationPrompt,
			"ai_usage_limit":              aiUsageLimit,
			"ai_usage_tokens":             aiUsageTokens,
			"auto_cleanup_enabled":        autoCleanupEnabled,
			"auto_show_all_content":       autoShowAllContent,
			"auto_update":                 autoUpdate,
			"baidu_app_id":                baiduAppId,
			"baidu_secret_key":            baiduSecretKey,
			"close_to_tray":               closeToTray,
			"custom_css_file":             customCssFile,
			"deepl_api_key":               deeplApiKey,
			"deepl_endpoint":              deeplEndpoint,
			"default_view_mode":           defaultViewMode,
			"freshrss_api_password":       freshrssApiPassword,
			"freshrss_auto_sync_interval": freshrssAutoSyncInterval,
			"freshrss_enabled":            freshrssEnabled,
			"freshrss_last_sync_time":     freshrssLastSyncTime,
			"freshrss_server_url":         freshrssServerUrl,
			"freshrss_sync_on_startup":    freshrssSyncOnStartup,
			"freshrss_username":           freshrssUsername,
			"full_text_fetch_enabled":     fullTextFetchEnabled,
			"google_translate_endpoint":   googleTranslateEndpoint,
			"hover_mark_as_read":          hoverMarkAsRead,
			"image_gallery_enabled":       imageGalleryEnabled,
			"language":                    language,
			"last_global_refresh":         lastGlobalRefresh,
			"last_network_test":           lastNetworkTest,
			"max_article_age_days":        maxArticleAgeDays,
			"max_cache_size_mb":           maxCacheSizeMb,
			"max_concurrent_refreshes":    maxConcurrentRefreshes,
			"media_cache_enabled":         mediaCacheEnabled,
			"media_cache_max_age_days":    mediaCacheMaxAgeDays,
			"media_cache_max_size_mb":     mediaCacheMaxSizeMb,
			"media_proxy_fallback":        mediaProxyFallback,
			"network_bandwidth_mbps":      networkBandwidthMbps,
			"network_latency_ms":          networkLatencyMs,
			"network_speed":               networkSpeed,
			"obsidian_enabled":            obsidianEnabled,
			"obsidian_vault":              obsidianVault,
			"obsidian_vault_path":         obsidianVaultPath,
			"proxy_enabled":               proxyEnabled,
			"proxy_host":                  proxyHost,
			"proxy_password":              proxyPassword,
			"proxy_port":                  proxyPort,
			"proxy_type":                  proxyType,
			"proxy_username":              proxyUsername,
			"refresh_mode":                refreshMode,
			"retry_timeout_seconds":       retryTimeoutSeconds,
			"rules":                       rules,
			"shortcuts":                   shortcuts,
			"shortcuts_enabled":           shortcutsEnabled,
			"show_article_preview_images": showArticlePreviewImages,
			"show_hidden_articles":        showHiddenArticles,
			"startup_on_boot":             startupOnBoot,
			"summary_enabled":             summaryEnabled,
			"summary_length":              summaryLength,
			"summary_provider":            summaryProvider,
			"summary_trigger_mode":        summaryTriggerMode,
			"target_language":             targetLanguage,
			"theme":                       theme,
			"translation_enabled":         translationEnabled,
			"translation_provider":        translationProvider,
			"update_interval":             updateInterval,
			"window_height":               windowHeight,
			"window_maximized":            windowMaximized,
			"window_width":                windowWidth,
			"window_x":                    windowX,
			"window_y":                    windowY,
		})
	case http.MethodPost:
		var req struct {
			AIAPIKey                 string `json:"ai_api_key"`
			AIChatEnabled            string `json:"ai_chat_enabled"`
			AICustomHeaders          string `json:"ai_custom_headers"`
			AIEndpoint               string `json:"ai_endpoint"`
			AIModel                  string `json:"ai_model"`
			AISummaryPrompt          string `json:"ai_summary_prompt"`
			AITranslationPrompt      string `json:"ai_translation_prompt"`
			AIUsageLimit             string `json:"ai_usage_limit"`
			AIUsageTokens            string `json:"ai_usage_tokens"`
			AutoCleanupEnabled       string `json:"auto_cleanup_enabled"`
			AutoShowAllContent       string `json:"auto_show_all_content"`
			AutoUpdate               string `json:"auto_update"`
			BaiduAppId               string `json:"baidu_app_id"`
			BaiduSecretKey           string `json:"baidu_secret_key"`
			CloseToTray              string `json:"close_to_tray"`
			CustomCssFile            string `json:"custom_css_file"`
			DeeplAPIKey              string `json:"deepl_api_key"`
			DeeplEndpoint            string `json:"deepl_endpoint"`
			DefaultViewMode          string `json:"default_view_mode"`
			FreshRSSAPIPassword      string `json:"freshrss_api_password"`
			FreshRSSAutoSyncInterval string `json:"freshrss_auto_sync_interval"`
			FreshRSSEnabled          string `json:"freshrss_enabled"`
			FreshRSSLastSyncTime     string `json:"freshrss_last_sync_time"`
			FreshRSSServerUrl        string `json:"freshrss_server_url"`
			FreshRSSSyncOnStartup    string `json:"freshrss_sync_on_startup"`
			FreshRSSUsername         string `json:"freshrss_username"`
			FullTextFetchEnabled     string `json:"full_text_fetch_enabled"`
			GoogleTranslateEndpoint  string `json:"google_translate_endpoint"`
			HoverMarkAsRead          string `json:"hover_mark_as_read"`
			ImageGalleryEnabled      string `json:"image_gallery_enabled"`
			Language                 string `json:"language"`
			LastGlobalRefresh        string `json:"last_global_refresh"`
			LastNetworkTest          string `json:"last_network_test"`
			MaxArticleAgeDays        string `json:"max_article_age_days"`
			MaxCacheSizeMb           string `json:"max_cache_size_mb"`
			MaxConcurrentRefreshes   string `json:"max_concurrent_refreshes"`
			MediaCacheEnabled        string `json:"media_cache_enabled"`
			MediaCacheMaxAgeDays     string `json:"media_cache_max_age_days"`
			MediaCacheMaxSizeMb      string `json:"media_cache_max_size_mb"`
			MediaProxyFallback       string `json:"media_proxy_fallback"`
			NetworkBandwidthMbps     string `json:"network_bandwidth_mbps"`
			NetworkLatencyMs         string `json:"network_latency_ms"`
			NetworkSpeed             string `json:"network_speed"`
			ObsidianEnabled          string `json:"obsidian_enabled"`
			ObsidianVault            string `json:"obsidian_vault"`
			ObsidianVaultPath        string `json:"obsidian_vault_path"`
			ProxyEnabled             string `json:"proxy_enabled"`
			ProxyHost                string `json:"proxy_host"`
			ProxyPassword            string `json:"proxy_password"`
			ProxyPort                string `json:"proxy_port"`
			ProxyType                string `json:"proxy_type"`
			ProxyUsername            string `json:"proxy_username"`
			RefreshMode              string `json:"refresh_mode"`
			RetryTimeoutSeconds      string `json:"retry_timeout_seconds"`
			Rules                    string `json:"rules"`
			Shortcuts                string `json:"shortcuts"`
			ShortcutsEnabled         string `json:"shortcuts_enabled"`
			ShowArticlePreviewImages string `json:"show_article_preview_images"`
			ShowHiddenArticles       string `json:"show_hidden_articles"`
			StartupOnBoot            string `json:"startup_on_boot"`
			SummaryEnabled           string `json:"summary_enabled"`
			SummaryLength            string `json:"summary_length"`
			SummaryProvider          string `json:"summary_provider"`
			SummaryTriggerMode       string `json:"summary_trigger_mode"`
			TargetLanguage           string `json:"target_language"`
			Theme                    string `json:"theme"`
			TranslationEnabled       string `json:"translation_enabled"`
			TranslationProvider      string `json:"translation_provider"`
			UpdateInterval           string `json:"update_interval"`
			WindowHeight             string `json:"window_height"`
			WindowMaximized          string `json:"window_maximized"`
			WindowWidth              string `json:"window_width"`
			WindowX                  string `json:"window_x"`
			WindowY                  string `json:"window_y"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := h.DB.SetEncryptedSetting("ai_api_key", req.AIAPIKey); err != nil {
			log.Printf("Failed to save ai_api_key: %v", err)
			http.Error(w, "Failed to save ai_api_key", http.StatusInternalServerError)
			return
		}

		if req.AIChatEnabled != "" {
			h.DB.SetSetting("ai_chat_enabled", req.AIChatEnabled)
		}

		if req.AICustomHeaders != "" {
			h.DB.SetSetting("ai_custom_headers", req.AICustomHeaders)
		}

		if req.AIEndpoint != "" {
			h.DB.SetSetting("ai_endpoint", req.AIEndpoint)
		}

		if req.AIModel != "" {
			h.DB.SetSetting("ai_model", req.AIModel)
		}

		if req.AISummaryPrompt != "" {
			h.DB.SetSetting("ai_summary_prompt", req.AISummaryPrompt)
		}

		if req.AITranslationPrompt != "" {
			h.DB.SetSetting("ai_translation_prompt", req.AITranslationPrompt)
		}

		if req.AIUsageLimit != "" {
			h.DB.SetSetting("ai_usage_limit", req.AIUsageLimit)
		}

		if req.AIUsageTokens != "" {
			h.DB.SetSetting("ai_usage_tokens", req.AIUsageTokens)
		}

		if req.AutoCleanupEnabled != "" {
			h.DB.SetSetting("auto_cleanup_enabled", req.AutoCleanupEnabled)
		}

		if req.AutoShowAllContent != "" {
			h.DB.SetSetting("auto_show_all_content", req.AutoShowAllContent)
		}

		if req.AutoUpdate != "" {
			h.DB.SetSetting("auto_update", req.AutoUpdate)
		}

		if req.BaiduAppId != "" {
			h.DB.SetSetting("baidu_app_id", req.BaiduAppId)
		}

		if err := h.DB.SetEncryptedSetting("baidu_secret_key", req.BaiduSecretKey); err != nil {
			log.Printf("Failed to save baidu_secret_key: %v", err)
			http.Error(w, "Failed to save baidu_secret_key", http.StatusInternalServerError)
			return
		}

		if req.CloseToTray != "" {
			h.DB.SetSetting("close_to_tray", req.CloseToTray)
		}

		if req.CustomCssFile != "" {
			h.DB.SetSetting("custom_css_file", req.CustomCssFile)
		}

		if err := h.DB.SetEncryptedSetting("deepl_api_key", req.DeeplAPIKey); err != nil {
			log.Printf("Failed to save deepl_api_key: %v", err)
			http.Error(w, "Failed to save deepl_api_key", http.StatusInternalServerError)
			return
		}

		if req.DeeplEndpoint != "" {
			h.DB.SetSetting("deepl_endpoint", req.DeeplEndpoint)
		}

		if req.DefaultViewMode != "" {
			h.DB.SetSetting("default_view_mode", req.DefaultViewMode)
		}

		if err := h.DB.SetEncryptedSetting("freshrss_api_password", req.FreshRSSAPIPassword); err != nil {
			log.Printf("Failed to save freshrss_api_password: %v", err)
			http.Error(w, "Failed to save freshrss_api_password", http.StatusInternalServerError)
			return
		}

		if req.FreshRSSAutoSyncInterval != "" {
			h.DB.SetSetting("freshrss_auto_sync_interval", req.FreshRSSAutoSyncInterval)
		}

		if req.FreshRSSEnabled != "" {
			h.DB.SetSetting("freshrss_enabled", req.FreshRSSEnabled)
		}

		if req.FreshRSSLastSyncTime != "" {
			h.DB.SetSetting("freshrss_last_sync_time", req.FreshRSSLastSyncTime)
		}

		if req.FreshRSSServerUrl != "" {
			h.DB.SetSetting("freshrss_server_url", req.FreshRSSServerUrl)
		}

		if req.FreshRSSSyncOnStartup != "" {
			h.DB.SetSetting("freshrss_sync_on_startup", req.FreshRSSSyncOnStartup)
		}

		if req.FreshRSSUsername != "" {
			h.DB.SetSetting("freshrss_username", req.FreshRSSUsername)
		}

		if req.FullTextFetchEnabled != "" {
			h.DB.SetSetting("full_text_fetch_enabled", req.FullTextFetchEnabled)
		}

		if req.GoogleTranslateEndpoint != "" {
			h.DB.SetSetting("google_translate_endpoint", req.GoogleTranslateEndpoint)
		}

		if req.HoverMarkAsRead != "" {
			h.DB.SetSetting("hover_mark_as_read", req.HoverMarkAsRead)
		}

		if req.ImageGalleryEnabled != "" {
			h.DB.SetSetting("image_gallery_enabled", req.ImageGalleryEnabled)
		}

		if req.Language != "" {
			h.DB.SetSetting("language", req.Language)
		}

		if req.LastGlobalRefresh != "" {
			h.DB.SetSetting("last_global_refresh", req.LastGlobalRefresh)
		}

		if req.LastNetworkTest != "" {
			h.DB.SetSetting("last_network_test", req.LastNetworkTest)
		}

		if req.MaxArticleAgeDays != "" {
			h.DB.SetSetting("max_article_age_days", req.MaxArticleAgeDays)
		}

		if req.MaxCacheSizeMb != "" {
			h.DB.SetSetting("max_cache_size_mb", req.MaxCacheSizeMb)
		}

		if req.MaxConcurrentRefreshes != "" {
			h.DB.SetSetting("max_concurrent_refreshes", req.MaxConcurrentRefreshes)
		}

		if req.MediaCacheEnabled != "" {
			h.DB.SetSetting("media_cache_enabled", req.MediaCacheEnabled)
		}

		if req.MediaCacheMaxAgeDays != "" {
			h.DB.SetSetting("media_cache_max_age_days", req.MediaCacheMaxAgeDays)
		}

		if req.MediaCacheMaxSizeMb != "" {
			h.DB.SetSetting("media_cache_max_size_mb", req.MediaCacheMaxSizeMb)
		}

		if req.MediaProxyFallback != "" {
			h.DB.SetSetting("media_proxy_fallback", req.MediaProxyFallback)
		}

		if req.NetworkBandwidthMbps != "" {
			h.DB.SetSetting("network_bandwidth_mbps", req.NetworkBandwidthMbps)
		}

		if req.NetworkLatencyMs != "" {
			h.DB.SetSetting("network_latency_ms", req.NetworkLatencyMs)
		}

		if req.NetworkSpeed != "" {
			h.DB.SetSetting("network_speed", req.NetworkSpeed)
		}

		if req.ObsidianEnabled != "" {
			h.DB.SetSetting("obsidian_enabled", req.ObsidianEnabled)
		}

		if req.ObsidianVault != "" {
			h.DB.SetSetting("obsidian_vault", req.ObsidianVault)
		}

		if req.ObsidianVaultPath != "" {
			h.DB.SetSetting("obsidian_vault_path", req.ObsidianVaultPath)
		}

		if req.ProxyEnabled != "" {
			h.DB.SetSetting("proxy_enabled", req.ProxyEnabled)
		}

		if req.ProxyHost != "" {
			h.DB.SetSetting("proxy_host", req.ProxyHost)
		}

		if err := h.DB.SetEncryptedSetting("proxy_password", req.ProxyPassword); err != nil {
			log.Printf("Failed to save proxy_password: %v", err)
			http.Error(w, "Failed to save proxy_password", http.StatusInternalServerError)
			return
		}

		if req.ProxyPort != "" {
			h.DB.SetSetting("proxy_port", req.ProxyPort)
		}

		if req.ProxyType != "" {
			h.DB.SetSetting("proxy_type", req.ProxyType)
		}

		if err := h.DB.SetEncryptedSetting("proxy_username", req.ProxyUsername); err != nil {
			log.Printf("Failed to save proxy_username: %v", err)
			http.Error(w, "Failed to save proxy_username", http.StatusInternalServerError)
			return
		}

		if req.RefreshMode != "" {
			h.DB.SetSetting("refresh_mode", req.RefreshMode)
		}

		if req.RetryTimeoutSeconds != "" {
			h.DB.SetSetting("retry_timeout_seconds", req.RetryTimeoutSeconds)
		}

		if req.Rules != "" {
			h.DB.SetSetting("rules", req.Rules)
		}

		if req.Shortcuts != "" {
			h.DB.SetSetting("shortcuts", req.Shortcuts)
		}

		if req.ShortcutsEnabled != "" {
			h.DB.SetSetting("shortcuts_enabled", req.ShortcutsEnabled)
		}

		if req.ShowArticlePreviewImages != "" {
			h.DB.SetSetting("show_article_preview_images", req.ShowArticlePreviewImages)
		}

		if req.ShowHiddenArticles != "" {
			h.DB.SetSetting("show_hidden_articles", req.ShowHiddenArticles)
		}

		if req.StartupOnBoot != "" {
			h.DB.SetSetting("startup_on_boot", req.StartupOnBoot)
		}

		if req.SummaryEnabled != "" {
			h.DB.SetSetting("summary_enabled", req.SummaryEnabled)
		}

		if req.SummaryLength != "" {
			h.DB.SetSetting("summary_length", req.SummaryLength)
		}

		if req.SummaryProvider != "" {
			h.DB.SetSetting("summary_provider", req.SummaryProvider)
		}

		if req.SummaryTriggerMode != "" {
			h.DB.SetSetting("summary_trigger_mode", req.SummaryTriggerMode)
		}

		if req.TargetLanguage != "" {
			h.DB.SetSetting("target_language", req.TargetLanguage)
		}

		if req.Theme != "" {
			h.DB.SetSetting("theme", req.Theme)
		}

		if req.TranslationEnabled != "" {
			h.DB.SetSetting("translation_enabled", req.TranslationEnabled)
		}

		if req.TranslationProvider != "" {
			h.DB.SetSetting("translation_provider", req.TranslationProvider)
		}

		if req.UpdateInterval != "" {
			h.DB.SetSetting("update_interval", req.UpdateInterval)
		}

		if req.WindowHeight != "" {
			h.DB.SetSetting("window_height", req.WindowHeight)
		}

		if req.WindowMaximized != "" {
			h.DB.SetSetting("window_maximized", req.WindowMaximized)
		}

		if req.WindowWidth != "" {
			h.DB.SetSetting("window_width", req.WindowWidth)
		}

		if req.WindowX != "" {
			h.DB.SetSetting("window_x", req.WindowX)
		}

		if req.WindowY != "" {
			h.DB.SetSetting("window_y", req.WindowY)
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

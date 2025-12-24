// Package config provides centralized default values for settings.
// The defaults are loaded from config/defaults.json which is shared between
// frontend and backend to ensure consistency.
package config

import (
	_ "embed"
	"encoding/json"
	"strconv"
)

//go:embed defaults.json
var defaultsJSON []byte

// Defaults holds all default settings values
type Defaults struct {
	UpdateInterval           int    `json:"update_interval"`
	RefreshMode              string `json:"refresh_mode"`
	Language                 string `json:"language"`
	Theme                    string `json:"theme"`
	DefaultViewMode          string `json:"default_view_mode"`
	StartupOnBoot            bool   `json:"startup_on_boot"`
	CloseToTray              bool   `json:"close_to_tray"`
	ShowHiddenArticles       bool   `json:"show_hidden_articles"`
	HoverMarkAsRead          bool   `json:"hover_mark_as_read"`
	TranslationEnabled       bool   `json:"translation_enabled"`
	TargetLanguage           string `json:"target_language"`
	TranslationProvider      string `json:"translation_provider"`
	DeepLAPIKey              string `json:"deepl_api_key"`
	DeepLEndpoint            string `json:"deepl_endpoint"`
	BaiduAppID               string `json:"baidu_app_id"`
	BaiduSecretKey           string `json:"baidu_secret_key"`
	AIAPIKey                 string `json:"ai_api_key"`
	AIEndpoint               string `json:"ai_endpoint"`
	AIModel                  string `json:"ai_model"`
	AITranslationPrompt      string `json:"ai_translation_prompt"`
	AISummaryPrompt          string `json:"ai_summary_prompt"`
	AIUsageTokens            string `json:"ai_usage_tokens"`
	AIUsageLimit             string `json:"ai_usage_limit"`
	AIChatEnabled            bool   `json:"ai_chat_enabled"`
	SummaryEnabled           bool   `json:"summary_enabled"`
	SummaryLength            string `json:"summary_length"`
	SummaryProvider          string `json:"summary_provider"`
	SummaryTriggerMode       string `json:"summary_trigger_mode"`
	AutoCleanupEnabled       bool   `json:"auto_cleanup_enabled"`
	MaxCacheSizeMB           int    `json:"max_cache_size_mb"`
	MaxArticleAgeDays        int    `json:"max_article_age_days"`
	MediaCacheEnabled        bool   `json:"media_cache_enabled"`
	MediaCacheMaxSizeMB      int    `json:"media_cache_max_size_mb"`
	MediaCacheMaxAgeDays     int    `json:"media_cache_max_age_days"`
	ProxyEnabled             bool   `json:"proxy_enabled"`
	ProxyType                string `json:"proxy_type"`
	ProxyHost                string `json:"proxy_host"`
	ProxyPort                string `json:"proxy_port"`
	ProxyUsername            string `json:"proxy_username"`
	ProxyPassword            string `json:"proxy_password"`
	Shortcuts                string `json:"shortcuts"`
	Rules                    string `json:"rules"`
	LastArticleUpdate        string `json:"last_article_update"`
	GoogleTranslateEndpoint  string `json:"google_translate_endpoint"`
	ShowArticlePreviewImages bool   `json:"show_article_preview_images"`
	ObsidianEnabled          bool   `json:"obsidian_enabled"`
	ObsidianVault            string `json:"obsidian_vault"`
	ObsidianVaultPath        string `json:"obsidian_vault_path"`
	WindowX                  string `json:"window_x"`
	WindowY                  string `json:"window_y"`
	WindowWidth              string `json:"window_width"`
	WindowHeight             string `json:"window_height"`
	WindowMaximized          string `json:"window_maximized"`
	ImageGalleryEnabled      bool   `json:"image_gallery_enabled"`
	FreshRSSSyncEnabled      bool   `json:"freshrss_enabled"`
	FreshRSSServerURL        string `json:"freshrss_server_url"`
	FreshRSSUsername         string `json:"freshrss_username"`
	FreshRSSAPIPassword      string `json:"freshrss_api_password"`
	FullTextFetchEnabled     bool   `json:"full_text_fetch_enabled"`
	RSSHubEnabled            bool   `json:"rsshub_enabled"`
	RSSHubInstanceURL        string `json:"rsshub_instance_url"`
	RSSHubFallbackEnabled    bool   `json:"rsshub_fallback_enabled"`
	RSSHubAPIKey             string `json:"rsshub_api_key"`
}

var defaults Defaults

func init() {
	if err := json.Unmarshal(defaultsJSON, &defaults); err != nil {
		panic("failed to parse defaults.json: " + err.Error())
	}
}

// Get returns the loaded defaults
func Get() Defaults {
	return defaults
}

// GetString returns a setting default as a string
func GetString(key string) string {
	switch key {
	case "update_interval":
		return strconv.Itoa(defaults.UpdateInterval)
	case "refresh_mode":
		return defaults.RefreshMode
	case "language":
		return defaults.Language
	case "theme":
		return defaults.Theme
	case "default_view_mode":
		return defaults.DefaultViewMode
	case "startup_on_boot":
		return strconv.FormatBool(defaults.StartupOnBoot)
	case "close_to_tray":
		return strconv.FormatBool(defaults.CloseToTray)
	case "show_hidden_articles":
		return strconv.FormatBool(defaults.ShowHiddenArticles)
	case "hover_mark_as_read":
		return strconv.FormatBool(defaults.HoverMarkAsRead)
	case "translation_enabled":
		return strconv.FormatBool(defaults.TranslationEnabled)
	case "target_language":
		return defaults.TargetLanguage
	case "translation_provider":
		return defaults.TranslationProvider
	case "deepl_api_key":
		return defaults.DeepLAPIKey
	case "deepl_endpoint":
		return defaults.DeepLEndpoint
	case "baidu_app_id":
		return defaults.BaiduAppID
	case "baidu_secret_key":
		return defaults.BaiduSecretKey
	case "ai_api_key":
		return defaults.AIAPIKey
	case "ai_endpoint":
		return defaults.AIEndpoint
	case "ai_model":
		return defaults.AIModel
	case "ai_translation_prompt":
		return defaults.AITranslationPrompt
	case "ai_summary_prompt":
		return defaults.AISummaryPrompt
	case "ai_usage_tokens":
		return defaults.AIUsageTokens
	case "ai_usage_limit":
		return defaults.AIUsageLimit
	case "ai_chat_enabled":
		return strconv.FormatBool(defaults.AIChatEnabled)
	case "summary_enabled":
		return strconv.FormatBool(defaults.SummaryEnabled)
	case "summary_length":
		return defaults.SummaryLength
	case "summary_provider":
		return defaults.SummaryProvider
	case "summary_trigger_mode":
		return defaults.SummaryTriggerMode
	case "auto_cleanup_enabled":
		return strconv.FormatBool(defaults.AutoCleanupEnabled)
	case "max_cache_size_mb":
		return strconv.Itoa(defaults.MaxCacheSizeMB)
	case "max_article_age_days":
		return strconv.Itoa(defaults.MaxArticleAgeDays)
	case "media_cache_enabled":
		return strconv.FormatBool(defaults.MediaCacheEnabled)
	case "media_cache_max_size_mb":
		return strconv.Itoa(defaults.MediaCacheMaxSizeMB)
	case "media_cache_max_age_days":
		return strconv.Itoa(defaults.MediaCacheMaxAgeDays)
	case "proxy_enabled":
		return strconv.FormatBool(defaults.ProxyEnabled)
	case "proxy_type":
		return defaults.ProxyType
	case "proxy_host":
		return defaults.ProxyHost
	case "proxy_port":
		return defaults.ProxyPort
	case "proxy_username":
		return defaults.ProxyUsername
	case "proxy_password":
		return defaults.ProxyPassword
	case "shortcuts":
		return defaults.Shortcuts
	case "rules":
		return defaults.Rules
	case "last_article_update":
		return defaults.LastArticleUpdate
	case "google_translate_endpoint":
		return defaults.GoogleTranslateEndpoint
	case "window_x":
		return defaults.WindowX
	case "window_y":
		return defaults.WindowY
	case "window_width":
		return defaults.WindowWidth
	case "window_height":
		return defaults.WindowHeight
	case "window_maximized":
		return defaults.WindowMaximized
	case "show_article_preview_images":
		return strconv.FormatBool(defaults.ShowArticlePreviewImages)
	case "obsidian_enabled":
		return strconv.FormatBool(defaults.ObsidianEnabled)
	case "obsidian_vault":
		return defaults.ObsidianVault
	case "obsidian_vault_path":
		return defaults.ObsidianVaultPath
	case "image_gallery_enabled":
		return strconv.FormatBool(defaults.ImageGalleryEnabled)
	case "freshrss_enabled":
		return strconv.FormatBool(defaults.FreshRSSSyncEnabled)
	case "freshrss_server_url":
		return defaults.FreshRSSServerURL
	case "freshrss_username":
		return defaults.FreshRSSUsername
	case "freshrss_api_password":
		return defaults.FreshRSSAPIPassword
	case "full_text_fetch_enabled":
		return strconv.FormatBool(defaults.FullTextFetchEnabled)
	case "rsshub_enabled":
		return strconv.FormatBool(defaults.RSSHubEnabled)
	case "rsshub_instance_url":
		return defaults.RSSHubInstanceURL
	case "rsshub_fallback_enabled":
		return strconv.FormatBool(defaults.RSSHubFallbackEnabled)
	case "rsshub_api_key":
		return defaults.RSSHubAPIKey
	default:
		return ""
	}
}

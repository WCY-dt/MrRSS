package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"MrRSS/internal/utils"
)

// HandleSettings handles both GET and POST requests for settings
func (h *Handler) HandleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetSettings(w, r)
	case http.MethodPost:
		h.handleUpdateSettings(w, r)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *Handler) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings := map[string]string{
		"update_interval":      h.getSettingOrDefault("update_interval", "10"),
		"translation_enabled":  h.getSettingOrDefault("translation_enabled", "false"),
		"target_language":      h.getSettingOrDefault("target_language", "en"),
		"translation_provider": h.getSettingOrDefault("translation_provider", "google"),
		"deepl_api_key":        h.getSettingOrDefault("deepl_api_key", ""),
		"auto_cleanup_enabled": h.getSettingOrDefault("auto_cleanup_enabled", "false"),
		"max_cache_size_mb":    h.getSettingOrDefault("max_cache_size_mb", "20"),
		"max_article_age_days": h.getSettingOrDefault("max_article_age_days", "30"),
		"language":             h.getSettingOrDefault("language", "en"),
		"theme":                h.getSettingOrDefault("theme", "auto"),
		"last_article_update":  h.getSettingOrDefault("last_article_update", ""),
		"show_hidden_articles": h.getSettingOrDefault("show_hidden_articles", "false"),
		"startup_on_boot":      h.getSettingOrDefault("startup_on_boot", "false"),
	}
	respondWithJSON(w, http.StatusOK, settings)
}

func (h *Handler) getSettingOrDefault(key, defaultValue string) string {
	if val, err := h.DB.GetSetting(key); err == nil && val != "" {
		return val
	}
	return defaultValue
}

func (h *Handler) handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req settingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Apply all settings
	h.applySettingIfNotEmpty("update_interval", req.UpdateInterval)
	h.applySettingIfNotEmpty("translation_enabled", req.TranslationEnabled)
	h.applySettingIfNotEmpty("target_language", req.TargetLanguage)
	h.applySettingIfNotEmpty("translation_provider", req.TranslationProvider)
	h.applySettingIfNotEmpty("auto_cleanup_enabled", req.AutoCleanupEnabled)
	h.applySettingIfNotEmpty("max_cache_size_mb", req.MaxCacheSizeMB)
	h.applySettingIfNotEmpty("max_article_age_days", req.MaxArticleAgeDays)
	h.applySettingIfNotEmpty("language", req.Language)
	h.applySettingIfNotEmpty("theme", req.Theme)
	h.applySettingIfNotEmpty("show_hidden_articles", req.ShowHiddenArticles)

	// Always update API key (can be cleared)
	h.DB.SetSetting("deepl_api_key", req.DeepLAPIKey)

	// Handle startup setting with application
	if req.StartupOnBoot != "" {
		h.applyStartupSetting(req.StartupOnBoot)
	}

	respondOK(w)
}

func (h *Handler) applySettingIfNotEmpty(key, value string) {
	if value != "" {
		h.DB.SetSetting(key, value)
	}
}

func (h *Handler) applyStartupSetting(value string) {
	currentValue, err := h.DB.GetSetting("startup_on_boot")
	if err != nil {
		log.Printf("Failed to get startup_on_boot setting: %v", err)
		h.DB.SetSetting("startup_on_boot", value)
		return
	}

	if currentValue == value {
		return // No change needed
	}

	h.DB.SetSetting("startup_on_boot", value)

	// Apply the startup setting
	if value == "true" {
		if err := utils.EnableStartup(); err != nil {
			log.Printf("Failed to enable startup: %v", err)
		}
	} else {
		if err := utils.DisableStartup(); err != nil {
			log.Printf("Failed to disable startup: %v", err)
		}
	}
}

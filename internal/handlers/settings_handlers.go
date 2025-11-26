package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"MrRSS/internal/utils"
)

// HandleSettings handles GET and POST requests for application settings.
func (h *Handler) HandleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		interval, _ := h.DB.GetSetting("update_interval")
		translationEnabled, _ := h.DB.GetSetting("translation_enabled")
		targetLang, _ := h.DB.GetSetting("target_language")
		provider, _ := h.DB.GetSetting("translation_provider")
		apiKey, _ := h.DB.GetSetting("deepl_api_key")
		autoCleanup, _ := h.DB.GetSetting("auto_cleanup_enabled")
		maxCacheSize, _ := h.DB.GetSetting("max_cache_size_mb")
		maxArticleAge, _ := h.DB.GetSetting("max_article_age_days")
		language, _ := h.DB.GetSetting("language")
		theme, _ := h.DB.GetSetting("theme")
		lastUpdate, _ := h.DB.GetSetting("last_article_update")
		showHidden, _ := h.DB.GetSetting("show_hidden_articles")
		startupOnBoot, _ := h.DB.GetSetting("startup_on_boot")
		shortcuts, _ := h.DB.GetSetting("shortcuts")
		json.NewEncoder(w).Encode(map[string]string{
			"update_interval":      interval,
			"translation_enabled":  translationEnabled,
			"target_language":      targetLang,
			"translation_provider": provider,
			"deepl_api_key":        apiKey,
			"auto_cleanup_enabled": autoCleanup,
			"max_cache_size_mb":    maxCacheSize,
			"max_article_age_days": maxArticleAge,
			"language":             language,
			"theme":                theme,
			"last_article_update":  lastUpdate,
			"show_hidden_articles": showHidden,
			"startup_on_boot":      startupOnBoot,
			"shortcuts":            shortcuts,
		})
	} else if r.Method == http.MethodPost {
		var req struct {
			UpdateInterval      string `json:"update_interval"`
			TranslationEnabled  string `json:"translation_enabled"`
			TargetLanguage      string `json:"target_language"`
			TranslationProvider string `json:"translation_provider"`
			DeepLAPIKey         string `json:"deepl_api_key"`
			AutoCleanupEnabled  string `json:"auto_cleanup_enabled"`
			MaxCacheSizeMB      string `json:"max_cache_size_mb"`
			MaxArticleAgeDays   string `json:"max_article_age_days"`
			Language            string `json:"language"`
			Theme               string `json:"theme"`
			ShowHiddenArticles  string `json:"show_hidden_articles"`
			StartupOnBoot       string `json:"startup_on_boot"`
			Shortcuts           string `json:"shortcuts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.UpdateInterval != "" {
			h.DB.SetSetting("update_interval", req.UpdateInterval)
		}
		if req.TranslationEnabled != "" {
			h.DB.SetSetting("translation_enabled", req.TranslationEnabled)
		}
		if req.TargetLanguage != "" {
			h.DB.SetSetting("target_language", req.TargetLanguage)
		}
		if req.TranslationProvider != "" {
			h.DB.SetSetting("translation_provider", req.TranslationProvider)
		}
		// Always update API key as it might be cleared
		h.DB.SetSetting("deepl_api_key", req.DeepLAPIKey)

		if req.AutoCleanupEnabled != "" {
			h.DB.SetSetting("auto_cleanup_enabled", req.AutoCleanupEnabled)
		}

		if req.MaxCacheSizeMB != "" {
			h.DB.SetSetting("max_cache_size_mb", req.MaxCacheSizeMB)
		}

		if req.MaxArticleAgeDays != "" {
			h.DB.SetSetting("max_article_age_days", req.MaxArticleAgeDays)
		}

		if req.Language != "" {
			h.DB.SetSetting("language", req.Language)
		}

		if req.Theme != "" {
			h.DB.SetSetting("theme", req.Theme)
		}

		if req.ShowHiddenArticles != "" {
			h.DB.SetSetting("show_hidden_articles", req.ShowHiddenArticles)
		}

		// Always update shortcuts as it might be cleared or modified
		h.DB.SetSetting("shortcuts", req.Shortcuts)

		if req.StartupOnBoot != "" {
			// Get current value to check if it changed
			currentValue, err := h.DB.GetSetting("startup_on_boot")
			if err != nil {
				log.Printf("Failed to get startup_on_boot setting: %v", err)
				// If we can't read the current value, save the new value but don't apply it
				h.DB.SetSetting("startup_on_boot", req.StartupOnBoot)
			} else if currentValue != req.StartupOnBoot {
				// Only apply if the value changed
				h.DB.SetSetting("startup_on_boot", req.StartupOnBoot)

				// Apply the startup setting
				if req.StartupOnBoot == "true" {
					if err := utils.EnableStartup(); err != nil {
						log.Printf("Failed to enable startup: %v", err)
					}
				} else {
					if err := utils.DisableStartup(); err != nil {
						log.Printf("Failed to disable startup: %v", err)
					}
				}
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

package translation

import (
	"encoding/json"
	"log"
	"net/http"

	"MrRSS/internal/aiusage"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/translation"
	"MrRSS/internal/utils"
)

// HandleTranslateArticle translates an article's title.
// @Summary      Translate article title
// @Description  Translate an article's title to the target language (uses AI or Google based on settings)
// @Tags         translation
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Translation request (article_id, title, target_language)"
// @Success      200  {object}  map[string]interface{}  "Translation result (translated_title, limit_reached)"
// @Failure      400  {object}  map[string]string  "Bad request (missing required fields)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /translate/article [post]
func HandleTranslateArticle(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ArticleID  int64  `json:"article_id"`
		Title      string `json:"title"`
		TargetLang string `json:"target_language"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.TargetLang == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Check if we should use AI translation or fallback to Google
	provider, _ := h.DB.GetSetting("translation_provider")
	isAIProvider := provider == "ai"

	var translatedTitle string
	var err error
	var limitReached = false

	if isAIProvider {
		// Check if AI usage limit is reached
		if h.AITracker.IsLimitReached() {
			log.Printf("AI usage limit reached, falling back to Google Translate")
			limitReached = true
			// Fallback to Google Translate
			googleTranslator := translation.NewGoogleFreeTranslatorWithDB(h.DB)
			translatedTitle, err = googleTranslator.Translate(req.Title, req.TargetLang)
		} else {
			// Apply rate limiting for AI requests
			h.AITracker.WaitForRateLimit()

			// Try AI translation first
			translatedTitle, err = h.Translator.Translate(req.Title, req.TargetLang)

			// If AI fails, fallback to Google Translate
			if err != nil {
				log.Printf("AI translation failed, falling back to Google Translate: %v", err)
				googleTranslator := translation.NewGoogleFreeTranslatorWithDB(h.DB)
				translatedTitle, err = googleTranslator.Translate(req.Title, req.TargetLang)
			}

			// Track AI usage only on success (whether AI or fallback)
			if err == nil {
				h.AITracker.TrackTranslation(req.Title, translatedTitle)
			}
		}
	} else {
		// Non-AI provider, no special handling needed
		translatedTitle, err = h.Translator.Translate(req.Title, req.TargetLang)
	}

	if err != nil {
		log.Printf("Error translating article %d: %v", req.ArticleID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the article with the translated title
	if err := h.DB.UpdateArticleTranslation(req.ArticleID, translatedTitle); err != nil {
		log.Printf("Error updating article translation: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"translated_title": translatedTitle,
		"limit_reached":    limitReached,
	})
}

// HandleClearTranslations clears all translated titles from the database.
// @Summary      Clear all translations
// @Description  Clear all translated article titles from the database
// @Tags         translation
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]bool  "Success status"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /translations/clear [post]
func HandleClearTranslations(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.DB.ClearAllTranslations(); err != nil {
		log.Printf("Error clearing translations: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleTranslateText translates any text to the target language.
// This is used for translating content, summaries, etc.
// @Summary      Translate text
// @Description  Translate any text to the target language (uses AI or Google based on settings)
// @Tags         translation
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Translation request (text, target_language)"
// @Success      200  {object}  map[string]string  "Translation result (translated_text, html)"
// @Failure      400  {object}  map[string]string  "Bad request (missing required fields)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /translate/text [post]
func HandleTranslateText(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Text       string `json:"text"`
		TargetLang string `json:"target_language"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Text == "" || req.TargetLang == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Check if we should use AI translation or fallback to Google
	provider, _ := h.DB.GetSetting("translation_provider")
	isAIProvider := provider == "ai"

	var translatedText string
	var err error

	if isAIProvider {
		// Check if AI usage limit is reached
		if h.AITracker.IsLimitReached() {
			log.Printf("AI usage limit reached, falling back to Google Translate")
			// Fallback to Google Translate
			googleTranslator := translation.NewGoogleFreeTranslatorWithDB(h.DB)
			translatedText, err = googleTranslator.Translate(req.Text, req.TargetLang)
		} else {
			// Apply rate limiting for AI requests
			h.AITracker.WaitForRateLimit()

			// Try AI translation first
			translatedText, err = h.Translator.Translate(req.Text, req.TargetLang)

			// If AI fails, fallback to Google Translate
			if err != nil {
				log.Printf("AI translation failed, falling back to Google Translate: %v", err)
				googleTranslator := translation.NewGoogleFreeTranslatorWithDB(h.DB)
				translatedText, err = googleTranslator.Translate(req.Text, req.TargetLang)
			}

			// Track AI usage only on success (whether AI or fallback)
			if err == nil {
				h.AITracker.TrackTranslation(req.Text, translatedText)
			}
		}
	} else {
		// Non-AI provider, no special handling needed
		translatedText, err = h.Translator.Translate(req.Text, req.TargetLang)
	}

	if err != nil {
		log.Printf("Error translating text: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert translated markdown to HTML
	htmlText := utils.ConvertMarkdownToHTML(translatedText)

	json.NewEncoder(w).Encode(map[string]string{
		"translated_text": translatedText,
		"html":            htmlText,
	})
}

// HandleResetAIUsage resets the AI usage counter.
// @Summary      Reset AI usage counter
// @Description  Reset the AI usage token counter to zero
// @Tags         translation
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]bool  "Success status"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /ai/usage/reset [post]
func HandleResetAIUsage(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.AITracker.ResetUsage(); err != nil {
		log.Printf("Error resetting AI usage: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleGetAIUsage returns the current AI usage statistics.
// @Summary      Get AI usage statistics
// @Description  Get current AI usage (tokens used, limit, and whether limit is reached)
// @Tags         translation
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "AI usage stats (usage, limit, limit_reached)"
// @Router       /ai/usage [get]
func HandleGetAIUsage(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	usage, _ := h.AITracker.GetCurrentUsage()
	limit, _ := h.AITracker.GetUsageLimit()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"usage":         usage,
		"limit":         limit,
		"limit_reached": h.AITracker.IsLimitReached(),
	})
}

// EstimateTokens exposes the token estimation function for testing/display.
func EstimateTokens(text string) int64 {
	return aiusage.EstimateTokens(text)
}

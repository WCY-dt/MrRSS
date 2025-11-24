package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// HandleTranslateArticle translates an article title
func (h *Handler) HandleTranslateArticle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req translateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Title == "" || req.TargetLang == "" {
		respondWithError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	// Translate the title
	translatedTitle, err := h.Translator.Translate(req.Title, req.TargetLang)
	if err != nil {
		log.Printf("Error translating article %d: %v", req.ArticleID, err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Update the article with the translated title
	if err := h.DB.UpdateArticleTranslation(req.ArticleID, translatedTitle); err != nil {
		log.Printf("Error updating article translation: %v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"translated_title": translatedTitle,
	})
}

// HandleClearTranslations clears all translated titles from the database
func (h *Handler) HandleClearTranslations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := h.DB.ClearAllTranslations(); err != nil {
		log.Printf("Error clearing translations: %v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]bool{"success": true})
}

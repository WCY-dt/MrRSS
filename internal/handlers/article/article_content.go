package article

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"MrRSS/internal/handlers/core"
)

// HandleGetArticleContent fetches the article content from RSS feed dynamically.
func HandleGetArticleContent(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	articleIDStr := r.URL.Query().Get("id")
	articleID, err := strconv.ParseInt(articleIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	// Get the article from database to access feed_id
	article, err := h.DB.GetArticleByID(articleID)
	if err != nil {
		log.Printf("Error getting article: %v", err)
		http.Error(w, "Failed to get article", http.StatusInternalServerError)
		return
	}

	// Use the cached content fetching method
	content, err := h.GetArticleContent(articleID)
	if err != nil {
		log.Printf("Error getting article content: %v", err)
		http.Error(w, "Failed to fetch article content", http.StatusInternalServerError)
		return
	}

	// Get feed URL to use as referer for image proxying
	feed, err := h.DB.GetFeedByID(article.FeedID)
	var feedURL string
	if err == nil && feed != nil {
		feedURL = feed.URL
	}

	json.NewEncoder(w).Encode(map[string]string{
		"content":  content,
		"feed_url": feedURL,
	})
}

// HandleFetchFullArticle fetches the full article content from the original URL using readability.
func HandleFetchFullArticle(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	articleIDStr := r.URL.Query().Get("id")
	articleID, err := strconv.ParseInt(articleIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	// Get the article from database
	article, err := h.DB.GetArticleByID(articleID)
	if err != nil {
		log.Printf("Error getting article: %v", err)
		http.Error(w, "Failed to get article", http.StatusInternalServerError)
		return
	}

	if article.URL == "" {
		http.Error(w, "Article has no URL", http.StatusBadRequest)
		return
	}

	// Check if full-text fetching is enabled (global setting only)
	// auto_expand_content only affects auto-expansion behavior, not manual button clicks
	fullTextEnabledStr, _ := h.DB.GetSetting("full_text_fetch_enabled")
	if fullTextEnabledStr != "true" {
		http.Error(w, "Full-text fetching is disabled", http.StatusForbidden)
		return
	}

	// Fetch full content
	fullContent, err := h.FetchFullArticleContent(article.URL)
	if err != nil {
		log.Printf("Error fetching full article content: %v", err)
		http.Error(w, "Failed to fetch full article content", http.StatusInternalServerError)
		return
	}

	// Get feed URL to use as referer for image proxying
	feed, err := h.DB.GetFeedByID(article.FeedID)
	var feedURL string
	if err == nil && feed != nil {
		feedURL = feed.URL
	}

	json.NewEncoder(w).Encode(map[string]string{
		"content":  fullContent,
		"feed_url": feedURL,
	})
}

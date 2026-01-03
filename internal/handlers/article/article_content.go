package article

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"MrRSS/internal/handlers/core"
)

// HandleGetArticleContent fetches the article content from RSS feed dynamically.
// @Summary      Get article content
// @Description  Fetch the full HTML content of an article (uses cache if available)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id   query     int64   true  "Article ID"
// @Success      200  {object}  map[string]string  "Article content (content, feed_url)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid article ID)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/content [get]
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
// @Summary      Fetch full article content
// @Description  Fetch the full article content from the original URL using readability extraction (requires full_text_fetch_enabled setting)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id   query     int64   true  "Article ID"
// @Success      200  {object}  map[string]string  "Full article content (content, feed_url)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid ID or missing URL)"
// @Failure      403  {object}  map[string]string  "Full-text fetching disabled"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/fetch-full [post]
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

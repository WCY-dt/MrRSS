package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"MrRSS/internal/models"

	"github.com/mmcdole/gofeed"
)

// HandleArticles returns articles based on filters
func (h *Handler) HandleArticles(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	category := r.URL.Query().Get("category")

	var feedID int64
	if feedIDStr := r.URL.Query().Get("feed_id"); feedIDStr != "" {
		feedID, _ = strconv.ParseInt(feedIDStr, 10, 64)
	}

	page := parseIntQueryParam(r, "page", defaultArticlesPerPage)
	limit := parseIntQueryParam(r, "limit", defaultArticleLimit)
	offset := (page - 1) * limit

	// Get show_hidden_articles setting
	showHiddenStr, _ := h.DB.GetSetting("show_hidden_articles")
	showHidden := showHiddenStr == "true"

	articles, err := h.DB.GetArticles(filter, feedID, category, showHidden, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, articles)
}

// HandleMarkRead marks an article as read or unread
func (h *Handler) HandleMarkRead(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64QueryParam(r, "id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	read := parseBoolQueryParam(r, "read", true)

	if err := h.DB.MarkArticleRead(id, read); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(w)
}

// HandleToggleFavorite toggles the favorite status of an article
func (h *Handler) HandleToggleFavorite(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64QueryParam(r, "id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.DB.ToggleFavorite(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(w)
}

// HandleMarkAllAsRead marks all articles as read
func (h *Handler) HandleMarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	feedIDStr := r.URL.Query().Get("feed_id")

	var err error
	if feedIDStr != "" {
		// Mark all as read for a specific feed
		feedID, parseErr := strconv.ParseInt(feedIDStr, 10, 64)
		if parseErr != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid feed_id parameter")
			return
		}
		err = h.DB.MarkAllAsReadForFeed(feedID)
	} else {
		// Mark all as read globally
		err = h.DB.MarkAllAsRead()
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondOK(w)
}

// HandleToggleHideArticle toggles the hidden status of an article
func (h *Handler) HandleToggleHideArticle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := parseInt64QueryParam(r, "id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.DB.ToggleArticleHidden(id); err != nil {
		log.Printf("Error toggling article hidden status: %v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// HandleGetArticleContent fetches the article content from RSS feed dynamically
func (h *Handler) HandleGetArticleContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	articleID, err := parseInt64QueryParam(r, "id")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Get the article
	article, err := h.getArticleByID(articleID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Article not found")
		return
	}

	// Get the feed URL
	feedURL, err := h.getFeedURL(article.FeedID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Feed not found")
		return
	}

	// Parse the feed to get fresh content
	ctx, cancel := context.WithTimeout(context.Background(), feedFetchTimeout)
	defer cancel()

	parsedFeed, err := h.Fetcher.ParseFeed(ctx, feedURL)
	if err != nil {
		log.Printf("Error parsing feed for article content: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch article content")
		return
	}

	// Find the article in the feed by URL
	content := h.findArticleContent(parsedFeed, article.URL)

	respondWithJSON(w, http.StatusOK, map[string]string{
		"content": content,
	})
}

func (h *Handler) getArticleByID(articleID int64) (*models.Article, error) {
	// Get all articles to find the one we need
	allArticles, err := h.DB.GetArticles("", 0, "", false, 1000, 0)
	if err != nil {
		return nil, err
	}

	for i := range allArticles {
		if allArticles[i].ID == articleID {
			return &allArticles[i], nil
		}
	}

	return nil, fmt.Errorf("article not found")
}

func (h *Handler) getFeedURL(feedID int64) (string, error) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		return "", err
	}

	for i := range feeds {
		if feeds[i].ID == feedID {
			return feeds[i].URL, nil
		}
	}

	return "", fmt.Errorf("feed not found")
}

func (h *Handler) findArticleContent(parsedFeed *gofeed.Feed, articleURL string) string {
	for _, item := range parsedFeed.Items {
		if item.Link == articleURL {
			if item.Content != "" {
				return item.Content
			}
			return item.Description
		}
	}
	return ""
}

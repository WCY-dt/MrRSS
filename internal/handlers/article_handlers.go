package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"MrRSS/internal/models"
)

// FilterCondition represents a single filter condition from the frontend
type FilterCondition struct {
	ID       int64    `json:"id"`
	Logic    string   `json:"logic"`    // "and", "or" (null for first condition)
	Negate   bool     `json:"negate"`   // NOT modifier for this condition
	Field    string   `json:"field"`    // "feed_name", "feed_category", "article_title", "published_after", "published_before"
	Operator string   `json:"operator"` // "contains", "exact" (null for date fields and multi-select)
	Value    string   `json:"value"`    // Single value for text/date fields
	Values   []string `json:"values"`   // Multiple values for feed_name and feed_category
}

// FilterRequest represents the request body for filtered articles
type FilterRequest struct {
	Conditions []FilterCondition `json:"conditions"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
}

// FilterResponse represents the response for filtered articles with pagination info
type FilterResponse struct {
	Articles   []models.Article `json:"articles"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	HasMore    bool             `json:"has_more"`
}

// HandleArticles returns articles with filtering and pagination.
func (h *Handler) HandleArticles(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	feedIDStr := r.URL.Query().Get("feed_id")
	category := r.URL.Query().Get("category")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	var feedID int64
	if feedIDStr != "" {
		feedID, _ = strconv.ParseInt(feedIDStr, 10, 64)
	}

	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := (page - 1) * limit

	// Get show_hidden_articles setting
	showHiddenStr, _ := h.DB.GetSetting("show_hidden_articles")
	showHidden := showHiddenStr == "true"

	articles, err := h.DB.GetArticles(filter, feedID, category, showHidden, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(articles)
}

// HandleProgress returns the current fetch progress.
func (h *Handler) HandleProgress(w http.ResponseWriter, r *http.Request) {
	progress := h.Fetcher.GetProgress()
	json.NewEncoder(w).Encode(progress)
}

// HandleMarkRead marks an article as read or unread.
func (h *Handler) HandleMarkRead(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	readStr := r.URL.Query().Get("read")
	read := true
	if readStr == "false" || readStr == "0" {
		read = false
	}

	if err := h.DB.MarkArticleRead(id, read); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandleToggleFavorite toggles the favorite status of an article.
func (h *Handler) HandleToggleFavorite(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.DB.ToggleFavorite(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandleGetUnreadCounts returns unread counts for all feeds.
func (h *Handler) HandleGetUnreadCounts(w http.ResponseWriter, r *http.Request) {
	// Get total unread count
	totalCount, err := h.DB.GetTotalUnreadCount()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get unread counts per feed
	feedCounts, err := h.DB.GetUnreadCountsForAllFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"total":       totalCount,
		"feed_counts": feedCounts,
	}
	json.NewEncoder(w).Encode(response)
}

// HandleMarkAllAsRead marks all articles as read.
func (h *Handler) HandleMarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	feedIDStr := r.URL.Query().Get("feed_id")

	var err error
	if feedIDStr != "" {
		// Mark all as read for a specific feed
		feedID, parseErr := strconv.ParseInt(feedIDStr, 10, 64)
		if parseErr != nil {
			http.Error(w, "Invalid feed_id parameter", http.StatusBadRequest)
			return
		}
		err = h.DB.MarkAllAsReadForFeed(feedID)
	} else {
		// Mark all as read globally
		err = h.DB.MarkAllAsRead()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandleRefresh triggers a refresh of all feeds.
func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	go h.Fetcher.FetchAll(context.Background())
	w.WriteHeader(http.StatusOK)
}

// HandleCleanupArticles triggers manual cleanup of articles.
func (h *Handler) HandleCleanupArticles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := h.DB.CleanupUnimportantArticles()
	if err != nil {
		log.Printf("Error cleaning up articles: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Cleaned up %d articles", count)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted": count,
	})
}

// HandleToggleHideArticle toggles the hidden status of an article.
func (h *Handler) HandleToggleHideArticle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	if err := h.DB.ToggleArticleHidden(id); err != nil {
		log.Printf("Error toggling article hidden status: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleGetArticleContent fetches the article content from RSS feed dynamically.
func (h *Handler) HandleGetArticleContent(w http.ResponseWriter, r *http.Request) {
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

	// Get all articles to find the one we need
	allArticles, err := h.DB.GetArticles("", 0, "", false, 1000, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var article *models.Article
	for i := range allArticles {
		if allArticles[i].ID == articleID {
			article = &allArticles[i]
			break
		}
	}

	if article == nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	// Get the feed to fetch fresh content
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var feedURL string
	for i := range feeds {
		if feeds[i].ID == article.FeedID {
			feedURL = feeds[i].URL
			break
		}
	}

	if feedURL == "" {
		http.Error(w, "Feed not found", http.StatusNotFound)
		return
	}

	// Parse the feed to get fresh content
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	parsedFeed, err := h.Fetcher.ParseFeed(ctx, feedURL)
	if err != nil {
		log.Printf("Error parsing feed for article content: %v", err)
		http.Error(w, "Failed to fetch article content", http.StatusInternalServerError)
		return
	}

	// Find the article in the feed by URL
	var content string
	for _, item := range parsedFeed.Items {
		if item.Link == article.URL {
			content = item.Content
			if content == "" {
				content = item.Description
			}
			break
		}
	}

	json.NewEncoder(w).Encode(map[string]string{
		"content": content,
	})
}

// HandleFilteredArticles returns articles filtered by advanced conditions from the database.
func (h *Handler) HandleFilteredArticles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set default pagination values
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 50
	}

	// Get show_hidden_articles setting
	showHiddenStr, _ := h.DB.GetSetting("show_hidden_articles")
	showHidden := showHiddenStr == "true"

	// Get all articles from database
	// Note: Using a high limit to fetch all articles for filtering
	// For very large datasets, consider implementing database-level filtering
	articles, err := h.DB.GetArticles("", 0, "", showHidden, 50000, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get feeds for category lookup
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a map of feed ID to category
	feedCategories := make(map[int64]string)
	for _, feed := range feeds {
		feedCategories[feed.ID] = feed.Category
	}

	// Apply filter conditions
	if len(req.Conditions) > 0 {
		var filteredArticles []models.Article
		for _, article := range articles {
			if evaluateArticleConditions(article, req.Conditions, feedCategories) {
				filteredArticles = append(filteredArticles, article)
			}
		}
		articles = filteredArticles
	}

	// Apply pagination
	total := len(articles)
	offset := (page - 1) * limit
	end := offset + limit

	// Handle edge cases for pagination
	var paginatedArticles []models.Article
	if offset >= total {
		// No more articles to show
		paginatedArticles = []models.Article{}
	} else {
		if end > total {
			end = total
		}
		paginatedArticles = articles[offset:end]
	}
	
	hasMore := end < total

	response := FilterResponse{
		Articles: paginatedArticles,
		Total:    total,
		Page:     page,
		Limit:    limit,
		HasMore:  hasMore,
	}

	json.NewEncoder(w).Encode(response)
}

// evaluateArticleConditions evaluates all filter conditions for an article
func evaluateArticleConditions(article models.Article, conditions []FilterCondition, feedCategories map[int64]string) bool {
	if len(conditions) == 0 {
		return true
	}

	result := evaluateSingleCondition(article, conditions[0], feedCategories)

	for i := 1; i < len(conditions); i++ {
		condition := conditions[i]
		conditionResult := evaluateSingleCondition(article, condition, feedCategories)

		switch condition.Logic {
		case "and":
			result = result && conditionResult
		case "or":
			result = result || conditionResult
		}
	}

	return result
}

// matchMultiSelectContains checks if fieldValue matches any of the selected values using contains logic
func matchMultiSelectContains(fieldValue string, values []string, singleValue string) bool {
	if len(values) > 0 {
		lowerField := strings.ToLower(fieldValue)
		for _, val := range values {
			if strings.Contains(lowerField, strings.ToLower(val)) {
				return true
			}
		}
		return false
	} else if singleValue != "" {
		return strings.Contains(strings.ToLower(fieldValue), strings.ToLower(singleValue))
	}
	return true
}

// evaluateSingleCondition evaluates a single filter condition for an article
func evaluateSingleCondition(article models.Article, condition FilterCondition, feedCategories map[int64]string) bool {
	var result bool

	switch condition.Field {
	case "feed_name":
		result = matchMultiSelectContains(article.FeedTitle, condition.Values, condition.Value)

	case "feed_category":
		feedCategory := feedCategories[article.FeedID]
		result = matchMultiSelectContains(feedCategory, condition.Values, condition.Value)

	case "article_title":
		if condition.Value == "" {
			result = true
		} else {
			lowerValue := strings.ToLower(condition.Value)
			lowerTitle := strings.ToLower(article.Title)
			if condition.Operator == "exact" {
				result = lowerTitle == lowerValue
			} else {
				result = strings.Contains(lowerTitle, lowerValue)
			}
		}

	case "published_after":
		if condition.Value == "" {
			result = true
		} else {
			afterDate, err := time.Parse("2006-01-02", condition.Value)
			if err != nil {
				log.Printf("Invalid date format for published_after filter: %s", condition.Value)
				result = true
			} else {
				result = article.PublishedAt.After(afterDate) || article.PublishedAt.Equal(afterDate)
			}
		}

	case "published_before":
		if condition.Value == "" {
			result = true
		} else {
			beforeDate, err := time.Parse("2006-01-02", condition.Value)
			if err != nil {
				log.Printf("Invalid date format for published_before filter: %s", condition.Value)
				result = true
			} else {
				// For "before Dec 24 (inclusive)", we want articles published on Dec 24 or earlier
				// We compare dates only (not times) - any article from Dec 24 should be included
				// Truncate to remove time component, preserving date in local timezone context
				articleDateOnly := article.PublishedAt.UTC().Truncate(24 * time.Hour)
				beforeDateOnly := beforeDate.Truncate(24 * time.Hour)
				// Include articles on the selected date or before
				result = !articleDateOnly.After(beforeDateOnly)
			}
		}

	case "is_read":
		// Filter by read/unread status
		if condition.Value == "" {
			result = true
		} else {
			wantRead := condition.Value == "true"
			result = article.IsRead == wantRead
		}

	case "is_favorite":
		// Filter by favorite/unfavorite status
		if condition.Value == "" {
			result = true
		} else {
			wantFavorite := condition.Value == "true"
			result = article.IsFavorite == wantFavorite
		}

	default:
		result = true
	}

	// Apply NOT modifier
	if condition.Negate {
		return !result
	}
	return result
}

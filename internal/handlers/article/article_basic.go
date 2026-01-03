package article

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"MrRSS/internal/handlers/core"
)

// HandleArticles returns articles with filtering and pagination.
// @Summary      Get articles with filtering
// @Description  Retrieve articles with optional filtering by feed, category, status, and pagination
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        filter    query     string  false  "Filter: 'all', 'unread', 'favorite', 'read_later'"  Enums(all, unread, favorite, read_later)
// @Param        feed_id   query     int64   false  "Filter by feed ID"
// @Param        category  query     string  false  "Filter by category name"
// @Param        page      query     int     false  "Page number (default: 1)"  minimum(1)
// @Param        limit     query     int     false  "Items per page (default: 50, max: 500)"  minimum(1)  maximum(500)
// @Success      200  {array}   models.Article  "List of articles"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles [get]
func HandleArticles(h *core.Handler, w http.ResponseWriter, r *http.Request) {
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

// HandleMarkRead marks an article as read or unread.
// @Summary      Mark article as read/unread
// @Description  Mark a specific article as read or unread
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id   query     int64   true  "Article ID"
// @Param        read query     string  true  "Read status: 'true', '1', 'false', or '0'"  Enums(true, 1, false, 0)
// @Success      200  {string}  string  "Article marked successfully"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/mark-read [post]
func HandleMarkRead(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	readStr := r.URL.Query().Get("read")
	read := true
	if readStr == "false" || readStr == "0" {
		read = false
	}

	syncReq, err := h.DB.MarkArticleReadWithSync(id, read)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

	// Trigger background sync if needed (non-blocking)
	if syncReq != nil {
		// Note: This is a simplified version. For production, use the new HandleMarkReadWithImmediateSync
		_ = syncReq
	}
}

// HandleToggleFavorite toggles the favorite status of an article.
// @Summary      Toggle article favorite status
// @Description  Toggle the favorite/starred status of an article
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id   query     int64   true  "Article ID"
// @Success      200  {string}  string  "Favorite status toggled successfully"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/toggle-favorite [post]
func HandleToggleFavorite(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	syncReq, err := h.DB.ToggleFavoriteWithSync(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

	// Trigger background sync if needed (non-blocking)
	if syncReq != nil {
		// Note: This is a simplified version. For production, use the new HandleToggleFavoriteWithImmediateSync
		_ = syncReq
	}
}

// HandleToggleHideArticle toggles the hidden status of an article.
// @Summary      Toggle article hidden status
// @Description  Toggle the hidden status of an article (hidden articles are filtered out by default)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id   query     int64   true  "Article ID"
// @Success      200  {object}  map[string]bool  "Success status"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/toggle-hide [post]
func HandleToggleHideArticle(h *core.Handler, w http.ResponseWriter, r *http.Request) {
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

// HandleToggleReadLater toggles the read later status of an article.
// @Summary      Toggle article read-later status
// @Description  Toggle the read-later status of an article (add to/remove from reading list)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id   query     int64   true  "Article ID"
// @Success      200  {object}  map[string]bool  "Success status"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/toggle-read-later [post]
func HandleToggleReadLater(h *core.Handler, w http.ResponseWriter, r *http.Request) {
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

	if err := h.DB.ToggleReadLater(id); err != nil {
		log.Printf("Error toggling article read later status: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// HandleImageGalleryArticles returns articles from image mode feeds with pagination.
// @Summary      Get image gallery articles
// @Description  Retrieve articles from image-mode feeds (visual/rss-gallery feeds) with pagination
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        feed_id  query     int64   false  "Filter by feed ID"
// @Param        page     query     int     false  "Page number (default: 1)"  minimum(1)
// @Param        limit    query     int     false  "Items per page (default: 50)"  minimum(1)
// @Success      200  {array}   models.Article  "List of image gallery articles"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/image-gallery [get]
func HandleImageGalleryArticles(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	feedIDStr := r.URL.Query().Get("feed_id")

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

	articles, err := h.DB.GetImageGalleryArticles(feedID, showHidden, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(articles)
}

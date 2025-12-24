package article_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"MrRSS/internal/database"
	ff "MrRSS/internal/feed"
	"MrRSS/internal/handlers/article"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
)

func setupHandler(t *testing.T) *core.Handler {
	t.Helper()
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	f := ff.NewFetcher(db, nil)
	return core.NewHandler(db, f, nil)
}

func TestHandleArticles_ListAndImageGallery(t *testing.T) {
	h := setupHandler(t)

	// Add a feed and articles
	feedID, err := h.DB.AddFeed(&models.Feed{Title: "F", URL: "http://x"})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	articles := []*models.Article{
		{FeedID: feedID, Title: "a1", URL: "u1", PublishedAt: time.Now()},
		{FeedID: feedID, Title: "a2", URL: "u2", PublishedAt: time.Now()},
	}
	if err := h.DB.SaveArticles(context.Background(), articles); err != nil {
		t.Fatalf("SaveArticles: %v", err)
	}

	// Call HandleArticles
	req := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
	w := httptest.NewRecorder()
	article.HandleArticles(h, w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Result().StatusCode)
	}
	var got []models.Article
	if err := json.NewDecoder(w.Result().Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) < 2 {
		t.Fatalf("expected >=2 articles, got %d", len(got))
	}

	// Image gallery: mark feed as image mode and add image article
	if err := h.DB.UpdateFeed(feedID, "F", "http://x", "", "", false, "", false, 0, true, false, "", "", "", "", "", "", "", "", "", "", "", "", ""); err != nil {
		t.Fatalf("UpdateFeed: %v", err)
	}
	imgArticle := &models.Article{FeedID: feedID, Title: "img", URL: "iu", ImageURL: "http://img", PublishedAt: time.Now()}
	if err := h.DB.SaveArticles(context.Background(), []*models.Article{imgArticle}); err != nil {
		t.Fatalf("SaveArticles img: %v", err)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/articles/image_gallery", nil)
	w2 := httptest.NewRecorder()
	article.HandleImageGalleryArticles(h, w2, req2)
	if w2.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 image gallery, got %d", w2.Result().StatusCode)
	}
	var imgs []models.Article
	if err := json.NewDecoder(w2.Result().Body).Decode(&imgs); err != nil {
		t.Fatalf("decode imgs: %v", err)
	}
	if len(imgs) == 0 {
		t.Fatalf("expected image articles, got 0")
	}
}

func TestArticleActions_MarkRead_Favorite_Hide_ReadLater(t *testing.T) {
	h := setupHandler(t)
	feedID, _ := h.DB.AddFeed(&models.Feed{Title: "F2", URL: "http://y"})

	a := &models.Article{FeedID: feedID, Title: "act", URL: "u", PublishedAt: time.Now()}
	if err := h.DB.SaveArticles(context.Background(), []*models.Article{a}); err != nil {
		t.Fatalf("SaveArticles: %v", err)
	}
	// fetch saved article id
	arts, err := h.DB.GetArticles("", feedID, "", true, 10, 0)
	if err != nil || len(arts) == 0 {
		t.Fatalf("GetArticles: %v", err)
	}
	id := arts[0].ID

	// Mark unread -> read
	req := httptest.NewRequest(http.MethodPost, "/api/articles/mark_read?id="+fmt.Sprint(id)+"&read=true", nil)
	w := httptest.NewRecorder()
	article.HandleMarkRead(h, w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("mark read failed: %d", w.Result().StatusCode)
	}

	// Toggle favorite
	req2 := httptest.NewRequest(http.MethodPost, "/api/articles/toggle_fav?id="+fmt.Sprint(id), nil)
	w2 := httptest.NewRecorder()
	article.HandleToggleFavorite(h, w2, req2)
	if w2.Result().StatusCode != http.StatusOK {
		t.Fatalf("toggle fav failed: %d", w2.Result().StatusCode)
	}

	// Toggle hide (invalid method GET -> 405)
	req3 := httptest.NewRequest(http.MethodGet, "/api/articles/toggle_hide?id="+fmt.Sprint(id), nil)
	w3 := httptest.NewRecorder()
	article.HandleToggleHideArticle(h, w3, req3)
	if w3.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET hide, got %d", w3.Result().StatusCode)
	}

	// Proper POST hide
	req4 := httptest.NewRequest(http.MethodPost, "/api/articles/toggle_hide?id="+fmt.Sprint(id), nil)
	w4 := httptest.NewRecorder()
	article.HandleToggleHideArticle(h, w4, req4)
	if w4.Result().StatusCode != http.StatusOK {
		t.Fatalf("toggle hide failed: %d", w4.Result().StatusCode)
	}

	// Toggle read later (POST)
	req5 := httptest.NewRequest(http.MethodPost, "/api/articles/toggle_read_later?id="+fmt.Sprint(id), nil)
	w5 := httptest.NewRecorder()
	article.HandleToggleReadLater(h, w5, req5)
	if w5.Result().StatusCode != http.StatusOK {
		t.Fatalf("toggle read later failed: %d", w5.Result().StatusCode)
	}
}

func TestHandleExportToObsidian(t *testing.T) {
	h := setupHandler(t)

	// Enable Obsidian integration
	if err := h.DB.SetSetting("obsidian_enabled", "true"); err != nil {
		t.Fatalf("SetSetting obsidian_enabled: %v", err)
	}
	if err := h.DB.SetSetting("obsidian_vault_path", t.TempDir()); err != nil {
		t.Fatalf("SetSetting obsidian_vault_path: %v", err)
	}

	// Add a feed and article
	feedID, err := h.DB.AddFeed(&models.Feed{Title: "Test Feed", URL: "http://example.com"})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	articleModel := &models.Article{
		FeedID:      feedID,
		Title:       "Test Article",
		URL:         "http://example.com/article",
		PublishedAt: time.Now(),
	}
	if err := h.DB.SaveArticles(context.Background(), []*models.Article{articleModel}); err != nil {
		t.Fatalf("SaveArticles: %v", err)
	}

	// Get the article ID
	articles, err := h.DB.GetArticles("", feedID, "", false, 10, 0)
	if err != nil || len(articles) == 0 {
		t.Fatalf("GetArticles: %v", err)
	}
	articleID := articles[0].ID

	// Test export request
	reqBody := fmt.Sprintf(`{"article_id": %d}`, articleID)
	req := httptest.NewRequest(http.MethodPost, "/api/articles/export/obsidian", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	article.HandleExportToObsidian(h, w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("Export failed: %d, body: %s", w.Result().StatusCode, w.Body.String())
	}

	// Verify response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != "true" {
		t.Fatalf("Export not successful: %v", response)
	}
}

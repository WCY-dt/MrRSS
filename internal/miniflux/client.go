package miniflux

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"MrRSS/internal/models"
)

// Client represents a Miniflux API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new Miniflux API client
func NewClient(serverURL, apiKey string) *Client {
	// Ensure URL ends with /v1
	serverURL = strings.TrimSuffix(serverURL, "/")
	if !strings.HasSuffix(serverURL, "/v1") {
		serverURL += "/v1"
	}

	return &Client{
		baseURL: serverURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			},
		},
	}
}

// TestConnection tests the connection to Miniflux server using /v1/me endpoint
func (c *Client) TestConnection(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/me", nil)
	if err != nil {
		return fmt.Errorf("create test request: %w", err)
	}

	req.Header.Set("X-Auth-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("test request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("test connection failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Feed represents a Miniflux feed
type Feed struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	FeedURL  string `json:"feed_url"`
	SiteURL  string `json:"site_url"`
	Category struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	} `json:"category"`
}

// GetFeeds retrieves all feeds from Miniflux
func (c *Client) GetFeeds(ctx context.Context) ([]Feed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/feeds", nil)
	if err != nil {
		return nil, fmt.Errorf("create feeds request: %w", err)
	}

	req.Header.Set("X-Auth-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("feeds request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("feeds request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var feeds []Feed
	if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
		return nil, fmt.Errorf("decode feeds response: %w", err)
	}

	return feeds, nil
}

// Entry represents a Miniflux entry
type Entry struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Content     string    `json:"content"`
	PublishedAt time.Time `json:"published_at"`
	Author      string    `json:"author"`
	Status      string    `json:"status"`
	Starred     bool      `json:"starred"`
}

// GetEntries retrieves entries from Miniflux with optional filters
func (c *Client) GetEntries(ctx context.Context, status string, limit int) ([]Entry, error) {
	url := fmt.Sprintf("%s/entries?status=%s&limit=%d", c.baseURL, status, limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create entries request: %w", err)
	}

	req.Header.Set("X-Auth-Token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("entries request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("entries request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Total   int     `json:"total"`
		Entries []Entry `json:"entries"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode entries response: %w", err)
	}

	return result.Entries, nil
}

// UpdateEntries updates the status of multiple entries
func (c *Client) UpdateEntries(ctx context.Context, entryIDs []int64, status string) error {
	payload := map[string]interface{}{
		"entry_ids": entryIDs,
		"status":    status,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal update request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", c.baseURL+"/entries", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create update request: %w", err)
	}

	req.Header.Set("X-Auth-Token", c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("update request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SyncService handles synchronization between MrRSS and Miniflux
type SyncService struct {
	client *Client
	db     Database
}

// Database interface for Miniflux sync operations
type Database interface {
	GetFeeds() ([]models.Feed, error)
	AddFeed(feed *models.Feed) (int64, error)
	SaveArticles(ctx context.Context, articles []*models.Article) error
}

// NewSyncService creates a new sync service
func NewSyncService(serverURL, apiKey string, db Database) *SyncService {
	return &SyncService{
		client: NewClient(serverURL, apiKey),
		db:     db,
	}
}

// Sync performs a bidirectional sync with Miniflux
func (s *SyncService) Sync(ctx context.Context) error {
	// Get feeds from Miniflux
	minifluxFeeds, err := s.client.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("get Miniflux feeds: %w", err)
	}

	// Get local feeds
	localFeeds, err := s.db.GetFeeds()
	if err != nil {
		return fmt.Errorf("get local feeds: %w", err)
	}

	// Create a map of local feed URLs for quick lookup
	localFeedMap := make(map[string]int64)
	for _, feed := range localFeeds {
		localFeedMap[feed.URL] = feed.ID
	}

	// Add missing feeds to local database
	for _, mf := range minifluxFeeds {
		if _, exists := localFeedMap[mf.FeedURL]; !exists {
			feed := &models.Feed{
				Title:       mf.Title,
				URL:         mf.FeedURL,
				Link:        mf.SiteURL,
				Category:    mf.Category.Title,
				LastUpdated: time.Now(),
			}

			_, err := s.db.AddFeed(feed)
			if err != nil {
				log.Printf("Failed to add feed %s: %v", mf.FeedURL, err)
				continue
			}
			log.Printf("Added feed: %s", mf.Title)
		}
	}

	// Get unread entries from Miniflux (limit 100)
	entries, err := s.client.GetEntries(ctx, "unread", 100)
	if err != nil {
		return fmt.Errorf("get unread entries: %w", err)
	}

	// Create or get Miniflux feed for synced articles
	minifluxFeedID, err := s.getOrCreateMinifluxFeed()
	if err != nil {
		return fmt.Errorf("create Miniflux feed: %w", err)
	}

	// Convert Miniflux entries to MrRSS articles
	articles := make([]*models.Article, 0, len(entries))
	for _, entry := range entries {
		article := &models.Article{
			FeedID:      minifluxFeedID,
			Title:       entry.Title,
			URL:         entry.URL,
			Content:     entry.Content,
			PublishedAt: entry.PublishedAt,
			IsRead:      false, // Miniflux unread entries
			IsFavorite:  entry.Starred,
			IsHidden:    false,
		}
		articles = append(articles, article)
	}

	// Save articles to database
	if len(articles) > 0 {
		if err := s.db.SaveArticles(ctx, articles); err != nil {
			return fmt.Errorf("save articles: %w", err)
		}
		log.Printf("Synced %d articles from Miniflux", len(articles))
	}

	log.Printf("Miniflux sync completed successfully")
	return nil
}

// getOrCreateMinifluxFeed creates or retrieves the special Miniflux sync feed
func (s *SyncService) getOrCreateMinifluxFeed() (int64, error) {
	// Check if Miniflux feed already exists
	feeds, err := s.db.GetFeeds()
	if err != nil {
		return 0, err
	}

	for _, feed := range feeds {
		if feed.URL == "miniflux://synced" {
			return feed.ID, nil
		}
	}

	// Create new Miniflux feed
	minifluxFeed := &models.Feed{
		Title:       "Miniflux Synced Articles",
		URL:         "miniflux://synced",
		Description: "Articles synced from Miniflux server",
		Category:    "Miniflux",
		LastUpdated: time.Now(),
	}

	return s.db.AddFeed(minifluxFeed)
}

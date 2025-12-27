// Package core contains the main Handler struct and core HTTP handlers for the application.
// It defines the Handler struct which holds dependencies like the database and fetcher.
package core

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"MrRSS/internal/aiusage"
	"MrRSS/internal/cache"
	"MrRSS/internal/database"
	"MrRSS/internal/discovery"
	"MrRSS/internal/feed"
	"MrRSS/internal/models"
	"MrRSS/internal/translation"
	"MrRSS/internal/utils"

	"codeberg.org/readeck/go-readability/v2"

	"github.com/mmcdole/gofeed"
)

// Discovery timeout constants
const (
	// SingleFeedDiscoveryTimeout is the timeout for discovering feeds from a single source
	SingleFeedDiscoveryTimeout = 90 * time.Second
	// BatchDiscoveryTimeout is the timeout for discovering feeds from all sources
	BatchDiscoveryTimeout = 5 * time.Minute
)

// DiscoveryState represents the current state of a discovery operation
type DiscoveryState struct {
	IsRunning  bool                       `json:"is_running"`
	Progress   discovery.Progress         `json:"progress"`
	Feeds      []discovery.DiscoveredBlog `json:"feeds,omitempty"`
	Error      string                     `json:"error,omitempty"`
	IsComplete bool                       `json:"is_complete"`
}

// Handler holds all dependencies for HTTP handlers.
type Handler struct {
	DB               *database.DB
	Fetcher          *feed.Fetcher
	Translator       translation.Translator
	AITracker        *aiusage.Tracker
	DiscoveryService *discovery.Service
	App              interface{}         // Wails app instance for browser integration (interface{} to avoid import in server mode)
	ContentCache     *cache.ContentCache // Cache for article content

	// Discovery state tracking for polling-based progress
	DiscoveryMu          sync.RWMutex
	SingleDiscoveryState *DiscoveryState
	BatchDiscoveryState  *DiscoveryState
}

// NewHandler creates a new Handler with the given dependencies.
func NewHandler(db *database.DB, fetcher *feed.Fetcher, translator translation.Translator) *Handler {
	return &Handler{
		DB:               db,
		Fetcher:          fetcher,
		Translator:       translator,
		AITracker:        aiusage.NewTracker(db),
		DiscoveryService: discovery.NewService(),
		ContentCache:     cache.NewContentCache(100, 30*time.Minute), // Cache up to 100 articles for 30 minutes
	}
}

// SetApp sets the Wails application instance for browser integration.
// This is called after app initialization in main.go.
func (h *Handler) SetApp(app interface{}) {
	h.App = app
}

// GetArticleContent fetches article content with caching
func (h *Handler) GetArticleContent(articleID int64) (string, error) {
	// First, check database cache (persistent cache)
	content, found, err := h.DB.GetArticleContent(articleID)
	if err == nil && found {
		// Also populate memory cache for faster subsequent access
		h.ContentCache.Set(articleID, content)
		return content, nil
	}

	// Check memory cache (in-memory cache, might be stale but fast)
	if content, found := h.ContentCache.Get(articleID); found {
		return content, nil
	}

	// Get the article from database
	article, err := h.DB.GetArticleByID(articleID)
	if err != nil {
		return "", err
	}

	// Get the feed
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		return "", err
	}

	var targetFeed *models.Feed
	for _, f := range feeds {
		if f.ID == article.FeedID {
			targetFeed = &f
			break
		}
	}

	if targetFeed == nil {
		return "", nil
	}

	// Try to get feed from cache first
	var parsedFeed *gofeed.Feed
	if cachedFeed, found := h.ContentCache.GetFeed(targetFeed.ID); found {
		parsedFeed = cachedFeed
	} else {
		// Parse the feed to get fresh content
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		parsedFeed, err = h.Fetcher.ParseFeedWithFeed(ctx, targetFeed, true) // High priority for content fetching
		if err != nil {
			return "", err
		}

		// Cache the feed for future use
		h.ContentCache.SetFeed(targetFeed.ID, parsedFeed)
	}

	// Find the article in the feed by multiple criteria for better matching
	matchingItem := h.findMatchingFeedItem(article, parsedFeed.Items)
	if matchingItem != nil {
		content := feed.ExtractContent(matchingItem)
		cleanContent := utils.CleanHTML(content)

		// Cache the content in both memory and database
		h.ContentCache.Set(articleID, cleanContent)
		if err := h.DB.SetArticleContent(articleID, cleanContent); err != nil {
			log.Printf("Error caching content to database: %v", err)
		}

		return cleanContent, nil
	}

	return "", nil
}

// FetchFullArticleContent fetches the full article content from the original URL using readability.
func (h *Handler) FetchFullArticleContent(url string) (string, error) {
	// Use FromURL which handles the HTTP request internally
	article, err := readability.FromURL(url, 30*time.Second)
	if err != nil {
		return "", fmt.Errorf("readability parse: %w", err)
	}

	// Render the article content as HTML
	var buf bytes.Buffer
	err = article.RenderHTML(&buf)
	if err != nil {
		return "", fmt.Errorf("render HTML: %w", err)
	}

	return buf.String(), nil
}

// findMatchingFeedItem finds the best matching feed item for an article using multiple criteria
func (h *Handler) findMatchingFeedItem(article *models.Article, items []*gofeed.Item) *gofeed.Item {
	// First pass: exact URL match
	for _, item := range items {
		if utils.URLsMatch(item.Link, article.URL) {
			return item
		}
	}

	// Second pass: URL + title match (for script-based feeds that might have URL variations)
	for _, item := range items {
		if utils.URLsMatch(item.Link, article.URL) && h.titlesMatch(item.Title, article.Title) {
			return item
		}
	}

	// Third pass: title + published time match (fallback for when URLs don't match)
	for _, item := range items {
		if h.titlesMatch(item.Title, article.Title) && h.publishedTimesMatch(item.PublishedParsed, &article.PublishedAt) {
			return item
		}
	}

	// Final fallback: just title match
	for _, item := range items {
		if h.titlesMatch(item.Title, article.Title) {
			return item
		}
	}

	return nil
}

// titlesMatch checks if two titles match, allowing for minor differences
func (h *Handler) titlesMatch(title1, title2 string) bool {
	if title1 == title2 {
		return true
	}
	// Normalize titles by removing extra whitespace and comparing
	normalized1 := strings.TrimSpace(strings.Join(strings.Fields(title1), " "))
	normalized2 := strings.TrimSpace(strings.Join(strings.Fields(title2), " "))
	return normalized1 == normalized2
}

// publishedTimesMatch checks if two published times match within a reasonable tolerance
func (h *Handler) publishedTimesMatch(time1, time2 *time.Time) bool {
	if time1 == nil || time2 == nil {
		return false
	}
	// Allow for 1 minute difference in published times
	diff := time1.Sub(*time2)
	if diff < 0 {
		diff = -diff
	}
	return diff <= time.Minute
}

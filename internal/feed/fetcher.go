package feed

import (
	"MrRSS/internal/database"
	"MrRSS/internal/models"
	"MrRSS/internal/rules"
	"MrRSS/internal/translation"
	"MrRSS/internal/utils"
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

// FeedParser interface to allow mocking
type FeedParser interface {
	ParseURL(url string) (*gofeed.Feed, error)
	ParseURLWithContext(url string, ctx context.Context) (*gofeed.Feed, error)
}

type Fetcher struct {
	db                *database.DB
	fp                FeedParser
	translator        translation.Translator
	scriptExecutor    *ScriptExecutor
	progress          Progress
	mu                sync.Mutex
	refreshCalculator *IntelligentRefreshCalculator
	// Queue tracking for individual feed refreshes
	queuedFeeds       map[int64]bool // Tracks feeds that are queued for refresh
	queueMu           sync.Mutex
}

func NewFetcher(db *database.DB, translator translation.Translator) *Fetcher {
	// Initialize script executor with scripts directory
	scriptsDir, err := utils.GetScriptsDir()
	var executor *ScriptExecutor
	if err == nil {
		executor = NewScriptExecutor(scriptsDir)
	}

	return &Fetcher{
		db:                db,
		fp:                gofeed.NewParser(),
		translator:        translator,
		scriptExecutor:    executor,
		refreshCalculator: NewIntelligentRefreshCalculator(db),
		queuedFeeds:       make(map[int64]bool),
	}
}

// GetIntelligentRefreshCalculator returns the refresh calculator
func (f *Fetcher) GetIntelligentRefreshCalculator() *IntelligentRefreshCalculator {
	return f.refreshCalculator
}

// GetStaggeredDelay calculates a staggered delay for feed refresh
func (f *Fetcher) GetStaggeredDelay(feedID int64, totalFeeds int) time.Duration {
	return GetStaggeredDelay(feedID, totalFeeds)
}

// getHTTPClient returns an HTTP client configured with proxy if needed
// Proxy precedence (highest to lowest):
// 1. Feed custom proxy (ProxyEnabled=true, ProxyURL != "")
// 2. Global proxy (ProxyEnabled=true, ProxyURL == "", global proxy_enabled=true)
// 3. No proxy (ProxyEnabled=false or no global proxy)
func (f *Fetcher) getHTTPClient(feed models.Feed) (*http.Client, error) {
	var proxyURL string

	// Check feed-level proxy settings
	if feed.ProxyEnabled && feed.ProxyURL != "" {
		// Feed has custom proxy configured - highest priority
		proxyURL = feed.ProxyURL
	} else if feed.ProxyEnabled {
		// Feed requests to use global proxy
		proxyEnabled, _ := f.db.GetSetting("proxy_enabled")
		if proxyEnabled == "true" {
			// Build global proxy URL from settings (use encrypted methods for credentials)
			proxyType, _ := f.db.GetSetting("proxy_type")
			proxyHost, _ := f.db.GetSetting("proxy_host")
			proxyPort, _ := f.db.GetSetting("proxy_port")
			proxyUsername, _ := f.db.GetEncryptedSetting("proxy_username")
			proxyPassword, _ := f.db.GetEncryptedSetting("proxy_password")
			proxyURL = BuildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)
		}
	}
	// If ProxyEnabled=false, proxyURL remains empty (no proxy)

	// Create HTTP client with or without proxy
	return CreateHTTPClient(proxyURL)
}

// setupTranslator configures the translator based on database settings.
// Now supports global proxy settings for all translation services.
func (f *Fetcher) setupTranslator() {
	provider, _ := f.db.GetSetting("translation_provider")

	var t translation.Translator
	switch provider {
	case "deepl":
		apiKey, _ := f.db.GetEncryptedSetting("deepl_api_key")
		if apiKey != "" {
			t = translation.NewDeepLTranslatorWithDB(apiKey, f.db)
		} else {
			t = translation.NewGoogleFreeTranslatorWithDB(f.db)
		}
	case "baidu":
		appID, _ := f.db.GetSetting("baidu_app_id")
		secretKey, _ := f.db.GetEncryptedSetting("baidu_secret_key")
		if appID != "" && secretKey != "" {
			t = translation.NewBaiduTranslatorWithDB(appID, secretKey, f.db)
		} else {
			t = translation.NewGoogleFreeTranslatorWithDB(f.db)
		}
	case "ai":
		apiKey, _ := f.db.GetEncryptedSetting("ai_api_key")
		endpoint, _ := f.db.GetSetting("ai_endpoint")
		model, _ := f.db.GetSetting("ai_model")
		if apiKey != "" {
			t = translation.NewAITranslatorWithDB(apiKey, endpoint, model, f.db)
		} else {
			t = translation.NewGoogleFreeTranslatorWithDB(f.db)
		}
	default:
		// Default to Google Free Translator with proxy support
		t = translation.NewGoogleFreeTranslatorWithDB(f.db)
	}
	f.translator = t
}

func (f *Fetcher) FetchAll(ctx context.Context) {
	f.mu.Lock()
	if f.progress.IsRunning {
		f.mu.Unlock()
		return
	}
	f.progress.IsRunning = true
	f.progress.Current = 0
	f.mu.Unlock()

	// Clear any queued individual feeds since we're doing a full refresh
	f.queueMu.Lock()
	f.queuedFeeds = make(map[int64]bool)
	f.queueMu.Unlock()

	// Setup translator based on settings
	f.setupTranslator()

	feeds, err := f.db.GetFeeds()
	if err != nil {
		log.Println("Error getting feeds:", err)
		f.mu.Lock()
		f.progress.IsRunning = false
		f.mu.Unlock()
		return
	}

	f.mu.Lock()
	f.progress.Total = len(feeds)
	f.mu.Unlock()

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // Limit concurrency

	for _, feed := range feeds {
		// Check for cancellation
		select {
		case <-ctx.Done():
			log.Println("FetchAll cancelled (loop)")
			goto Finish
		default:
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(fd models.Feed) {
			defer wg.Done()
			defer func() { <-sem }()

			// Check for cancellation inside goroutine before starting
			select {
			case <-ctx.Done():
				return
			default:
			}

			f.FetchFeed(ctx, fd)
			f.mu.Lock()
			f.progress.Current++
			f.mu.Unlock()
		}(feed)
	}

Finish:
	wg.Wait()

	f.mu.Lock()
	f.progress.IsRunning = false
	f.mu.Unlock()

	// Update last article update time
	f.db.SetSetting("last_article_update", time.Now().Format(time.RFC3339))
}

func (f *Fetcher) FetchFeed(ctx context.Context, feed models.Feed) {
	var parsedFeed *gofeed.Feed
	var err error

	// Check if this feed uses a custom script
	if feed.ScriptPath != "" {
		// Execute the custom script to fetch feed
		if f.scriptExecutor == nil {
			log.Printf("Script executor not initialized for feed %s", feed.Title)
			f.db.UpdateFeedError(feed.ID, "Script executor not initialized")
			return
		}
		parsedFeed, err = f.scriptExecutor.ExecuteScript(ctx, feed.ScriptPath)
		if err != nil {
			log.Printf("Error executing script for feed %s: %v", feed.Title, err)
			f.db.UpdateFeedError(feed.ID, err.Error())
			return
		}
	} else {
		// Get HTTP client with proxy support
		httpClient, err := f.getHTTPClient(feed)
		if err != nil {
			log.Printf("Error creating HTTP client for feed %s: %v", feed.Title, err)
			f.db.UpdateFeedError(feed.ID, err.Error())
			return
		}

		// Set HTTP client on parser if it's a gofeed parser
		if gp, ok := f.fp.(*gofeed.Parser); ok {
			gp.Client = httpClient
		}

		// Use parser to fetch feed
		parsedFeed, err = f.fp.ParseURLWithContext(feed.URL, ctx)
		if err != nil {
			log.Printf("Error parsing feed %s: %v", feed.URL, err)
			f.db.UpdateFeedError(feed.ID, err.Error())
			return
		}
	}

	// Clear any previous error on successful fetch
	f.db.UpdateFeedError(feed.ID, "")

	// Update Feed Image if available and not set
	if feed.ImageURL == "" && parsedFeed.Image != nil {
		f.db.UpdateFeedImage(feed.ID, parsedFeed.Image.URL)
	}

	// Update Feed Link if available and not set
	if feed.Link == "" && parsedFeed.Link != "" {
		f.db.UpdateFeedLink(feed.ID, parsedFeed.Link)
	}

	// Process articles
	articlesToSave := f.processArticles(feed, parsedFeed.Items)

	// Check context before heavy DB operation
	select {
	case <-ctx.Done():
		return
	default:
	}

	if len(articlesToSave) > 0 {
		if err := f.db.SaveArticles(ctx, articlesToSave); err != nil {
			log.Printf("Error saving articles for feed %s: %v", feed.Title, err)
		} else {
			// Apply rules to newly saved articles
			// We fetch the recent articles for this feed since SaveArticles doesn't return IDs
			// This is limited to the number of articles we just saved
			savedArticles, err := f.db.GetArticles("", feed.ID, "", false, len(articlesToSave), 0)
			if err == nil && len(savedArticles) > 0 {
				engine := rules.NewEngine(f.db)
				affected, err := engine.ApplyRulesToArticles(savedArticles)
				if err != nil {
					log.Printf("Error applying rules for feed %s: %v", feed.Title, err)
				} else if affected > 0 {
					log.Printf("Applied rules to %d articles in feed %s", affected, feed.Title)
				}
			}
		}
	}
	log.Printf("Updated feed: %s", feed.Title)
}

// FetchSingleFeed fetches a single feed with progress tracking.
// This is used when adding a new feed, refreshing a single feed from the context menu,
// or when the scheduler triggers individual feed refreshes.
func (f *Fetcher) FetchSingleFeed(ctx context.Context, feed models.Feed) {
	// Add feed to queue
	f.queueMu.Lock()
	if f.queuedFeeds[feed.ID] {
		// Feed is already queued, skip duplicate
		f.queueMu.Unlock()
		log.Printf("Feed %s is already queued for refresh, skipping", feed.Title)
		return
	}
	f.queuedFeeds[feed.ID] = true
	queuedCount := len(f.queuedFeeds)
	f.queueMu.Unlock()

	// Update progress to reflect the new queue state
	f.mu.Lock()
	if !f.progress.IsRunning {
		f.progress.IsRunning = true
		f.progress.Total = queuedCount
		f.progress.Current = 0
	} else {
		// Already running, just update total to include this new feed
		f.progress.Total = f.progress.Current + queuedCount
	}
	f.mu.Unlock()

	// Setup translator based on settings
	f.setupTranslator()

	// Fetch the feed
	f.FetchFeed(ctx, feed)

	// Remove from queue and update progress
	f.queueMu.Lock()
	delete(f.queuedFeeds, feed.ID)
	remainingCount := len(f.queuedFeeds)
	f.queueMu.Unlock()

	f.mu.Lock()
	f.progress.Current++
	if remainingCount == 0 {
		// No more feeds in queue, mark as complete
		f.progress.IsRunning = false
	} else {
		// Update total to reflect current queue state
		f.progress.Total = f.progress.Current + remainingCount
	}
	f.mu.Unlock()

	// Update last article update time when queue is empty
	if remainingCount == 0 {
		f.db.SetSetting("last_article_update", time.Now().Format(time.RFC3339))
		log.Printf("All queued feed updates complete")
	} else {
		log.Printf("Feed update complete: %s (%d remaining)", feed.Title, remainingCount)
	}
}

// FetchFeedsByIDs fetches multiple feeds by their IDs with progress tracking.
// This is used after OPML import or when refreshing specific feeds.
func (f *Fetcher) FetchFeedsByIDs(ctx context.Context, feedIDs []int64) {
	f.mu.Lock()
	if f.progress.IsRunning {
		f.mu.Unlock()
		// Wait for current operation to complete with timeout
		if !f.waitForProgressComplete(5 * time.Minute) {
			log.Println("FetchFeedsByIDs: Timeout waiting for previous operation")
			return
		}
		f.mu.Lock()
	}
	f.progress.IsRunning = true
	f.progress.Total = len(feedIDs)
	f.progress.Current = 0
	f.mu.Unlock()

	// Clear any queued individual feeds since we're doing a batch refresh
	f.queueMu.Lock()
	f.queuedFeeds = make(map[int64]bool)
	f.queueMu.Unlock()

	// Setup translator based on settings
	f.setupTranslator()

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // Limit concurrency

	for _, feedID := range feedIDs {
		// Check for cancellation
		select {
		case <-ctx.Done():
			log.Println("FetchFeedsByIDs cancelled")
			goto Finish
		default:
		}

		feed, err := f.db.GetFeedByID(feedID)
		if err != nil {
			log.Printf("Error getting feed %d: %v", feedID, err)
			f.mu.Lock()
			f.progress.Current++
			f.mu.Unlock()
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(fd models.Feed) {
			defer wg.Done()
			defer func() { <-sem }()

			select {
			case <-ctx.Done():
				return
			default:
			}

			f.FetchFeed(ctx, fd)
			f.mu.Lock()
			f.progress.Current++
			f.mu.Unlock()
		}(*feed)
	}

Finish:
	wg.Wait()

	f.mu.Lock()
	f.progress.IsRunning = false
	f.mu.Unlock()

	// Update last article update time
	f.db.SetSetting("last_article_update", time.Now().Format(time.RFC3339))
	log.Printf("Batch feed update complete for %d feeds", len(feedIDs))
}

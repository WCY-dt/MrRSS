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
	"strconv"
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
	highPriorityFp    FeedParser // High priority parser for content fetching
	translator        translation.Translator
	scriptExecutor    *ScriptExecutor
	progress          Progress
	mu                sync.Mutex
	refreshCalculator *IntelligentRefreshCalculator
	// Queue tracking for individual feed refreshes
	queuedFeeds map[int64]bool // Tracks feeds that are queued for refresh
	queueMu     sync.Mutex
	// Priority system
	priorityMu sync.Mutex // Protects priority operations
}

func NewFetcher(db *database.DB, translator translation.Translator) *Fetcher {
	// Initialize script executor with scripts directory
	scriptsDir, err := utils.GetScriptsDir()
	var executor *ScriptExecutor
	if err == nil {
		executor = NewScriptExecutor(scriptsDir)
	}

	// Create HTTP client for feed parsing
	httpClient, err := CreateHTTPClient("")
	if err != nil {
		// Fallback to default client if proxy setup fails
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	// Create parser with custom HTTP client to support localhost and other endpoints
	parser := gofeed.NewParser()
	parser.Client = httpClient

	// Create high priority parser with shorter timeout for content fetching
	highPriorityParser := gofeed.NewParser()
	highPriorityParser.Client = httpClient

	return &Fetcher{
		db:                db,
		fp:                parser,
		highPriorityFp:    highPriorityParser,
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

// getConcurrencyLimit returns the maximum number of concurrent feed refreshes
// based on network detection or defaults to 5 if not configured
// For large numbers of feeds, concurrency is reduced to prevent connection exhaustion
func (f *Fetcher) getConcurrencyLimit(feedCount int) int {
	concurrencyStr, err := f.db.GetSetting("max_concurrent_refreshes")
	if err != nil || concurrencyStr == "" {
		concurrency := 5 // Default concurrency
		// Reduce concurrency for large feed counts to prevent connection exhaustion
		if feedCount > 20 {
			concurrency = 2
		} else if feedCount > 10 {
			concurrency = 3
		}
		return concurrency
	}

	concurrency, err := strconv.Atoi(concurrencyStr)
	if err != nil || concurrency < 1 {
		concurrency := 5 // Default on parse error or invalid value
		// Reduce concurrency for large feed counts
		if feedCount > 20 {
			concurrency = 2
		} else if feedCount > 10 {
			concurrency = 3
		}
		return concurrency
	}

	// Cap at reasonable limits
	if concurrency > 20 {
		concurrency = 20
	}

	// Reduce concurrency for large feed counts to prevent connection exhaustion
	if feedCount > 50 {
		concurrency = min(concurrency, 3)
	} else if feedCount > 20 {
		concurrency = min(concurrency, 5)
	}

	return concurrency
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
			// Set custom headers if available
			if aiTranslator, ok := t.(*translation.AITranslator); ok {
				customHeaders, _ := f.db.GetSetting("ai_custom_headers")
				aiTranslator.SetCustomHeaders(customHeaders)
			}
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
	f.progress.Errors = make(map[int64]string)
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
	concurrency := f.getConcurrencyLimit(len(feeds))
	sem := make(chan struct{}, concurrency) // Limit concurrency based on network speed

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
	// Use ParseFeedWithFeed with normal priority for feed refresh
	parsedFeed, err := f.ParseFeedWithFeed(ctx, &feed, false) // Normal priority for refresh
	if err != nil {
		log.Printf("Error parsing feed %s: %v", feed.URL, err)
		f.db.UpdateFeedError(feed.ID, err.Error())
		// Add error to progress for immediate feedback
		f.mu.Lock()
		if f.progress.Errors == nil {
			f.progress.Errors = make(map[int64]string)
		}
		f.progress.Errors[feed.ID] = err.Error()
		f.mu.Unlock()
		return
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
	articlesWithContent := f.processArticles(feed, parsedFeed.Items)

	// Check context before heavy DB operation
	select {
	case <-ctx.Done():
		return
	default:
	}

	if len(articlesWithContent) > 0 {
		// Extract just the articles for saving
		articlesToSave := make([]*models.Article, len(articlesWithContent))
		for i, awc := range articlesWithContent {
			articlesToSave[i] = awc.Article
		}

		if err := f.db.SaveArticles(ctx, articlesToSave); err != nil {
			log.Printf("Error saving articles for feed %s: %v", feed.Title, err)
		} else {
			// Cache article content from RSS feed
			f.cacheArticleContents(articlesWithContent)

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
					utils.DebugLog("Applied rules to %d articles in feed %s", affected, feed.Title)
				}
			}
		}
	}
	utils.DebugLog("Updated feed: %s", feed.Title)
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
		utils.DebugLog("Feed %s is already queued for refresh, skipping", feed.Title)
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
		f.progress.Errors = make(map[int64]string)
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
		utils.DebugLog("Feed update complete: %s (%d remaining)", feed.Title, remainingCount)
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
	concurrency := f.getConcurrencyLimit(len(feedIDs))
	sem := make(chan struct{}, concurrency) // Limit concurrency based on network speed

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
	utils.DebugLog("Batch feed update complete for %d feeds", len(feedIDs))
}

// cacheArticleContents caches article contents from RSS feeds
// This is called after articles are saved to the database
func (f *Fetcher) cacheArticleContents(articlesWithContent []*ArticleWithContent) {
	for _, awc := range articlesWithContent {
		// Only cache if content is not empty and URL is present
		if awc.Content == "" || awc.Article.URL == "" {
			continue
		}

		// Get article ID by URL (article was just saved, so it should exist)
		articleID, err := f.db.GetArticleIDByURL(awc.Article.URL)
		if err != nil {
			// Article might not exist yet (race condition) or other error
			utils.DebugLog("Could not find article ID for URL %s: %v", awc.Article.URL, err)
			continue
		}

		// Cache the content (this will overwrite any existing cache as required)
		if err := f.db.SetArticleContent(articleID, awc.Content); err != nil {
			log.Printf("Error caching content for article %d: %v", articleID, err)
		} else {
			utils.DebugLog("Cached content for article %d", articleID)
		}
	}
}

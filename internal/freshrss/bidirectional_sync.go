package freshrss

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/models"
)

// SyncResult represents the result of a sync operation
type SyncResult struct {
	PullSuccess      bool
	PullChangesCount int
	PushSuccess      bool
	PushChangesCount int
	Errors           []string
	Duration         time.Duration
	LastSyncTime     time.Time
}

// BidirectionalSyncService handles bidirectional synchronization
type BidirectionalSyncService struct {
	client *Client
	db     *database.DB
}

// NewBidirectionalSyncService creates a new bidirectional sync service
func NewBidirectionalSyncService(serverURL, username, password string, db *database.DB) *BidirectionalSyncService {
	return &BidirectionalSyncService{
		client: NewClient(serverURL, username, password),
		db:     db,
	}
}

// Sync performs a full bidirectional sync
// This is called for manual/scheduled sync
// Logic: Pull remote changes first, then push local changes
func (s *BidirectionalSyncService) Sync(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{
		LastSyncTime: time.Now(),
	}
	startTime := time.Now()
	defer func() { result.Duration = time.Since(startTime) }()

	// Stage 1: Login to FreshRSS
	if err := s.client.Login(ctx); err != nil {
		return result, fmt.Errorf("login failed: %w", err)
	}

	// Stage 2: Pull from server (feeds, articles, starred status, read status)
	log.Printf("Stage 1: Pull from server")
	pullChanges, err := s.pullFromServer(ctx)
	if err != nil {
		log.Printf("Stage 1 ERROR: pull failed: %v", err)
		result.Errors = append(result.Errors, fmt.Sprintf("pull failed: %v", err))
		result.PullSuccess = false
		result.PushSuccess = false
		return result, err
	}
	log.Printf("Stage 1 SUCCESS: %d changes pulled", pullChanges)
	result.PullSuccess = true
	result.PullChangesCount = pullChanges

	// Stage 3: Push local changes to server
	log.Printf("Stage 2: Push to server")
	pushChanges, err := s.pushToServer(ctx)
	if err != nil {
		log.Printf("Stage 2 ERROR: push failed: %v", err)
		result.Errors = append(result.Errors, fmt.Sprintf("push failed: %v", err))
		result.PushSuccess = false
	} else {
		log.Printf("Stage 2 SUCCESS: %d changes pushed", pushChanges)
		result.PushSuccess = true
		result.PushChangesCount = pushChanges
	}

	return result, nil
}

// SyncFeed syncs articles for a single FreshRSS feed/stream
// This is called when user right-clicks a FreshRSS feed and selects "Sync Feed"
func (s *BidirectionalSyncService) SyncFeed(ctx context.Context, streamID string) (int, error) {
	// Login to FreshRSS
	if err := s.client.Login(ctx); err != nil {
		return 0, fmt.Errorf("login failed: %w", err)
	}

	log.Printf("[SyncFeed] Syncing stream: %s", streamID)

	// Fetch articles from this stream
	// Exclude already read articles to reduce data transfer
	excludeTypes := []string{"user/-/state/com.google/read"}
	contents, err := s.client.GetStreamContents(ctx, streamID, excludeTypes, 1000, "")
	if err != nil {
		return 0, fmt.Errorf("get stream contents: %w", err)
	}

	if len(contents.Items) == 0 {
		log.Printf("[SyncFeed] No new articles in stream: %s", streamID)
		return 0, nil
	}

	// Save articles to database
	count, err := s.saveArticlesFromServer(ctx, contents.Items)
	if err != nil {
		return 0, fmt.Errorf("save articles: %w", err)
	}

	log.Printf("[SyncFeed] Synced %d articles for stream: %s", count, streamID)
	return count, nil
}

// SyncArticleStatus syncs a single article's status immediately
// This is called when user manually marks an article as read/unread or starred/unstarred
// Logic: Immediately push local status to server, overwriting remote
// If sync fails, the change is added to the queue for later retry
func (s *BidirectionalSyncService) SyncArticleStatus(ctx context.Context, articleID int64, articleURL string, action database.SyncAction) error {
	// Login to FreshRSS
	if err := s.client.Login(ctx); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Get the article to check if we have FreshRSS Item ID
	article, err := s.db.GetArticleByID(articleID)
	if err != nil {
		log.Printf("[Immediate Sync] Failed to get article: %v", err)
		return err
	}

	// Use FreshRSS Item ID if available, otherwise fall back to URL
	identifier := articleURL
	if article.FreshRSSItemID != "" {
		identifier = article.FreshRSSItemID
		log.Printf("[Immediate Sync] Using FreshRSS Item ID: %s for article %d", identifier, articleID)
	} else {
		log.Printf("[Immediate Sync] No FreshRSS Item ID, using URL: %s for article %d", articleURL, articleID)
	}

	// Perform the action immediately
	log.Printf("[Immediate Sync] Syncing article status: %s (%s) -> %s", articleURL, identifier, action)

	var syncErr error
	switch action {
	case database.SyncActionMarkRead:
		syncErr = s.client.MarkAsReadBatch(ctx, []string{identifier})
	case database.SyncActionMarkUnread:
		syncErr = s.client.MarkAsUnreadBatch(ctx, []string{identifier})
	case database.SyncActionStar:
		syncErr = s.client.StarBatch(ctx, []string{identifier})
	case database.SyncActionUnstar:
		syncErr = s.client.UnstarBatch(ctx, []string{identifier})
	}

	if syncErr != nil {
		log.Printf("[Immediate Sync] ERROR: %v", syncErr)
		// Add to queue for retry
		if queueErr := s.db.EnqueueSyncChange(articleID, articleURL, action); queueErr != nil {
			log.Printf("[Immediate Sync] Failed to enqueue for retry: %v", queueErr)
		} else {
			log.Printf("[Immediate Sync] Enqueued for retry due to sync failure")
		}
		return syncErr
	}

	log.Printf("[Immediate Sync] SUCCESS: %s -> %s", articleURL, action)
	return nil
}

// pullFromServer pulls changes from FreshRSS server
func (s *BidirectionalSyncService) pullFromServer(ctx context.Context) (int, error) {
	totalChanges := 0
	log.Printf("pullFromServer: Starting pull from server")

	// Step 1: Get subscriptions and create feeds
	subscriptions, err := s.client.GetSubscriptions(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get subscriptions: %v", err)
		if subscriptions == nil {
			subscriptions = []Subscription{}
		}
	}

	// Only proceed if we have subscriptions
	if len(subscriptions) > 0 {
		feedsCreated, err := s.createFeedsFromSubscriptions(ctx, subscriptions)
		if err != nil {
			log.Printf("Warning: Failed to create feeds: %v", err)
		} else {
			totalChanges += feedsCreated
			log.Printf("Created/updated %d feeds from FreshRSS", feedsCreated)
		}

		// Step 2: Get articles from each subscription
		articlesPerFeed := 100
		totalArticles := 0

		for _, sub := range subscriptions {
			feedURL := strings.TrimPrefix(sub.ID, "feed/")

			result, err := s.client.GetStreamContents(ctx, sub.ID, nil, articlesPerFeed, "")
			if err != nil {
				log.Printf("Warning: Failed to get articles for feed %s: %v", feedURL, err)
				continue
			}

			if len(result.Items) > 0 {
				saved, err := s.saveArticlesFromServer(ctx, result.Items)
				if err != nil {
					log.Printf("Warning: Failed to save articles for feed %s: %v", feedURL, err)
				} else {
					totalArticles += saved
				}
			}

			time.Sleep(50 * time.Millisecond)
		}

		totalChanges += totalArticles
		log.Printf("Saved %d total articles from %d feeds", totalArticles, len(subscriptions))
	} else {
		log.Printf("No subscriptions to sync from FreshRSS")
	}

	// Step 3: Apply starred status from server
	log.Printf("pullFromServer: Step 3 - Applying starred status")
	starredArticles, err := s.client.GetStarredArticles(ctx, 1000)
	if err != nil {
		log.Printf("Warning: Failed to get starred articles: %v", err)
	} else {
		for _, article := range starredArticles {
			_, err := s.applyServerStatus(article.URL, true, "is_favorite")
			if err != nil {
				log.Printf("Warning: Failed to apply starred status for %s: %v", article.URL, err)
			}
		}
		log.Printf("Applied starred status to %d articles from server", len(starredArticles))
	}

	// Step 4: Apply read status from server (optional - can be skipped if API doesn't support it well)
	log.Printf("pullFromServer: Step 4 - Applying read status")
	readArticles, err := s.client.GetReadArticles(ctx, 500)
	if err != nil {
		log.Printf("Warning: Failed to get read articles: %v", err)
	} else {
		for _, article := range readArticles {
			_, err := s.applyServerStatus(article.URL, true, "is_read")
			if err != nil {
				log.Printf("Warning: Failed to apply read status for %s: %v", article.URL, err)
			}
		}
		log.Printf("Applied read status to %d articles from server", len(readArticles))
	}

	log.Printf("Pull from server: %d total changes applied", totalChanges)
	log.Printf("pullFromServer: Completed")

	return totalChanges, nil
}

// applyServerStatus applies server status to local article
// If local has a pending sync for the same action type, clear it to avoid conflicts
func (s *BidirectionalSyncService) applyServerStatus(articleURL string, status bool, column string) (int, error) {
	// Check if article exists locally
	localArticle, err := s.db.GetArticleByURL(articleURL)
	if err != nil {
		// Article doesn't exist locally, skip
		return 0, nil
	}

	// Check if there's a pending sync change for the SAME action type
	// Only clear conflicting pending syncs
	pendingItems, err := s.db.GetPendingSyncChanges(1000)
	hasConflictingSync := false
	for _, item := range pendingItems {
		if item.ArticleURL == articleURL {
			isConflictingAction := false
			switch column {
			case "is_favorite":
				// Star/unstar actions conflict with favorite status changes
				isConflictingAction = (item.Action == database.SyncActionStar || item.Action == database.SyncActionUnstar)
			case "is_read":
				// Mark read/unread actions conflict with read status changes
				isConflictingAction = (item.Action == database.SyncActionMarkRead || item.Action == database.SyncActionMarkUnread)
			}

			if isConflictingAction {
				hasConflictingSync = true
				log.Printf("[Conflict detected] Article %s has pending %s, server says %s=%v - clearing pending sync",
					articleURL, item.Action, column, status)
				break
			}
		}
	}

	// Apply server status
	switch column {
	case "is_favorite":
		err = s.db.SetArticleFavorite(localArticle.ID, status)
	case "is_read":
		err = s.db.MarkArticleRead(localArticle.ID, status)
	}
	if err != nil {
		return 0, err
	}

	// Clear conflicting pending sync if any
	if hasConflictingSync {
		_ = s.db.ClearPendingSyncForArticle(localArticle.ID)
	}

	return 1, nil
}

// createFeedsFromSubscriptions creates local feeds from FreshRSS subscriptions
func (s *BidirectionalSyncService) createFeedsFromSubscriptions(ctx context.Context, subscriptions []Subscription) (int, error) {
	feedsCreated := 0

	// Get all existing feeds to check for duplicates
	existingFeeds, err := s.db.GetFeeds()
	if err != nil {
		log.Printf("Warning: Failed to get existing feeds: %v", err)
	}

	// Create a map of feed URLs to existing feeds
	feedURLMap := make(map[string]*models.Feed)
	titleMap := make(map[string]int64)
	for i := range existingFeeds {
		feedURLMap[existingFeeds[i].URL] = &existingFeeds[i]
		titleMap[existingFeeds[i].Title] = existingFeeds[i].ID
	}

	for _, sub := range subscriptions {
		feedURL := sub.URL

		// Extract category from subscription categories
		category := ""
		if len(sub.Categories) > 0 {
			for _, cat := range sub.Categories {
				if strings.HasPrefix(cat.ID, "user/-/label/") {
					category = cat.Label
					break
				}
			}
		}

		// Adjust title if there's a conflict with an existing non-FreshRSS feed
		feedTitle := sub.Title
		if existingID, exists := titleMap[feedTitle]; exists {
			if existingFeed, ok := feedURLMap[feedURL]; !ok || existingFeed.ID != existingID {
				feedTitle = feedTitle + " (FreshRSS)"
				log.Printf("Title conflict detected for '%s', using adjusted title '%s'", sub.Title, feedTitle)
			}
		}

		// Check if feed already exists
		if existingFeed, exists := feedURLMap[feedURL]; exists {
			// Feed exists, check if we need to update it
			needsUpdate := false

			if existingFeed.Title != feedTitle {
				needsUpdate = true
			}

			if category != "" && existingFeed.Category != category {
				needsUpdate = true
			}

			if needsUpdate {
				updateCategory := existingFeed.Category
				if category != "" {
					updateCategory = category
				}

				err := s.db.UpdateFeed(
					existingFeed.ID,
					feedTitle,
					existingFeed.URL,
					updateCategory,
					existingFeed.ScriptPath,
					existingFeed.HideFromTimeline,
					existingFeed.ProxyURL,
					existingFeed.ProxyEnabled,
					existingFeed.RefreshInterval,
					existingFeed.IsImageMode,
					existingFeed.Type,
					existingFeed.XPathItem,
					existingFeed.XPathItemTitle,
					existingFeed.XPathItemContent,
					existingFeed.XPathItemUri,
					existingFeed.XPathItemAuthor,
					existingFeed.XPathItemTimestamp,
					existingFeed.XPathItemTimeFormat,
					existingFeed.XPathItemThumbnail,
					existingFeed.XPathItemCategories,
					existingFeed.XPathItemUid,
					existingFeed.ArticleViewMode,
					existingFeed.AutoExpandContent,
					existingFeed.EmailAddress,
					existingFeed.EmailIMAPServer,
					existingFeed.EmailUsername,
					existingFeed.EmailPassword,
					existingFeed.EmailFolder,
					existingFeed.EmailIMAPPort,
				)
				if err != nil {
					log.Printf("Warning: Failed to update feed %s: %v", feedURL, err)
				} else {
					feedsCreated++
					log.Printf("Updated feed '%s' with category '%s'", feedTitle, updateCategory)
				}
			}
			continue
		}

		// Create new feed
		newFeed := &models.Feed{
			URL:              feedURL,
			Title:            feedTitle,
			Link:             feedURL,
			Description:      "",
			Category:         category,
			IsFreshRSSSource: true,
			FreshRSSStreamID: sub.ID,
		}

		_, err = s.db.AddFeed(newFeed)
		if err != nil {
			log.Printf("Warning: Failed to create feed %s: %v", feedURL, err)
		} else {
			feedsCreated++
			if category != "" {
				log.Printf("Created feed '%s' in category '%s'", feedTitle, category)
			} else {
				log.Printf("Created feed '%s' (no category)", feedTitle)
			}
		}
	}

	// Delete local FreshRSS feeds that no longer exist on the server
	remoteFeedURLs := make(map[string]bool)
	for _, sub := range subscriptions {
		remoteFeedURLs[sub.URL] = true
	}

	for _, feed := range existingFeeds {
		if feed.IsFreshRSSSource {
			if !remoteFeedURLs[feed.URL] {
				log.Printf("Deleting local FreshRSS feed '%s' (removed from server)", feed.Title)
				err := s.db.DeleteFeed(feed.ID)
				if err != nil {
					log.Printf("Warning: Failed to delete feed '%s': %v", feed.Title, err)
				} else {
					feedsCreated++
				}
			}
		}
	}

	return feedsCreated, nil
}

// saveArticlesFromServer saves articles from FreshRSS to local database
func (s *BidirectionalSyncService) saveArticlesFromServer(ctx context.Context, articles []Article) (int, error) {
	if len(articles) == 0 {
		return 0, nil
	}

	// Get all existing feeds to map stream IDs to feed IDs
	existingFeeds, err := s.db.GetFeeds()
	if err != nil {
		return 0, fmt.Errorf("get existing feeds: %w", err)
	}

	feedStreamIDMap := make(map[string]int64)
	feedURLMap := make(map[string]int64)
	for i := range existingFeeds {
		var streamID string
		if existingFeeds[i].IsFreshRSSSource && existingFeeds[i].FreshRSSStreamID != "" {
			streamID = existingFeeds[i].FreshRSSStreamID
		} else {
			streamID = "feed/" + existingFeeds[i].URL
		}
		feedStreamIDMap[streamID] = existingFeeds[i].ID
		feedURLMap[existingFeeds[i].URL] = existingFeeds[i].ID
	}

	// Convert FreshRSS articles to models.Article
	mrssArticles := make([]*models.Article, 0, len(articles))
	articleContentMap := make(map[string]string)
	skippedCount := 0

	for _, article := range articles {
		var feedID int64

		// Try to find feed ID using article's Origin stream ID first
		if article.OriginStreamID != "" {
			if id, exists := feedStreamIDMap[article.OriginStreamID]; exists {
				feedID = id
			}
		}

		// Fallback: try to find feed ID using categories
		if feedID == 0 && len(article.Categories) > 0 {
			for _, cat := range article.Categories {
				if strings.HasPrefix(cat, "user/-/label/") {
					feedURL := strings.TrimPrefix(cat, "user/-/label/")
					feedURL = strings.TrimPrefix(feedURL, "feed/")
					if id, exists := feedURLMap[feedURL]; exists {
						feedID = id
						break
					}
				}
			}
		}

		if feedID == 0 {
			skippedCount++
			if skippedCount <= 5 {
				log.Printf("Warning: Could not find feed for article '%s' (stream ID: %s)",
					article.Title, article.OriginStreamID)
			}
			continue
		}

		// Check article status from categories
		isRead := false
		isStarred := false
		for _, cat := range article.Categories {
			if cat == "user/-/state/com.google/read" {
				isRead = true
			}
			if cat == "user/-/state/com.google/starred" {
				isStarred = true
			}
		}

		// Check if article already exists (by URL)
		existingArticle, err := s.db.GetArticleByURL(article.URL)

		if err == nil && existingArticle != nil {
			// Article already exists, update read/starred status and FreshRSS Item ID if needed
			updated := false

			// Update FreshRSS Item ID if missing
			if existingArticle.FreshRSSItemID == "" && article.ID != "" {
				err := s.db.UpdateFreshRSSItemID(existingArticle.ID, article.ID)
				if err != nil {
					log.Printf("Warning: Failed to update FreshRSS Item ID for article %s: %v", article.URL, err)
				} else {
					log.Printf("Updated FreshRSS Item ID for existing article %s: %s", article.URL, article.ID)
					updated = true
				}
			}

			// Update read status
			if isRead && !existingArticle.IsRead {
				err := s.db.MarkArticleRead(existingArticle.ID, true)
				if err != nil {
					log.Printf("Warning: Failed to mark article as read: %v", err)
				} else {
					updated = true
				}
			}

			// Update favorite status
			if isStarred && !existingArticle.IsFavorite {
				err := s.db.SetArticleFavorite(existingArticle.ID, true)
				if err != nil {
					log.Printf("Warning: Failed to mark article as favorite: %v", err)
				} else {
					updated = true
				}
			}

			if updated {
				log.Printf("Updated existing article from FreshRSS: %s", article.URL)
			}
			continue
		}

		// Create new article
		mrssArticle := &models.Article{
			FeedID:         feedID,
			Title:          article.Title,
			URL:            article.URL,
			Summary:        "",
			PublishedAt:    article.Published,
			IsRead:         isRead,
			IsFavorite:     isStarred,
			FreshRSSItemID: article.ID, // Save FreshRSS/Google Reader item ID
		}

		mrssArticles = append(mrssArticles, mrssArticle)

		// Store content for later insertion
		if article.Content != "" {
			articleContentMap[article.URL] = article.Content
		}
	}

	if skippedCount > 5 {
		log.Printf("Warning: Skipped %d articles due to missing feed mapping", skippedCount)
	}

	if len(mrssArticles) == 0 {
		return 0, nil
	}

	// Save articles to database
	err = s.db.SaveArticles(ctx, mrssArticles)
	if err != nil {
		return 0, fmt.Errorf("save articles: %w", err)
	}

	// Save article contents
	contentSavedCount := 0
	for url, content := range articleContentMap {
		savedArticle, err := s.db.GetArticleByURL(url)
		if err != nil {
			log.Printf("Warning: Could not find saved article with URL %s to set content", url)
			continue
		}

		if err := s.db.SetArticleContent(savedArticle.ID, content); err != nil {
			log.Printf("Warning: Failed to save content for article ID %d: %v", savedArticle.ID, err)
		} else {
			contentSavedCount++
		}
	}

	if contentSavedCount > 0 {
		log.Printf("Saved content for %d articles", contentSavedCount)
	}

	return len(mrssArticles), nil
}

// pushToServer pushes local changes to FreshRSS server
// This compares local vs remote state and immediately syncs any differences
func (s *BidirectionalSyncService) pushToServer(ctx context.Context) (int, error) {
	totalChanges := 0

	// First, process any failed items from the queue (retry mechanism)
	pendingChanges, err := s.db.GetPendingSyncChanges(500)
	if err != nil {
		log.Printf("Warning: Failed to get pending changes: %v", err)
	} else if len(pendingChanges) > 0 {
		log.Printf("[Push] Retrying %d failed sync items from queue", len(pendingChanges))
		for i, item := range pendingChanges {
			log.Printf("  [%d] ArticleID=%d URL=%s Action=%s", i, item.ArticleID, item.ArticleURL, item.Action)
		}
		changes, err := s.pushPendingItems(ctx, pendingChanges)
		if err != nil {
			log.Printf("[Push] Error pushing pending items: %v", err)
			// Continue anyway, don't fail entire sync
		} else {
			totalChanges += changes
		}
	}

	// Get all FreshRSS feeds
	feeds, err := s.db.GetFeeds()
	if err != nil {
		return totalChanges, fmt.Errorf("get feeds: %w", err)
	}

	// Filter to only FreshRSS feeds
	freshRSSFeeds := make([]models.Feed, 0)
	for _, feed := range feeds {
		if feed.IsFreshRSSSource {
			freshRSSFeeds = append(freshRSSFeeds, feed)
		}
	}

	if len(freshRSSFeeds) == 0 {
		log.Printf("[Push] No FreshRSS feeds to sync")
		return totalChanges, nil
	}

	log.Printf("[Push] Checking %d FreshRSS feeds for local changes to push", len(freshRSSFeeds))

	// Get remote read and starred articles
	remoteReadArticles := make(map[string]bool)    // URL -> is read
	remoteStarredArticles := make(map[string]bool) // URL -> is starred

	// Fetch remote read status
	readArticles, err := s.client.GetReadArticles(ctx, 1000)
	if err != nil {
		log.Printf("[Push] Warning: Failed to get remote read articles: %v", err)
	} else {
		for _, article := range readArticles {
			remoteReadArticles[article.URL] = true
		}
		log.Printf("[Push] Fetched %d read articles from remote", len(remoteReadArticles))
	}

	// Fetch remote starred status
	starredArticles, err := s.client.GetStarredArticles(ctx, 1000)
	if err != nil {
		log.Printf("[Push] Warning: Failed to get remote starred articles: %v", err)
	} else {
		for _, article := range starredArticles {
			remoteStarredArticles[article.URL] = true
		}
		log.Printf("[Push] Fetched %d starred articles from remote", len(remoteStarredArticles))
	}

	// For each FreshRSS feed, check articles and sync differences
	readIDs := make([]string, 0)
	unreadIDs := make([]string, 0)
	starIDs := make([]string, 0)
	unstarIDs := make([]string, 0)

	for _, feed := range freshRSSFeeds {
		// Get articles for this feed
		articles, err := s.db.GetArticles("all", feed.ID, "", false, 1000, 0)
		if err != nil {
			log.Printf("[Push] Warning: Failed to get articles for feed %s: %v", feed.Title, err)
			continue
		}

		for _, article := range articles {
			// Use FreshRSS Item ID if available, otherwise fall back to URL
			identifier := article.URL
			if article.FreshRSSItemID != "" {
				identifier = article.FreshRSSItemID
			}

			// Check read status differences
			remoteIsRead := remoteReadArticles[article.URL]
			if article.IsRead && !remoteIsRead {
				// Local is read, remote is not - push read status
				readIDs = append(readIDs, identifier)
			} else if !article.IsRead && remoteIsRead {
				// Local is unread, remote is read - push unread status
				unreadIDs = append(unreadIDs, identifier)
			}

			// Check starred status differences
			remoteIsStarred := remoteStarredArticles[article.URL]
			if article.IsFavorite && !remoteIsStarred {
				// Local is starred, remote is not - push star status
				starIDs = append(starIDs, identifier)
			} else if !article.IsFavorite && remoteIsStarred {
				// Local is unstarred, remote is starred - push unstar status
				unstarIDs = append(unstarIDs, identifier)
			}
		}
	}

	// Execute batch operations
	if len(readIDs) > 0 {
		log.Printf("[Push] Marking %d articles as read (local has read, remote doesn't)", len(readIDs))
		if err := s.client.MarkAsReadBatch(ctx, readIDs); err != nil {
			log.Printf("[Push] ERROR marking as read: %v", err)
			return totalChanges, fmt.Errorf("mark read batch: %w", err)
		}
		log.Printf("[Push] Successfully marked %d articles as read", len(readIDs))
		totalChanges += len(readIDs)
	}

	if len(unreadIDs) > 0 {
		log.Printf("[Push] Marking %d articles as unread (local has unread, remote doesn't)", len(unreadIDs))
		if err := s.client.MarkAsUnreadBatch(ctx, unreadIDs); err != nil {
			log.Printf("[Push] ERROR marking as unread: %v", err)
			return totalChanges, fmt.Errorf("mark unread batch: %w", err)
		}
		log.Printf("[Push] Successfully marked %d articles as unread", len(unreadIDs))
		totalChanges += len(unreadIDs)
	}

	if len(starIDs) > 0 {
		log.Printf("[Push] Starring %d articles (local has starred, remote doesn't)", len(starIDs))
		if err := s.client.StarBatch(ctx, starIDs); err != nil {
			log.Printf("[Push] ERROR starring: %v", err)
			return totalChanges, fmt.Errorf("star batch: %w", err)
		}
		log.Printf("[Push] Successfully starred %d articles", len(starIDs))
		totalChanges += len(starIDs)
	}

	if len(unstarIDs) > 0 {
		log.Printf("[Push] Unstarring %d articles (local has unstarred, remote doesn't)", len(unstarIDs))
		if err := s.client.UnstarBatch(ctx, unstarIDs); err != nil {
			log.Printf("[Push] ERROR unstarring: %v", err)
			return totalChanges, fmt.Errorf("unstar batch: %w", err)
		}
		log.Printf("[Push] Successfully unstarred %d articles", len(unstarIDs))
		totalChanges += len(unstarIDs)
	}

	log.Printf("[Push] Total %d changes synced to server", totalChanges)
	return totalChanges, nil
}

// pushPendingItems pushes items that failed previously (from the queue)
func (s *BidirectionalSyncService) pushPendingItems(ctx context.Context, pendingChanges []database.SyncQueueItem) (int, error) {
	totalChanges := 0

	// Group changes by action type
	readIDs := make([]string, 0)
	unreadIDs := make([]string, 0)
	starIDs := make([]string, 0)
	unstarIDs := make([]string, 0)
	itemIDs := make([]int64, 0)

	// Get article IDs to fetch FreshRSS item IDs
	articleIDs := make([]int64, len(pendingChanges))
	for i, item := range pendingChanges {
		articleIDs[i] = item.ArticleID
	}

	// Fetch articles to get their FreshRSS item IDs
	articles, err := s.db.GetArticlesByIDs(articleIDs)
	if err != nil {
		log.Printf("[PushPending] Failed to fetch articles: %v", err)
		return totalChanges, err
	}

	// Create a map of articles by their ID
	articleByID := make(map[int64]models.Article)
	for _, article := range articles {
		articleByID[article.ID] = article
	}

	// Use FreshRSS item ID if available, otherwise fall back to URL
	for _, item := range pendingChanges {
		itemIDs = append(itemIDs, item.ID)

		article, exists := articleByID[item.ArticleID]
		identifier := item.ArticleURL // Default fallback

		if exists && article.FreshRSSItemID != "" {
			identifier = article.FreshRSSItemID
			log.Printf("  Using FreshRSS Item ID: %s for article %d", identifier, item.ArticleID)
		} else {
			log.Printf("  Warning: No FreshRSS Item ID for article %d, using URL: %s", item.ArticleID, item.ArticleURL)
		}

		switch item.Action {
		case database.SyncActionMarkRead:
			readIDs = append(readIDs, identifier)
		case database.SyncActionMarkUnread:
			unreadIDs = append(unreadIDs, identifier)
		case database.SyncActionStar:
			starIDs = append(starIDs, identifier)
		case database.SyncActionUnstar:
			unstarIDs = append(unstarIDs, identifier)
		}
	}

	// Execute batch operations
	if len(readIDs) > 0 {
		log.Printf("[PushPending] Marking %d articles as read", len(readIDs))
		if err := s.client.MarkAsReadBatch(ctx, readIDs); err != nil {
			log.Printf("[PushPending] ERROR marking as read: %v", err)
			return totalChanges, fmt.Errorf("mark read batch: %w", err)
		}
		log.Printf("[PushPending] Successfully marked %d articles as read", len(readIDs))
		totalChanges += len(readIDs)
	}

	if len(unreadIDs) > 0 {
		if err := s.client.MarkAsUnreadBatch(ctx, unreadIDs); err != nil {
			return totalChanges, fmt.Errorf("mark unread batch: %w", err)
		}
		totalChanges += len(unreadIDs)
	}

	if len(starIDs) > 0 {
		if err := s.client.StarBatch(ctx, starIDs); err != nil {
			return totalChanges, fmt.Errorf("star batch: %w", err)
		}
		totalChanges += len(starIDs)
	}

	if len(unstarIDs) > 0 {
		if err := s.client.UnstarBatch(ctx, unstarIDs); err != nil {
			return totalChanges, fmt.Errorf("unstar batch: %w", err)
		}
		totalChanges += len(unstarIDs)
	}

	// Mark all as synced
	if err := s.db.MarkSynced(itemIDs); err != nil {
		log.Printf("Warning: Failed to mark items as synced: %v", err)
	}

	log.Printf("[PushPending] Successfully synced %d items from queue", totalChanges)

	// Clean up old synced items
	_ = s.db.DeleteOldSyncedItems(7 * 24 * time.Hour)

	return totalChanges, nil
}

// GetPendingCount returns the number of pending sync changes
func (s *BidirectionalSyncService) GetPendingCount() (int, error) {
	return s.db.GetPendingSyncCount()
}

// GetFailedItems returns items that failed to sync
func (s *BidirectionalSyncService) GetFailedItems(limit int) ([]database.SyncQueueItem, error) {
	return s.db.GetFailedSyncItems(limit)
}

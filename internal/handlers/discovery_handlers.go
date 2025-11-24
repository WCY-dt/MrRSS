package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"MrRSS/internal/discovery"
	"MrRSS/internal/models"
)

// HandleDiscoverBlogs discovers blogs from a feed's friend links (SSE with progress)
func (h *Handler) HandleDiscoverBlogs(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleDiscoverBlogs: Request received, method=%s", r.Method)

	if r.Method != http.MethodGet {
		log.Printf("HandleDiscoverBlogs: Invalid method: %s", r.Method)
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	feedID, err := parseInt64QueryParam(r, "feed_id")
	if err != nil {
		log.Printf("HandleDiscoverBlogs: Invalid feed_id parameter: %v", err)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("HandleDiscoverBlogs: Setting up SSE for feed_id=%d", feedID)

	// Set up SSE
	flusher, err := h.setupSSE(w)
	if err != nil {
		log.Printf("HandleDiscoverBlogs: Failed to setup SSE: %v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("HandleDiscoverBlogs: SSE established, sending initial ping")
	// Send initial ping to establish connection
	sendSSEProgress(w, flusher, "Connected, starting discovery...")

	// Get the target feed
	log.Printf("HandleDiscoverBlogs: Fetching feed from database")
	targetFeed, err := h.DB.GetFeedByID(feedID)
	if err != nil {
		log.Printf("HandleDiscoverBlogs: Feed not found: %v", err)
		sendSSEError(w, flusher, "Feed not found")
		return
	}

	log.Printf("HandleDiscoverBlogs: Found feed: %s", targetFeed.Title)

	// Get existing feed URLs for deduplication
	subscribedURLs, err := h.DB.GetAllFeedURLs()
	if err != nil {
		log.Printf("Error getting subscribed URLs: %v", err)
		subscribedURLs = make(map[string]bool)
	}

	// Discover blogs
	ctx, cancel := context.WithTimeout(r.Context(), singleDiscoveryTimeout)
	defer cancel()

	log.Printf("Starting blog discovery for feed: %s (%s), link: %s", targetFeed.Title, targetFeed.URL, targetFeed.Link)

	progressCallback := func(message string) {
		sendSSEProgress(w, flusher, message)
	}

	homepage := targetFeed.Link
	if homepage == "" {
		log.Printf("No link in database for feed %s, will extract from feed URL", targetFeed.Title)
	}

	discovered, err := h.DiscoveryService.DiscoverFromFeed(ctx, targetFeed.URL, homepage, progressCallback)
	if err != nil {
		log.Printf("Error discovering blogs: %v", err)
		sendSSEError(w, flusher, fmt.Sprintf("Failed to discover blogs: %v", err))
		return
	}

	// Filter out already-subscribed feeds
	filtered := h.filterSubscribedFeeds(discovered, subscribedURLs)

	// Mark the feed as discovered
	if err := h.DB.MarkFeedDiscovered(feedID); err != nil {
		log.Printf("Error marking feed as discovered: %v", err)
	}

	log.Printf("Discovered %d blogs, %d after filtering", len(discovered), len(filtered))
	sendSSEComplete(w, flusher, filtered)
}

// HandleDiscoverAllFeeds discovers feeds from all subscriptions that haven't been discovered yet
func (h *Handler) HandleDiscoverAllFeeds(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleDiscoverAllFeeds: Request received, method=%s", r.Method)

	if r.Method != http.MethodGet {
		log.Printf("HandleDiscoverAllFeeds: Invalid method: %s", r.Method)
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	log.Printf("HandleDiscoverAllFeeds: Setting up SSE")

	// Set up SSE
	flusher, err := h.setupSSE(w)
	if err != nil {
		log.Printf("HandleDiscoverAllFeeds: Failed to setup SSE: %v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("HandleDiscoverAllFeeds: SSE established, sending initial ping")
	// Send initial ping to establish connection
	sendSSEProgress(w, flusher, "Connected, preparing batch discovery...")

	// Get feeds that need discovery
	log.Printf("HandleDiscoverAllFeeds: Fetching feeds for discovery")
	feedsToDiscover, err := h.getFeedsForDiscovery()
	if err != nil {
		log.Printf("HandleDiscoverAllFeeds: Error getting feeds: %v", err)
		sendSSEError(w, flusher, err.Error())
		return
	}

	if len(feedsToDiscover) == 0 {
		sendSSEProgress(w, flusher, "All feeds have already been discovered")
		sendSSEComplete(w, flusher, []discovery.DiscoveredBlog{})
		return
	}

	sendSSEProgress(w, flusher, fmt.Sprintf("Starting discovery for %d feeds", len(feedsToDiscover)))

	// Get existing feed URLs for deduplication
	subscribedURLs, err := h.DB.GetAllFeedURLs()
	if err != nil {
		log.Printf("Error getting subscribed URLs: %v", err)
		subscribedURLs = make(map[string]bool)
	}

	// Discover feeds with timeout
	ctx, cancel := context.WithTimeout(r.Context(), batchDiscoveryTimeout)
	defer cancel()

	allDiscovered := h.discoverFromMultipleFeeds(ctx, w, flusher, feedsToDiscover, subscribedURLs)

	log.Printf("Batch discovery complete: discovered %d feeds from %d sources",
		h.countDiscoveredFeeds(allDiscovered), len(feedsToDiscover))

	// Flatten and send results
	allFeeds := h.flattenDiscoveredFeeds(allDiscovered)
	sendSSEProgress(w, flusher, fmt.Sprintf("Completed: Found %d new feeds from %d sources",
		len(allFeeds), len(feedsToDiscover)))
	sendSSEComplete(w, flusher, allFeeds)
}

func (h *Handler) setupSSE(w http.ResponseWriter) (http.Flusher, error) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("streaming unsupported")
	}

	// Write initial comment to establish connection
	// This ensures the browser knows it's a valid SSE stream
	fmt.Fprintf(w, ": SSE connection established\n\n")
	flusher.Flush()

	return flusher, nil
}

func (h *Handler) getFeedsForDiscovery() ([]models.Feed, error) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		return nil, err
	}

	var feedsToDiscover []models.Feed
	for _, feed := range feeds {
		if !feed.DiscoveryCompleted {
			feedsToDiscover = append(feedsToDiscover, feed)
		}
	}

	return feedsToDiscover, nil
}

func (h *Handler) filterSubscribedFeeds(discovered []discovery.DiscoveredBlog, subscribedURLs map[string]bool) []discovery.DiscoveredBlog {
	filtered := make([]discovery.DiscoveredBlog, 0)
	for _, blog := range discovered {
		if !subscribedURLs[blog.RSSFeed] {
			filtered = append(filtered, blog)
		} else {
			log.Printf("Filtering out already-subscribed feed: %s (%s)", blog.Name, blog.RSSFeed)
		}
	}
	return filtered
}

func (h *Handler) discoverFromMultipleFeeds(
	ctx context.Context,
	w http.ResponseWriter,
	flusher http.Flusher,
	feeds []models.Feed,
	subscribedURLs map[string]bool,
) map[string][]discovery.DiscoveredBlog {
	allDiscovered := make(map[string][]discovery.DiscoveredBlog)
	log.Printf("Starting batch discovery for %d feeds", len(feeds))

	for i, feed := range feeds {
		select {
		case <-ctx.Done():
			log.Println("Batch discovery cancelled: timeout")
			sendSSEProgress(w, flusher, "Discovery timed out")
			return allDiscovered
		default:
		}

		sendSSEProgress(w, flusher, fmt.Sprintf("Processing feed %d/%d: %s", i+1, len(feeds), feed.Title))
		log.Printf("Discovering from feed: %s (%s), link: %s", feed.Title, feed.URL, feed.Link)

		progressCallback := func(message string) {
			sendSSEProgress(w, flusher, fmt.Sprintf("  %s", message))
		}

		homepage := feed.Link
		if homepage == "" {
			log.Printf("No link in database for feed %s, will extract from feed URL", feed.Title)
		}

		discovered, err := h.DiscoveryService.DiscoverFromFeed(ctx, feed.URL, homepage, progressCallback)
		if err != nil {
			log.Printf("Error discovering from feed %s: %v", feed.Title, err)
			sendSSEProgress(w, flusher, fmt.Sprintf("  Error: %v", err))
			continue
		}

		// Filter and store results
		filtered := h.filterSubscribedFeeds(discovered, subscribedURLs)
		if len(filtered) > 0 {
			allDiscovered[feed.Title] = filtered
			sendSSEProgress(w, flusher, fmt.Sprintf("  Found %d new feeds", len(filtered)))
		} else {
			sendSSEProgress(w, flusher, "  No new feeds found")
		}

		// Mark the feed as discovered
		if err := h.DB.MarkFeedDiscovered(feed.ID); err != nil {
			log.Printf("Error marking feed as discovered: %v", err)
		}
	}

	return allDiscovered
}

func (h *Handler) countDiscoveredFeeds(discovered map[string][]discovery.DiscoveredBlog) int {
	count := 0
	for _, feeds := range discovered {
		count += len(feeds)
	}
	return count
}

func (h *Handler) flattenDiscoveredFeeds(discovered map[string][]discovery.DiscoveredBlog) []discovery.DiscoveredBlog {
	var allFeeds []discovery.DiscoveredBlog
	for _, feeds := range discovered {
		allFeeds = append(allFeeds, feeds...)
	}
	return allFeeds
}

// SSE helper functions for sending progress updates
func sendSSEProgress(w http.ResponseWriter, flusher http.Flusher, message string) {
	data, err := json.Marshal(map[string]string{
		"type":    "progress",
		"message": message,
	})
	if err != nil {
		log.Printf("sendSSEProgress: JSON marshal error: %v", err)
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
	log.Printf("sendSSEProgress: Sent progress message: %s", message)
}

func sendSSEError(w http.ResponseWriter, flusher http.Flusher, message string) {
	log.Printf("sendSSEError: Sending error: %s", message)
	data, err := json.Marshal(map[string]string{
		"type":    "error",
		"message": message,
	})
	if err != nil {
		log.Printf("sendSSEError: JSON marshal error: %v", err)
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

func sendSSEComplete(w http.ResponseWriter, flusher http.Flusher, feeds []discovery.DiscoveredBlog) {
	log.Printf("sendSSEComplete: Sending %d feeds", len(feeds))
	data, err := json.Marshal(map[string]interface{}{
		"type":  "complete",
		"feeds": feeds,
	})
	if err != nil {
		log.Printf("sendSSEComplete: JSON marshal error: %v", err)
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

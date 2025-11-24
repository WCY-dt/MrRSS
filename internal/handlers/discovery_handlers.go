package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"MrRSS/internal/discovery"
	"MrRSS/internal/models"
)

// HandleDiscoverBlogs discovers blogs from a feed's friend links (SSE with progress)
func (h *Handler) HandleDiscoverBlogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	feedIDStr := r.URL.Query().Get("feed_id")
	feedID, err := strconv.ParseInt(feedIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid feed_id")
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Get the target feed
	targetFeed, err := h.DB.GetFeedByID(feedID)
	if err != nil {
		sendSSEError(w, flusher, "Feed not found")
		return
	}

	// Get existing feed URLs for deduplication
	subscribedURLs, err := h.DB.GetAllFeedURLs()
	if err != nil {
		log.Printf("Error getting subscribed URLs: %v", err)
		subscribedURLs = make(map[string]bool)
	}

	// Discover blogs
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Send progress updates via SSE
	progressCallback := func(message string) {
		sendSSEProgress(w, flusher, message)
	}

	homepage := targetFeed.Link
	discovered, err := h.DiscoveryService.DiscoverFromFeed(ctx, targetFeed.URL, homepage, progressCallback)
	if err != nil {
		sendSSEError(w, flusher, fmt.Sprintf("Failed to discover blogs: %v", err))
		return
	}

	// Filter out already-subscribed feeds
	filtered := make([]discovery.DiscoveredBlog, 0)
	for _, blog := range discovered {
		if !subscribedURLs[blog.RSSFeed] {
			filtered = append(filtered, blog)
		} else {
			log.Printf("Filtering out already-subscribed feed: %s (%s)", blog.Name, blog.RSSFeed)
		}
	}

	// Mark the feed as discovered
	if err := h.DB.MarkFeedDiscovered(feedID); err != nil {
		log.Printf("Error marking feed as discovered: %v", err)
	}

	// Send final results
	sendSSEComplete(w, flusher, filtered)
}

// HandleDiscoverAllFeeds discovers feeds from all subscriptions that haven't been discovered yet
func (h *Handler) HandleDiscoverAllFeeds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Get feeds that need discovery
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		sendSSEError(w, flusher, err.Error())
		return
	}

	var feedsToDiscover []models.Feed
	for _, feed := range feeds {
		if !feed.DiscoveryCompleted {
			feedsToDiscover = append(feedsToDiscover, feed)
		}
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
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	allDiscovered := make(map[string][]discovery.DiscoveredBlog)

	for i, feed := range feedsToDiscover {
		select {
		case <-ctx.Done():
			sendSSEProgress(w, flusher, "Discovery timed out")
			break
		default:
		}

		sendSSEProgress(w, flusher, fmt.Sprintf("Processing feed %d/%d: %s", i+1, len(feedsToDiscover), feed.Title))

		progressCallback := func(message string) {
			sendSSEProgress(w, flusher, fmt.Sprintf("  %s", message))
		}

		homepage := feed.Link
		discovered, err := h.DiscoveryService.DiscoverFromFeed(ctx, feed.URL, homepage, progressCallback)
		if err != nil {
			sendSSEProgress(w, flusher, fmt.Sprintf("  Error: %v", err))
			continue
		}

		// Filter and store results
		filtered := make([]discovery.DiscoveredBlog, 0)
		for _, blog := range discovered {
			if !subscribedURLs[blog.RSSFeed] {
				filtered = append(filtered, blog)
			}
		}

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

	// Flatten and send results
	var allFeeds []discovery.DiscoveredBlog
	for _, feeds := range allDiscovered {
		allFeeds = append(allFeeds, feeds...)
	}

	sendSSEProgress(w, flusher, fmt.Sprintf("Completed: Found %d new feeds from %d sources",
		len(allFeeds), len(feedsToDiscover)))
	sendSSEComplete(w, flusher, allFeeds)
}

// SSE helper functions
func sendSSEProgress(w http.ResponseWriter, flusher http.Flusher, message string) {
	data, _ := json.Marshal(map[string]string{
		"type":    "progress",
		"message": message,
	})
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

func sendSSEError(w http.ResponseWriter, flusher http.Flusher, message string) {
	data, _ := json.Marshal(map[string]string{
		"type":    "error",
		"message": message,
	})
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

func sendSSEComplete(w http.ResponseWriter, flusher http.Flusher, feeds []discovery.DiscoveredBlog) {
	data, _ := json.Marshal(map[string]interface{}{
		"type":  "complete",
		"feeds": feeds,
	})
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

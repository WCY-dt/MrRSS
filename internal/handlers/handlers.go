// Package handlers contains the HTTP handlers for the application.
// It defines the Handler struct which holds dependencies like the database and fetcher.
package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/discovery"
	"MrRSS/internal/feed"
	"MrRSS/internal/translation"
	"MrRSS/internal/version"
)

// Constants for configuration and defaults
const (
	// GitHub API and repository configuration
	githubAPILatestRelease = "https://api.github.com/repos/WCY-dt/MrRSS/releases/latest"
	allowedURLPrefix       = "https://github.com/WCY-dt/MrRSS/releases/download/"

	// Default values for settings
	defaultUpdateInterval  = 10
	defaultArticleLimit    = 50
	defaultArticlesPerPage = 1

	// File handling
	downloadBufferSize = 32 * 1024 // 32KB

	// Timeout durations
	feedFetchTimeout       = 30 * time.Second
	batchDiscoveryTimeout  = 5 * time.Minute
	singleDiscoveryTimeout = 60 * time.Second

	// Cleanup delays
	windowsCleanupDelay = 10 * time.Second
	linuxCleanupDelay   = 10 * time.Second
	macosCleanupDelay   = 15 * time.Second
	shutdownDelay       = 2 * time.Second
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	DB               *database.DB
	Fetcher          *feed.Fetcher
	Translator       translation.Translator
	DiscoveryService *discovery.Service
}

func NewHandler(db *database.DB, fetcher *feed.Fetcher, translator translation.Translator) *Handler {
	return &Handler{
		DB:               db,
		Fetcher:          fetcher,
		Translator:       translator,
		DiscoveryService: discovery.NewService(),
	}
}

func (h *Handler) StartBackgroundScheduler(ctx context.Context) {
	// Run initial cleanup only if auto_cleanup is enabled
	go func() {
		autoCleanup, _ := h.DB.GetSetting("auto_cleanup_enabled")
		if autoCleanup == "true" {
			log.Println("Running initial article cleanup...")
			count, err := h.DB.CleanupOldArticles()
			if err != nil {
				log.Printf("Error during initial cleanup: %v", err)
			} else {
				log.Printf("Initial cleanup: removed %d old articles", count)
			}
		}
	}()

	for {
		intervalStr, err := h.DB.GetSetting("update_interval")
		interval := 10
		if err == nil {
			if i, err := strconv.Atoi(intervalStr); err == nil && i > 0 {
				interval = i
			}
		}

		log.Printf("Next auto-update in %d minutes", interval)

		select {
		case <-ctx.Done():
			log.Println("Stopping background scheduler")
			return
		case <-time.After(time.Duration(interval) * time.Minute):
			h.Fetcher.FetchAll(ctx)
			// Run cleanup after fetching new articles only if auto_cleanup is enabled
			go func() {
				autoCleanup, _ := h.DB.GetSetting("auto_cleanup_enabled")
				if autoCleanup == "true" {
					count, err := h.DB.CleanupOldArticles()
					if err != nil {
						log.Printf("Error during automatic cleanup: %v", err)
					} else if count > 0 {
						log.Printf("Automatic cleanup: removed %d old articles", count)
					}
				}
			}()
		}
	}
}

func (h *Handler) HandleProgress(w http.ResponseWriter, r *http.Request) {
	progress := h.Fetcher.GetProgress()
	respondWithJSON(w, http.StatusOK, progress)
}

func (h *Handler) HandleGetUnreadCounts(w http.ResponseWriter, r *http.Request) {
	// Get total unread count
	totalCount, err := h.DB.GetTotalUnreadCount()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get unread counts per feed
	feedCounts, err := h.DB.GetUnreadCountsForAllFeeds()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"total":       totalCount,
		"feed_counts": feedCounts,
	}
	respondWithJSON(w, http.StatusOK, response)
}

func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	go h.Fetcher.FetchAll(context.Background())
	respondOK(w)
}

func (h *Handler) HandleCleanupArticles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	count, err := h.DB.CleanupUnimportantArticles()
	if err != nil {
		log.Printf("Error cleaning up articles: %v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("Cleaned up %d articles", count)
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"deleted": count,
	})
}

// HandleVersion returns the current application version
func (h *Handler) HandleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"version": version.Version,
	})
}

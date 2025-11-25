package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"MrRSS/internal/database"
	"MrRSS/internal/feed"
	"MrRSS/internal/handlers"
	"MrRSS/internal/translation"
	"MrRSS/internal/utils"
)

//go:embed frontend/dist/*
var frontendFiles embed.FS

type CombinedHandler struct {
	apiMux     *http.ServeMux
	fileServer http.Handler
}

func (h *CombinedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		h.apiMux.ServeHTTP(w, r)
		return
	}
	h.fileServer.ServeHTTP(w, r)
}

func main() {
	// Get proper paths for data files
	logPath, err := utils.GetLogPath()
	if err != nil {
		log.Printf("Warning: Could not get log path: %v. Using current directory.", err)
		logPath = "debug.log"
	}
	
	f, _ := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()
	log.SetOutput(f)

	log.Println("Starting application...")
	log.Printf("Log file: %s", logPath)

	// Get database path
	dbPath, err := utils.GetDBPath()
	if err != nil {
		log.Printf("Error getting database path: %v", err)
		log.Fatal(err)
	}
	log.Printf("Database path: %s", dbPath)

	// Initialize database
	log.Println("Initializing Database...")
	db, err := database.NewDB(dbPath)
	if err != nil {
		log.Printf("Error initializing database: %v", err)
		log.Fatal(err)
	}
	
	// Run database schema initialization synchronously to ensure it's ready
	log.Println("Running DB migrations...")
	if err := db.Init(); err != nil {
		log.Printf("Error initializing database schema: %v", err)
		log.Fatal(err)
	}
	log.Println("Database initialized successfully")

	translator := translation.NewGoogleFreeTranslator()
	fetcher := feed.NewFetcher(db, translator)
	h := handlers.NewHandler(db, fetcher, translator)

	// API Routes
	log.Println("Setting up API routes...")
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/feeds", h.HandleFeeds)
	apiMux.HandleFunc("/api/feeds/add", h.HandleAddFeed)
	apiMux.HandleFunc("/api/feeds/delete", h.HandleDeleteFeed)
	apiMux.HandleFunc("/api/feeds/update", h.HandleUpdateFeed)
	apiMux.HandleFunc("/api/feeds/discover", h.HandleDiscoverBlogs)
	apiMux.HandleFunc("/api/feeds/discover-all", h.HandleDiscoverAllFeeds)
	// Polling-based discovery endpoints (Wails-compatible)
	apiMux.HandleFunc("/api/feeds/discover/start", h.HandleStartSingleDiscovery)
	apiMux.HandleFunc("/api/feeds/discover/progress", h.HandleGetSingleDiscoveryProgress)
	apiMux.HandleFunc("/api/feeds/discover/clear", h.HandleClearSingleDiscovery)
	apiMux.HandleFunc("/api/feeds/discover-all/start", h.HandleStartBatchDiscovery)
	apiMux.HandleFunc("/api/feeds/discover-all/progress", h.HandleGetBatchDiscoveryProgress)
	apiMux.HandleFunc("/api/feeds/discover-all/clear", h.HandleClearBatchDiscovery)
	apiMux.HandleFunc("/api/articles", h.HandleArticles)
	apiMux.HandleFunc("/api/articles/filter", h.HandleFilteredArticles)
	apiMux.HandleFunc("/api/articles/read", h.HandleMarkRead)
	apiMux.HandleFunc("/api/articles/favorite", h.HandleToggleFavorite)
	apiMux.HandleFunc("/api/articles/cleanup", h.HandleCleanupArticles)
	apiMux.HandleFunc("/api/articles/translate", h.HandleTranslateArticle)
	apiMux.HandleFunc("/api/articles/clear-translations", h.HandleClearTranslations)
	apiMux.HandleFunc("/api/articles/toggle-hide", h.HandleToggleHideArticle)
	apiMux.HandleFunc("/api/articles/content", h.HandleGetArticleContent)
	apiMux.HandleFunc("/api/articles/unread-counts", h.HandleGetUnreadCounts)
	apiMux.HandleFunc("/api/articles/mark-all-read", h.HandleMarkAllAsRead)
	apiMux.HandleFunc("/api/settings", h.HandleSettings)
	apiMux.HandleFunc("/api/refresh", h.HandleRefresh)
	apiMux.HandleFunc("/api/progress", h.HandleProgress)
	apiMux.HandleFunc("/api/opml/import", h.HandleOPMLImport)
	apiMux.HandleFunc("/api/opml/export", h.HandleOPMLExport)
	apiMux.HandleFunc("/api/check-updates", h.HandleCheckUpdates)
	apiMux.HandleFunc("/api/download-update", h.HandleDownloadUpdate)
	apiMux.HandleFunc("/api/install-update", h.HandleInstallUpdate)
	apiMux.HandleFunc("/api/version", h.HandleVersion)

	// Static Files
	log.Println("Setting up static files...")
	frontendFS, err := fs.Sub(frontendFiles, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	fileServer := http.FileServer(http.FS(frontendFS))

	combinedHandler := &CombinedHandler{
		apiMux:     apiMux,
		fileServer: fileServer,
	}

	// Start background scheduler
	log.Println("Starting background scheduler...")
	bgCtx, bgCancel := context.WithCancel(context.Background())

	log.Println("Starting Wails...")
	err = wails.Run(&options.App{
		Title:            "MrRSS",
		Width:            1024,
		Height:           768,
		WindowStartState: options.Maximised,
		LogLevel:         logger.DEBUG,
		AssetServer: &assetserver.Options{
			Assets:  frontendFS,
			Handler: combinedHandler,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnShutdown: func(ctx context.Context) {
			log.Println("Shutting down...")

			// Stop background tasks first
			bgCancel()
			// Give some time for tasks to finish
			time.Sleep(500 * time.Millisecond)

			// Close DB with timeout
			done := make(chan struct{})
			go func() {
				if err := db.Close(); err != nil {
					log.Printf("Error closing database: %v", err)
				}
				close(done)
			}()

			select {
			case <-done:
				log.Println("Database closed")
			case <-time.After(2 * time.Second):
				log.Println("Database close timed out")
			}
		},
		OnStartup: func(ctx context.Context) {
			log.Println("App started")

			// Start background scheduler after a longer delay to allow UI to show first
			go func() {
				time.Sleep(5 * time.Second)
				log.Println("Starting background scheduler...")
				h.StartBackgroundScheduler(bgCtx)
			}()
		},
	})

	if err != nil {
		log.Printf("Error running Wails: %v", err)
		log.Fatal(err)
	}
	log.Println("Application finished")
}

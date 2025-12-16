package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"MrRSS/internal/database"
	"MrRSS/internal/feed"
	article "MrRSS/internal/handlers/article"
	handlers "MrRSS/internal/handlers/core"
	discovery "MrRSS/internal/handlers/discovery"
	feedhandlers "MrRSS/internal/handlers/feed"
	media "MrRSS/internal/handlers/media"
	networkhandlers "MrRSS/internal/handlers/network"
	opml "MrRSS/internal/handlers/opml"
	platform "MrRSS/internal/handlers/platform"
	rules "MrRSS/internal/handlers/rules"
	script "MrRSS/internal/handlers/script"
	settings "MrRSS/internal/handlers/settings"
	summary "MrRSS/internal/handlers/summary"
	translationhandlers "MrRSS/internal/handlers/translation"
	update "MrRSS/internal/handlers/update"
	window "MrRSS/internal/handlers/window"
	"MrRSS/internal/network"
	"MrRSS/internal/translation"
	"MrRSS/internal/tray"
	"MrRSS/internal/utils"
)

//go:embed frontend/dist/*
var frontendFiles embed.FS

//go:embed build/appicon.png
var trayIconPNG []byte

//go:embed build/windows/icon.ico
var trayIconICO []byte

type windowState struct {
	width  int
	height int
	x      int
	y      int
	valid  atomic.Bool
}

type atomicContext struct {
	value atomic.Value // stores context.Context
}

func (ac *atomicContext) Store(ctx context.Context) {
	ac.value.Store(ctx)
}

func (ac *atomicContext) Load() context.Context {
	val := ac.value.Load()
	if val == nil {
		return nil
	}
	return val.(context.Context)
}

// getTrayIcon returns the appropriate icon bytes for the current platform
func getTrayIcon() []byte {
	// Windows requires .ico format, other platforms use .png
	if utils.IsWindows() {
		return trayIconICO
	}
	return trayIconPNG
}

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

	// Log portable mode status
	if utils.IsPortableMode() {
		log.Println("Running in PORTABLE mode")
	} else {
		log.Println("Running in NORMAL mode")
	}

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

	translator := translation.NewDynamicTranslator(db)
	fetcher := feed.NewFetcher(db, translator)
	h := handlers.NewHandler(db, fetcher, translator)

	// Use platform-specific icon format
	trayIconBytes := getTrayIcon()
	trayManager := tray.NewManager(h, trayIconBytes)
	var quitRequested atomic.Bool
	var lastWindowState windowState

	// API Routes
	log.Println("Setting up API routes...")
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/feeds", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleFeeds(h, w, r) })
	apiMux.HandleFunc("/api/feeds/add", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleAddFeed(h, w, r) })
	apiMux.HandleFunc("/api/feeds/delete", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleDeleteFeed(h, w, r) })
	apiMux.HandleFunc("/api/feeds/update", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleUpdateFeed(h, w, r) })
	apiMux.HandleFunc("/api/feeds/refresh", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleRefreshFeed(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover", func(w http.ResponseWriter, r *http.Request) { discovery.HandleDiscoverBlogs(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover-all", func(w http.ResponseWriter, r *http.Request) { discovery.HandleDiscoverAllFeeds(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover/start", func(w http.ResponseWriter, r *http.Request) { discovery.HandleStartSingleDiscovery(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover/progress", func(w http.ResponseWriter, r *http.Request) { discovery.HandleGetSingleDiscoveryProgress(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover/clear", func(w http.ResponseWriter, r *http.Request) { discovery.HandleClearSingleDiscovery(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover-all/start", func(w http.ResponseWriter, r *http.Request) { discovery.HandleStartBatchDiscovery(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover-all/progress", func(w http.ResponseWriter, r *http.Request) { discovery.HandleGetBatchDiscoveryProgress(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover-all/clear", func(w http.ResponseWriter, r *http.Request) { discovery.HandleClearBatchDiscovery(h, w, r) })
	apiMux.HandleFunc("/api/articles", func(w http.ResponseWriter, r *http.Request) { article.HandleArticles(h, w, r) })
	apiMux.HandleFunc("/api/articles/images", func(w http.ResponseWriter, r *http.Request) { article.HandleImageGalleryArticles(h, w, r) })
	apiMux.HandleFunc("/api/articles/filter", func(w http.ResponseWriter, r *http.Request) { article.HandleFilteredArticles(h, w, r) })
	apiMux.HandleFunc("/api/articles/read", func(w http.ResponseWriter, r *http.Request) { article.HandleMarkRead(h, w, r) })
	apiMux.HandleFunc("/api/articles/favorite", func(w http.ResponseWriter, r *http.Request) { article.HandleToggleFavorite(h, w, r) })
	apiMux.HandleFunc("/api/articles/cleanup", func(w http.ResponseWriter, r *http.Request) { article.HandleCleanupArticles(h, w, r) })
	apiMux.HandleFunc("/api/articles/translate", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleTranslateArticle(h, w, r) })
	apiMux.HandleFunc("/api/articles/translate-text", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleTranslateText(h, w, r) })
	apiMux.HandleFunc("/api/articles/clear-translations", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleClearTranslations(h, w, r) })
	apiMux.HandleFunc("/api/articles/toggle-hide", func(w http.ResponseWriter, r *http.Request) { article.HandleToggleHideArticle(h, w, r) })
	apiMux.HandleFunc("/api/articles/toggle-read-later", func(w http.ResponseWriter, r *http.Request) { article.HandleToggleReadLater(h, w, r) })
	apiMux.HandleFunc("/api/articles/content", func(w http.ResponseWriter, r *http.Request) { article.HandleGetArticleContent(h, w, r) })
	apiMux.HandleFunc("/api/articles/unread-counts", func(w http.ResponseWriter, r *http.Request) { article.HandleGetUnreadCounts(h, w, r) })
	apiMux.HandleFunc("/api/articles/mark-all-read", func(w http.ResponseWriter, r *http.Request) { article.HandleMarkAllAsRead(h, w, r) })
	apiMux.HandleFunc("/api/articles/clear-read-later", func(w http.ResponseWriter, r *http.Request) { article.HandleClearReadLater(h, w, r) })
	apiMux.HandleFunc("/api/articles/summarize", func(w http.ResponseWriter, r *http.Request) { summary.HandleSummarizeArticle(h, w, r) })
	apiMux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) { settings.HandleSettings(h, w, r) })
	apiMux.HandleFunc("/api/refresh", func(w http.ResponseWriter, r *http.Request) { article.HandleRefresh(h, w, r) })
	apiMux.HandleFunc("/api/progress", func(w http.ResponseWriter, r *http.Request) { article.HandleProgress(h, w, r) })
	apiMux.HandleFunc("/api/opml/import", func(w http.ResponseWriter, r *http.Request) { opml.HandleOPMLImport(h, w, r) })
	apiMux.HandleFunc("/api/opml/export", func(w http.ResponseWriter, r *http.Request) { opml.HandleOPMLExport(h, w, r) })
	apiMux.HandleFunc("/api/check-updates", func(w http.ResponseWriter, r *http.Request) { update.HandleCheckUpdates(h, w, r) })
	apiMux.HandleFunc("/api/download-update", func(w http.ResponseWriter, r *http.Request) { update.HandleDownloadUpdate(h, w, r) })
	apiMux.HandleFunc("/api/install-update", func(w http.ResponseWriter, r *http.Request) { update.HandleInstallUpdate(h, w, r) })
	apiMux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) { update.HandleVersion(h, w, r) })
	apiMux.HandleFunc("/api/rules/apply", func(w http.ResponseWriter, r *http.Request) { rules.HandleApplyRule(h, w, r) })
	apiMux.HandleFunc("/api/scripts/dir", func(w http.ResponseWriter, r *http.Request) { script.HandleGetScriptsDir(h, w, r) })
	apiMux.HandleFunc("/api/scripts/open", func(w http.ResponseWriter, r *http.Request) { script.HandleOpenScriptsDir(h, w, r) })
	apiMux.HandleFunc("/api/scripts/list", func(w http.ResponseWriter, r *http.Request) { script.HandleListScripts(h, w, r) })
	apiMux.HandleFunc("/api/media/proxy", func(w http.ResponseWriter, r *http.Request) { media.HandleMediaProxy(h, w, r) })
	apiMux.HandleFunc("/api/media/cleanup", func(w http.ResponseWriter, r *http.Request) { media.HandleMediaCacheCleanup(h, w, r) })
	apiMux.HandleFunc("/api/media/info", func(w http.ResponseWriter, r *http.Request) { media.HandleMediaCacheInfo(h, w, r) })
	apiMux.HandleFunc("/api/window/state", func(w http.ResponseWriter, r *http.Request) { window.HandleGetWindowState(h, w, r) })
	apiMux.HandleFunc("/api/window/save", func(w http.ResponseWriter, r *http.Request) { window.HandleSaveWindowState(h, w, r) })
	apiMux.HandleFunc("/api/network/detect", func(w http.ResponseWriter, r *http.Request) { networkhandlers.HandleDetectNetwork(h, w, r) })
	apiMux.HandleFunc("/api/network/info", func(w http.ResponseWriter, r *http.Request) { networkhandlers.HandleGetNetworkInfo(h, w, r) })
	apiMux.HandleFunc("/api/platform/info", func(w http.ResponseWriter, r *http.Request) { platform.HandleGetPlatformInfo(h, w, r) })

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

	shouldCloseToTray := func() bool {
		val, err := db.GetSetting("close_to_tray")
		return err == nil && val == "true"
	}

	startTray := func(ctx context.Context) {
		if trayManager == nil || trayManager.IsRunning() {
			return
		}
		trayManager.Start(ctx, func() {
			quitRequested.Store(true)
			runtime.Quit(ctx)
		}, func() {
			if lastWindowState.valid.Load() {
				// Validate window state before restoring
				width := lastWindowState.width
				height := lastWindowState.height
				x := lastWindowState.x
				y := lastWindowState.y

				// Ensure minimum window size
				if width < 400 {
					width = 1024
				}
				if height < 300 {
					height = 768
				}

				// Ensure window is at least partially on screen
				// Allow some negative values for multi-monitor setups, but not extreme ones
				if x < -1000 || x > 3000 {
					x = 100
				}
				if y < -1000 || y > 3000 {
					y = 100
				}

				log.Printf("Restoring window state: x=%d, y=%d, width=%d, height=%d", x, y, width, height)
				runtime.WindowSetSize(ctx, width, height)
				runtime.WindowSetPosition(ctx, x, y)
			} else {
				// No valid state, use safe defaults
				log.Println("No valid window state, using defaults")
				runtime.WindowSetSize(ctx, 1024, 768)
				runtime.WindowCenter(ctx)
			}
			runtime.WindowShow(ctx)
			runtime.WindowUnminimise(ctx)
		})
	}

	storeWindowState := func(ctx context.Context) {
		if ctx == nil {
			return
		}

		// Don't store window state if in fullscreen mode (macOS issue workaround)
		isFullscreen := runtime.WindowIsFullscreen(ctx)
		if isFullscreen {
			log.Println("Window is in fullscreen mode, skipping state storage")
			return
		}

		w, h := runtime.WindowGetSize(ctx)
		x, y := runtime.WindowGetPosition(ctx)

		// Only store state if it's valid (reasonable size and position)
		if w >= 400 && h >= 300 && w <= 4000 && h <= 3000 {
			if x > -1000 && x < 3000 && y > -1000 && y < 3000 {
				lastWindowState.width = w
				lastWindowState.height = h
				lastWindowState.x = x
				lastWindowState.y = y
				lastWindowState.valid.Store(true)
				log.Printf("Stored window state: x=%d, y=%d, width=%d, height=%d", x, y, w, h)
			} else {
				log.Printf("Window position invalid (x=%d, y=%d), not storing", x, y)
			}
		} else {
			log.Printf("Window size invalid (width=%d, height=%d), not storing", w, h)
		}
	}

	// Start background scheduler
	log.Println("Starting background scheduler...")
	bgCtx, bgCancel := context.WithCancel(context.Background())

	// Store the app context for single instance callback (thread-safe)
	var appCtx atomicContext

	log.Println("Starting Wails...")
	err = wails.Run(&options.App{
		Title:    "MrRSS",
		Width:    1024,
		Height:   768,
		LogLevel: logger.DEBUG,
		AssetServer: &assetserver.Options{
			Assets:  frontendFS,
			Handler: combinedHandler,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: false, // Set to false to fix window dragging issues on macOS
				HideTitle:                  true,
				HideTitleBar:               false,
				FullSizeContent:            true,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameAqua,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "MrRSS",
				Message: "A modern, privacy-focused RSS reader\n\nCopyright © 2025 MrRSS Team",
				Icon:    trayIconPNG,
			},
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "com.mrrss.app",
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				log.Printf("Second instance detected, bringing window to front")
				ctx := appCtx.Load()
				if ctx != nil {
					// Restore window state if it was stored (minimized to tray)
					if lastWindowState.valid.Load() {
						width := lastWindowState.width
						height := lastWindowState.height
						x := lastWindowState.x
						y := lastWindowState.y

						// Ensure minimum window size
						if width < 400 {
							width = 1024
						}
						if height < 300 {
							height = 768
						}

						// Ensure window is at least partially on screen
						if x < -1000 || x > 3000 {
							x = 100
						}
						if y < -1000 || y > 3000 {
							y = 100
						}

						log.Printf("Restoring window state: x=%d, y=%d, width=%d, height=%d", x, y, width, height)
						runtime.WindowSetSize(ctx, width, height)
						runtime.WindowSetPosition(ctx, x, y)
					}
					// Show and unminimize the window
					runtime.WindowShow(ctx)
					runtime.WindowUnminimise(ctx)
				}
			},
		},
		OnShutdown: func(ctx context.Context) {
			log.Println("Shutting down...")

			if trayManager != nil {
				trayManager.Stop()
			}

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
			// Store context for single instance callback (thread-safe)
			appCtx.Store(ctx)

			// Try to restore window state from database
			restoredFromDB := false
			if x, err := db.GetSetting("window_x"); err == nil && x != "" {
				if y, err := db.GetSetting("window_y"); err == nil && y != "" {
					if width, err := db.GetSetting("window_width"); err == nil && width != "" {
						if height, err := db.GetSetting("window_height"); err == nil && height != "" {
							// Parse values
							var xInt, yInt, widthInt, heightInt int
							if _, err := fmt.Sscanf(x, "%d", &xInt); err == nil {
								if _, err := fmt.Sscanf(y, "%d", &yInt); err == nil {
									if _, err := fmt.Sscanf(width, "%d", &widthInt); err == nil {
										if _, err := fmt.Sscanf(height, "%d", &heightInt); err == nil {
											// Validate values
											if widthInt >= 400 && heightInt >= 300 && widthInt <= 4000 && heightInt <= 3000 {
												if xInt > -1000 && xInt < 3000 && yInt > -1000 && yInt < 3000 {
													log.Printf("Restoring window state from database: x=%d, y=%d, width=%d, height=%d", xInt, yInt, widthInt, heightInt)
													runtime.WindowSetSize(ctx, widthInt, heightInt)
													runtime.WindowSetPosition(ctx, xInt, yInt)
													restoredFromDB = true
													// Store in memory for minimize-restore
													lastWindowState.width = widthInt
													lastWindowState.height = heightInt
													lastWindowState.x = xInt
													lastWindowState.y = yInt
													lastWindowState.valid.Store(true)
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}

			if !restoredFromDB {
				// Use default size if restoration failed
				log.Println("No saved window state found, using defaults")
				runtime.WindowSetSize(ctx, 1024, 768)
				runtime.WindowCenter(ctx)
			}

			runtime.WindowShow(ctx)
			log.Println("Window initialized")

			if shouldCloseToTray() {
				startTray(ctx)
			}

			// Detect network speed on startup in background
			go func() {
				log.Println("Detecting network speed...")
				detector := network.NewDetector()
				detectCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()

				result := detector.DetectSpeed(detectCtx)
				if result.DetectionSuccess {
					db.SetSetting("network_speed", string(result.SpeedLevel))
					db.SetSetting("network_bandwidth_mbps", fmt.Sprintf("%.2f", result.BandwidthMbps))
					db.SetSetting("network_latency_ms", strconv.FormatInt(result.LatencyMs, 10))
					db.SetSetting("max_concurrent_refreshes", strconv.Itoa(result.MaxConcurrency))
					db.SetSetting("last_network_test", result.DetectionTime.Format(time.RFC3339))
					log.Printf("Network detection complete: %s (max concurrency: %d)", result.SpeedLevel, result.MaxConcurrency)
				} else {
					log.Printf("Network detection failed: %s", result.ErrorMessage)
				}
			}()

			// Start background scheduler after a longer delay to allow UI to show first
			go func() {
				time.Sleep(5 * time.Second)
				log.Println("Starting background scheduler...")
				h.StartBackgroundScheduler(bgCtx)
			}()
		},
		OnBeforeClose: func(ctx context.Context) bool {
			if quitRequested.Load() {
				return false
			}

			// Check if window is in fullscreen mode
			isFullscreen := runtime.WindowIsFullscreen(ctx)

			// On macOS, if closing from fullscreen, exit fullscreen first
			// This prevents the black screen issue
			if utils.IsMacOS() && isFullscreen {
				log.Println("Exiting fullscreen before closing on macOS")
				runtime.WindowUnfullscreen(ctx)
				// Don't prevent close, let it proceed normally
				// The minimize-to-tray logic will be skipped since we're in fullscreen
			}

			if shouldCloseToTray() && !isFullscreen {
				storeWindowState(ctx)
				// Fallback start in case tray failed to start on startup
				startTray(ctx)
				if trayManager != nil && trayManager.IsRunning() {
					runtime.WindowHide(ctx)
				} else {
					runtime.WindowMinimise(ctx)
				}
				return true
			}

			return false
		},
	})

	if err != nil {
		log.Printf("Error running Wails: %v", err)
		log.Fatal(err)
	}
	log.Println("Application finished")
}

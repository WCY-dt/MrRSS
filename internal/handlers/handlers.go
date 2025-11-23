// Package handlers contains the HTTP handlers for the application.
// It defines the Handler struct which holds dependencies like the database and fetcher.
package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/feed"
	"MrRSS/internal/opml"
	"MrRSS/internal/translation"
	"MrRSS/internal/version"
)

type Handler struct {
	DB         *database.DB
	Fetcher    *feed.Fetcher
	Translator translation.Translator
}

func NewHandler(db *database.DB, fetcher *feed.Fetcher, translator translation.Translator) *Handler {
	return &Handler{
		DB:         db,
		Fetcher:    fetcher,
		Translator: translator,
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

func (h *Handler) HandleFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(feeds)
}

func (h *Handler) HandleAddFeed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL      string `json:"url"`
		Category string `json:"category"`
		Title    string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.Fetcher.AddSubscription(req.URL, req.Category, req.Title); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleDeleteFeed(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.DB.DeleteFeed(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleUpdateFeed(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID       int64  `json:"id"`
		Title    string `json:"title"`
		URL      string `json:"url"`
		Category string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.DB.UpdateFeed(req.ID, req.Title, req.URL, req.Category); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleArticles(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	feedIDStr := r.URL.Query().Get("feed_id")
	category := r.URL.Query().Get("category")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	var feedID int64
	if feedIDStr != "" {
		feedID, _ = strconv.ParseInt(feedIDStr, 10, 64)
	}

	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := (page - 1) * limit

	articles, err := h.DB.GetArticles(filter, feedID, category, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(articles)
}

func (h *Handler) HandleProgress(w http.ResponseWriter, r *http.Request) {
	progress := h.Fetcher.GetProgress()
	json.NewEncoder(w).Encode(progress)
}

func (h *Handler) HandleMarkRead(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)

	readStr := r.URL.Query().Get("read")
	read := true
	if readStr == "false" || readStr == "0" {
		read = false
	}

	if err := h.DB.MarkArticleRead(id, read); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleToggleFavorite(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.DB.ToggleFavorite(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	go h.Fetcher.FetchAll(context.Background())
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleOPMLImport(w http.ResponseWriter, r *http.Request) {
	log.Printf("HandleOPMLImport: ContentLength: %d", r.ContentLength)
	contentType := r.Header.Get("Content-Type")
	log.Printf("HandleOPMLImport: Content-Type: %s", contentType)

	var file io.Reader

	if strings.Contains(contentType, "multipart/form-data") {
		f, header, err := r.FormFile("file")
		if err != nil {
			log.Printf("Error getting form file: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer f.Close()
		log.Printf("HandleOPMLImport: Received file %s, size: %d", header.Filename, header.Size)

		if header.Size == 0 {
			http.Error(w, "Uploaded file is empty", http.StatusBadRequest)
			return
		}
		file = f
	} else {
		// Handle raw body upload
		file = r.Body
		defer r.Body.Close()
	}

	feeds, err := opml.Parse(file)
	if err != nil {
		log.Printf("Error parsing OPML: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		for _, f := range feeds {
			h.Fetcher.ImportSubscription(f.Title, f.URL, f.Category)
		}
		h.Fetcher.FetchAll(context.Background())
	}()

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleOPMLExport(w http.ResponseWriter, r *http.Request) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := opml.Generate(feeds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename=subscriptions.opml")
	w.Header().Set("Content-Type", "text/xml")
	w.Write(data)
}

func (h *Handler) HandleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		interval, _ := h.DB.GetSetting("update_interval")
		translationEnabled, _ := h.DB.GetSetting("translation_enabled")
		targetLang, _ := h.DB.GetSetting("target_language")
		provider, _ := h.DB.GetSetting("translation_provider")
		apiKey, _ := h.DB.GetSetting("deepl_api_key")
		autoCleanup, _ := h.DB.GetSetting("auto_cleanup_enabled")
		maxCacheSize, _ := h.DB.GetSetting("max_cache_size_mb")
		maxArticleAge, _ := h.DB.GetSetting("max_article_age_days")
		language, _ := h.DB.GetSetting("language")
		theme, _ := h.DB.GetSetting("theme")
		json.NewEncoder(w).Encode(map[string]string{
			"update_interval":       interval,
			"translation_enabled":   translationEnabled,
			"target_language":       targetLang,
			"translation_provider":  provider,
			"deepl_api_key":         apiKey,
			"auto_cleanup_enabled":  autoCleanup,
			"max_cache_size_mb":     maxCacheSize,
			"max_article_age_days":  maxArticleAge,
			"language":              language,
			"theme":                 theme,
		})
	} else if r.Method == http.MethodPost {
		var req struct {
			UpdateInterval      string `json:"update_interval"`
			TranslationEnabled  string `json:"translation_enabled"`
			TargetLanguage      string `json:"target_language"`
			TranslationProvider string `json:"translation_provider"`
			DeepLAPIKey         string `json:"deepl_api_key"`
			AutoCleanupEnabled  string `json:"auto_cleanup_enabled"`
			MaxCacheSizeMB      string `json:"max_cache_size_mb"`
			MaxArticleAgeDays   string `json:"max_article_age_days"`
			Language            string `json:"language"`
			Theme               string `json:"theme"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.UpdateInterval != "" {
			h.DB.SetSetting("update_interval", req.UpdateInterval)
		}
		if req.TranslationEnabled != "" {
			h.DB.SetSetting("translation_enabled", req.TranslationEnabled)
		}
		if req.TargetLanguage != "" {
			h.DB.SetSetting("target_language", req.TargetLanguage)
		}
		if req.TranslationProvider != "" {
			h.DB.SetSetting("translation_provider", req.TranslationProvider)
		}
		// Always update API key as it might be cleared
		h.DB.SetSetting("deepl_api_key", req.DeepLAPIKey)
		
		if req.AutoCleanupEnabled != "" {
			h.DB.SetSetting("auto_cleanup_enabled", req.AutoCleanupEnabled)
		}
		
		if req.MaxCacheSizeMB != "" {
			h.DB.SetSetting("max_cache_size_mb", req.MaxCacheSizeMB)
		}
		
		if req.MaxArticleAgeDays != "" {
			h.DB.SetSetting("max_article_age_days", req.MaxArticleAgeDays)
		}
		
		if req.Language != "" {
			h.DB.SetSetting("language", req.Language)
		}
		
		if req.Theme != "" {
			h.DB.SetSetting("theme", req.Theme)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) HandleCleanupArticles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	count, err := h.DB.CleanupUnimportantArticles()
	if err != nil {
		log.Printf("Error cleaning up articles: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	log.Printf("Cleaned up %d articles", count)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted": count,
	})
}

func (h *Handler) HandleTranslateArticle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req struct {
		ArticleID    int64  `json:"article_id"`
		Title        string `json:"title"`
		TargetLang   string `json:"target_language"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if req.Title == "" || req.TargetLang == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	
	// Translate the title
	translatedTitle, err := h.Translator.Translate(req.Title, req.TargetLang)
	if err != nil {
		log.Printf("Error translating article %d: %v", req.ArticleID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Update the article with the translated title
	if err := h.DB.UpdateArticleTranslation(req.ArticleID, translatedTitle); err != nil {
		log.Printf("Error updating article translation: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	json.NewEncoder(w).Encode(map[string]string{
		"translated_title": translatedTitle,
	})
}

// HandleCheckUpdates checks for the latest version on GitHub
func (h *Handler) HandleCheckUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentVersion := version.Version
	const githubAPI = "https://api.github.com/repos/WCY-dt/MrRSS/releases/latest"

	resp, err := http.Get(githubAPI)
	if err != nil {
		log.Printf("Error checking for updates: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"current_version": currentVersion,
			"error":           "Failed to check for updates",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("GitHub API returned status: %d", resp.StatusCode)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"current_version": currentVersion,
			"error":           "Failed to fetch latest release",
		})
		return
	}

	var release struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		HTMLURL     string `json:"html_url"`
		Body        string `json:"body"`
		PublishedAt string `json:"published_at"`
		Assets      []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		log.Printf("Error decoding release info: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"current_version": currentVersion,
			"error":           "Failed to parse release information",
		})
		return
	}

	// Remove 'v' prefix if present for comparison
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	hasUpdate := compareVersions(latestVersion, currentVersion) > 0

	// Find the appropriate download URL based on platform
	var downloadURL string
	var assetName string
	var assetSize int64
	platform := runtime.GOOS
	arch := runtime.GOARCH

	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		
		// Match platform-specific installer/package with architecture
		// Asset naming convention: MrRSS-{version}-{platform}-{arch}-installer.{ext}
		platformArch := platform + "-" + arch
		
		if platform == "windows" {
			// For Windows, prefer installer.exe, fallback to .zip
			if strings.Contains(name, platformArch) && strings.HasSuffix(name, "-installer.exe") {
				downloadURL = asset.BrowserDownloadURL
				assetName = asset.Name
				assetSize = asset.Size
				break
			}
		} else if platform == "linux" {
			// For Linux, prefer .AppImage, fallback to .tar.gz
			if strings.Contains(name, platformArch) && strings.HasSuffix(name, ".appimage") {
				downloadURL = asset.BrowserDownloadURL
				assetName = asset.Name
				assetSize = asset.Size
				break
			}
		} else if platform == "darwin" {
			// For macOS, use universal build (supports both arm64 and amd64)
			if strings.Contains(name, "darwin-universal") && strings.HasSuffix(name, ".dmg") {
				downloadURL = asset.BrowserDownloadURL
				assetName = asset.Name
				assetSize = asset.Size
				break
			}
		}
	}

	response := map[string]interface{}{
		"current_version": currentVersion,
		"latest_version":  latestVersion,
		"has_update":      hasUpdate,
		"platform":        platform,
		"arch":            arch,
	}

	if downloadURL != "" {
		response["download_url"] = downloadURL
		response["asset_name"] = assetName
		response["asset_size"] = assetSize
	}

	json.NewEncoder(w).Encode(response)
}

// compareVersions compares two semantic versions (e.g., "1.1.0" vs "1.0.0")
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) {
			p1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			p2, _ = strconv.Atoi(parts2[i])
		}

		if p1 > p2 {
			return 1
		} else if p1 < p2 {
			return -1
		}
	}

	return 0
}

// HandleVersion returns the current application version
func (h *Handler) HandleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"version": version.Version,
	})
}

// HandleDownloadUpdate downloads the update file
func (h *Handler) HandleDownloadUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DownloadURL string `json:"download_url"`
		AssetName   string `json:"asset_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate download URL is from the official GitHub repository releases
	const allowedURLPrefix = "https://github.com/WCY-dt/MrRSS/releases/download/"
	if !strings.HasPrefix(req.DownloadURL, allowedURLPrefix) {
		log.Printf("Invalid download URL attempted: %s", req.DownloadURL)
		http.Error(w, "Invalid download URL", http.StatusBadRequest)
		return
	}

	// Validate asset name to prevent path traversal
	if strings.Contains(req.AssetName, "..") || strings.Contains(req.AssetName, "/") || strings.Contains(req.AssetName, "\\") {
		log.Printf("Invalid asset name attempted: %s", req.AssetName)
		http.Error(w, "Invalid asset name", http.StatusBadRequest)
		return
	}

	// Create temp directory for download
	tempDir := os.TempDir()
	filePath := filepath.Join(tempDir, req.AssetName)

	// Download the file
	log.Printf("Downloading update from: %s", req.DownloadURL)
	resp, err := http.Get(req.DownloadURL)
	if err != nil {
		log.Printf("Error downloading update: %v", err)
		http.Error(w, "Failed to download update", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Download failed with status: %d", resp.StatusCode)
		http.Error(w, "Failed to download update", http.StatusInternalServerError)
		return
	}

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		http.Error(w, "Failed to create download file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Write the body to file with progress tracking
	totalSize := resp.ContentLength
	var bytesWritten int64
	
	// Create a buffer for efficient copying
	buffer := make([]byte, 32*1024) // 32KB buffer
	
	for {
		nr, er := resp.Body.Read(buffer)
		if nr > 0 {
			nw, ew := out.Write(buffer[0:nr])
			if nw > 0 {
				bytesWritten += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	if err != nil {
		log.Printf("Error writing file: %v", err)
		http.Error(w, "Failed to write download file", http.StatusInternalServerError)
		return
	}

	log.Printf("Update downloaded successfully to: %s (%.2f MB)", filePath, float64(bytesWritten)/(1024*1024))

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"file_path":   filePath,
		"total_bytes": totalSize,
		"bytes_written": bytesWritten,
	})
}

// HandleInstallUpdate triggers the installation of the downloaded update
func (h *Handler) HandleInstallUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate file path is within temp directory to prevent path traversal
	tempDir := os.TempDir()
	cleanPath := filepath.Clean(req.FilePath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(tempDir)) {
		log.Printf("Invalid file path attempted: %s", req.FilePath)
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Validate file exists and is a regular file
	fileInfo, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		http.Error(w, "Update file not found", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("Error stating file: %v", err)
		http.Error(w, "Error accessing update file", http.StatusInternalServerError)
		return
	}
	if !fileInfo.Mode().IsRegular() {
		log.Printf("File is not a regular file: %s", cleanPath)
		http.Error(w, "Invalid file type", http.StatusBadRequest)
		return
	}

	platform := runtime.GOOS
	log.Printf("Installing update from: %s on platform: %s", cleanPath, platform)

	// Helper function to schedule cleanup of installer file
	scheduleCleanup := func(filePath string, delay time.Duration) {
		go func() {
			time.Sleep(delay)
			if err := os.Remove(filePath); err != nil {
				log.Printf("Failed to remove installer: %v", err)
			} else {
				log.Printf("Successfully removed installer: %s", filePath)
			}
		}()
	}

	// Launch installer based on platform
	var cmd *exec.Cmd
	switch platform {
	case "windows":
		// Launch the installer - validate file extension
		if !strings.HasSuffix(strings.ToLower(cleanPath), ".exe") {
			http.Error(w, "Invalid file type for Windows", http.StatusBadRequest)
			return
		}
		cmd = exec.Command(cleanPath, "/S")
		scheduleCleanup(cleanPath, 3*time.Second)
	case "linux":
		// Make AppImage executable and run it - validate file extension
		if !strings.HasSuffix(strings.ToLower(cleanPath), ".appimage") {
			http.Error(w, "Invalid file type for Linux", http.StatusBadRequest)
			return
		}
		if err := os.Chmod(cleanPath, 0755); err != nil {
			log.Printf("Error making file executable: %v", err)
			http.Error(w, "Failed to prepare installer", http.StatusInternalServerError)
			return
		}
		cmd = exec.Command(cleanPath)
		scheduleCleanup(cleanPath, 3*time.Second)
	case "darwin":
		// Open the DMG file - validate file extension
		if !strings.HasSuffix(strings.ToLower(cleanPath), ".dmg") {
			http.Error(w, "Invalid file type for macOS", http.StatusBadRequest)
			return
		}
		cmd = exec.Command("open", cleanPath)
		scheduleCleanup(cleanPath, 5*time.Second)
	default:
		http.Error(w, "Unsupported platform", http.StatusBadRequest)
		return
	}

	// Start the installer in the background
	if err := cmd.Start(); err != nil {
		log.Printf("Error starting installer: %v", err)
		http.Error(w, "Failed to start installer", http.StatusInternalServerError)
		return
	}

	log.Printf("Installer started successfully")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Installation started. Application will exit shortly.",
	})

	// Schedule graceful shutdown to allow the response to be sent
	// and give time for proper cleanup
	go func() {
		time.Sleep(2 * time.Second)
		log.Println("Initiating graceful shutdown for update installation...")
		// Note: In a production app, this should trigger the Wails shutdown handler
		// which will properly clean up resources. For now, we use os.Exit.
		os.Exit(0)
	}()
}

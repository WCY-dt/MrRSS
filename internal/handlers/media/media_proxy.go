package media

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"MrRSS/internal/cache"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/utils"
)

// validateMediaURL validates that the URL is HTTP/HTTPS and properly formatted
func validateMediaURL(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return errors.New("invalid URL format")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("URL must use HTTP or HTTPS")
	}

	return nil
}

// HandleMediaProxy serves cached media or downloads and caches it
func HandleMediaProxy(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if media cache is enabled
	mediaCacheEnabled, _ := h.DB.GetSetting("media_cache_enabled")
	if mediaCacheEnabled != "true" {
		http.Error(w, "Media cache is disabled", http.StatusForbidden)
		return
	}

	// Get URL from query parameter
	mediaURL := r.URL.Query().Get("url")
	if mediaURL == "" {
		http.Error(w, "Missing url parameter", http.StatusBadRequest)
		return
	}

	// Validate mediaURL (must be HTTP/HTTPS and valid format)
	if err := validateMediaURL(mediaURL); err != nil {
		http.Error(w, "Invalid url parameter: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get optional referer from query parameter
	referer := r.URL.Query().Get("referer")

	// Get media cache directory
	cacheDir, err := utils.GetMediaCacheDir()
	if err != nil {
		log.Printf("Failed to get media cache directory: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Initialize media cache
	mediaCache, err := cache.NewMediaCache(cacheDir)
	if err != nil {
		log.Printf("Failed to initialize media cache: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get media (from cache or download)
	data, contentType, err := mediaCache.Get(mediaURL, referer)
	if err != nil {
		log.Printf("Failed to get media %s: %v", mediaURL, err)
		http.Error(w, "Failed to fetch media", http.StatusInternalServerError)
		return
	}

	// Set appropriate headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year

	// Write response
	if _, err := w.Write(data); err != nil {
		log.Printf("Failed to write media response: %v", err)
	}
}

// HandleMediaCacheCleanup performs manual cleanup of media cache
func HandleMediaCacheCleanup(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get media cache directory
	cacheDir, err := utils.GetMediaCacheDir()
	if err != nil {
		log.Printf("Failed to get media cache directory: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Initialize media cache
	mediaCache, err := cache.NewMediaCache(cacheDir)
	if err != nil {
		log.Printf("Failed to initialize media cache: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if this is a manual cleanup (clean all) or automatic cleanup (respect settings)
	cleanAll := r.URL.Query().Get("all") == "true"

	var maxAgeDays int
	var maxSizeMB int

	if cleanAll {
		// Manual cleanup: remove all files
		maxAgeDays = 0
		maxSizeMB = 0 // Will skip size-based cleanup
	} else {
		// Automatic cleanup: use settings
		maxAgeDaysStr, _ := h.DB.GetSetting("media_cache_max_age_days")
		maxSizeMBStr, _ := h.DB.GetSetting("media_cache_max_size_mb")

		maxAgeDays, err = strconv.Atoi(maxAgeDaysStr)
		if err != nil || maxAgeDays < 0 {
			maxAgeDays = 7 // Default
		}

		maxSizeMB, err = strconv.Atoi(maxSizeMBStr)
		if err != nil || maxSizeMB <= 0 {
			maxSizeMB = 100 // Default
		}
	}

	// Cleanup by age
	ageCount, err := mediaCache.CleanupOldFiles(maxAgeDays)
	if err != nil {
		log.Printf("Failed to cleanup old media files: %v", err)
	}

	// Cleanup by size (only for automatic cleanup)
	sizeCount := 0
	if !cleanAll {
		sizeCount, err = mediaCache.CleanupBySize(maxSizeMB)
		if err != nil {
			log.Printf("Failed to cleanup media files by size: %v", err)
		}
	}

	totalCleaned := ageCount + sizeCount
	log.Printf("Media cache cleanup: removed %d files (clean_all: %v)", totalCleaned, cleanAll)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success":       true,
		"files_cleaned": totalCleaned,
	}
	json.NewEncoder(w).Encode(response)
}

// HandleWebpageProxy proxies webpage content to bypass CSP restrictions in iframes
func HandleWebpageProxy(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get URL from query parameter
	webpageURL := r.URL.Query().Get("url")
	if webpageURL == "" {
		http.Error(w, "Missing url parameter", http.StatusBadRequest)
		return
	}

	// Validate webpageURL (must be HTTP/HTTPS and valid format)
	if err := validateMediaURL(webpageURL); err != nil {
		http.Error(w, "Invalid url parameter: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Create HTTP client with proxy settings if enabled
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Check if proxy is enabled and configure client
	proxyEnabled, _ := h.DB.GetSetting("proxy_enabled")
	if proxyEnabled == "true" {
		proxyType, _ := h.DB.GetSetting("proxy_type")
		proxyHost, _ := h.DB.GetSetting("proxy_host")
		proxyPort, _ := h.DB.GetSetting("proxy_port")
		proxyUsername, _ := h.DB.GetSetting("proxy_username")
		proxyPassword, _ := h.DB.GetSetting("proxy_password")

		proxyURLStr := utils.BuildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)
		if proxyURLStr != "" {
			proxyURL, err := url.Parse(proxyURLStr)
			if err != nil {
				log.Printf("Failed to parse proxy URL: %v", err)
			} else {
				transport := &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				}
				client.Transport = transport
			}
		}
	}

	// Create request to the target URL
	req, err := http.NewRequest("GET", webpageURL, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}

	// Set User-Agent to mimic a regular browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// Forward some headers from the original request
	if referer := r.Header.Get("Referer"); referer != "" {
		req.Header.Set("Referer", referer)
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch webpage %s: %v", webpageURL, err)
		http.Error(w, "Failed to fetch webpage", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("Webpage returned status %d: %s", resp.StatusCode, webpageURL)
		http.Error(w, "Webpage returned error", resp.StatusCode)
		return
	}

	// Get content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/html; charset=utf-8"
	}

	// Read the entire response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		http.Error(w, "Failed to read webpage content", http.StatusInternalServerError)
		return
	}

	// If this is HTML content, add base tag to fix relative links
	if strings.Contains(strings.ToLower(contentType), "text/html") {
		// Parse the URL to get the base URL
		parsedURL, err := url.Parse(webpageURL)
		if err == nil {
			baseURL := parsedURL.Scheme + "://" + parsedURL.Host
			baseTag := fmt.Sprintf("<base href=\"%s\">", baseURL)

			// Convert body to string for manipulation
			bodyStr := string(bodyBytes)

			// Find the <head> tag and insert base tag after it
			headIndex := strings.Index(strings.ToLower(bodyStr), "<head>")
			if headIndex == -1 {
				// If no <head>, look for <html>
				htmlIndex := strings.Index(strings.ToLower(bodyStr), "<html>")
				if htmlIndex != -1 {
					// Insert after <html>
					htmlEndIndex := htmlIndex + strings.Index(bodyStr[htmlIndex:], ">") + 1
					bodyStr = bodyStr[:htmlEndIndex] + "<head>" + baseTag + "</head>" + bodyStr[htmlEndIndex:]
				}
			} else {
				// Insert after <head>
				headEndIndex := headIndex + strings.Index(bodyStr[headIndex:], ">") + 1
				bodyStr = bodyStr[:headEndIndex] + baseTag + bodyStr[headEndIndex:]
			}

			// Convert back to bytes
			bodyBytes = []byte(bodyStr)
		}
	}

	// Set response headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Frame-Options", "SAMEORIGIN") // Allow framing from same origin
	w.Header().Set("Content-Length", strconv.Itoa(len(bodyBytes)))

	// Write modified response body
	_, err = w.Write(bodyBytes)
	if err != nil {
		log.Printf("Failed to write response body: %v", err)
	}
}

// HandleMediaCacheInfo returns information about the media cache
func HandleMediaCacheInfo(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get media cache directory
	cacheDir, err := utils.GetMediaCacheDir()
	if err != nil {
		log.Printf("Failed to get media cache directory: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Initialize media cache
	mediaCache, err := cache.NewMediaCache(cacheDir)
	if err != nil {
		log.Printf("Failed to initialize media cache: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get cache size
	cacheSize, err := mediaCache.GetCacheSize()
	if err != nil {
		log.Printf("Failed to get cache size: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to MB
	cacheSizeMB := float64(cacheSize) / (1024 * 1024)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"cache_size_mb": cacheSizeMB,
	}
	json.NewEncoder(w).Encode(response)
}

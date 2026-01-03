package media

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
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

// proxyImagesInHTML replaces image URLs in HTML with proxied versions
func proxyImagesInHTML(htmlContent, referer string) string {
	if htmlContent == "" || referer == "" {
		return htmlContent
	}

	// Parse the referer URL once for resolving relative URLs
	baseURL, err := url.Parse(referer)
	if err != nil {
		log.Printf("Failed to parse referer URL: %v", err)
		return htmlContent
	}

	// Use regex to find and replace img src attributes
	// This handles various formats: src="url", src='url', src=url (unquoted)
	re := regexp.MustCompile(`<img[^>]*src\s*=\s*(?:['"]\s*)?([^'"\s>]+)(?:\s*['"])?[^>]*>`)
	htmlContent = re.ReplaceAllStringFunc(htmlContent, func(match string) string {
		// Extract the src URL from the match
		re := regexp.MustCompile(`src\s*=\s*(?:['"]\s*)?([^'"\s>]+)(?:\s*['"])?`)
		srcMatch := re.FindStringSubmatch(match)
		if len(srcMatch) < 2 {
			return match // No valid src found, return unchanged
		}

		srcURL := srcMatch[1]

		// Skip data URLs, blob URLs, and already proxied URLs
		if strings.HasPrefix(srcURL, "data:") ||
			strings.HasPrefix(srcURL, "blob:") ||
			strings.Contains(srcURL, "/api/media/proxy") {
			return match
		}

		// CRITICAL FIX: Decode HTML entities before processing the URL
		// HTML attributes contain &amp; which should be decoded to & before URL encoding
		// For example: ?key=val&amp;other=val becomes ?key=val&other=val
		srcURL = html.UnescapeString(srcURL)

		// Resolve relative URLs against the referer
		// Handles: images/photo.jpg, ./img.png, ../assets/image.gif, /static/img.png
		if !strings.HasPrefix(srcURL, "http://") && !strings.HasPrefix(srcURL, "https://") {
			parsedURL, err := url.Parse(srcURL)
			if err != nil {
				log.Printf("Failed to parse image URL %s: %v", srcURL, err)
				return match
			}
			srcURL = baseURL.ResolveReference(parsedURL).String()
		}

		// Build proxied URL
		proxyURL := fmt.Sprintf("/api/media/proxy?url=%s&referer=%s",
			url.QueryEscape(srcURL),
			url.QueryEscape(referer))

		// Replace the src attribute
		return strings.Replace(match, srcMatch[0], fmt.Sprintf(`src="%s"`, proxyURL), 1)
	})

	return htmlContent
}

// HandleMediaProxy serves cached media or downloads and caches it
// HandleMediaProxy proxies media files with optional caching
// @Summary      Proxy media file
// @Description  Proxy and cache media files (images, videos, audio) from external URLs
// @Tags         media
// @Accept       json
// @Produce      application/octet-stream
// @Param        url      query     string  true  "Media URL to proxy"
// @Param        referer  query     string  false  "Referer URL for hotlink protection"
// @Success      200  {file}  file  "Media file"
// @Failure      400  {object}  map[string]string  "Bad request (missing or invalid URL)"
// @Failure      403  {object}  map[string]string  "Media proxy is disabled"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /media/proxy [get]
func HandleMediaProxy(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	// Check if media cache is enabled
	mediaCacheEnabled, _ := h.DB.GetSetting("media_cache_enabled")
	mediaProxyFallback, _ := h.DB.GetSetting("media_proxy_fallback")

	// If neither cache nor fallback is enabled, return error
	if mediaCacheEnabled != "true" && mediaProxyFallback != "true" {
		http.Error(w, "Media proxy is disabled", http.StatusForbidden)
		return
	}

	// Get optional referer from query parameter
	referer := r.URL.Query().Get("referer")

	// Try cache first if enabled
	if mediaCacheEnabled == "true" {
		// Get media cache directory
		cacheDir, err := utils.GetMediaCacheDir()
		if err != nil {
			log.Printf("Failed to get media cache directory: %v", err)
			// Continue to fallback if enabled
		} else {
			// Initialize media cache
			mediaCache, err := cache.NewMediaCache(cacheDir)
			if err != nil {
				log.Printf("Failed to initialize media cache: %v", err)
				// Continue to fallback if enabled
			} else {
				// Get media (from cache or download)
				data, contentType, err := mediaCache.Get(mediaURL, referer)
				if err == nil {
					// Success! Serve from cache
					w.Header().Set("Content-Type", contentType)
					w.Header().Set("Content-Length", strconv.Itoa(len(data)))
					w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
					w.Header().Set("X-Media-Source", "cache")
					w.Write(data)
					return
				}
				log.Printf("Cache failed for %s: %v, trying fallback", mediaURL, err)
			}
		}
	}

	// Fallback: Direct proxy if enabled
	if mediaProxyFallback == "true" {
		err := proxyMediaDirectly(mediaURL, referer, w)
		if err == nil {
			return // Success
		}
		log.Printf("Direct proxy failed for %s: %v", mediaURL, err)
	}

	// All methods failed
	http.Error(w, "Failed to fetch media", http.StatusInternalServerError)
}

// HandleMediaCacheCleanup performs manual cleanup of media cache
// @Summary      Cleanup media cache
// @Description  Clean up the media cache by age and size
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        all  query     bool  false  "Clean all files (ignores age/size settings)"  default(false)
// @Success      200  {object}  map[string]interface{}  "Cleanup result (success, files_cleaned)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /media/cache/cleanup [post]
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
// @Summary      Proxy webpage content
// @Description  Proxy webpage HTML content to bypass CSP restrictions in iframes
// @Tags         media
// @Accept       json
// @Produce      html
// @Param        url  query     string  true  "Webpage URL to proxy"
// @Success      200  {string}  string  "Webpage HTML content"
// @Failure      400  {object}  map[string]string  "Bad request (missing or invalid URL)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /media/proxy-webpage [get]
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

			// CRITICAL FIX: Proxy images in the HTML content
			// Check if media cache is enabled
			mediaCacheEnabled, _ := h.DB.GetSetting("media_cache_enabled")
			if mediaCacheEnabled == "true" {
				// Proxy all images in the HTML using the webpage URL as referer
				bodyStr = proxyImagesInHTML(bodyStr, webpageURL)
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

// proxyMediaDirectly proxies media directly without caching
func proxyMediaDirectly(mediaURL, referer string, w http.ResponseWriter) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", mediaURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to bypass anti-hotlinking
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	if referer != "" {
		req.Header.Set("Referer", referer)
	}

	// Add additional headers
	req.Header.Set("Accept", "image/webp,image/apng,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch media: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = getContentTypeFromPath(mediaURL)
	}

	// Set response headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.Header().Set("X-Media-Source", "direct-proxy")

	// Stream the response directly to avoid loading large files into memory
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to stream response: %w", err)
	}

	return nil
}

// HandleMediaCacheInfo returns information about the media cache
// HandleMediaCacheInfo returns information about the media cache
// @Summary      Get media cache info
// @Description  Get media cache statistics (size in MB)
// @Tags         media
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Cache info (cache_size_mb)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /media/cache/info [get]
func HandleMediaCacheInfo(h *core.Handler, w http.ResponseWriter, r *http.Request) {

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

// getContentTypeFromPath determines content type from file extension
func getContentTypeFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".ogg":
		return "video/ogg"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".flac":
		return "audio/flac"
	default:
		return "application/octet-stream"
	}
}

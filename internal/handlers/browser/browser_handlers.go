//go:build !server

// Package browser provides HTTP handlers for browser-related operations using Wails v3 Browser API.
package browser

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/wailsapp/wails/v3/pkg/application"

	handlers "MrRSS/internal/handlers/core"
	"MrRSS/internal/utils"
)

// HandleOpenURL opens a URL in the user's default web browser using Wails v3 Browser API.
// @Summary      Open URL in browser
// @Description  Open a URL in the user's default web browser
// @Tags         browser
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Open URL request (url)"
// @Success      200  {object}  map[string]string  "Success status (status) or redirect URL (redirect)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid URL)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /browser/open [post]
func HandleOpenURL(h *handlers.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate URL
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Parse and validate URL scheme
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		log.Printf("Invalid URL format: %v", err)
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	// Only allow http and https schemes for security
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		log.Printf("Invalid URL scheme: %s", parsedURL.Scheme)
		http.Error(w, "Only HTTP and HTTPS URLs are allowed", http.StatusBadRequest)
		return
	}

	// Check if app instance is available
	if h.App == nil {
		// specific check for server mode to redirect to client side
		if utils.IsServerMode() {
			log.Printf("Server mode detected, instructing client to open URL: %s", req.URL)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"redirect": req.URL})
			return
		}

		log.Printf("App instance not available for browser integration")
		http.Error(w, "Browser integration not available", http.StatusInternalServerError)
		return
	}

	// Open URL using Wails v3 Browser API
	// Note: app.Browser is a field of type *application.BrowserManager
	wailsApp, ok := h.App.(*application.App)
	if !ok {
		log.Printf("Browser integration not available - invalid app type")
		http.Error(w, "Browser integration not available", http.StatusInternalServerError)
		return
	}

	err = wailsApp.Browser.OpenURL(req.URL)
	if err != nil {
		log.Printf("Failed to open URL in browser: %v", err)
		http.Error(w, "Failed to open URL in browser", http.StatusInternalServerError)
		return
	}

	log.Printf("Opened URL in browser: %s", req.URL)

	// Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

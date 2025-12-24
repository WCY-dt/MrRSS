package miniflux

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/miniflux"
)

// HandleSync performs synchronization with Miniflux server
func HandleSync(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get Miniflux settings
	serverURL, _ := h.DB.GetSetting("miniflux_server_url")
	apiKey, _ := h.DB.GetSetting("miniflux_api_key")

	if serverURL == "" || apiKey == "" {
		http.Error(w, "Miniflux settings incomplete", http.StatusBadRequest)
		return
	}

	// Create sync service
	syncService := miniflux.NewSyncService(serverURL, apiKey, h.DB)

	// Perform sync with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	go func() {
		if err := syncService.Sync(ctx); err != nil {
			log.Printf("Miniflux sync failed: %v", err)
		} else {
			log.Printf("Miniflux sync completed successfully")
			// Trigger a refresh of all feeds to update the article list
			go h.Fetcher.FetchAll(context.Background())
		}
	}()

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Miniflux sync completed successfully",
	})
}

// HandleTestConnection tests the connection to Miniflux server
func HandleTestConnection(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		ServerURL string `json:"server_url"`
		APIKey    string `json:"api_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.ServerURL == "" || req.APIKey == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Miniflux settings incomplete",
		})
		return
	}

	// Test connection
	client := miniflux.NewClient(req.ServerURL, req.APIKey)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.TestConnection(ctx)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// Get feeds count
	feeds, err := client.GetFeeds(ctx)
	feedCount := 0
	if err == nil {
		feedCount = len(feeds)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "Connection successful",
		"feedCount": feedCount,
	})
}

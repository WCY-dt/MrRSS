package rsshub

import (
	"encoding/json"
	"net/http"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/rsshub"
)

// HandleAddFeed adds a new RSSHub feed subscription
func HandleAddFeed(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Route    string `json:"route"`
		Category string `json:"category"`
		Title    string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate route
	if req.Route == "" {
		http.Error(w, "Route is required", http.StatusBadRequest)
		return
	}

	// Add RSSHub subscription using specialized handler
	feedID, err := h.Fetcher.AddRSSHubSubscription(req.Route, req.Category, req.Title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"feed_id": feedID,
	})
}

// HandleTestConnection tests the RSSHub endpoint and API key
func HandleTestConnection(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Endpoint string `json:"endpoint"`
		APIKey   string `json:"api_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Test with a simple, common route
	client := rsshub.NewClient(req.Endpoint, req.APIKey)
	err := client.ValidateRoute("nytimes")

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Connection successful",
	})
}

// HandleValidateRoute validates a specific RSSHub route
func HandleValidateRoute(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Route string `json:"route"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Route == "" {
		http.Error(w, "Route is required", http.StatusBadRequest)
		return
	}

	// Get RSSHub settings
	endpoint, _ := h.DB.GetSetting("rsshub_endpoint")
	if endpoint == "" {
		endpoint = "https://rsshub.app"
	}
	apiKey, _ := h.DB.GetEncryptedSetting("rsshub_api_key")

	client := rsshub.NewClient(endpoint, apiKey)
	err := client.ValidateRoute(req.Route)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": false,
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":   true,
		"message": "Route is valid",
	})
}

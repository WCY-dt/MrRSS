package rsshub

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"MrRSS/internal/handlers/core"
)

// TestConfigRequest represents the request body for testing RSSHub configuration
type TestConfigRequest struct {
	InstanceURL string `json:"instance_url"`
	APIKey      string `json:"api_key"`
}

// TestConfigResponse represents the response for RSSHub configuration test
type TestConfigResponse struct {
	Success      bool   `json:"success"`
	StatusCode   int    `json:"status_code,omitempty"`
	ResponseTime int64  `json:"response_time_ms,omitempty"`
	Error        string `json:"error,omitempty"`
	Message      string `json:"message,omitempty"`
}

// HandleTestConfig handles POST /api/rsshub/config/test to test RSSHub configuration
func HandleTestConfig(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TestConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response := TestConfigResponse{
		Success: false,
	}

	if req.InstanceURL == "" {
		response.Error = "Instance URL is required"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	urlStr := strings.TrimSpace(req.InstanceURL)

	if !strings.Contains(urlStr, "://") {

		if strings.Contains(urlStr, ":") && !strings.HasPrefix(urlStr, "/") {
			urlStr = "http://" + urlStr
		} else {
			urlStr = "https://" + urlStr
		}
	}

	instanceURL, err := url.Parse(urlStr)
	if err != nil {
		response.Error = fmt.Sprintf("Invalid instance URL: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Normalize URL (remove trailing slash, ensure scheme)
	if instanceURL.Scheme == "" {
		instanceURL.Scheme = "https"
	}
	instanceURL.Path = strings.TrimSuffix(instanceURL.Path, "/")
	baseURL := instanceURL.String()

	// test a normal endpoints to ensure authorize is needed
	testURL := baseURL + "/nytimes"

	// Test connection with both authentication methods
	startTime := time.Now()

	_, statusCode, err := testRSSHubConnection(testURL, req.APIKey, "query")
	if err != nil {

		_, statusCode, err = testRSSHubConnection(testURL, req.APIKey, "header")
	}

	responseTime := time.Since(startTime).Milliseconds()
	response.ResponseTime = responseTime
	response.StatusCode = statusCode

	if err != nil {
		response.Error = err.Error()
		response.Success = false
	} else if statusCode == http.StatusOK {
		response.Success = true
		response.Message = "Connection test successful"
	} else {
		response.Error = fmt.Sprintf("Unexpected status code: %d", statusCode)
		response.Success = false
	}

	w.Header().Set("Content-Type", "application/json")
	if !response.Success {
		w.WriteHeader(http.StatusOK) // Still return 200, but success=false in body
	}
	json.NewEncoder(w).Encode(response)
}

// testRSSHubConnection tests the RSSHub connection with the specified authentication method
func testRSSHubConnection(baseURL, apiKey, authMethod string) (success bool, statusCode int, err error) {

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	reqURL := baseURL
	if apiKey != "" && authMethod == "query" {
		u, _ := url.Parse(baseURL)
		q := u.Query()
		q.Set("key", apiKey)
		u.RawQuery = q.Encode()
		reqURL = u.String()
	}

	// Create request
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return false, 0, fmt.Errorf("failed to create request: %w", err)
	}

	if apiKey != "" && authMethod == "header" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	// Set user agent, and rsshub do not need real ua, so I just call it MrRSS hhhh
	// if in the future need, maybe we can offer a ua.txt , and get it randomly

	req.Header.Set("User-Agent", "MrRSS/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return false, 0, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes := make([]byte, 1024)
	n, _ := io.ReadFull(resp.Body, bodyBytes)
	bodyStr := string(bodyBytes[:n])

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == 503 {
		if apiKey == "" {
			return false, resp.StatusCode, fmt.Errorf("authentication required - this instance requires an API key")
		}
		return false, resp.StatusCode, fmt.Errorf("authentication failed - check API key")
	}

	if resp.StatusCode != http.StatusOK {

		if strings.Contains(bodyStr, "error") || strings.Contains(bodyStr, "Error") {
			return false, resp.StatusCode, fmt.Errorf("RSSHub returned error (status %d)", resp.StatusCode)
		}
		return false, resp.StatusCode, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return true, resp.StatusCode, nil
}

package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"MrRSS/internal/config"
	"MrRSS/internal/handlers/core"
)

// TestResult represents the result of AI configuration test
type TestResult struct {
	ConfigValid       bool   `json:"config_valid"`
	ConnectionSuccess bool   `json:"connection_success"`
	ModelAvailable    bool   `json:"model_available"`
	ResponseTimeMs    int64  `json:"response_time_ms"`
	TestTime          string `json:"test_time"`
	ErrorMessage      string `json:"error_message,omitempty"`
}

// HandleTestAIConfig handles POST /api/ai/test to test AI configuration
// @Summary      Test AI configuration
// @Description  Test AI service configuration (endpoint, API key, model availability)
// @Tags         ai
// @Accept       json
// @Produce      json
// @Success      200  {object}  ai.TestResult  "Test result (config_valid, connection_success, model_available, response_time_ms, test_time, error_message)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /ai/test [post]
func HandleTestAIConfig(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	result := TestResult{
		TestTime: time.Now().Format(time.RFC3339),
	}

	// Get AI settings
	apiKey, _ := h.DB.GetEncryptedSetting("ai_api_key")
	endpoint, _ := h.DB.GetSetting("ai_endpoint")
	model, _ := h.DB.GetSetting("ai_model")

	// Use defaults if not set
	defaults := config.Get()
	if endpoint == "" {
		endpoint = defaults.AIEndpoint
	}
	if model == "" {
		model = defaults.AIModel
	}

	// Validate configuration
	result.ConfigValid = true
	validationErrors := []string{}

	if endpoint == "" {
		validationErrors = append(validationErrors, "endpoint is required")
		result.ConfigValid = false
	}

	if model == "" {
		validationErrors = append(validationErrors, "model is required")
		result.ConfigValid = false
	}

	if !result.ConfigValid {
		result.ErrorMessage = "Configuration incomplete: " + strings.Join(validationErrors, ", ")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// Validate endpoint URL format
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		result.ConfigValid = false
		result.ErrorMessage = "Invalid endpoint URL: " + err.Error()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// Both HTTP and HTTPS are allowed
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		result.ConfigValid = false
		result.ErrorMessage = "API endpoint must use HTTP or HTTPS"
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	// Test connection with a simple request
	startTime := time.Now()

	// Try OpenAI format first
	connectionSuccess, modelAvailable, err := testOpenAIConnection(endpoint, apiKey, model, h)
	if err != nil {
		// If OpenAI format fails, try Ollama format
		connectionSuccess, modelAvailable, err = testOllamaConnection(endpoint, apiKey, model, h)
		if err != nil {
			result.ConnectionSuccess = false
			result.ModelAvailable = false
			result.ErrorMessage = fmt.Sprintf("Connection failed: %v", err)
		} else {
			result.ConnectionSuccess = connectionSuccess
			result.ModelAvailable = modelAvailable
		}
	} else {
		result.ConnectionSuccess = connectionSuccess
		result.ModelAvailable = modelAvailable
	}

	result.ResponseTimeMs = time.Since(startTime).Milliseconds()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleGetAITestInfo handles GET /api/ai/test/info to get last test result
// @Summary      Get AI test info
// @Description  Get the last AI configuration test result (returns empty/default result as tests are not stored persistently)
// @Tags         ai
// @Accept       json
// @Produce      json
// @Success      200  {object}  ai.TestResult  "Empty test result (config_valid, connection_success, model_available, response_time_ms, test_time)"
// @Router       /ai/test/info [get]
func HandleGetAITestInfo(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return default/empty result - tests are not stored persistently
	// The frontend will trigger a new test if needed
	result := TestResult{
		ConfigValid:       false,
		ConnectionSuccess: false,
		ModelAvailable:    false,
		ResponseTimeMs:    0,
		TestTime:          "",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// testOpenAIConnection tests the AI connection using OpenAI format
func testOpenAIConnection(endpoint, apiKey, model string, h *core.Handler) (connectionSuccess, modelAvailable bool, err error) {
	requestBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": "test"},
		},
		"max_tokens": 5,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return false, false, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := sendTestRequest(endpoint, apiKey, jsonBody, h)
	if err != nil {
		return false, false, err
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return true, false, fmt.Errorf("authentication failed - check API key")
	}

	if resp.StatusCode == http.StatusNotFound {
		return true, false, fmt.Errorf("model '%s' not found", model)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return true, false, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return true, false, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for API error in response body
	if result.Error != nil {
		if strings.Contains(result.Error.Type, "invalid_request_error") {
			return true, false, fmt.Errorf("API error: %s", result.Error.Message)
		}
		return true, false, fmt.Errorf("API error: %s", result.Error.Message)
	}

	// Success if we got choices back
	return true, len(result.Choices) > 0, nil
}

// testOllamaConnection tests the AI connection using Ollama format
func testOllamaConnection(endpoint, apiKey, model string, h *core.Handler) (connectionSuccess, modelAvailable bool, err error) {
	requestBody := map[string]interface{}{
		"model":  model,
		"prompt": "test",
		"stream": false,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return false, false, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := sendTestRequest(endpoint, apiKey, jsonBody, h)
	if err != nil {
		return false, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return true, false, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
		Error    string `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return true, false, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != "" {
		return true, false, fmt.Errorf("Ollama error: %s", result.Error)
	}

	// Success if response is done
	return true, result.Done, nil
}

// sendTestRequest sends the HTTP test request with proper headers and proxy support
func sendTestRequest(endpoint, apiKey string, jsonBody []byte, h *core.Handler) (*http.Response, error) {
	// Create HTTP client with proxy support if configured
	client, err := createHTTPClientWithProxy(h)
	if err != nil {
		log.Printf("Failed to create HTTP client with proxy: %v", err)
		client = &http.Client{Timeout: 30 * time.Second}
	} else {
		// Set timeout for test request
		client.Timeout = 30 * time.Second
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	return client.Do(req)
}

// createHTTPClientWithProxy creates an HTTP client with global proxy settings if enabled
func createHTTPClientWithProxy(h *core.Handler) (*http.Client, error) {
	// Check if global proxy is enabled
	proxyEnabled, _ := h.DB.GetSetting("proxy_enabled")
	if proxyEnabled != "true" {
		return &http.Client{}, nil
	}

	// Build proxy URL from global settings
	proxyType, _ := h.DB.GetSetting("proxy_type")
	proxyHost, _ := h.DB.GetSetting("proxy_host")
	proxyPort, _ := h.DB.GetSetting("proxy_port")
	proxyUsername, _ := h.DB.GetEncryptedSetting("proxy_username")
	proxyPassword, _ := h.DB.GetEncryptedSetting("proxy_password")

	// Build proxy URL
	proxyURL := buildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)

	if proxyURL == "" {
		return &http.Client{}, nil
	}

	// Parse proxy URL
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(u),
		},
	}, nil
}

// buildProxyURL builds a proxy URL from components
func buildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword string) string {
	if proxyHost == "" || proxyPort == "" {
		return ""
	}

	var urlBuilder strings.Builder
	urlBuilder.WriteString(strings.ToLower(proxyType))
	urlBuilder.WriteString("://")

	if proxyUsername != "" && proxyPassword != "" {
		urlBuilder.WriteString(url.QueryEscape(proxyUsername))
		urlBuilder.WriteString(":")
		urlBuilder.WriteString(url.QueryEscape(proxyPassword))
		urlBuilder.WriteString("@")
	}

	urlBuilder.WriteString(proxyHost)
	urlBuilder.WriteString(":")
	urlBuilder.WriteString(proxyPort)

	return urlBuilder.String()
}

// isLocalEndpoint checks if a host is a local endpoint
func isLocalEndpoint(host string) bool {
	// Remove port if present
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		if !strings.Contains(host[idx:], "]") {
			host = host[:idx]
		}
	}
	// Remove brackets from IPv6 addresses
	host = strings.Trim(host, "[]")

	return host == "localhost" ||
		host == "127.0.0.1" ||
		host == "::1" ||
		strings.HasPrefix(host, "127.") ||
		host == "0.0.0.0"
}

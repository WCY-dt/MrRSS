package rsshub

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MrRSS/internal/database"
	corepkg "MrRSS/internal/handlers/core"
)

func setupHandler(t *testing.T) *corepkg.Handler {
	t.Helper()
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	return corepkg.NewHandler(db, nil, nil)
}

func TestHandleTestConfig_MethodNotAllowed(t *testing.T) {
	h := setupHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/rsshub/config/test", nil)
	rr := httptest.NewRecorder()

	HandleTestConfig(h, rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleTestConfig_InvalidRequestBody(t *testing.T) {
	h := setupHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/api/rsshub/config/test", strings.NewReader("invalid json"))
	rr := httptest.NewRecorder()

	HandleTestConfig(h, rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleTestConfig_MissingInstanceURL(t *testing.T) {
	h := setupHandler(t)
	reqBody := TestConfigRequest{
		InstanceURL: "",
		APIKey:      "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/rsshub/config/test", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	HandleTestConfig(h, rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, rr.Code)
	}

	var resp TestConfigResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if resp.Success {
		t.Fatal("expected success=false")
	}
	if !strings.Contains(resp.Error, "Instance URL is required") {
		t.Fatalf("expected error about missing URL, got: %s", resp.Error)
	}
}

func TestHandleTestConfig_InvalidURL(t *testing.T) {
	h := setupHandler(t)
	reqBody := TestConfigRequest{
		InstanceURL: "://invalid-url",
		APIKey:      "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/rsshub/config/test", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	HandleTestConfig(h, rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, rr.Code)
	}

	var resp TestConfigResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if resp.Success {
		t.Fatal("expected success=false")
	}
	if !strings.Contains(resp.Error, "Invalid instance URL") {
		t.Fatalf("expected error about invalid URL, got: %s", resp.Error)
	}
}

func TestHandleTestConfig_Unauthorized(t *testing.T) {
	// Create a mock RSSHub server that returns 401 for /nytimes (when API key is provided)
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If testing /nytimes endpoint (when API key is provided), return 401
		if r.URL.Path == "/nytimes" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}
	}))
	defer mockServer.Close()

	h := setupHandler(t)
	reqBody := TestConfigRequest{
		InstanceURL: mockServer.URL,
		APIKey:      "invalid-key",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/rsshub/config/test", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	HandleTestConfig(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d", http.StatusOK, rr.Code)
	}

	var resp TestConfigResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if resp.Success {
		t.Fatal("expected success=false")
	}
	if !strings.Contains(resp.Error, "authentication failed") {
		t.Fatalf("expected error about authentication, got: %s", resp.Error)
	}
}

func TestHandleTestConfig_RequiresAuthButNoKey(t *testing.T) {
	// Create a mock RSSHub server that requires authentication for /nytimes endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// /nytimes endpoint requires authentication
		if r.URL.Path == "/nytimes" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}
	}))
	defer mockServer.Close()

	h := setupHandler(t)
	reqBody := TestConfigRequest{
		InstanceURL: mockServer.URL,
		APIKey:      "", // No API key provided
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/rsshub/config/test", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	HandleTestConfig(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d", http.StatusOK, rr.Code)
	}

	var resp TestConfigResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if resp.Success {
		t.Fatal("expected success=false when instance requires auth but no key provided")
	}
	if !strings.Contains(resp.Error, "authentication required") {
		t.Fatalf("expected error about authentication required, got: %s", resp.Error)
	}
}

func TestHandleTestConfig_URLNormalization(t *testing.T) {
	// Create a mock RSSHub server that handles /nytimes endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle /nytimes endpoint (test endpoint)
		if r.URL.Path == "/nytimes" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"></rss>`))
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Not Found"))
		}
	}))
	defer mockServer.Close()

	h := setupHandler(t)

	tests := []struct {
		name          string
		inputURL      string
		shouldSucceed bool
	}{
		{
			name:          "URL without scheme",
			inputURL:      strings.TrimPrefix(mockServer.URL, "http://"),
			shouldSucceed: true,
		},
		{
			name:          "URL with trailing slash",
			inputURL:      mockServer.URL + "/",
			shouldSucceed: true,
		},
		{
			name:          "Normal URL",
			inputURL:      mockServer.URL,
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := TestConfigRequest{
				InstanceURL: tt.inputURL,
				APIKey:      "",
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/rsshub/config/test", bytes.NewReader(body))
			rr := httptest.NewRecorder()

			HandleTestConfig(h, rr, req)

			var resp TestConfigResponse
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("decode failed: %v", err)
			}

			if tt.shouldSucceed && !resp.Success {
				t.Fatalf("expected success but got error: %s", resp.Error)
			}
		})
	}
}

func TestTestRSSHubConnection_QueryParameterAuth(t *testing.T) {
	// Create a mock server that checks for query parameter
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.URL.Query().Get("key")
		if apiKey == "test-key" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer mockServer.Close()

	success, statusCode, err := testRSSHubConnection(mockServer.URL, "test-key", "query")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !success {
		t.Fatal("expected success=true")
	}
	if statusCode != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", statusCode)
	}
}

func TestTestRSSHubConnection_HeaderAuth(t *testing.T) {
	// Create a mock server that checks for Authorization header
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "Bearer test-key" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer mockServer.Close()

	success, statusCode, err := testRSSHubConnection(mockServer.URL, "test-key", "header")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !success {
		t.Fatal("expected success=true")
	}
	if statusCode != http.StatusOK {
		t.Fatalf("expected status code 200, got %d", statusCode)
	}
}

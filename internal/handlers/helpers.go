package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

// HTTP response helpers
func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(w, statusCode, errorResponse{Error: message})
}

func respondWithSuccess(w http.ResponseWriter, message string) {
	respondWithJSON(w, http.StatusOK, successResponse{Success: true, Message: message})
}

func respondOK(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

// Validation helpers
func parseIntQueryParam(r *http.Request, key string, defaultValue int) int {
	valStr := r.URL.Query().Get(key)
	if val, err := strconv.Atoi(valStr); err == nil && val > 0 {
		return val
	}
	return defaultValue
}

func parseInt64QueryParam(r *http.Request, key string) (int64, error) {
	valStr := r.URL.Query().Get(key)
	if valStr == "" {
		return 0, fmt.Errorf("%s is required", key)
	}
	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return val, nil
}

func parseBoolQueryParam(r *http.Request, key string, defaultValue bool) bool {
	valStr := r.URL.Query().Get(key)
	if valStr == "false" || valStr == "0" {
		return false
	}
	if valStr == "true" || valStr == "1" {
		return true
	}
	return defaultValue
}

func validateFilePath(filePath, baseDir string) error {
	cleanPath := filepath.Clean(filePath)
	cleanBase := filepath.Clean(baseDir)
	if !strings.HasPrefix(cleanPath, cleanBase) {
		return fmt.Errorf("invalid file path: outside base directory")
	}
	return nil
}

func validateDownloadURL(url string) error {
	if !strings.HasPrefix(url, allowedURLPrefix) {
		return fmt.Errorf("invalid download URL: must be from official repository")
	}
	return nil
}

func validateAssetName(name string) error {
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("invalid asset name: path traversal detected")
	}
	return nil
}

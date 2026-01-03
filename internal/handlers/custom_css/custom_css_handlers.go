//go:build !server

package custom_css

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/utils"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const customCSSFileName = "custom_article.css"

// HandleUploadCSSDialog opens a file dialog to select CSS file for upload.
// @Summary      Upload CSS dialog (desktop mode)
// @Description  Open a file dialog to select a CSS file for upload (desktop mode only)
// @Tags         custom-css
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string  "Upload success (status, message)"
// @Success      501  {object}  map[string]string  "Not implemented in server mode"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /custom-css/dialog [post]
func HandleUploadCSSDialog(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if h.App == nil {
		log.Printf("File dialog not available")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "File dialog not available",
		})
		return
	}

	// Type assert to *application.App to access Dialog
	app, ok := h.App.(*application.App)
	if !ok {
		log.Printf("File dialog not available: app is not *application.App type")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "File dialog not available",
		})
		return
	}

	filePath, err := app.Dialog.OpenFileWithOptions(&application.OpenFileDialogOptions{
		Title: "Select CSS File",
		Filters: []application.FileFilter{
			{
				DisplayName: "CSS Files (*.css)",
				Pattern:     "*.css",
			},
			{
				DisplayName: "All Files (*)",
				Pattern:     "*",
			},
		},
		CanChooseFiles:       true,
		AllowsOtherFileTypes: true,
	}).PromptForSingleSelection()
	if err != nil {
		log.Printf("Error opening file dialog: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to open file dialog",
		})
		return
	}

	if filePath == "" {
		// User cancelled the dialog
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "cancelled"})
		return
	}

	// Read the selected file
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening selected file: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to open selected file",
		})
		return
	}
	defer file.Close()

	// Get file info for validation
	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("Error getting file info: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to get file info",
		})
		return
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".css" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Only CSS files are allowed",
		})
		return
	}

	// Validate file size (max 1MB)
	if fileInfo.Size() > 1<<20 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "CSS file is too large (max 1MB)",
		})
		return
	}

	// Get data directory
	dataDir, err := utils.GetDataDir()
	if err != nil {
		log.Printf("Error getting data directory: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to get data directory",
		})
		return
	}

	// Save CSS file
	cssFilePath := filepath.Join(dataDir, customCSSFileName)
	destFile, err := os.Create(cssFilePath)
	if err != nil {
		log.Printf("Error creating CSS file: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to save CSS file",
		})
		return
	}
	defer destFile.Close()

	// Copy file content
	written, err := io.Copy(destFile, file)
	if err != nil {
		log.Printf("Error writing CSS file: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to write CSS file",
		})
		return
	}

	log.Printf("CSS file uploaded via dialog: %s (%d bytes)", filePath, written)

	// Update setting in database
	if err := h.DB.SetSetting("custom_css_file", customCSSFileName); err != nil {
		log.Printf("Error saving custom_css_file setting: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to update settings",
		})
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "CSS file uploaded successfully",
	})
}

// HandleUploadCSS handles CSS file upload and saves it to the data directory
// @Summary      Upload custom CSS file
// @Description  Upload a custom CSS file to style article content (max 1MB, .css files only)
// @Tags         custom-css
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "CSS file to upload"
// @Success      200  {object}  map[string]string  "Upload success (status, message)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid file or size)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /custom-css/upload [post]
func HandleUploadCSS(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error getting form file: %v", err)
		http.Error(w, "Failed to get file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".css" {
		http.Error(w, "Only CSS files are allowed", http.StatusBadRequest)
		return
	}

	// Validate file size (max 1MB)
	if header.Size > 1<<20 {
		http.Error(w, "CSS file is too large (max 1MB)", http.StatusBadRequest)
		return
	}

	// Get data directory
	dataDir, err := utils.GetDataDir()
	if err != nil {
		log.Printf("Error getting data directory: %v", err)
		http.Error(w, "Failed to get data directory", http.StatusInternalServerError)
		return
	}

	// Save CSS file
	cssFilePath := filepath.Join(dataDir, customCSSFileName)
	destFile, err := os.Create(cssFilePath)
	if err != nil {
		log.Printf("Error creating CSS file: %v", err)
		http.Error(w, "Failed to save CSS file", http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	// Copy file content
	written, err := io.Copy(destFile, file)
	if err != nil {
		log.Printf("Error writing CSS file: %v", err)
		http.Error(w, "Failed to write CSS file", http.StatusInternalServerError)
		return
	}

	log.Printf("CSS file uploaded successfully: %s (%d bytes)", header.Filename, written)

	// Update setting in database
	if err := h.DB.SetSetting("custom_css_file", customCSSFileName); err != nil {
		log.Printf("Error saving custom_css_file setting: %v", err)
		http.Error(w, "Failed to update settings", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success", "message": "CSS file uploaded successfully"}`))
}

// HandleGetCSS returns the custom CSS file content
// @Summary      Get custom CSS
// @Description  Get the content of the uploaded custom CSS file
// @Tags         custom-css
// @Accept       json
// @Produce      text/css
// @Success      200  {string}  string  "CSS file content"
// @Failure      404  {object}  map[string]string  "No custom CSS file configured"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /custom-css [get]
func HandleGetCSS(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get custom_css_file setting
	cssFileName, err := h.DB.GetSetting("custom_css_file")
	if err != nil || cssFileName == "" {
		http.Error(w, "No custom CSS file configured", http.StatusNotFound)
		return
	}

	// Get data directory
	dataDir, err := utils.GetDataDir()
	if err != nil {
		log.Printf("Error getting data directory: %v", err)
		http.Error(w, "Failed to get data directory", http.StatusInternalServerError)
		return
	}

	// Read CSS file
	cssFilePath := filepath.Join(dataDir, cssFileName)
	cssContent, err := os.ReadFile(cssFilePath)
	if err != nil {
		log.Printf("Error reading CSS file: %v", err)
		http.Error(w, "Failed to read CSS file", http.StatusInternalServerError)
		return
	}

	// Set content type and return CSS
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(cssContent)
}

// HandleDeleteCSS deletes the custom CSS file and clears the setting
// @Summary      Delete custom CSS
// @Description  Delete the custom CSS file and clear the setting
// @Tags         custom-css
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string  "Delete success (status, message)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /custom-css [delete]
func HandleDeleteCSS(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get custom_css_file setting
	cssFileName, err := h.DB.GetSetting("custom_css_file")
	if err != nil || cssFileName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "success", "message": "No custom CSS file to delete"}`))
		return
	}

	// Get data directory
	dataDir, err := utils.GetDataDir()
	if err != nil {
		log.Printf("Error getting data directory: %v", err)
		http.Error(w, "Failed to get data directory", http.StatusInternalServerError)
		return
	}

	// Delete CSS file
	cssFilePath := filepath.Join(dataDir, cssFileName)
	if err := os.Remove(cssFilePath); err != nil && !os.IsNotExist(err) {
		log.Printf("Error deleting CSS file: %v", err)
		http.Error(w, "Failed to delete CSS file", http.StatusInternalServerError)
		return
	}

	// Clear setting in database
	if err := h.DB.SetSetting("custom_css_file", ""); err != nil {
		log.Printf("Error clearing custom_css_file setting: %v", err)
		http.Error(w, "Failed to update settings", http.StatusInternalServerError)
		return
	}

	log.Printf("Custom CSS file deleted: %s", cssFileName)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "success", "message": "CSS file deleted successfully"}`))
}

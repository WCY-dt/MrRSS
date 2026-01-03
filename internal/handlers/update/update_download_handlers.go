package update

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"MrRSS/internal/handlers/core"
)

// HandleDownloadUpdate downloads the update file.
// @Summary      Download update
// @Description  Download the update file from GitHub releases to the temp directory
// @Tags         update
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Download request (download_url, asset_name)"
// @Success      200  {object}  map[string]interface{}  "Download success (success, file_path, total_bytes, bytes_written)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid URL or asset name)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /update/download [post]
func HandleDownloadUpdate(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		DownloadURL string `json:"download_url"`
		AssetName   string `json:"asset_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate download URL is from the official GitHub repository releases
	const allowedURLPrefix = "https://github.com/WCY-dt/MrRSS/releases/download/"
	if !strings.HasPrefix(req.DownloadURL, allowedURLPrefix) {
		log.Printf("Invalid download URL attempted: %s", req.DownloadURL)
		http.Error(w, "Invalid download URL", http.StatusBadRequest)
		return
	}

	// Validate asset name to prevent path traversal
	if strings.Contains(req.AssetName, "..") || strings.Contains(req.AssetName, "/") || strings.Contains(req.AssetName, "\\") {
		log.Printf("Invalid asset name attempted: %s", req.AssetName)
		http.Error(w, "Invalid asset name", http.StatusBadRequest)
		return
	}

	// Create temp directory for download
	tempDir := os.TempDir()
	filePath := filepath.Join(tempDir, req.AssetName)

	// Download the file
	log.Printf("Downloading update from: %s", req.DownloadURL)
	resp, err := http.Get(req.DownloadURL)
	if err != nil {
		log.Printf("Error downloading update: %v", err)
		http.Error(w, "Failed to download update", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Download failed with status: %d", resp.StatusCode)
		http.Error(w, "Failed to download update", http.StatusInternalServerError)
		return
	}

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		http.Error(w, "Failed to create download file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Write the body to file with progress tracking
	totalSize := resp.ContentLength
	var bytesWritten int64

	// Create a buffer for efficient copying
	buffer := make([]byte, 32*1024) // 32KB buffer

	for {
		nr, er := resp.Body.Read(buffer)
		if nr > 0 {
			nw, ew := out.Write(buffer[0:nr])
			if nw > 0 {
				bytesWritten += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	if err != nil {
		log.Printf("Error writing file: %v", err)
		os.Remove(filePath) // Clean up partial file
		http.Error(w, "Failed to write download file", http.StatusInternalServerError)
		return
	}

	// Ensure all data is flushed to disk
	if err := out.Sync(); err != nil {
		log.Printf("Error syncing file: %v", err)
		os.Remove(filePath) // Clean up
		http.Error(w, "Failed to save download file", http.StatusInternalServerError)
		return
	}

	// Verify the file size matches expected size
	if totalSize > 0 && bytesWritten != totalSize {
		log.Printf("Download incomplete: expected %d bytes, got %d bytes", totalSize, bytesWritten)
		os.Remove(filePath) // Clean up incomplete file
		http.Error(w, "Download incomplete", http.StatusInternalServerError)
		return
	}

	log.Printf("Update downloaded successfully to: %s (%.2f MB)", filePath, float64(bytesWritten)/(1024*1024))

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":       true,
		"file_path":     filePath,
		"total_bytes":   totalSize,
		"bytes_written": bytesWritten,
	})
}

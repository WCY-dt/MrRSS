package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"MrRSS/internal/version"
)

// HandleCheckUpdates checks for the latest version on GitHub
func (h *Handler) HandleCheckUpdates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	updateInfo, err := h.checkForUpdates()
	if err != nil {
		log.Printf("Error checking for updates: %v", err)
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"current_version": version.Version,
			"error":           err.Error(),
		})
		return
	}

	respondWithJSON(w, http.StatusOK, updateInfo)
}

func (h *Handler) checkForUpdates() (map[string]interface{}, error) {
	currentVersion := version.Version

	resp, err := http.Get(githubAPILatestRelease)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status: %d", resp.StatusCode)
	}

	var release struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		HTMLURL     string `json:"html_url"`
		Body        string `json:"body"`
		PublishedAt string `json:"published_at"`
		Assets      []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release information: %w", err)
	}

	// Remove 'v' prefix if present for comparison
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	hasUpdate := compareVersions(latestVersion, currentVersion) > 0

	// Find the appropriate download URL based on platform
	platform := runtime.GOOS
	arch := runtime.GOARCH
	downloadURL, assetName, assetSize := h.findPlatformAsset(release.Assets, platform, arch)

	response := map[string]interface{}{
		"current_version": currentVersion,
		"latest_version":  latestVersion,
		"has_update":      hasUpdate,
		"platform":        platform,
		"arch":            arch,
	}

	if downloadURL != "" {
		response["download_url"] = downloadURL
		response["asset_name"] = assetName
		response["asset_size"] = assetSize
	}

	return response, nil
}

func (h *Handler) findPlatformAsset(assets []struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}, platform, arch string) (string, string, int64) {
	platformArch := platform + "-" + arch

	for _, asset := range assets {
		name := strings.ToLower(asset.Name)

		// Match platform-specific installer/package with architecture
		if platform == "windows" && strings.Contains(name, platformArch) && strings.HasSuffix(name, "-installer.exe") {
			return asset.BrowserDownloadURL, asset.Name, asset.Size
		}
		if platform == "linux" && strings.Contains(name, platformArch) && strings.HasSuffix(name, ".appimage") {
			return asset.BrowserDownloadURL, asset.Name, asset.Size
		}
		if platform == "darwin" && strings.Contains(name, "darwin-universal") && strings.HasSuffix(name, ".dmg") {
			return asset.BrowserDownloadURL, asset.Name, asset.Size
		}
	}

	return "", "", 0
}

// compareVersions compares two semantic versions (e.g., "1.1.0" vs "1.0.0")
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var p1, p2 int
		if i < len(parts1) {
			p1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			p2, _ = strconv.Atoi(parts2[i])
		}

		if p1 > p2 {
			return 1
		} else if p1 < p2 {
			return -1
		}
	}

	return 0
}

// HandleDownloadUpdate downloads the update file
func (h *Handler) HandleDownloadUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req downloadUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate download URL
	if err := validateDownloadURL(req.DownloadURL); err != nil {
		log.Printf("Invalid download URL attempted: %s", req.DownloadURL)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate asset name
	if err := validateAssetName(req.AssetName); err != nil {
		log.Printf("Invalid asset name attempted: %s", req.AssetName)
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Download the file
	filePath, bytesWritten, totalSize, err := h.downloadFile(req.DownloadURL, req.AssetName)
	if err != nil {
		log.Printf("Error downloading update: %v", err)
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	log.Printf("Update downloaded successfully to: %s (%.2f MB)", filePath, float64(bytesWritten)/(1024*1024))

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success":       true,
		"file_path":     filePath,
		"total_bytes":   totalSize,
		"bytes_written": bytesWritten,
	})
}

func (h *Handler) downloadFile(downloadURL, assetName string) (string, int64, int64, error) {
	// Create temp directory for download
	tempDir := os.TempDir()
	filePath := filepath.Join(tempDir, assetName)

	log.Printf("Downloading update from: %s", downloadURL)
	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, 0, fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Create the file
	out, err := os.Create(filePath)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to create download file: %w", err)
	}
	defer out.Close()

	// Write the body to file with progress tracking
	totalSize := resp.ContentLength
	var bytesWritten int64

	buffer := make([]byte, downloadBufferSize)
	for {
		nr, er := resp.Body.Read(buffer)
		if nr > 0 {
			nw, ew := out.Write(buffer[0:nr])
			if nw > 0 {
				bytesWritten += int64(nw)
			}
			if ew != nil {
				os.Remove(filePath)
				return "", 0, 0, fmt.Errorf("failed to write download file: %w", ew)
			}
			if nr != nw {
				os.Remove(filePath)
				return "", 0, 0, io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				os.Remove(filePath)
				return "", 0, 0, fmt.Errorf("error reading response: %w", er)
			}
			break
		}
	}

	// Ensure all data is flushed to disk
	if err := out.Sync(); err != nil {
		os.Remove(filePath)
		return "", 0, 0, fmt.Errorf("failed to save download file: %w", err)
	}

	// Verify the file size matches expected size
	if totalSize > 0 && bytesWritten != totalSize {
		os.Remove(filePath)
		return "", 0, 0, fmt.Errorf("download incomplete: expected %d bytes, got %d bytes", totalSize, bytesWritten)
	}

	return filePath, bytesWritten, totalSize, nil
}

// HandleInstallUpdate triggers the installation of the downloaded update
func (h *Handler) HandleInstallUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req installUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate and prepare file
	cleanPath, err := h.validateInstallerFile(req.FilePath)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Start installer
	if err := h.startInstaller(cleanPath); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithSuccess(w, "Installation started. Application will exit shortly.")

	// Schedule graceful shutdown
	go func() {
		time.Sleep(shutdownDelay)
		log.Println("Initiating graceful shutdown for update installation...")
		os.Exit(0)
	}()
}

func (h *Handler) validateInstallerFile(filePath string) (string, error) {
	// Validate file path is within temp directory
	tempDir := os.TempDir()
	if err := validateFilePath(filePath, tempDir); err != nil {
		log.Printf("Invalid file path attempted: %s", filePath)
		return "", err
	}

	cleanPath := filepath.Clean(filePath)

	// Validate file exists and is a regular file
	fileInfo, err := os.Stat(cleanPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("update file not found")
	}
	if err != nil {
		log.Printf("Error stating file: %v", err)
		return "", fmt.Errorf("error accessing update file")
	}
	if !fileInfo.Mode().IsRegular() {
		log.Printf("File is not a regular file: %s", cleanPath)
		return "", fmt.Errorf("invalid file type")
	}

	return cleanPath, nil
}

func (h *Handler) startInstaller(installerPath string) error {
	platform := runtime.GOOS
	log.Printf("Installing update from: %s on platform: %s", installerPath, platform)

	var cmd *exec.Cmd
	var cleanupDelay time.Duration

	switch platform {
	case "windows":
		if err := h.validateFileExtension(installerPath, ".exe"); err != nil {
			return err
		}
		cmd = exec.Command("cmd.exe", "/C", "start", "/B", installerPath)
		cleanupDelay = windowsCleanupDelay

	case "linux":
		if err := h.validateFileExtension(installerPath, ".appimage"); err != nil {
			return err
		}
		if err := os.Chmod(installerPath, 0755); err != nil {
			log.Printf("Error making file executable: %v", err)
			return fmt.Errorf("failed to prepare installer")
		}
		cmd = exec.Command(installerPath)
		cleanupDelay = linuxCleanupDelay

	case "darwin":
		if err := h.validateFileExtension(installerPath, ".dmg"); err != nil {
			return err
		}
		cmd = exec.Command("open", installerPath)
		cleanupDelay = macosCleanupDelay

	default:
		return fmt.Errorf("unsupported platform: %s", platform)
	}

	// Start the installer in the background
	if err := cmd.Start(); err != nil {
		log.Printf("Error starting installer: %v", err)
		return fmt.Errorf("failed to start installer")
	}

	log.Printf("Installer started successfully, PID: %d", cmd.Process.Pid)

	// Schedule cleanup
	h.scheduleFileCleanup(installerPath, cleanupDelay)

	return nil
}

func (h *Handler) validateFileExtension(filePath, expectedExt string) error {
	if !strings.HasSuffix(strings.ToLower(filePath), expectedExt) {
		return fmt.Errorf("invalid file type: expected %s", expectedExt)
	}
	return nil
}

func (h *Handler) scheduleFileCleanup(filePath string, delay time.Duration) {
	go func() {
		time.Sleep(delay)
		if err := os.Remove(filePath); err != nil {
			log.Printf("Failed to remove installer: %v", err)
		} else {
			log.Printf("Successfully removed installer: %s", filePath)
		}
	}()
}

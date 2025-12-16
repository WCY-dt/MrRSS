package platform

import (
	"encoding/json"
	"net/http"
	"runtime"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/utils"
)

// PlatformInfo represents platform and OS information
type PlatformInfo struct {
	OS                string `json:"os"`
	Arch              string `json:"arch"`
	IsMacOS           bool   `json:"is_macos"`
	IsWindows         bool   `json:"is_windows"`
	MacOSVersion      string `json:"macos_version,omitempty"`
	IsMacOSSequoia    bool   `json:"is_macos_sequoia"`
	NeedsUIThrottling bool   `json:"needs_ui_throttling"` // True for macOS 15+ due to WKWebView issues
}

// HandleGetPlatformInfo returns platform and OS information
func HandleGetPlatformInfo(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := PlatformInfo{
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		IsMacOS:   utils.IsMacOS(),
		IsWindows: utils.IsWindows(),
	}

	// Add macOS-specific information
	if info.IsMacOS {
		info.MacOSVersion = utils.GetMacOSVersionString()
		info.IsMacOSSequoia = utils.IsMacOSSequoia()
		// macOS 15 (Sequoia) and later have UI refresh issues with WKWebView
		info.NeedsUIThrottling = utils.IsMacOS15OrLater()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// +build darwin

package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// macOSVersion represents a macOS version
type macOSVersion struct {
	Major int
	Minor int
	Patch int
}

// GetMacOSVersion returns the current macOS version
// Returns (major, minor, patch, error)
func GetMacOSVersion() (int, int, int, error) {
	// Use sw_vers command to get macOS version
	cmd := exec.Command("sw_vers", "-productVersion")
	var out bytes.Buffer
	cmd.Stdout = &out
	
	if err := cmd.Run(); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to get macOS version: %w", err)
	}
	
	versionStr := strings.TrimSpace(out.String())
	
	// Parse version string (e.g., "15.0.1" or "14.5")
	re := regexp.MustCompile(`^(\d+)(?:\.(\d+))?(?:\.(\d+))?`)
	matches := re.FindStringSubmatch(versionStr)
	
	if len(matches) < 2 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s", versionStr)
	}
	
	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %s", matches[1])
	}
	
	minor := 0
	if len(matches) > 2 && matches[2] != "" {
		minor, _ = strconv.Atoi(matches[2])
	}
	
	patch := 0
	if len(matches) > 3 && matches[3] != "" {
		patch, _ = strconv.Atoi(matches[3])
	}
	
	return major, minor, patch, nil
}

// IsMacOS15OrLater checks if the current macOS version is 15 (Sequoia) or later
// macOS 15 is also known as macOS Sequoia or "macOS 26" in some references
func IsMacOS15OrLater() bool {
	major, _, _, err := GetMacOSVersion()
	if err != nil {
		// If we can't detect, assume false (don't apply workarounds)
		return false
	}
	return major >= 15
}

// IsMacOSSequoia checks if the current macOS version is Sequoia (15.x)
// This is the version that has UI refresh rate issues with WKWebView
func IsMacOSSequoia() bool {
	major, _, _, err := GetMacOSVersion()
	if err != nil {
		return false
	}
	return major == 15
}

// GetMacOSVersionString returns a formatted version string
func GetMacOSVersionString() string {
	major, minor, patch, err := GetMacOSVersion()
	if err != nil {
		return "unknown"
	}
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

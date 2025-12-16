//go:build !darwin

package utils

import "fmt"

// GetMacOSVersion returns an error on non-macOS platforms
func GetMacOSVersion() (int, int, int, error) {
	return 0, 0, 0, fmt.Errorf("not running on macOS")
}

// IsMacOS15OrLater always returns false on non-macOS platforms
func IsMacOS15OrLater() bool {
	return false
}

// IsMacOSSequoia always returns false on non-macOS platforms
func IsMacOSSequoia() bool {
	return false
}

// GetMacOSVersionString returns "N/A" on non-macOS platforms
func GetMacOSVersionString() string {
	return "N/A"
}

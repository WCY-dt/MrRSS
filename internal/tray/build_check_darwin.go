//go:build darwin && !nosystray

package tray

// This file should NEVER compile on macOS.
// If you see a compilation error from this file, it means the build tags are incorrect.
//
// SOLUTION: Always use `-tags nosystray` when building on macOS:
//   - wails dev -tags nosystray
//   - wails build -skipbindings -tags nosystray
//   - make build (automatically applies correct tags)
//
// This prevents AppDelegate symbol conflicts between Wails and fyne.io/systray.

import (
	_ "embed"
)

//go:embed build_error_macos.txt
var buildErrorMessage string

func init() {
	// This should never execute
	panic(buildErrorMessage)
}

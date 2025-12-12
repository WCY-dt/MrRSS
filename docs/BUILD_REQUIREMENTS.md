# Build Requirements

This document describes the system-level dependencies required for building MrRSS on different platforms.

## Overview

MrRSS uses several native libraries that require CGO (C bindings for Go):

- **Wails v2**: For the desktop application framework
- **fyne.io/systray**: For system tray integration (cross-platform, compatible with Wails)
- **SQLite**: Pure Go implementation (`modernc.org/sqlite`), no C dependencies

## Important: CGO Requirement

⚠️ **CRITICAL**: `fyne.io/systray` requires CGO to be enabled. You must set:

```bash
export CGO_ENABLED=1
```

Or when building:

```bash
CGO_ENABLED=1 wails build
```

## Platform-Specific Requirements

### Linux

#### Development Dependencies

```bash
sudo apt-get update
sudo apt-get install -y \
  gcc \
  pkg-config \
  libgtk-3-dev \
  libwebkit2gtk-4.1-dev \
  libayatana-appindicator3-dev
```

**Dependency Breakdown**:

- `gcc`: C compiler (required for CGO)
- `pkg-config`: Build tool for finding libraries
- `libgtk-3-dev`: GTK3 development headers (for Wails UI)
- `libwebkit2gtk-4.1-dev`: WebKit2GTK development headers (for Wails webview)
- `libayatana-appindicator3-dev`: AppIndicator development headers (for systray)

**Note for Linux Mint**: Also install `libxapp-dev`

#### Legacy AppIndicator Support

If you need to support the older `libappindicator3` library instead of `libayatana-appindicator3`, install:

```bash
sudo apt-get install libappindicator3-dev
```

And build with the `legacy_appindicator` tag:

```bash
go build -tags=legacy_appindicator
```

#### Runtime Dependencies

End users running the compiled binary will need:

- `libgtk-3-0`
- `libwebkit2gtk-4.1-0`
- `libayatana-appindicator3-1` (or `libappindicator3-1` for legacy)

### Windows

#### Development Dependencies

Install via Chocolatey:

```powershell
choco install mingw nsis -y
```

**Dependency Breakdown**:

- `mingw`: MinGW-w64 GCC compiler (required for CGO)
- `nsis`: Nullsoft Scriptable Install System (for creating installers)

#### Alternative: Manual Installation

If not using Chocolatey:

1. Install [MinGW-w64](https://www.mingw-w64.org/)
2. Install [NSIS](https://nsis.sourceforge.io/) (optional, for installers)
3. Add MinGW `bin` directory to PATH

#### Build Flags

To avoid opening a console at application startup:

```bash
go build -ldflags "-H=windowsgui"
```

Or with Wails:

```bash
wails build -ldflags "-H=windowsgui"
```

#### Runtime Dependencies

Windows binaries are statically linked and don't require additional runtime dependencies.

### macOS

#### Development Dependencies

Install Xcode Command Line Tools (if not already installed):

```bash
xcode-select --install
```

**Note**: macOS uses [DarwinKit](https://github.com/progrium/darwinkit) for native system tray support via NSStatusItem. This avoids the AppDelegate symbol conflicts that occur with `fyne.io/systray` on macOS.

#### Building on macOS

macOS requires the `-tags nosystray` build tag to exclude `fyne.io/systray` package (which conflicts with Wails' AppDelegate):

```bash
# Development
wails dev -tags nosystray

# Production build
wails build -skipbindings -tags nosystray

# Or use make (automatically uses correct tags)
make build
```

#### Application Bundle

macOS requires an application bundle structure:

```plaintext
MrRSS.app/
  Contents/
    Info.plist
    MacOS/
      MrRSS
    Resources/
      MrRSS.icns
```

Wails automatically creates this structure during build.

#### Info.plist Settings

Add these keys for better macOS integration:

```xml
<!-- High resolution support -->
<key>NSHighResolutionCapable</key>
<string>True</string>

<!-- Hide from Dock (optional, for menu bar only apps) -->
<key>LSUIElement</key>
<string>1</string>
```

#### Runtime Dependencies

macOS binaries are self-contained and don't require additional runtime dependencies.

## Building with Wails

### Standard Build

```bash
# Development build with hot reload
wails dev

# Production build
wails build -clean -ldflags "-s -w"

# Platform-specific build
wails build -platform linux/amd64
wails build -platform windows/amd64
wails build -platform darwin/universal
```

### Important Wails Flags

- `-clean`: Clean build directory before building
- `-ldflags "-s -w"`: Strip debug information for smaller binaries
- `-skipbindings`: Skip generating bindings (MrRSS uses HTTP API)
- `-platform <os/arch>`: Cross-compile for specific platform

### Cross-Compilation

**Note**: Cross-compilation with CGO is complex. For best results:

- Build Linux binaries on Linux
- Build Windows binaries on Windows
- Build macOS binaries on macOS

GitHub Actions handles this automatically using platform-specific runners.

## GitHub Actions

Our CI/CD pipeline automatically installs all required dependencies:

### Test Workflow

- Installs Linux dependencies for backend tests
- Sets `CGO_ENABLED=1`

### Release Workflow

- Platform-specific dependency installation
- Cross-platform builds using native runners
- Artifact creation (installers, AppImages, DMGs)

## Troubleshooting

### "CGO is disabled" Error

**Solution**: Enable CGO before building:

```bash
export CGO_ENABLED=1
wails build
```

### Linux: "Package webkit2gtk-4.1 was not found"

**Solution**: Install webkit2gtk-4.1 development headers:

```bash
sudo apt-get install libwebkit2gtk-4.1-dev
```

### Linux: "Package ayatana-appindicator3-0.1 was not found"

**Solution**: Install libayatana-appindicator3 development headers:

```bash
sudo apt-get install libayatana-appindicator3-dev
```

For older distributions, use the legacy version:

```bash
sudo apt-get install libappindicator3-dev
go build -tags=legacy_appindicator
```

### Windows: "gcc: command not found"

**Solution**: Install MinGW:

```powershell
choco install mingw -y
```

Or download from [mingw-w64.org](https://www.mingw-w64.org/) and add to PATH.

### macOS: Missing Xcode Command Line Tools

**Solution**: Install Xcode Command Line Tools:

```bash
xcode-select --install
```

## Development Environment Setup

### Quick Setup Scripts

**Linux/macOS**:

```bash
# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend
npm install
cd ..

# Run development server
wails dev
```

**Windows (PowerShell)**:

```powershell
# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend
npm install
cd ..

# Run development server
wails dev
```

## Related Documentation

- [Architecture Overview](ARCHITECTURE.md)
- [Code Patterns](CODE_PATTERNS.md)
- [Testing Guide](TESTING.md)

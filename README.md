<div align="center">
  <img src="imgs/logo.svg" alt="MrRSS Logo" width="120" height="120">
  <h1>MrRSS</h1>
  <p><strong>A Modern, Cross-Platform Desktop RSS Reader</strong></p>

   [![Version](https://img.shields.io/badge/version-1.2.14-blue.svg)](https://github.com/WCY-dt/MrRSS/releases)
  [![License](https://img.shields.io/badge/license-GPLv3-green.svg)](LICENSE)
   [![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)
  [![Wails](https://img.shields.io/badge/Wails-v2-red)](https://wails.io/)
  [![Vue.js](https://img.shields.io/badge/Vue.js-3.x-4FC08D?logo=vue.js)](https://vuejs.org/)

  <p>
    <strong>English</strong> | <a href="README_zh.md">简体中文</a>
  </p>
</div>

---

## ✨ Features

- 🖥️ **Cross-Platform**: Native desktop app for Windows, macOS, and Linux
- 🌐 **Auto-Translation**: Translate article titles and content using translate service or AI
- 🔍 **Smart Feed Discovery**: Automatically discover new feeds from friend links and related sources
- ⭐ **Favorites & Reading Tracking**: Save articles and track read/unread status
- 📝 **Article Summarization**: Generate article summaries using local TF-IDF/TextRank algorithms or AI
- 🎨 **Modern UI**: Clean, responsive interface with dark mode support
- 📦 **OPML Import/Export**: Easy migration from other RSS readers
- ⌨️ **Keyboard Shortcuts**: Boost productivity with customizable keyboard shortcuts for quick navigation and actions
- 📋 **Smart Filtering Rules**: Automate feed organization with powerful custom rules
- 🏭 **Custom Automation Scripts**: Support for running user-defined automation scripts to fetch feeds

## 📸 Screenshots

![Screenshot](imgs/og.png)

## 🚀 Quick Start

### Download and Install

#### Option 1: Download Pre-built Installer (Recommended)

Download the latest installer for your platform from the [Releases](https://github.com/WCY-dt/MrRSS/releases/latest) page:

**Standard Installation:**

- **Windows:** `MrRSS-{version}-windows-amd64-installer.exe` / `MrRSS-{version}-windows-arm64-installer.exe`
- **macOS:** `MrRSS-{version}-darwin-universal.dmg`
- **Linux:** `MrRSS-{version}-linux-amd64.AppImage` / `MrRSS-{version}-linux-arm64.AppImage`

**Portable Version** (no installation required, all data in one folder):

- **Windows:** `MrRSS-{version}-windows-{arch}-portable.zip`
- **Linux:** `MrRSS-{version}-linux-{arch}-portable.tar.gz`
- **macOS:** `MrRSS-{version}-darwin-{arch}-portable.zip`

#### Option 2: Build from Source

<details>

<summary>Click to expand the build from source guide</summary>

<div markdown="1">

### Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://go.dev/) (1.24 or higher)
- [Node.js](https://nodejs.org/) (LTS version with npm)
- [Wails](https://wails.io/docs/gettingstarted/installation) CLI

**Platform-specific requirements:**

- **Linux**: System dependencies required (GTK, WebKit, AppIndicator)
- **Windows**: MinGW required (for CGO/systray support)
- **macOS**: Xcode Command Line Tools

For detailed installation instructions, see [Build Requirements](docs/BUILD_REQUIREMENTS.md)

```bash
# Quick setup for Linux (Ubuntu/Debian):
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev libayatana-appindicator3-dev gcc pkg-config

# For older Ubuntu versions (before 24.04):
# sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev libappindicator3-dev gcc pkg-config
```

### Installation

1. **Clone the repository**

   ```bash
   git clone https://github.com/WCY-dt/MrRSS.git
   cd MrRSS
   ```

2. **Install frontend dependencies**

   ```bash
   cd frontend
   npm install
   cd ..
   ```

3. **Build the application**

   ```bash
   # Using Makefile (recommended - auto-detects platform)
   make build

   # Or manually with wails
   # Linux/Windows:
   wails build -skipbindings

   # macOS (IMPORTANT - use -tags nosystray):
   wails build -skipbindings -tags nosystray
   ```

   > **⚠️ macOS**: Always use `make build` or add `-tags nosystray` to prevent build errors.

   The executable will be created in the `build/bin` directory.

4. **Run the application**

   - Windows: `build/bin/MrRSS.exe`
   - macOS: `build/bin/MrRSS.app`
   - Linux: `build/bin/MrRSS`

</div>

</details>

### Data Storage

**Normal Mode** (default):

- **Windows:** `%APPDATA%\MrRSS\` (e.g., `C:\Users\YourName\AppData\Roaming\MrRSS\`)
- **macOS:** `~/Library/Application Support/MrRSS/`
- **Linux:** `~/.local/share/MrRSS/`

**Portable Mode** (when `portable.txt` exists):

- All data stored in `data/` folder

This ensures your data persists across application updates and reinstalls.

## 🛠️ Development

<details>

<summary>Click to expand the development guide</summary>

<div markdown="1">

### Running in Development Mode

Start the application with hot reloading:

```bash
wails dev
```

> **⚠️ Important for macOS Developers**: Always use `make dev` or add `-tags nosystray` flag when building on macOS to prevent AppDelegate conflicts. See [Build Requirements](docs/BUILD_REQUIREMENTS.md#macos) for details.

### Code Quality Tools

#### Using Make

We provide a `Makefile` with common development tasks (works on Linux/macOS/Windows with Make installed):

```bash
# Show all available commands
make help

# Run full check (lint + test + build)
make check

# Clean build artifacts
make clean

# Setup development environment
make setup
```

#### Using Scripts

Cross-platform scripts are available in the `scripts/` directory:

**Linux/macOS:**

```bash
# Run all checks
./scripts/check.sh

# Pre-release checks
./scripts/pre-release.sh
```

**Windows (PowerShell):**

```powershell
# Run all checks
.\scripts\check.ps1

# Pre-release checks
.\scripts\pre-release.ps1
```

### Pre-commit Hooks

This project uses pre-commit hooks to ensure code quality:

```bash
# Install hooks
pre-commit install

# Run on all files
pre-commit run --all-files
```

### Running Tests

```bash
# Backend tests
go test ./...

# Frontend tests
cd frontend
npm test
```

</div>

</details>

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

Before contributing:

1. Read the [Code of Conduct](CODE_OF_CONDUCT.md)
2. Check existing issues or create a new one
3. Fork the repository and create a feature branch
4. Make your changes and add tests
5. Submit a pull request

## 🔒 Security

If you discover a security vulnerability, please follow our [Security Policy](SECURITY.md).

## 📝 License

This project is licensed under the GPLv3 License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Wails](https://wails.io/) - Go + Web framework
- UI powered by [Vue.js 3](https://vuejs.org/) and [Tailwind CSS](https://tailwindcss.com/)
- Icons from [Phosphor Icons](https://phosphoricons.com/)
- RSS parsing with [gofeed](https://github.com/mmcdole/gofeed)

## 📮 Contact & Support

- **Issues**: [GitHub Issues](https://github.com/WCY-dt/MrRSS/issues)
- **Repository**: [github.com/WCY-dt/MrRSS](https://github.com/WCY-dt/MrRSS)

---

<div align="center">
  <p>Made with ❤️ by the MrRSS Team</p>
  <p>⭐ Star us on GitHub if you find this project useful!</p>
</div>

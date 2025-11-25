<div align="center">
  <img src="imgs/logo.svg" alt="MrRSS Logo" width="120" height="120">
  <h1>MrRSS</h1>
  <p><strong>A Modern, Cross-Platform Desktop RSS Reader</strong></p>
  
   [![Version](https://img.shields.io/badge/version-1.2.1-blue.svg)](https://github.com/WCY-dt/MrRSS/releases)
  [![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
  [![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
  [![Wails](https://img.shields.io/badge/Wails-v2-red)](https://wails.io/)
  [![Vue.js](https://img.shields.io/badge/Vue.js-3.x-4FC08D?logo=vue.js)](https://vuejs.org/)

  <p>
    <strong>English</strong> | <a href="README_zh.md">简体中文</a>
  </p>
</div>

---

## ✨ Features

- 🖥️ **Cross-Platform**: Native desktop app for Windows, macOS, and Linux
- 🌐 **Auto-Translation**: Translate article titles using Google Translate or DeepL API
- 🔍 **Smart Feed Discovery**: Automatically discover new feeds from friend links and related sources
- ⭐ **Favorites & Reading Tracking**: Save articles and track read/unread status
- 🎨 **Modern UI**: Clean, responsive interface with dark mode support
- 📦 **OPML Import/Export**: Easy migration from other RSS readers

## 📸 Screenshots

![Screenshot](imgs/og.png)

## 🚀 Quick Start

### Download and Install

#### Option 1: Download Pre-built Installer (Recommended)

Download the latest installer for your platform from the [Releases](https://github.com/WCY-dt/MrRSS/releases) page:

- **Windows:** `MrRSS-{version}-windows-amd64-installer.exe` / `MrRSS-{version}-windows-arm64-installer.exe`
- **macOS:** `MrRSS-{version}-darwin-universal.dmg`
- **Linux:** `MrRSS-{version}-linux-amd64.AppImage` / `MrRSS-{version}-linux-arm64.AppImage`

#### Option 2: Build from Source

### Prerequisites

Before you begin, ensure you have the following installed:

- [Go](https://go.dev/) (1.21 or higher)
- [Node.js](https://nodejs.org/) (LTS version with npm)
- [Wails](https://wails.io/docs/gettingstarted/installation) CLI

**Linux only:** You also need to install system dependencies:

```bash
# Ubuntu/Debian
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev gcc pkg-config

# For older Ubuntu versions (before 24.04), use:
# sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev gcc pkg-config
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
   wails build
   ```

   The executable will be created in the `build/bin` directory.

4. **Run the application**

   - Windows: `build/bin/MrRSS.exe`
   - macOS: `build/bin/MrRSS.app`
   - Linux: `build/bin/MrRSS`

### Data Storage

MrRSS stores all user data (database, settings) in platform-specific directories:

- **Windows:** `%APPDATA%\MrRSS\` (e.g., `C:\Users\YourName\AppData\Roaming\MrRSS\`)
- **macOS:** `~/Library/Application Support/MrRSS/`
- **Linux:** `~/.local/share/MrRSS/`

This ensures your data persists across application updates and reinstalls.

## 🛠️ Development

### Running in Development Mode

Start the application with hot reloading:

```bash
wails dev
```

### Running Tests

```bash
# Backend tests
go test ./...

# Frontend tests
cd frontend
npm test
```

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

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

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

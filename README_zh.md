<div align="center">
  <img src="imgs/logo.svg" alt="MrRSS Logo" width="120" height="120">
  <h1>MrRSS</h1>
  <p><strong>一个现代化的跨平台桌面 RSS 阅读器</strong></p>
  
   [![Version](https://img.shields.io/badge/version-1.2.1-blue.svg)](https://github.com/WCY-dt/MrRSS/releases)
  [![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
  [![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
  [![Wails](https://img.shields.io/badge/Wails-v2-red)](https://wails.io/)
  [![Vue.js](https://img.shields.io/badge/Vue.js-3.x-4FC08D?logo=vue.js)](https://vuejs.org/)

  <p>
    <a href="README.md">English</a> | <strong>简体中文</strong>
  </p>
</div>

---

## ✨ 功能特性

- 🖥️ **跨平台**: 适用于 Windows、macOS 和 Linux 的原生桌面应用
- 🌐 **自动翻译**: 使用 Google Translate 或 DeepL API 翻译文章标题
- 🔍 **智能订阅源发现**: 自动从友链和相关来源发现新的订阅源
- ⭐ **收藏与阅读追踪**: 保存文章并追踪已读/未读状态
- 🎨 **现代 UI**: 简洁、响应式的界面，支持深色模式
- 📦 **OPML 导入/导出**: 轻松从其他 RSS 阅读器迁移

## 📸 截图

![Screenshot](imgs/og.png)

## 🚀 快速开始

### 下载与安装

#### 选项 1: 下载预构建安装包（推荐）

从 [Releases](https://github.com/WCY-dt/MrRSS/releases) 页面下载适合您平台的最新安装包：

- **Windows:** `MrRSS-{version}-windows-amd64-installer.exe` / `MrRSS-{version}-windows-arm64-installer.exe`
- **macOS:** `MrRSS-{version}-darwin-universal.dmg`
- **Linux:** `MrRSS-{version}-linux-amd64.AppImage` / `MrRSS-{version}-linux-arm64.AppImage`

#### 选项 2: 源码构建

### 前置要求

在开始之前，请确保已安装以下环境：

- [Go](https://go.dev/) (1.21 或更高版本)
- [Node.js](https://nodejs.org/) (LTS 版本，带 npm)
- [Wails](https://wails.io/docs/gettingstarted/installation) CLI

**仅限 Linux:** 还需要安装系统依赖：

```bash
# Ubuntu/Debian
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev gcc pkg-config

# 对于旧版 Ubuntu（24.04 之前），请使用：
# sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev gcc pkg-config
```

### 安装步骤

1. **克隆仓库**

   ```bash
   git clone https://github.com/WCY-dt/MrRSS.git
   cd MrRSS
   ```

2. **安装前端依赖**

   ```bash
   cd frontend
   npm install
   cd ..
   ```

3. **构建应用**

   ```bash
   wails build
   ```

   可执行文件将在 `build/bin` 目录下生成。

4. **运行应用**

   - Windows: `build/bin/MrRSS.exe`
   - macOS: `build/bin/MrRSS.app`
   - Linux: `build/bin/MrRSS`

### 数据存储

MrRSS 将所有用户数据（数据库、设置）存储在特定平台的目录中：

- **Windows:** `%APPDATA%\MrRSS\` (例如 `C:\Users\YourName\AppData\Roaming\MrRSS\`)
- **macOS:** `~/Library/Application Support/MrRSS/`
- **Linux:** `~/.local/share/MrRSS/`

这确保了您的数据在应用更新和重新安装时得以保留。

## 🛠️ 开发指南

### 开发模式运行

启动带有热重载的应用：

```bash
wails dev
```

### 运行测试

```bash
# 后端测试
go test ./...

# 前端测试
cd frontend
npm test
```

## 🤝 贡献

我们欢迎贡献！详情请参阅我们的 [贡献指南](CONTRIBUTING.md)。

在贡献之前：

1. 阅读 [行为准则](CODE_OF_CONDUCT.md)
2. 检查现有 issue 或创建一个新 issue
3. Fork 仓库并创建功能分支
4. 进行更改并添加测试
5. 提交 Pull Request

## 🔒 安全

如果您发现安全漏洞，请遵循我们的 [安全策略](SECURITY.md)。

## 📝 许可证

本项目采用 MIT 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- 基于 [Wails](https://wails.io/) 构建 - Go + Web 框架
- UI 由 [Vue.js 3](https://vuejs.org/) 和 [Tailwind CSS](https://tailwindcss.com/) 驱动
- 图标来自 [Phosphor Icons](https://phosphoricons.com/)
- RSS 解析使用 [gofeed](https://github.com/mmcdole/gofeed)

## 📮 联系与支持

- **Issues**: [GitHub Issues](https://github.com/WCY-dt/MrRSS/issues)
- **仓库**: [github.com/WCY-dt/MrRSS](https://github.com/WCY-dt/MrRSS)

---

<div align="center">
  <p>Made with ❤️ by the MrRSS Team</p>
  <p>⭐ 如果您觉得这个项目有用，请在 GitHub 上给我们点星！</p>
</div>

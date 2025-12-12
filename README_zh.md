<div align="center">
  <img src="imgs/logo.svg" alt="MrRSS Logo" width="120" height="120">
  <h1>MrRSS</h1>
  <p><strong>一个现代化的跨平台桌面 RSS 阅读器</strong></p>

   [![Version](https://img.shields.io/badge/version-1.2.14-blue.svg)](https://github.com/WCY-dt/MrRSS/releases)
  [![License](https://img.shields.io/badge/license-GPLv3-green.svg)](LICENSE)
   [![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev/)
  [![Wails](https://img.shields.io/badge/Wails-v2-red)](https://wails.io/)
  [![Vue.js](https://img.shields.io/badge/Vue.js-3.x-4FC08D?logo=vue.js)](https://vuejs.org/)

  <p>
    <a href="README.md">English</a> | <strong>简体中文</strong>
  </p>
</div>

---

## ✨ 功能特性

- 🖥️ **跨平台**: 适用于 Windows、macOS 和 Linux 的原生桌面应用
- 🌐 **自动翻译**: 使用翻译 API 或 AI 服务翻译文章标题和内容
- 🔍 **智能订阅源发现**: 自动从友链和相关来源发现新的订阅源
- ⭐ **收藏与阅读追踪**: 保存文章并追踪已读/未读状态
- 📝 **文章摘要**: 使用本地 TF-IDF/TextRank 算法或 AI 生成文章摘要
- 🎨 **现代 UI**: 简洁、响应式的界面，支持深色模式
- 📦 **OPML 导入/导出**: 轻松从其他 RSS 阅读器迁移
- ⌨️ **快捷键**: 通过可自定义的键盘快捷键提升效率，快速导航和操作
- 📋 **智能过滤规则**: 使用强大的自定义规则自动组织订阅源
- 🏭 **自定义自动化脚本**: 支持运行用户定义的自动化脚本以获取订阅源

## 📸 截图

![Screenshot](imgs/og.png)

## 🚀 快速开始

### 下载与安装

#### 选项 1: 下载预构建安装包（推荐）

从 [Releases](https://github.com/WCY-dt/MrRSS/releases/latest) 页面下载适合您平台的最新安装包：

**标准安装版：**

- **Windows:** `MrRSS-{version}-windows-amd64-installer.exe` / `MrRSS-{version}-windows-arm64-installer.exe`
- **macOS:** `MrRSS-{version}-darwin-universal.dmg`
- **Linux:** `MrRSS-{version}-linux-amd64.AppImage` / `MrRSS-{version}-linux-arm64.AppImage`

**便携版**（无需安装，所有数据在一个文件夹内）：

- **Windows:** `MrRSS-{version}-windows-{arch}-portable.zip`
- **Linux:** `MrRSS-{version}-linux-{arch}-portable.tar.gz`
- **macOS:** `MrRSS-{version}-darwin-{arch}-portable.zip`

#### 选项 2: 源码构建

<details>

<summary>点击展开源码构建指南</summary>

<div markdown="1">

##### 前置要求

在开始之前，请确保已安装以下环境：

- [Go](https://go.dev/) (1.24 或更高版本)
- [Node.js](https://nodejs.org/) (LTS 版本，带 npm)
- [Wails](https://wails.io/docs/gettingstarted/installation) CLI

**平台特定要求：**

- **Linux**: 需要系统依赖（GTK、WebKit、AppIndicator）
- **Windows**: 需要 MinGW（用于 CGO/systray 支持）
- **macOS**: 需要 Xcode 命令行工具

详细安装说明请参见[构建要求](docs/BUILD_REQUIREMENTS.md)

```bash
# Linux 快速设置（Ubuntu/Debian）：
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev libayatana-appindicator3-dev gcc pkg-config

# 对于旧版 Ubuntu（24.04 之前）：
# sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev libappindicator3-dev gcc pkg-config
```

##### 安装步骤

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
   # 使用 Makefile（推荐）
   make build

   # 或手动使用 wails（需要 -skipbindings 标志）
   wails build -skipbindings
   ```

   可执行文件将在 `build/bin` 目录下生成。

4. **运行应用**

   - Windows: `build/bin/MrRSS.exe`
   - macOS: `build/bin/MrRSS.app`
   - Linux: `build/bin/MrRSS`

</div>

</details>

### 数据存储

**正常模式**（默认）：

- **Windows:** `%APPDATA%\MrRSS\` (例如 `C:\Users\YourName\AppData\Roaming\MrRSS\`)
- **macOS:** `~/Library/Application Support/MrRSS/`
- **Linux:** `~/.local/share/MrRSS/`

**便携模式**（当 `portable.txt` 文件存在时）：

- 所有数据存储在 `data/` 文件夹中

这确保了您的数据在应用更新和重新安装时得以保留。

## 🛠️ 开发指南

<details>

<summary>点击展开开发指南</summary>

<div markdown="1">

### 开发模式运行

启动带有热重载的应用：

```bash
wails dev
```

> **⚠️ macOS 开发者注意**：在 macOS 上构建时，请始终使用 `make dev` 或添加 `-tags nosystray` 标签以避免 AppDelegate 冲突。详见[构建要求](docs/BUILD_REQUIREMENTS.md#macos)。

### 代码质量工具

#### 使用 Make

我们提供了 `Makefile` 来处理常见的开发任务（在 Linux/macOS/Windows 上都可用）：

```bash
# 显示所有可用命令
make help

# 运行完整检查（lint + 测试 + 构建）
make check

# 清理构建产物
make clean

# 设置开发环境
make setup
```

#### 使用脚本

`scripts/` 目录中提供了跨平台的自动化脚本：

**Linux/macOS:**

```bash
# 运行所有质量检查
./scripts/check.sh

# 发布前检查
./scripts/pre-release.sh
```

**Windows (PowerShell):**

```powershell
# 运行所有质量检查
.\scripts\check.ps1

# 发布前检查
.\scripts\pre-release.ps1
```

### Pre-commit Hooks

本项目使用 pre-commit hooks 来确保代码质量：

```bash
# 安装 hooks
pre-commit install

# 在所有文件上运行
pre-commit run --all-files
```

### 运行测试

```bash
# 后端测试
go test ./...

# 前端测试
cd frontend
npm test
```

</div>

</details>

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

本项目采用 GPLv3 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件。

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

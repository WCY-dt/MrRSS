# Changelog

All notable changes to MrRSS will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.3.8] - 2025-12-24

### Added

- Supported AI setting tests in settings page to verify connectivity and credentials. (#297)
- Supported fetching feeds which require JavaScript rendering using headless browser. (#298)

### Changed

- AI generated summaries are now stored in the database to avoid redundant requests and improve performance. (#295)
- Reduce the frequency of automatic record of window status to improve performance.
- Improved the conversion from HTML to Markdown when exporting articles to Obsidian. (#299)

### Fixed

- Fixed the issue where docker image failed to access local files due to permission issues. (#296)
- Fixed the issue where articles failed to open in default browser. (#294)
- Fixed the issue where importing and exporting OPML files did not work correctly. (#271)
- Fixed the issue of CSP blocking some external resources.

## [1.3.7] - 2025-12-23

### Added

- Supported server mode for self-hosted web application deployment. (#267) (@caoli5288)
- Supported drag-and-drop to reorder feeds or change feed categories. (#288)
- Supported AI Chat on article content. And of course **it's disabled by default**! (#286)
- Supported exporting articles to Obsidian. (#289)
- Supported extracting full article content from original website when RSS feed only provides summary. (#266)

### Changed

- AI summarization is now triggered manually on default to avoid excessive API usage. Users can enable automatic summarization in settings if desired. (#287)
- Added Plugin setting tab in settings page and moved FreshRSS synchronization settings there.
- Improved icons and translations for better user experience.
- Enhanced the conversion from HTML to Markdown when exporting articles to Obsidian. (#299)

### Fixed

- Fixed the issue where concurrent feed refreshes exceed network capacity limit. (#262)

## [1.3.6] - 2025-12-22

### Added

- Supported importing feeds with HTML+XPath / XML+XPath type from OPML files. (#264)
- Supported FreshRSS synchronization. (#245)

### Changed

- Improved error display for customized scripts when adding/editing feeds. (#264)
- Network connection test now supports proxy settings. (#256)

### Fixed

- Fixed the issue where different articles display the same content due to incorrect URL matching. (#257)
- Fixed the issue where import and export of OPML files did not work correctly on macOS. (#263)
- Fixed the issue where localhost cannot be processed correctly. (#257)

### Removed

- Removed single instance lock on Linux platform to avoid D-Bus related issues. (#246)

## [1.3.5] - 2025-12-20

**BREAKING**: AI-based summarization and translation now need a full path instead of just endpoint URL.

> e.g. for OpenAI services, use `https://api.openai.com/v1/chat/completions`. for Ollama, use `http://localhost:11434/api/generate`.

### Added

- Supported ollama and other local LLMs for AI-based translation and summarization. (#251)
- Supported limits and quotas for AI services to control usage and costs. (#252)
- Supported hover to mark articles as read in article list. (#250)
- Supported deelpx translation service. (#247)

### Changed

- Improved AI settings UI/UX for better user experience.
- Refactored docs and workflows to improve maintainability and clarity.
- AI translation and summarization are now cached to reduce redundant requests and improve performance.
- Recent articles are now cached to improve loading speed.
- When AI functionality gets errors, fallback to local summarization/translation automatically.

### Fixed

- Fixed the issue where some opml files cannot be imported and outported correctly. (#249)
- Fixed the issue where proxy settings were not applied correctly for feed fetching. (#256)
- Fixed the issue where software print too much debug logs in production builds.
- Fixed the issue where network connection test fails when some test endpoints are unreachable. (#256)
- Fixed the issue where summarization failures will affect article content rendering. (#242)
- Fixed the issue where article content fetching blocked by feed refreshes.
- Fixed the issue of dark mode styles on Linux platform.

## [1.3.4] - 2025-12-18

### Fixed

- Fixed the issue where window title bar buttons on MacOS overlapping with content area.
- Fixed the issue where window cannot be dragged on MacOS. (#242)

## [1.3.3] - 2025-12-18

### Added

- Supported copying article link and title to clipboard from article actions menu. (#155)

### Changed

- Replace the following functionality with a native implementation using wails3 (#242)
  - Open link in default browser
  - Window events handling (minimize, maximize, close) and management
  - Native window context menu and title bar on MacOS

### Fixed

- Fixed the issue where super loooooooong article titles causing layout breaking in article list.
- Fixed the issue where cutting long article titles in chinese does not work correctly.

## [1.3.2] - 2025-12-17

### Fixed

- Fixed the issue where MacOS window cannot be closed correctly after maximizing. (#221)
- Fixed the issue where images in article content rendering mode cannot be displayed correctly. (#222)
- Fixed the issue where windows app cannot be packaged correctly due to wrong version number format.

## [1.3.0] - 2025-12-17

### Changed

- **BREAKING**: Upgraded from Wails v2 to Wails v3 (alpha) framework (#234)
  - Migrated to new API
  - Replaced external systray library with Wails v3 built-in system tray
  - Updated single instance handling to use v3 API
  - Updated event handling to use v3 hooks
  - Updated build system to use Taskfile and Wails v3 CLI
  - Updated dependencies to work with WebKit2GTK 4.1 and libsoup 3.0
- Changed GitHub Actions workflows compatibility with Wails v3

## [1.2.20] - 2025-12-16

### Changed

- Added more tests for backend and frontend components to improve code reliability.
- Updated dependencies to latest versions for security and performance improvements.

### Fixed

- Fixed issues related to MacOS platform (#212)
  - Updated icons for better appearance.
  - Added more white space on top of the main window for better visual balance.
  - Disabled icon name on tray.
  - Fixed the issue where window cannot be dragged.
  - Fixed the issue where application not closing correctly after maximizing.

## [1.2.19] - 2025-12-15

### Fixed

- Fixed the issue where some settings were not saved and applied correctly. (#201)
- Fixed the issue where macOS application failing to launch after installation.

## [1.2.18] - 2025-12-14

### Added

- Supported image gallery for browsing all images in articles. (#190)
- Supported network latency and bandwidth testing in settings. (#194)

### Changed

- Added number of concurrent feed refreshes according to network situation.

### Fixed

- Fixed the issue where software can open multiple instances. (#198)
- Fixed the issue where number of feeds left to refresh is not accurately displayed during feed refresh. (#194)

## [1.2.17] - 2025-12-13

### Added

- Supported upgrade in portable mode. (#191)

### Fixed

- Fixed the issue where settings cannot be saved and applied by downgrade TailwindCSS version.

## [1.2.16] - 2025-12-13

### Added

- Add toggle button to hide/show article content translations. (#186)

### Changed

- Updated all dependencies to latest versions for security and performance improvements.

### Fixed

- Fixed the issue where MacOS cannot complile correctly for system tray support. (#181)
- Fixed the issue where Linux-ARM64 AppImage cannot run correctly.

## [1.2.15] - 2025-12-13

### Changed

- Supported alpha, beta, and pre-release version tags. (#182)
- Enhanced credential encryption mechanism to improve security during database migration and storage. (#160)

## [1.2.14] - 2025-12-12

### Added

- Supported portable mode for running MrRSS from USB drives with all data stored in a single folder. (#167)
- Supported minimizing to system tray on close action.
- Supported hiding preview images in article list for a more compact view. (#157)

### Fixed

- Fixed the issue where some images wrapped in links cannot be operated correctly.
- Fixed the issue where single-line link cannot be translated correctly.
- Fixed the issue where some links cannot be opened in the default browser.
- Fixed the issue where icons on MacOS were not displayed correctly. (#173)
- Fixed the issue where the window size and position were not restored correctly. (#173)

## [1.2.13] - 2025-12-11

### Added

- Supported media cache system to bypass anti-hotlinking restrictions and cache images/videos locally. (#152)
- Supported proxy settings for network requests.
- Supported intelligent refresh scheduling based on feed update frequency. (#151)
- Supported customizing proxy and refresh settings per feed. (#151)
- Supported read all articles for a specific feed or category. (#156)

### Changed

- Google Translate endpoint is now customizable in settings. (#158)

### Fixed

- Fixed the issue where title and summary cannot be selected and copied in article content rendering mode. (#155)
- Fixed the issue where some articles are rendered with incorrect formatting in article content rendering mode.

## [1.2.12] - 2025-12-10

### Changed

- Settings now support validation and show error messages for invalid inputs. (#147)

### Fixed

- Links in article content rendering mode can now be translated correctly. (#148)
- Fixed the issue where some images were not displayed in article content rendering mode. (#148)

## [1.2.11] - 2025-12-08

### Added

- Supported selecting existing categories when adding or editing a new feed.
- When playing audio or video in article content rendering mode, playback controls are now available.
- Supported customizing the AI prompt for article summarization and translation.

### Changed

- Improved styles for article content rendering mode.

### Fixed

- Fixed the issue where some feeds cannot be handled due to invalid styles in RSS XML.
- Fixed the issue where inline elements (e.g. code, formulas) were not handled correctly in translation.
- Fixed the issue where toast notifications not supporting dark mode caused visibility problems.
- Fixed the issue related to importing OPML files.

## [1.2.10] - 2025-12-07

### Added

- Supported audio and video embedding in article content rendering mode.

### Changed

- Enhanced styling of article content for better readability.

## [1.2.9] - 2025-12-05

### Added

- Supported Baidu Translation and AI-based translation.
- Supported AI-based article summarization using OpenAI-compatible APIs.

### Changed

- Errors from translation services are now logged and displayed to users for better troubleshooting.

### Fixed

- Fixed the issue where the default settings were not applied correctly on first launch.
- Fixed the issue where `PubMed` feed parsing failed.

## [1.2.8] - 2025-12-04

### Added

- Implemented Read Later functionality, articles marked for read later are automatically set to unread.

### Changed

- Last update time now displayed as inline sub-text instead of separate row.
- Added toggle filter shortcut (default: `f`).
- Nav icons use fill style when active for stronger visual feedback.
- Category headers are now sticky for scroll context retention.
- Feed refresh now skipped on startup if last article update interval is within set threshold.
- After each article refresh completes, the app now checks for updates. If a new version is detected, it automatically downloads and installs in the background.
- Changed some default settings.

### Fixed

- Fixed styling issues, including incorrect icon colors in dark mode, inconsistent font sizes, and misaligned elements.

## [1.2.7] - 2025-12-03

### Added

- Supported hiding feeds from timeline.

### Fixed

- Fixed initialization problem by adding progress tracking for single feed and OPML import.

## [1.2.6] - 2025-12-02

### Added

- Added TF-IDF and TextRank algorithms for generating article summaries.
- Added auto-translation support for summary, title, and content in rendering.
- Enhanced multimedia support in content rendering mode.

### Changed

- Improved image viewer with drag support and better zooming.
- Refactored both frontend and backend code for better maintainability.

### Fixed

- Fixed the issue where searching box scrolls with the feed list.

## [1.2.5] - 2025-11-27

### Added

- Supported for user-defined scripts to fetch and parse non-standard RSS feeds.
- Improved shortcuts for popup window actions.
- Supported sorting articles list in settings by various criteria.
- Supported for refreshing individual feeds via right-click context menu.
- Supported for searching feeds in the feed list.

### Changed

- Article list will not refresh during feed refresh, fixing a bug causing the article list to occasionally crash.
- Generate article titles from content when RSS feed items are missing titles.

### Fixed

- Fixed issue where some UI elements did not scale properly.
- Fixed bug causing view mode performe incorrectly when switching articles rapidly.

### Removed

- Removed search box for article list because the filter function covers the same use case.

## [1.2.4] - 2025-11-27

### Changed

- Refactored frontend codebase and landing page for better maintainability and user experience.
- Added tests for critical components to improve code reliability.
- Updated dependencies to latest versions for security and performance improvements.
- Better documentation for developers and contributors.
- Improved CI/CD pipeline for faster and more reliable builds and deployments.

## [1.2.2] - 2025-11-26

### Added

- Added keyboard shortcuts for common actions and corresponding settings in the Settings tab.
- Supported customizing rules with "If [condition], then [action]" format for advanced users.

### Changed

- Improved landing page UI/UX for better user engagement.
- Improved documentation for new users.

## [1.2.1] - 2025-11-26

### Added

- Adds advanced article filtering via a modal accessible from a filter button next to the search box.

## [1.2.0] - 2025-11-25

### Added

- Implements automated feed discovery from friend links with intelligent batch scanning, comprehensive deduplication, real-time progress tracking

### Changed

- Major restructuring of backend code for improved maintainability

## [1.1.8] - 2025-11-24

### Added

- Feed icons now display in the Settings > Feeds tab feed list for better visual identification
- Website homepage link stored for each feed (from RSS feed metadata)

### Changed

- "Open Website" context menu action now opens the feed's website homepage (if available) instead of RSS feed URL, with fallback to RSS URL
- All hardcoded strings now properly use i18n translations for better internationalization support

### Fixed

- Replaced native `prompt()` with custom `showInput()` dialog for consistent UI

## [1.1.7] - 2025-11-24

### Added

- Unread count badge displayed on each feed in the sidebar and "All Articles" button
- "Mark All as Read" button next to the refresh button in article list and feed context menu
- When feeds fail to load, display error message in feed list instead of silently failing
- Implemented input dialog for moving feeds to a new category

### Changed

- Fixed unfavorite icon for better visibility

## [1.1.6] - 2025-11-23

### Added

- "Open Website" option in feed right-click menu
- Startup on boot setting in General settings tab (default off)

### Changed

- "Hide Article" context menu item now shows as a danger button
- Improved settings tab switching style with hover effects and animations
- Fixed unfavorite icon visibility in article detail view

### Fixed

- Fixed software update installation process - updates can now be properly installed

## [1.1.5] - 2025-11-23

### Added

- Switch between viewing the original webpage and RSS content within the app
- Article hiding functionality
- Last article update time display in settings

### Changed

- Improved UI text and image selection prevention

### Fixed

- Fixed issue where translation settings changes didn't clear existing translations

## [1.1.4] - 2025-11-23

### Added

- Auto cleanup sub-settings:
  - Max cache size setting (default 20MB) - controls maximum database size before cleanup
  - Max article age setting (default 30 days) - automatically delete articles older than specified days (except favorites)
- Download progress bar during update download
- App automatically closes after starting installer to prevent conflicts during update
- Automatic cleanup of installation packages after update installation

### Changed

- Settings now auto-save immediately when changed (no need to click save button)

### Removed

- "Save Settings" button at bottom of settings page (replaced with auto-save)

## [1.1.3] - 2025-11-22

### Added

- Automatically detects user's operating system and CPU architecture and downloads appropriate installer from GitHub releases. Then launches installer and prepares for update
- Multi-Platform Support:
  - Windows: x64 (amd64), ARM64
  - Linux: x64 (amd64), ARM64 (aarch64)
  - macOS: Universal (Intel & Apple Silicon)
- Visual feedback during update download and installation

## [1.1.2] - 2025-11-22

### Added

- Initial release preparation
- OPML import/export functionality
- Feed category organization
- Automatically detect and apply system theme preference
- Better defaults for translation settings
- Version check functionality in Settings → About tab

### Changed

- Simplified update check UI
- Improved theme switching mechanism
- Better handling of translation provider selection

### Fixed

- Various bug fixes and stability improvements
- UI refinements for better user experience
- Theme switching issues between light and dark modes
- Translation default language selection
- Update notification display

## [1.1.0] - 2025-11-20

### Added

- **Initial Public Release** of MrRSS
- **Cross-Platform Support**: Native desktop app for Windows, macOS, and Linux
- **RSS Feed Management**: Add, edit, and delete RSS feeds
- **Article Reading**: Clean, distraction-free reading interface
- **Smart Organization**: Organize feeds into categories
- **Favorites & Reading Tracking**: Save articles and track read/unread status
- **Modern UI**: Clean, responsive interface with dark mode support
- **Auto-Translation**: Translate article titles using translation services or AI-based translation
- **OPML Support**: Import and export feed subscriptions
- **Auto-Update**: Configurable interval for fetching new articles
- **Database Cleanup**: Automatic removal of old articles
- **Multi-Language Support**: English and Chinese interface
- **Theme Support**: Light, dark, and auto (system) themes

---

## Release Notes

### Version Numbering

MrRSS follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

### Download

Downloads for all platforms are available on the [GitHub Releases](https://github.com/WCY-dt/MrRSS/releases) page.

### Upgrade Notes

When upgrading from a previous version:

1. Your data (feeds, articles, settings) is preserved in platform-specific directories
2. Database migrations are applied automatically on first launch
3. For major version upgrades, please review the changelog for breaking changes

### Support

- Report bugs: [GitHub Issues](https://github.com/WCY-dt/MrRSS/issues)
- Feature requests: [GitHub Issues](https://github.com/WCY-dt/MrRSS/issues)
- Documentation: [README](README.md)

# Changelog

All notable changes to MrRSS will be documented in this file.


The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
- **Auto-Translation**: Translate article titles using Google Translate or DeepL API
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

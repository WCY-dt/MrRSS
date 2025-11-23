# AI Agent Guidelines for MrRSS

This document provides comprehensive guidance for AI agents (like GitHub Copilot, ChatGPT, Claude, etc.) working on the MrRSS project.

## Project Overview

**MrRSS** is a modern, privacy-focused, cross-platform desktop RSS reader built with:

- **Backend**: Go 1.21+ with Wails v2 framework
- **Frontend**: Vue.js 3 (Composition API) with Tailwind CSS
- **Database**: SQLite with `modernc.org/sqlite` driver
- **Build Tool**: Wails CLI v2.11+
- **Current Version**: 1.1.4

### Core Functionality

- **Feed Management**: RSS/Atom feed subscription, parsing with `gofeed`, and real-time updates
- **Article Management**: Read/unread tracking, favorites, pagination, and filtering
- **Organization**: Category-based feed organization with expandable categories
- **Translation**: Auto-translation using Google Translate (free) or DeepL API
- **Data Portability**: OPML import/export for easy migration
- **Internationalization**: Full UI support for English and Chinese
- **Auto-Refresh**: Configurable interval for automatic feed updates (default 10 minutes)
- **Auto-Cleanup**: Configurable article retention (by age and cache size)
- **Update System**: In-app update checking and installation with progress tracking
- **Theming**: Light/Dark/Auto modes with system preference detection

## Project Structure

```plaintext
MrRSS/
├── main.go                      # Application entry point, Wails app initialization
├── wails.json                   # Wails configuration, version info (2 version fields)
├── go.mod / go.sum              # Go dependencies
├── internal/                    # Backend Go code (not exposed directly)
│   ├── database/
│   │   ├── sqlite.go           # SQLite operations, schema, settings management
│   │   └── sqlite_test.go      # Database tests
│   ├── feed/
│   │   ├── fetcher.go          # RSS/Atom parsing with gofeed, concurrent fetching
│   │   └── fetcher_test.go     # Feed parsing tests
│   ├── handlers/
│   │   └── handlers.go         # HTTP handlers, app logic, update system
│   ├── models/
│   │   └── models.go           # Data structures (Feed, Article, etc.)
│   ├── opml/
│   │   ├── handler.go          # OPML import/export logic
│   │   └── handler_test.go     # OPML tests
│   ├── translation/
│   │   ├── translator.go       # Translation interface
│   │   ├── google_free.go      # Google Translate (free, no API key)
│   │   ├── deepl.go            # DeepL API integration
│   │   └── translator_test.go  # Translation tests
│   ├── utils/
│   │   └── paths.go            # Platform-specific paths for data storage
│   └── version/
│       └── version.go          # Version constant (update this!)
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── ArticleList.vue     # Article list with virtual scrolling
│   │   │   ├── ArticleDetail.vue   # Article reader view
│   │   │   ├── Sidebar.vue         # Feed/category navigation
│   │   │   ├── ContextMenu.vue     # Right-click context menu
│   │   │   ├── Toast.vue           # Toast notifications
│   │   │   └── modals/             # Modal dialogs
│   │   │       ├── SettingsModal.vue       # Main settings container
│   │   │       ├── settings/
│   │   │       │   ├── GeneralTab.vue     # General settings with auto-save
│   │   │       │   ├── FeedsTab.vue       # Feed management
│   │   │       │   └── AboutTab.vue       # About, version, updates
│   │   │       ├── AddFeedModal.vue       # Add new feed
│   │   │       ├── EditFeedModal.vue      # Edit feed
│   │   │       └── ConfirmDialog.vue      # Confirmation dialogs
│   │   ├── store.js            # Global state management (reactive)
│   │   ├── i18n.js             # Internationalization (en/zh)
│   │   ├── App.vue             # Root component
│   │   ├── main.js             # Vue app initialization
│   │   └── style.css           # Global styles, theme variables
│   ├── package.json            # Frontend dependencies, version
│   ├── vite.config.js          # Vite build configuration
│   ├── tailwind.config.js      # Tailwind CSS configuration
│   └── wailsjs/                # Auto-generated Go→JS bindings (don't edit)
├── test/
│   └── testdata/               # Test files (OPML samples, etc.)
├── build/                       # Build scripts and installers
│   ├── windows/                # Windows-specific build files
│   ├── linux/                  # Linux AppImage scripts
│   └── macos/                  # macOS DMG creation scripts
├── website/                     # GitHub Pages website source
├── CHANGELOG.md                 # Version history (update this!)
├── README.md                    # English documentation (version badge)
├── README_zh.md                 # Chinese documentation (version badge)
└── AGENTS.md                    # This file
```

## Key Technologies & Patterns

### Backend (Go)

**Wails Framework**:

- Use `context.Context` for all exported methods
- Methods are automatically exposed to frontend
- Use struct methods for organization

**Database**:

- SQLite with `modernc.org/sqlite` driver (pure Go implementation)
- WAL mode enabled for better concurrency
- Optimized with indexes on common query patterns
- Use prepared statements for all queries
- Settings stored in `settings` table with key-value pairs
- No migrations (development phase) - users backup data before updates

**Settings System**:

```go
// Settings are stored as strings in database
db.GetSetting("key")           // Returns string value
db.SetSetting("key", "value")  // Stores string value

// Common settings:
// - update_interval: minutes between auto-refresh (default: "10")
// - auto_cleanup_enabled: "true" or "false" (default: "false")
// - max_cache_size_mb: max DB size in MB (default: "20")
// - max_article_age_days: days to keep articles (default: "30")
// - translation_enabled: "true" or "false"
// - target_language: language code (e.g., "zh", "en")
// - translation_provider: "google" or "deepl"
// - language: UI language "en" or "zh"
// - theme: "light", "dark", or "auto"
```

**RSS Parsing**:

- Use `github.com/mmcdole/gofeed` for parsing
- Support both RSS and Atom formats
- Handle malformed feeds gracefully

**Translation**:

- Google Translate: Free, no API key
- DeepL: Requires API key
- Store translations in database

### Frontend (Vue.js)

**Vue 3 Composition API**:

```vue
<script setup>
import { ref, computed, onMounted } from 'vue';
import { store } from '../store.js';

const items = ref([]);
const filteredItems = computed(() => /* ... */);

onMounted(async () => {
  // Initialize
});
</script>
```

**State Management**:

- Use `store.js` for global state
- Reactive with Vue's `reactive()`
- Key state properties:
  - `feeds`: Array of feed objects
  - `articles`: Array of article objects
  - `selectedFeed`: Currently selected feed ID
  - `selectedArticle`: Currently selected article
  - `theme`: Current theme ("light", "dark", "auto")
  - `i18n`: Internationalization object with `t()` method

**Auto-Save Pattern** (Settings):

```vue
<script setup>
import { watch, onUnmounted } from 'vue';

let saveTimeout = null;

// Debounced save function (500ms delay)
function debouncedAutoSave() {
    if (saveTimeout) clearTimeout(saveTimeout);
    saveTimeout = setTimeout(autoSave, 500);
}

// Watch entire settings object
watch(() => props.settings, debouncedAutoSave, { deep: true });

// Cleanup on unmount
onUnmounted(() => {
    if (saveTimeout) clearTimeout(saveTimeout);
});
</script>
```

**Styling**:

- Tailwind CSS utility classes
- Theme variables in CSS
- Dark mode support via `data-theme`

**Internationalization**:

- `i18n.js` provides translation function
- Use `store.i18n.t('key')` in templates
- Support English (`en`) and Chinese (`zh`)

## Development Guidelines

### Code Style

**Go**:

- Follow `gofmt` formatting
- Use meaningful variable names
- Handle errors explicitly
- Add comments for exported functions
- Keep functions focused and small

**Vue.js**:

- Use `<script setup>` syntax
- Props validation with `defineProps`
- Emit declarations with `defineEmits`
- Composition API over Options API
- Keep components under 300 lines

**Commit Messages**:

Follow Conventional Commits:

```plaintext
<type>(<scope>): <description>

[optional body]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### Testing

**Backend**:

```bash
go test ./...
go test -cover ./internal/database
```

**Frontend**:

```bash
cd frontend
npm test
npm run lint
```

**Manual Testing**:

```bash
wails dev  # Development mode with hot reload
wails build  # Production build
```

### Building

**Development**:

```bash
wails dev
```

**Production**:

```bash
wails build -clean -ldflags "-s -w"
```

**Platform-Specific**:

- Windows: Built for amd64 and arm64
- macOS: Universal binary (Intel + Apple Silicon)
- Linux: Built for amd64

## Common Tasks

### Adding a New Feature

1. **Plan**: Define the feature scope and requirements
2. **Backend** (if needed):
   - Add methods to handlers
   - Update database schema if needed
   - Add tests
3. **Frontend**:
   - Create/update components
   - Add i18n strings
   - Update store if needed
4. **Test**: Manual + automated testing
5. **Document**: Update README if user-facing

### Adding Translations

1. Edit `frontend/src/i18n.js`
2. Add keys to both `en` and `zh` sections:
   ```javascript
   export const translations = {
       en: {
           newKey: 'English text',
           // ...
       },
       zh: {
           newKey: '中文文本',
           // ...
       }
   };
   ```
3. Use `store.i18n.t('newKey')` in templates
4. Test language switching in UI

### Auto-Cleanup Feature

The auto-cleanup system has configurable settings:

1. **Enable/Disable**: `auto_cleanup_enabled` setting
2. **Max Article Age**: `max_article_age_days` (default: 30 days)
   - Deletes articles older than this threshold
   - **Exception**: Favorited articles are never deleted
3. **Max Cache Size**: `max_cache_size_mb` (default: 20 MB)
   - Reserved for future size-based cleanup
   - Currently tracks DB size but doesn't enforce limit

**Cleanup Logic** (`internal/database/sqlite.go`):
```go
// Articles older than max_article_age_days are deleted
// EXCEPT favorited articles which are kept forever
cutoffDate := time.Now().AddDate(0, 0, -maxAgeDays)
DELETE FROM articles WHERE published_at < ? AND is_favorite = 0
```

**Cleanup Triggers**:
- Manual: User clicks "Clean Database" in Settings → Feeds tab
- Automatic: After each auto-refresh (if enabled in settings)
- On startup: If auto-cleanup enabled

### Update System

The in-app update system checks GitHub releases:

1. **Check Updates**: Queries GitHub API for latest release
2. **Platform Detection**: Identifies OS and architecture
3. **Asset Matching**: Finds appropriate installer:
   - Windows: `MrRSS-{version}-windows-{arch}-installer.exe`
   - Linux: `MrRSS-{version}-linux-{arch}.AppImage`
   - macOS: `MrRSS-{version}-darwin-universal.dmg`
4. **Download**: Shows progress bar, simulated progress updates
5. **Install**: Launches installer, schedules cleanup, exits app
6. **Cleanup**: Installer file deleted 3-5 seconds after launch

**Security Measures**:
- URL validation (must be from official GitHub repo)
- File path validation (prevents traversal attacks)
- Extension validation (platform-specific)
- Safe file cleanup using `os.Remove()` instead of shell commands

### Database Changes

1. Edit `internal/database/sqlite.go`
2. Update `Init()` function with new schema
3. No migrations (development phase)
4. Users should back up data before updates

### UI Changes

1. Use existing Tailwind classes
2. Follow dark mode pattern: `class="bg-bg-primary text-text-primary"`
3. Ensure responsive design
4. Test in both themes
5. Add icons from Phosphor Icons (`ph ph-*`)

## Important Conventions

### Naming

**Go**:

- Exported: `PascalCase` (e.g., `FetchFeed`)
- Unexported: `camelCase` (e.g., `parseXML`)
- Interfaces: Usually noun (e.g., `Reader`, `Handler`)

**Vue**:

- Components: `PascalCase` files (e.g., `ArticleList.vue`)
- Props/emits: `camelCase` (e.g., `feedId`, `onUpdate`)
- CSS classes: `kebab-case` (e.g., `article-card`)

### File Organization

- One component per file
- Co-locate tests with code
- Group related functionality
- Keep files under 500 lines

### Error Handling

**Go**:

```go
if err != nil {
    return fmt.Errorf("context: %w", err)
}
```

**Vue**:

```javascript
try {
    // operation
} catch (e) {
    console.error(e);
    window.showToast(store.i18n.t('error'), 'error');
}
```

## Security Considerations

1. **Input Validation**: Validate all user inputs
2. **SQL Injection**: Use parameterized queries
3. **XSS**: Vue escapes by default, don't use `v-html`
4. **API Keys**: Store locally, never commit
5. **Feed Content**: Displayed in iframe with sandbox

## Performance Tips

1. **Backend**:
   - Use goroutines for concurrent feed fetching
   - Cache feed data when appropriate
   - Optimize database queries

2. **Frontend**:
   - Virtual scrolling for large lists
   - Lazy load images
   - Debounce search inputs
   - Use `v-show` vs `v-if` appropriately

## Debugging

**Backend**:

- Use `fmt.Println` or `log.Printf`
- Check Wails console output
- Use Go debugger (Delve)

**Frontend**:

- Browser DevTools console
- Vue DevTools extension
- Check Network tab for API calls

**Common Issues**:

1. **CORS**: Not applicable (native app)
2. **Feed parsing**: Check feed URL format
3. **Translation**: Verify API key for DeepL
4. **Build errors**: Check Go/Node versions

## Version Management

**IMPORTANT**: When upgrading the software version, you MUST update ALL of the following files:

### Required Version Updates

1. **`internal/version/version.go`**
   ```go
   const Version = "X.Y.Z"  // Update this constant
   ```

2. **`wails.json`** (2 locations in same file)
   ```json
   {
     "version": "X.Y.Z",
     "info": {
       "productVersion": "X.Y.Z",
       ...
     }
   }
   ```

3. **`frontend/package.json`**
   ```json
   {
     "version": "X.Y.Z",
     ...
   }
   ```

4. **`frontend/src/components/modals/settings/AboutTab.vue`**
   ```vue
   const appVersion = ref('X.Y.Z');  // Update default version
   ```

5. **`README.md`** (version badge)
   ```markdown
   [![Version](https://img.shields.io/badge/version-X.Y.Z-blue.svg)](...)
   ```

6. **`README_zh.md`** (version badge)
   ```markdown
   [![Version](https://img.shields.io/badge/version-X.Y.Z-blue.svg)](...)
   ```

7. **`CHANGELOG.md`** (add new version entry at top)
   ```markdown
   ## [X.Y.Z] - YYYY-MM-DD
   
   ### Added
   - New feature descriptions
   
   ### Changed
   - Changes to existing functionality
   
   ### Fixed
   - Bug fixes
   ```

### Version Numbering

Follow **Semantic Versioning** (MAJOR.MINOR.PATCH):

- **MAJOR** (X.0.0): Incompatible API changes, breaking changes
- **MINOR** (1.X.0): Backwards-compatible new features
- **PATCH** (1.1.X): Backwards-compatible bug fixes

### Version Update Checklist

- [ ] Update `internal/version/version.go`
- [ ] Update `wails.json` (both fields)
- [ ] Update `frontend/package.json`
- [ ] Update `AboutTab.vue` default version
- [ ] Update `README.md` badge
- [ ] Update `README_zh.md` badge  
- [ ] Update `CHANGELOG.md` with changes
- [ ] Test that version displays correctly in About tab
- [ ] Run all tests to ensure nothing broke

## Version Management

- Version in `wails.json`: `"version"` and `"productVersion"`
- Version in frontend: `SettingsModal.vue` About tab
- Follow Semantic Versioning (MAJOR.MINOR.PATCH)
- Current: 1.1.0

## Resources

- **Wails Docs**: [https://wails.io/docs/](https://wails.io/docs/)
- **Vue.js Docs**: [https://vuejs.org/](https://vuejs.org/)
- **Tailwind CSS**: [https://tailwindcss.com/](https://tailwindcss.com/)
- **Go Docs**: [https://golang.org/doc/](https://golang.org/doc/)
- **gofeed**: [https://github.com/mmcdole/gofeed](https://github.com/mmcdole/gofeed)

## Getting Help

1. Check existing issues on GitHub
2. Read documentation (README, CONTRIBUTING)
3. Search discussions
4. Create new issue with template

---

**Remember**: When in doubt, follow existing patterns in the codebase. Consistency is key!

# GitHub Copilot Instructions for MrRSS

> **Documentation**: [AGENTS.md](../AGENTS.md) | [Architecture](../docs/ARCHITECTURE.md) | [Code Patterns](../docs/CODE_PATTERNS.md) | [Testing](../docs/TESTING.md) | [Build Requirements](../docs/BUILD_REQUIREMENTS.md)

## Project Context

MrRSS is a modern, privacy-focused, cross-platform desktop RSS reader built with Wails (Go + Vue.js).

**Core Principles**: Privacy-first, cross-platform, modern UI, high performance, accessible

## Tech Stack

- **Backend**: Go 1.24+ with Wails v2.11+ framework, SQLite with `modernc.org/sqlite`, `fyne.io/systray` for system tray (disabled on macOS)
- **Frontend**: Vue 3.5+ Composition API, Pinia, Tailwind CSS 3.3+, Vite 5+
- **Tools**: Wails CLI v2.11+, npm, Go modules
- **Icons**: Phosphor Icons | **I18n**: vue-i18n (English/Chinese)

## Quick Patterns Reference

### Backend (Go)

**Key Principles**:
- Always use `context.Context` for exported methods
- Error wrapping with `fmt.Errorf("operation failed: %w", err)`
- Prepared statements for all database queries
- Proper cleanup with `defer`
- Input validation before processing

📚 **Full Patterns**: See [CODE_PATTERNS.md](../docs/CODE_PATTERNS.md#backend-patterns-go)

### Frontend (Vue 3)

When writing Vue components, follow these patterns:

```vue
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useAppStore } from '@/stores/app';
import { useI18n } from 'vue-i18n';

// Props with proper typing
interface Props {
  item: Article;
  isActive?: boolean;
}
const props = withDefaults(defineProps<Props>(), {
  isActive: false
});

// Store and i18n
const store = useAppStore();
const { t } = useI18n();

// Reactive state
const isLoading = ref(false);

// Async operations with error handling
async function loadData() {
  isLoading.value = true;
  try {
    const data = await fetch('/api/articles').then(r => r.json());
    // Process data...
  } catch (error) {
    console.error('Failed to load:', error);
    window.showToast(t('error'), 'error');
  } finally {
    isLoading.value = false;
  }
}

onMounted(() => loadData());
</script>

<template>
  <div class="component-container">
    <!-- Content with proper i18n -->
  </div>
</template>
```

📚 **Full Patterns**: See [CODE_PATTERNS.md](../docs/CODE_PATTERNS.md#frontend-patterns-vue)

## Internationalization

Always use i18n for user-facing strings:

```vue
<!-- Template -->
<h1>{{ t('welcome') }}</h1>
<button :title="t('clickToOpen')">{{ t('open') }}</button>

<!-- Script -->
window.showToast(t('successMessage'), 'success');
```

## UI Components

### Toast HTML Structure

```html
<div class="bg-bg-primary border border-border rounded-lg p-4">
  <h3 class="text-text-primary font-semibold">{{ t('title') }}</h3>
  <p class="text-text-secondary text-sm">{{ t('description') }}</p>
</div>

<!-- Status Indicators -->
<div class="status-indicator status-unread">{{ t('unread') }}</div>
<div class="status-indicator status-read">{{ t('read') }}</div>
<div class="status-indicator status-favorite">{{ t('favorite') }}</div>

<!-- Modal/Dialog -->
<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
  <div class="bg-bg-primary w-full max-w-2xl rounded-2xl shadow-2xl border border-border">
    <div class="modal-header">
      <h2 class="text-xl font-bold">{{ t('modalTitle') }}</h2>
      <button @click="close" class="btn-icon">
        <i class="ph ph-x"></i>
      </button>
    </div>
  </div>
</div>
```

### Toast Notifications

```javascript
// Success message
window.showToast(message, 'success');

// Error message
window.showToast(t('operationFailed'), 'error');

// Info message with custom duration
window.showToast(t('updateAvailable'), 'info', 5000);
```

### Confirm Dialogs

```javascript
const confirmed = await window.showConfirm(
  t('confirmDelete'),
  t('deleteWarning'),
  true  // isDanger - shows red confirmation button
);

if (confirmed) {
  // Proceed with dangerous operation
}
```

### Context Menu Pattern

```vue
<script setup>
import { useContextMenu } from '@/composables/ui/useContextMenu';

const { contextMenu, openContextMenu, closeContextMenu } = useContextMenu();

// Define menu items
const menuItems = [
  { label: t('edit'), action: 'edit', icon: 'ph-pencil' },
  { label: t('delete'), action: 'delete', icon: 'ph-trash', danger: true },
  { type: 'divider' },
  { label: t('markAsRead'), action: 'mark-read' }
];

// Handle right-click
function handleRightClick(event: MouseEvent, item: Article) {
  event.preventDefault();
  openContextMenu(event, menuItems, item);
}

// Handle menu action
function handleMenuAction(action: string, item: Article) {
  switch (action) {
    case 'edit':
      // Handle edit
      break;
    case 'delete':
      // Handle delete
      break;
    case 'mark-read':
      // Handle mark as read
      break;
  }
}
</script>

<template>
  <div @contextmenu="handleRightClick($event, item)">
    <!-- Item content -->
  </div>
</template>
```

## Database Schema and Operations

### Core Tables

```sql
-- Feeds table
CREATE TABLE feeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    link TEXT,  -- Website homepage
    description TEXT,
    category TEXT,
    image_url TEXT,
    last_updated DATETIME,
    last_error TEXT,
    discovery_completed BOOLEAN DEFAULT FALSE
);

-- Articles table
CREATE TABLE articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    image_url TEXT,
    content TEXT,
    published_at DATETIME NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    is_favorite BOOLEAN DEFAULT FALSE,
    is_hidden BOOLEAN DEFAULT FALSE,
    translated_title TEXT,
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

-- Settings table (key-value store)
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- Indexes for performance
CREATE INDEX idx_articles_feed_id ON articles(feed_id);
CREATE INDEX idx_articles_published_at ON articles(published_at);
CREATE INDEX idx_articles_is_read ON articles(is_read);
CREATE INDEX idx_articles_is_favorite ON articles(is_favorite);
```

### Cleanup Logic

Auto-cleanup preserves favorites:

```go
func (db *DB) CleanupOldArticles(maxAgeDays int) (int64, error) {
    cutoffDate := time.Now().AddDate(0, 0, -maxAgeDays)

    // Delete old articles EXCEPT favorites
    result, err := db.conn.Exec(`
        DELETE FROM articles
        WHERE published_at < ? AND is_favorite = 0
    `, cutoffDate)

    if err != nil {
        return 0, fmt.Errorf("cleanup articles: %w", err)
    }

    // Run VACUUM to reclaim space
    _, _ = db.conn.Exec("VACUUM")

    return result.RowsAffected()
}
```

## Settings Management (CRITICAL)

**IMPORTANT**: When adding/modifying/deleting a setting, you MUST update ALL 8 locations below. Missing even ONE will cause bugs!

### Required Updates Checklist

Use this checklist for EVERY settings change:

- [ ] **1. Default Values** (2 files)
  - `config/defaults.json` - Frontend default values
  - `internal/config/defaults.json` - Backend default values

- [ ] **2. Backend Type Definition**
  - `internal/config/config.go` - Add field to `Defaults` struct
  - `internal/config/config.go` - Add case in `GetString()` switch

- [ ] **3. Database Initialization**
  - `internal/database/db.go` - Add key to `settingsKeys` array

- [ ] **4. Backend API Handler** (4 sub-tasks)
  - `internal/handlers/settings/settings_handlers.go` - GET: Add `GetSetting()` call
  - `internal/handlers/settings/settings_handlers.go` - GET: Add to response map
  - `internal/handlers/settings/settings_handlers.go` - POST: Add field to request struct
  - `internal/handlers/settings/settings_handlers.go` - POST: Add `SetSetting()` call

- [ ] **5. Frontend Type Definition**
  - `frontend/src/types/settings.ts` - Add field to `SettingsData` interface

- [ ] **6. Frontend Settings Management** (2 sub-tasks)
  - `frontend/src/composables/core/useSettings.ts` - Add to initial `settings` ref
  - `frontend/src/composables/core/useSettings.ts` - Add to `fetchSettings()` parsing

- [ ] **7. Frontend Auto-Save**
  - `frontend/src/composables/core/useSettingsAutoSave.ts` - Add to POST body

- [ ] **8. UI Component** (if user-facing)
  - Create/update in `frontend/src/components/modals/settings/`

### Quick Example: Adding `new_feature_enabled`

```typescript
// 1. config/defaults.json & internal/config/defaults.json
"new_feature_enabled": false

// 2. internal/config/config.go
type Defaults struct {
    NewFeatureEnabled bool `json:"new_feature_enabled"`
}
case "new_feature_enabled":
    return strconv.FormatBool(defaults.NewFeatureEnabled)

// 3. internal/database/db.go
settingsKeys := []string{"new_feature_enabled", /*...*/}

// 4. internal/handlers/settings/settings_handlers.go
// GET:
newFeature, _ := h.DB.GetSetting("new_feature_enabled")
"new_feature_enabled": newFeature,
// POST struct:
NewFeatureEnabled string `json:"new_feature_enabled"`
// POST handler:
if req.NewFeatureEnabled != "" {
    h.DB.SetSetting("new_feature_enabled", req.NewFeatureEnabled)
}

// 5. frontend/src/types/settings.ts
export interface SettingsData {
  new_feature_enabled: boolean;
}

// 6. frontend/src/composables/core/useSettings.ts
const settings = ref({
  new_feature_enabled: settingsDefaults.new_feature_enabled,
});
new_feature_enabled: data.new_feature_enabled === 'true',

// 7. frontend/src/composables/core/useSettingsAutoSave.ts
new_feature_enabled: (settings.value.new_feature_enabled ??
                     settingsDefaults.new_feature_enabled).toString(),
```

📚 **Detailed Guide**: See [CODE_PATTERNS.md](../docs/CODE_PATTERNS.md#settings-management)

## Security Best Practices

### Input Validation

Always validate user inputs, especially URLs and file paths:

```go
// Validate URL format and scheme
func validateFeedURL(urlStr string) error {
    u, err := url.Parse(urlStr)
    if err != nil {
        return fmt.Errorf("invalid URL: %w", err)
    }

    if u.Scheme != "http" && u.Scheme != "https" {
        return errors.New("URL must use HTTP or HTTPS")
    }

    return nil
}

// Validate file path to prevent directory traversal
func validateFilePath(baseDir, filePath string) error {
    cleanPath := filepath.Clean(filePath)
    if !strings.HasPrefix(cleanPath, filepath.Clean(baseDir)) {
        return errors.New("invalid file path: path traversal detected")
    }
    return nil
}
```

### Safe Command Execution

**NEVER** use shell command concatenation:

```go
// ❌ BAD: Command injection vulnerability
cmd := exec.Command("sh", "-c", "rm " + filePath)

// ✅ GOOD: Use Go standard library
if err := os.Remove(filePath); err != nil {
    return fmt.Errorf("remove file: %w", err)
}

// ✅ GOOD: If external command is necessary, use separate args
cmd := exec.Command("installer.exe", "/S") // No concatenation
```

### File Operations

Always clean up temporary files and use proper error handling:

```go
// Schedule cleanup with timeout
scheduleCleanup := func(filePath string, delay time.Duration) {
    go func() {
        time.Sleep(delay)
        if err := os.Remove(filePath); err != nil {
            log.Printf("Failed to cleanup %s: %v", filePath, err)
        } else {
            log.Printf("Cleaned up temporary file: %s", filePath)
        }
    }()
}
```

## Version Management (CRITICAL)

When updating version, modify ALL of these files:

1. `internal/version/version.go` - Version constant
2. `wails.json` - "version" and "info.productVersion" fields
3. `frontend/package.json` - "version" field
4. `frontend/package-lock.json` - "version" field
5. `frontend/src/components/modals/settings/about/AboutTab.vue` - appVersion ref default
6. `website/package.json` - "version" field
7. `website/package-lock.json` - "version" field
8. `README.md` - Version badge
9. `README_zh.md` - Version badge
10. `CHANGELOG.md` - Add new version entry

📚 **Detailed Guide**: See [VERSION_MANAGEMENT.md](../docs/VERSION_MANAGEMENT.md)

## Don'ts

❌ **Don't**:

- Use `var` declarations in Vue (use `ref` or `reactive`)
- Hardcode user-facing strings (always use i18n `t()`)
- Use inline styles (use Tailwind classes or scoped styles)
- Forget error handling in async operations
- Use `any` type without strong justification
- Commit API keys, secrets, or sensitive data
- Use `v-html` for user content (XSS risk)
- Make breaking changes without migration path
- Use shell command concatenation (security risk)
- Create multiple deep watchers when one suffices
- Forget to clean up timers/intervals on component unmount
- Delete favorited articles during cleanup operations
- Use synchronous operations in UI thread for long tasks

## Do's

✅ **Do**:

- Use TypeScript with proper type annotations
- Follow existing code patterns and conventions
- Write comprehensive tests for new features
- Keep functions small and focused (single responsibility)
- Use meaningful variable and function names
- Handle all edge cases and error conditions
- Validate inputs thoroughly (URLs, file paths, user data)
- Log errors with appropriate context for debugging
- Use semantic HTML with proper ARIA attributes
- Debounce frequent operations (auto-save, search, etc.)
- Use `os.Remove()` instead of shell commands for file operations
- Clean up resources (timers, goroutines, event listeners) properly
- Preserve favorited articles during any cleanup operation
- Use prepared statements for all database queries
- Implement proper loading states and progress indicators
- Follow semantic versioning (MAJOR.MINOR.PATCH)
- Document exported functions and complex logic
- Use goroutines for concurrent operations
- Implement graceful degradation for network failures

## Quick Reference

**Build Commands**:
- Development: `wails dev`
- Production Build: `make build` or `wails build -skipbindings`
- Important: Always use `-skipbindings` flag with `wails build` (MrRSS uses HTTP API, not Wails bindings)

**Store Access**: `const store = useAppStore()`
**i18n**: `const { t } = useI18n()`
**Theme**: `store.theme` returns `'light'` or `'dark'`
**Language**: `store.i18n.locale.value` returns `'en'` or `'zh'`
**Toast**: `window.showToast(message, type)`
**Confirm**: `await window.showConfirm(title, message, isDanger)`
**Settings API**: `GET/POST /api/settings`
**Articles API**: `GET /api/articles` with query params
**Progress Polling**: `GET /api/progress` for async operations

---

When generating code, prioritize:

1. **Correctness**: Code that works and handles errors properly
2. **Consistency**: Follow existing patterns in the codebase
3. **Clarity**: Easy to understand and maintain
4. **Performance**: Efficient queries, minimal re-renders, proper cleanup
5. **Security**: Input validation, safe file operations, no injection vulnerabilities
6. **User Experience**: Loading states, progress indicators, error messages

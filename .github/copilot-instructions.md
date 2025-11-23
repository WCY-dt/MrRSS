# GitHub Copilot Instructions for MrRSS

## Project Context

MrRSS is a cross-platform desktop RSS reader built with Wails (Go + Vue.js). It emphasizes simplicity, privacy, and modern UI design.

**Current Version**: 1.1.4

## Tech Stack

- **Backend**: Go 1.21+, Wails v2, SQLite (`modernc.org/sqlite`)
- **Frontend**: Vue 3 (Composition API), Tailwind CSS, Vite
- **Tools**: npm, Wails CLI, Go modules

## Key Features to Understand

- RSS/Atom feed parsing with concurrent fetching
- Auto-translation (Google Translate free, DeepL with API key)
- OPML import/export for feed portability
- Configurable auto-cleanup by article age
- In-app update system with progress tracking
- Auto-save settings (no save button needed)
- Multi-language support (English/Chinese)
- Light/Dark/Auto themes with system detection

## Code Patterns

### Backend (Go)

When writing Go code:

```go
// Always use context for exported methods
func (h *Handler) MethodName(ctx context.Context, param string) (Result, error) {
    if param == "" {
        return Result{}, errors.New("param is required")
    }
    
    // Implementation
    return result, nil
}

// Use prepared statements for SQL
stmt, err := db.Prepare("SELECT * FROM table WHERE id = ?")
if err != nil {
    return nil, fmt.Errorf("prepare: %w", err)
}
defer stmt.Close()

// Handle errors explicitly
if err != nil {
    return nil, fmt.Errorf("operation failed: %w", err)
}
```

### Frontend (Vue)

When writing Vue components:

```vue
<script setup>
import { ref, computed, onMounted } from 'vue';
import { store } from '../store.js';

// Props with validation
const props = defineProps({
    item: { type: Object, required: true },
    isActive: { type: Boolean, default: false }
});

// Emits declaration
const emit = defineEmits(['update', 'delete']);

// Reactive state
const isLoading = ref(false);
const items = ref([]);

// Computed properties
const filteredItems = computed(() => {
    return items.value.filter(item => /* condition */);
});

// Methods
async function loadData() {
    isLoading.value = true;
    try {
        // API call via Wails bindings
        const data = await SomeBackendMethod();
        items.value = data;
    } catch (e) {
        console.error(e);
        window.showToast(store.i18n.t('errorLoadingData'), 'error');
    } finally {
        isLoading.value = false;
    }
}

// Lifecycle
onMounted(() => {
    loadData();
});
</script>

<template>
    <div class="container">
        <h2 class="text-lg font-semibold">{{ store.i18n.t('title') }}</h2>
        
        <div v-if="isLoading">{{ store.i18n.t('loading') }}</div>
        <div v-else-if="items.length === 0">{{ store.i18n.t('noItems') }}</div>
        <div v-else v-for="item in filteredItems" :key="item.id">
            {{ item.name }}
        </div>
    </div>
</template>

<style scoped>
.container {
    @apply p-4 bg-bg-primary rounded-lg;
}
</style>
```

## Styling Guidelines

### Tailwind CSS Patterns

Use these semantic class combinations:

```html
<!-- Buttons -->
<button class="btn-primary">Primary Action</button>
<button class="btn-secondary">Secondary Action</button>
<button class="btn-danger">Dangerous Action</button>

<!-- Cards -->
<div class="bg-bg-primary border border-border rounded-lg p-4">
    <h3 class="text-text-primary font-semibold">Title</h3>
    <p class="text-text-secondary text-sm">Description</p>
</div>

<!-- Inputs -->
<input class="input-field" type="text" />

<!-- Modal -->
<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
    <div class="bg-bg-primary w-full max-w-2xl rounded-2xl shadow-2xl border border-border">
        <!-- Content -->
    </div>
</div>
```

### Theme Variables

Use CSS variables for theming:

```css
/* Colors follow theme */
background-color: var(--color-bg-primary);
color: var(--color-text-primary);

/* Or use Tailwind classes */
class="bg-bg-primary text-text-primary"
```

## Internationalization

Always use i18n for user-facing strings:

```vue
<!-- Template -->
<h1>{{ store.i18n.t('welcome') }}</h1>
<button :title="store.i18n.t('clickToOpen')">

<!-- Script -->
window.showToast(store.i18n.t('successMessage'), 'success');
```

To add new strings, edit `frontend/src/i18n.js`:

```javascript
export const translations = {
    en: {
        // English translations
        newKey: 'New String',
    },
    zh: {
        // Chinese translations
        newKey: '新字符串',
    }
};
```

## Common Patterns

### Settings Auto-Save Pattern

Settings now auto-save without a save button. Use debouncing:

```vue
<script setup>
import { watch, onUnmounted } from 'vue';

let saveTimeout = null;

async function autoSave() {
    await fetch('/api/settings', {
        method: 'POST',
        body: JSON.stringify(settings)
    });
    // Apply settings immediately
    store.applySettings(settings);
}

function debouncedAutoSave() {
    if (saveTimeout) clearTimeout(saveTimeout);
    saveTimeout = setTimeout(autoSave, 500); // 500ms debounce
}

// Single deep watcher
watch(() => props.settings, debouncedAutoSave, { deep: true });

// Cleanup to prevent memory leaks
onUnmounted(() => {
    if (saveTimeout) {
        clearTimeout(saveTimeout);
        saveTimeout = null;
    }
});
</script>
```

### API Calls (Frontend → Backend)

MrRSS uses HTTP fetch for API calls, not Wails bindings:

```javascript
// Settings API
const res = await fetch('/api/settings');
const settings = await res.json();

// Update settings
await fetch('/api/settings', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(settingsObject)
});

// Check for updates
const res = await fetch('/api/check-updates');
const updateInfo = await res.json();
```

### API Calls (Frontend → Backend)

```javascript
// Import from generated bindings
import { MethodName } from './wailsjs/go/internal/handlers/Handler.js';

// Call backend method
try {
    const result = await MethodName(param1, param2);
    // Handle result
} catch (error) {
    console.error('Error:', error);
    window.showToast(store.i18n.t('error'), 'error');
}
```

### Progress Bar Pattern

For download/upload operations:

```vue
<template>
    <button @click="handleDownload" :disabled="downloading">
        <i v-if="downloading" class="ph ph-circle-notch animate-spin"></i>
        {{ downloading ? `${store.i18n.t('downloading')} ${progress}%` : store.i18n.t('download') }}
    </button>
    
    <!-- Progress bar -->
    <div v-if="downloading" class="w-full bg-bg-tertiary rounded-full h-2 overflow-hidden">
        <div class="bg-accent h-full transition-all duration-300" :style="{ width: progress + '%' }"></div>
    </div>
</template>

<script setup>
const downloading = ref(false);
const progress = ref(0);

// Simulated progress (for operations without real progress tracking)
async function handleDownload() {
    downloading.value = true;
    progress.value = 0;
    
    const progressInterval = setInterval(() => {
        if (progress.value < 90) progress.value += 10;
    }, 500);
    
    try {
        await fetch('/api/download', { method: 'POST' });
        clearInterval(progressInterval);
        progress.value = 100;
    } catch (e) {
        clearInterval(progressInterval);
        // Handle error
    } finally {
        downloading.value = false;
    }
}
</script>
```

### Database Settings Pattern

Settings are stored as string key-value pairs:

```go
// Backend: Get and set settings
func (h *Handler) HandleSettings(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        autoCleanup, _ := h.DB.GetSetting("auto_cleanup_enabled")
        maxAge, _ := h.DB.GetSetting("max_article_age_days")
        // Convert strings to appropriate types
        json.NewEncoder(w).Encode(map[string]string{
            "auto_cleanup_enabled": autoCleanup,
            "max_article_age_days": maxAge,
        })
    } else if r.Method == http.MethodPost {
        // Save settings
        h.DB.SetSetting("auto_cleanup_enabled", req.AutoCleanupEnabled)
        h.DB.SetSetting("max_article_age_days", req.MaxArticleAgeDays)
    }
}
```

```javascript
// Dispatch event
window.dispatchEvent(new CustomEvent('event-name', {
    detail: { data: value }
}));

// Listen for event
window.addEventListener('event-name', (e) => {
    const data = e.detail.data;
    // Handle event
});
```

### Toast Notifications

```javascript
window.showToast(message, type);  // type: 'success' | 'error' | 'info' | 'warning'
```

### Confirm Dialogs

```javascript
const confirmed = await window.showConfirm(
    store.i18n.t('confirmTitle'),
    store.i18n.t('confirmMessage'),
    true  // isDanger (shows red button)
);

if (confirmed) {
    // Proceed with action
}
```

## Database Operations

### Query Pattern

```go
// Use prepared statements
func (db *Database) GetArticles(feedID int) ([]models.Article, error) {
    rows, err := db.conn.Query(`
        SELECT id, title, url, content, published_at
        FROM articles
        WHERE feed_id = ?
        ORDER BY published_at DESC
    `, feedID)
    if err != nil {
        return nil, fmt.Errorf("query: %w", err)
    }
    defer rows.Close()
    
    var articles []models.Article
    for rows.Next() {
        var article models.Article
        err := rows.Scan(&article.ID, &article.Title, &article.URL, &article.Content, &article.PublishedAt)
        if err != nil {
            return nil, fmt.Errorf("scan: %w", err)
        }
        articles = append(articles, article)
    }
    
    return articles, rows.Err()
}
```

### Cleanup Logic

Auto-cleanup deletes old articles but **preserves favorites**:

```go
func (db *DB) CleanupOldArticles() (int64, error) {
    // Get configurable threshold
    maxAgeDaysStr, _ := db.GetSetting("max_article_age_days")
    maxAgeDays := 30 // default
    if days, err := strconv.Atoi(maxAgeDaysStr); err == nil && days > 0 {
        maxAgeDays = days
    }
    
    cutoffDate := time.Now().AddDate(0, 0, -maxAgeDays)
    
    // Delete old articles EXCEPT favorites
    result, err := db.Exec(`
        DELETE FROM articles 
        WHERE published_at < ? 
        AND is_favorite = 0
    `, cutoffDate)
    
    // Run VACUUM to reclaim space
    _, _ = db.Exec("VACUUM")
    
    return result.RowsAffected()
}
```

## Database Operations

### Query Pattern

```go
// Use prepared statements
func (db *Database) GetArticles(feedID int) ([]models.Article, error) {
    rows, err := db.conn.Query(`
        SELECT id, title, url, content, published_at
        FROM articles
        WHERE feed_id = ?
        ORDER BY published_at DESC
    `, feedID)
    if err != nil {
        return nil, fmt.Errorf("query: %w", err)
    }
    defer rows.Close()
    
    var articles []models.Article
    for rows.Next() {
        var article models.Article
        err := rows.Scan(&article.ID, &article.Title, &article.URL, &article.Content, &article.PublishedAt)
        if err != nil {
            return nil, fmt.Errorf("scan: %w", err)
        }
        articles = append(articles, article)
    }
    
    return articles, rows.Err()
}
```

## Testing

### Backend Tests

```go
func TestFunctionName(t *testing.T) {
    // Setup
    input := "test"
    expected := "result"
    
    // Execute
    result, err := FunctionName(input)
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Frontend Tests

```javascript
describe('Component', () => {
    it('should render correctly', () => {
        // Test implementation
    });
});
```

## Documentation

When adding new features:

1. **Code Comments**: Document exported functions

   ```go
   // FetchFeed retrieves and parses an RSS feed from the given URL.
   // Returns an error if the URL is invalid or the feed cannot be parsed.
   func FetchFeed(url string) (*Feed, error) {
   ```

2. **README Updates**: Document user-facing features

3. **Changelog**: Update CHANGELOG.md for releases

## Security Best Practices

### Input Validation

Always validate user inputs, especially file paths and URLs:

```go
// Validate URL is from official repository
const allowedURLPrefix = "https://github.com/WCY-dt/MrRSS/releases/download/"
if !strings.HasPrefix(req.DownloadURL, allowedURLPrefix) {
    return errors.New("invalid download URL")
}

// Validate file path to prevent traversal
if strings.Contains(req.AssetName, "..") || 
   strings.Contains(req.AssetName, "/") || 
   strings.Contains(req.AssetName, "\\") {
    return errors.New("invalid asset name")
}

// Validate file is within temp directory
cleanPath := filepath.Clean(req.FilePath)
if !strings.HasPrefix(cleanPath, filepath.Clean(tempDir)) {
    return errors.New("invalid file path")
}
```

### Command Execution

**NEVER** use shell command concatenation. Always use safe alternatives:

```go
// ❌ BAD: Command injection vulnerability
cmd := exec.Command("sh", "-c", "rm " + filePath)

// ✅ GOOD: Use os.Remove instead
if err := os.Remove(filePath); err != nil {
    log.Printf("Failed to remove file: %v", err)
}

// ✅ GOOD: If you must use exec, pass args separately
cmd := exec.Command("installer.exe", "/S") // Not concatenated
```

### File Operations

Always clean up temporary files and use goroutines for delayed operations:

```go
// Schedule cleanup with proper error handling
scheduleCleanup := func(filePath string, delay time.Duration) {
    go func() {
        time.Sleep(delay)
        if err := os.Remove(filePath); err != nil {
            log.Printf("Failed to remove: %v", err)
        } else {
            log.Printf("Successfully removed: %s", filePath)
        }
    }()
}

scheduleCleanup(installerPath, 3*time.Second)
```

## Documentation

When adding new features:

1. **Code Comments**: Document exported functions

   ```go
   // FetchFeed retrieves and parses an RSS feed from the given URL.
   // Returns an error if the URL is invalid or the feed cannot be parsed.
   func FetchFeed(url string) (*Feed, error) {
   ```

2. **README Updates**: Document user-facing features

3. **Changelog**: Update for releases

## Don'ts

❌ **Don't**:

- Use `var` in Vue (use `ref` or `reactive`)
- Hardcode strings (use i18n)
- Use inline styles (use Tailwind classes)
- Forget error handling
- Use `any` type without good reason
- Commit API keys or secrets
- Use `v-html` (XSS risk)
- Make breaking changes without discussion
- Use shell command concatenation (command injection risk)
- Create multiple watchers when one deep watcher suffices
- Forget to clean up timers/intervals on component unmount
- Delete favorited articles during cleanup

## Do's

✅ **Do**:

- Use TypeScript-style JSDoc for better IDE support
- Follow existing code patterns
- Write tests for new features
- Keep functions small and focused
- Use meaningful variable names
- Handle edge cases
- Validate inputs (especially file paths and URLs)
- Log errors appropriately with context
- Use semantic HTML
- Debounce frequent operations (e.g., auto-save)
- Use `os.Remove()` instead of shell commands for file operations
- Clean up resources (timers, goroutines) properly
- Preserve favorited articles during any cleanup operation
- Update all 7 version files when bumping version

## File Naming

- Components: `PascalCase.vue` (e.g., `ArticleList.vue`)
- Go files: `lowercase.go` (e.g., `fetcher.go`)
- Test files: `*_test.go`
- Utilities: `kebab-case.js` (e.g., `date-utils.js`)

## Useful Commands

```bash
# Development
wails dev                          # Run in dev mode with hot reload
wails build                        # Build for production
wails build -clean                 # Clean build (removes previous artifacts)

# Testing
go test ./...                      # Run all Go tests
go test ./internal/database -v     # Run database tests with verbose output
cd frontend && npm test            # Run frontend tests (if available)

# Linting
go vet ./...                       # Go linter
go fmt ./...                       # Format Go code
cd frontend && npm run lint        # Frontend linter (if configured)

# Dependencies
go mod tidy                        # Clean Go dependencies
cd frontend && npm install         # Install npm packages

# Building Frontend Only
cd frontend && npm run build       # Build frontend assets
```

## Version Management

**CRITICAL**: When updating version, modify these 7 files:

1. `internal/version/version.go` - Version constant
2. `wails.json` - Two fields: "version" and "info.productVersion"
3. `frontend/package.json` - "version" field
4. `frontend/src/components/modals/settings/AboutTab.vue` - appVersion ref default
5. `README.md` - Version badge
6. `README_zh.md` - Version badge
7. `CHANGELOG.md` - Add new version entry

**Always use Semantic Versioning**: MAJOR.MINOR.PATCH (e.g., 1.1.4)

## Quick Reference

**Get current theme**: `store.theme` (returns `'light'` or `'dark'`)
**Get current language**: `store.i18n.locale.value` (returns `'en'` or `'zh')`)
**Global store**: `import { store } from './store.js'`
**Show notification**: `window.showToast(message, type)`
**Confirm action**: `await window.showConfirm(title, message, isDanger)`

---

When generating code, prioritize:

1. **Correctness**: Code that works
2. **Consistency**: Follow existing patterns
3. **Clarity**: Easy to understand
4. **Maintainability**: Easy to modify later

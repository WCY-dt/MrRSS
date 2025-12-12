# MrRSS Architecture Documentation

## Overview

MrRSS is built with a modern, modular architecture using:

- **Backend**: Go 1.24+ with Wails v2.11+ framework
- **System Tray**: `fyne.io/systray` for cross-platform system tray integration (disabled on macOS due to AppDelegate conflicts with Wails)
- **Frontend**: Vue 3.5+ Composition API with TypeScript
- **Database**: SQLite with pure Go implementation (`modernc.org/sqlite`)
- **Communication**: HTTP REST API (not Wails bindings)

## Backend Architecture

### Handler Organization

Handlers are organized by feature domains in `internal/handlers/`:

```plaintext
handlers/
├── core/          # Core handler initialization and scheduling
├── article/       # Article CRUD and filtering
├── feed/          # Feed management
├── discovery/     # Feed discovery
├── opml/          # OPML import/export
├── rules/         # Filtering rules
├── script/        # Custom script execution
├── settings/      # Settings management
├── summary/       # Article summarization
├── translation/   # Translation services
└── update/        # Application updates
```

### Core Components

#### Database Layer (`internal/database/`)

- `db.go` - Database initialization and core operations
- `article_db.go` - Article CRUD operations
- `feed_db.go` - Feed CRUD operations
- `settings_db.go` - Key-value settings store
- `cleanup_db.go` - Auto-cleanup logic (preserves favorites)

**Key Features**:

- SQLite with WAL mode for better concurrency
- Prepared statements for all queries
- Indexed queries for performance
- Automatic cleanup with favorite preservation

#### Feed Processing (`internal/feed/`)

- `fetcher.go` - RSS/Atom parsing with `gofeed`, concurrent fetching
- `script_executor.go` - Custom script execution for non-standard feeds

**Supported Scripts**:

- Python (`.py`)
- Shell (`.sh`)
- PowerShell (`.ps1`)
- Node.js (`.js`)
- Ruby (`.rb`)

#### Discovery System (`internal/discovery/`)

- `feed_discovery.go` - Main discovery orchestration
- `html_parser.go` - HTML parsing for RSS links
- `rss_detector.go` - RSS feed detection logic
- `service.go` - Discovery service with progress tracking

**Features**:

- Discover feeds from URLs
- Batch discovery from friend links
- Real-time progress tracking
- Comprehensive deduplication

#### Summarization (`internal/summary/`)

- `summarizer.go` - TF-IDF and TextRank-based summarization
- `ai_summarizer.go` - AI-based summarization using OpenAI-compatible APIs
- `scoring.go` - Sentence scoring algorithms
- `text_utils.go` - Text processing utilities

**Local Algorithms**:

- TF-IDF for term importance
- TextRank for sentence ranking
- Combined scoring (0.5 TF-IDF + 0.5 TextRank)
- Smart sentence selection preserving narrative flow

**AI Summarization**:

- Supports OpenAI-compatible APIs (GPT, Claude, etc.)
- Configurable API endpoint and model
- Token-efficient prompts

#### Translation (`internal/translation/`)

- `translator.go` - Translation interface and factory
- `google.go` - Google Translate (free, no API key)
- `deepl.go` - DeepL API integration
- `baidu.go` - Baidu Translation API integration
- `ai.go` - AI-based translation integration

## Frontend Architecture

### Component Organization

Components are organized by feature in `frontend/src/components/`:

```plaintext
components/
├── article/       # Article display and rendering
│   ├── ArticleList.vue
│   ├── ArticleItem.vue
│   ├── ArticleDetail.vue
│   ├── ArticleContent.vue
│   ├── ArticleToolbar.vue
│   └── parts/     # Content rendering parts
│       ├── ArticleTitle.vue
│       ├── ArticleSummary.vue
│       ├── ArticleBody.vue
│       └── ArticleLoading.vue
├── sidebar/       # Feed list sidebar
│   ├── Sidebar.vue
│   ├── SidebarFeed.vue
│   ├── SidebarCategory.vue
│   └── SidebarNavItem.vue
├── common/        # Reusable components
│   ├── Toast.vue
│   ├── ContextMenu.vue
│   └── ImageViewer.vue
└── modals/        # Modal dialogs
    ├── SettingsModal.vue
    ├── settings/  # Settings tabs
    ├── feed/      # Feed modals
    ├── filter/    # Filter modals
    ├── rules/     # Rules editor
    ├── discovery/ # Discovery modal
    └── common/    # Common modals
```

### Composables Organization

Composables provide reusable logic in `frontend/src/composables/`:

```plaintext
composables/
├── article/       # Article-related logic
│   ├── useArticleDetail.ts
│   ├── useArticleList.ts
│   ├── useArticleContent.ts
│   └── useArticleSummary.ts
├── feed/          # Feed management
│   ├── useFeedManagement.ts
│   └── useFeedRefresh.ts
├── discovery/     # Feed discovery
│   └── useFeedDiscovery.ts
├── filter/        # Article filtering
│   └── useArticleFilter.ts
├── rules/         # Filtering rules
│   └── useRules.ts
├── ui/            # UI utilities
│   ├── useContextMenu.ts
│   ├── useKeyboardShortcuts.ts
│   └── useToast.ts
└── core/          # Core utilities
    └── useSettings.ts
```

### State Management

Pinia store (`frontend/src/stores/app.ts`) manages global state:

- Articles list and selection
- Feeds and categories
- Filter states
- Theme and locale
- Refresh progress
- Unread counts

### Multimedia Support

Enhanced content rendering (`ArticleContent.vue` + `ArticleContent.css`):

- **Images**: Clickable for viewer, right-click context menu, download support
- **Audio**: Full-width player with podcast container styling
- **Video**: Responsive player with proper aspect ratio
- **Iframes**: 16:9 aspect ratio for YouTube/Vimeo embeds
- **Rich Text**: Tables, blockquotes, code blocks, definition lists

### Translation Integration

Auto-translation features:

- Title translation (on-demand)
- Content paragraph translation (inline display)
- Summary translation
- Supports Google Translate, DeepL, Baidu Translation, and AI-based translation

## Communication Flow

### HTTP API Pattern

Frontend uses direct HTTP fetch (not Wails bindings):

```javascript
// GET request
const response = await fetch('/api/articles');
const articles = await response.json();

// POST request
await fetch('/api/settings', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(data)
});
```

### Backend Handler Pattern

```go
func (h *Handler) GetArticles(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters
    feedID := r.URL.Query().Get("feed_id")

    // Database operation
    articles, err := h.DB.GetArticles(feedID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // JSON response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(articles)
}
```

## Data Flow

### Feed Refresh Flow

1. User triggers refresh (manual or scheduled)
2. Backend starts concurrent feed fetching
3. For each feed:
   - Execute script (if custom script) OR fetch RSS/Atom
   - Parse feed with `gofeed`
   - Extract articles
   - Store new articles in database
4. Update progress tracking
5. Frontend polls progress endpoint
6. UI updates with new articles

### Article Display Flow

1. User selects feed/category/filter
2. Frontend fetches articles from API
3. Article list displays with virtual scrolling
4. User selects article
5. Content loaded based on view mode:
   - **Rendered**: Parse and display HTML content
   - **Webpage**: Load original URL in iframe
6. Optional: Generate summary, translate content

### Discovery Flow

1. User initiates discovery (single URL or batch)
2. Backend creates discovery session
3. For each source:
   - Fetch HTML content
   - Parse for RSS links
   - Detect common RSS patterns
   - Verify feeds
4. Progress tracking via polling
5. Frontend displays discovered feeds
6. User selects feeds to import

## Security Considerations

### Input Validation

- URL validation for feeds and websites
- File path validation to prevent directory traversal
- Script path sandboxing within scripts directory

### Safe Operations

- Use `os.Remove()` instead of shell commands
- Prepared SQL statements prevent injection
- No shell command concatenation
- XSS prevention in content rendering

### Script Execution

- Timeout enforcement (30 seconds)
- Working directory restricted to scripts folder
- Path traversal prevention
- Separate execution context per script

## Performance Optimizations

### Database

- SQLite WAL mode for concurrent access
- Indexed columns for frequent queries
- Prepared statement caching
- Periodic VACUUM for space reclamation

### Frontend

- Virtual scrolling for large article lists
- Debounced operations (search, auto-save)
- Lazy loading of article content
- Efficient state updates with Pinia

### Concurrency

- Goroutines for parallel feed fetching
- Background task scheduling
- Progress tracking without blocking
- Graceful timeout handling

## Related Documentation

- [Code Patterns](CODE_PATTERNS.md) - Common coding patterns and examples
- [Testing Guide](TESTING.md) - Testing strategies and examples
- [Custom Scripts](CUSTOM_SCRIPTS.md) - Guide for writing custom feed scripts

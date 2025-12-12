package database

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"MrRSS/internal/config"

	_ "modernc.org/sqlite"
)

// DB wraps sql.DB with initialization state tracking.
type DB struct {
	*sql.DB
	ready chan struct{}
	once  sync.Once
}

// NewDB creates a new database connection with optimized settings.
func NewDB(dataSourceName string) (*DB, error) {
	// Add busy_timeout to prevent "database is locked" errors
	// Also enable WAL mode for better concurrency
	// Add performance optimizations: increase cache size, set synchronous=NORMAL
	if !strings.Contains(dataSourceName, "?") {
		dataSourceName += "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=cache_size(-32000)&_pragma=synchronous(NORMAL)"
	} else {
		dataSourceName += "&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=cache_size(-32000)&_pragma=synchronous(NORMAL)"
	}

	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Set connection pool limits for better performance
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{
		DB:    db,
		ready: make(chan struct{}),
	}, nil
}

// Init initializes the database schema and settings.
func (db *DB) Init() error {
	var err error
	db.once.Do(func() {
		defer close(db.ready)

		if err = db.Ping(); err != nil {
			return
		}

		if err = initSchema(db.DB); err != nil {
			return
		}

		// Create settings table if not exists
		_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT
		)`)

		// Insert default settings if they don't exist (using centralized defaults from config)
		settingsKeys := []string{
			"update_interval", "refresh_mode", "translation_enabled", "target_language", "translation_provider",
			"deepl_api_key", "baidu_app_id", "baidu_secret_key", "ai_api_key", "ai_endpoint", "ai_model",
			"ai_system_prompt", "auto_cleanup_enabled", "max_cache_size_mb", "max_article_age_days", "language", "theme",
			"last_article_update", "show_hidden_articles", "default_view_mode", "summary_enabled", "summary_length",
			"summary_provider", "summary_ai_api_key", "summary_ai_endpoint", "summary_ai_model", "summary_ai_system_prompt",
			"media_cache_enabled", "media_cache_max_size_mb", "media_cache_max_age_days",
			"proxy_enabled", "proxy_type", "proxy_host", "proxy_port", "proxy_username", "proxy_password",
			"shortcuts", "rules", "startup_on_boot", "close_to_tray", "google_translate_endpoint", "show_article_preview_images",
			"window_x", "window_y", "window_width", "window_height", "window_maximized",
		}
		for _, key := range settingsKeys {
			defaultVal := config.GetString(key)
			_, _ = db.Exec(fmt.Sprintf(`INSERT OR IGNORE INTO settings (key, value) VALUES ('%s', '%s')`, key, defaultVal))
		}

		// Migration: Add link column to feeds table if it doesn't exist
		// Note: SQLite doesn't support IF NOT EXISTS for ALTER TABLE ADD COLUMN.
		// Error is ignored - if column exists, the operation fails harmlessly.
		_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN link TEXT DEFAULT ''`)

		// Migration: Add discovery_completed column to feeds table
		// Error is ignored - if column exists, the operation fails harmlessly.
		_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN discovery_completed BOOLEAN DEFAULT 0`)

		// Migration: Add script_path column to feeds table for custom script support
		// Error is ignored - if column exists, the operation fails harmlessly.
		_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN script_path TEXT DEFAULT ''`)

		// Migration: Add hide_from_timeline column to feeds table
		// Error is ignored - if column exists, the operation fails harmlessly.
		_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN hide_from_timeline BOOLEAN DEFAULT 0`)

		// Migration: Add proxy and refresh interval columns to feeds table
		// Error is ignored - if column exists, the operation fails harmlessly.
		_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN proxy_url TEXT DEFAULT ''`)
		_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN proxy_enabled BOOLEAN DEFAULT 0`)
		_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN refresh_interval INTEGER DEFAULT 0`)
	})
	return err
}

// WaitForReady blocks until the database is initialized.
func (db *DB) WaitForReady() {
	<-db.ready
}

func initSchema(db *sql.DB) error {
	// First, run migrations to ensure all columns exist
	// This must happen BEFORE creating indexes that depend on those columns
	if err := runMigrations(db); err != nil {
		return err
	}

	query := `
	CREATE TABLE IF NOT EXISTS feeds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		url TEXT UNIQUE,
		link TEXT DEFAULT '',
		description TEXT,
		category TEXT DEFAULT '',
		image_url TEXT DEFAULT '',
		last_updated DATETIME,
		last_error TEXT DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS articles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		feed_id INTEGER,
		title TEXT,
		url TEXT UNIQUE,
		image_url TEXT,
		audio_url TEXT DEFAULT '',
		video_url TEXT DEFAULT '',
		translated_title TEXT,
		published_at DATETIME,
		is_read BOOLEAN DEFAULT 0,
		is_favorite BOOLEAN DEFAULT 0,
		is_hidden BOOLEAN DEFAULT 0,
		is_read_later BOOLEAN DEFAULT 0,
		FOREIGN KEY(feed_id) REFERENCES feeds(id)
	);

	-- Create indexes for better query performance
	CREATE INDEX IF NOT EXISTS idx_articles_feed_id ON articles(feed_id);
	CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_articles_is_read ON articles(is_read);
	CREATE INDEX IF NOT EXISTS idx_articles_is_favorite ON articles(is_favorite);
	CREATE INDEX IF NOT EXISTS idx_articles_is_hidden ON articles(is_hidden);
	CREATE INDEX IF NOT EXISTS idx_articles_is_read_later ON articles(is_read_later);
	CREATE INDEX IF NOT EXISTS idx_feeds_category ON feeds(category);

	-- Composite indexes for common query patterns
	CREATE INDEX IF NOT EXISTS idx_articles_feed_published ON articles(feed_id, published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_articles_read_published ON articles(is_read, published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_articles_fav_published ON articles(is_favorite, published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_articles_readlater_published ON articles(is_read_later, published_at DESC);
	`
	_, err := db.Exec(query)
	return err
}

// runMigrations applies database migrations for existing databases
func runMigrations(db *sql.DB) error {
	// Migration: Add content and is_hidden columns if they don't exist
	// SQLite doesn't support IF NOT EXISTS for ALTER TABLE, so we ignore errors if columns already exist
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN content TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN is_hidden BOOLEAN DEFAULT 0`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN last_error TEXT DEFAULT ''`)

	// Migration: Add is_read_later column for read later feature
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN is_read_later BOOLEAN DEFAULT 0`)

	// Migration: Add audio_url column for podcast support
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN audio_url TEXT DEFAULT ''`)

	// Migration: Add video_url column for YouTube video support
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN video_url TEXT DEFAULT ''`)

	return nil
}

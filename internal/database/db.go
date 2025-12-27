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
		// Note: settingsKeys is auto-generated from settings_schema.json
		settingsKeys := config.SettingsKeys()
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

		// Migration: Add is_image_mode column to feeds table for image gallery feature
		// Error is ignored - if column exists, the operation fails harmlessly.
		_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN is_image_mode BOOLEAN DEFAULT 0`)

		// Migration: Add position column to feeds table for custom ordering
		// Error is ignored - if column exists, the operation fails harmlessly.
		_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN position INTEGER DEFAULT 0`)

		// Migration: Add article_view_mode column to feeds table for per-feed view mode override
		// Error is ignored - if column exists, the operation fails harmlessly.
		_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN article_view_mode TEXT DEFAULT 'global'`)

		// Migration: Add auto_expand_content column to feeds table for per-feed content expansion override
		// Error is ignored - if column exists, the operation fails harmlessly.
		_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN auto_expand_content TEXT DEFAULT 'global'`)
	})
	return err
}

// WaitForReady blocks until the database is initialized.
func (db *DB) WaitForReady() {
	<-db.ready
}

func initSchema(db *sql.DB) error {
	// First create tables
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

	-- Translation cache table to avoid redundant API calls
	CREATE TABLE IF NOT EXISTS translation_cache (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_text_hash TEXT NOT NULL,
		source_text TEXT NOT NULL,
		target_lang TEXT NOT NULL,
		translated_text TEXT NOT NULL,
		provider TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(source_text_hash, target_lang, provider)
	);

	-- Article content cache table to store full article content
	CREATE TABLE IF NOT EXISTS article_contents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_id INTEGER NOT NULL UNIQUE,
		content TEXT NOT NULL,
		fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
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

	-- Translation cache index
	CREATE INDEX IF NOT EXISTS idx_translation_cache_lookup ON translation_cache(source_text_hash, target_lang, provider);

	-- Article content cache index
	CREATE INDEX IF NOT EXISTS idx_article_contents_article_id ON article_contents(article_id);
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	// Then run migrations to ensure all columns exist
	// This must happen AFTER creating tables
	if err := runMigrations(db); err != nil {
		return err
	}

	return nil
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

	// Migration: Add XPath support fields to feeds table
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN type TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_title TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_content TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_uri TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_author TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_timestamp TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_time_format TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_thumbnail TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_categories TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_uid TEXT DEFAULT ''`)

	// Migration: Add summary column for caching AI-generated summaries
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN summary TEXT DEFAULT ''`)

	// Migration: Add article_contents table for caching article content
	// This uses a separate table to keep the articles table lightweight
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS article_contents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_id INTEGER NOT NULL UNIQUE,
		content TEXT NOT NULL,
		fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
	)`)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_article_contents_article_id ON article_contents(article_id)`)

	return nil
}

// TranslationCache represents a cached translation entry
type TranslationCache struct {
	ID             int64
	SourceTextHash string
	SourceText     string
	TargetLang     string
	TranslatedText string
	Provider       string
	CreatedAt      string
}

// GetCachedTranslation retrieves a translation from cache if available
func (db *DB) GetCachedTranslation(sourceTextHash, targetLang, provider string) (string, bool, error) {
	var translatedText string
	err := db.QueryRow(
		`SELECT translated_text FROM translation_cache
		 WHERE source_text_hash = ? AND target_lang = ? AND provider = ?`,
		sourceTextHash, targetLang, provider,
	).Scan(&translatedText)

	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return translatedText, true, nil
}

// SetCachedTranslation stores a translation in cache
func (db *DB) SetCachedTranslation(sourceTextHash, sourceText, targetLang, translatedText, provider string) error {
	_, err := db.Exec(
		`INSERT OR REPLACE INTO translation_cache
		 (source_text_hash, source_text, target_lang, translated_text, provider, created_at)
		 VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		sourceTextHash, sourceText, targetLang, translatedText, provider,
	)
	return err
}

// CleanupTranslationCache removes cached translations older than maxAgeDays
func (db *DB) CleanupTranslationCache(maxAgeDays int) (int64, error) {
	result, err := db.Exec(
		`DELETE FROM translation_cache WHERE created_at < datetime('now', ?)`,
		fmt.Sprintf("-%d days", maxAgeDays),
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

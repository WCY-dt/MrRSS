package database

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"MrRSS/internal/models"

	_ "modernc.org/sqlite"
)

type DB struct {
	*sql.DB
	ready chan struct{}
	once  sync.Once
}

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
		
		// Insert default settings if they don't exist
		_, _ = db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES ('update_interval', '10')`)
		_, _ = db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES ('translation_enabled', 'false')`)
		_, _ = db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES ('target_language', 'zh')`)
		_, _ = db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES ('translation_provider', 'google')`)
		_, _ = db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES ('deepl_api_key', '')`)
		_, _ = db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES ('auto_cleanup_enabled', 'false')`)
		_, _ = db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES ('max_cache_size_mb', '20')`)
		_, _ = db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES ('max_article_age_days', '30')`)
		_, _ = db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES ('language', 'en')`)
		_, _ = db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES ('theme', 'auto')`)
	})
	return err
}

func (db *DB) WaitForReady() {
	<-db.ready
}

func initSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS feeds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		url TEXT UNIQUE,
		description TEXT,
		category TEXT DEFAULT '',
		image_url TEXT DEFAULT '',
		last_updated DATETIME
	);

	CREATE TABLE IF NOT EXISTS articles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		feed_id INTEGER,
		title TEXT,
		url TEXT UNIQUE,
		image_url TEXT,
		translated_title TEXT,
		published_at DATETIME,
		is_read BOOLEAN DEFAULT 0,
		is_favorite BOOLEAN DEFAULT 0,
		FOREIGN KEY(feed_id) REFERENCES feeds(id)
	);

	-- Create indexes for better query performance
	CREATE INDEX IF NOT EXISTS idx_articles_feed_id ON articles(feed_id);
	CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_articles_is_read ON articles(is_read);
	CREATE INDEX IF NOT EXISTS idx_articles_is_favorite ON articles(is_favorite);
	CREATE INDEX IF NOT EXISTS idx_feeds_category ON feeds(category);
	
	-- Composite indexes for common query patterns
	CREATE INDEX IF NOT EXISTS idx_articles_feed_published ON articles(feed_id, published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_articles_read_published ON articles(is_read, published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_articles_fav_published ON articles(is_favorite, published_at DESC);
	`
	_, err := db.Exec(query)
	return err
}

func (db *DB) AddFeed(feed *models.Feed) error {
	db.WaitForReady()
	query := `INSERT OR IGNORE INTO feeds (title, url, description, category, image_url, last_updated) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, feed.Title, feed.URL, feed.Description, feed.Category, feed.ImageURL, time.Now())
	return err
}

func (db *DB) DeleteFeed(id int64) error {
	db.WaitForReady()
	// First delete associated articles
	_, err := db.Exec("DELETE FROM articles WHERE feed_id = ?", id)
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM feeds WHERE id = ?", id)
	return err
}

func (db *DB) GetFeeds() ([]models.Feed, error) {
	db.WaitForReady()
	rows, err := db.Query("SELECT id, title, url, description, category, image_url, last_updated FROM feeds")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []models.Feed
	for rows.Next() {
		var f models.Feed
		var category, imageURL sql.NullString
		if err := rows.Scan(&f.ID, &f.Title, &f.URL, &f.Description, &category, &imageURL, &f.LastUpdated); err != nil {
			return nil, err
		}
		f.Category = category.String
		f.ImageURL = imageURL.String
		feeds = append(feeds, f)
	}
	return feeds, nil
}

func (db *DB) SaveArticle(article *models.Article) error {
	db.WaitForReady()
	query := `INSERT OR IGNORE INTO articles (feed_id, title, url, image_url, published_at, translated_title, is_read, is_favorite) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, article.FeedID, article.Title, article.URL, article.ImageURL, article.PublishedAt, article.TranslatedTitle, article.IsRead, article.IsFavorite)
	return err
}

func (db *DB) SaveArticles(ctx context.Context, articles []*models.Article) error {
	db.WaitForReady()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO articles (feed_id, title, url, image_url, published_at, translated_title, is_read, is_favorite) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, article := range articles {
		// Check context before each insert
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		_, err := stmt.ExecContext(ctx, article.FeedID, article.Title, article.URL, article.ImageURL, article.PublishedAt, article.TranslatedTitle, article.IsRead, article.IsFavorite)
		if err != nil {
			log.Println("Error saving article in batch:", err)
			// Continue even if one fails
		}
	}

	return tx.Commit()
}

func (db *DB) GetArticles(filter string, feedID int64, category string, limit, offset int) ([]models.Article, error) {
	db.WaitForReady()
	baseQuery := `
		SELECT a.id, a.feed_id, a.title, a.url, a.image_url, a.published_at, a.is_read, a.is_favorite, a.translated_title, f.title 
		FROM articles a 
		JOIN feeds f ON a.feed_id = f.id 
	`
	var args []interface{}
	whereClauses := []string{}

	if filter == "unread" {
		whereClauses = append(whereClauses, "a.is_read = 0")
	} else if filter == "favorites" {
		whereClauses = append(whereClauses, "a.is_favorite = 1")
	}

	if feedID > 0 {
		whereClauses = append(whereClauses, "a.feed_id = ?")
		args = append(args, feedID)
	}

	if category != "" {
		// Simple prefix match for category hierarchy
		whereClauses = append(whereClauses, "(f.category = ? OR f.category LIKE ?)")
		args = append(args, category, category+"/%")
	}

	query := baseQuery
	if len(whereClauses) > 0 {
		query += " WHERE " + whereClauses[0]
		for i := 1; i < len(whereClauses); i++ {
			query += " AND " + whereClauses[i]
		}
	}
	query += " ORDER BY a.published_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var a models.Article
		var imageURL, translatedTitle sql.NullString
		if err := rows.Scan(&a.ID, &a.FeedID, &a.Title, &a.URL, &imageURL, &a.PublishedAt, &a.IsRead, &a.IsFavorite, &translatedTitle, &a.FeedTitle); err != nil {
			log.Println("Error scanning article:", err)
			continue
		}
		a.ImageURL = imageURL.String
		a.TranslatedTitle = translatedTitle.String
		articles = append(articles, a)
	}
	return articles, nil
}

func (db *DB) UpdateFeed(id int64, title, url, category string) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE feeds SET title = ?, url = ?, category = ? WHERE id = ?", title, url, category, id)
	return err
}

func (db *DB) UpdateFeedCategory(id int64, category string) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE feeds SET category = ? WHERE id = ?", category, id)
	return err
}

func (db *DB) UpdateFeedImage(id int64, imageURL string) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE feeds SET image_url = ? WHERE id = ?", imageURL, id)
	return err
}

func (db *DB) MarkArticleRead(id int64, read bool) error {
	db.WaitForReady()
	isRead := 0
	if read {
		isRead = 1
	}
	_, err := db.Exec("UPDATE articles SET is_read = ? WHERE id = ?", isRead, id)
	return err
}

func (db *DB) ToggleFavorite(id int64) error {
	db.WaitForReady()
	// First get current state
	var isFav bool
	err := db.QueryRow("SELECT is_favorite FROM articles WHERE id = ?", id).Scan(&isFav)
	if err != nil {
		return err
	}
	_, err = db.Exec("UPDATE articles SET is_favorite = ? WHERE id = ?", !isFav, id)
	return err
}

func (db *DB) GetSetting(key string) (string, error) {
	db.WaitForReady()
	var value string
	err := db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
}

func (db *DB) SetSetting(key, value string) error {
	db.WaitForReady()
	_, err := db.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", key, value)
	return err
}

// CleanupOldArticles removes articles based on age and status
// - Articles older than configured days: delete except read OR favorited
// - Also checks database size against max_cache_size_mb setting
func (db *DB) CleanupOldArticles() (int64, error) {
	db.WaitForReady()
	
	// Get max article age from settings (default 30 days)
	maxAgeDaysStr, err := db.GetSetting("max_article_age_days")
	maxAgeDays := 30
	if err == nil {
		if days, err := strconv.Atoi(maxAgeDaysStr); err == nil && days > 0 {
			maxAgeDays = days
		}
	}
	
	cutoffDate := time.Now().AddDate(0, 0, -maxAgeDays)
	
	// Delete articles older than configured age that are not favorited
	result, err := db.Exec(`
		DELETE FROM articles 
		WHERE published_at < ? 
		AND is_favorite = 0
	`, cutoffDate)
	if err != nil {
		return 0, err
	}
	
	count, _ := result.RowsAffected()
	
	// Run VACUUM to reclaim space
	_, _ = db.Exec("VACUUM")
	
	return count, nil
}

// CleanupUnimportantArticles removes all articles except read and favorited ones
func (db *DB) CleanupUnimportantArticles() (int64, error) {
	db.WaitForReady()
	
	result, err := db.Exec(`
		DELETE FROM articles 
		WHERE is_read = 0 
		AND is_favorite = 0
	`)
	if err != nil {
		return 0, err
	}
	
	count, _ := result.RowsAffected()
	
	// Run VACUUM to reclaim space
	_, _ = db.Exec("VACUUM")
	
	return count, nil
}

// UpdateArticleTranslation updates the translated_title field for an article
func (db *DB) UpdateArticleTranslation(id int64, translatedTitle string) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET translated_title = ? WHERE id = ?", translatedTitle, id)
	return err
}

// GetDatabaseSizeMB returns the current database size in megabytes
func (db *DB) GetDatabaseSizeMB() (float64, error) {
	db.WaitForReady()
	
	var pageCount, pageSize int64
	err := db.QueryRow("PRAGMA page_count").Scan(&pageCount)
	if err != nil {
		return 0, err
	}
	
	err = db.QueryRow("PRAGMA page_size").Scan(&pageSize)
	if err != nil {
		return 0, err
	}
	
	sizeBytes := pageCount * pageSize
	sizeMB := float64(sizeBytes) / (1024 * 1024)
	
	return sizeMB, nil
}

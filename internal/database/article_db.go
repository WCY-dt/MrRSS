package database

import (
	"context"
	"database/sql"
	"log"

	"MrRSS/internal/models"
)

// SaveArticle saves a single article to the database.
func (db *DB) SaveArticle(article *models.Article) error {
	db.WaitForReady()
	query := `INSERT OR IGNORE INTO articles (feed_id, title, url, image_url, audio_url, video_url, published_at, translated_title, is_read, is_favorite, is_hidden, is_read_later) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, article.FeedID, article.Title, article.URL, article.ImageURL, article.AudioURL, article.VideoURL, article.PublishedAt, article.TranslatedTitle, article.IsRead, article.IsFavorite, article.IsHidden, article.IsReadLater)
	return err
}

// SaveArticles saves multiple articles in a transaction.
func (db *DB) SaveArticles(ctx context.Context, articles []*models.Article) error {
	db.WaitForReady()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO articles (feed_id, title, url, image_url, audio_url, video_url, published_at, translated_title, is_read, is_favorite, is_hidden, is_read_later) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
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

		_, err := stmt.ExecContext(ctx, article.FeedID, article.Title, article.URL, article.ImageURL, article.AudioURL, article.VideoURL, article.PublishedAt, article.TranslatedTitle, article.IsRead, article.IsFavorite, article.IsHidden, article.IsReadLater)
		if err != nil {
			log.Println("Error saving article in batch:", err)
			// Continue even if one fails
		}
	}

	return tx.Commit()
}

// GetArticles retrieves articles with filtering, pagination, and sorting.
func (db *DB) GetArticles(filter string, feedID int64, category string, showHidden bool, limit, offset int) ([]models.Article, error) {
	db.WaitForReady()
	baseQuery := `
		SELECT a.id, a.feed_id, a.title, a.url, a.image_url, a.audio_url, a.video_url, a.published_at, a.is_read, a.is_favorite, a.is_hidden, a.is_read_later, a.translated_title, f.title
		FROM articles a
		JOIN feeds f ON a.feed_id = f.id
	`
	var args []interface{}
	whereClauses := []string{}

	// Always filter hidden articles unless showHidden is true
	if !showHidden {
		whereClauses = append(whereClauses, "a.is_hidden = 0")
	}

	switch filter {
	case "unread":
		whereClauses = append(whereClauses, "a.is_read = 0")
		// Exclude feeds marked as hide_from_timeline when viewing unread (unless specific feed/category selected)
		if feedID <= 0 && category == "" {
			whereClauses = append(whereClauses, "COALESCE(f.hide_from_timeline, 0) = 0")
		}
	case "favorites":
		whereClauses = append(whereClauses, "a.is_favorite = 1")
	case "readLater":
		whereClauses = append(whereClauses, "a.is_read_later = 1")
	case "all":
		// Exclude feeds marked as hide_from_timeline when viewing all articles (unless specific feed/category selected)
		if feedID <= 0 && category == "" {
			whereClauses = append(whereClauses, "COALESCE(f.hide_from_timeline, 0) = 0")
		}
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
		var imageURL, audioURL, videoURL, translatedTitle sql.NullString
		if err := rows.Scan(&a.ID, &a.FeedID, &a.Title, &a.URL, &imageURL, &audioURL, &videoURL, &a.PublishedAt, &a.IsRead, &a.IsFavorite, &a.IsHidden, &a.IsReadLater, &translatedTitle, &a.FeedTitle); err != nil {
			log.Println("Error scanning article:", err)
			continue
		}
		a.ImageURL = imageURL.String
		a.AudioURL = audioURL.String
		a.VideoURL = videoURL.String
		a.TranslatedTitle = translatedTitle.String
		articles = append(articles, a)
	}
	return articles, nil
}

// GetArticleByID retrieves a single article by its ID.
// This is more efficient than GetArticles when you only need one article.
func (db *DB) GetArticleByID(id int64) (*models.Article, error) {
	db.WaitForReady()
	query := `
		SELECT a.id, a.feed_id, a.title, a.url, a.image_url, a.audio_url, a.video_url, a.published_at, a.is_read, a.is_favorite, a.is_hidden, a.is_read_later, a.translated_title, f.title
		FROM articles a
		JOIN feeds f ON a.feed_id = f.id
		WHERE a.id = ?
	`
	row := db.QueryRow(query, id)

	var a models.Article
	var imageURL, audioURL, videoURL, translatedTitle sql.NullString
	if err := row.Scan(&a.ID, &a.FeedID, &a.Title, &a.URL, &imageURL, &audioURL, &videoURL, &a.PublishedAt, &a.IsRead, &a.IsFavorite, &a.IsHidden, &a.IsReadLater, &translatedTitle, &a.FeedTitle); err != nil {
		return nil, err
	}
	a.ImageURL = imageURL.String
	a.AudioURL = audioURL.String
	a.VideoURL = videoURL.String
	a.TranslatedTitle = translatedTitle.String
	return &a, nil
}

// MarkArticleRead marks an article as read or unread.
// When marking as read, also removes from read later list.
func (db *DB) MarkArticleRead(id int64, read bool) error {
	db.WaitForReady()
	isRead := 0
	if read {
		isRead = 1
		// When marking as read, also remove from read later
		_, err := db.Exec("UPDATE articles SET is_read = 1, is_read_later = 0 WHERE id = ?", id)
		return err
	}
	_, err := db.Exec("UPDATE articles SET is_read = ? WHERE id = ?", isRead, id)
	return err
}

// ToggleFavorite toggles the favorite status of an article.
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

// SetArticleFavorite sets the favorite status of an article.
func (db *DB) SetArticleFavorite(id int64, favorite bool) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET is_favorite = ? WHERE id = ?", favorite, id)
	return err
}

// UpdateArticleTranslation updates the translated_title field for an article.
func (db *DB) UpdateArticleTranslation(id int64, translatedTitle string) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET translated_title = ? WHERE id = ?", translatedTitle, id)
	return err
}

// ClearAllTranslations clears all translated titles from articles.
func (db *DB) ClearAllTranslations() error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET translated_title = ''")
	return err
}

// ToggleArticleHidden toggles the is_hidden status of an article.
func (db *DB) ToggleArticleHidden(id int64) error {
	db.WaitForReady()
	// First get current state
	var isHidden bool
	err := db.QueryRow("SELECT is_hidden FROM articles WHERE id = ?", id).Scan(&isHidden)
	if err != nil {
		return err
	}
	_, err = db.Exec("UPDATE articles SET is_hidden = ? WHERE id = ?", !isHidden, id)
	return err
}

// SetArticleHidden sets the hidden status of an article.
func (db *DB) SetArticleHidden(id int64, hidden bool) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET is_hidden = ? WHERE id = ?", hidden, id)
	return err
}

// ToggleReadLater toggles the read later status of an article.
// When adding to read later, also marks article as unread.
func (db *DB) ToggleReadLater(id int64) error {
	db.WaitForReady()
	// First get current state
	var isReadLater bool
	err := db.QueryRow("SELECT is_read_later FROM articles WHERE id = ?", id).Scan(&isReadLater)
	if err != nil {
		return err
	}
	newState := !isReadLater
	// If adding to read later, also mark as unread
	if newState {
		_, err = db.Exec("UPDATE articles SET is_read_later = 1, is_read = 0 WHERE id = ?", id)
	} else {
		_, err = db.Exec("UPDATE articles SET is_read_later = 0 WHERE id = ?", id)
	}
	return err
}

// SetArticleReadLater sets the read later status of an article.
// When adding to read later, also marks article as unread.
func (db *DB) SetArticleReadLater(id int64, readLater bool) error {
	db.WaitForReady()
	// If adding to read later, also mark as unread
	if readLater {
		_, err := db.Exec("UPDATE articles SET is_read_later = 1, is_read = 0 WHERE id = ?", id)
		return err
	}
	_, err := db.Exec("UPDATE articles SET is_read_later = 0 WHERE id = ?", id)
	return err
}

// UpdateArticleContent updates the content field for an article.
func (db *DB) UpdateArticleContent(id int64, content string) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET content = ? WHERE id = ?", content, id)
	return err
}

// GetTotalUnreadCount returns the total number of unread articles.
func (db *DB) GetTotalUnreadCount() (int, error) {
	db.WaitForReady()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM articles WHERE is_read = 0 AND is_hidden = 0").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetUnreadCountByFeed returns the number of unread articles for a specific feed.
func (db *DB) GetUnreadCountByFeed(feedID int64) (int, error) {
	db.WaitForReady()
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM articles WHERE feed_id = ? AND is_read = 0 AND is_hidden = 0", feedID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetUnreadCountsForAllFeeds returns a map of feed_id to unread count.
func (db *DB) GetUnreadCountsForAllFeeds() (map[int64]int, error) {
	db.WaitForReady()
	rows, err := db.Query(`
		SELECT feed_id, COUNT(*)
		FROM articles
		WHERE is_read = 0 AND is_hidden = 0
		GROUP BY feed_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[int64]int)
	for rows.Next() {
		var feedID int64
		var count int
		if err := rows.Scan(&feedID, &count); err != nil {
			log.Println("Error scanning unread count:", err)
			continue
		}
		counts[feedID] = count
	}
	return counts, rows.Err()
}

// MarkAllAsReadForFeed marks all articles in a feed as read.
func (db *DB) MarkAllAsReadForFeed(feedID int64) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET is_read = 1 WHERE feed_id = ? AND is_hidden = 0", feedID)
	return err
}

// MarkAllAsRead marks all articles as read.
func (db *DB) MarkAllAsRead() error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET is_read = 1 WHERE is_hidden = 0")
	return err
}

// MarkAllAsReadForCategory marks all articles in a category as read.
func (db *DB) MarkAllAsReadForCategory(category string) error {
	db.WaitForReady()
	// Get all feed IDs in this category
	// Handle empty category (uncategorized) by matching NULL or empty string
	var query string
	if category == "" {
		query = `UPDATE articles SET is_read = 1
			WHERE feed_id IN (SELECT id FROM feeds WHERE category IS NULL OR category = '') AND is_hidden = 0`
		_, err := db.Exec(query)
		return err
	}
	query = `UPDATE articles SET is_read = 1
		WHERE feed_id IN (SELECT id FROM feeds WHERE category = ?) AND is_hidden = 0`
	_, err := db.Exec(query, category)
	return err
}

// ClearReadLater removes all articles from the read later list.
func (db *DB) ClearReadLater() error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET is_read_later = 0 WHERE is_read_later = 1")
	return err
}

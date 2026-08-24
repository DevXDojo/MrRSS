package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"MrRSS/internal/models"
)

// GetImageGalleryArticles retrieves articles from image mode feeds with pagination.
// If feedID is provided, it gets articles only from that feed (assuming it's an image mode feed).
// If category is provided, it gets articles from all image mode feeds in that category.
// Otherwise, it gets articles from all image mode feeds.
// If onlyUnread is true, only returns unread articles.
func (db *DB) GetImageGalleryArticles(feedID int64, category string, showHidden bool, onlyUnread bool, limit, offset int) ([]models.Article, error) {
	db.WaitForReady()
	baseQuery := `
		SELECT a.id, a.feed_id, a.title, a.url, a.image_url, a.audio_url, a.video_url, a.published_at, a.is_read, a.is_favorite, a.is_hidden, a.is_read_later, a.translated_title, a.summary, f.title, a.author
		FROM articles a
		JOIN feeds f ON a.feed_id = f.id
		WHERE COALESCE(f.is_image_mode, 0) = 1
	`
	var args []interface{}

	// Always filter hidden articles unless showHidden is true
	if !showHidden {
		baseQuery += " AND a.is_hidden = 0"
	}

	// Only get articles with image_url
	baseQuery += " AND a.image_url IS NOT NULL AND a.image_url != ''"

	// Filter for only unread articles if requested
	if onlyUnread {
		baseQuery += " AND a.is_read = 0"
	}

	if feedID > 0 {
		baseQuery += " AND a.feed_id = ?"
		args = append(args, feedID)
	} else if category == "\x00" {
		// Special value "\x00" means explicit uncategorized filtering
		baseQuery += " AND (f.category IS NULL OR f.category = '')"
	} else if category != "" {
		// For categories, use prefix match to support nested categories
		baseQuery += " AND (f.category = ? OR f.category LIKE ?)"
		args = append(args, category, category+"/%")
	}
	// Note: When category is empty string, it means no category filter was provided,
	// so we should not filter by category at all (show all image mode articles from all categories).

	baseQuery += " ORDER BY a.published_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]models.Article, 0)
	for rows.Next() {
		var a models.Article
		var imageURL, audioURL, videoURL, translatedTitle, summary, author sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.FeedID, &a.Title, &a.URL, &imageURL, &audioURL, &videoURL, &publishedAt, &a.IsRead, &a.IsFavorite, &a.IsHidden, &a.IsReadLater, &translatedTitle, &summary, &a.FeedTitle, &author); err != nil {
			log.Println("Error scanning article:", err)
			continue
		}
		a.ImageURL = imageURL.String
		a.AudioURL = audioURL.String
		a.VideoURL = videoURL.String
		if publishedAt.Valid {
			a.PublishedAt = publishedAt.Time
		} else {
			a.PublishedAt = time.Time{}
		}
		a.TranslatedTitle = translatedTitle.String
		a.Summary = summary.String
		a.Author = author.String
		articles = append(articles, a)
	}
	return articles, nil
}

// SearchArticlesWithAI executes a search query with an AI-generated WHERE clause.
// The whereClause should be pre-validated to prevent SQL injection.
func (db *DB) SearchArticlesWithAI(whereClause string, limit int) ([]models.Article, error) {
	db.WaitForReady()

	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	// Build the complete query with the AI-generated WHERE clause
	query := fmt.Sprintf(`
		SELECT a.id, a.feed_id, a.title, a.url, a.image_url, a.audio_url, a.video_url,
			   a.published_at, a.is_read, a.is_favorite, a.is_hidden, a.is_read_later,
			   a.translated_title, a.summary, a.freshrss_item_id, f.title, a.author
		FROM articles a
		JOIN feeds f ON a.feed_id = f.id
		WHERE %s
		ORDER BY a.published_at DESC
		LIMIT %d
	`, whereClause, limit)

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var a models.Article
		var imageURL, audioURL, videoURL, translatedTitle, summary, freshrssItemID, author sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.FeedID, &a.Title, &a.URL, &imageURL, &audioURL, &videoURL, &publishedAt, &a.IsRead, &a.IsFavorite, &a.IsHidden, &a.IsReadLater, &translatedTitle, &summary, &freshrssItemID, &a.FeedTitle, &author); err != nil {
			log.Println("Error scanning article in AI search:", err)
			continue
		}
		a.ImageURL = imageURL.String
		a.AudioURL = audioURL.String
		a.VideoURL = videoURL.String
		if publishedAt.Valid {
			a.PublishedAt = publishedAt.Time
		} else {
			a.PublishedAt = time.Time{}
		}
		a.TranslatedTitle = translatedTitle.String
		a.Summary = summary.String
		a.FreshRSSItemID = freshrssItemID.String
		a.Author = author.String
		articles = append(articles, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}

	return articles, nil
}

// SearchArticlesWithSQL executes a complete SQL query for AI search with content and relevance scoring.
// The query should include all necessary SELECT fields and be pre-validated.
func (db *DB) SearchArticlesWithSQL(query string) ([]models.Article, error) {
	db.WaitForReady()

	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var a models.Article
		var imageURL, audioURL, videoURL, translatedTitle, summary, freshrssItemID, author sql.NullString
		var publishedAt sql.NullTime
		var relevanceScore float64
		if err := rows.Scan(&a.ID, &a.FeedID, &a.Title, &a.URL, &imageURL, &audioURL, &videoURL, &publishedAt, &a.IsRead, &a.IsFavorite, &a.IsHidden, &a.IsReadLater, &translatedTitle, &summary, &freshrssItemID, &a.FeedTitle, &author, &relevanceScore); err != nil {
			log.Println("Error scanning article in AI search with SQL:", err)
			continue
		}
		a.ImageURL = imageURL.String
		a.AudioURL = audioURL.String
		a.VideoURL = videoURL.String
		if publishedAt.Valid {
			a.PublishedAt = publishedAt.Time
		} else {
			a.PublishedAt = time.Time{}
		}
		a.TranslatedTitle = translatedTitle.String
		a.Summary = summary.String
		a.FreshRSSItemID = freshrssItemID.String
		a.Author = author.String
		articles = append(articles, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration failed: %w", err)
	}

	return articles, nil
}

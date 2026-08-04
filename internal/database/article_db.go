package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"MrRSS/internal/models"
	"MrRSS/internal/utils/urlutil"
)

// SaveArticle saves a single article to the database.
func (db *DB) SaveArticle(article *models.Article) error {
	db.WaitForReady()

	// Generate unique_id for deduplication
	uniqueID := urlutil.GenerateArticleUniqueID(article.Title, article.FeedID, article.PublishedAt, article.HasValidPublishedTime)
	query := `INSERT OR IGNORE INTO articles (feed_id, title, url, image_url, audio_url, video_url, published_at, translated_title, is_read, is_favorite, is_hidden, is_read_later, summary, original_summary, unique_id, author) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, article.FeedID, article.Title, article.URL, article.ImageURL, article.AudioURL, article.VideoURL, article.PublishedAt, article.TranslatedTitle, article.IsRead, article.IsFavorite, article.IsHidden, article.IsReadLater, article.Summary, article.OriginalSummary, uniqueID, article.Author)
	return err
}

// SaveArticles saves multiple articles in a transaction.
// Includes progressive cleanup check to prevent database from exceeding size limit during refresh.
func (db *DB) SaveArticles(ctx context.Context, articles []*models.Article) error {
	db.WaitForReady()

	// Progressive cleanup: check if we need to clean up before saving
	if len(articles) > 10 {
		// Only check for larger batches to avoid overhead
		shouldCleanup, _ := db.ShouldCleanupBeforeSave()
		if shouldCleanup {
			log.Printf("Database approaching size limit, running progressive cleanup...")
			go func() {
				deleted, err := db.CleanupBySize()
				if err != nil {
					log.Printf("Progressive cleanup error: %v", err)
				} else if deleted > 0 {
					log.Printf("Progressive cleanup removed %d articles", deleted)
				}
			}()
		}
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Use DO UPDATE instead of INSERT OR REPLACE because REPLACE deletes the
	// existing row before inserting a new one. With foreign keys enabled, that
	// implicit delete would cascade to article contents and chat history.
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO articles (
			feed_id, title, url, image_url, audio_url, video_url, published_at,
			translated_title, is_read, is_favorite, is_hidden, is_read_later,
			summary, original_summary, unique_id, author
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(unique_id) DO UPDATE SET
			feed_id = excluded.feed_id,
			title = excluded.title,
			url = excluded.url,
			image_url = excluded.image_url,
			audio_url = excluded.audio_url,
			video_url = excluded.video_url,
			published_at = CASE WHEN ? THEN excluded.published_at ELSE articles.published_at END,
			translated_title = excluded.translated_title,
			is_read = excluded.is_read,
			is_favorite = excluded.is_favorite,
			is_hidden = excluded.is_hidden,
			is_read_later = excluded.is_read_later,
			original_summary = excluded.original_summary,
			author = excluded.author
	`)
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

		// Generate unique_id for deduplication
		uniqueID := urlutil.GenerateArticleUniqueID(article.Title, article.FeedID, article.PublishedAt, article.HasValidPublishedTime)

		// Fetch the existing row to preserve user-controlled status fields and to
		// detect whether the article actually changed (for no-pubDate feeds).
		var existingIsRead, existingIsFavorite, existingIsHidden, existingIsReadLater int
		var existingTitle, existingURL, existingImageURL, existingAudioURL, existingVideoURL, existingOriginalSummary, existingAuthor string
		err := tx.QueryRowContext(ctx, `
			SELECT is_read, is_favorite, is_hidden, is_read_later,
				COALESCE(title, ''), COALESCE(url, ''), COALESCE(image_url, ''),
				COALESCE(audio_url, ''), COALESCE(video_url, ''),
				COALESCE(original_summary, ''), COALESCE(author, '')
			FROM articles WHERE unique_id = ?`, uniqueID).Scan(
			&existingIsRead, &existingIsFavorite, &existingIsHidden, &existingIsReadLater,
			&existingTitle, &existingURL, &existingImageURL, &existingAudioURL, &existingVideoURL,
			&existingOriginalSummary, &existingAuthor)
		isRead := article.IsRead
		isFavorite := article.IsFavorite
		isHidden := article.IsHidden
		isReadLater := article.IsReadLater
		// Articles with a real pubDate always update their published_at.
		updatePublishedAt := article.HasValidPublishedTime
		if err == nil {
			// Article exists, preserve its status
			isRead = existingIsRead == 1
			isFavorite = existingIsFavorite == 1
			isHidden = existingIsHidden == 1
			isReadLater = existingIsReadLater == 1
			// For articles without a pubDate, only move published_at to the current
			// time when the article actually changed (e.g. edited content); otherwise
			// keep the original time so it doesn't jump to the top on every refresh.
			articleChanged := article.Title != existingTitle ||
				article.URL != existingURL ||
				article.ImageURL != existingImageURL ||
				article.AudioURL != existingAudioURL ||
				article.VideoURL != existingVideoURL ||
				article.OriginalSummary != existingOriginalSummary ||
				article.Author != existingAuthor
			updatePublishedAt = article.HasValidPublishedTime || articleChanged
		}

		updateTime := 0
		if updatePublishedAt {
			updateTime = 1
		}
		_, err = stmt.ExecContext(ctx, article.FeedID, article.Title, article.URL, article.ImageURL, article.AudioURL, article.VideoURL, article.PublishedAt, article.TranslatedTitle, isRead, isFavorite, isHidden, isReadLater, article.Summary, article.OriginalSummary, uniqueID, article.Author, updateTime)
		if err != nil {
			log.Println("Error saving article in batch:", err)
			// Continue even if one fails
		}
	}

	return tx.Commit()
}

// GetArticles retrieves articles with filtering, pagination, and sorting.
// Optimized to filter feeds first for category queries, reducing JOIN overhead.
func (db *DB) GetArticles(filter string, feedID int64, category string, showHidden bool, limit, offset int) ([]models.Article, error) {
	return db.GetArticlesWithUnreadFilter(filter, feedID, category, showHidden, false, limit, offset)
}

// GetArticlesWithUnreadFilter returns articles with optional read-state filtering.
func (db *DB) GetArticlesWithUnreadFilter(filter string, feedID int64, category string, showHidden bool, onlyUnread bool, limit, offset int) ([]models.Article, error) {
	db.WaitForReady()

	// Optimization: For category queries, first get the feed IDs, then query articles
	// This avoids JOINing all articles and then filtering by category
	var feedIDFilter []int64
	var useFeedIDFilter bool

	if category != "" {
		var categoryQuery string
		var categoryArgs []interface{}

		if category == "\x00" {
			// Special value "\x00" means explicit uncategorized filtering
			categoryQuery = "SELECT id FROM feeds WHERE category IS NULL OR category = ''"
		} else {
			// Simple prefix match for category hierarchy
			categoryQuery = "SELECT id FROM feeds WHERE category = ? OR category LIKE ?"
			categoryArgs = []interface{}{category, category + "/%"}
		}

		rows, err := db.Query(categoryQuery, categoryArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to query feeds by category: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				log.Println("Error scanning feed ID:", err)
				continue
			}
			feedIDFilter = append(feedIDFilter, id)
		}

		// If no feeds found in this category, return empty result early
		if len(feedIDFilter) == 0 {
			return []models.Article{}, nil
		}

		useFeedIDFilter = true
	}

	// Build the main query
	baseQuery := `
		SELECT a.id, a.feed_id, a.title, a.url, a.image_url, a.audio_url, a.video_url, a.published_at, a.is_read, a.is_favorite, a.is_hidden, a.is_read_later, a.translated_title, a.summary, a.freshrss_item_id, f.title, a.author
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

	if onlyUnread && filter != "unread" {
		whereClauses = append(whereClauses, "a.is_read = 0")
	}

	// Apply feed ID filter
	if useFeedIDFilter {
		// Use optimized IN clause with pre-filtered feed IDs
		placeholders := make([]string, len(feedIDFilter))
		for i, id := range feedIDFilter {
			placeholders[i] = "?"
			args = append(args, id)
		}
		whereClauses = append(whereClauses, "a.feed_id IN ("+strings.Join(placeholders, ",")+")")
	} else if feedID > 0 {
		whereClauses = append(whereClauses, "a.feed_id = ?")
		args = append(args, feedID)
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
		var imageURL, audioURL, videoURL, translatedTitle, summary, freshrssItemID, author sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.FeedID, &a.Title, &a.URL, &imageURL, &audioURL, &videoURL, &publishedAt, &a.IsRead, &a.IsFavorite, &a.IsHidden, &a.IsReadLater, &translatedTitle, &summary, &freshrssItemID, &a.FeedTitle, &author); err != nil {
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
		a.FreshRSSItemID = freshrssItemID.String
		a.Author = author.String
		articles = append(articles, a)
	}
	return articles, nil
}

// GetArticleByID retrieves a single article by its ID.
// This is more efficient than GetArticles when you only need one article.
func (db *DB) GetArticleByID(id int64) (*models.Article, error) {
	db.WaitForReady()
	query := `
		SELECT a.id, a.feed_id, a.title, a.url, a.image_url, a.audio_url, a.video_url, a.published_at, a.is_read, a.is_favorite, a.is_hidden, a.is_read_later, a.translated_title, a.summary, a.freshrss_item_id, f.title, a.author
		FROM articles a
		JOIN feeds f ON a.feed_id = f.id
		WHERE a.id = ?
	`
	row := db.QueryRow(query, id)

	var a models.Article
	var imageURL, audioURL, videoURL, translatedTitle, summary, freshrssItemID, author sql.NullString
	var publishedAt sql.NullTime
	if err := row.Scan(&a.ID, &a.FeedID, &a.Title, &a.URL, &imageURL, &audioURL, &videoURL, &publishedAt, &a.IsRead, &a.IsFavorite, &a.IsHidden, &a.IsReadLater, &translatedTitle, &summary, &freshrssItemID, &a.FeedTitle, &author); err != nil {
		return nil, err
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
	return &a, nil
}

// GetArticlesByIDs retrieves multiple articles by their IDs
func (db *DB) GetArticlesByIDs(ids []int64) ([]models.Article, error) {
	db.WaitForReady()
	if len(ids) == 0 {
		return []models.Article{}, nil
	}

	// Build placeholder string for IN clause
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `
		SELECT a.id, a.feed_id, a.title, a.url, a.image_url, a.audio_url, a.video_url, a.published_at, a.is_read, a.is_favorite, a.is_hidden, a.is_read_later, a.translated_title, a.summary, a.freshrss_item_id, f.title, a.author
		FROM articles a
		JOIN feeds f ON a.feed_id = f.id
		WHERE a.id IN (` + strings.Join(placeholders, ",") + `)
	`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := []models.Article{}
	for rows.Next() {
		var a models.Article
		var imageURL, audioURL, videoURL, translatedTitle, summary, freshrssItemID, author sql.NullString
		var publishedAt sql.NullTime

		err := rows.Scan(&a.ID, &a.FeedID, &a.Title, &a.URL, &imageURL, &audioURL, &videoURL, &publishedAt, &a.IsRead, &a.IsFavorite, &a.IsHidden, &a.IsReadLater, &translatedTitle, &summary, &freshrssItemID, &a.FeedTitle, &author)
		if err != nil {
			return nil, err
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

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

// GetArticleIDByUniqueID retrieves an article's ID by its unique identifier.
// This is the preferred method for looking up articles as it uses the title+feed_id+published_date based deduplication.
// Note: Uses date only (YYYY-MM-DD) rather than full timestamp for better deduplication.
func (db *DB) GetArticleIDByUniqueID(title string, feedID int64, publishedAt time.Time, hasValidPublishedTime bool) (int64, error) {
	db.WaitForReady()
	uniqueID := urlutil.GenerateArticleUniqueID(title, feedID, publishedAt, hasValidPublishedTime)
	var id int64
	err := db.QueryRow("SELECT id FROM articles WHERE unique_id = ?", uniqueID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

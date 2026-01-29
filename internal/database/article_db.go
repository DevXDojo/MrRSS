package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"MrRSS/internal/models"
	"MrRSS/internal/utils"
)

// SaveArticle saves a single article to the database.
func (db *DB) SaveArticle(article *models.Article) error {
	db.WaitForReady()

	// Generate unique_id for deduplication
	uniqueID := utils.GenerateArticleUniqueID(article.Title, article.FeedID, article.PublishedAt, article.HasValidPublishedTime)
	query := `INSERT OR IGNORE INTO articles (feed_id, title, url, image_url, audio_url, video_url, published_at, translated_title, is_read, is_favorite, is_hidden, is_read_later, summary, unique_id, author) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, article.FeedID, article.Title, article.URL, article.ImageURL, article.AudioURL, article.VideoURL, article.PublishedAt, article.TranslatedTitle, article.IsRead, article.IsFavorite, article.IsHidden, article.IsReadLater, article.Summary, uniqueID, article.Author)
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

	stmt, err := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO articles (feed_id, title, url, image_url, audio_url, video_url, published_at, translated_title, is_read, is_favorite, is_hidden, is_read_later, summary, unique_id, author) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
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
		uniqueID := utils.GenerateArticleUniqueID(article.Title, article.FeedID, article.PublishedAt, article.HasValidPublishedTime)
		_, err := stmt.ExecContext(ctx, article.FeedID, article.Title, article.URL, article.ImageURL, article.AudioURL, article.VideoURL, article.PublishedAt, article.TranslatedTitle, article.IsRead, article.IsFavorite, article.IsHidden, article.IsReadLater, article.Summary, uniqueID, article.Author)
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

// ClearAllSummaries clears all summaries from articles.
func (db *DB) ClearAllSummaries() error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET summary = ''")
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

// GetFavoriteCountsForAllFeeds returns a map of feed_id to favorite article count.
func (db *DB) GetFavoriteCountsForAllFeeds() (map[int64]int, error) {
	db.WaitForReady()
	rows, err := db.Query(`
		SELECT feed_id, COUNT(*)
		FROM articles
		WHERE is_favorite = 1 AND is_hidden = 0
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
			log.Println("Error scanning favorite count:", err)
			continue
		}
		counts[feedID] = count
	}
	return counts, rows.Err()
}

// GetReadLaterCountsForAllFeeds returns a map of feed_id to read_later article count.
func (db *DB) GetReadLaterCountsForAllFeeds() (map[int64]int, error) {
	db.WaitForReady()
	rows, err := db.Query(`
		SELECT feed_id, COUNT(*)
		FROM articles
		WHERE is_read_later = 1 AND is_hidden = 0
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
			log.Println("Error scanning read_later count:", err)
			continue
		}
		counts[feedID] = count
	}
	return counts, rows.Err()
}

// GetImageModeCountsForAllFeeds returns a map of feed_id to image article count.
func (db *DB) GetImageModeCountsForAllFeeds() (map[int64]int, error) {
	db.WaitForReady()
	rows, err := db.Query(`
		SELECT feed_id, COUNT(*)
		FROM articles
		WHERE (image_url IS NOT NULL AND image_url != '') AND is_hidden = 0
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
			log.Println("Error scanning image mode count:", err)
			continue
		}
		counts[feedID] = count
	}
	return counts, rows.Err()
}

// GetImageUnreadCountsForAllFeeds returns a map of feed_id to unread image article count.
func (db *DB) GetImageUnreadCountsForAllFeeds() (map[int64]int, error) {
	db.WaitForReady()
	rows, err := db.Query(`
		SELECT feed_id, COUNT(*)
		FROM articles
		WHERE (image_url IS NOT NULL AND image_url != '') AND is_read = 0 AND is_hidden = 0
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
			log.Println("Error scanning image unread count:", err)
			continue
		}
		counts[feedID] = count
	}
	return counts, rows.Err()
}

// GetFavoriteUnreadCountsForAllFeeds returns a map of feed_id to favorite AND unread article count.
func (db *DB) GetFavoriteUnreadCountsForAllFeeds() (map[int64]int, error) {
	db.WaitForReady()
	rows, err := db.Query(`
		SELECT feed_id, COUNT(*)
		FROM articles
		WHERE is_favorite = 1 AND is_read = 0 AND is_hidden = 0
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
			log.Println("Error scanning favorite unread count:", err)
			continue
		}
		counts[feedID] = count
	}
	return counts, rows.Err()
}

// GetReadLaterUnreadCountsForAllFeeds returns a map of feed_id to read_later AND unread article count.
func (db *DB) GetReadLaterUnreadCountsForAllFeeds() (map[int64]int, error) {
	db.WaitForReady()
	rows, err := db.Query(`
		SELECT feed_id, COUNT(*)
		FROM articles
		WHERE is_read_later = 1 AND is_read = 0 AND is_hidden = 0
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
			log.Println("Error scanning read_later unread count:", err)
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

// MarkArticlesRelativeToPublishedTime marks articles as read based on their published time relative to a reference article.
// direction: "above" marks articles with published_at > reference_published_at (newer articles)
// direction: "below" marks articles with published_at < reference_published_at (older articles)
// feedID: optional, if provided only marks articles from that feed
// category: optional, if provided only marks articles from that category
// Returns the number of articles marked as read.
func (db *DB) MarkArticlesRelativeToPublishedTime(referencePublishedAt time.Time, direction string, feedID int64, category string) (int, error) {
	db.WaitForReady()

	var operator string

	switch direction {
	case "above":
		operator = ">"
	case "below":
		operator = "<"
	default:
		return 0, fmt.Errorf("invalid direction: %s", direction)
	}

	baseQuery := "UPDATE articles SET is_read = 1 WHERE is_read = 0 AND is_hidden = 0 AND published_at IS NOT NULL AND published_at " + operator + " ?"
	args := []interface{}{referencePublishedAt}

	if feedID > 0 {
		baseQuery += " AND feed_id = ?"
		args = append(args, feedID)
	}

	if category != "" {
		if category == "\x00" {
			// Special value for uncategorized
			baseQuery += " AND feed_id IN (SELECT id FROM feeds WHERE category IS NULL OR category = '')"
		} else {
			baseQuery += " AND feed_id IN (SELECT id FROM feeds WHERE category = ?)"
			args = append(args, category)
		}
	}

	result, err := db.Exec(baseQuery, args...)
	if err != nil {
		return 0, err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// GetImageGalleryArticles retrieves articles from image mode feeds with pagination.
// If feedID is provided, it gets articles only from that feed (assuming it's an image mode feed).
// If category is provided, it gets articles from all image mode feeds in that category.
// Otherwise, it gets articles from all image mode feeds.
func (db *DB) GetImageGalleryArticles(feedID int64, category string, showHidden bool, limit, offset int) ([]models.Article, error) {
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

// UpdateArticleSummary updates the cached summary for an article.
func (db *DB) UpdateArticleSummary(id int64, summary string) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET summary = ? WHERE id = ?", summary, id)
	return err
}

// GetArticleIDByUniqueID retrieves an article's ID by its unique identifier.
// This is the preferred method for looking up articles as it uses the title+feed_id+published_date based deduplication.
// Note: Uses date only (YYYY-MM-DD) rather than full timestamp for better deduplication.
func (db *DB) GetArticleIDByUniqueID(title string, feedID int64, publishedAt time.Time, hasValidPublishedTime bool) (int64, error) {
	db.WaitForReady()
	uniqueID := utils.GenerateArticleUniqueID(title, feedID, publishedAt, hasValidPublishedTime)
	var id int64
	err := db.QueryRow("SELECT id FROM articles WHERE unique_id = ?", uniqueID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
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

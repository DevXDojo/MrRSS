package database

import (
	"database/sql"
	"log"
	"time"
)

// This file adds FreshRSS sync tracking to article operations
// When FreshRSS sync is enabled, article changes trigger immediate sync

// SyncRequest represents a request to sync an article status to FreshRSS
type SyncRequest struct {
	ArticleID  int64
	ArticleURL string
	Action     SyncAction
}

// articleInfo represents basic article information for sync operations
type articleInfo struct {
	id     int64
	url    string
	feedID int64
}

// MarkArticleReadWithSync marks an article as read/unread and returns sync request if FreshRSS is enabled
func (db *DB) MarkArticleReadWithSync(id int64, read bool) (*SyncRequest, error) {
	// Get article URL and feed_id first
	var url string
	var feedID int64
	err := db.QueryRow("SELECT url, feed_id FROM articles WHERE id = ?", id).Scan(&url, &feedID)
	if err != nil {
		return nil, err
	}

	// Mark as read
	err = db.MarkArticleRead(id, read)
	if err != nil {
		return nil, err
	}

	// Track article read statistic if marking as read
	if read {
		_ = db.IncrementStat("article_read")
	}

	// Check if this article belongs to a FreshRSS feed
	var isFreshRSSFeed bool
	err = db.QueryRow("SELECT COALESCE(is_freshrss_source, 0) FROM feeds WHERE id = ?", feedID).Scan(&isFreshRSSFeed)
	if err != nil {
		return nil, err
	}

	// Return sync request only if FreshRSS is enabled and this is a FreshRSS feed
	enabled, _ := db.GetSetting("freshrss_enabled")
	if enabled == "true" && isFreshRSSFeed {
		action := SyncActionMarkRead
		if !read {
			action = SyncActionMarkUnread
		}
		log.Printf("[FreshRSS Sync] Article %d needs sync: %s", id, action)
		return &SyncRequest{
			ArticleID:  id,
			ArticleURL: url,
			Action:     action,
		}, nil
	}

	return nil, nil
}

// SetArticleFavoriteWithSync sets the favorite status and returns sync request if FreshRSS is enabled
func (db *DB) SetArticleFavoriteWithSync(id int64, favorite bool) (*SyncRequest, error) {
	// Get article URL and feed_id first
	var url string
	var feedID int64
	err := db.QueryRow("SELECT url, feed_id FROM articles WHERE id = ?", id).Scan(&url, &feedID)
	if err != nil {
		return nil, err
	}

	// Set favorite
	err = db.SetArticleFavorite(id, favorite)
	if err != nil {
		return nil, err
	}

	// Check if this article belongs to a FreshRSS feed
	var isFreshRSSFeed bool
	err = db.QueryRow("SELECT COALESCE(is_freshrss_source, 0) FROM feeds WHERE id = ?", feedID).Scan(&isFreshRSSFeed)
	if err != nil {
		return nil, err
	}

	// Return sync request only if FreshRSS is enabled and this is a FreshRSS feed
	enabled, _ := db.GetSetting("freshrss_enabled")
	if enabled == "true" && isFreshRSSFeed {
		action := SyncActionStar
		if !favorite {
			action = SyncActionUnstar
		}
		log.Printf("[FreshRSS Sync] Article %d needs sync: %s", id, action)
		return &SyncRequest{
			ArticleID:  id,
			ArticleURL: url,
			Action:     action,
		}, nil
	}

	return nil, nil
}

// ToggleFavoriteWithSync toggles favorite status and returns sync request if FreshRSS is enabled
func (db *DB) ToggleFavoriteWithSync(id int64) (*SyncRequest, error) {
	// Get current state, URL, and feed_id
	var isFav bool
	var url string
	var feedID int64
	err := db.QueryRow("SELECT is_favorite, url, feed_id FROM articles WHERE id = ?", id).Scan(&isFav, &url, &feedID)
	if err != nil {
		return nil, err
	}

	// Toggle favorite
	err = db.ToggleFavorite(id)
	if err != nil {
		return nil, err
	}

	// Track favorite action (only when adding to favorites, not removing)
	if !isFav {
		_ = db.IncrementStat("article_favorite")
	}

	// Check if this article belongs to a FreshRSS feed
	var isFreshRSSFeed bool
	err = db.QueryRow("SELECT COALESCE(is_freshrss_source, 0) FROM feeds WHERE id = ?", feedID).Scan(&isFreshRSSFeed)
	if err != nil {
		return nil, err
	}

	// Return sync request only if FreshRSS is enabled and this is a FreshRSS feed
	enabled, _ := db.GetSetting("freshrss_enabled")
	if enabled == "true" && isFreshRSSFeed {
		action := SyncActionStar
		if isFav {
			// Was favorited, now unfavorited
			action = SyncActionUnstar
		}
		log.Printf("[FreshRSS Sync] Article %d needs sync: %s", id, action)
		return &SyncRequest{
			ArticleID:  id,
			ArticleURL: url,
			Action:     action,
		}, nil
	}

	return nil, nil
}

// GetArticleByURL retrieves an article by its URL for sync purposes
func (db *DB) GetArticleByURL(url string) (*Article, error) {
	db.WaitForReady()

	query := `
		SELECT id, feed_id, title, url, is_read, is_favorite, published_at, freshrss_item_id
		FROM articles
		WHERE url = ?
		LIMIT 1
	`

	var article Article
	var publishedAt interface{}
	var freshRSSItemID sql.NullString
	err := db.QueryRow(query, url).Scan(
		&article.ID,
		&article.FeedID,
		&article.Title,
		&article.URL,
		&article.IsRead,
		&article.IsFavorite,
		&publishedAt,
		&freshRSSItemID,
	)

	if err != nil {
		return nil, err
	}

	article.FreshRSSItemID = freshRSSItemID.String
	return &article, nil
}

// Article represents a simplified article for sync operations
type Article struct {
	ID             int64
	FeedID         int64
	Title          string
	URL            string
	IsRead         bool
	IsFavorite     bool
	PublishedAt    interface{}
	FreshRSSItemID string
}

// MarkArticlesReadWithSync marks multiple articles as read and returns sync requests if FreshRSS is enabled
func (db *DB) MarkArticlesReadWithSync(ids []int64, read bool) ([]SyncRequest, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	// Get URLs and feed_ids for all articles
	type articleInfo struct {
		url    string
		feedID int64
	}
	articles := make(map[int64]articleInfo)
	for _, id := range ids {
		var url string
		var feedID int64
		err := db.QueryRow("SELECT url, feed_id FROM articles WHERE id = ?", id).Scan(&url, &feedID)
		if err == nil {
			articles[id] = articleInfo{url: url, feedID: feedID}
		}
	}

	// Mark all as read
	isRead := 0
	if read {
		isRead = 1
	}

	for _, id := range ids {
		_, err := db.Exec("UPDATE articles SET is_read = ? WHERE id = ?", isRead, id)
		if err != nil {
			return nil, err
		}
	}

	// Check if FreshRSS is enabled
	enabled, _ := db.GetSetting("freshrss_enabled")
	var syncRequests []SyncRequest
	if enabled == "true" {
		action := SyncActionMarkRead
		if !read {
			action = SyncActionMarkUnread
		}

		// Collect sync requests
		for id, info := range articles {
			// Check if this article belongs to a FreshRSS feed
			var isFreshRSSFeed bool
			err := db.QueryRow("SELECT COALESCE(is_freshrss_source, 0) FROM feeds WHERE id = ?", info.feedID).Scan(&isFreshRSSFeed)
			if err == nil && isFreshRSSFeed {
				syncRequests = append(syncRequests, SyncRequest{
					ArticleID:  id,
					ArticleURL: info.url,
					Action:     action,
				})
				log.Printf("[FreshRSS Sync] Article %d needs sync: %s", id, action)
			}
		}
	}

	return syncRequests, nil
}

// GetFreshRSSIDForArticle converts a local article ID to a FreshRSS-compatible ID
// FreshRSS uses the format: tag:google.com,2005:reader/item/{itemID}
func (db *DB) GetFreshRSSIDForArticle(articleID int64) (string, error) {
	db.WaitForReady()

	// Check if we have a stored FreshRSS ID
	var freshRSSID sql.NullString
	err := db.QueryRow("SELECT url FROM articles WHERE id = ?", articleID).Scan(&freshRSSID)
	if err != nil {
		return "", err
	}

	// For now, use the article URL as the ID
	// FreshRSS can accept either the full item ID or URL-based identification
	// In a full implementation, we would store the FreshRSS item ID when syncing articles
	return freshRSSID.String, nil
}

// ShouldSyncWithFreshRSS checks if FreshRSS sync is enabled and configured
func (db *DB) ShouldSyncWithFreshRSS() bool {
	enabled, _ := db.GetSetting("freshrss_enabled")
	if enabled != "true" {
		return false
	}

	serverURL, _ := db.GetSetting("freshrss_server_url")
	username, _ := db.GetSetting("freshrss_username")
	password, _ := db.GetEncryptedSetting("freshrss_api_password")

	return serverURL != "" && username != "" && password != ""
}

// GetFreshRSSConfig retrieves FreshRSS configuration
func (db *DB) GetFreshRSSConfig() (serverURL, username, password string, err error) {
	serverURL, _ = db.GetSetting("freshrss_server_url")
	username, _ = db.GetSetting("freshrss_username")
	password, err = db.GetEncryptedSetting("freshrss_api_password")
	return
}

// UpdateFreshRSSItemID updates the FreshRSS item ID for an article
func (db *DB) UpdateFreshRSSItemID(articleID int64, freshRSSItemID string) error {
	db.WaitForReady()

	query := `UPDATE articles SET freshrss_item_id = ? WHERE id = ?`
	_, err := db.Exec(query, freshRSSItemID, articleID)
	if err != nil {
		return err
	}

	log.Printf("[UpdateFreshRSSItemID] Updated article %d with FreshRSS Item ID: %s", articleID, freshRSSItemID)
	return nil
}

// MarkAllAsReadWithSync marks all articles as read and returns sync requests if FreshRSS is enabled
func (db *DB) MarkAllAsReadWithSync() ([]SyncRequest, error) {
	// Get all unread article IDs and URLs
	query := `SELECT id, url, feed_id FROM articles WHERE is_read = 0 AND is_hidden = 0`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []articleInfo
	for rows.Next() {
		var info articleInfo
		if err := rows.Scan(&info.id, &info.url, &info.feedID); err != nil {
			return nil, err
		}
		articles = append(articles, info)
	}

	// Mark all as read
	if err := db.MarkAllAsRead(); err != nil {
		return nil, err
	}

	// Collect sync requests for FreshRSS feeds
	return db.collectSyncRequests(articles), nil
}

// MarkAllAsReadForFeedWithSync marks all articles in a feed as read and returns sync requests if FreshRSS is enabled
func (db *DB) MarkAllAsReadForFeedWithSync(feedID int64) ([]SyncRequest, error) {
	// Get all unread article IDs and URLs for this feed
	query := `SELECT id, url, feed_id FROM articles WHERE feed_id = ? AND is_read = 0 AND is_hidden = 0`
	rows, err := db.Query(query, feedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []articleInfo
	for rows.Next() {
		var info articleInfo
		if err := rows.Scan(&info.id, &info.url, &info.feedID); err != nil {
			return nil, err
		}
		articles = append(articles, info)
	}

	// Mark all as read for this feed
	if err := db.MarkAllAsReadForFeed(feedID); err != nil {
		return nil, err
	}

	// Collect sync requests for FreshRSS feeds
	return db.collectSyncRequests(articles), nil
}

// MarkAllAsReadForCategoryWithSync marks all articles in a category as read and returns sync requests if FreshRSS is enabled
func (db *DB) MarkAllAsReadForCategoryWithSync(category string) ([]SyncRequest, error) {
	// Get all feed IDs in this category
	var feedIDs []int64

	// Handle empty category (uncategorized)
	var rows *sql.Rows
	var err error
	if category == "" {
		rows, err = db.Query("SELECT id FROM feeds WHERE (category IS NULL OR category = '')")
	} else {
		rows, err = db.Query("SELECT id FROM feeds WHERE category = ?", category)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var feedID int64
		if err := rows.Scan(&feedID); err != nil {
			return nil, err
		}
		feedIDs = append(feedIDs, feedID)
	}

	// Get all unread article IDs and URLs for feeds in this category
	var articles []articleInfo

	for _, feedID := range feedIDs {
		query := `SELECT id, url, feed_id FROM articles WHERE feed_id = ? AND is_read = 0 AND is_hidden = 0`
		rows, err := db.Query(query, feedID)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var info articleInfo
			if err := rows.Scan(&info.id, &info.url, &info.feedID); err != nil {
				rows.Close()
				return nil, err
			}
			articles = append(articles, info)
		}
		rows.Close()
	}

	// Mark all as read for this category
	if err := db.MarkAllAsReadForCategory(category); err != nil {
		return nil, err
	}

	// Collect sync requests for FreshRSS feeds
	return db.collectSyncRequests(articles), nil
}

// MarkArticlesRelativeToPublishedTimeWithSync marks articles relative to a reference time and returns sync requests
func (db *DB) MarkArticlesRelativeToPublishedTimeWithSync(referencePublishedAt time.Time, direction string, feedID int64, category string) (int, []SyncRequest, error) {
	// Build query based on parameters
	var operator string
	if direction == "above" {
		operator = ">"
	} else {
		operator = "<"
	}

	// Build WHERE clause
	var whereClause string
	var args []interface{}
	baseClause := "is_read = 0 AND is_hidden = 0 AND published_at " + operator + " ?"
	args = append(args, referencePublishedAt)

	if feedID > 0 {
		whereClause = baseClause + " AND feed_id = ?"
		args = append(args, feedID)
	} else if category != "" {
		// Get feed IDs in this category
		var feedIDs []int64
		var rows *sql.Rows
		var err error
		if category == "" {
			rows, err = db.Query("SELECT id FROM feeds WHERE (category IS NULL OR category = '')")
		} else {
			rows, err = db.Query("SELECT id FROM feeds WHERE category = ?", category)
		}
		if err != nil {
			return 0, nil, err
		}

		for rows.Next() {
			var fid int64
			if err := rows.Scan(&fid); err != nil {
				rows.Close()
				return 0, nil, err
			}
			feedIDs = append(feedIDs, fid)
		}
		rows.Close()

		// Build IN clause for feed IDs
		if len(feedIDs) > 0 {
			whereClause = baseClause + " AND feed_id IN ("
			for i := range feedIDs {
				if i > 0 {
					whereClause += ","
				}
				whereClause += "?"
				args = append(args, feedIDs[i])
			}
			whereClause += ")"
		} else {
			// No feeds in category, nothing to mark
			return 0, nil, nil
		}
	} else {
		whereClause = baseClause
	}

	// Get articles to mark
	query := `SELECT id, url, feed_id FROM articles WHERE ` + whereClause
	rows, err := db.Query(query, args...)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	var articles []articleInfo
	for rows.Next() {
		var info articleInfo
		if err := rows.Scan(&info.id, &info.url, &info.feedID); err != nil {
			return 0, nil, err
		}
		articles = append(articles, info)
	}

	// Mark articles as read
	count, err := db.MarkArticlesRelativeToPublishedTime(referencePublishedAt, direction, feedID, category)
	if err != nil {
		return 0, nil, err
	}

	// Collect sync requests for FreshRSS feeds
	syncRequests := db.collectSyncRequests(articles)
	return count, syncRequests, nil
}

// collectSyncRequests collects sync requests for articles that belong to FreshRSS feeds
func (db *DB) collectSyncRequests(articles []articleInfo) []SyncRequest {
	enabled, _ := db.GetSetting("freshrss_enabled")
	if enabled != "true" {
		return nil
	}

	var syncRequests []SyncRequest
	action := SyncActionMarkRead

	for _, article := range articles {
		// Check if this article belongs to a FreshRSS feed
		var isFreshRSSFeed bool
		err := db.QueryRow("SELECT COALESCE(is_freshrss_source, 0) FROM feeds WHERE id = ?", article.feedID).Scan(&isFreshRSSFeed)
		if err == nil && isFreshRSSFeed {
			syncRequests = append(syncRequests, SyncRequest{
				ArticleID:  article.id,
				ArticleURL: article.url,
				Action:     action,
			})
			log.Printf("[FreshRSS Sync] Article %d needs sync: %s", article.id, action)
		}
	}

	return syncRequests
}

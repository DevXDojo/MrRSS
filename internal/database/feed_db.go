package database

import (
	"database/sql"
	"time"

	"MrRSS/internal/models"
)

// AddFeed adds a new feed or updates an existing one.
// Returns the feed ID and any error encountered.
// IMPORTANT: We allow the same URL from different sources (FreshRSS vs local),
// so we check both url AND is_freshrss_source when looking for existing feeds.
func (db *DB) AddFeed(feed *models.Feed) (int64, error) {
	db.WaitForReady()

	// Check if feed already exists with same URL AND same source type
	var existingID int64
	var existingIsFreshRSS bool
	err := db.QueryRow("SELECT id, is_freshrss_source FROM feeds WHERE url = ?", feed.URL).Scan(&existingID, &existingIsFreshRSS)

	if err == sql.ErrNoRows {
		// Feed doesn't exist, insert new
		// Get next position in category if not specified
		position := feed.Position
		if position == 0 {
			position, err = db.GetNextPositionInCategory(feed.Category)
			if err != nil {
				return 0, err
			}
		}

		// 36 columns to insert (added is_freshrss_source and freshrss_stream_id)
		query := `INSERT INTO feeds (
			title, url, link, description, category, image_url, position,
			script_path, hide_from_timeline, proxy_url, proxy_enabled, refresh_interval,
			is_image_mode, type,
			xpath_item, xpath_item_title, xpath_item_content, xpath_item_uri,
			xpath_item_author, xpath_item_timestamp, xpath_item_time_format,
			xpath_item_thumbnail, xpath_item_categories, xpath_item_uid,
			article_view_mode, auto_expand_content,
			email_address, email_imap_server, email_imap_port,
			email_username, email_password, email_folder, email_last_uid,
			is_freshrss_source, freshrss_stream_id,
			last_updated
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		result, err := db.Exec(query,
			feed.Title, feed.URL, feed.Link, feed.Description, feed.Category, feed.ImageURL, position,
			feed.ScriptPath, feed.HideFromTimeline, feed.ProxyURL, feed.ProxyEnabled, feed.RefreshInterval,
			feed.IsImageMode, feed.Type,
			feed.XPathItem, feed.XPathItemTitle, feed.XPathItemContent, feed.XPathItemUri,
			feed.XPathItemAuthor, feed.XPathItemTimestamp, feed.XPathItemTimeFormat,
			feed.XPathItemThumbnail, feed.XPathItemCategories, feed.XPathItemUid,
			feed.ArticleViewMode, feed.AutoExpandContent,
			feed.EmailAddress, feed.EmailIMAPServer, feed.EmailIMAPPort,
			feed.EmailUsername, feed.EmailPassword, feed.EmailFolder, feed.EmailLastUID,
			feed.IsFreshRSSSource, feed.FreshRSSStreamID,
			time.Now())
		if err != nil {
			return 0, err
		}
		newID, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		return newID, nil
	} else if err != nil {
		return 0, err
	}

	// Feed with same URL exists, check if source type matches
	if existingIsFreshRSS != feed.IsFreshRSSSource {
		// Different source types - create a new feed instead of updating
		// This allows FreshRSS and local feeds with the same URL to coexist
		// Get next position in category if not specified
		position := feed.Position
		if position == 0 {
			position, err = db.GetNextPositionInCategory(feed.Category)
			if err != nil {
				return 0, err
			}
		}

		query := `INSERT INTO feeds (
			title, url, link, description, category, image_url, position,
			script_path, hide_from_timeline, proxy_url, proxy_enabled, refresh_interval,
			is_image_mode, type,
			xpath_item, xpath_item_title, xpath_item_content, xpath_item_uri,
			xpath_item_author, xpath_item_timestamp, xpath_item_time_format,
			xpath_item_thumbnail, xpath_item_categories, xpath_item_uid,
			article_view_mode, auto_expand_content,
			email_address, email_imap_server, email_imap_port,
			email_username, email_password, email_folder, email_last_uid,
			is_freshrss_source, freshrss_stream_id,
			last_updated
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		result, err := db.Exec(query,
			feed.Title, feed.URL, feed.Link, feed.Description, feed.Category, feed.ImageURL, position,
			feed.ScriptPath, feed.HideFromTimeline, feed.ProxyURL, feed.ProxyEnabled, feed.RefreshInterval,
			feed.IsImageMode, feed.Type,
			feed.XPathItem, feed.XPathItemTitle, feed.XPathItemContent, feed.XPathItemUri,
			feed.XPathItemAuthor, feed.XPathItemTimestamp, feed.XPathItemTimeFormat,
			feed.XPathItemThumbnail, feed.XPathItemCategories, feed.XPathItemUid,
			feed.ArticleViewMode, feed.AutoExpandContent,
			feed.EmailAddress, feed.EmailIMAPServer, feed.EmailIMAPPort,
			feed.EmailUsername, feed.EmailPassword, feed.EmailFolder, feed.EmailLastUID,
			feed.IsFreshRSSSource, feed.FreshRSSStreamID,
			time.Now())
		if err != nil {
			return 0, err
		}
		newID, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		return newID, nil
	}

	// Same URL and same source type - update existing feed
	// (note: we don't update is_freshrss_source or freshrss_stream_id for existing feeds)
	query := `UPDATE feeds SET title = ?, link = ?, description = ?, category = ?, image_url = ?, position = ?, script_path = ?, hide_from_timeline = ?, proxy_url = ?, proxy_enabled = ?, refresh_interval = ?, is_image_mode = ?, type = ?, xpath_item = ?, xpath_item_title = ?, xpath_item_content = ?, xpath_item_uri = ?, xpath_item_author = ?, xpath_item_timestamp = ?, xpath_item_time_format = ?, xpath_item_thumbnail = ?, xpath_item_categories = ?, xpath_item_uid = ?, article_view_mode = ?, auto_expand_content = ?, email_address = ?, email_imap_server = ?, email_imap_port = ?, email_username = ?, email_password = ?, email_folder = ?, email_last_uid = ?, last_updated = ? WHERE id = ?`
	_, err = db.Exec(query, feed.Title, feed.Link, feed.Description, feed.Category, feed.ImageURL, feed.Position, feed.ScriptPath, feed.HideFromTimeline, feed.ProxyURL, feed.ProxyEnabled, feed.RefreshInterval, feed.IsImageMode, feed.Type, feed.XPathItem, feed.XPathItemTitle, feed.XPathItemContent, feed.XPathItemUri, feed.XPathItemAuthor, feed.XPathItemTimestamp, feed.XPathItemTimeFormat, feed.XPathItemThumbnail, feed.XPathItemCategories, feed.XPathItemUid, feed.ArticleViewMode, feed.AutoExpandContent, feed.EmailAddress, feed.EmailIMAPServer, feed.EmailIMAPPort, feed.EmailUsername, feed.EmailPassword, feed.EmailFolder, feed.EmailLastUID, time.Now(), existingID)
	return existingID, err
}

// DeleteFeed deletes a feed and all its articles.
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

// GetFeeds returns all feeds ordered by category and position.
func (db *DB) GetFeeds() ([]models.Feed, error) {
	db.WaitForReady()
	rows, err := db.Query(`
		SELECT
			f.id, f.title, f.url, f.link, f.description, f.category, f.image_url,
			COALESCE(f.position, 0), f.last_updated, f.last_error,
			COALESCE(f.discovery_completed, 0), COALESCE(f.script_path, ''),
			COALESCE(f.hide_from_timeline, 0), COALESCE(f.proxy_url, ''),
			COALESCE(f.proxy_enabled, 0), COALESCE(f.refresh_interval, 0),
			COALESCE(f.is_image_mode, 0), COALESCE(f.type, ''),
			COALESCE(f.xpath_item, ''), COALESCE(f.xpath_item_title, ''),
			COALESCE(f.xpath_item_content, ''), COALESCE(f.xpath_item_uri, ''),
			COALESCE(f.xpath_item_author, ''), COALESCE(f.xpath_item_timestamp, ''),
			COALESCE(f.xpath_item_time_format, ''), COALESCE(f.xpath_item_thumbnail, ''),
			COALESCE(f.xpath_item_categories, ''), COALESCE(f.xpath_item_uid, ''),
			COALESCE(f.article_view_mode, 'global'),
			COALESCE(f.auto_expand_content, 'global'),
			COALESCE(f.email_address, ''), COALESCE(f.email_imap_server, ''),
			COALESCE(f.email_imap_port, 993), COALESCE(f.email_username, ''),
			COALESCE(f.email_password, ''), COALESCE(f.email_folder, 'INBOX'),
			COALESCE(f.email_last_uid, 0), COALESCE(f.is_freshrss_source, 0),
			COALESCE(f.freshrss_stream_id, ''),
			(SELECT MAX(a.published_at) FROM articles a WHERE a.feed_id = f.id) as latest_article_time,
			CAST(COALESCE((
				SELECT
					CASE
						-- If we have articles spanning more than 365 days, use last 365 days
						WHEN EXISTS (SELECT 1 FROM articles a2 WHERE a2.feed_id = f.id AND a2.published_at IS NOT NULL AND a2.published_at != '' AND CAST(julianday('now') - julianday(substr(a2.published_at, 1, 19)) AS REAL) > 365.0)
						THEN (SELECT CAST(COUNT(*) * 30.0 / 365.0 AS REAL) FROM articles a3 WHERE a3.feed_id = f.id AND a3.published_at IS NOT NULL AND a3.published_at != '' AND julianday(substr(a3.published_at, 1, 19)) >= julianday(datetime('now', '-365 days')))
						-- Otherwise, calculate based on the actual time span of all articles
						ELSE (
							SELECT CAST(COUNT(*) * 30.0 AS REAL) /
								CASE
									WHEN CAST(julianday(substr(MAX(a4.published_at), 1, 19)) - julianday(substr(MIN(a4.published_at), 1, 19)) AS REAL) < 1.0 THEN 1.0
									ELSE CAST(julianday(substr(MAX(a4.published_at), 1, 19)) - julianday(substr(MIN(a4.published_at), 1, 19)) AS REAL)
								END
							FROM articles a4
							WHERE a4.feed_id = f.id AND a4.published_at IS NOT NULL AND a4.published_at != ''
						)
					END
			), 0) AS REAL) as articles_per_month
		FROM feeds f
		ORDER BY f.category ASC, f.position ASC, f.id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []models.Feed
	for rows.Next() {
		var f models.Feed
		var link, category, imageURL, lastError, scriptPath, proxyURL, feedType, xpathItem, xpathItemTitle, xpathItemContent, xpathItemUri, xpathItemAuthor, xpathItemTimestamp, xpathItemTimeFormat, xpathItemThumbnail, xpathItemCategories, xpathItemUid, articleViewMode, autoExpandContent, emailAddress, emailIMAPServer, emailUsername, emailPassword, emailFolder, freshRSSStreamID, latestArticleTimeStr sql.NullString
		var lastUpdated sql.NullTime
		if err := rows.Scan(
			&f.ID, &f.Title, &f.URL, &link, &f.Description, &category, &imageURL,
			&f.Position, &lastUpdated, &lastError, &f.DiscoveryCompleted, &scriptPath,
			&f.HideFromTimeline, &proxyURL, &f.ProxyEnabled, &f.RefreshInterval,
			&f.IsImageMode, &feedType, &xpathItem, &xpathItemTitle, &xpathItemContent,
			&xpathItemUri, &xpathItemAuthor, &xpathItemTimestamp, &xpathItemTimeFormat,
			&xpathItemThumbnail, &xpathItemCategories, &xpathItemUid, &articleViewMode,
			&autoExpandContent, &emailAddress, &emailIMAPServer, &f.EmailIMAPPort,
			&emailUsername, &emailPassword, &emailFolder, &f.EmailLastUID,
			&f.IsFreshRSSSource, &freshRSSStreamID, &latestArticleTimeStr, &f.ArticlesPerMonth,
		); err != nil {
			return nil, err
		}

		// Round articles_per_month to integer for display
		f.ArticlesPerMonth = float64(int(f.ArticlesPerMonth + 0.5))

		f.Link = link.String
		f.Category = category.String
		f.ImageURL = imageURL.String
		if lastUpdated.Valid {
			f.LastUpdated = lastUpdated.Time
		} else {
			f.LastUpdated = time.Time{}
		}
		f.LastError = lastError.String
		f.ScriptPath = scriptPath.String
		f.ProxyURL = proxyURL.String
		f.Type = feedType.String
		f.XPathItem = xpathItem.String
		f.XPathItemTitle = xpathItemTitle.String
		f.XPathItemContent = xpathItemContent.String
		f.XPathItemUri = xpathItemUri.String
		f.XPathItemAuthor = xpathItemAuthor.String
		f.XPathItemTimestamp = xpathItemTimestamp.String
		f.XPathItemTimeFormat = xpathItemTimeFormat.String
		f.XPathItemThumbnail = xpathItemThumbnail.String
		f.XPathItemCategories = xpathItemCategories.String
		f.XPathItemUid = xpathItemUid.String
		f.ArticleViewMode = articleViewMode.String
		if f.ArticleViewMode == "" {
			f.ArticleViewMode = "global"
		}
		f.AutoExpandContent = autoExpandContent.String
		if f.AutoExpandContent == "" {
			f.AutoExpandContent = "global"
		}
		f.EmailAddress = emailAddress.String
		f.EmailIMAPServer = emailIMAPServer.String
		f.EmailUsername = emailUsername.String
		f.EmailPassword = emailPassword.String
		f.EmailFolder = emailFolder.String
		if f.EmailFolder == "" {
			f.EmailFolder = "INBOX"
		}
		if f.EmailIMAPPort == 0 {
			f.EmailIMAPPort = 993
		}
		f.FreshRSSStreamID = freshRSSStreamID.String

		// Set latest article time from string
		// Format from database: "2025-11-15 18:39:02 +0000 UTC" (Go's time.String() format)
		if latestArticleTimeStr.Valid && latestArticleTimeStr.String != "" {
			timeStr := latestArticleTimeStr.String
			var parsedTime time.Time
			var err error

			// Try the format stored in database (Go's time.String() format with UTC zone)
			// Format: "2006-01-02 15:04:05 +0000 UTC"
			if parsedTime, err = time.Parse("2006-01-02 15:04:05 -0700 MST", timeStr); err == nil {
				f.LatestArticleTime = &parsedTime
			} else if parsedTime, err = time.Parse(time.RFC3339, timeStr); err == nil {
				// RFC3339 format
				f.LatestArticleTime = &parsedTime
			} else if parsedTime, err = time.Parse("2006-01-02T15:04:05Z", timeStr); err == nil {
				// RFC3339 variant with explicit Z
				f.LatestArticleTime = &parsedTime
			} else if parsedTime, err = time.Parse("2006-01-02T15:04:05", timeStr); err == nil {
				// Without timezone
				f.LatestArticleTime = &parsedTime
			} else if parsedTime, err = time.Parse("2006-01-02 15:04:05", timeStr); err == nil {
				// SQLite default format
				f.LatestArticleTime = &parsedTime
			}
		}

		// Determine last update status based on last_error
		if f.LastError != "" {
			f.LastUpdateStatus = "failed"
		} else {
			f.LastUpdateStatus = "success"
		}

		feeds = append(feeds, f)
	}
	return feeds, nil
}

// GetFeedByID retrieves a specific feed by its ID.
func (db *DB) GetFeedByID(id int64) (*models.Feed, error) {
	db.WaitForReady()
	row := db.QueryRow("SELECT id, title, url, link, description, category, image_url, COALESCE(position, 0), last_updated, last_error, COALESCE(discovery_completed, 0), COALESCE(script_path, ''), COALESCE(hide_from_timeline, 0), COALESCE(proxy_url, ''), COALESCE(proxy_enabled, 0), COALESCE(refresh_interval, 0), COALESCE(is_image_mode, 0), COALESCE(type, ''), COALESCE(xpath_item, ''), COALESCE(xpath_item_title, ''), COALESCE(xpath_item_content, ''), COALESCE(xpath_item_uri, ''), COALESCE(xpath_item_author, ''), COALESCE(xpath_item_timestamp, ''), COALESCE(xpath_item_time_format, ''), COALESCE(xpath_item_thumbnail, ''), COALESCE(xpath_item_categories, ''), COALESCE(xpath_item_uid, ''), COALESCE(article_view_mode, 'global'), COALESCE(auto_expand_content, 'global'), COALESCE(email_address, ''), COALESCE(email_imap_server, ''), COALESCE(email_imap_port, 993), COALESCE(email_username, ''), COALESCE(email_password, ''), COALESCE(email_folder, 'INBOX'), COALESCE(email_last_uid, 0), COALESCE(is_freshrss_source, 0), COALESCE(freshrss_stream_id, '') FROM feeds WHERE id = ?", id)

	var f models.Feed
	var link, category, imageURL, lastError, scriptPath, proxyURL, feedType, xpathItem, xpathItemTitle, xpathItemContent, xpathItemUri, xpathItemAuthor, xpathItemTimestamp, xpathItemTimeFormat, xpathItemThumbnail, xpathItemCategories, xpathItemUid, articleViewMode, autoExpandContent, emailAddress, emailIMAPServer, emailUsername, emailPassword, emailFolder, freshRSSStreamID sql.NullString
	var lastUpdated sql.NullTime
	if err := row.Scan(&f.ID, &f.Title, &f.URL, &link, &f.Description, &category, &imageURL, &f.Position, &lastUpdated, &lastError, &f.DiscoveryCompleted, &scriptPath, &f.HideFromTimeline, &proxyURL, &f.ProxyEnabled, &f.RefreshInterval, &f.IsImageMode, &feedType, &xpathItem, &xpathItemTitle, &xpathItemContent, &xpathItemUri, &xpathItemAuthor, &xpathItemTimestamp, &xpathItemTimeFormat, &xpathItemThumbnail, &xpathItemCategories, &xpathItemUid, &articleViewMode, &autoExpandContent, &emailAddress, &emailIMAPServer, &f.EmailIMAPPort, &emailUsername, &emailPassword, &emailFolder, &f.EmailLastUID, &f.IsFreshRSSSource, &freshRSSStreamID); err != nil {
		return nil, err
	}
	f.Link = link.String
	f.Category = category.String
	f.ImageURL = imageURL.String
	if lastUpdated.Valid {
		f.LastUpdated = lastUpdated.Time
	} else {
		f.LastUpdated = time.Time{}
	}
	f.LastError = lastError.String
	f.ScriptPath = scriptPath.String
	f.ProxyURL = proxyURL.String
	f.Type = feedType.String
	f.XPathItem = xpathItem.String
	f.XPathItemTitle = xpathItemTitle.String
	f.XPathItemContent = xpathItemContent.String
	f.XPathItemUri = xpathItemUri.String
	f.XPathItemAuthor = xpathItemAuthor.String
	f.XPathItemTimestamp = xpathItemTimestamp.String
	f.XPathItemTimeFormat = xpathItemTimeFormat.String
	f.XPathItemThumbnail = xpathItemThumbnail.String
	f.XPathItemCategories = xpathItemCategories.String
	f.XPathItemUid = xpathItemUid.String
	f.ArticleViewMode = articleViewMode.String
	if f.ArticleViewMode == "" {
		f.ArticleViewMode = "global"
	}
	f.AutoExpandContent = autoExpandContent.String
	if f.AutoExpandContent == "" {
		f.AutoExpandContent = "global"
	}
	f.EmailAddress = emailAddress.String
	f.EmailIMAPServer = emailIMAPServer.String
	f.EmailUsername = emailUsername.String
	f.EmailPassword = emailPassword.String
	f.EmailFolder = emailFolder.String
	if f.EmailFolder == "" {
		f.EmailFolder = "INBOX"
	}
	if f.EmailIMAPPort == 0 {
		f.EmailIMAPPort = 993
	}
	f.FreshRSSStreamID = freshRSSStreamID.String

	return &f, nil
}

// GetAllFeedURLs returns a set of all subscribed RSS feed URLs for deduplication.
func (db *DB) GetAllFeedURLs() (map[string]bool, error) {
	db.WaitForReady()
	rows, err := db.Query("SELECT url FROM feeds")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := make(map[string]bool)
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		urls[url] = true
	}
	return urls, rows.Err()
}

// UpdateFeed updates feed title, URL, category, script_path, hide_from_timeline, proxy settings, refresh_interval, is_image_mode, XPath fields, article_view_mode, auto_expand_content, and email settings.
func (db *DB) UpdateFeed(id int64, title, url, category, scriptPath string, hideFromTimeline bool, proxyURL string, proxyEnabled bool, refreshInterval int, isImageMode bool, feedType string, xpathItem, xpathItemTitle, xpathItemContent, xpathItemUri, xpathItemAuthor, xpathItemTimestamp, xpathItemTimeFormat, xpathItemThumbnail, xpathItemCategories, xpathItemUid, articleViewMode, autoExpandContent, emailAddress, emailIMAPServer, emailUsername, emailPassword, emailFolder string, emailIMAPPort int) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE feeds SET title = ?, url = ?, category = ?, script_path = ?, hide_from_timeline = ?, proxy_url = ?, proxy_enabled = ?, refresh_interval = ?, is_image_mode = ?, type = ?, xpath_item = ?, xpath_item_title = ?, xpath_item_content = ?, xpath_item_uri = ?, xpath_item_author = ?, xpath_item_timestamp = ?, xpath_item_time_format = ?, xpath_item_thumbnail = ?, xpath_item_categories = ?, xpath_item_uid = ?, article_view_mode = ?, auto_expand_content = ?, email_address = ?, email_imap_server = ?, email_imap_port = ?, email_username = ?, email_password = ?, email_folder = ? WHERE id = ?", title, url, category, scriptPath, hideFromTimeline, proxyURL, proxyEnabled, refreshInterval, isImageMode, feedType, xpathItem, xpathItemTitle, xpathItemContent, xpathItemUri, xpathItemAuthor, xpathItemTimestamp, xpathItemTimeFormat, xpathItemThumbnail, xpathItemCategories, xpathItemUid, articleViewMode, autoExpandContent, emailAddress, emailIMAPServer, emailIMAPPort, emailUsername, emailPassword, emailFolder, id)
	return err
}

// UpdateFeedWithPosition updates a feed including its position field.
func (db *DB) UpdateFeedWithPosition(id int64, title, url, category, scriptPath string, position int, hideFromTimeline bool, proxyURL string, proxyEnabled bool, refreshInterval int, isImageMode bool, feedType string, xpathItem, xpathItemTitle, xpathItemContent, xpathItemUri, xpathItemAuthor, xpathItemTimestamp, xpathItemTimeFormat, xpathItemThumbnail, xpathItemCategories, xpathItemUid, articleViewMode, autoExpandContent, emailAddress, emailIMAPServer, emailUsername, emailPassword, emailFolder string, emailIMAPPort int) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE feeds SET title = ?, url = ?, category = ?, script_path = ?, position = ?, hide_from_timeline = ?, proxy_url = ?, proxy_enabled = ?, refresh_interval = ?, is_image_mode = ?, type = ?, xpath_item = ?, xpath_item_title = ?, xpath_item_content = ?, xpath_item_uri = ?, xpath_item_author = ?, xpath_item_timestamp = ?, xpath_item_time_format = ?, xpath_item_thumbnail = ?, xpath_item_categories = ?, xpath_item_uid = ?, article_view_mode = ?, auto_expand_content = ?, email_address = ?, email_imap_server = ?, email_imap_port = ?, email_username = ?, email_password = ?, email_folder = ? WHERE id = ?", title, url, category, scriptPath, position, hideFromTimeline, proxyURL, proxyEnabled, refreshInterval, isImageMode, feedType, xpathItem, xpathItemTitle, xpathItemContent, xpathItemUri, xpathItemAuthor, xpathItemTimestamp, xpathItemTimeFormat, xpathItemThumbnail, xpathItemCategories, xpathItemUid, articleViewMode, autoExpandContent, emailAddress, emailIMAPServer, emailIMAPPort, emailUsername, emailPassword, emailFolder, id)
	return err
}

// UpdateFeedCategory updates a feed's category.
func (db *DB) UpdateFeedCategory(id int64, category string) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE feeds SET category = ? WHERE id = ?", category, id)
	return err
}

// UpdateFeedImage updates a feed's image URL.
func (db *DB) UpdateFeedImage(id int64, imageURL string) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE feeds SET image_url = ? WHERE id = ?", imageURL, id)
	return err
}

// UpdateFeedLink updates a feed's homepage link.
func (db *DB) UpdateFeedLink(id int64, link string) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE feeds SET link = ? WHERE id = ?", link, id)
	return err
}

// UpdateFeedError updates a feed's error message.
func (db *DB) UpdateFeedError(id int64, errorMsg string) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE feeds SET last_error = ? WHERE id = ?", errorMsg, id)
	return err
}

// ClearAllFeedErrors clears error messages for all feeds.
func (db *DB) ClearAllFeedErrors() error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE feeds SET last_error = ''")
	return err
}

// UpdateFeedLastUpdated updates a feed's last_updated timestamp.
func (db *DB) UpdateFeedLastUpdated(id int64) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE feeds SET last_updated = datetime('now') WHERE id = ?", id)
	return err
}

// UpdateFeedEmailLastUID updates a newsletter feed's last processed email UID.
func (db *DB) UpdateFeedEmailLastUID(id int64, lastUID int) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE feeds SET email_last_uid = ? WHERE id = ?", lastUID, id)
	return err
}

// MarkFeedDiscovered marks a feed as having completed discovery.
func (db *DB) MarkFeedDiscovered(id int64) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE feeds SET discovery_completed = 1 WHERE id = ?", id)
	return err
}

// UpdateFeedPosition updates a feed's category and position.
func (db *DB) UpdateFeedPosition(id int64, category string, position int) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE feeds SET category = ?, position = ? WHERE id = ?", category, position, id)
	return err
}

// ReorderFeed reorders feeds within a category after moving a feed.
// The newPosition parameter is the visual index (0-based) where the feed should appear.
// This adjusts the positions of other feeds to maintain consistent ordering.
func (db *DB) ReorderFeed(feedID int64, newCategory string, newIndex int) error {
	db.WaitForReady()

	// Get the feed being moved
	var oldCategory string
	var oldPosition int
	err := db.QueryRow("SELECT COALESCE(category, ''), COALESCE(position, 0) FROM feeds WHERE id = ?", feedID).Scan(&oldCategory, &oldPosition)
	if err != nil {
		return err
	}

	// Get all feeds in the target category, ordered by position
	rows, err := db.Query("SELECT id, COALESCE(position, 0) FROM feeds WHERE category = ? ORDER BY position ASC, id ASC", newCategory)
	if err != nil {
		return err
	}
	defer rows.Close()

	type feedPosition struct {
		id       int64
		position int
	}
	var feeds []feedPosition
	for rows.Next() {
		var f feedPosition
		if err := rows.Scan(&f.id, &f.position); err != nil {
			return err
		}
		feeds = append(feeds, f)
	}

	// Find the old index of the feed being moved
	oldIndex := -1
	for i, f := range feeds {
		if f.id == feedID {
			oldIndex = i
			break
		}
	}

	// If moving within the same category
	if oldCategory == newCategory {
		// Adjust the newIndex if the feed is being moved within the same category
		// and the new index accounts for the feed being removed (which the frontend does)
		// No additional adjustment needed here

		// Remove the feed from its old position and insert at the new position
		var updatedFeeds []feedPosition
		for i, f := range feeds {
			if i != oldIndex {
				updatedFeeds = append(updatedFeeds, f)
			}
		}

		// Insert at the new position
		if newIndex > len(updatedFeeds) {
			newIndex = len(updatedFeeds)
		}

		var finalFeeds []feedPosition
		finalFeeds = append(finalFeeds, updatedFeeds[:newIndex]...)
		finalFeeds = append(finalFeeds, feedPosition{id: feedID})
		finalFeeds = append(finalFeeds, updatedFeeds[newIndex:]...)

		// Update all positions in the category
		for i, f := range finalFeeds {
			_, err = db.Exec("UPDATE feeds SET position = ? WHERE id = ?", i, f.id)
			if err != nil {
				return err
			}
		}
	} else {
		// Moving to different category
		// 1. Shift feeds in old category after old position up by 1
		_, err = db.Exec(`
			UPDATE feeds SET position = position - 1
			WHERE category = ? AND position > ?
		`, oldCategory, oldPosition)
		if err != nil {
			return err
		}

		// 2. Shift feeds in new category at and after new index down by 1
		// First, get the feeds in the new category again (without the moved feed)
		var newCategoryFeeds []feedPosition
		for _, f := range feeds {
			if f.id != feedID {
				newCategoryFeeds = append(newCategoryFeeds, f)
			}
		}

		// Insert the moved feed at the new index
		if newIndex > len(newCategoryFeeds) {
			newIndex = len(newCategoryFeeds)
		}

		var finalFeeds []feedPosition
		finalFeeds = append(finalFeeds, newCategoryFeeds[:newIndex]...)
		finalFeeds = append(finalFeeds, feedPosition{id: feedID})
		finalFeeds = append(finalFeeds, newCategoryFeeds[newIndex:]...)

		// Update all feeds in the new category
		for i, f := range finalFeeds {
			_, err = db.Exec("UPDATE feeds SET position = ?, category = ? WHERE id = ?", i, newCategory, f.id)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// GetNextPositionInCategory returns the next available position in a category.
func (db *DB) GetNextPositionInCategory(category string) (int, error) {
	db.WaitForReady()
	var maxPos int
	err := db.QueryRow("SELECT COALESCE(MAX(position), -1) FROM feeds WHERE category = ?", category).Scan(&maxPos)
	if err != nil {
		return 0, err
	}
	return maxPos + 1, nil
}

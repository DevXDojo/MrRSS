package database

import (
	"database/sql"
	"fmt"

	"MrRSS/internal/models"
)

// GetTags retrieves all tags ordered by position.
func (db *DB) GetTags() ([]models.Tag, error) {
	db.WaitForReady()

	query := `SELECT id, name, color, position FROM tags ORDER BY position ASC, id ASC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.Position)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// GetTagByID retrieves a single tag by ID.
func (db *DB) GetTagByID(id int64) (*models.Tag, error) {
	db.WaitForReady()

	query := `SELECT id, name, color, position FROM tags WHERE id = ?`
	var tag models.Tag
	err := db.QueryRow(query, id).Scan(&tag.ID, &tag.Name, &tag.Color, &tag.Position)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &tag, nil
}

// AddTag creates a new tag and returns its ID.
func (db *DB) AddTag(tag *models.Tag) (int64, error) {
	db.WaitForReady()

	// If position is 0, set it to the end
	if tag.Position == 0 {
		var maxPosition int
		err := db.QueryRow("SELECT COALESCE(MAX(position), 0) FROM tags").Scan(&maxPosition)
		if err != nil {
			return 0, err
		}
		tag.Position = maxPosition + 1
	}

	query := `INSERT INTO tags (name, color, position) VALUES (?, ?, ?)`
	result, err := db.Exec(query, tag.Name, tag.Color, tag.Position)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateTag updates an existing tag.
func (db *DB) UpdateTag(id int64, name, color string, position int) error {
	db.WaitForReady()

	query := `UPDATE tags SET name = ?, color = ?, position = ? WHERE id = ?`
	_, err := db.Exec(query, name, color, position, id)
	return err
}

// DeleteTag deletes a tag by ID.
// Note: ON DELETE CASCADE will automatically remove feed_tags associations.
func (db *DB) DeleteTag(id int64) error {
	db.WaitForReady()

	query := `DELETE FROM tags WHERE id = ?`
	_, err := db.Exec(query, id)
	return err
}

// ReorderTag changes the position of a tag.
func (db *DB) ReorderTag(id int64, newPosition int) error {
	db.WaitForReady()

	// Get the tag being moved
	tag, err := db.GetTagByID(id)
	if err != nil {
		return err
	}
	if tag == nil {
		return fmt.Errorf("tag not found")
	}

	oldPosition := tag.Position

	// If position hasn't changed, do nothing
	if oldPosition == newPosition {
		return nil
	}

	// Begin transaction for atomic reordering
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if newPosition < oldPosition {
		// Moving up: shift items between new and old position down
		_, err = tx.Exec(`
			UPDATE tags SET position = position + 1
			WHERE position >= ? AND position < ?
		`, newPosition, oldPosition)
	} else {
		// Moving down: shift items between old and new position up
		_, err = tx.Exec(`
			UPDATE tags SET position = position - 1
			WHERE position > ? AND position <= ?
		`, oldPosition, newPosition)
	}

	if err != nil {
		return err
	}

	// Update the tag's position
	_, err = tx.Exec(`UPDATE tags SET position = ? WHERE id = ?`, newPosition, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetFeedTags retrieves all tags for a specific feed.
func (db *DB) GetFeedTags(feedID int64) ([]models.Tag, error) {
	db.WaitForReady()

	query := `
		SELECT t.id, t.name, t.color, t.position
		FROM tags t
		INNER JOIN feed_tags ft ON t.id = ft.tag_id
		WHERE ft.feed_id = ?
		ORDER BY t.position ASC, t.id ASC
	`
	rows, err := db.Query(query, feedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.Position)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// SetFeedTags replaces all tags for a feed with the provided tag IDs.
// This is done in a transaction to ensure atomicity.
func (db *DB) SetFeedTags(feedID int64, tagIDs []int64) error {
	db.WaitForReady()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing feed-tag associations
	_, err = tx.Exec("DELETE FROM feed_tags WHERE feed_id = ?", feedID)
	if err != nil {
		return err
	}

	// Insert new associations
	if len(tagIDs) > 0 {
		stmt, err := tx.Prepare("INSERT INTO feed_tags (feed_id, tag_id) VALUES (?, ?)")
		if err != nil {
			return err
		}
		defer stmt.Close()

		for _, tagID := range tagIDs {
			_, err = stmt.Exec(feedID, tagID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// GetFeedsByTag retrieves all feeds that have a specific tag.
func (db *DB) GetFeedsByTag(tagID int64) ([]models.Feed, error) {
	db.WaitForReady()

	query := `
		SELECT f.id, f.title, f.url, f.link, f.description, f.category, f.image_url, f.position,
			f.last_updated, f.last_error, f.discovery_completed, f.script_path, f.hide_from_timeline,
			f.proxy_url, f.proxy_enabled, f.refresh_interval, f.is_image_mode, f.type,
			f.xpath_item, f.xpath_item_title, f.xpath_item_content, f.xpath_item_uri,
			f.xpath_item_author, f.xpath_item_timestamp, f.xpath_item_time_format,
			f.xpath_item_thumbnail, f.xpath_item_categories, f.xpath_item_uid,
			f.article_view_mode, f.auto_expand_content,
			f.email_address, f.email_imap_server, f.email_imap_port,
			f.email_username, f.email_password, f.email_folder, f.email_last_uid,
			f.is_freshrss_source, f.freshrss_stream_id
		FROM feeds f
		INNER JOIN feed_tags ft ON f.id = ft.feed_id
		WHERE ft.tag_id = ?
		ORDER BY f.category ASC, f.position ASC, f.id ASC
	`
	rows, err := db.Query(query, tagID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []models.Feed
	for rows.Next() {
		var feed models.Feed
		var lastUpdatedAt, lastError sql.NullString
		err := rows.Scan(
			&feed.ID, &feed.Title, &feed.URL, &feed.Link, &feed.Description, &feed.Category,
			&feed.ImageURL, &feed.Position, &lastUpdatedAt, &lastError, &feed.DiscoveryCompleted,
			&feed.ScriptPath, &feed.HideFromTimeline, &feed.ProxyURL, &feed.ProxyEnabled,
			&feed.RefreshInterval, &feed.IsImageMode, &feed.Type, &feed.XPathItem,
			&feed.XPathItemTitle, &feed.XPathItemContent, &feed.XPathItemUri, &feed.XPathItemAuthor,
			&feed.XPathItemTimestamp, &feed.XPathItemTimeFormat, &feed.XPathItemThumbnail,
			&feed.XPathItemCategories, &feed.XPathItemUid, &feed.ArticleViewMode,
			&feed.AutoExpandContent, &feed.EmailAddress, &feed.EmailIMAPServer,
			&feed.EmailIMAPPort, &feed.EmailUsername, &feed.EmailPassword, &feed.EmailFolder,
			&feed.EmailLastUID, &feed.IsFreshRSSSource, &feed.FreshRSSStreamID,
		)
		if err != nil {
			return nil, err
		}

		if lastError.Valid {
			feed.LastError = lastError.String
		}

		feeds = append(feeds, feed)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return feeds, nil
}

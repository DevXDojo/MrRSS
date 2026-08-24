package database

import (
	"fmt"
	"time"
)

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

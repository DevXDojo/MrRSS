package database

import "log"

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

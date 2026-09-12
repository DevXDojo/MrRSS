package database

// CleanupFreshRSSData removes all FreshRSS-related feeds, articles, and sync queue items
// This should be called when FreshRSS is disabled or its settings are changed
func (db *DB) CleanupFreshRSSData() error {
	return db.CleanupReaderData("freshrss")
}

// getFreshRSSFeedIDs returns IDs of all FreshRSS feeds
func (db *DB) getFreshRSSFeedIDs() ([]int64, error) {
	rows, err := db.Query("SELECT id FROM feeds WHERE is_freshrss_source = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feedIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		feedIDs = append(feedIDs, id)
	}

	return feedIDs, rows.Err()
}

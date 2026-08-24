package database

import (
	"database/sql"
	"time"
)

// StatRecord represents a single statistics record
type StatRecord struct {
	ID        int64
	EventDate string // Format: YYYY-MM-DD
	EventType string // "feed_refresh", "article_read", "ai_chat", "ai_summary"
	Count     int
	CreatedAt time.Time
}

// InitStatisticsTable creates the statistics table if it doesn't exist
// IMPORTANT: This table is NEVER cleaned up by database cleanup functions
func InitStatisticsTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS statistics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_date TEXT NOT NULL,
		event_type TEXT NOT NULL,
		count INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(event_date, event_type)
	);

	-- Create indexes for faster queries
	CREATE INDEX IF NOT EXISTS idx_statistics_event_date ON statistics(event_date);
	CREATE INDEX IF NOT EXISTS idx_statistics_event_type ON statistics(event_type);
	CREATE INDEX IF NOT EXISTS idx_statistics_date_type ON statistics(event_date, event_type);
	`
	_, err := db.Exec(query)
	return err
}

// IncrementStat increments a statistic counter for a specific date and event type
func (db *DB) IncrementStat(eventType string) error {
	db.WaitForReady()

	today := time.Now().Format("2006-01-02")

	query := `
	INSERT INTO statistics (event_date, event_type, count)
	VALUES (?, ?, 1)
	ON CONFLICT(event_date, event_type) DO UPDATE SET
		count = count + 1,
		created_at = CURRENT_TIMESTAMP
	`
	_, err := db.Exec(query, today, eventType)
	return err
}

// GetStatsByDateRange retrieves statistics for a specific date range
func (db *DB) GetStatsByDateRange(startDate, endDate string) ([]StatRecord, error) {
	db.WaitForReady()

	query := `
	SELECT event_date, event_type, count
	FROM statistics
	WHERE event_date >= ? AND event_date <= ?
	ORDER BY event_date DESC, event_type ASC
	`
	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []StatRecord
	for rows.Next() {
		var stat StatRecord
		err := rows.Scan(&stat.EventDate, &stat.EventType, &stat.Count)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

// GetStatsAggregated retrieves aggregated statistics grouped by event type
func (db *DB) GetStatsAggregated(startDate, endDate string) (map[string]int, error) {
	db.WaitForReady()

	query := `
	SELECT event_type, SUM(count) as total
	FROM statistics
	WHERE event_date >= ? AND event_date <= ?
	GROUP BY event_type
	`
	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var eventType string
		var total int
		err := rows.Scan(&eventType, &total)
		if err != nil {
			return nil, err
		}
		stats[eventType] = total
	}

	return stats, nil
}

// GetStatsByDate retrieves statistics grouped by date for a specific event type
func (db *DB) GetStatsByDate(eventType, startDate, endDate string) (map[string]int, error) {
	db.WaitForReady()

	query := `
	SELECT event_date, SUM(count) as total
	FROM statistics
	WHERE event_type = ? AND event_date >= ? AND event_date <= ?
	GROUP BY event_date
	ORDER BY event_date DESC
	`
	rows, err := db.Query(query, eventType, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var eventDate string
		var total int
		err := rows.Scan(&eventDate, &total)
		if err != nil {
			return nil, err
		}
		stats[eventDate] = total
	}

	return stats, nil
}

// GetDailyStatsForPeriod retrieves daily statistics for all event types in a period
func (db *DB) GetDailyStatsForPeriod(startDate, endDate string) (map[string]map[string]int, error) {
	db.WaitForReady()

	query := `
	SELECT event_date, event_type, count
	FROM statistics
	WHERE event_date >= ? AND event_date <= ?
	ORDER BY event_date DESC, event_type ASC
	`
	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Result structure: map[date][event_type]count
	result := make(map[string]map[string]int)
	for rows.Next() {
		var eventDate, eventType string
		var count int
		err := rows.Scan(&eventDate, &eventType, &count)
		if err != nil {
			return nil, err
		}

		if result[eventDate] == nil {
			result[eventDate] = make(map[string]int)
		}
		result[eventDate][eventType] = count
	}

	return result, nil
}

// GetTotalStats retrieves all-time total statistics
func (db *DB) GetTotalStats() (map[string]int, error) {
	db.WaitForReady()

	query := `
	SELECT event_type, SUM(count) as total
	FROM statistics
	GROUP BY event_type
	`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var eventType string
		var total int
		err := rows.Scan(&eventType, &total)
		if err != nil {
			return nil, err
		}
		stats[eventType] = total
	}

	return stats, nil
}

// Event type constants
const (
	StatEventFeedRefresh     = "feed_refresh"
	StatEventArticleRead     = "article_read"
	StatEventArticleView     = "article_view"
	StatEventAIChat          = "ai_chat"
	StatEventAISummary       = "ai_summary"
	StatEventArticleFavorite = "article_favorite"
)

// GetAvailableMonths retrieves list of months (YYYY-MM) that have statistics data
func (db *DB) GetAvailableMonths() ([]string, error) {
	db.WaitForReady()

	query := `
		SELECT DISTINCT substr(event_date, 1, 7) as month
		FROM statistics
		ORDER BY month DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var months []string
	for rows.Next() {
		var month string
		if err := rows.Scan(&month); err != nil {
			return nil, err
		}
		months = append(months, month)
	}

	return months, nil
}

// ResetAllStatistics deletes all statistics data from the database
func (db *DB) ResetAllStatistics() error {
	db.WaitForReady()

	query := `DELETE FROM statistics`
	_, err := db.Exec(query)
	return err
}

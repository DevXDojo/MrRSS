package database

import (
	"database/sql"
	"strings"
)

// ArticleContent represents a cached article content entry
type ArticleContent struct {
	ID        int64
	ArticleID int64
	Content   string
	FetchedAt string
}

// GetArticleContent retrieves cached content for an article
func (db *DB) GetArticleContent(articleID int64) (string, bool, error) {
	db.WaitForReady()
	var content string
	err := db.QueryRow(
		`SELECT content FROM article_contents WHERE article_id = ?`,
		articleID,
	).Scan(&content)

	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return content, true, nil
}

// SetArticleContent stores or updates content for an article
func (db *DB) SetArticleContent(articleID int64, content string) error {
	db.WaitForReady()
	_, err := db.Exec(
		`INSERT OR REPLACE INTO article_contents (article_id, content, fetched_at)
		 VALUES (?, ?, CURRENT_TIMESTAMP)`,
		articleID, content,
	)
	return err
}

// DeleteArticleContent removes cached content for an article
func (db *DB) DeleteArticleContent(articleID int64) error {
	db.WaitForReady()
	_, err := db.Exec(
		`DELETE FROM article_contents WHERE article_id = ?`,
		articleID,
	)
	return err
}

// CleanupOldArticleContents removes article content cache entries older than maxAgeDays
func (db *DB) CleanupOldArticleContents(maxAgeDays int) (int64, error) {
	db.WaitForReady()
	result, err := db.Exec(
		`DELETE FROM article_contents WHERE fetched_at < datetime('now', '-' || ? || ' days')`,
		maxAgeDays,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetArticleContentCount returns the total number of cached article content entries
func (db *DB) GetArticleContentCount() (int64, error) {
	db.WaitForReady()
	var count int64
	err := db.QueryRow(`SELECT COUNT(*) FROM article_contents`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetArticleContentsBatch retrieves cached content for multiple articles at once
// Returns a map of article ID to content for articles that have cached content
func (db *DB) GetArticleContentsBatch(articleIDs []int64) (map[int64]string, error) {
	db.WaitForReady()
	if len(articleIDs) == 0 {
		return make(map[int64]string), nil
	}

	// Build query with placeholders
	placeholders := make([]string, len(articleIDs))
	args := make([]interface{}, len(articleIDs))
	for i, id := range articleIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `SELECT article_id, content FROM article_contents WHERE article_id IN (` +
		strings.Join(placeholders, ",") + `)`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]string)
	for rows.Next() {
		var articleID int64
		var content string
		if err := rows.Scan(&articleID, &content); err != nil {
			return nil, err
		}
		result[articleID] = content
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

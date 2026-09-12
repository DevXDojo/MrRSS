package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// SyncAction represents the type of sync action for an article
type SyncAction string

const (
	SyncActionMarkRead   SyncAction = "mark_read"
	SyncActionMarkUnread SyncAction = "mark_unread"
	SyncActionStar       SyncAction = "star"
	SyncActionUnstar     SyncAction = "unstar"
)

// SyncQueueItem represents an item in the FreshRSS sync queue
type SyncQueueItem struct {
	ID         int64
	ArticleID  int64
	ArticleURL string
	Action     SyncAction
	CreatedAt  time.Time
	SyncedAt   *time.Time
	SyncError  *string
}

// InitFreshRSSSyncTable creates the freshrss_sync_queue table if it doesn't exist
func InitFreshRSSSyncTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS freshrss_sync_queue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_id INTEGER NOT NULL,
		article_url TEXT NOT NULL,
		sync_action TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		synced_at INTEGER,
		sync_error TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_freshrss_sync_article ON freshrss_sync_queue(article_id);
	CREATE INDEX IF NOT EXISTS idx_freshrss_sync_synced ON freshrss_sync_queue(synced_at);
	CREATE INDEX IF NOT EXISTS idx_freshrss_sync_url ON freshrss_sync_queue(article_url);
	`

	_, err := db.Exec(query)
	return err
}

// EnqueueSyncChange adds a state change to the sync queue
func (db *DB) EnqueueSyncChange(articleID int64, articleURL string, action SyncAction) error {
	db.WaitForReady()

	query := `
	INSERT INTO freshrss_sync_queue (article_id, article_url, sync_action, created_at)
	VALUES (?, ?, ?, ?)
	`

	result, err := db.Exec(query, articleID, articleURL, string(action), time.Now().Unix())
	if err != nil {
		return fmt.Errorf("enqueue sync change: %w", err)
	}

	// Get the ID of the inserted row for verification
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("[EnqueueSyncChange] Warning: could not get insert ID: %v", err)
	} else {
		log.Printf("[EnqueueSyncChange] Successfully enqueued sync item ID=%d: articleID=%d url=%s action=%s",
			id, articleID, articleURL, action)
	}

	return nil
}

// GetPendingSyncChanges retrieves all pending sync changes that haven't been synced yet
func (db *DB) GetPendingSyncChanges(limit int, providers ...string) ([]SyncQueueItem, error) {
	db.WaitForReady()
	provider := ""
	if len(providers) > 0 {
		provider = providers[0]
	}

	query := `
	SELECT id, article_id, article_url, sync_action, created_at, synced_at, sync_error
	FROM freshrss_sync_queue
	WHERE synced_at IS NULL AND (? = '' OR article_id IN (SELECT a.id FROM articles a JOIN feeds f ON f.id = a.feed_id WHERE f.is_freshrss_source = 1 AND f.sync_provider = ?))
	ORDER BY created_at ASC
	LIMIT ?
	`

	rows, err := db.Query(query, provider, provider, limit)
	if err != nil {
		return nil, fmt.Errorf("get pending sync changes: %w", err)
	}
	defer rows.Close()

	var items []SyncQueueItem
	for rows.Next() {
		var item SyncQueueItem
		var syncedAt sql.NullInt64
		var syncError sql.NullString
		var action string
		var createdAt int64

		err := rows.Scan(
			&item.ID,
			&item.ArticleID,
			&item.ArticleURL,
			&action,
			&createdAt,
			&syncedAt,
			&syncError,
		)
		if err != nil {
			return nil, fmt.Errorf("scan sync queue item: %w", err)
		}

		item.Action = SyncAction(action)
		item.CreatedAt = time.Unix(createdAt, 0)

		if syncedAt.Valid {
			t := time.Unix(syncedAt.Int64, 0)
			item.SyncedAt = &t
		}

		if syncError.Valid {
			item.SyncError = &syncError.String
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sync queue items: %w", err)
	}

	log.Printf("[GetPendingSyncChanges] Retrieved %d pending items (limit=%d)", len(items), limit)
	if len(items) > 0 {
		for i, item := range items {
			log.Printf("  [%d] ID=%d ArticleID=%d URL=%s Action=%s", i, item.ID, item.ArticleID, item.ArticleURL, item.Action)
		}
	}

	return items, nil
}

// GetPendingSyncChangesByAction retrieves pending sync changes grouped by action type
func (db *DB) GetPendingSyncChangesByAction(action SyncAction, limit int) ([]SyncQueueItem, error) {
	db.WaitForReady()

	query := `
	SELECT id, article_id, article_url, sync_action, created_at, synced_at, sync_error
	FROM freshrss_sync_queue
	WHERE synced_at IS NULL AND sync_action = ?
	ORDER BY created_at ASC
	LIMIT ?
	`

	rows, err := db.Query(query, string(action), limit)
	if err != nil {
		return nil, fmt.Errorf("get pending sync changes by action: %w", err)
	}
	defer rows.Close()

	var items []SyncQueueItem
	for rows.Next() {
		var item SyncQueueItem
		var syncedAt sql.NullInt64
		var syncError sql.NullString
		var actionStr string
		var createdAt int64

		err := rows.Scan(
			&item.ID,
			&item.ArticleID,
			&item.ArticleURL,
			&actionStr,
			&createdAt,
			&syncedAt,
			&syncError,
		)
		if err != nil {
			return nil, fmt.Errorf("scan sync queue item: %w", err)
		}

		item.Action = SyncAction(actionStr)
		item.CreatedAt = time.Unix(createdAt, 0)

		if syncedAt.Valid {
			t := time.Unix(syncedAt.Int64, 0)
			item.SyncedAt = &t
		}

		if syncError.Valid {
			item.SyncError = &syncError.String
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sync queue items: %w", err)
	}

	return items, nil
}

// MarkSynced marks sync queue items as successfully synced
func (db *DB) MarkSynced(itemIDs []int64) error {
	db.WaitForReady()

	if len(itemIDs) == 0 {
		return nil
	}

	query := `UPDATE freshrss_sync_queue SET synced_at = ? WHERE id = ?`
	now := time.Now().Unix()

	for _, id := range itemIDs {
		_, err := db.Exec(query, now, id)
		if err != nil {
			return fmt.Errorf("mark synced item %d: %w", id, err)
		}
	}

	return nil
}

// MarkSyncFailed marks a sync queue item as failed with an error message
func (db *DB) MarkSyncFailed(itemID int64, errMsg string) error {
	db.WaitForReady()

	query := `UPDATE freshrss_sync_queue SET sync_error = ? WHERE id = ?`

	_, err := db.Exec(query, errMsg, itemID)
	if err != nil {
		return fmt.Errorf("mark sync failed: %w", err)
	}

	return nil
}

// ClearPendingSyncForArticle removes all pending sync changes for a specific article
// This is useful when resolving conflicts by accepting server state
func (db *DB) ClearPendingSyncForArticle(articleID int64) error {
	db.WaitForReady()

	query := `DELETE FROM freshrss_sync_queue WHERE article_id = ? AND synced_at IS NULL`

	_, err := db.Exec(query, articleID)
	if err != nil {
		return fmt.Errorf("clear pending sync for article: %w", err)
	}

	return nil
}

// GetPendingSyncCount returns the count of pending sync changes
func (db *DB) GetPendingSyncCount(providers ...string) (int, error) {
	db.WaitForReady()
	provider := ""
	if len(providers) > 0 {
		provider = providers[0]
	}

	var count int
	query := `SELECT COUNT(*) FROM freshrss_sync_queue WHERE synced_at IS NULL AND (? = '' OR article_id IN (SELECT a.id FROM articles a JOIN feeds f ON f.id = a.feed_id WHERE f.is_freshrss_source = 1 AND f.sync_provider = ?))`

	err := db.QueryRow(query, provider, provider).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get pending sync count: %w", err)
	}

	return count, nil
}

// DeleteOldSyncedItems removes old successfully synced items from the queue
func (db *DB) DeleteOldSyncedItems(olderThan time.Duration) error {
	db.WaitForReady()

	cutoff := time.Now().Add(-olderThan).Unix()
	query := `DELETE FROM freshrss_sync_queue WHERE synced_at IS NOT NULL AND synced_at < ?`

	_, err := db.Exec(query, cutoff)
	if err != nil {
		return fmt.Errorf("delete old synced items: %w", err)
	}

	return nil
}

// GetFailedSyncItems returns sync items that failed to sync
func (db *DB) GetFailedSyncItems(limit int, providers ...string) ([]SyncQueueItem, error) {
	db.WaitForReady()
	provider := ""
	if len(providers) > 0 {
		provider = providers[0]
	}

	query := `
	SELECT id, article_id, article_url, sync_action, created_at, synced_at, sync_error
	FROM freshrss_sync_queue
	WHERE sync_error IS NOT NULL AND (? = '' OR article_id IN (SELECT a.id FROM articles a JOIN feeds f ON f.id = a.feed_id WHERE f.is_freshrss_source = 1 AND f.sync_provider = ?))
	ORDER BY created_at DESC
	LIMIT ?
	`

	rows, err := db.Query(query, provider, provider, limit)
	if err != nil {
		return nil, fmt.Errorf("get failed sync items: %w", err)
	}
	defer rows.Close()

	var items []SyncQueueItem
	for rows.Next() {
		var item SyncQueueItem
		var syncedAt sql.NullInt64
		var syncError sql.NullString
		var action string
		var createdAt int64

		err := rows.Scan(
			&item.ID,
			&item.ArticleID,
			&item.ArticleURL,
			&action,
			&createdAt,
			&syncedAt,
			&syncError,
		)
		if err != nil {
			return nil, fmt.Errorf("scan sync queue item: %w", err)
		}

		item.Action = SyncAction(action)
		item.CreatedAt = time.Unix(createdAt, 0)

		if syncedAt.Valid {
			t := time.Unix(syncedAt.Int64, 0)
			item.SyncedAt = &t
		}

		if syncError.Valid {
			item.SyncError = &syncError.String
		}

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sync queue items: %w", err)
	}

	return items, nil
}

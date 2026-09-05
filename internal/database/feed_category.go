package database

import (
	"context"
	"errors"
	"fmt"
)

var ErrSyncedCategory = errors.New("category contains FreshRSS subscriptions")

// ChangeFeedCategory dissolves a folder tree or unsubscribes its feeds atomically.
func (db *DB) ChangeFeedCategory(ctx context.Context, category, action string) (int, error) {
	if action != "dissolve" && action != "unsubscribe" || action == "dissolve" && category == "" {
		return 0, fmt.Errorf("invalid category action")
	}
	db.WaitForReady()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin category change: %w", err)
	}
	defer tx.Rollback()
	// Compare literal path prefixes: percent and underscore in names are not SQL wildcards.
	stmt, err := tx.PrepareContext(ctx, `SELECT id, COALESCE(is_freshrss_source, 0) FROM feeds WHERE COALESCE(category, '') = ? OR (? <> '' AND substr(category, 1, length(?) + 1) = ? || '/') ORDER BY position, id`)
	if err != nil {
		return 0, fmt.Errorf("prepare category selection: %w", err)
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx, category, category, category, category)
	if err != nil {
		return 0, fmt.Errorf("select category feeds: %w", err)
	}
	var ids []int64
	for rows.Next() {
		var id int64
		var synced bool
		if err := rows.Scan(&id, &synced); err != nil {
			rows.Close()
			return 0, fmt.Errorf("read category feed: %w", err)
		}
		if synced {
			rows.Close()
			return 0, ErrSyncedCategory
		}
		ids = append(ids, id)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return 0, fmt.Errorf("read category feeds: %w", err)
	}
	if action == "dissolve" {
		var position int
		if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(position), -1) + 1 FROM feeds WHERE COALESCE(category, '') = ''`).Scan(&position); err != nil {
			return 0, fmt.Errorf("get uncategorized position: %w", err)
		}
		update, err := tx.PrepareContext(ctx, `UPDATE feeds SET category = '', position = ? WHERE id = ?`)
		if err != nil {
			return 0, fmt.Errorf("prepare category dissolution: %w", err)
		}
		defer update.Close()
		for i, id := range ids {
			if _, err := update.ExecContext(ctx, position+i, id); err != nil {
				return 0, fmt.Errorf("dissolve category: %w", err)
			}
		}
	} else {
		articles, err := tx.PrepareContext(ctx, `DELETE FROM articles WHERE feed_id = ?`)
		if err != nil {
			return 0, fmt.Errorf("prepare article deletion: %w", err)
		}
		defer articles.Close()
		feeds, err := tx.PrepareContext(ctx, `DELETE FROM feeds WHERE id = ?`)
		if err != nil {
			return 0, fmt.Errorf("prepare feed deletion: %w", err)
		}
		defer feeds.Close()
		for _, id := range ids {
			if _, err := articles.ExecContext(ctx, id); err != nil {
				return 0, fmt.Errorf("delete category articles: %w", err)
			}
			if _, err := feeds.ExecContext(ctx, id); err != nil {
				return 0, fmt.Errorf("delete category feeds: %w", err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit category change: %w", err)
	}
	return len(ids), nil
}

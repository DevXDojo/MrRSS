package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/andybalholm/cascadia"
	"strings"
)

type FeedContentOptions struct {
	ContentSelector string `json:"content_selector"`
	RemoveSelector  string `json:"remove_selector"`
}

func (o FeedContentOptions) Validate() error {
	for _, selector := range []string{o.ContentSelector, o.RemoveSelector} {
		if len(selector) > 4096 {
			return fmt.Errorf("selector exceeds 4096 bytes")
		}
		if strings.TrimSpace(selector) != "" {
			if _, err := cascadia.Compile(selector); err != nil {
				return fmt.Errorf("invalid CSS selector: %w", err)
			}
		}
	}
	return nil
}

func (db *DB) GetFeedContentOptions(ctx context.Context, feedID int64) (FeedContentOptions, error) {
	db.WaitForReady()
	var options FeedContentOptions
	stmt, err := db.PrepareContext(ctx, `SELECT content_selector, remove_selector FROM feed_content_options WHERE feed_id = ?`)
	if err != nil {
		return options, fmt.Errorf("prepare content options: %w", err)
	}
	defer stmt.Close()
	err = stmt.QueryRowContext(ctx, feedID).Scan(&options.ContentSelector, &options.RemoveSelector)
	if errors.Is(err, sql.ErrNoRows) {
		return options, nil
	}
	return options, err
}

func (db *DB) SetFeedContentOptions(ctx context.Context, feedID int64, options FeedContentOptions) error {
	db.WaitForReady()
	if err := options.Validate(); err != nil {
		return err
	}
	stmt, err := db.PrepareContext(ctx, `INSERT INTO feed_content_options (feed_id, content_selector, remove_selector) VALUES (?, ?, ?)
 ON CONFLICT(feed_id) DO UPDATE SET content_selector = excluded.content_selector, remove_selector = excluded.remove_selector`)
	if err != nil {
		return fmt.Errorf("prepare content options: %w", err)
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, feedID, options.ContentSelector, options.RemoveSelector)
	return err
}

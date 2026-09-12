// Package tray provides the unread indicator independently of the desktop shell.
package tray

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"
)

type UnreadSource interface {
	GetSettingContext(context.Context, string) (string, error)
	GetTotalUnreadCountContext(context.Context) (int, error)
}

// UnreadLabel hides both zero counts and counts disabled in the reader settings.
func UnreadLabel(ctx context.Context, source UnreadSource) (string, error) {
	showCounts, err := source.GetSettingContext(ctx, "show_unread_counts")
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if showCounts == "false" {
		return "", nil
	}
	count, err := source.GetTotalUnreadCountContext(ctx)
	if err != nil || count <= 0 {
		return "", err
	}
	return strconv.Itoa(count), nil
}

// WatchUnread keeps the tray current even while the WebView is hidden. Only
// changed, successfully read values are applied; a transient error keeps the
// last known count. The caller marshals apply to its native UI thread.
func WatchUnread(ctx context.Context, source UnreadSource, apply func(string)) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	previous := ""
	for {
		if ctx.Err() != nil {
			return
		}
		readCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		label, err := UnreadLabel(readCtx, source)
		cancel()
		if err == nil && ctx.Err() == nil && label != previous {
			apply(label)
			previous = label
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

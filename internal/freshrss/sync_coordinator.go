package freshrss

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"MrRSS/internal/database"
)

var ErrSyncRunning = errors.New("reader sync is already running")
var activeReaderSyncs sync.Map

type readerSyncKey struct {
	db       *database.DB
	provider string
}

func AcquireSync(db *database.DB, provider string) (release func(), acquired bool) {
	key := readerSyncKey{db, provider}
	if _, running := activeReaderSyncs.LoadOrStore(key, true); running {
		return nil, false
	}
	return func() { activeReaderSyncs.Delete(key) }, true
}

func IsSyncRunning(db *database.DB, provider string) bool {
	_, running := activeReaderSyncs.Load(readerSyncKey{db, provider})
	return running
}

// StartSync shares validation, concurrency and cancellation between manual and
// unattended syncs. A provider can run only one full or single-feed sync.
func StartSync(parent context.Context, db *database.DB, provider, streamID string) error {
	if !database.ValidReaderProvider(provider) {
		return fmt.Errorf("invalid reader provider")
	}
	if parent == nil {
		parent = context.Background()
	}
	if err := parent.Err(); err != nil {
		return err
	}
	enabled, err := db.GetSetting(provider + "_enabled")
	if err != nil || enabled != "true" {
		return fmt.Errorf("reader sync is disabled")
	}
	serverURL, username, password, err := db.GetReaderConfig(provider)
	if err != nil || serverURL == "" || username == "" || password == "" {
		return fmt.Errorf("reader settings incomplete")
	}
	if streamID != "" {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM feeds WHERE is_freshrss_source = 1 AND sync_provider = ? AND freshrss_stream_id = ?", provider, streamID).Scan(&count); err != nil || count == 0 {
			return fmt.Errorf("stream does not belong to reader")
		}
	}
	release, acquired := AcquireSync(db, provider)
	if !acquired {
		return ErrSyncRunning
	}
	service := NewBidirectionalSyncServiceForProvider(serverURL, username, password, provider, db)
	go func() {
		defer release()
		ctx, cancel := context.WithTimeout(parent, 10*time.Minute)
		defer cancel()
		var syncErr error
		if streamID != "" {
			_, syncErr = service.SyncFeed(ctx, streamID)
		} else {
			_, syncErr = service.Sync(ctx)
		}
		if parent.Err() != nil {
			return
		}
		if streamID == "" {
			if err := db.SetSetting(provider+"_last_sync_time", time.Now().Format(time.RFC3339Nano)); err != nil {
				log.Printf("Save reader sync time: %v", err)
			}
		}
		if syncErr != nil {
			log.Printf("%s synchronization failed: %v", provider, syncErr)
		} else {
			log.Printf("%s synchronization completed", provider)
		}
	}()
	return nil
}

func readerSyncDue(now time.Time, intervalText, lastSync string, startup bool) bool {
	if startup {
		return true
	}
	minutes, err := strconv.Atoi(intervalText)
	if err != nil || minutes <= 0 || minutes > 525600 {
		return false
	}
	last, err := time.Parse(time.RFC3339Nano, lastSync)
	return err != nil || !now.Before(last.Add(time.Duration(minutes)*time.Minute))
}

// A short polling interval applies settings changes without rebuilding tickers.
// Startup synchronization is delayed to avoid the initial RSS/cleanup burst.
func RunSyncScheduler(ctx context.Context, db *database.DB) {
	startupTimer := time.NewTimer(45 * time.Second)
	defer startupTimer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-startupTimer.C:
	}
	startup := true
	lastAttempt := make(map[string]time.Time)
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		now := time.Now()
		for _, provider := range []string{"freshrss", "miniflux"} {
			enabled, _ := db.GetSetting(provider + "_enabled")
			if enabled != "true" || IsSyncRunning(db, provider) {
				continue
			}
			interval, _ := db.GetSetting(provider + "_auto_sync_interval")
			last, _ := db.GetSetting(provider + "_last_sync_time")
			onStartup, _ := db.GetSetting(provider + "_sync_on_startup")
			if !readerSyncDue(now, interval, last, startup && onStartup == "true") || now.Sub(lastAttempt[provider]) < time.Minute {
				continue
			}
			lastAttempt[provider] = now
			if err := StartSync(ctx, db, provider, ""); err != nil && !errors.Is(err, ErrSyncRunning) {
				log.Printf("Scheduled %s sync could not start: %v", provider, err)
			}
		}
		startup = false
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

package database

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestNotificationSchemaUpgradeIsRepeatableAndPreservesData(t *testing.T) {
	path := filepath.Join(t.TempDir(), "upgrade.db")
	db, err := NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`INSERT INTO feeds(id,title,url) VALUES(1,'Saved feed','https://example.test/rss'); INSERT INTO articles(id,feed_id,title,is_favorite,is_read) VALUES(1,1,'Saved favorite',1,1); DROP TABLE notification_state; DROP TABLE notification_deliveries;`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	for i := 0; i < 2; i++ {
		db, err = NewDB(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = db.Init(); err != nil {
			t.Fatal(err)
		}
		var favorite, read bool
		if err = db.QueryRow(`SELECT is_favorite,is_read FROM articles WHERE id=1`).Scan(&favorite, &read); err != nil || !favorite || !read {
			t.Fatal("existing article changed", err)
		}
		if _, err = db.NotificationDeliveries(context.Background(), false, time.Now()); err != nil {
			t.Fatal(err)
		}
		db.Close()
	}
}

func TestNotificationQueueAndCursorRollbackTogether(t *testing.T) {
	db, err := NewDB(filepath.Join(t.TempDir(), "atomic.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.Init(); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	now := time.Now()
	if err = db.SaveNotificationConfig(ctx, "", map[string]bool{"r:c": true}, nil, now); err != nil {
		t.Fatal(err)
	}
	// Simulate a failure after queue inserts but before updating the cursor.
	if _, err = db.Exec(`CREATE TRIGGER reject_cursor BEFORE UPDATE ON notification_state BEGIN SELECT RAISE(ABORT,'test failure'); END;`); err != nil {
		t.Fatal(err)
	}
	if err = db.QueueNotification(ctx, "r:c", "c", "r", "Rule", 99, now, []string{"message"}); err == nil {
		t.Fatal("expected transaction failure")
	}
	rows, err := db.NotificationDeliveries(ctx, false, now)
	if err != nil || len(rows) != 0 {
		t.Fatal("partial queue persisted", rows, err)
	}
	state, err := db.GetNotificationState(ctx, "r:c")
	if err != nil || state.Cursor != 0 {
		t.Fatal("cursor advanced despite rollback", state, err)
	}
}

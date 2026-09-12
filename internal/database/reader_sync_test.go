package database

import (
	"path/filepath"
	"testing"
)

func TestReaderProviderMigrationPreservesLegacyMiniflux(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reader.db")
	db, err := NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	for key, value := range map[string]string{"freshrss_provider": "miniflux", "freshrss_enabled": "true", "freshrss_server_url": "https://mini.example", "freshrss_username": "reader"} {
		if err := db.SetSetting(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.SetEncryptedSetting("freshrss_api_password", "test-password"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("DELETE FROM settings WHERE key = 'reader_providers_migrated'"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO feeds (id,title,url,is_freshrss_source) VALUES (1,'Remote','https://example.org/rss',1),(2,'Local','https://local.example/rss',0)"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO articles (id,feed_id,title,url,is_read,is_favorite,is_read_later,freshrss_item_id) VALUES (1,1,'Saved','https://example.org/1',1,1,1,'item/1')"); err != nil {
		t.Fatal(err)
	}
	if err := db.EnqueueSyncChange(1, "https://example.org/1", SyncActionStar); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("ALTER TABLE feeds DROP COLUMN sync_provider"); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	// Reopening exercises the real upgrade path, including settings defaults.
	for pass := 0; pass < 2; pass++ {
		db, err = NewDB(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := db.Init(); err != nil {
			t.Fatal(err)
		}
		url, user, password, err := db.GetReaderConfig("miniflux")
		if err != nil || url != "https://mini.example" || user != "reader" || password != "test-password" {
			t.Fatal("Miniflux credentials were not migrated")
		}
		if enabled, _ := db.GetSetting("freshrss_enabled"); enabled != "false" {
			t.Fatal("FreshRSS must remain disabled")
		}
		if !db.FeedSyncEnabled(1) || db.FeedSyncProvider(1) != "miniflux" || db.FeedSyncProvider(2) != "" {
			t.Fatal("incorrect feed ownership")
		}
		article, err := db.GetArticleByID(1)
		if err != nil || !article.IsRead || !article.IsFavorite || !article.IsReadLater || article.FreshRSSItemID != "item/1" {
			t.Fatal("article state changed")
		}
		if count, err := db.GetPendingSyncCount("miniflux"); err != nil || count != 1 {
			t.Fatal("queue ownership lost")
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

package database

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"MrRSS/internal/models"
)

func setupIdentityDB(t *testing.T) *DB {
	t.Helper()
	db, err := NewDB(filepath.Join(t.TempDir(), "identity.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO feeds (title,url) VALUES ('test','https://example.org/rss')"); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestArticleIdentityKeepsSameTitlePosts(t *testing.T) {
	for _, remote := range []bool{false, true} {
		t.Run(fmt.Sprint(remote), func(t *testing.T) {
			db := setupIdentityDB(t)
			var feedID int64
			if err := db.QueryRow("SELECT id FROM feeds LIMIT 1").Scan(&feedID); err != nil {
				t.Fatal(err)
			}
			var articles []*models.Article
			for i := 1; i <= 3; i++ {
				a := &models.Article{FeedID: feedID, Title: "Alice reposted", URL: fmt.Sprintf("https://example.org/post?id=%d", i), PublishedAt: time.Now(), HasValidPublishedTime: true}
				if remote {
					a.FreshRSSItemID = fmt.Sprint(i)
				}
				articles = append(articles, a)
			}
			if err := db.SaveArticles(context.Background(), articles); err != nil {
				t.Fatal(err)
			}
			if err := db.SaveArticles(context.Background(), articles); err != nil {
				t.Fatal(err)
			}
			var count int
			if err := db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count); err != nil || count != 3 {
				t.Fatalf("count=%d err=%v", count, err)
			}
		})
	}
}

func TestArticleIdentityMigrationPreservesUserData(t *testing.T) {
	db := setupIdentityDB(t)
	var feedID int64
	if err := db.QueryRow("SELECT id FROM feeds LIMIT 1").Scan(&feedID); err != nil {
		t.Fatal(err)
	}
	_, err := db.Exec(`INSERT INTO articles (id,feed_id,title,url,unique_id,is_read,is_favorite,freshrss_item_id) VALUES (101,?,'same','https://example.org/1','legacy-one',1,1,'remote-1'),(102,?,'same','https://example.org/2',NULL,0,0,'remote-2')`, feedID, feedID)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SetArticleContent(101, "keep content"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CreateChatSession(101, "keep chat"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("DELETE FROM settings WHERE key='article_identity_v2'"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := migrateArticleIdentity(db.DB); err != nil {
			t.Fatal(err)
		}
	}
	article := &models.Article{FeedID: feedID, Title: "edited title", URL: "https://example.org/1", FreshRSSItemID: "remote-1"}
	if err := db.SaveArticles(context.Background(), []*models.Article{article}); err != nil {
		t.Fatal(err)
	}
	id, err := db.GetSavedArticleID(article)
	if err != nil || id != 101 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	var read, favorite bool
	if err := db.QueryRow("SELECT is_read,is_favorite FROM articles WHERE id=101").Scan(&read, &favorite); err != nil || !read || !favorite {
		t.Fatalf("state lost: %v", err)
	}
	if content, found, err := db.GetArticleContent(101); err != nil || !found || content != "keep content" {
		t.Fatalf("content lost: %q %v", content, err)
	}
	if sessions, err := db.GetChatSessionsByArticle(101); err != nil || len(sessions) != 1 {
		t.Fatalf("chat lost: %v", err)
	}
}

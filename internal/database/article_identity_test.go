package database

import (
	"context"
	"database/sql"
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

func TestLegacyURLConstraintMigrationPreservesReferences(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "legacy.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`PRAGMA foreign_keys=ON;
	 CREATE TABLE articles (id INTEGER PRIMARY KEY, url TEXT UNIQUE, extra TEXT);
	 CREATE INDEX articles_extra ON articles(extra);
	 CREATE TABLE contents (article_id INTEGER REFERENCES articles(id) ON DELETE CASCADE, content TEXT);
	 INSERT INTO articles VALUES(7,'https://example.org/post','retained');
	 INSERT INTO contents VALUES(7,'cached article');`)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := dropLegacyArticleURLConstraint(db); err != nil {
			t.Fatal(err)
		}
	}
	var content, extra string
	if err := db.QueryRow("SELECT content,extra FROM contents JOIN articles ON articles.id=contents.article_id WHERE articles.id=7").Scan(&content, &extra); err != nil || content != "cached article" || extra != "retained" {
		t.Fatalf("lost data: %q %q %v", content, extra, err)
	}
	if _, err := db.Exec("INSERT INTO articles VALUES(8,'https://example.org/post','second')"); err != nil {
		t.Fatal(err)
	}
	var indexes int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='articles_extra'").Scan(&indexes); err != nil || indexes != 1 {
		t.Fatalf("index lost: %v", err)
	}
	if _, err := db.Exec("INSERT INTO contents VALUES(99,'invalid reference')"); err == nil {
		t.Fatal("foreign key protection was not restored")
	}
}

func TestArticleIdentityMigrationPreservesLinklessLegacyKey(t *testing.T) {
	db := setupIdentityDB(t)
	_, err := db.Exec(`INSERT INTO articles(id,feed_id,title,url,published_at,unique_id) VALUES(103,(SELECT id FROM feeds LIMIT 1),'linkless','',CURRENT_TIMESTAMP,'original-key'); DELETE FROM settings WHERE key='article_identity_v2'`)
	if err != nil {
		t.Fatal(err)
	}
	if err := migrateArticleIdentity(db.DB); err != nil {
		t.Fatal(err)
	}
	var key string
	if err := db.QueryRow("SELECT unique_id FROM articles WHERE id=103").Scan(&key); err != nil || key != "original-key" {
		t.Fatalf("key=%q err=%v", key, err)
	}
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
			for _, article := range articles {
				article.Title = "Edited title"
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

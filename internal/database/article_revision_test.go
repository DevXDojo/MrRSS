package database

import (
	"MrRSS/internal/models"
	"context"
	"testing"
	"time"
)

func TestArticleRevisionAfterCommittedSaves(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	id, err := db.AddFeed(&models.Feed{Title: "Test", URL: "https://example.org/rss"})
	if err != nil {
		t.Fatal(err)
	}
	article := &models.Article{FeedID: id, Title: "A", URL: "https://example.org/a", PublishedAt: time.Now(), HasValidPublishedTime: true}
	if err := db.SaveArticle(article); err != nil {
		t.Fatal(err)
	}
	if db.ArticleRevision() != 1 {
		t.Fatal("insert did not advance revision")
	}
	if err := db.SaveArticle(article); err != nil {
		t.Fatal(err)
	}
	if db.ArticleRevision() != 1 {
		t.Fatal("ignored duplicate advanced revision")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := db.SaveArticles(ctx, []*models.Article{article}); err == nil {
		t.Fatal("cancelled save succeeded")
	}
	if db.ArticleRevision() != 1 {
		t.Fatal("failed save advanced revision")
	}
	if err := db.SaveArticles(context.Background(), []*models.Article{article}); err != nil {
		t.Fatal(err)
	}
	if db.ArticleRevision() != 2 {
		t.Fatal("committed batch did not advance revision")
	}
}

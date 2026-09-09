package database

import (
	"context"
	"testing"
	"time"

	"MrRSS/internal/models"
)

func TestMarkOldUnreadArticlesReadPreservesProtectedArticles(t *testing.T) {
	db, err := NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}

	feedID, err := db.AddFeed(&models.Feed{Title: "Feed", URL: "https://example.com/feed"})
	if err != nil {
		t.Fatal(err)
	}
	old := time.Now().AddDate(0, 0, -10)
	articles := []*models.Article{
		{FeedID: feedID, Title: "old", URL: "https://example.com/old", PublishedAt: old},
		{FeedID: feedID, Title: "favorite", URL: "https://example.com/favorite", PublishedAt: old, IsFavorite: true},
		{FeedID: feedID, Title: "later", URL: "https://example.com/later", PublishedAt: old, IsReadLater: true},
		{FeedID: feedID, Title: "new", URL: "https://example.com/new", PublishedAt: time.Now()},
	}
	if err := db.SaveArticles(context.Background(), articles); err != nil {
		t.Fatal(err)
	}

	count, err := db.MarkOldUnreadArticlesRead(time.Now().AddDate(0, 0, -7))
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("marked %d articles, want 1", count)
	}

	got, err := db.GetArticles("all", 0, "", true, 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	states := make(map[string]bool, len(got))
	for _, article := range got {
		states[article.Title] = article.IsRead
	}
	if !states["old"] || states["favorite"] || states["later"] || states["new"] {
		t.Fatalf("unexpected read states: %#v", states)
	}
}

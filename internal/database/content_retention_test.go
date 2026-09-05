package database_test

import (
	"MrRSS/internal/models"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestAutomaticCleanupPreservesRecentAndSavedContents(t *testing.T) {
	db := setupTestDB(t)
	id, err := db.AddFeed(&models.Feed{Title: "f", URL: "https://example.org"})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SetSetting("max_cache_size_mb", "1"); err != nil {
		t.Fatal(err)
	}
	articles := []*models.Article{}
	for i := 0; i < 3; i++ {
		articles = append(articles, &models.Article{FeedID: id, Title: fmt.Sprint(i), URL: fmt.Sprintf("https://example.org/%d", i), PublishedAt: time.Now().AddDate(-2, 0, 0), IsFavorite: i == 1, IsReadLater: i == 2})
	}
	if err := db.SaveArticles(context.Background(), articles); err != nil {
		t.Fatal(err)
	}
	ids := []int64{}
	for i := 0; i < 3; i++ {
		var articleID int64
		if err := db.QueryRow("SELECT id FROM articles WHERE title=?", fmt.Sprint(i)).Scan(&articleID); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, articleID)
		if err := db.SetArticleContent(articleID, strings.Repeat("content ", 70000)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec("UPDATE article_contents SET fetched_at=datetime('now','-60 days') WHERE article_id IN (?,?)", ids[1], ids[2]); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CleanupBySize(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.CleanupOldArticleContents(30); err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		if _, found, err := db.GetArticleContent(id); err != nil || !found {
			t.Fatalf("lost recent/saved content %d: %v", id, err)
		}
	}
}

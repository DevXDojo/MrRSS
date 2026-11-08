package database

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"MrRSS/internal/models"
)

func TestConcurrentArticleBatches(t *testing.T) {
	db := setupIdentityDB(t)
	var feedID int64
	if err := db.QueryRow("SELECT id FROM feeds LIMIT 1").Scan(&feedID); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	start := make(chan struct{})
	errs := make(chan error, 8)
	var wg sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			var articles []*models.Article
			for i := 0; i < 40; i++ {
				articles = append(articles, &models.Article{FeedID: feedID, Title: "same", URL: fmt.Sprintf("https://example.org/%d/%d", worker, i), PublishedAt: time.Now()})
			}
			<-start
			errs <- db.SaveArticles(ctx, articles)
		}(worker)
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count); err != nil || count != 320 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}

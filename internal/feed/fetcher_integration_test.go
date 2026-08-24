package feed

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/models"
)

func TestFetchFeed_SavesArticlesAndAppliesRules(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}

	// RSS server with two items; one title contains favme to match rule
	rss := `<?xml version="1.0"?><rss><channel><title>ITest</title>` +
		`<item><title>favme item</title><link>/1</link><guid>1</guid><pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate></item>` +
		`<item><title>other</title><link>/2</link><guid>2</guid><pubDate>Mon, 02 Jan 2006 15:04:05 MST</pubDate></item>` +
		`</channel></rss>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(rss))
	}))
	defer srv.Close()

	// Use fetcher
	ffetcher := NewFetcher(db)

	// Insert feed into DB
	id, err := db.AddFeed(&models.Feed{Title: "itest", URL: srv.URL})
	if err != nil {
		t.Fatalf("AddFeed error: %v", err)
	}

	// Insert a simple rule to favorite articles with title containing 'favme'
	rules := []map[string]interface{}{
		{
			"id":      1,
			"name":    "fav rule",
			"enabled": true,
			"conditions": []map[string]interface{}{
				{"field": "article_title", "operator": "contains", "value": "favme"},
			},
			"actions": []string{"favorite"},
		},
	}
	rb, _ := json.Marshal(rules)
	db.SetSetting("rules", string(rb))

	// Fetch the feed
	feedRow, err := db.GetFeedByID(id)
	if err != nil {
		t.Fatalf("GetFeedByID error: %v", err)
	}

	ffetcher.FetchFeed(context.Background(), *feedRow)

	// Allow a moment for DB writes
	time.Sleep(50 * time.Millisecond)

	articles, err := db.GetArticles("all", id, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles error: %v", err)
	}
	if len(articles) != 2 {
		t.Fatalf("expected 2 articles saved, got %d", len(articles))
	}

	// Ensure rule applied (favorite)
	favCount := 0
	for _, a := range articles {
		if a.IsFavorite {
			favCount++
		}
	}
	if favCount == 0 {
		t.Fatalf("expected at least one favorite article from rules, got 0")
	}
}

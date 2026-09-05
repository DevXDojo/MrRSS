package article_test

import (
	"MrRSS/internal/handlers/article"
	"MrRSS/internal/models"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestReloadInvalidatesBothCachesAndReadsFeedOnce(t *testing.T) {
	h := setupHandler(t)
	t.Cleanup(func() { h.Fetcher.GetTaskManager().Stop(); h.Fetcher.GetCleanupManager().Stop(); h.DB.Close() })
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/rss+xml")
		fmt.Fprint(w, `<rss version="2.0"><channel><title>f</title><item><title>Entry</title><link>https://example.org/entry</link><description><![CDATA[<p>Fresh source content</p>]]></description></item></channel></rss>`)
	}))
	defer server.Close()
	feedID, err := h.DB.AddFeed(&models.Feed{Title: "f", URL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if err := h.DB.SaveArticles(context.Background(), []*models.Article{{FeedID: feedID, Title: "Entry", URL: "https://example.org/entry", PublishedAt: time.Now()}}); err != nil {
		t.Fatal(err)
	}
	var id int64
	if err := h.DB.QueryRow("SELECT id FROM articles WHERE feed_id=?", feedID).Scan(&id); err != nil {
		t.Fatal(err)
	}
	h.DB.SetArticleContent(id, "stale")
	h.ContentCache.Set(id, "stale")
	rr := httptest.NewRecorder()
	article.HandleReloadArticleContent(h, rr, httptest.NewRequest("POST", fmt.Sprintf("/api/articles/reload-content?id=%d", id), nil))
	if rr.Code != 200 {
		t.Fatal(rr.Body.String())
	}
	content, _, err := h.GetArticleContent(id)
	if err != nil || !strings.Contains(content, "Fresh source content") || requests != 1 {
		t.Fatalf("reload content=%s requests=%d err=%v", content, requests, err)
	}
	// Previously cached empty strings must not suppress recovery either.
	h.DB.SetArticleContent(id, "")
	h.ContentCache.Set(id, "")
	content, _, err = h.GetArticleContent(id)
	if err != nil || !strings.Contains(content, "Fresh source content") || requests != 2 {
		t.Fatal("empty cache suppressed recovery", err)
	}
}
func TestArchivedArticleUsesSavedDescriptionWhenSourceFails(t *testing.T) {
	h := setupHandler(t)
	t.Cleanup(func() { h.Fetcher.GetTaskManager().Stop(); h.Fetcher.GetCleanupManager().Stop(); h.DB.Close() })
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(404) }))
	defer server.Close()
	id, err := h.DB.AddFeed(&models.Feed{Title: "f", URL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	h.DB.SetSetting("full_text_fetch_enabled", "false")
	if err := h.DB.SaveArticles(context.Background(), []*models.Article{{FeedID: id, Title: "Archived", URL: server.URL + "/article", PublishedAt: time.Now(), OriginalSummary: "<p>Saved source description</p>"}}); err != nil {
		t.Fatal(err)
	}
	var articleID int64
	h.DB.QueryRow("SELECT id FROM articles WHERE feed_id=?", id).Scan(&articleID)
	content, _, err := h.GetArticleContent(articleID)
	if err != nil || !strings.Contains(content, "Saved source description") {
		t.Fatal(content, err)
	}
}

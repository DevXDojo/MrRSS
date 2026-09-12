package feed_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"MrRSS/internal/feed"
	fh "MrRSS/internal/handlers/feed"
	"MrRSS/internal/models"
)

func TestPreviewDoesNotSubscribeAndUsesSelectedProxy(t *testing.T) {
	h := setupHandler(t)
	defer h.DB.Close()
	requests := 0
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Host != "feed.example" {
			t.Errorf("unexpected proxy request: %s", r.URL)
		}
		_, _ = w.Write([]byte(`<rss version="2.0"><channel><title>Preview feed</title><item><title>One</title><link>https://feed.example/1</link><description>Body</description></item></channel></rss>`))
	}))
	defer proxy.Close()
	body, _ := json.Marshal(map[string]any{"url": "http://feed.example/rss", "proxy_enabled": true, "proxy_url": proxy.URL})
	w := httptest.NewRecorder()
	fh.HandlePreviewFeed(h, w, httptest.NewRequest(http.MethodPost, "/api/feeds/preview", bytes.NewReader(body)))
	if w.Code != http.StatusOK || requests != 1 {
		t.Fatalf("status/requests: %d/%d", w.Code, requests)
	}
	var preview feed.Preview
	if err := json.Unmarshal(w.Body.Bytes(), &preview); err != nil {
		t.Fatal(err)
	}
	if preview.Title != "Preview feed" || len(preview.Articles) != 1 {
		t.Fatalf("bad preview: %+v", preview)
	}
	for _, table := range []string{"feeds", "articles"} {
		var count int
		if err := h.DB.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("preview modified %s", table)
		}
	}
	id, err := h.Fetcher.AddSubscriptionWithOptions(context.Background(), models.Feed{URL: "http://feed.example/rss", ProxyEnabled: true, ProxyURL: proxy.URL})
	if err != nil || id <= 0 || requests != 2 {
		t.Fatalf("subscription did not use preview proxy: %d, %v", requests, err)
	}
}

func TestPreviewRejectsNonFeedSchemes(t *testing.T) {
	h := setupHandler(t)
	defer h.DB.Close()
	for _, url := range []string{"", "file:///private/feed.xml", "script://script.py", "email://inbox@example.com"} {
		body, _ := json.Marshal(map[string]string{"url": url})
		w := httptest.NewRecorder()
		fh.HandlePreviewFeed(h, w, httptest.NewRequest(http.MethodPost, "/api/feeds/preview", bytes.NewReader(body)))
		if w.Code != http.StatusBadRequest {
			t.Errorf("accepted %q: %d", url, w.Code)
		}
	}
}

package core

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/feed"
	"MrRSS/internal/models"

	"github.com/mmcdole/gofeed"
)

func TestNewHandler_ConstructsHandler(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init failed: %v", err)
	}

	f := feed.NewFetcher(db)
	h := NewHandler(db, f, nil, nil)

	if h.DB == nil {
		t.Fatal("Handler DB is nil")
	}
	if h.Fetcher == nil {
		t.Fatal("Handler Fetcher is nil")
	}
	if h.DiscoveryService == nil {
		t.Fatal("DiscoveryService should be initialized")
	}
}

func TestCreateArticleHTTPClientUsesFeedProxy(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init failed: %v", err)
	}

	h := NewHandler(db, feed.NewFetcher(db), nil, nil)
	client, err := h.createArticleHTTPClient(&models.Feed{
		ProxyEnabled: true,
		ProxyURL:     "http://127.0.0.1:3128",
	})
	if err != nil {
		t.Fatalf("createArticleHTTPClient failed: %v", err)
	}

	proxyURL := proxyURLFromClient(t, client)
	if proxyURL != "http://127.0.0.1:3128" {
		t.Fatalf("proxy URL = %q", proxyURL)
	}
}

func TestCreateArticleHTTPClientUsesGlobalProxyWhenFeedRequestsIt(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init failed: %v", err)
	}
	_ = db.SetSetting("proxy_enabled", "true")
	_ = db.SetSetting("proxy_type", "http")
	_ = db.SetSetting("proxy_host", "127.0.0.1")
	_ = db.SetSetting("proxy_port", "8080")

	h := NewHandler(db, feed.NewFetcher(db), nil, nil)
	client, err := h.createArticleHTTPClient(&models.Feed{ProxyEnabled: true})
	if err != nil {
		t.Fatalf("createArticleHTTPClient failed: %v", err)
	}

	proxyURL := proxyURLFromClient(t, client)
	if proxyURL != "http://127.0.0.1:8080" {
		t.Fatalf("proxy URL = %q", proxyURL)
	}
}

func proxyURLFromClient(t *testing.T, client *http.Client) string {
	t.Helper()
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport type %T", client.Transport)
	}
	if transport.Proxy == nil {
		t.Fatalf("expected proxy to be configured")
	}
	reqURL, _ := url.Parse("https://example.com/article")
	req := &http.Request{URL: reqURL}
	proxy, err := transport.Proxy(req)
	if err != nil {
		t.Fatalf("proxy function returned error: %v", err)
	}
	if proxy == nil {
		t.Fatalf("proxy function returned nil")
	}
	return proxy.String()
}

func TestFindMatchingFeedItem(t *testing.T) {
	t.Parallel()

	publishedAt := time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)
	oneMinuteLater := publishedAt.Add(time.Minute)
	h := &Handler{}

	tests := []struct {
		name      string
		article   *models.Article
		items     []*gofeed.Item
		wantIndex int
	}{
		{
			name: "empty URLs fall back to title and published time",
			article: &models.Article{
				Title:       "Target article",
				URL:         "",
				PublishedAt: publishedAt,
			},
			items: []*gofeed.Item{
				{Title: "First article", Link: "", PublishedParsed: &publishedAt},
				{Title: "Target article", Link: "", PublishedParsed: &oneMinuteLater},
			},
			wantIndex: 1,
		},
		{
			name: "matching title wins when multiple items share a URL",
			article: &models.Article{
				Title:       "Target article",
				URL:         "https://example.com/shared",
				PublishedAt: publishedAt,
			},
			items: []*gofeed.Item{
				{Title: "First article", Link: "https://example.com/shared"},
				{Title: "Target article", Link: "https://example.com/shared"},
			},
			wantIndex: 1,
		},
		{
			name: "non-empty URL remains a fallback when the title changes",
			article: &models.Article{
				Title:       "Stored title",
				URL:         "https://example.com/target?utm_source=rss",
				PublishedAt: publishedAt,
			},
			items: []*gofeed.Item{
				{Title: "First article", Link: "https://example.com/first"},
				{Title: "Updated source title", Link: "https://example.com/target"},
			},
			wantIndex: 1,
		},
		{
			name: "title-only fallback tolerates whitespace differences",
			article: &models.Article{
				Title:       "Target article",
				PublishedAt: publishedAt,
			},
			items: []*gofeed.Item{
				{Title: "First article"},
				{Title: "  Target   article  "},
			},
			wantIndex: 1,
		},
		{
			name: "unmatched article returns nil",
			article: &models.Article{
				Title:       "Missing article",
				PublishedAt: publishedAt,
			},
			items: []*gofeed.Item{
				{Title: "First article"},
				{Title: "Second article"},
			},
			wantIndex: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := h.findMatchingFeedItem(tt.article, tt.items)
			if tt.wantIndex == -1 {
				if got != nil {
					t.Fatalf("findMatchingFeedItem() = %q, want nil", got.Title)
				}
				return
			}

			if got != tt.items[tt.wantIndex] {
				t.Fatalf("findMatchingFeedItem() = %v, want item %d", got, tt.wantIndex)
			}
		})
	}
}

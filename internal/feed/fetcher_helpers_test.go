package feed

import (
	"MrRSS/internal/database"
	"MrRSS/internal/models"
	"MrRSS/internal/utils/httputil"
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/mmcdole/gofeed"
)

type MockParser struct {
	Feed *gofeed.Feed
	Err  error
}

func (m *MockParser) ParseURL(url string) (*gofeed.Feed, error) {
	return m.Feed, m.Err
}

func (m *MockParser) ParseURLWithContext(url string, ctx context.Context) (*gofeed.Feed, error) {
	return m.Feed, m.Err
}

func setupDBForFeedTests(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	return db
}

func TestGetHTTPClientProxyPrecedence(t *testing.T) {
	db := setupDBForFeedTests(t)
	f := NewFetcher(db)

	// Helper function to get the underlying http.Transport
	getTransport := func(client *http.Client) *http.Transport {
		if uat, ok := client.Transport.(*httputil.UserAgentTransport); ok {
			return uat.Original.(*http.Transport)
		}
		return client.Transport.(*http.Transport)
	}

	feed := models.Feed{ProxyEnabled: true, ProxyURL: "http://10.0.0.1:3128"}
	client, err := f.getHTTPClient(feed)
	if err != nil {
		t.Fatalf("getHTTPClient error: %v", err)
	}
	tr := getTransport(client)
	if tr.Proxy == nil {
		t.Fatalf("expected proxy function for feed-level proxy")
	}
	pu, _ := tr.Proxy(&http.Request{URL: &url.URL{Scheme: "http", Host: "example.com"}})
	if pu == nil || pu.String() != "http://10.0.0.1:3128" {
		t.Fatalf("unexpected proxy url: %v", pu)
	}

	feed2 := models.Feed{ProxyEnabled: true, ProxyURL: ""}
	db.SetSetting("proxy_enabled", "true")
	db.SetSetting("proxy_type", "http")
	db.SetSetting("proxy_host", "127.0.0.1")
	db.SetSetting("proxy_port", "8080")
	db.SetEncryptedSetting("proxy_username", "u")
	db.SetEncryptedSetting("proxy_password", "p")

	client2, err := f.getHTTPClient(feed2)
	if err != nil {
		t.Fatalf("getHTTPClient error: %v", err)
	}
	tr2 := getTransport(client2)
	if tr2.Proxy == nil {
		t.Fatalf("expected proxy function for global proxy")
	}
	pu2, _ := tr2.Proxy(&http.Request{URL: &url.URL{Scheme: "http", Host: "example.com"}})
	if pu2 == nil || pu2.Host == "" {
		t.Fatalf("unexpected global proxy url: %v", pu2)
	}

	feed3 := models.Feed{ProxyEnabled: false}
	client3, err := f.getHTTPClient(feed3)
	if err != nil {
		t.Fatalf("getHTTPClient error: %v", err)
	}
	tr3 := getTransport(client3)
	if tr3.Proxy != nil {
		if pu3, _ := tr3.Proxy(&http.Request{URL: &url.URL{Scheme: "http", Host: "example.com"}}); pu3 != nil {
			t.Fatalf("expected no proxy when disabled, got %v", pu3)
		}
	}
}

func TestSetupTranslatorSelectsNonNil(t *testing.T) {
	// This test is no longer relevant since translation was removed from feed refresh
	// Translation is now handled on-demand in the frontend
	t.Skip("Translation setup removed from Fetcher - now handled on-demand")
}

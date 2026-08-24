package feed_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"MrRSS/internal/database"
	ff "MrRSS/internal/feed"
	"MrRSS/internal/models"
)

// Test that FetchAll respects concurrency limits
func TestFetchAll_RespectsConcurrency(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db.Init: %v", err)
	}

	// Set concurrency to 2
	db.SetSetting("max_concurrent_refreshes", "2")

	var active int32
	var peak int32

	// create multiple feed servers
	for i := 0; i < 5; i++ {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&active, 1)
			cur := atomic.LoadInt32(&active)
			for {
				prev := atomic.LoadInt32(&peak)
				if cur > prev {
					if atomic.CompareAndSwapInt32(&peak, prev, cur) {
						break
					}
				} else {
					break
				}
			}
			// simulate work
			time.Sleep(150 * time.Millisecond)
			atomic.AddInt32(&active, -1)
			w.Header().Set("Content-Type", "application/rss+xml")
			w.Write([]byte(`<?xml version="1.0"?><rss><channel><title>t</title><item><title>a</title><link>http://x</link><guid>1</guid></item></channel></rss>`))
		}))
		defer srv.Close()

		// add feed entry
		_, err := db.AddFeed(&models.Feed{Title: "f", URL: srv.URL})
		if err != nil {
			t.Fatalf("AddFeed: %v", err)
		}
	}

	f := ff.NewFetcher(db)

	// Run FetchAll with context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		f.FetchAll(ctx)
		close(done)
	}()

	select {
	case <-done:
		// ok
	case <-time.After(6 * time.Second):
		t.Fatalf("FetchAll did not finish in time")
	}

	if peak > 2 {
		t.Fatalf("expected peak concurrency <= 2, got %d", peak)
	}
}

// Test that FetchAll cancels promptly when context is cancelled
func TestFetchAll_RespectsCancellation(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db.Init: %v", err)
	}

	// create one slow server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(`<?xml version="1.0"?><rss><channel><title>t</title></channel></rss>`))
	}))
	defer srv.Close()

	_, err = db.AddFeed(&models.Feed{Title: "slow", URL: srv.URL})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	f := ff.NewFetcher(db)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		f.FetchAll(ctx)
		close(done)
	}()

	// cancel quickly
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// ok
	case <-time.After(3 * time.Second):
		t.Fatalf("FetchAll did not return after cancellation")
	}
}

// Test that FetchAll skips feeds with custom refresh intervals
func TestFetchAll_SkipsCustomIntervalFeeds(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db.Init: %v", err)
	}

	var globalRefreshCount int32
	var customRefreshCount int32

	// Add 2 feeds with global interval (RefreshInterval = 0)
	for i := 0; i < 2; i++ {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&globalRefreshCount, 1)
			w.Header().Set("Content-Type", "application/rss+xml")
			w.Write([]byte(`<?xml version="1.0"?><rss><channel><title>Global Feed</title><item><title>Test</title></item></channel></rss>`))
		}))

		_, err := db.AddFeed(&models.Feed{
			Title:           "Global Feed",
			URL:             srv.URL,
			RefreshInterval: 0, // Use global setting
		})
		if err != nil {
			t.Fatalf("AddFeed: %v", err)
		}
		defer srv.Close()
	}

	// Add 3 feeds with custom interval (RefreshInterval = 60)
	for i := 0; i < 3; i++ {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&customRefreshCount, 1)
			w.Header().Set("Content-Type", "application/rss+xml")
			w.Write([]byte(`<?xml version="1.0"?><rss><channel><title>Custom Feed</title><item><title>Test</title></item></channel></rss>`))
		}))

		_, err := db.AddFeed(&models.Feed{
			Title:           "Custom Feed",
			URL:             srv.URL,
			RefreshInterval: 60, // Custom 60-minute interval
		})
		if err != nil {
			t.Fatalf("AddFeed: %v", err)
		}
		defer srv.Close()
	}

	f := ff.NewFetcher(db)

	ctx := context.Background()
	f.FetchAll(ctx)

	// Wait for tasks to complete
	time.Sleep(500 * time.Millisecond)

	// Verify that only global feeds were refreshed
	globalCount := atomic.LoadInt32(&globalRefreshCount)
	customCount := atomic.LoadInt32(&customRefreshCount)

	// We expect at least 2 global feeds to be refreshed (may be more due to retries)
	if globalCount < 2 {
		t.Errorf("Expected at least 2 global feeds to be refreshed, got %d", globalCount)
	}

	// Custom interval feeds should not be refreshed at all
	if customCount != 0 {
		t.Errorf("Expected 0 custom interval feeds to be refreshed during FetchAll, got %d", customCount)
	}
}

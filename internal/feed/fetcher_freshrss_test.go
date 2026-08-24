package feed_test

import (
	"context"
	"testing"
	"time"

	"MrRSS/internal/database"
	ff "MrRSS/internal/feed"
	"MrRSS/internal/models"
)

// TestFetchAll_OnlyFreshRSSFeeds_IncrementsStatistics tests that when all feeds are FreshRSS sources,
// the global refresh still increments statistics and updates last refresh time
func TestFetchAll_OnlyFreshRSSFeeds_IncrementsStatistics(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db.Init: %v", err)
	}

	// Add only FreshRSS feeds
	for i := 0; i < 3; i++ {
		_, err := db.AddFeed(&models.Feed{
			Title:            "FreshRSS Feed",
			URL:              "freshrss:stream",
			IsFreshRSSSource: true,
		})
		if err != nil {
			t.Fatalf("AddFeed: %v", err)
		}
	}

	f := ff.NewFetcher(db)

	// Get initial statistics
	initialStats, err := db.GetTotalStats()
	if err != nil {
		t.Fatalf("GetTotalStats: %v", err)
	}
	initialRefreshCount := initialStats["feed_refresh"]

	// Get initial last refresh time
	initialLastRefresh, _ := db.GetSetting("last_global_refresh")

	// Wait a bit to ensure timestamp difference
	time.Sleep(100 * time.Millisecond)

	// Run FetchAll
	ctx := context.Background()
	f.FetchAll(ctx)

	// Wait for task manager to complete
	f.GetTaskManager().Wait(5 * time.Second)

	// Verify statistics were incremented
	finalStats, err := db.GetTotalStats()
	if err != nil {
		t.Fatalf("GetTotalStats: %v", err)
	}
	finalRefreshCount := finalStats["feed_refresh"]

	if finalRefreshCount != initialRefreshCount+1 {
		t.Fatalf("Expected feed refresh count to increment from %d to %d, got %d",
			initialRefreshCount, initialRefreshCount+1, finalRefreshCount)
	}

	// Verify last refresh time was updated
	finalLastRefresh, err := db.GetSetting("last_global_refresh")
	if err != nil {
		t.Fatalf("GetSetting(last_global_refresh): %v", err)
	}

	if finalLastRefresh == "" {
		t.Fatalf("Expected last_global_refresh to be set, got empty string")
	}

	// Parse times to verify they're different
	if initialLastRefresh != "" {
		initialTime, err := time.Parse(time.RFC3339, initialLastRefresh)
		if err != nil {
			t.Fatalf("Failed to parse initial last refresh time: %v", err)
		}
		finalTime, err := time.Parse(time.RFC3339, finalLastRefresh)
		if err != nil {
			t.Fatalf("Failed to parse final last refresh time: %v", err)
		}
		if !finalTime.After(initialTime) {
			t.Fatalf("Expected last refresh time to be updated, but it wasn't")
		}
	}
}

// TestFetchAll_NoFeeds_DoesNotIncrementStatistics tests that when there are no feeds,
// the global refresh does not increment statistics
func TestFetchAll_NoFeeds_DoesNotIncrementStatistics(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db.Init: %v", err)
	}

	f := ff.NewFetcher(db)

	// Get initial statistics
	initialStats, err := db.GetTotalStats()
	if err != nil {
		t.Fatalf("GetTotalStats: %v", err)
	}
	initialRefreshCount := initialStats["feed_refresh"]

	// Run FetchAll with no feeds
	ctx := context.Background()
	f.FetchAll(ctx)

	// Wait for task manager to complete
	f.GetTaskManager().Wait(5 * time.Second)

	// Verify statistics were NOT incremented
	finalStats, err := db.GetTotalStats()
	if err != nil {
		t.Fatalf("GetTotalStats: %v", err)
	}
	finalRefreshCount := finalStats["feed_refresh"]

	if finalRefreshCount != initialRefreshCount {
		t.Fatalf("Expected feed refresh count to remain %d when no feeds, got %d",
			initialRefreshCount, finalRefreshCount)
	}
}

// TestFetchAll_MixedFeeds_IncrementsStatistics tests that with a mix of FreshRSS and regular feeds,
// the global refresh increments statistics
func TestFetchAll_MixedFeeds_IncrementsStatistics(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db.Init: %v", err)
	}

	// Add a FreshRSS feed
	_, err = db.AddFeed(&models.Feed{
		Title:            "FreshRSS Feed",
		URL:              "freshrss:stream",
		IsFreshRSSSource: true,
	})
	if err != nil {
		t.Fatalf("AddFeed FreshRSS: %v", err)
	}

	// Add a regular feed (will fail to fetch, but that's ok for this test)
	_, err = db.AddFeed(&models.Feed{
		Title:            "Regular Feed",
		URL:              "http://invalid-url-that-will-fail.local/feed.xml",
		IsFreshRSSSource: false,
	})
	if err != nil {
		t.Fatalf("AddFeed regular: %v", err)
	}

	f := ff.NewFetcher(db)

	// Get initial statistics
	initialStats, err := db.GetTotalStats()
	if err != nil {
		t.Fatalf("GetTotalStats: %v", err)
	}
	initialRefreshCount := initialStats["feed_refresh"]

	// Run FetchAll
	ctx := context.Background()
	f.FetchAll(ctx)

	// Wait for task manager to complete
	f.GetTaskManager().Wait(5 * time.Second)

	// Verify statistics were incremented
	finalStats, err := db.GetTotalStats()
	if err != nil {
		t.Fatalf("GetTotalStats: %v", err)
	}
	finalRefreshCount := finalStats["feed_refresh"]

	if finalRefreshCount != initialRefreshCount+1 {
		t.Fatalf("Expected feed refresh count to increment from %d to %d, got %d",
			initialRefreshCount, initialRefreshCount+1, finalRefreshCount)
	}
}

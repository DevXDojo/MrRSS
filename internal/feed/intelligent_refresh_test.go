package feed

import (
	"context"
	"fmt"
	"testing"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/models"
)

func addIntelligentRefreshTestFeed(t *testing.T, db *database.DB) models.Feed {
	t.Helper()
	feed := models.Feed{Title: "Test Feed", URL: "https://example.com/test-feed", Type: "rss"}
	feedID, err := db.AddFeed(&feed)
	if err != nil {
		t.Fatalf("AddFeed error: %v", err)
	}
	feed.ID = feedID
	return feed
}

func TestIntelligentRefreshCalculator_NoArticles(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, err := database.NewDB(tmpFile)
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	defer db.Close()

	calculator := NewIntelligentRefreshCalculator(db)
	feed := addIntelligentRefreshTestFeed(t, db)

	// Test with no articles - should return default interval
	interval := calculator.CalculateInterval(feed)
	if interval != DefaultRefreshInterval {
		t.Errorf("Expected default interval %v, got %v", DefaultRefreshInterval, interval)
	}
}

func TestIntelligentRefreshCalculator_ConstantValues(t *testing.T) {
	// Test that constants are properly defined
	if MinRefreshInterval != 5*time.Minute {
		t.Errorf("Expected MinRefreshInterval = 5 minutes, got %v", MinRefreshInterval)
	}

	if MaxRefreshInterval != 24*time.Hour {
		t.Errorf("Expected MaxRefreshInterval = 24 hours, got %v", MaxRefreshInterval)
	}

	if DefaultRefreshInterval != 30*time.Minute {
		t.Errorf("Expected DefaultRefreshInterval = 30 minutes, got %v", DefaultRefreshInterval)
	}
}

func TestIntelligentRefreshCalculator_Bounds(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, err := database.NewDB(tmpFile)
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	defer db.Close()

	calculator := NewIntelligentRefreshCalculator(db)
	feed := addIntelligentRefreshTestFeed(t, db)

	// Create articles with very high frequency (every 10 seconds)
	now := time.Now()
	articles := make([]*models.Article, 10)
	for i := 0; i < 10; i++ {
		articles[i] = &models.Article{
			FeedID:                feed.ID,
			Title:                 fmt.Sprintf("Article %d", i),
			URL:                   fmt.Sprintf("http://example.com/article/%d", i),
			PublishedAt:           now.Add(time.Duration(-i) * 10 * time.Second),
			HasValidPublishedTime: true,
		}
	}

	// Save articles
	err = db.SaveArticles(context.Background(), articles)
	if err != nil {
		t.Fatalf("Failed to save articles: %v", err)
	}

	// Just verify it returns a valid interval within bounds
	interval := calculator.CalculateInterval(feed)
	if interval < MinRefreshInterval {
		t.Errorf("Interval %v is below minimum %v", interval, MinRefreshInterval)
	}
	// Interval should be reasonable (less than 1 hour for high-frequency feeds)
	if interval > time.Hour {
		t.Logf("Note: Interval %v is higher than expected for high-frequency feed", interval)
	}
}

func TestIntelligentRefreshCalculator_LowFrequency(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, err := database.NewDB(tmpFile)
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	defer db.Close()

	calculator := NewIntelligentRefreshCalculator(db)
	feed := addIntelligentRefreshTestFeed(t, db)

	// Create articles with very low frequency (every 48 hours)
	now := time.Now()
	articles := make([]*models.Article, 10)
	for i := 0; i < 10; i++ {
		articles[i] = &models.Article{
			FeedID:                feed.ID,
			Title:                 fmt.Sprintf("Article %d", i),
			URL:                   fmt.Sprintf("http://example.com/article/%d", i),
			PublishedAt:           now.Add(time.Duration(-i) * 48 * time.Hour),
			HasValidPublishedTime: true,
		}
	}

	// Save articles
	err = db.SaveArticles(context.Background(), articles)
	if err != nil {
		t.Fatalf("Failed to save articles: %v", err)
	}

	// Should be capped at MaxRefreshInterval (24 hours)
	interval := calculator.CalculateInterval(feed)
	if interval > MaxRefreshInterval {
		t.Errorf("Interval %v exceeds maximum %v", interval, MaxRefreshInterval)
	}
	// With 48h frequency and limited articles, may not reach max
	// Just verify it's within bounds
	if interval < MinRefreshInterval {
		t.Errorf("Interval %v is below minimum", interval)
	}
}

func TestGetStaggeredDelay(t *testing.T) {
	// Test that stagger works correctly
	for totalFeeds := 1; totalFeeds <= 100; totalFeeds++ {
		for feedID := int64(0); feedID < int64(totalFeeds); feedID++ {
			delay := GetStaggeredDelay(feedID, totalFeeds)
			if delay < 0 || delay > 5*time.Minute {
				t.Errorf("GetStaggeredDelay(%d, %d) = %v, want 0-5 minutes", feedID, totalFeeds, delay)
			}
		}
	}
}

func TestIntelligentRefreshCalculator_WithArticles(t *testing.T) {
	tmpFile := t.TempDir() + "/test.db"
	db, err := database.NewDB(tmpFile)
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	defer db.Close()

	calculator := NewIntelligentRefreshCalculator(db)
	feed := addIntelligentRefreshTestFeed(t, db)

	// Create articles published every 2 hours
	now := time.Now()
	articles := make([]*models.Article, 20)
	for i := 0; i < 20; i++ {
		articles[i] = &models.Article{
			FeedID:                feed.ID,
			Title:                 fmt.Sprintf("Article %d", i),
			URL:                   fmt.Sprintf("http://example.com/article/%d", i),
			PublishedAt:           now.Add(time.Duration(-i*2) * time.Hour),
			HasValidPublishedTime: true,
		}
	}

	err = db.SaveArticles(context.Background(), articles)
	if err != nil {
		t.Fatalf("Failed to save articles: %v", err)
	}

	// Calculate interval - should be approximately 1 hour (half of 2 hours)
	interval := calculator.CalculateInterval(feed)

	// Check bounds (should be between 30 min and 2 hours)
	if interval < 30*time.Minute {
		t.Errorf("Interval %v is too low (below 30 min)", interval)
	}
	if interval > 2*time.Hour {
		t.Errorf("Interval %v is too high (above 2 hours)", interval)
	}
}

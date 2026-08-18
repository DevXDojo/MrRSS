package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"MrRSS/internal/models"
)

func TestDatabaseInitialization(t *testing.T) {
	// Create temporary database
	dbFile := "test_init.db"
	defer os.Remove(dbFile)

	// Test database creation and initialization
	db, err := NewDB(dbFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Initialize database
	err = db.Init()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Verify tables were created
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('feeds', 'articles', 'settings')").Scan(&tableCount)
	if err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}
	if tableCount != 3 {
		t.Errorf("Expected 3 tables, got %d", tableCount)
	}

	// Verify indexes were created
	var indexCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name LIKE 'idx_%'").Scan(&indexCount)
	if err != nil {
		t.Fatalf("Failed to query indexes: %v", err)
	}
	if indexCount < 8 {
		t.Errorf("Expected at least 8 indexes, got %d", indexCount)
	}

	// Schema version table removed in development - skip version check
}

func TestDatabasePerformanceWithIndexes(t *testing.T) {
	// Create temporary database
	dbFile := "test_perf.db"
	defer os.Remove(dbFile)

	db, err := NewDB(dbFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	err = db.Init()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Add test feed
	feed := &models.Feed{
		Title:       "Test Feed",
		URL:         "https://example.com/feed",
		Description: "Test Description",
		Category:    "test",
	}
	_, err = db.AddFeed(feed)
	if err != nil {
		t.Fatalf("Failed to add feed: %v", err)
	}

	// Get feed ID
	feeds, err := db.GetFeeds()
	if err != nil || len(feeds) == 0 {
		t.Fatalf("Failed to get feeds: %v", err)
	}
	feedID := feeds[0].ID

	// Insert many test articles
	ctx := context.Background()
	numArticles := 1000
	articles := make([]*models.Article, numArticles)
	for i := 0; i < numArticles; i++ {
		articles[i] = &models.Article{
			FeedID:      feedID,
			Title:       fmt.Sprintf("Article %d", i),
			URL:         fmt.Sprintf("https://example.com/article-%d", i),
			PublishedAt: time.Now().Add(-time.Duration(i) * time.Minute),
			IsRead:      i%2 == 0,
			IsFavorite:  i%10 == 0,
		}
	}

	// Measure insert time
	startInsert := time.Now()
	err = db.SaveArticles(ctx, articles)
	if err != nil {
		t.Fatalf("Failed to save articles: %v", err)
	}
	insertDuration := time.Since(startInsert)
	t.Logf("Inserted %d articles in %v", numArticles, insertDuration)

	// Measure query time with filter
	startQuery := time.Now()
	results, err := db.GetArticles("unread", feedID, "", false, 50, 0)
	if err != nil {
		t.Fatalf("Failed to get articles: %v", err)
	}
	queryDuration := time.Since(startQuery)
	t.Logf("Queried articles in %v, got %d results", queryDuration, len(results))

	// Query should be fast with indexes (under 50ms for 1000 articles)
	if queryDuration > 50*time.Millisecond {
		t.Logf("Warning: Query took longer than expected: %v", queryDuration)
	}

	// Test category query
	startCategoryQuery := time.Now()
	results, err = db.GetArticles("", 0, "test", false, 50, 0)
	if err != nil {
		t.Fatalf("Failed to get articles by category: %v", err)
	}
	categoryQueryDuration := time.Since(startCategoryQuery)
	t.Logf("Queried articles by category in %v, got %d results", categoryQueryDuration, len(results))

	// Test favorites query
	startFavQuery := time.Now()
	results, err = db.GetArticles("favorites", 0, "", false, 50, 0)
	if err != nil {
		t.Fatalf("Failed to get favorite articles: %v", err)
	}
	favQueryDuration := time.Since(startFavQuery)
	t.Logf("Queried favorite articles in %v, got %d results", favQueryDuration, len(results))
}

func TestMigrationIdempotency(t *testing.T) {
	// Create temporary database
	dbFile := "test_migration.db"
	defer os.Remove(dbFile)

	db, err := NewDB(dbFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Initialize database multiple times to verify idempotency
	for i := 0; i < 3; i++ {
		err = db.Init()
		if err != nil {
			t.Fatalf("Failed to initialize database on iteration %d: %v", i, err)
		}
	}

	// Schema version table removed in development - migrations not used
	// Just verify tables still exist after multiple inits
	var tableCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('feeds', 'articles', 'settings')").Scan(&tableCount)
	if err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}
	if tableCount != 3 {
		t.Errorf("Expected 3 tables after multiple inits, got %d", tableCount)
	}
}

func BenchmarkGetArticles(b *testing.B) {
	// Create temporary database
	dbFile := "bench.db"
	defer os.Remove(dbFile)

	db, err := NewDB(dbFile)
	if err != nil {
		b.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	err = db.Init()
	if err != nil {
		b.Fatalf("Failed to initialize database: %v", err)
	}

	// Add test feed
	feed := &models.Feed{
		Title:       "Bench Feed",
		URL:         "https://example.com/bench",
		Description: "Bench Description",
	}
	_, err = db.AddFeed(feed)
	if err != nil {
		b.Fatalf("Failed to add feed: %v", err)
	}

	feeds, _ := db.GetFeeds()
	feedID := feeds[0].ID

	// Insert test articles
	ctx := context.Background()
	articles := make([]*models.Article, 500)
	for i := 0; i < 500; i++ {
		articles[i] = &models.Article{
			FeedID:      feedID,
			Title:       fmt.Sprintf("Bench Article %d", i),
			URL:         fmt.Sprintf("https://example.com/bench-%d", i),
			PublishedAt: time.Now().Add(-time.Duration(i) * time.Minute),
			IsRead:      i%3 == 0,
		}
	}
	db.SaveArticles(ctx, articles)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.GetArticles("", feedID, "", false, 50, 0)
		if err != nil {
			b.Fatalf("Failed to get articles: %v", err)
		}
	}
}

func TestCleanupOldArticles(t *testing.T) {
	// Create temporary database
	dbFile := "test_cleanup.db"
	defer os.Remove(dbFile)

	db, err := NewDB(dbFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	err = db.Init()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Add test feed
	feed := &models.Feed{
		Title:       "Test Feed",
		URL:         "https://example.com/test",
		Description: "Test Description",
	}
	_, err = db.AddFeed(feed)
	if err != nil {
		t.Fatalf("Failed to add feed: %v", err)
	}

	feeds, _ := db.GetFeeds()
	feedID := feeds[0].ID

	// Insert test articles with different ages and statuses
	now := time.Now()
	articles := []*models.Article{
		// Old (>30 days) not favorited - should be deleted
		{FeedID: feedID, Title: "Old 1", URL: "https://example.com/old1", PublishedAt: now.AddDate(0, -2, 0), IsRead: false, IsFavorite: false},
		{FeedID: feedID, Title: "Old 2", URL: "https://example.com/old2", PublishedAt: now.AddDate(0, -2, 0), IsRead: true, IsFavorite: false},
		// Old (>30 days) favorited - should be kept
		{FeedID: feedID, Title: "Old Fav", URL: "https://example.com/oldfav", PublishedAt: now.AddDate(0, -2, 0), IsRead: false, IsFavorite: true},
		// Within 30 days - should all be kept
		{FeedID: feedID, Title: "Week Old Unread", URL: "https://example.com/weekold", PublishedAt: now.AddDate(0, 0, -8), IsRead: false, IsFavorite: false},
		{FeedID: feedID, Title: "Week Old Read", URL: "https://example.com/weekoldread", PublishedAt: now.AddDate(0, 0, -8), IsRead: true, IsFavorite: false},
		{FeedID: feedID, Title: "Recent", URL: "https://example.com/recent", PublishedAt: now.AddDate(0, 0, -1), IsRead: false, IsFavorite: false},
	}

	for _, article := range articles {
		err = db.SaveArticle(article)
		if err != nil {
			t.Fatalf("Failed to save article: %v", err)
		}
	}

	// Verify initial count
	allArticles, _ := db.GetArticles("", feedID, "", false, 100, 0)
	if len(allArticles) != 6 {
		t.Errorf("Expected 6 articles initially, got %d", len(allArticles))
	}
	for _, a := range allArticles {
		t.Logf("Before cleanup: %s (read: %v, fav: %v, published: %v)", a.Title, a.IsRead, a.IsFavorite, a.PublishedAt)
	}

	// Run cleanup (default is 30 days)
	count, err := db.CleanupOldArticles()
	if err != nil {
		t.Fatalf("Failed to cleanup articles: %v", err)
	}

	t.Logf("Cleaned up %d articles", count)

	// Verify cleanup results
	remainingArticles, _ := db.GetArticles("", feedID, "", false, 100, 0)
	t.Logf("Remaining articles: %d", len(remainingArticles))

	// Should have: Old Fav (1) + Week Old Unread (1) + Week Old Read (1) + Recent (1) = 4
	// Old 1 and Old 2 should be deleted (>30 days and not favorited)
	if len(remainingArticles) != 4 {
		t.Errorf("Expected 4 articles after cleanup, got %d", len(remainingArticles))
		for _, a := range remainingArticles {
			t.Logf("  - %s (read: %v, fav: %v, published: %v)", a.Title, a.IsRead, a.IsFavorite, a.PublishedAt)
		}
	}

	// Verify the right articles remain
	titles := make(map[string]bool)
	for _, a := range remainingArticles {
		titles[a.Title] = true
	}

	expectedTitles := []string{"Old Fav", "Week Old Unread", "Week Old Read", "Recent"}
	for _, expected := range expectedTitles {
		if !titles[expected] {
			t.Errorf("Expected article '%s' to remain after cleanup", expected)
		}
	}
}

func TestCleanupUnimportantArticles(t *testing.T) {
	dbFile := filepath.Join(t.TempDir(), "test_cleanup_unimportant.db")

	db, err := NewDB(dbFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	err = db.Init()
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Add test feed
	feed := &models.Feed{
		Title:       "Test Feed",
		URL:         "https://example.com/test",
		Description: "Test Description",
	}
	_, err = db.AddFeed(feed)
	if err != nil {
		t.Fatalf("Failed to add feed: %v", err)
	}

	feeds, _ := db.GetFeeds()
	feedID := feeds[0].ID

	// Insert test articles
	articles := []*models.Article{
		{FeedID: feedID, Title: "Unread Unfav", URL: "https://example.com/1", PublishedAt: time.Now(), IsRead: false, IsFavorite: false},
		{FeedID: feedID, Title: "Read Unfav", URL: "https://example.com/2", PublishedAt: time.Now(), IsRead: true, IsFavorite: false},
		{FeedID: feedID, Title: "Unread Fav", URL: "https://example.com/3", PublishedAt: time.Now(), IsRead: false, IsFavorite: true},
		{FeedID: feedID, Title: "Unread Read Later", URL: "https://example.com/4", PublishedAt: time.Now(), IsReadLater: true},
		{FeedID: feedID, Title: "Read Fav", URL: "https://example.com/5", PublishedAt: time.Now(), IsRead: true, IsFavorite: true},
	}

	for _, article := range articles {
		err = db.SaveArticle(article)
		if err != nil {
			t.Fatalf("Failed to save article: %v", err)
		}
	}

	savedArticles, err := db.GetArticles("", feedID, "", false, 100, 0)
	if err != nil {
		t.Fatalf("Failed to load saved articles: %v", err)
	}
	articleIDs := make(map[string]int64, len(savedArticles))
	for _, article := range savedArticles {
		articleIDs[article.Title] = article.ID
		if err := db.SetArticleContent(article.ID, "cached content for "+article.Title); err != nil {
			t.Fatalf("Failed to cache content for %q: %v", article.Title, err)
		}
	}
	if _, err := db.Exec(`UPDATE article_contents SET fetched_at = datetime('now', '-8 days')`); err != nil {
		t.Fatalf("Failed to age article content cache: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO translation_cache (
			source_text_hash, source_text, target_lang, translated_text, provider, created_at
		) VALUES ('old-hash', 'source', 'en', 'translated', 'test', datetime('now', '-8 days'))
	`); err != nil {
		t.Fatalf("Failed to insert translation cache: %v", err)
	}

	// Run cleanup
	count, err := db.CleanupUnimportantArticles()
	if err != nil {
		t.Fatalf("Failed to cleanup articles: %v", err)
	}

	// Should delete 1 article (Unread Unfav)
	if count != 1 {
		t.Errorf("Expected to delete 1 article, deleted %d", count)
	}

	// Verify remaining articles
	remainingArticles, _ := db.GetArticles("", feedID, "", false, 100, 0)
	if len(remainingArticles) != 4 {
		t.Errorf("Expected 4 articles after cleanup, got %d", len(remainingArticles))
	}

	// Verify the right articles remain
	titles := make(map[string]bool)
	for _, a := range remainingArticles {
		titles[a.Title] = true
	}

	expectedTitles := []string{"Read Unfav", "Unread Fav", "Unread Read Later", "Read Fav"}
	for _, expected := range expectedTitles {
		if !titles[expected] {
			t.Errorf("Expected article '%s' to remain after cleanup", expected)
		}
		content, found, err := db.GetArticleContent(articleIDs[expected])
		if err != nil || !found || content == "" {
			t.Errorf("Expected cached content for %q to remain, found=%v, err=%v", expected, found, err)
		}
	}

	if _, found, err := db.GetArticleContent(articleIDs["Unread Unfav"]); err != nil {
		t.Fatalf("Failed to check deleted article content: %v", err)
	} else if found {
		t.Error("Expected deleted article content to be removed by cascade")
	}

	var translationCacheCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM translation_cache WHERE source_text_hash = 'old-hash'`).Scan(&translationCacheCount); err != nil {
		t.Fatalf("Failed to count translation cache: %v", err)
	}
	if translationCacheCount != 1 {
		t.Fatalf("Expected unrelated translation cache to remain, got %d", translationCacheCount)
	}

	secondCount, err := db.CleanupUnimportantArticles()
	if err != nil {
		t.Fatalf("Second cleanup failed: %v", err)
	}
	if secondCount != 0 {
		t.Fatalf("Expected idempotent cleanup to delete 0 articles, deleted %d", secondCount)
	}
}

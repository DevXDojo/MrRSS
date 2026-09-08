package database

import (
	"os"
	"testing"
	"time"

	"MrRSS/internal/models"
)

func TestUnreadCounts(t *testing.T) {
	// Create temporary database
	dbFile := "test_unread.db"
	defer os.Remove(dbFile)

	db, err := NewDB(dbFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Create test feed
	feed := &models.Feed{
		Title:    "Test Feed",
		URL:      "https://example.com/feed",
		Category: "Test",
	}
	_, err = db.AddFeed(feed)
	if err != nil {
		t.Fatalf("Failed to create feed: %v", err)
	}

	// Get feed ID
	feeds, err := db.GetFeeds()
	if err != nil || len(feeds) == 0 {
		t.Fatalf("Failed to get feeds: %v", err)
	}
	feedID := feeds[0].ID

	// Add unread articles
	unreadArticles := []models.Article{
		{
			FeedID:      feedID,
			Title:       "Unread 1",
			URL:         "https://example.com/1",
			PublishedAt: time.Now(),
			IsRead:      false,
			IsFavorite:  false,
			IsHidden:    false,
		},
		{
			FeedID:      feedID,
			Title:       "Unread 2",
			URL:         "https://example.com/2",
			PublishedAt: time.Now(),
			IsRead:      false,
			IsFavorite:  false,
			IsHidden:    false,
		},
		{
			FeedID:      feedID,
			Title:       "Read Article",
			URL:         "https://example.com/3",
			PublishedAt: time.Now(),
			IsRead:      true,
			IsFavorite:  false,
			IsHidden:    false,
		},
	}

	for _, article := range unreadArticles {
		if err := db.SaveArticle(&article); err != nil {
			t.Fatalf("Failed to save article: %v", err)
		}
	}

	// Test GetTotalUnreadCount
	totalCount, err := db.GetTotalUnreadCount()
	if err != nil {
		t.Fatalf("Failed to get total unread count: %v", err)
	}
	if totalCount != 2 {
		t.Errorf("Expected 2 unread articles, got %d", totalCount)
	}

	// Test GetUnreadCountByFeed
	feedCount, err := db.GetUnreadCountByFeed(feedID)
	if err != nil {
		t.Fatalf("Failed to get unread count by feed: %v", err)
	}
	if feedCount != 2 {
		t.Errorf("Expected 2 unread articles for feed, got %d", feedCount)
	}

	// Test GetUnreadCountsForAllFeeds
	feedCounts, err := db.GetUnreadCountsForAllFeeds()
	if err != nil {
		t.Fatalf("Failed to get unread counts for all feeds: %v", err)
	}
	if feedCounts[feedID] != 2 {
		t.Errorf("Expected 2 unread articles in feed counts map, got %d", feedCounts[feedID])
	}

	// Test MarkAllAsReadForFeed
	if err := db.MarkAllAsReadForFeed(feedID); err != nil {
		t.Fatalf("Failed to mark all as read for feed: %v", err)
	}

	// Verify all are now read
	totalCount, err = db.GetTotalUnreadCount()
	if err != nil {
		t.Fatalf("Failed to get total unread count after marking: %v", err)
	}
	if totalCount != 0 {
		t.Errorf("Expected 0 unread articles after marking all as read, got %d", totalCount)
	}

	if err := db.MarkAllAsUnreadForFeed(feedID); err != nil {
		t.Fatalf("Failed to reset feed articles to unread: %v", err)
	}
	totalCount, err = db.GetTotalUnreadCount()
	if err != nil {
		t.Fatalf("Failed to get total unread count after resetting: %v", err)
	}
	if totalCount != 2 {
		t.Errorf("Expected 2 unread articles after resetting feed, got %d", totalCount)
	}
}

func TestMarkAllAsRead(t *testing.T) {
	// Create temporary database
	dbFile := "test_mark_all.db"
	defer os.Remove(dbFile)

	db, err := NewDB(dbFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Create two test feeds
	feed1 := &models.Feed{Title: "Feed 1", URL: "https://example.com/feed1", Category: "Test"}
	feed2 := &models.Feed{Title: "Feed 2", URL: "https://example.com/feed2", Category: "Test"}
	db.AddFeed(feed1)
	db.AddFeed(feed2)

	feeds, _ := db.GetFeeds()
	feed1ID := feeds[0].ID
	feed2ID := feeds[1].ID

	// Add unread articles to both feeds
	articles := []models.Article{
		{FeedID: feed1ID, Title: "F1 Unread 1", URL: "https://example.com/f1-1", PublishedAt: time.Now(), IsRead: false, IsHidden: false},
		{FeedID: feed1ID, Title: "F1 Unread 2", URL: "https://example.com/f1-2", PublishedAt: time.Now(), IsRead: false, IsHidden: false},
		{FeedID: feed2ID, Title: "F2 Unread 1", URL: "https://example.com/f2-1", PublishedAt: time.Now(), IsRead: false, IsHidden: false},
	}

	for _, article := range articles {
		db.SaveArticle(&article)
	}

	// Test MarkAllAsRead (global)
	if err := db.MarkAllAsRead(); err != nil {
		t.Fatalf("Failed to mark all as read: %v", err)
	}

	// Verify all feeds have 0 unread
	totalCount, _ := db.GetTotalUnreadCount()
	if totalCount != 0 {
		t.Errorf("Expected 0 unread articles after global mark all as read, got %d", totalCount)
	}
}

func TestMarkAllAsReadExcludesHidden(t *testing.T) {
	// Create temporary database
	dbFile := "test_mark_all_hidden.db"
	defer os.Remove(dbFile)

	db, err := NewDB(dbFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Create test feed
	feed := &models.Feed{Title: "Test Feed", URL: "https://example.com/feed", Category: "Test"}
	db.AddFeed(feed)

	feeds, _ := db.GetFeeds()
	feedID := feeds[0].ID

	// Add articles including hidden ones
	articles := []models.Article{
		{FeedID: feedID, Title: "Unread Visible", URL: "https://example.com/1", PublishedAt: time.Now(), IsRead: false, IsHidden: false},
		{FeedID: feedID, Title: "Unread Hidden", URL: "https://example.com/2", PublishedAt: time.Now(), IsRead: false, IsHidden: true},
	}

	for _, article := range articles {
		db.SaveArticle(&article)
	}

	// Mark all as read
	if err := db.MarkAllAsRead(); err != nil {
		t.Fatalf("Failed to mark all as read: %v", err)
	}

	// Verify visible article is now read
	visibleArticles, _ := db.GetArticles("", feedID, "", false, 100, 0)
	for _, a := range visibleArticles {
		if a.Title == "Unread Visible" && !a.IsRead {
			t.Errorf("Visible article should be marked as read")
		}
	}

	// Verify hidden article is still unread
	hiddenArticles, _ := db.GetArticles("", feedID, "", true, 100, 0)
	for _, a := range hiddenArticles {
		if a.Title == "Unread Hidden" && a.IsRead {
			t.Errorf("Hidden article should remain unread")
		}
	}
}

func TestUnreadCountsWithHiddenArticles(t *testing.T) {
	// Create temporary database
	dbFile := "test_hidden.db"
	defer os.Remove(dbFile)

	db, err := NewDB(dbFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if err := db.Init(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Create test feed
	feed := &models.Feed{Title: "Test Feed", URL: "https://example.com/feed", Category: "Test"}
	db.AddFeed(feed)

	feeds, _ := db.GetFeeds()
	feedID := feeds[0].ID

	// Add articles, including hidden ones
	articles := []models.Article{
		{FeedID: feedID, Title: "Unread Visible", URL: "https://example.com/1", PublishedAt: time.Now(), IsRead: false, IsHidden: false},
		{FeedID: feedID, Title: "Unread Hidden", URL: "https://example.com/2", PublishedAt: time.Now(), IsRead: false, IsHidden: true},
	}

	for _, article := range articles {
		db.SaveArticle(&article)
	}

	// Test that hidden articles are not counted
	totalCount, _ := db.GetTotalUnreadCount()
	if totalCount != 1 {
		t.Errorf("Expected 1 unread article (hidden should be excluded), got %d", totalCount)
	}

	feedCount, _ := db.GetUnreadCountByFeed(feedID)
	if feedCount != 1 {
		t.Errorf("Expected 1 unread article for feed (hidden should be excluded), got %d", feedCount)
	}
}

package feed

import (
	"MrRSS/internal/database"
	"context"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestFetchFeedWithAudioEnclosure(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	fetcher := NewFetcher(db)

	mockFeed := &gofeed.Feed{
		Title:       "Test Podcast",
		Description: "Test Podcast Description",
		Items: []*gofeed.Item{
			{
				Title:       "Podcast Episode 1",
				Link:        "http://test.com/episode1",
				Description: "Episode Description",
				Content:     "Episode Content",
				Enclosures: []*gofeed.Enclosure{
					{
						URL:    "https://test.com/audio/episode1.mp3",
						Type:   "audio/mpeg",
						Length: "12345678",
					},
				},
			},
		},
	}
	fetcher.fp = &MockParser{Feed: mockFeed}

	_, err = fetcher.AddSubscription("http://test.com/rss", "Test Category", "")
	if err != nil {
		t.Fatalf("AddSubscription failed: %v", err)
	}

	feeds, _ := db.GetFeeds()

	fetcher.FetchFeed(context.Background(), feeds[0])

	articles, err := db.GetArticles("", 0, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles failed: %v", err)
	}

	if len(articles) != 1 {
		t.Errorf("Expected 1 article, got %d", len(articles))
	}
	if articles[0].Title != "Podcast Episode 1" {
		t.Errorf("Expected article title 'Podcast Episode 1', got '%s'", articles[0].Title)
	}
	expectedAudioURL := "https://test.com/audio/episode1.mp3"
	if articles[0].AudioURL != expectedAudioURL {
		t.Errorf("Expected audio URL '%s', got '%s'", expectedAudioURL, articles[0].AudioURL)
	}
}

func TestFetchFeedWithImageEnclosure(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	fetcher := NewFetcher(db)

	mockFeed := &gofeed.Feed{
		Title:       "Test Feed",
		Description: "Test Description",
		Items: []*gofeed.Item{
			{
				Title:           "Article with PNG",
				Link:            "http://test.com/article1",
				Description:     "Article Description",
				Content:         "Article Content",
				PublishedParsed: &[]time.Time{time.Now().Add(time.Hour)}[0],
				Enclosures: []*gofeed.Enclosure{
					{
						URL:    "https://test.com/images/image1.png",
						Type:   "image/png",
						Length: "12345",
					},
				},
			},
			{
				Title:           "Article with JPEG",
				Link:            "http://test.com/article2",
				Description:     "Article Description",
				Content:         "Article Content",
				PublishedParsed: &[]time.Time{time.Now()}[0],
				Enclosures: []*gofeed.Enclosure{
					{
						URL:    "https://test.com/images/image2.jpg",
						Type:   "image/jpeg",
						Length: "23456",
					},
				},
			},
		},
	}
	fetcher.fp = &MockParser{Feed: mockFeed}

	_, err = fetcher.AddSubscription("http://test.com/rss", "Test Category", "")
	if err != nil {
		t.Fatalf("AddSubscription failed: %v", err)
	}

	feeds, _ := db.GetFeeds()

	fetcher.FetchFeed(context.Background(), feeds[0])

	articles, err := db.GetArticles("", 0, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles failed: %v", err)
	}

	if len(articles) != 2 {
		t.Errorf("Expected 2 articles, got %d", len(articles))
	}

	if articles[0].ImageURL != "https://test.com/images/image1.png" {
		t.Errorf("Expected PNG image URL, got '%s'", articles[0].ImageURL)
	}
	if articles[1].ImageURL != "https://test.com/images/image2.jpg" {
		t.Errorf("Expected JPEG image URL, got '%s'", articles[1].ImageURL)
	}
}

func TestFetchFeedWithMultipleEnclosures(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	fetcher := NewFetcher(db)

	mockFeed := &gofeed.Feed{
		Title:       "Test Podcast",
		Description: "Test Podcast Description",
		Items: []*gofeed.Item{
			{
				Title:       "Podcast Episode with Cover",
				Link:        "http://test.com/episode1",
				Description: "Episode Description",
				Content:     "Episode Content",
				Enclosures: []*gofeed.Enclosure{
					{
						URL:    "https://test.com/images/cover.jpg",
						Type:   "image/jpeg",
						Length: "12345",
					},
					{
						URL:    "https://test.com/audio/episode1.mp3",
						Type:   "audio/mpeg",
						Length: "98765432",
					},
				},
			},
		},
	}
	fetcher.fp = &MockParser{Feed: mockFeed}

	_, err = fetcher.AddSubscription("http://test.com/rss", "Test Category", "")
	if err != nil {
		t.Fatalf("AddSubscription failed: %v", err)
	}

	feeds, _ := db.GetFeeds()

	fetcher.FetchFeed(context.Background(), feeds[0])

	articles, err := db.GetArticles("", 0, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles failed: %v", err)
	}

	if len(articles) != 1 {
		t.Errorf("Expected 1 article, got %d", len(articles))
	}

	expectedImageURL := "https://test.com/images/cover.jpg"
	if articles[0].ImageURL != expectedImageURL {
		t.Errorf("Expected image URL '%s', got '%s'", expectedImageURL, articles[0].ImageURL)
	}

	expectedAudioURL := "https://test.com/audio/episode1.mp3"
	if articles[0].AudioURL != expectedAudioURL {
		t.Errorf("Expected audio URL '%s', got '%s'", expectedAudioURL, articles[0].AudioURL)
	}
}

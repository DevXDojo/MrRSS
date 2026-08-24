package opml

import (
	"MrRSS/internal/models"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	xmlData := `
	<opml version="1.0">
		<head>
			<title>Test Subscriptions</title>
		</head>
		<body>
			<outline text="Tech" title="Tech">
				<outline type="rss" text="Hacker News" title="Hacker News" xmlUrl="https://news.ycombinator.com/rss" htmlUrl="https://news.ycombinator.com/"/>
			</outline>
			<outline type="rss" text="Go Blog" title="Go Blog" xmlUrl="https://blog.golang.org/feed.atom"/>
		</body>
	</opml>`

	r := strings.NewReader(xmlData)
	feeds, err := Parse(r)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(feeds) != 2 {
		t.Errorf("Expected 2 feeds, got %d", len(feeds))
	}

	if feeds[0].Title != "Hacker News" {
		t.Errorf("Expected first feed title 'Hacker News', got '%s'", feeds[0].Title)
	}
	if feeds[0].Category != "Tech" {
		t.Errorf("Expected first feed category 'Tech', got '%s'", feeds[0].Category)
	}

	if feeds[1].Title != "Go Blog" {
		t.Errorf("Expected second feed title 'Go Blog', got '%s'", feeds[1].Title)
	}
	if feeds[1].Category != "" {
		t.Errorf("Expected second feed category '', got '%s'", feeds[1].Category)
	}
}

// TestParseMinifluxFormat tests Miniflux OPML export format
// Miniflux uses feedURL attribute instead of xmlUrl
func TestParseMinifluxFormat(t *testing.T) {
	xmlData := `
	<?xml version="1.0" encoding="UTF-8"?>
	<opml version="2.0">
		<head>
			<title>Miniflux Subscriptions</title>
		</head>
		<body>
			<outline text="Tech" title="Tech">
				<outline type="rss" text="Hacker News" title="Hacker News" feedURL="https://news.ycombinator.com/rss" htmlUrl="https://news.ycombinator.com/"/>
			</outline>
			<outline type="rss" text="Go Blog" title="Go Blog" feedURL="https://blog.golang.org/feed.atom"/>
		</body>
	</opml>`

	r := strings.NewReader(xmlData)
	feeds, err := Parse(r)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(feeds) != 2 {
		t.Errorf("Expected 2 feeds, got %d", len(feeds))
	}

	if feeds[0].Title != "Hacker News" {
		t.Errorf("Expected first feed title 'Hacker News', got '%s'", feeds[0].Title)
	}
	if feeds[0].URL != "https://news.ycombinator.com/rss" {
		t.Errorf("Expected first feed URL 'https://news.ycombinator.com/rss', got '%s'", feeds[0].URL)
	}
	if feeds[0].Category != "Tech" {
		t.Errorf("Expected first feed category 'Tech', got '%s'", feeds[0].Category)
	}

	if feeds[1].Title != "Go Blog" {
		t.Errorf("Expected second feed title 'Go Blog', got '%s'", feeds[1].Title)
	}
	if feeds[1].URL != "https://blog.golang.org/feed.atom" {
		t.Errorf("Expected second feed URL 'https://blog.golang.org/feed.atom', got '%s'", feeds[1].URL)
	}
}

func TestGenerate(t *testing.T) {
	feeds := []models.Feed{
		{Title: "Feed 1", URL: "http://feed1.com/rss", Category: "Cat1"},
		{Title: "Feed 2", URL: "http://feed2.com/rss", Category: ""},
	}

	data, err := Generate(feeds)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	xmlStr := string(data)
	if !strings.Contains(xmlStr, `xmlUrl="http://feed1.com/rss"`) {
		t.Error("Generated XML missing Feed 1 URL")
	}
	if !strings.Contains(xmlStr, `xmlUrl="http://feed2.com/rss"`) {
		t.Error("Generated XML missing Feed 2 URL")
	}
}

// TestParseSelfExportFormat tests that MrRSS's own exported OPML format (which includes both xmlUrl and feedURL)
// can be parsed correctly without the feedURL empty string overwriting the valid xmlUrl.
func TestParseSelfExportFormat(t *testing.T) {
	xmlData := `
	<?xml version="1.0" encoding="UTF-8"?>
	<opml version="1.0">
		<head>
			<title>MrRSS Subscriptions</title>
		</head>
		<body>
			<outline text="v2ex hot" title="v2ex hot" type="" xmlUrl="rsshub://v2ex/topics/hot" htmlUrl="" feedURL="" description="" category=""/>
			<outline text="Hacker News" title="Hacker News" xmlUrl="https://hnrss.org/frontpage" feedURL="" />
		</body>
	</opml>`

	r := strings.NewReader(xmlData)
	feeds, err := Parse(r)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(feeds) != 2 {
		t.Errorf("Expected 2 feeds, got %d", len(feeds))
	}

	if len(feeds) > 0 {
		if feeds[0].Title != "v2ex hot" {
			t.Errorf("Expected first feed title 'v2ex hot', got '%s'", feeds[0].Title)
		}
		if feeds[0].URL != "rsshub://v2ex/topics/hot" {
			t.Errorf("Expected first feed URL 'rsshub://v2ex/topics/hot', got '%s'", feeds[0].URL)
		}
	}

	if len(feeds) > 1 {
		if feeds[1].Title != "Hacker News" {
			t.Errorf("Expected second feed title 'Hacker News', got '%s'", feeds[1].Title)
		}
		if feeds[1].URL != "https://hnrss.org/frontpage" {
			t.Errorf("Expected second feed URL 'https://hnrss.org/frontpage', got '%s'", feeds[1].URL)
		}
	}
}

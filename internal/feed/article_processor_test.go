package feed

import (
	"MrRSS/internal/database"
	"MrRSS/internal/models"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
)

func TestExtractMediaThumbnail(t *testing.T) {
	tests := []struct {
		name     string
		item     *gofeed.Item
		expected string
	}{
		{
			name: "YouTube feed with media:group structure",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"group": []ext.Extension{
							{
								Name:  "group",
								Value: "",
								Children: map[string][]ext.Extension{
									"thumbnail": {
										{
											Name:  "thumbnail",
											Value: "",
											Attrs: map[string]string{
												"url":    "https://i4.ytimg.com/vi/KZcE7HgtFsA/hqdefault.jpg",
												"width":  "480",
												"height": "360",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: "https://i4.ytimg.com/vi/KZcE7HgtFsA/hqdefault.jpg",
		},
		{
			name: "Direct media:thumbnail structure",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"thumbnail": []ext.Extension{
							{
								Name:  "thumbnail",
								Value: "",
								Attrs: map[string]string{
									"url": "https://example.com/thumb.jpg",
								},
							},
						},
					},
				},
			},
			expected: "https://example.com/thumb.jpg",
		},
		{
			name:     "No media extensions",
			item:     &gofeed.Item{},
			expected: "",
		},
		{
			name: "Media extensions without thumbnail",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"title": []ext.Extension{
							{
								Name:  "title",
								Value: "Some Title",
							},
						},
					},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMediaThumbnail(tt.item)
			if result != tt.expected {
				t.Errorf("extractMediaThumbnail() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractMediaTitle(t *testing.T) {
	tests := []struct {
		name     string
		item     *gofeed.Item
		expected string
	}{
		{
			name: "YouTube feed with media:group structure",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"group": []ext.Extension{
							{
								Name:  "group",
								Value: "",
								Children: map[string][]ext.Extension{
									"title": {
										{
											Name:  "title",
											Value: "WORST Place to be a Pilot: West Papua's Extreme Bush Flying",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: "WORST Place to be a Pilot: West Papua's Extreme Bush Flying",
		},
		{
			name: "Direct media:title structure",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"title": []ext.Extension{
							{
								Name:  "title",
								Value: "Direct Media Title",
							},
						},
					},
				},
			},
			expected: "Direct Media Title",
		},
		{
			name:     "No media extensions",
			item:     &gofeed.Item{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMediaTitle(tt.item)
			if result != tt.expected {
				t.Errorf("extractMediaTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractMediaDescription(t *testing.T) {
	tests := []struct {
		name     string
		item     *gofeed.Item
		expected string
	}{
		{
			name: "YouTube feed with media:group structure",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"group": []ext.Extension{
							{
								Name:  "group",
								Value: "",
								Children: map[string][]ext.Extension{
									"description": {
										{
											Name:  "description",
											Value: "I'm joining bush pilot Matt Dearden as we fly into extreme airstrips...",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: "I'm joining bush pilot Matt Dearden as we fly into extreme airstrips...",
		},
		{
			name: "Direct media:description structure",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"description": []ext.Extension{
							{
								Name:  "description",
								Value: "Direct Media Description",
							},
						},
					},
				},
			},
			expected: "Direct Media Description",
		},
		{
			name:     "No media extensions",
			item:     &gofeed.Item{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMediaDescription(tt.item)
			if result != tt.expected {
				t.Errorf("extractMediaDescription() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractImageURLWithMediaRSS(t *testing.T) {
	tests := []struct {
		name     string
		item     *gofeed.Item
		expected string
	}{
		{
			name: "YouTube feed with media:thumbnail",
			item: &gofeed.Item{
				Extensions: ext.Extensions{
					"media": {
						"group": []ext.Extension{
							{
								Name:  "group",
								Value: "",
								Children: map[string][]ext.Extension{
									"thumbnail": {
										{
											Name:  "thumbnail",
											Value: "",
											Attrs: map[string]string{
												"url": "https://i4.ytimg.com/vi/KZcE7HgtFsA/hqdefault.jpg",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: "https://i4.ytimg.com/vi/KZcE7HgtFsA/hqdefault.jpg",
		},
		{
			name: "Item with Image takes precedence",
			item: &gofeed.Item{
				Image: &gofeed.Image{
					URL: "https://example.com/item-image.jpg",
				},
				Extensions: ext.Extensions{
					"media": {
						"group": []ext.Extension{
							{
								Name:  "group",
								Value: "",
								Children: map[string][]ext.Extension{
									"thumbnail": {
										{
											Name:  "thumbnail",
											Value: "",
											Attrs: map[string]string{
												"url": "https://i4.ytimg.com/vi/KZcE7HgtFsA/hqdefault.jpg",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: "https://example.com/item-image.jpg",
		},
		{
			name: "Fallback to enclosure",
			item: &gofeed.Item{
				Enclosures: []*gofeed.Enclosure{
					{
						URL:  "https://example.com/enclosure-image.png",
						Type: "image/png",
					},
				},
			},
			expected: "https://example.com/enclosure-image.png",
		},
		{
			name: "Fallback to HTML img tag",
			item: &gofeed.Item{
				Description: `<p>Some text</p><img src="https://example.com/html-image.jpg" alt="test">`,
			},
			expected: "https://example.com/html-image.jpg",
		},
		{
			name: "Relative URL in item.Image",
			item: &gofeed.Item{
				Image: &gofeed.Image{
					URL: "/assets/post/images/test.svg",
				},
			},
			expected: "https://example.com/assets/post/images/test.svg",
		},
		{
			name: "Relative URL in description",
			item: &gofeed.Item{
				Description: `<p>Some text</p><img src="/images/relative.jpg" alt="test">`,
			},
			expected: "https://example.com/images/relative.jpg",
		},
		{
			name: "Relative URL with path in item.Image",
			item: &gofeed.Item{
				Image: &gofeed.Image{
					URL: "images/test.png",
				},
			},
			expected: "https://example.com/images/test.png",
		},
		{
			name: "Absolute URL remains unchanged",
			item: &gofeed.Item{
				Image: &gofeed.Image{
					URL: "https://cdn.example.com/image.jpg",
				},
			},
			expected: "https://cdn.example.com/image.jpg",
		},
		{
			name:     "No image available",
			item:     &gofeed.Item{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractImageURL(tt.item, "https://example.com/feed.xml")
			if result != tt.expected {
				t.Errorf("extractImageURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProcessArticlesWithYouTubeFeed(t *testing.T) {
	// Create a mock database
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	// Create a mock fetcher
	f := &Fetcher{
		db: db,
		// translator will be nil for this test
	}

	// Create a mock feed
	feed := models.Feed{
		ID: 1,
	}

	// Create mock YouTube feed items
	publishedTime := time.Now()
	items := []*gofeed.Item{
		{
			Title: "WORST Place to be a Pilot",
			Link:  "https://www.youtube.com/watch?v=KZcE7HgtFsA",
			Extensions: ext.Extensions{
				"media": {
					"group": []ext.Extension{
						{
							Name:  "group",
							Value: "",
							Children: map[string][]ext.Extension{
								"title": {
									{
										Name:  "title",
										Value: "WORST Place to be a Pilot: West Papua's Extreme Bush Flying",
									},
								},
								"description": {
									{
										Name:  "description",
										Value: "I'm joining bush pilot Matt Dearden as we fly into some of the world's most extreme and unforgiving airstrips.",
									},
								},
								"thumbnail": {
									{
										Name:  "thumbnail",
										Value: "",
										Attrs: map[string]string{
											"url":    "https://i4.ytimg.com/vi/KZcE7HgtFsA/hqdefault.jpg",
											"width":  "480",
											"height": "360",
										},
									},
								},
							},
						},
					},
				},
			},
			PublishedParsed: &publishedTime,
		},
	}

	// Process the articles
	articlesWithContent := f.processArticles(feed, items)

	// Verify results
	if len(articlesWithContent) != 1 {
		t.Fatalf("Expected 1 article, got %d", len(articlesWithContent))
	}

	article := articlesWithContent[0].Article

	// Should use the longer media:title
	expectedTitle := "WORST Place to be a Pilot: West Papua's Extreme Bush Flying"
	if article.Title != expectedTitle {
		t.Errorf("Expected title '%s', got '%s'", expectedTitle, article.Title)
	}

	// Should extract media:thumbnail
	expectedImageURL := "https://i4.ytimg.com/vi/KZcE7HgtFsA/hqdefault.jpg"
	if article.ImageURL != expectedImageURL {
		t.Errorf("Expected image URL '%s', got '%s'", expectedImageURL, article.ImageURL)
	}

	// Should have correct URL
	expectedURL := "https://www.youtube.com/watch?v=KZcE7HgtFsA"
	if article.URL != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, article.URL)
	}

	// Should extract video URL for embedded player
	expectedVideoURL := "https://www.youtube.com/embed/KZcE7HgtFsA"
	if article.VideoURL != expectedVideoURL {
		t.Errorf("Expected video URL '%s', got '%s'", expectedVideoURL, article.VideoURL)
	}
}

func TestExtractAllImageURLsFromHTMLHandlesLazyAndQuotedVariants(t *testing.T) {
	html := `
		<p>
			<img src="https://example.com/a.jpg?x=1&amp;y=2">
			<img src='https://example.com/b.jpg'>
			<img src=https://example.com/c.jpg>
			<img src="/placeholder.gif" data-src="https://cdn.example.com/lazy.jpg">
			<img src="/placeholder.gif" data-original='https://cdn.example.com/original.jpg'>
			<img src="https://example.com/a.jpg?x=1&amp;y=2">
		</p>`

	got := ExtractAllImageURLsFromHTML(html)
	want := []string{
		"https://example.com/a.jpg?x=1&y=2",
		"https://example.com/b.jpg",
		"https://example.com/c.jpg",
		"https://cdn.example.com/lazy.jpg",
		"https://cdn.example.com/original.jpg",
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d URLs, got %d: %#v", len(want), len(got), got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("URL %d expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestExtractFirstImageURLFromHTMLUsesLazyImageWhenPresent(t *testing.T) {
	html := `<img src="/placeholder.gif" data-original="https://cdn.example.com/real.jpg">`

	if got := ExtractFirstImageURLFromHTML(html); got != "https://cdn.example.com/real.jpg" {
		t.Fatalf("expected lazy image URL, got %q", got)
	}
}

func TestExtractBilibiliVideoURL(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Bilibili RSSHub iframe in description",
			content:  `<iframe width="640" height="360" src="https://www.bilibili.com/blackboard/html5mobileplayer.html?aid=116142274314455&amp;cid=undefined&amp;bvid=BV11bAUzBEqG" frameborder="0" allowfullscreen=""></iframe><br/><img src="https://i1.hdslb.com/bfs/archive/fe5ffe1fdb3ac021529814058b04e47b74dd4468.jpg"/><br/>2026第363期 02.27 - 03.05 决战！我要夺冠！！！《下一个是谁6》06 - 完美收官`,
			expected: "https://www.bilibili.com/blackboard/html5mobileplayer.html?aid=116142274314455&cid=undefined&bvid=BV11bAUzBEqG",
		},
		{
			name:     "Bilibili iframe with single quotes",
			content:  `<iframe width='640' height='360' src='https://www.bilibili.com/blackboard/html5mobileplayer.html?aid=123&bvid=BV1234567890' frameborder='0'></iframe>`,
			expected: "https://www.bilibili.com/blackboard/html5mobileplayer.html?aid=123&bvid=BV1234567890",
		},
		{
			name:     "No iframe in content",
			content:  `<p>Just some text content</p><img src="https://example.com/image.jpg">`,
			expected: "",
		},
		{
			name:     "Empty content",
			content:  "",
			expected: "",
		},
		{
			name:     "YouTube iframe (should not match)",
			content:  `<iframe src="https://www.youtube.com/embed/KZcE7HgtFsA"></iframe>`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBilibiliVideoURL(tt.content)
			if result != tt.expected {
				t.Errorf("extractBilibiliVideoURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestProcessArticlesWithBilibiliFeed(t *testing.T) {
	// Create a mock database
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}

	// Create a mock fetcher
	f := &Fetcher{
		db: db,
	}

	// Create a mock feed
	feed := models.Feed{
		ID:  1,
		URL: "https://rsshub.example.com/bilibili/weekly",
	}

	// Create mock Bilibili RSSHub feed items
	publishedTime := time.Now()
	items := []*gofeed.Item{
		{
			Title:           "决战！我要夺冠！！！《下一个是谁6》06",
			Link:            "https://www.bilibili.com/video/BV11bAUzBEqG",
			Description:     `<iframe width="640" height="360" src="https://www.bilibili.com/blackboard/html5mobileplayer.html?aid=116142274314455&amp;cid=undefined&amp;bvid=BV11bAUzBEqG" frameborder="0" allowfullscreen=""></iframe><br/><img src="https://i1.hdslb.com/bfs/archive/fe5ffe1fdb3ac021529814058b04e47b74dd4468.jpg"/><br/>2026第363期 02.27 - 03.05 决战！我要夺冠！！！《下一个是谁6》06 - 完美收官`,
			PublishedParsed: &publishedTime,
		},
	}

	// Process the articles
	articlesWithContent := f.processArticles(feed, items)

	// Verify results
	if len(articlesWithContent) != 1 {
		t.Fatalf("Expected 1 article, got %d", len(articlesWithContent))
	}

	article := articlesWithContent[0].Article

	// Should have correct title
	expectedTitle := "决战！我要夺冠！！！《下一个是谁6》06"
	if article.Title != expectedTitle {
		t.Errorf("Expected title '%s', got '%s'", expectedTitle, article.Title)
	}

	// Should extract image from description
	expectedImageURL := "https://i1.hdslb.com/bfs/archive/fe5ffe1fdb3ac021529814058b04e47b74dd4468.jpg"
	if article.ImageURL != expectedImageURL {
		t.Errorf("Expected image URL '%s', got '%s'", expectedImageURL, article.ImageURL)
	}

	// Should have correct URL
	expectedURL := "https://www.bilibili.com/video/BV11bAUzBEqG"
	if article.URL != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, article.URL)
	}

	// Should extract Bilibili video URL
	expectedVideoURL := "https://www.bilibili.com/blackboard/html5mobileplayer.html?aid=116142274314455&cid=undefined&bvid=BV11bAUzBEqG"
	if article.VideoURL != expectedVideoURL {
		t.Errorf("Expected video URL '%s', got '%s'", expectedVideoURL, article.VideoURL)
	}

	// Content should still contain iframe (CleanHTML doesn't remove iframes)
	content := articlesWithContent[0].Content
	if !strings.Contains(content, "iframe") {
		t.Error("Content should still contain iframe tag")
	}
}

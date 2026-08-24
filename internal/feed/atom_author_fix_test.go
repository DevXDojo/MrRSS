package feed

import (
	"testing"

	"github.com/mmcdole/gofeed"
)

func TestFixFeedAuthors(t *testing.T) {
	tests := []struct {
		name     string
		rawXML   string
		feed     *gofeed.Feed
		expected []string
	}{
		{
			name: "Simple text Atom authors",
			rawXML: `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>WeRss</title>
  <entry>
    <id>1</id>
    <title>Article 1</title>
    <link href="https://example.com/1"/>
    <author>杭州网</author>
  </entry>
  <entry>
    <id>2</id>
    <title>Article 2</title>
    <link href="https://example.com/2"/>
    <author>老安大叔</author>
  </entry>
  <entry>
    <id>3</id>
    <title>Article 3</title>
    <link href="https://example.com/3"/>
    <author>杭州民建</author>
  </entry>
</feed>`,
			feed: &gofeed.Feed{
				FeedType: "atom",
				Title:    "WeRss",
				Items: []*gofeed.Item{
					{
						Title:  "Article 1",
						Link:   "https://example.com/1",
						Author: &gofeed.Person{},
					},
					{
						Title:  "Article 2",
						Link:   "https://example.com/2",
						Author: &gofeed.Person{},
					},
					{
						Title:  "Article 3",
						Link:   "https://example.com/3",
						Author: &gofeed.Person{},
					},
				},
			},
			expected: []string{"杭州网", "老安大叔", "杭州民建"},
		},
		{
			name: "Standard Atom authors (should not modify)",
			rawXML: `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Test Feed</title>
  <entry>
    <id>1</id>
    <title>Article 1</title>
    <link href="https://example.com/1"/>
    <author>
      <name>John Doe</name>
    </author>
  </entry>
</feed>`,
			feed: &gofeed.Feed{
				FeedType: "atom",
				Title:    "Test Feed",
				Items: []*gofeed.Item{
					{
						Title:  "Article 1",
						Link:   "https://example.com/1",
						Author: &gofeed.Person{Name: "John Doe"},
					},
				},
			},
			expected: []string{"John Doe"},
		},
		{
			name: "RSS feed with simple text author",
			rawXML: `<?xml version="1.0" encoding="utf-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <item>
      <title>Article 1</title>
      <link>https://example.com/1</link>
      <author>test@example.com</author>
    </item>
  </channel>
</rss>`,
			feed: &gofeed.Feed{
				FeedType: "rss",
				Title:    "Test Feed",
				Items: []*gofeed.Item{
					{
						Title:  "Article 1",
						Link:   "https://example.com/1",
						Author: &gofeed.Person{},
					},
				},
			},
			expected: []string{"test@example.com"},
		},
		{
			name: "Mixed feed - some with authors, some without",
			rawXML: `<?xml version="1.0" encoding="utf-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Test Feed</title>
  <entry>
    <id>1</id>
    <title>Article 1</title>
    <link href="https://example.com/1"/>
    <author>
      <name>Existing Author</name>
    </author>
  </entry>
  <entry>
    <id>2</id>
    <title>Article 2</title>
    <link href="https://example.com/2"/>
    <author>New Author</author>
  </entry>
</feed>`,
			feed: &gofeed.Feed{
				FeedType: "atom",
				Title:    "Test Feed",
				Items: []*gofeed.Item{
					{
						Title:  "Article 1",
						Link:   "https://example.com/1",
						Author: &gofeed.Person{Name: "Existing Author"},
					},
					{
						Title:  "Article 2",
						Link:   "https://example.com/2",
						Author: &gofeed.Person{},
					},
				},
			},
			expected: []string{"Existing Author", "New Author"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixFeedAuthors(tt.feed, tt.rawXML)

			for i, item := range tt.feed.Items {
				if i >= len(tt.expected) {
					break
				}
				expected := tt.expected[i]
				if item.Author == nil {
					if expected != "" {
						t.Errorf("Item %d: expected author to be non-nil with name '%s', but got nil", i, expected)
					}
				} else if item.Author.Name != expected {
					t.Errorf("Item %d: expected author name '%s', got '%s'", i, expected, item.Author.Name)
				}
			}
		})
	}
}

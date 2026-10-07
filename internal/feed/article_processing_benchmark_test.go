package feed

import (
	"MrRSS/internal/models"
	"github.com/mmcdole/gofeed"
	"strings"
	"testing"
)

func BenchmarkProcessTextFeed(b *testing.B) {
	content := strings.Repeat(`<p>Article text with a <a href="https://example.org">link</a> and an <img src="https://example.org/image.png"></p>`, 80)
	items := make([]*gofeed.Item, 20)
	for i := range items {
		items[i] = &gofeed.Item{Title: "Article title", Link: "https://example.org/article", Content: content}
	}
	f := &Fetcher{}
	b.ReportAllocs()
	for b.Loop() {
		f.processArticles(models.Feed{ID: 1, URL: "https://example.org/feed"}, items)
	}
}

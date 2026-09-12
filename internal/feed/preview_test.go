package feed

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestFeedPreviewSortsBoundsAndSanitizes(t *testing.T) {
	parsed := &gofeed.Feed{Title: "Feed", Description: "<p>Description</p>"}
	for i := range 25 {
		published := time.Date(2026, 1, i+1, 12, 0, 0, 0, time.UTC)
		parsed.Items = append(parsed.Items, &gofeed.Item{Title: fmt.Sprint(i), Link: "/posts/1", PublishedParsed: &published, Content: `<p>Body<img src="/image.jpg" onerror="bad()"></p><script>evil()</script>`})
	}
	preview := buildPreview(parsed, "https://example.com/feed.xml")
	if preview.Total != 25 || len(preview.Articles) != 20 || preview.Articles[0].Title != "24" || preview.Description != "Description" {
		t.Fatalf("incorrect preview: %+v", preview)
	}
	if parsed.Items[0].Title != "0" {
		t.Fatal("reordered source feed")
	}
	first := preview.Articles[0]
	if first.URL != "https://example.com/posts/1" || !strings.Contains(first.ContentHTML, "https://example.com/image.jpg") || strings.Contains(first.ContentHTML, "onerror") || strings.Contains(first.ContentHTML, "evil()") {
		t.Fatalf("unsafe article: %+v", first)
	}
	preview = buildPreview(&gofeed.Feed{Items: []*gofeed.Item{{Content: strings.Repeat("字", 200000), Link: "javascript:bad()"}}}, "https://example.com")
	if !preview.Articles[0].Truncated || preview.Articles[0].URL != "" {
		t.Fatal("did not bound or sanitize article")
	}
	if got := buildPreview(&gofeed.Feed{}, "https://example.com"); got.Articles == nil || got.Total != 0 {
		t.Fatal("empty feed must have an empty article array")
	}
}

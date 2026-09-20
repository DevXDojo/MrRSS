package textutil

import (
	"MrRSS/internal/adfilter"
	"strings"
	"testing"
)

func TestArticleFilteringPreservesSafeUnfilteredVariant(t *testing.T) {
	source := `<article><h1>Editor notes</h1><p>Editorial discussion of advertisements.</p><div class="promotion">Buy a product</div><img src="/photo.jpg" onerror="bad()"><script>bad()</script></article>`
	options := adfilter.Options{Enabled: true, Rules: "example.org##.promotion"}
	result := PrepareFilteredArticle(source, "https://example.org/story", options)
	if result.Filter.Removed != 1 || strings.Contains(result.Content, "Buy a product") || !strings.Contains(result.OriginalContent, "Buy a product") {
		t.Fatalf("%+v", result)
	}
	for _, body := range []string{result.Content, result.OriginalContent} {
		if strings.Contains(body, "onerror") || strings.Contains(body, "<script") || !strings.Contains(body, "https://example.org/photo.jpg") {
			t.Fatalf("unsafe variant %s", body)
		}
	}
	options.Enabled = false
	restored := PrepareFilteredArticle(source, "https://example.org/story", options)
	if restored.Content != result.OriginalContent || restored.Filter.Removed != 0 {
		t.Fatalf("source was modified: %+v", restored)
	}
}

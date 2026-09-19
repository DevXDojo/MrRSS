package feed

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MrRSS/internal/models"
	"github.com/antchfx/htmlquery"
)

func TestXPathPreviewPreservesPathsWithoutActiveContent(t *testing.T) {
	source := `<html><body><script>alert(1)</script><ul><li><a onclick="evil()" href="/one">One &amp; two</a><img src="https://example.com/a.png" onerror="evil()"></li><li><a href="javascript:evil()">Second</a></li></ul><iframe src="https://evil.test"></iframe><svg onload="evil()"></svg></body></html>`
	preview, err := buildXPathPreview(source, "https://example.com/")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(preview)
	if err != nil {
		t.Fatal(err)
	}
	for _, unsafe := range []string{"evil", "onclick", "onerror", "iframe", "script", "svg"} {
		if strings.Contains(string(encoded), unsafe) {
			t.Errorf("preview leaked active content %q", unsafe)
		}
	}
	doc, err := htmlquery.Parse(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	var visit func(*XPathPreviewNode)
	visit = func(node *XPathPreviewNode) {
		if node.Path != "" && node.Tag != "" {
			matched, err := htmlquery.Query(doc, node.Path)
			if err != nil || matched == nil || matched.Data != node.Tag {
				t.Errorf("invalid original path %q: %v", node.Path, err)
			}
			if node.Tag == "li" {
				items, err := htmlquery.QueryAll(doc, node.Group)
				if err != nil || len(items) != 2 {
					t.Errorf("group = %q, matches = %d", node.Group, len(items))
				}
			}
		}
		for _, child := range node.Children {
			visit(child)
		}
	}
	visit(preview)
}

func TestXPathPreviewBoundsAndCancellation(t *testing.T) {
	f := &Fetcher{}
	for _, address := range []string{"file:///etc/passwd", "javascript:alert(1)", "https://user:secret@example.com"} {
		if _, err := f.PreviewXPathPage(context.Background(), &models.Feed{URL: address}); err == nil {
			t.Fatalf("accepted %s", address)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", (2<<20)+1)))
	}))
	defer server.Close()
	if _, err := f.PreviewXPathPage(context.Background(), &models.Feed{URL: server.URL}); err == nil {
		t.Fatal("accepted oversized preview")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := f.PreviewXPathPage(ctx, &models.Feed{URL: server.URL}); err == nil {
		t.Fatal("ignored cancellation")
	}
	if _, err := buildXPathPreview(strings.Repeat("<div>", 90), "https://example.com"); err == nil {
		t.Fatal("accepted excessively deep page")
	}
}

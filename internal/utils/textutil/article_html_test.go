package textutil

import (
	"strings"
	"testing"
)

func TestArticleMarkdownAndCodeFormatting(t *testing.T) {
	got := PrepareArticleContent("# Heading\n\n| A | B |\n|---|---|\n|1|2|\n\n```go\nfmt.Println(1)\n```", "https://example.org/article")
	for _, want := range []string{"<h1", "<table>", `class="language-go"`, "fmt.Println(1)"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q: %s", want, got)
		}
	}
	example := PrepareArticleContent("<pre><code># Do not reinterpret this example</code></pre>", "https://example.org")
	if strings.Contains(example, "<h1") {
		t.Fatal(example)
	}
}
func TestArticleHTMLRemovesActiveContentAndResolvesLinks(t *testing.T) {
	got := PrepareArticleContent(`<p style="display:none" onclick=evil()>Text</p><img src="../photo.jpg" onerror=evil()><a href="java&#x73;cript:alert(1)">Bad</a><script>
evil()
</script><iframe src="https://evil.example"></iframe>`, "https://example.org/news/post")
	for _, bad := range []string{"onclick", "onerror", "javascript:", "<script", "<iframe", "style="} {
		if strings.Contains(got, bad) {
			t.Errorf("unsafe %s in %s", bad, got)
		}
	}
	if !strings.Contains(got, "https://example.org/photo.jpg") {
		t.Fatal(got)
	}
}

func TestArticleHTMLPreservesInlineRasterAndMath(t *testing.T) {
	content := `<img src="data:image/png;base64,aGVsbG8="><img data-src="/lazy.jpg"><math><mi>x</mi><mo>+</mo><mn>1</mn></math>`
	got := PrepareArticleContent(content, "https://example.org/article")
	for _, want := range []string{"data:image/png;base64,aGVsbG8=", "https://example.org/lazy.jpg", "<math>", "<mi>x</mi>"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s: %s", want, got)
		}
	}
}

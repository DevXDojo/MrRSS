package textutil

import (
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
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

func TestArticleHTMLPreservesImageNoReferrer(t *testing.T) {
	for _, policy := range []string{"no-referrer", "NO-REFERRER", " no-referrer "} {
		got := PrepareArticleContent(`<img data-src="/photo.jpg" referrerpolicy="`+policy+`" onerror="alert(1)">`, "https://example.org/article")
		if !strings.Contains(got, `referrerpolicy="no-referrer"`) || !strings.Contains(got, `src="https://example.org/photo.jpg"`) || strings.Contains(got, "onerror") {
			t.Errorf("unexpected sanitized image: %s", got)
		}
	}
	for _, content := range []string{
		`<a href="/link" referrerpolicy="no-referrer">link</a>`,
	} {
		if got := PrepareArticleContent(content, "https://example.org"); strings.Contains(got, "referrerpolicy") {
			t.Errorf("unexpected policy retained: %s", got)
		}
	}
}

func TestArticleImagesDefaultToNoReferrer(t *testing.T) {
	for _, content := range []string{
		`<img src="http://img.example/get?src=http://mmbiz.qpic.cn/photo">`,
		`<img src="/photo.jpg" referrerpolicy="unsafe-url">`,
		`<img data-src="//img.example/photo" referrerpolicy="invalid">`,
		`![photo](https://img.example/photo)`,
	} {
		// A heading makes the final case an unambiguous Markdown source.
		if strings.HasPrefix(content, "![") {
			content = "# Article\n\n" + content
		}
		got := PrepareArticleContent(content, "https://example.org/article")
		if strings.Count(got, `referrerpolicy="no-referrer"`) != 1 || strings.Contains(got, "unsafe-url") {
			t.Errorf("image must use no-referrer: %s", got)
		}
	}
}

func TestArticleImageSourcePriorityAndValidation(t *testing.T) {
	base, _ := url.Parse("https://example.org/news/article")
	for _, test := range []struct {
		name       string
		attributes map[string]string
		want       string
	}{
		{"existing lazy priority", map[string]string{"data-src": "/data.jpg", "data-original": "/original.jpg", "zoomfile": "/full.jpg", "src": "/none.gif"}, "https://example.org/data.jpg"},
		{"original lazy priority", map[string]string{"data-original": "/original.jpg", "data-lazy-src": "/lazy.jpg"}, "https://example.org/original.jpg"},
		{"remaining lazy priority", map[string]string{"data-lazy-src": "/lazy.jpg", "data-actualsrc": "/actual.jpg", "data-original-src": "/original-src.jpg"}, "https://example.org/lazy.jpg"},
		{"actual before original src", map[string]string{"data-actualsrc": "/actual.jpg", "data-original-src": "/original-src.jpg"}, "https://example.org/actual.jpg"},
		{"original src before discuz", map[string]string{"data-original-src": "/original-src.jpg", "zoomfile": "/full.jpg"}, "https://example.org/original-src.jpg"},
		{"discuz full before thumbnail", map[string]string{"zoomfile": "../full.jpg", "file": "/thumb.jpg", "src": "/none.gif"}, "https://example.org/full.jpg"},
		{"invalid candidates continue", map[string]string{"data-src": "javascript:alert(1)", "data-original": "http:///missing-host.jpg", "zoomfile": "data:image/png;base64,AAAA", "file": "//cdn.example/photo.jpg", "src": "/none.gif"}, "https://cdn.example/photo.jpg"},
		{"malformed candidates continue", map[string]string{"data-src": "%zz", "zoomfile": "#image", "file": "blob:https://example.org/id", "src": "/photo.jpg"}, "https://example.org/photo.jpg"},
		{"already decoded entity remains literal", map[string]string{"zoomfile": "/image?a=1&amp;b=2"}, "https://example.org/image?a=1&amp;b=2"},
		{"ordinary source", map[string]string{"src": "images/photo.jpg"}, "https://example.org/news/images/photo.jpg"},
		{"inline source preserved by caller", map[string]string{"src": "data:image/png;base64,aGVsbG8="}, ""},
		{"no source", map[string]string{}, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := ResolveArticleImageSource(test.attributes, base); got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
	if got := ResolveArticleImageSource(map[string]string{"zoomfile": "/relative.jpg", "file": "https://cdn.example/full.jpg"}, nil); got != "https://cdn.example/full.jpg" {
		t.Errorf("nil base must skip relative candidates: %q", got)
	}
}

func TestArticleHTMLPromotesDiscuzImagesBeforeSanitizing(t *testing.T) {
	content := `<p>Article body.</p><ignore_js_op><img style="cursor:pointer" id="aimg_5366234" src="static/image/common/none.gif" onclick="zoom(this)" zoomfile="https://att.huarenjie.com/attachment/forum/202609/29/004010ve8fzabf4ee88kua.jpg" file="https://att.huarenjie.com/attachment/forum/202609/29/004010ve8fzabf4ee88kua.jpg.thumb.jpg"></ignore_js_op><img data-src="java&#x73;cript:alert(1)" zoomfile="../full.jpg?size=large&amp;type=image" file="/thumb.jpg" src="/none.gif"><img src="data:image/png;base64,aGVsbG8="><img data-original-src="/original.jpg" src="/none.gif">`
	got := PrepareArticleContent(content, "https://example.org/news/article")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(got))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"https://att.huarenjie.com/attachment/forum/202609/29/004010ve8fzabf4ee88kua.jpg",
		"https://example.org/full.jpg?size=large&type=image",
		"data:image/png;base64,aGVsbG8=",
		"https://example.org/original.jpg",
	}
	if doc.Find("img").Length() != len(want) {
		t.Fatalf("unexpected image count: %s", got)
	}
	doc.Find("img").Each(func(i int, image *goquery.Selection) {
		if image.AttrOr("src", "") != want[i] || image.AttrOr("referrerpolicy", "") != "no-referrer" {
			t.Errorf("unexpected image %d: %s", i, got)
		}
	})
	for _, unsafe := range []string{"onclick", "zoomfile", "file=", "javascript:", "none.gif", ".thumb.jpg", "data-original-src"} {
		if strings.Contains(got, unsafe) {
			t.Errorf("unexpected %q in sanitized article: %s", unsafe, got)
		}
	}
}

package adfilter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func filterHTML(t *testing.T, source string, options Options) (string, Report) {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	report := Filter(doc, "https://news.example.org/story", options)
	output, err := doc.Find("body").Html()
	if err != nil {
		t.Fatal(err)
	}
	return output, report
}

func TestAdvertisingCorpus(t *testing.T) {
	positives := []string{
		`<ins class="adsbygoogle" data-ad-client="ca-pub-123" data-ad-slot="456">AD</ins>`,
		`<div id="div-gpt-ad-123">AD</div>`,
		`<img src="https://pagead2.googlesyndication.com/AD.png">`,
		`<iframe src="https://ad.doubleclick.net/AD"></iframe>`,
		`<img width="1" height="1" alt="" src="https://track.example/pixel.gif?AD=1">`,
		`<img width="0" height="0" src="/track?AD=1">`,
	}
	for i, source := range positives {
		t.Run(fmt.Sprintf("ad-%02d", i), func(t *testing.T) {
			result, report := filterHTML(t, `<p>Editorial paragraph.</p>`+source, Options{Enabled: true})
			if strings.Contains(result, "AD") || report.Removed != 1 || !strings.Contains(result, "Editorial paragraph.") {
				t.Fatalf("%+v %s", report, result)
			}
		})
	}
	negatives := []string{
		`<div class="ad-banner"><small>Advertisement</small><a href="https://shop.example/">AD</a></div>`,
		`<aside class="advertisement"><span>广告</span><p>AD</p></aside>`,
		`<div class="ad-container" aria-label="Advertisement">AD</div>`,
		`<section aria-label="广告"><img src="/AD.png"></section>`,
		`<figure class="sponsored-ad"><span>Sponsored</span><img src="/AD.png"></figure>`,
		`<div class="ad-unit"><a href="https://www.googleadservices.com/AD">Buy</a></div>`,
		`<article class="advertisement"><h1>How advertising works</h1><p>Advertisement</p></article>`,
		`<div class="advice">Advice about advertising and promotion</div>`,
		`<div id="download">Download the research paper</div>`,
		`<div class="header shadow gradient">Normal content</div>`,
		`<p>Sponsored by our community. Thank you!</p>`,
		`<span class="sponsored">Sponsored research: results</span>`,
		`<div class="ad">Alzheimer disease (AD) study results</div>`,
		`<div class="ad-banner">A tutorial on the ad-banner class</div>`,
		`<p><a href="https://googleadservices.com/documentation">Reference</a></p>`,
		`<img src="https://googlesyndication.com.example.org/photo.png">`,
		`<img src="https://example.org/googlesyndication.com/photo.png">`,
		`<img src="https://criteo.com/team-photo.jpg">`,
		`<img src="/banner.jpg" width="728" height="90" alt="A diagram">`,
		`<img src="/pixel-art.png" width="100" height="100">`,
		`<img src="/pixel.gif" width="1" height="1" alt="Pixel art sample">`,
		`<img src="/dot.png" width="1" height="1">`,
		`<img src="data:image/gif;base64,R0lG" width="1" height="1">`,
		`<img src="/pixel.gif" width="1" height="1" data-src="/real-photo.jpg">`,
		`<pre><code>&lt;ins class="adsbygoogle"&gt;</code><img src="https://ad.doubleclick.net/demo"></pre>`,
		`<blockquote><div class="ad"><span>Advertisement</span>Quoted campaign copy</div></blockquote>`,
		`<main id="div-gpt-ad-example"><h1>Documentation</h1><p>Ad examples explained.</p></main>`,
		`<div class="ad"><small>Advertisement</small><h2>Study of advertising</h2></div>`,
		`<div class="ad"><small>Advertisement</small>` + strings.Repeat("Editorial context. ", 50) + `</div>`,
		`<ins class="adsbygoogle">An HTML insertion describing this class</ins>`,
	}
	for i, source := range negatives {
		t.Run(fmt.Sprintf("keep-%02d", i), func(t *testing.T) {
			result, report := filterHTML(t, source, Options{Enabled: true})
			expected, _ := filterHTML(t, source, Options{})
			if result != expected || report.Removed != 0 {
				t.Fatalf("false positive: %+v\n%s\n%s", report, result, expected)
			}
		})
	}
}

func TestExceptionsAndCustomRules(t *testing.T) {
	source := `<p>Keep editorial.</p><div id="div-gpt-ad-123"><span>Advertisement</span><img class="keep" src="/picture"></div><aside class="promo">Offer</aside>`
	cases := []struct {
		options Options
		removed int
	}{
		{Options{Enabled: false}, 0},
		{Options{Enabled: true, Allowlist: "example.org"}, 0},
		{Options{Enabled: true, Allowlist: "example.org.evil"}, 1},
		{Options{Enabled: true, Rules: "example.org##.promo"}, 2},
		{Options{Enabled: true, Rules: "example.org#@#.keep\nexample.org##.promo"}, 1},
		{Options{Enabled: true, Rules: "other.example##.promo"}, 1},
		{Options{Enabled: true, Rules: "example.org##body"}, 1},
		{Options{Enabled: true, Rules: "##.promo"}, 0},
	}
	for _, c := range cases {
		_, report := filterHTML(t, source, c.options)
		if report.Removed != c.removed {
			t.Errorf("%+v: %+v", c.options, report)
		}
	}
}

func TestRuleValidation(t *testing.T) {
	for _, raw := range []string{"##.ad", "https://example.org##.ad", "example.org##[", "example.org##div:has(p)", "example.org##div:contains(offer)", "example.org##", "example.org##div{display:none}", strings.Repeat("example.org##.ad\n", 101)} {
		if _, err := ParseRules(raw); err == nil {
			t.Errorf("accepted %q", raw)
		}
	}
	if rules, err := ParseRules("! Local rules\nexample.org##.promo > a[href]\nexample.org#@#.keep"); err != nil || len(rules) != 2 {
		t.Fatalf("%v %v", rules, err)
	}
	if Validate(Options{Allowlist: "example.org\nhttps://wrong.org"}) == nil {
		t.Fatal("URL is not a domain")
	}
}

func TestBoundedAndIdempotent(t *testing.T) {
	source := strings.Repeat("<div>", 140) + `<div class="ad"><span>Advertisement</span>Keep on overflow</div>` + strings.Repeat("</div>", 140)
	result, report := filterHTML(t, source, Options{Enabled: true})
	if report.Removed != 0 || report.Guarded == 0 || !strings.Contains(result, "Keep on overflow") {
		t.Fatalf("%+v", report)
	}
	result, report = filterHTML(t, `<div id="div-gpt-ad-123">Sale</div><p>Keep</p>`, Options{Enabled: true})
	again, second := filterHTML(t, result, Options{Enabled: true})
	if result != again || report.Removed != 1 || second.Removed != 0 {
		t.Fatalf("%+v %+v", report, second)
	}
}

func BenchmarkFilter(b *testing.B) {
	source := strings.Repeat(`<section><p>Editorial content with references.</p><img src="/photo.jpg"><div class="ad"><small>Advertisement</small>Sale</div></section>`, 500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(source))
		Filter(doc, "https://news.example.org", Options{Enabled: true})
	}
}

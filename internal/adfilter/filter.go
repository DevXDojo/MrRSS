package adfilter

import (
	"net/url"
	"path"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type Report struct {
	Enabled bool           `json:"enabled"`
	Removed int            `json:"removed"`
	Reasons map[string]int `json:"reasons"`
	Guarded int            `json:"guarded"`
}

type Content struct {
	Content         string `json:"content"`
	OriginalContent string `json:"original_content,omitempty"`
	Filter          Report `json:"ad_filter"`
}

// Filter mutates only the caller's temporary parsed document. Never pass a
// shared source cache. Exceptions protect the element, its descendants, and any
// ancestor whose removal would erase that exception.
func Filter(doc *goquery.Document, base string, options Options) Report {
	report := Report{Enabled: options.Applies(base), Reasons: map[string]int{}}
	if !report.Enabled {
		return report
	}
	if Validate(options) != nil || !boundedDocument(doc) {
		report.Guarded++
		return report
	}
	rules, err := ParseRules(options.Rules)
	if err != nil {
		return report
	} // Imported invalid settings fail open, preserving content.
	protected := map[*html.Node]bool{}
	host := Host(base)
	for _, rule := range rules {
		if !rule.Exception || !domainMatches(host, rule.Domain) {
			continue
		}
		doc.FindMatcher(rule.Selector).Each(func(_ int, s *goquery.Selection) {
			protected[s.Get(0)] = true
			s.Find("*").Each(func(_ int, child *goquery.Selection) { protected[child.Get(0)] = true })
			s.Parents().Each(func(_ int, parent *goquery.Selection) { protected[parent.Get(0)] = true })
		})
	}
	custom := map[*html.Node]bool{}
	for _, rule := range rules {
		if rule.Exception || !domainMatches(host, rule.Domain) {
			continue
		}
		doc.FindMatcher(rule.Selector).Each(func(_ int, s *goquery.Selection) { custom[s.Get(0)] = true })
	}
	var visit func(*html.Node)
	visit = func(node *html.Node) {
		if node.Type != html.ElementNode {
			return
		}
		s := goquery.NewDocumentFromNode(node).Selection
		reason := ""
		if custom[node] {
			reason = "custom"
		} else if !protected[node] {
			reason = classify(s, base)
		}
		if reason != "" && !protected[node] {
			if structuralRoot(node) || (reason != "custom" && editorialGuard(s)) {
				report.Guarded++
			} else {
				if node.Parent != nil {
					node.Parent.RemoveChild(node)
				}
				report.Removed++
				report.Reasons[reason]++
				return
			}
		}
		// Code and quotations are editorial evidence, including embedded examples.
		if node.Data == "pre" || node.Data == "code" || node.Data == "blockquote" {
			return
		}
		for child := node.FirstChild; child != nil; {
			next := child.NextSibling
			visit(child)
			child = next
		}
	}
	doc.Find("html").Each(func(_ int, s *goquery.Selection) { visit(s.Get(0)) })
	return report
}

func boundedDocument(doc *goquery.Document) bool {
	count := 0
	var visit func(*html.Node, int) bool
	visit = func(n *html.Node, depth int) bool {
		count++
		if count > 20000 || depth > 128 {
			return false
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if !visit(c, depth+1) {
				return false
			}
		}
		return true
	}
	return visit(doc.Get(0), 0)
}

func structuralRoot(node *html.Node) bool {
	return node.Data == "html" || node.Data == "body" || node.Data == "head"
}

func editorialGuard(s *goquery.Selection) bool {
	tag := s.Get(0).Data
	if tag == "article" || tag == "main" || tag == "pre" || tag == "code" || tag == "blockquote" {
		return true
	}
	if s.Find("article,main,pre,code,blockquote,h1,h2").Length() > 0 {
		return true
	}
	text := strings.TrimSpace(s.Text())
	return utf8.RuneCountInString(text) > 600 || s.Find("p").Length() > 3
}

func classify(s *goquery.Selection, base string) string {
	tag := s.Get(0).Data
	if structuralRoot(s.Get(0)) {
		return ""
	}
	// Resource rules never apply to ordinary links, which can be citations.
	if tag == "img" || tag == "iframe" || tag == "script" || tag == "source" {
		source := s.AttrOr("data-src", s.AttrOr("data-original", s.AttrOr("src", "")))
		if source != "" && AdResource(source, base) {
			return "ad_resource"
		}
	}
	if tag == "img" && trackingPixel(s, base) {
		return "tracking_pixel"
	}
	if tag != "div" && tag != "aside" && tag != "section" && tag != "ins" && tag != "span" && tag != "figure" {
		return ""
	}
	classes := strings.Fields(strings.ToLower(s.AttrOr("class", "")))
	if has(classes, "adsbygoogle") && strings.HasPrefix(s.AttrOr("data-ad-client", ""), "ca-pub-") && s.AttrOr("data-ad-slot", "") != "" {
		return "ad_network"
	}
	id := strings.ToLower(s.AttrOr("id", ""))
	if strings.HasPrefix(id, "div-gpt-ad-") || strings.HasPrefix(id, "google_ads_iframe_") {
		return "ad_network"
	}
	// Ambiguous marketing blocks are handled by contextual AI review, not keywords.
	return ""
}

func has(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func trackingPixel(s *goquery.Selection, base string) bool {
	width, e1 := strconv.Atoi(s.AttrOr("width", ""))
	height, e2 := strconv.Atoi(s.AttrOr("height", ""))
	if e1 != nil || e2 != nil || width < 0 || height < 0 || width > 1 || height > 1 {
		return false
	}
	// Lazy image placeholders are not beacons; preserve their actual picture.
	for _, key := range []string{"data-src", "data-original", "data-lazy-src", "data-srcset"} {
		if s.AttrOr(key, "") != "" {
			return false
		}
	}
	if strings.TrimSpace(s.AttrOr("alt", "")) != "" {
		return false
	}
	source := s.AttrOr("src", "")
	if strings.HasPrefix(strings.ToLower(source), "data:") {
		return false
	}
	// Tiny images alone are ambiguous: require a tracking endpoint or a known ad host.
	u, err := url.Parse(source)
	if err != nil {
		return false
	}
	endpoint := strings.ToLower(path.Base(u.Path))
	return AdResource(source, base) || endpoint == "pixel.gif" || (u.RawQuery != "" && (endpoint == "pixel" || endpoint == "track" || endpoint == "beacon"))
}

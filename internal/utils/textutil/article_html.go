package textutil

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var markdownBlock = regexp.MustCompile("(?m)^ {0,3}(#{1,6} |```|~~~|[-*+] |[0-9]+\\. |>|\\|.*\\|)")
var htmlElement = regexp.MustCompile(`(?i)<[a-z][a-z0-9]*(?:\s|/?>)`)

// PrepareArticleContent renders Markdown sources and normalizes safe reader HTML.
// Existing HTML and pre/code examples remain HTML rather than being reinterpreted.
func PrepareArticleContent(content, baseURL string) string {
	if !htmlElement.MatchString(content) && markdownBlock.MatchString(content) {
		content = RenderMarkdown(content)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return ""
	}
	doc.Find("script,style,link,meta,base,object,embed,form,input,button,textarea,select,svg,math,template").Remove()
	base, _ := url.Parse(baseURL)
	doc.Find("*").Each(func(_ int, sel *goquery.Selection) {
		node := sel.Get(0)
		if node.Type != html.ElementNode {
			return
		}
		attrs := make([]html.Attribute, 0, len(node.Attr))
		for _, attr := range node.Attr {
			if attr.Namespace != "" {
				continue
			}
			key := strings.ToLower(attr.Key)
			switch key {
			case "href", "src", "poster":
				value, err := url.Parse(strings.TrimSpace(attr.Val))
				if err != nil {
					continue
				}
				if base != nil {
					value = base.ResolveReference(value)
				}
				if value.Scheme != "http" && value.Scheme != "https" && !(key == "href" && value.Scheme == "mailto") {
					continue
				}
				attr.Val = value.String()
			case "class":
				allowed := []string{}
				for _, c := range strings.Fields(attr.Val) {
					if strings.HasPrefix(c, "language-") || c == "math" || c == "math-display" {
						allowed = append(allowed, c)
					}
				}
				if len(allowed) == 0 {
					continue
				}
				attr.Val = strings.Join(allowed, " ")
			case "alt", "title", "width", "height", "colspan", "rowspan", "id", "controls", "preload", "type", "data-math":
			default:
				continue
			}
			attrs = append(attrs, attr)
		}
		node.Attr = attrs
		if node.Data == "iframe" {
			source, _ := url.Parse(sel.AttrOr("src", ""))
			if source == nil || !allowedEmbedHost(source.Hostname()) {
				sel.Remove()
				return
			}
			sel.SetAttr("sandbox", "allow-scripts allow-same-origin allow-presentation")
			sel.SetAttr("allowfullscreen", "")
		}
	})
	result, _ := doc.Find("body").Html()
	return strings.TrimSpace(result)
}

func allowedEmbedHost(host string) bool {
	for _, allowed := range []string{"youtube.com", "youtube-nocookie.com", "player.vimeo.com", "player.bilibili.com"} {
		if host == allowed || strings.HasSuffix(host, "."+allowed) {
			return true
		}
	}
	return false
}

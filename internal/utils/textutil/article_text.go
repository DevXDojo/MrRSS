package textutil

import (
	"github.com/PuerkitoBio/goquery"
	"strings"
)

// ArticlePlainText extracts readable evidence without script/style contents or
// markup attributes, preserving word boundaries and decoding HTML entities.
func ArticlePlainText(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return ""
	}
	doc.Find("script,style,template,noscript").Remove()
	doc.Find("p,div,li,h1,h2,h3,h4,br,tr,blockquote,pre").Each(func(_ int, s *goquery.Selection) { s.PrependHtml(" "); s.AppendHtml(" ") })
	return strings.Join(strings.Fields(doc.Find("body").Text()), " ")
}

package textutil

import (
	"MrRSS/internal/adfilter"
	"github.com/PuerkitoBio/goquery"
	"strings"
)

// PrepareFilteredArticle always sanitizes both variants. The original source
// stays in the existing cache; filtering is a reversible presentation operation.
func PrepareFilteredArticle(source, base string, options adfilter.Options) adfilter.Content {
	if !htmlElement.MatchString(source) && markdownBlock.MatchString(source) {
		source = RenderMarkdown(source)
	}
	original := PrepareArticleContent(source, base)
	result := adfilter.Content{Content: original, Filter: adfilter.Report{Enabled: options.Applies(base), Reasons: map[string]int{}}}
	if !result.Filter.Enabled {
		return result
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(source))
	if err != nil {
		return result
	}
	result.Filter = adfilter.Filter(doc, base, options)
	if result.Filter.Removed == 0 {
		return result
	}
	filtered, err := doc.Find("body").Html()
	if err != nil {
		return result
	}
	result.Content = PrepareArticleContent(filtered, base)
	result.OriginalContent = original
	return result
}

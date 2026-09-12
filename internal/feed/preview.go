package feed

import (
	"context"
	"net/url"
	"sort"
	"strings"
	"time"

	"MrRSS/internal/models"
	"MrRSS/internal/utils/textutil"
	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

type PreviewArticle struct {
	Title       string     `json:"title"`
	URL         string     `json:"url"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	ContentHTML string     `json:"content_html"`
	Truncated   bool       `json:"truncated"`
}

type Preview struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Articles    []PreviewArticle `json:"articles"`
	Total       int              `json:"total"`
}

// PreviewSubscription parses an unsaved feed. It never processes articles into
// the database, changes read state, or schedules a refresh.
func (f *Fetcher) PreviewSubscription(ctx context.Context, source *models.Feed) (*Preview, error) {
	parsed, err := f.ParseFeedWithFeed(ctx, source, false)
	if err != nil {
		return nil, err
	}
	return buildPreview(parsed, source.URL), nil
}

func previewPublished(item *gofeed.Item) *time.Time {
	if item.PublishedParsed != nil {
		return item.PublishedParsed
	}
	return item.UpdatedParsed
}

func buildPreview(parsed *gofeed.Feed, baseURL string) *Preview {
	items := append([]*gofeed.Item(nil), parsed.Items...)
	sort.SliceStable(items, func(i, j int) bool {
		left, right := previewPublished(items[i]), previewPublished(items[j])
		if left == nil {
			return false
		}
		return right == nil || left.After(*right)
	})
	preview := &Preview{Title: parsed.Title, Description: parsed.Description, Total: len(items), Articles: []PreviewArticle{}}
	if doc, err := goquery.NewDocumentFromReader(strings.NewReader(textutil.PrepareArticleContent(parsed.Description, baseURL))); err == nil {
		preview.Description = strings.TrimSpace(doc.Text())
	}
	if len(items) > 20 {
		items = items[:20]
	}
	for _, item := range items {
		content := ExtractContent(item)
		truncated := false
		// Bound preview size before parsing/sanitizing large article bodies.
		if len(content) > 512<<10 {
			content = strings.ToValidUTF8(content[:512<<10], "")
			truncated = true
		}
		articleURL := previewArticleURL(baseURL, item.Link)
		articleBase := articleURL
		if articleBase == "" {
			articleBase = baseURL
		}
		preview.Articles = append(preview.Articles, PreviewArticle{
			Title: item.Title, URL: articleURL, PublishedAt: previewPublished(item),
			ContentHTML: textutil.PrepareArticleContent(content, articleBase), Truncated: truncated,
		})
	}
	return preview
}

func previewArticleURL(baseURL, link string) string {
	if strings.TrimSpace(link) == "" {
		return ""
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}
	reference, err := base.Parse(link)
	if err != nil || (reference.Scheme != "https" && reference.Scheme != "http") {
		return ""
	}
	return reference.String()
}

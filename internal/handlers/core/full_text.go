package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"MrRSS/internal/database"
	"MrRSS/internal/models"
	"MrRSS/internal/utils/textutil"
	"codeberg.org/readeck/go-readability/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html/charset"
)

const maxArticleBytes = 10 << 20

func (h *Handler) FetchFullArticleContent(articleURL string) (string, error) {
	return h.FetchFullArticleContentWithFeed(articleURL, nil)
}

func (h *Handler) FetchFullArticleContentWithFeed(articleURL string, source *models.Feed) (string, error) {
	return h.FetchFullArticleContentContext(context.Background(), articleURL, source)
}

// FetchFullArticleContentContext fetches and extracts a page using the feed's network settings.
func (h *Handler) FetchFullArticleContentContext(ctx context.Context, articleURL string, source *models.Feed) (string, error) {
	parsedURL, err := url.Parse(articleURL)
	if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return "", fmt.Errorf("invalid article URL")
	}
	client, err := h.createArticleHTTPClient(source)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, articleURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9,*/*;q=0.8")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetch page: HTTP %d", resp.StatusCode)
	}
	reader, err := charset.NewReader(io.LimitReader(resp.Body, maxArticleBytes+1), resp.Header.Get("Content-Type"))
	if err != nil {
		return "", fmt.Errorf("decode page: %w", err)
	}
	body, err := io.ReadAll(io.LimitReader(reader, maxArticleBytes+1))
	if err != nil {
		return "", fmt.Errorf("read page: %w", err)
	}
	if len(body) > maxArticleBytes {
		return "", fmt.Errorf("article exceeds 10 MB")
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("parse page: %w", err)
	}
	base := resp.Request.URL
	if baseHref, exists := doc.Find("base[href]").First().Attr("href"); exists {
		if ref, err := base.Parse(baseHref); err == nil && (ref.Scheme == "http" || ref.Scheme == "https") {
			base = ref
		}
	}
	options := database.FeedContentOptions{}
	if source != nil {
		options, err = h.DB.GetFeedContentOptions(ctx, source.ID)
		if err != nil {
			return "", fmt.Errorf("read content options: %w", err)
		}
	}
	if err := options.Validate(); err != nil {
		return "", err
	}
	if strings.TrimSpace(options.RemoveSelector) != "" {
		matcher, _ := cascadia.Compile(options.RemoveSelector)
		doc.FindMatcher(matcher).Remove()
	}
	normalizeArticleImages(doc, base)
	if strings.TrimSpace(options.ContentSelector) != "" {
		matcher, _ := cascadia.Compile(options.ContentSelector)
		selected := doc.FindMatcher(matcher)
		if selected.Length() == 0 {
			return "", fmt.Errorf("content selector matched no elements")
		}
		var fragments strings.Builder
		selected.Each(func(_ int, selection *goquery.Selection) {
			// Nested matches are already included by their matching ancestor.
			if selection.ParentsMatcher(matcher).Length() > 0 {
				return
			}
			fragment, _ := goquery.OuterHtml(selection)
			fragments.WriteString(fragment)
		})
		content := textutil.PrepareArticleContent(fragments.String(), base.String())
		if content == "" {
			return "", fmt.Errorf("content selector produced empty content")
		}
		return content, nil
	}

	page, err := doc.Html()
	if err != nil {
		return "", fmt.Errorf("render page: %w", err)
	}
	extracted, extractErr := readability.FromReader(strings.NewReader(page), base)
	var output bytes.Buffer
	if extractErr == nil {
		extractErr = extracted.RenderHTML(&output)
	}
	content := output.String()
	if extractErr != nil || strings.TrimSpace(content) == "" {
		// Explicit semantic article containers are a useful fallback for short pages.
		content, _ = doc.Find("article,main,[role=main]").First().Html()
	}
	content = textutil.PrepareArticleContent(content, base.String())
	if strings.TrimSpace(content) == "" {
		return "", fmt.Errorf("no readable article content")
	}
	return content, nil
}

func normalizeArticleImages(doc *goquery.Document, base *url.URL) {
	doc.Find("img").Each(func(_ int, img *goquery.Selection) {
		source := ""
		for _, attr := range []string{"data-src", "data-original", "data-lazy-src", "data-actualsrc", "data-original-src", "src"} {
			value := strings.TrimSpace(img.AttrOr(attr, ""))
			if value != "" && !strings.HasPrefix(value, "data:") {
				source = value
				break
			}
		}
		if source == "" {
			srcset := img.AttrOr("data-srcset", img.AttrOr("srcset", ""))
			candidates := strings.Split(srcset, ",")
			if len(candidates) > 0 {
				parts := strings.Fields(candidates[len(candidates)-1])
				if len(parts) > 0 {
					source = parts[0]
				}
			}
		}
		if resolved, err := base.Parse(source); source != "" && err == nil && (resolved.Scheme == "http" || resolved.Scheme == "https") {
			img.SetAttr("src", resolved.String())
			img.RemoveAttr("srcset").RemoveAttr("sizes").RemoveAttr("loading")
		}
	})
}

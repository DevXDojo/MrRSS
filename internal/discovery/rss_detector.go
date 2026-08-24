package discovery

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// discoverRSSFeedsWithProgress discovers RSS feeds with progress updates
func (s *Service) discoverRSSFeedsWithProgress(ctx context.Context, blogURLs []string, progressCb ProgressCallback) []DiscoveredBlog {
	var wg sync.WaitGroup
	results := make(chan DiscoveredBlog, len(blogURLs))
	sem := make(chan struct{}, MaxConcurrentRSSChecks)

	// Track progress
	var progressMu sync.Mutex
	processed := 0
	foundCount := 0
	total := len(blogURLs)

OuterLoop:
	for _, blogURL := range blogURLs {
		select {
		case <-ctx.Done():
			break OuterLoop
		default:
		}

		wg.Add(1)
		sem <- struct{}{}

		go func(u string) {
			defer wg.Done()
			defer func() { <-sem }()

			// Report progress
			if progressCb != nil {
				progressMu.Lock()
				processed++
				currentProcessed := processed
				currentFound := foundCount
				progressMu.Unlock()

				progressCb(Progress{
					Stage:      "checking_rss",
					Message:    "Checking RSS feed",
					Detail:     u,
					Current:    currentProcessed,
					Total:      total,
					FoundCount: currentFound,
				})
			}

			if blog, err := s.discoverBlogRSS(ctx, u); err == nil {
				progressMu.Lock()
				foundCount++
				progressMu.Unlock()
				results <- blog
			}
		}(blogURL)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var discovered []DiscoveredBlog
	for blog := range results {
		discovered = append(discovered, blog)
	}

	return discovered
}

// discoverBlogRSS discovers RSS feed for a single blog
func (s *Service) discoverBlogRSS(ctx context.Context, blogURL string) (DiscoveredBlog, error) {
	// Try to find RSS feed URL
	rssURL, err := s.findRSSFeed(ctx, blogURL)
	if err != nil {
		return DiscoveredBlog{}, err
	}

	// Parse the RSS feed to get blog info
	feed, err := s.feedParser.ParseURLWithContext(rssURL, ctx)
	if err != nil {
		return DiscoveredBlog{}, err
	}

	// Extract recent articles (max 3)
	var recentArticles []RecentArticle
	for i := 0; i < len(feed.Items) && i < 3; i++ {
		item := feed.Items[i]
		dateStr := ""
		if item.PublishedParsed != nil {
			// Format as relative time or date
			dateStr = item.PublishedParsed.Format("2006-01-02")
		}
		recentArticles = append(recentArticles, RecentArticle{
			Title: item.Title,
			Date:  dateStr,
		})
	}

	// Get favicon
	iconURL := s.getFavicon(blogURL)

	return DiscoveredBlog{
		Name:           feed.Title,
		Homepage:       blogURL,
		RSSFeed:        rssURL,
		IconURL:        iconURL,
		RecentArticles: recentArticles,
	}, nil
}

// findRSSFeed finds the RSS feed URL for a blog
func (s *Service) findRSSFeed(ctx context.Context, blogURL string) (string, error) {
	// Common RSS feed paths to try
	u, err := url.Parse(blogURL)
	if err != nil {
		return "", err
	}

	baseURL := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	// First, try to parse HTML and find RSS link in <head> - this is usually the most reliable
	doc, err := s.fetchHTML(ctx, blogURL)
	if err == nil {
		var foundFeed string
		doc.Find("link[type='application/rss+xml'], link[type='application/atom+xml'], link[rel='alternate'][type*='xml']").Each(func(i int, sel *goquery.Selection) {
			if foundFeed != "" {
				return
			}
			if href, exists := sel.Attr("href"); exists {
				foundFeed = s.resolveURL(blogURL, href)
			}
		})

		if foundFeed != "" && s.isValidFeed(ctx, foundFeed) {
			return foundFeed, nil
		}
	}

	// Expanded common RSS/Atom feed paths
	commonPaths := []string{
		"/rss.xml",
		"/feed.xml",
		"/atom.xml",
		"/feed",
		"/rss",
		"/feeds/posts/default", // Blogger
		"/index.xml",           // Hugo
		"/feed/",
		"/rss/",
		"/atom/",
		"/blog/feed",
		"/blog/rss",
		"/blog/feed.xml",
		"/blog/rss.xml",
		"/posts/feed",
		"/posts/rss.xml",
		"/?feed=rss2",     // WordPress
		"/feed/?type=rss", // Some WordPress
		"/rss2.xml",
		"/feed.atom",
		"/feed.rss",
	}

	// Try common paths concurrently for faster discovery
	type feedResult struct {
		url   string
		valid bool
	}
	resultCh := make(chan feedResult, len(commonPaths))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, MaxConcurrentPathChecks)

	for _, path := range commonPaths {
		feedURL := baseURL + path
		wg.Add(1)
		go func(fURL string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if s.isValidFeed(ctx, fURL) {
				resultCh <- feedResult{url: fURL, valid: true}
			}
		}(feedURL)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Return the first valid feed found
	for result := range resultCh {
		if result.valid {
			return result.url, nil
		}
	}

	return "", errRSSFeedNotFound
}

// isValidFeed checks if a URL is a valid RSS/Atom feed
func (s *Service) isValidFeed(ctx context.Context, feedURL string) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", feedURL, nil)
	if err != nil {
		return false
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Try GET if HEAD doesn't work
		req, err = http.NewRequestWithContext(ctx, "GET", feedURL, nil)
		if err != nil {
			return false
		}

		resp2, err := s.client.Do(req)
		if err != nil {
			return false
		}
		defer resp2.Body.Close()

		if resp2.StatusCode != http.StatusOK {
			return false
		}

		// Read first few bytes to check if it's XML
		buf := make([]byte, 512)
		n, err := io.ReadAtLeast(resp2.Body, buf, 1)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return false
		}
		if n == 0 {
			return false
		}
		content := string(buf[:n])

		// Check for XML declaration and RSS/Atom tags
		if strings.Contains(content, "<?xml") ||
			strings.Contains(content, "<rss") ||
			strings.Contains(content, "<feed") ||
			strings.Contains(content, "<atom") {
			return true
		}
		return false
	}

	contentType := resp.Header.Get("Content-Type")
	return strings.Contains(contentType, "xml") ||
		strings.Contains(contentType, "rss") ||
		strings.Contains(contentType, "atom")
}

// getFavicon gets the favicon URL for a blog
func (s *Service) getFavicon(blogURL string) string {
	u, err := url.Parse(blogURL)
	if err != nil {
		return ""
	}

	// Use Google's favicon service as fallback
	return fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s", u.Host)
}

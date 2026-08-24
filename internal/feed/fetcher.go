package feed

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/models"
	"MrRSS/internal/rsshub"
	"MrRSS/internal/rules"
	"MrRSS/internal/utils"
	"MrRSS/internal/utils/fileutil"
	"MrRSS/internal/utils/httputil"

	"github.com/mmcdole/gofeed"
)

// FeedParser interface to allow mocking
type FeedParser interface {
	ParseURL(url string) (*gofeed.Feed, error)
	ParseURLWithContext(url string, ctx context.Context) (*gofeed.Feed, error)
}

type Fetcher struct {
	db                *database.DB
	fp                FeedParser
	highPriorityFp    FeedParser // High priority parser for content fetching
	scriptExecutor    *ScriptExecutor
	emailFetcher      *EmailFetcher
	progress          Progress
	mu                sync.Mutex
	refreshCalculator *IntelligentRefreshCalculator
	taskManager       *TaskManager
	cleanupManager    *CleanupManager
}

func NewFetcher(db *database.DB) *Fetcher {
	// Initialize script executor with scripts directory
	scriptsDir, err := fileutil.GetScriptsDir()
	var executor *ScriptExecutor
	if err == nil {
		executor = NewScriptExecutor(scriptsDir)
	}

	// Create HTTP client for feed parsing with proper User-Agent
	// This is critical because many RSS servers block requests without a proper User-Agent
	httpClient, err := httputil.CreateHTTPClientWithUserAgent(
		"",
		30*time.Second,
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	)
	if err != nil {
		// Fallback to default client if proxy setup fails
		log.Printf("Warning: Failed to create HTTP client with User-Agent: %v, using default client", err)
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	// Create parser with custom HTTP client to support localhost and other endpoints
	parser := gofeed.NewParser()
	parser.Client = httpClient

	// Create high priority parser with shorter timeout for content fetching
	highPriorityParser := gofeed.NewParser()
	highPriorityParser.Client = httpClient

	fetcher := &Fetcher{
		db:                db,
		fp:                parser,
		highPriorityFp:    highPriorityParser,
		scriptExecutor:    executor,
		emailFetcher:      NewEmailFetcher(db),
		refreshCalculator: NewIntelligentRefreshCalculator(db),
	}

	// Initialize task manager with default capacity (increased from 5 to 10)
	fetcher.taskManager = NewTaskManager(fetcher, 10)
	fetcher.taskManager.Start()

	// Initialize cleanup manager
	fetcher.cleanupManager = NewCleanupManager(fetcher)
	fetcher.cleanupManager.Start()

	return fetcher
}

// GetIntelligentRefreshCalculator returns the refresh calculator
func (f *Fetcher) GetIntelligentRefreshCalculator() *IntelligentRefreshCalculator {
	return f.refreshCalculator
}

// GetStaggeredDelay calculates a staggered delay for feed refresh
func (f *Fetcher) GetStaggeredDelay(feedID int64, totalFeeds int) time.Duration {
	return GetStaggeredDelay(feedID, totalFeeds)
}

// GetTaskManager returns the task manager
func (f *Fetcher) GetTaskManager() *TaskManager {
	return f.taskManager
}

// GetCleanupManager returns the cleanup manager
func (f *Fetcher) GetCleanupManager() *CleanupManager {
	return f.cleanupManager
}

// transformRSSHubURL converts rsshub:// route to full URL
func (f *Fetcher) transformRSSHubURL(url string) (string, error) {
	if !rsshub.IsRSSHubURL(url) {
		return url, nil
	}

	// Check if RSSHub is enabled
	enabledStr, _ := f.db.GetSetting("rsshub_enabled")
	if enabledStr != "true" {
		return "", fmt.Errorf("RSSHub integration is disabled. Please enable it in settings")
	}

	endpoint, _ := f.db.GetSetting("rsshub_endpoint")
	if endpoint == "" {
		endpoint = "https://rsshub.app"
	}
	apiKey, _ := f.db.GetEncryptedSetting("rsshub_api_key")

	route := rsshub.ExtractRoute(url)
	client := rsshub.NewClient(endpoint, apiKey)
	return client.BuildURL(route), nil
}

// getDataDir returns the data directory path
func (f *Fetcher) getDataDir() (string, error) {
	return fileutil.GetDataDir()
}

// getConcurrencyLimit returns the maximum number of concurrent feed refreshes
// based on network detection or defaults to 10 if not configured
func (f *Fetcher) getConcurrencyLimit() int {
	concurrencyStr, err := f.db.GetSetting("max_concurrent_refreshes")
	if err != nil || concurrencyStr == "" {
		return 10 // Default concurrency increased from 5 to 10
	}

	concurrency, err := strconv.Atoi(concurrencyStr)
	if err != nil || concurrency < 1 {
		return 10 // Default on parse error or invalid value
	}

	// Cap at reasonable limits (increased to 50 for faster networks)
	if concurrency > 50 {
		concurrency = 50
	}

	return concurrency
}

// getHTTPClient returns an HTTP client configured with proxy if needed
// Proxy precedence (highest to lowest):
// 1. Feed custom proxy (ProxyEnabled=true, ProxyURL != "")
// 2. Global proxy (ProxyEnabled=true, ProxyURL == "", global proxy_enabled=true)
// 3. No proxy (ProxyEnabled=false or no global proxy)
func (f *Fetcher) getHTTPClient(feed models.Feed) (*http.Client, error) {
	var proxyURL string

	// Check feed-level proxy settings
	if feed.ProxyEnabled && feed.ProxyURL != "" {
		// Feed has custom proxy configured - highest priority
		proxyURL = feed.ProxyURL
	} else if feed.ProxyEnabled {
		// Feed requests to use global proxy
		proxyEnabled, _ := f.db.GetSetting("proxy_enabled")
		if proxyEnabled == "true" {
			// Build global proxy URL from settings (use encrypted methods for credentials)
			proxyType, _ := f.db.GetSetting("proxy_type")
			proxyHost, _ := f.db.GetSetting("proxy_host")
			proxyPort, _ := f.db.GetSetting("proxy_port")
			proxyUsername, _ := f.db.GetEncryptedSetting("proxy_username")
			proxyPassword, _ := f.db.GetEncryptedSetting("proxy_password")
			proxyURL = BuildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)
		}
	}
	// If ProxyEnabled=false, proxyURL remains empty (no proxy)

	// Create HTTP client with browser-like headers to bypass Cloudflare and anti-bot protections
	// This is critical for RSSHub feeds and other services with anti-bot protection
	return httputil.CreateHTTPClientWithUserAgent(
		proxyURL,
		30*time.Second,
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	)
}

func (f *Fetcher) FetchAll(ctx context.Context) {
	// Get all feeds
	feeds, err := f.db.GetFeeds()
	if err != nil {
		log.Println("Error getting feeds:", err)
		return
	}

	if len(feeds) == 0 {
		log.Println("No feeds to refresh")
		// Mark progress as completed since there's nothing to do
		f.taskManager.MarkCompleted()
		return
	}

	// Filter out FreshRSS feeds, never-refresh feeds, and custom interval feeds
	filteredFeeds := make([]models.Feed, 0, len(feeds))
	freshRSSCount := 0
	neverRefreshCount := 0
	customIntervalCount := 0
	for _, feed := range feeds {
		if feed.IsFreshRSSSource {
			freshRSSCount++
		} else if feed.RefreshInterval == -2 {
			// Skip feeds with never refresh mode
			neverRefreshCount++
		} else if feed.RefreshInterval > 0 {
			// Skip feeds with custom refresh interval (they are scheduled individually)
			customIntervalCount++
		} else {
			// Only include feeds using global setting (RefreshInterval == 0)
			filteredFeeds = append(filteredFeeds, feed)
		}
	}

	// If all feeds are FreshRSS feeds, never-refresh feeds, or custom interval feeds, no standard refresh needed
	if len(filteredFeeds) == 0 {
		if freshRSSCount > 0 && neverRefreshCount > 0 && customIntervalCount > 0 {
			log.Printf("All feeds are either FreshRSS sources (%d), never-refresh feeds (%d), or custom interval feeds (%d), skipping standard refresh", freshRSSCount, neverRefreshCount, customIntervalCount)
		} else if freshRSSCount > 0 && neverRefreshCount > 0 {
			log.Printf("All feeds are either FreshRSS sources (%d) or never-refresh feeds (%d), skipping standard refresh", freshRSSCount, neverRefreshCount)
		} else if freshRSSCount > 0 && customIntervalCount > 0 {
			log.Printf("All feeds are either FreshRSS sources (%d) or custom interval feeds (%d), skipping standard refresh", freshRSSCount, customIntervalCount)
		} else if neverRefreshCount > 0 && customIntervalCount > 0 {
			log.Printf("All feeds are either never-refresh feeds (%d) or custom interval feeds (%d), skipping standard refresh", neverRefreshCount, customIntervalCount)
		} else if freshRSSCount > 0 {
			log.Printf("All %d feeds are FreshRSS sources (refreshed via sync only), skipping standard refresh", freshRSSCount)
		} else if neverRefreshCount > 0 {
			log.Printf("All %d feeds are never-refresh feeds, skipping standard refresh", neverRefreshCount)
		} else {
			log.Printf("All %d feeds are custom interval feeds, skipping standard refresh", customIntervalCount)
		}

		// Update last global refresh time even if no standard feeds
		newUpdateTime := time.Now().Format(time.RFC3339)
		log.Printf("Global refresh started (FreshRSS only), updating last_global_refresh to: %s", newUpdateTime)
		if err := f.db.SetSetting("last_global_refresh", newUpdateTime); err != nil {
			log.Printf("ERROR: Failed to update last_global_refresh: %v", err)
		}

		// Track global refresh in statistics even if no standard feeds
		if err := f.db.IncrementStat("feed_refresh"); err != nil {
			log.Printf("ERROR: Failed to track feed refresh: %v", err)
		}

		// Mark progress as completed since there's nothing to do
		f.taskManager.MarkCompleted()
		return
	}

	// Log what's being refreshed
	if neverRefreshCount > 0 && customIntervalCount > 0 {
		log.Printf("Standard refresh: %d feeds (skipped %d FreshRSS feeds, %d never-refresh feeds, %d custom interval feeds)", len(filteredFeeds), freshRSSCount, neverRefreshCount, customIntervalCount)
	} else if neverRefreshCount > 0 {
		log.Printf("Standard refresh: %d feeds (skipped %d FreshRSS feeds, %d never-refresh feeds)", len(filteredFeeds), freshRSSCount, neverRefreshCount)
	} else if customIntervalCount > 0 {
		log.Printf("Standard refresh: %d feeds (skipped %d FreshRSS feeds, %d custom interval feeds)", len(filteredFeeds), freshRSSCount, customIntervalCount)
	} else {
		log.Printf("Standard refresh: %d feeds (skipped %d FreshRSS feeds)", len(filteredFeeds), freshRSSCount)
	}

	// Update task manager capacity based on network
	concurrency := f.getConcurrencyLimit()
	f.taskManager.SetPoolCapacity(concurrency)

	// Use task manager for global refresh (all feeds go to queue tail)
	f.taskManager.AddGlobalRefresh(ctx, filteredFeeds)
}

func (f *Fetcher) FetchFeed(ctx context.Context, feed models.Feed) {
	// Use ParseFeedWithFeed with normal priority for feed refresh
	parsedFeed, err := f.ParseFeedWithFeed(ctx, &feed, false) // Normal priority for refresh
	if err != nil {
		log.Printf("Error parsing feed %s: %v", feed.URL, err)
		f.db.UpdateFeedError(feed.ID, err.Error())
		// Add error to progress for immediate feedback
		f.mu.Lock()
		if f.progress.Errors == nil {
			f.progress.Errors = make(map[int64]string)
		}
		f.progress.Errors[feed.ID] = err.Error()
		f.mu.Unlock()
		return
	}

	// Clear any previous error on successful fetch
	f.db.UpdateFeedError(feed.ID, "")

	// Update Feed Image if available and not set
	if feed.ImageURL == "" && parsedFeed.Image != nil {
		f.db.UpdateFeedImage(feed.ID, parsedFeed.Image.URL)
	}

	// Update Feed Link if available and not set
	if feed.Link == "" && parsedFeed.Link != "" {
		f.db.UpdateFeedLink(feed.ID, parsedFeed.Link)
	}

	// Process articles
	articlesWithContent := f.processArticles(feed, parsedFeed.Items)

	// Check context before heavy DB operation
	select {
	case <-ctx.Done():
		return
	default:
	}

	if len(articlesWithContent) > 0 {
		// Extract just the articles for saving
		articlesToSave := make([]*models.Article, len(articlesWithContent))
		for i, awc := range articlesWithContent {
			articlesToSave[i] = awc.Article
		}

		if err := f.db.SaveArticles(ctx, articlesToSave); err != nil {
			log.Printf("Error saving articles for feed %s: %v", feed.Title, err)
		} else {
			// Cache article content from RSS feed
			f.cacheArticleContents(articlesWithContent)

			// Apply rules to newly saved articles
			// We fetch the recent articles for this feed since SaveArticles doesn't return IDs
			// This is limited to the number of articles we just saved
			savedArticles, err := f.db.GetArticles("", feed.ID, "", false, len(articlesToSave), 0)
			if err == nil && len(savedArticles) > 0 {
				engine := rules.NewEngine(f.db)
				affected, err := engine.ApplyRulesToArticles(savedArticles)
				if err != nil {
					log.Printf("Error applying rules for feed %s: %v", feed.Title, err)
				} else if affected > 0 {
					utils.DebugLog("Applied rules to %d articles in feed %s", affected, feed.Title)
				}
			}
		}
	}
	utils.DebugLog("Updated feed: %s", feed.Title)
}

// fetchFeedWithContext is the internal fetch method used by TaskManager
// Returns error instead of storing in progress.Errors
func (f *Fetcher) fetchFeedWithContext(ctx context.Context, feed models.Feed) error {
	// Use ParseFeedWithFeed with normal priority for feed refresh
	parsedFeed, err := f.ParseFeedWithFeed(ctx, &feed, false)
	if err != nil {
		return err
	}

	// Check context after parsing
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Clear any previous error on successful fetch
	f.db.UpdateFeedError(feed.ID, "")

	// Update Feed Image if available and not set
	if feed.ImageURL == "" && parsedFeed.Image != nil {
		f.db.UpdateFeedImage(feed.ID, parsedFeed.Image.URL)
	}

	// Update Feed Link if available and not set
	if feed.Link == "" && parsedFeed.Link != "" {
		f.db.UpdateFeedLink(feed.ID, parsedFeed.Link)
	}

	// Check context before processing articles
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Process articles
	articlesWithContent := f.processArticles(feed, parsedFeed.Items)

	// Check context before heavy DB operation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if len(articlesWithContent) > 0 {
		// Extract just the articles for saving
		articlesToSave := make([]*models.Article, len(articlesWithContent))
		for i, awc := range articlesWithContent {
			articlesToSave[i] = awc.Article
		}

		if err := f.db.SaveArticles(ctx, articlesToSave); err != nil {
			return err
		}

		// Post-processing operations (content caching and rule application)
		// These are non-critical and run asynchronously to avoid blocking the feed refresh
		// Even if they fail or are slow, the feed has already been successfully saved
		go func() {
			// Cache article content from RSS feed
			f.cacheArticleContents(articlesWithContent)

			// Apply rules to newly saved articles
			savedArticles, err := f.db.GetArticles("", feed.ID, "", false, len(articlesToSave), 0)
			if err != nil {
				log.Printf("Error getting articles for rule application: %v", err)
				return
			}
			if len(savedArticles) == 0 {
				return
			}

			engine := rules.NewEngine(f.db)
			affected, err := engine.ApplyRulesToArticles(savedArticles)
			if err != nil {
				log.Printf("Error applying rules for feed %s: %v", feed.Title, err)
			} else if affected > 0 {
				utils.DebugLog("Applied rules to %d articles in feed %s", affected, feed.Title)
			}
		}()
	}
	return nil
}

// FetchSingleFeed fetches a single feed with progress tracking.
// This is used when adding a new feed, refreshing a single feed from the context menu,
// or when the scheduler triggers individual feed refreshes.
// For manual operations (add/edit/refresh), place at queue head.
// For scheduled operations, place at queue tail.
func (f *Fetcher) FetchSingleFeed(ctx context.Context, feed models.Feed, isManual bool) {
	if isManual {
		// Manual operations go to queue head
		f.taskManager.AddToQueueHead(ctx, feed, TaskReasonManualRefresh)
	} else {
		// Scheduled operations go to queue tail
		f.taskManager.AddToQueueTail(ctx, feed, TaskReasonScheduledCustom)
	}
}

// FetchFeedForArticle fetches a feed immediately when article content is missing.
// This bypasses the queue and pool limits.
func (f *Fetcher) FetchFeedForArticle(ctx context.Context, feed models.Feed) {
	f.taskManager.ExecuteImmediately(ctx, feed)
}

// FetchFeedsByIDs fetches multiple feeds by their IDs with progress tracking.
// This is used after OPML import or when editing feeds.
// All feeds are added to queue head (high priority).
func (f *Fetcher) FetchFeedsByIDs(ctx context.Context, feedIDs []int64) {
	if len(feedIDs) == 0 {
		return
	}

	// Fetch feeds by IDs
	for _, feedID := range feedIDs {
		feed, err := f.db.GetFeedByID(feedID)
		if err != nil {
			log.Printf("Error getting feed %d: %v", feedID, err)
			continue
		}
		// Add to queue head as high priority (manual add/edit)
		f.taskManager.AddToQueueHead(ctx, *feed, TaskReasonManualAdd)
	}
}

// cacheArticleContents caches article contents from RSS feeds
// This is called after articles are saved to the database
func (f *Fetcher) cacheArticleContents(articlesWithContent []*ArticleWithContent) {
	for _, awc := range articlesWithContent {
		// Only cache if content is not empty and URL is present
		if awc.Content == "" || awc.Article.URL == "" {
			continue
		}

		// Get article ID by unique_id (article was just saved, so it should exist)
		articleID, err := f.db.GetArticleIDByUniqueID(awc.Article.Title, awc.Article.FeedID, awc.Article.PublishedAt, awc.Article.HasValidPublishedTime)
		if err != nil {
			// Article might not exist yet (race condition) or other error
			utils.DebugLog("Could not find article ID for %s: %v", awc.Article.Title, err)
			continue
		}

		// Cache the content (this will overwrite any existing cache as required)
		if err := f.db.SetArticleContent(articleID, awc.Content); err != nil {
			log.Printf("Error caching content for article %d: %v", articleID, err)
		} else {
			utils.DebugLog("Cached content for article %d", articleID)
		}
	}
}

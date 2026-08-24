// Package core contains the main Handler struct and core HTTP handlers for the application.
// It defines the Handler struct which holds dependencies like the database and fetcher.
package core

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"MrRSS/internal/ai"
	"MrRSS/internal/cache"
	"MrRSS/internal/database"
	"MrRSS/internal/discovery"
	"MrRSS/internal/feed"
	"MrRSS/internal/models"
	svc "MrRSS/internal/service"
	"MrRSS/internal/statistics"
	"MrRSS/internal/translation"
	"MrRSS/internal/utils/httputil"
	"MrRSS/internal/utils/textutil"
	"MrRSS/internal/utils/urlutil"

	"codeberg.org/readeck/go-readability/v2"

	"github.com/mmcdole/gofeed"
)

// Discovery timeout constants
const (
	// SingleFeedDiscoveryTimeout is the timeout for discovering feeds from a single source
	SingleFeedDiscoveryTimeout = 90 * time.Second
	// BatchDiscoveryTimeout is the timeout for discovering feeds from all sources
	BatchDiscoveryTimeout = 5 * time.Minute
)

// DiscoveryState represents the current state of a discovery operation
type DiscoveryState struct {
	IsRunning  bool                       `json:"is_running"`
	Progress   discovery.Progress         `json:"progress"`
	Feeds      []discovery.DiscoveredBlog `json:"feeds,omitempty"`
	Error      string                     `json:"error,omitempty"`
	IsComplete bool                       `json:"is_complete"`
}

// Handler holds all dependencies for HTTP handlers.
// It now uses a service registry for better separation of concerns.
type Handler struct {
	// Services registry provides access to all business logic services
	Services *svc.Registry

	// Direct access to core dependencies (for backward compatibility)
	DB                *database.DB
	Fetcher           *feed.Fetcher
	Translator        translation.Translator
	AIProfileProvider *ai.ProfileProvider // AI profile provider for feature-specific configurations
	AITracker         *ai.UsageTracker
	DiscoveryService  *discovery.Service
	App               interface{}         // Wails app instance for browser integration (interface{} to avoid import in server mode)
	ContentCache      *cache.ContentCache // Cache for article content
	Stats             *statistics.Service // Statistics tracking service

	// Discovery state tracking for polling-based progress
	DiscoveryMu          sync.RWMutex
	SingleDiscoveryState *DiscoveryState
	BatchDiscoveryState  *DiscoveryState
}

// NewHandler creates a new Handler with the given dependencies.
func NewHandler(db *database.DB, fetcher *feed.Fetcher, translator translation.Translator, profileProvider *ai.ProfileProvider) *Handler {
	// Create service registry
	registry := svc.NewRegistry(db, fetcher, translator)

	h := &Handler{
		Services:          registry,
		DB:                db,
		Fetcher:           fetcher,
		Translator:        translator,
		AIProfileProvider: profileProvider,
		AITracker:         registry.AITracker(),
		DiscoveryService:  registry.DiscoveryService(),
		ContentCache:      registry.ContentCache(),
		Stats:             registry.Stats(),
	}

	return h
}

// CallAppMethod calls a method on the Wails app instance if available
func (h *Handler) CallAppMethod(method string, args ...interface{}) error {
	if h.App == nil {
		return fmt.Errorf("app instance not set")
	}

	// Use reflection or type assertion to call the method
	// This is a simplified version - you may need to adjust based on your actual Wails app structure
	// For now, just log that we want to call this method
	log.Printf("Would call app method: %s with args: %v", method, args)
	return nil
}

// SetApp sets the Wails application instance for browser integration.
// This is called after app initialization in main.go.
func (h *Handler) SetApp(app interface{}) {
	h.App = app
}

// Statistics returns the statistics service
func (h *Handler) Statistics() *statistics.Service {
	return h.Stats
}

// GetArticleContent fetches article content with caching
// Returns (content, wasCached, error)
func (h *Handler) GetArticleContent(articleID int64) (string, bool, error) {
	// First, check database cache (persistent cache)
	content, found, err := h.DB.GetArticleContent(articleID)
	if err == nil && found {
		// Also populate memory cache for faster subsequent access
		h.ContentCache.Set(articleID, content)
		return content, true, nil
	}

	// Check memory cache (in-memory cache, might be stale but fast)
	if content, found := h.ContentCache.Get(articleID); found {
		return content, true, nil
	}

	// Get the article from database
	article, err := h.DB.GetArticleByID(articleID)
	if err != nil {
		return "", false, err
	}

	// Get the feed
	targetFeed, err := h.DB.GetFeedByID(article.FeedID)
	if err != nil {
		return "", false, err
	}

	if targetFeed == nil {
		return "", false, nil
	}

	// Trigger immediate feed refresh using the new task manager
	// This bypasses the queue and pool limits
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Fetch the feed immediately (article click triggered)
	h.Fetcher.FetchFeedForArticle(ctx, *targetFeed)

	// Parse the feed to get fresh content
	parsedFeed, err := h.Fetcher.ParseFeedWithFeed(ctx, targetFeed, true) // High priority for content fetching
	if err != nil {
		return "", false, err
	}

	// Cache the feed for future use
	h.ContentCache.SetFeed(targetFeed.ID, parsedFeed)

	// Find the article in the feed by multiple criteria for better matching
	matchingItem := h.findMatchingFeedItem(article, parsedFeed.Items)
	if matchingItem != nil {
		content := feed.ExtractContent(matchingItem)
		cleanContent := textutil.CleanHTML(content)

		// Cache the content in both memory and database
		h.ContentCache.Set(articleID, cleanContent)
		if err := h.DB.SetArticleContent(articleID, cleanContent); err != nil {
			log.Printf("Error caching content to database: %v", err)
		}

		return cleanContent, false, nil
	}

	return "", false, nil
}

// FetchFullArticleContent fetches the full article content from the original URL using readability.
func (h *Handler) FetchFullArticleContent(articleURL string) (string, error) {
	return h.FetchFullArticleContentWithFeed(articleURL, nil)
}

// FetchFullArticleContentWithFeed fetches full content using the same proxy semantics as feed refresh.
func (h *Handler) FetchFullArticleContentWithFeed(articleURL string, feedConfig *models.Feed) (string, error) {
	parsedURL, err := url.ParseRequestURI(articleURL)
	if err != nil {
		return "", fmt.Errorf("parse article URL: %w", err)
	}

	client, err := h.createArticleHTTPClient(feedConfig)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodGet, articleURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("fetch page: HTTP %d", resp.StatusCode)
	}

	article, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		return "", fmt.Errorf("readability parse: %w", err)
	}

	// Render the article content as HTML
	var buf bytes.Buffer
	err = article.RenderHTML(&buf)
	if err != nil {
		return "", fmt.Errorf("render HTML: %w", err)
	}

	return buf.String(), nil
}

func (h *Handler) createArticleHTTPClient(feedConfig *models.Feed) (*http.Client, error) {
	var proxyURL string
	if feedConfig != nil && feedConfig.ProxyEnabled && feedConfig.ProxyURL != "" {
		proxyURL = feedConfig.ProxyURL
	} else if feedConfig != nil && feedConfig.ProxyEnabled {
		proxyEnabled, _ := h.DB.GetSetting("proxy_enabled")
		if proxyEnabled == "true" {
			proxyURL = h.globalProxyURL()
		}
	} else if feedConfig == nil {
		proxyEnabled, _ := h.DB.GetSetting("proxy_enabled")
		if proxyEnabled == "true" {
			proxyURL = h.globalProxyURL()
		}
	}

	return httputil.CreateHTTPClient(proxyURL, 30*time.Second)
}

func (h *Handler) globalProxyURL() string {
	proxyType, _ := h.DB.GetSetting("proxy_type")
	proxyHost, _ := h.DB.GetSetting("proxy_host")
	proxyPort, _ := h.DB.GetSetting("proxy_port")
	proxyUsername, _ := h.DB.GetEncryptedSetting("proxy_username")
	proxyPassword, _ := h.DB.GetEncryptedSetting("proxy_password")
	return httputil.BuildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)
}

// findMatchingFeedItem finds the best matching feed item for an article using multiple criteria
func (h *Handler) findMatchingFeedItem(article *models.Article, items []*gofeed.Item) *gofeed.Item {
	// First pass: URL + title match. Empty URLs are not valid identifiers and
	// must never match each other, otherwise the first item in the feed wins.
	for _, item := range items {
		if nonEmptyURLsMatch(item.Link, article.URL) && h.titlesMatch(item.Title, article.Title) {
			return item
		}
	}

	// Second pass: title + published time match (fallback for missing or changed URLs)
	for _, item := range items {
		if h.titlesMatch(item.Title, article.Title) && h.publishedTimesMatch(item.PublishedParsed, &article.PublishedAt) {
			return item
		}
	}

	// Third pass: URL-only match for feeds whose titles change between refreshes
	for _, item := range items {
		if nonEmptyURLsMatch(item.Link, article.URL) {
			return item
		}
	}

	// Final fallback: just title match
	for _, item := range items {
		if h.titlesMatch(item.Title, article.Title) {
			return item
		}
	}

	return nil
}

func nonEmptyURLsMatch(url1, url2 string) bool {
	url1 = strings.TrimSpace(url1)
	url2 = strings.TrimSpace(url2)
	return url1 != "" && url2 != "" && urlutil.URLsMatch(url1, url2)
}

// titlesMatch checks if two titles match, allowing for minor differences
func (h *Handler) titlesMatch(title1, title2 string) bool {
	if title1 == title2 {
		return true
	}
	// Normalize titles by removing extra whitespace and comparing
	normalized1 := strings.TrimSpace(strings.Join(strings.Fields(title1), " "))
	normalized2 := strings.TrimSpace(strings.Join(strings.Fields(title2), " "))
	return normalized1 == normalized2
}

// publishedTimesMatch checks if two published times match within a reasonable tolerance
func (h *Handler) publishedTimesMatch(time1, time2 *time.Time) bool {
	if time1 == nil || time2 == nil {
		return false
	}
	// Allow for 1 minute difference in published times
	diff := time1.Sub(*time2)
	if diff < 0 {
		diff = -diff
	}
	return diff <= time.Minute
}

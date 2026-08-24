// Package service provides business logic services with clear interfaces
// to reduce coupling and improve testability.
package service

import (
	"context"

	"MrRSS/internal/models"
)

// ArticleService defines article-related operations
type ArticleService interface {
	// GetArticles retrieves articles based on query options
	GetArticles(ctx context.Context, opts ArticleQueryOptions) ([]models.Article, error)

	// GetArticleByID retrieves a single article by ID
	GetArticleByID(ctx context.Context, id int64) (*models.Article, error)

	// MarkRead marks an article as read/unread
	MarkRead(ctx context.Context, id int64, read bool) error

	// MarkFavorite marks an article as favorite/unfavorite
	MarkFavorite(ctx context.Context, id int64, favorite bool) error

	// MarkHidden marks an article as hidden/shown
	MarkHidden(ctx context.Context, id int64, hidden bool) error

	// GetContent retrieves the full content of an article
	GetContent(ctx context.Context, id int64) (string, error)

	// Summarize generates a summary for an article
	Summarize(ctx context.Context, id int64) (string, error)
}

// FeedService defines feed-related operations
type FeedService interface {
	// GetFeeds retrieves all feeds
	GetFeeds(ctx context.Context) ([]models.Feed, error)

	// GetFeedByID retrieves a single feed by ID
	GetFeedByID(ctx context.Context, id int64) (*models.Feed, error)

	// AddFeed adds a new feed
	AddFeed(ctx context.Context, feed *models.Feed) (int64, error)

	// UpdateFeed updates an existing feed
	UpdateFeed(ctx context.Context, feed *models.Feed) error

	// DeleteFeed deletes a feed
	DeleteFeed(ctx context.Context, id int64) error

	// RefreshFeed refreshes a single feed
	RefreshFeed(ctx context.Context, id int64) error

	// RefreshAll refreshes all feeds
	RefreshAll(ctx context.Context) error
}

// TranslationService defines translation-related operations
type TranslationService interface {
	// Translate translates text to target language
	Translate(ctx context.Context, text, targetLang string) (string, error)

	// TranslateArticle translates an article
	TranslateArticle(ctx context.Context, articleID int64, targetLang string) error
}

// AIService defines AI-related operations
type AIService interface {
	// Summarize generates a summary
	Summarize(ctx context.Context, content string) (string, error)

	// Chat handles AI chat conversations
	Chat(ctx context.Context, sessionID int64, message string) (string, error)

	// Search performs semantic search
	Search(ctx context.Context, query string) ([]models.Article, error)

	// TestConfig tests AI configuration
	TestConfig(ctx context.Context) error
}

// DiscoveryService defines feed discovery operations
type DiscoveryService interface {
	// DiscoverFromURL discovers feeds from a URL
	DiscoverFromURL(ctx context.Context, url string) ([]DiscoveredFeed, error)

	// DiscoverFromBatch discovers feeds from multiple URLs
	DiscoverFromBatch(ctx context.Context, urls []string) ([]DiscoveredFeed, error)

	// GetProgress returns discovery progress
	GetProgress() DiscoveryProgress
}

// SettingsService defines settings-related operations
type SettingsService interface {
	// Get retrieves a setting value
	Get(key string) (string, error)

	// Set sets a setting value
	Set(key, value string) error

	// GetEncrypted retrieves an encrypted setting value
	GetEncrypted(key string) (string, error)

	// SetEncrypted sets an encrypted setting value
	SetEncrypted(key, value string) error

	// GetAll retrieves all settings
	GetAll() (map[string]string, error)

	// SaveAll saves multiple settings
	SaveAll(settings map[string]string) error
}

// ArticleQueryOptions represents options for querying articles
type ArticleQueryOptions struct {
	Filter     string
	FeedID     int64
	Category   string
	ShowHidden bool
	Limit      int
	Offset     int
}

// DiscoveredFeed represents a discovered feed
type DiscoveredFeed struct {
	URL         string
	Title       string
	Description string
	Type        string
}

// DiscoveryProgress represents the progress of a discovery operation
type DiscoveryProgress struct {
	Total     int
	Completed int
	Current   string
	Status    string
}

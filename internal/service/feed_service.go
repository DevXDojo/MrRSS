package service

import (
	"context"

	"MrRSS/internal/database"
	"MrRSS/internal/models"
)

// feedService implements FeedService interface
type feedService struct {
	registry *Registry
	db       *database.DB
}

// NewFeedService creates a new feed service
func NewFeedService(registry *Registry, db *database.DB) FeedService {
	return &feedService{
		registry: registry,
		db:       db,
	}
}

// GetFeeds retrieves all feeds
func (s *feedService) GetFeeds(ctx context.Context) ([]models.Feed, error) {
	return s.db.GetFeeds()
}

// GetFeedByID retrieves a single feed by ID
func (s *feedService) GetFeedByID(ctx context.Context, id int64) (*models.Feed, error) {
	return s.db.GetFeedByID(id)
}

// AddFeed adds a new feed
func (s *feedService) AddFeed(ctx context.Context, feed *models.Feed) (int64, error) {
	return s.db.AddFeed(feed)
}

// UpdateFeed updates an existing feed
func (s *feedService) UpdateFeed(ctx context.Context, feed *models.Feed) error {
	// Use UpdateFeedWithPosition with default values for now
	return s.db.UpdateFeedWithPosition(
		feed.ID,
		feed.Title,
		feed.URL,
		feed.Category,
		feed.ScriptPath,
		feed.Position,
		feed.HideFromTimeline,
		feed.ProxyURL,
		feed.ProxyEnabled,
		feed.RefreshInterval,
		feed.IsImageMode,
		feed.Type,
		feed.XPathItem,
		feed.XPathItemTitle,
		feed.XPathItemContent,
		feed.XPathItemUri,
		feed.XPathItemAuthor,
		feed.XPathItemTimestamp,
		feed.XPathItemTimeFormat,
		feed.XPathItemThumbnail,
		feed.XPathItemCategories,
		feed.XPathItemUid,
		feed.ArticleViewMode,
		feed.AutoExpandContent,
		feed.EmailAddress,
		feed.EmailIMAPServer,
		feed.EmailUsername,
		feed.EmailPassword,
		feed.EmailFolder,
		feed.EmailIMAPPort,
	)
}

// DeleteFeed deletes a feed
func (s *feedService) DeleteFeed(ctx context.Context, id int64) error {
	return s.db.DeleteFeed(id)
}

// RefreshFeed refreshes a single feed
func (s *feedService) RefreshFeed(ctx context.Context, id int64) error {
	feed, err := s.db.GetFeedByID(id)
	if err != nil {
		return err
	}
	if feed == nil {
		return nil
	}

	s.registry.Fetcher().FetchFeedForArticle(ctx, *feed)
	return nil
}

// RefreshAll refreshes all feeds
func (s *feedService) RefreshAll(ctx context.Context) error {
	s.registry.Fetcher().FetchAll(ctx)
	return nil
}

package service

import (
	"context"

	"MrRSS/internal/database"
	"MrRSS/internal/models"
)

// articleService implements ArticleService interface
type articleService struct {
	registry *Registry
	db       *database.DB
}

// NewArticleService creates a new article service
func NewArticleService(registry *Registry, db *database.DB) ArticleService {
	return &articleService{
		registry: registry,
		db:       db,
	}
}

// GetArticles retrieves articles based on query options
func (s *articleService) GetArticles(ctx context.Context, opts ArticleQueryOptions) ([]models.Article, error) {
	return s.db.GetArticles(opts.Filter, opts.FeedID, opts.Category, opts.ShowHidden, opts.Limit, opts.Offset)
}

// GetArticleByID retrieves a single article by ID
func (s *articleService) GetArticleByID(ctx context.Context, id int64) (*models.Article, error) {
	return s.db.GetArticleByID(id)
}

// MarkRead marks an article as read/unread
func (s *articleService) MarkRead(ctx context.Context, id int64, read bool) error {
	return s.db.MarkArticleRead(id, read)
}

// MarkFavorite marks an article as favorite/unfavorite
func (s *articleService) MarkFavorite(ctx context.Context, id int64, favorite bool) error {
	return s.db.SetArticleFavorite(id, favorite)
}

// MarkHidden marks an article as hidden/shown
func (s *articleService) MarkHidden(ctx context.Context, id int64, hidden bool) error {
	return s.db.SetArticleHidden(id, hidden)
}

// GetContent retrieves the full content of an article
func (s *articleService) GetContent(ctx context.Context, id int64) (string, error) {
	// Check memory cache first
	content, found := s.registry.ContentCache().Get(id)
	if found && content != "" {
		return content, nil
	}

	// Fallback to database
	content, found, err := s.db.GetArticleContent(id)
	if err != nil {
		return "", err
	}
	if found {
		s.registry.ContentCache().Set(id, content)
		return content, nil
	}

	return "", nil
}

// Summarize generates a summary for an article
func (s *articleService) Summarize(ctx context.Context, id int64) (string, error) {
	// Get article content first
	content, err := s.GetContent(ctx, id)
	if err != nil {
		return "", err
	}

	// Use AI service to generate summary
	return s.registry.AI().Summarize(ctx, content)
}

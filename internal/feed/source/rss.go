package source

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

// RSSSource fetches feeds via standard HTTP RSS/Atom requests.
type RSSSource struct {
	parser *gofeed.Parser
	client *http.Client
}

// NewRSSSource creates a new RSS source.
// Uses a default HTTP client with 30s timeout.
func NewRSSSource() *RSSSource {
	client := &http.Client{Timeout: 30 * time.Second}

	parser := gofeed.NewParser()
	parser.Client = client

	return &RSSSource{
		parser: parser,
		client: client,
	}
}

// Type returns the source type identifier.
func (s *RSSSource) Type() Type {
	return TypeRSS
}

// Validate checks if the configuration is valid for RSS source.
func (s *RSSSource) Validate(config *Config) error {
	if config == nil {
		return errors.New("config is nil")
	}
	if config.URL == "" {
		return errors.New("URL is required for RSS source")
	}
	return nil
}

// Fetch retrieves and parses the RSS/Atom feed from the URL.
func (s *RSSSource) Fetch(ctx context.Context, config *Config) (*gofeed.Feed, error) {
	if err := s.Validate(config); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Use context-aware parsing if available
	feed, err := s.parser.ParseURLWithContext(config.URL, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed from %s: %w", config.URL, err)
	}

	return feed, nil
}

// SetHTTPClient updates the HTTP client used for requests.
func (s *RSSSource) SetHTTPClient(client *http.Client) {
	s.client = client
	s.parser.Client = client
}

package service

import (
	"context"

	"MrRSS/internal/discovery"
)

// discoveryServiceWrapper wraps the existing discovery.Service
type discoveryServiceWrapper struct {
	service *discovery.Service
}

// NewDiscoveryServiceWrapper creates a new discovery service wrapper
func NewDiscoveryServiceWrapper(service *discovery.Service) DiscoveryService {
	return &discoveryServiceWrapper{
		service: service,
	}
}

// DiscoverFromURL discovers feeds from a URL (wraps DiscoverFromFeed)
func (s *discoveryServiceWrapper) DiscoverFromURL(ctx context.Context, url string) ([]DiscoveredFeed, error) {
	feeds, err := s.service.DiscoverFromFeed(ctx, url)
	if err != nil {
		return nil, err
	}

	// Convert to our format
	result := make([]DiscoveredFeed, len(feeds))
	for i, f := range feeds {
		result[i] = DiscoveredFeed{
			URL:         f.RSSFeed,
			Title:       f.Name,
			Description: f.Homepage,
			Type:        "rss",
		}
	}
	return result, nil
}

// DiscoverFromBatch discovers feeds from multiple URLs
// Note: This is a simplified implementation that processes URLs one by one
func (s *discoveryServiceWrapper) DiscoverFromBatch(ctx context.Context, urls []string) ([]DiscoveredFeed, error) {
	var allFeeds []DiscoveredFeed
	for _, url := range urls {
		feeds, err := s.DiscoverFromURL(ctx, url)
		if err != nil {
			continue // Skip errors, collect what we can
		}
		allFeeds = append(allFeeds, feeds...)
	}
	return allFeeds, nil
}

// GetProgress returns discovery progress
// Note: The actual discovery service doesn't store progress globally
// This returns a placeholder empty progress
func (s *discoveryServiceWrapper) GetProgress() DiscoveryProgress {
	return DiscoveryProgress{
		Total:     0,
		Completed: 0,
		Current:   "",
		Status:    "",
	}
}

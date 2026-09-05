package service

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"MrRSS/internal/ai"
	"MrRSS/internal/config"
	"MrRSS/internal/database"
	"MrRSS/internal/models"
	"MrRSS/internal/utils/httputil"
)

// aiService implements AIService interface
type aiService struct {
	registry *Registry
	db       *database.DB
}

// NewAIService creates a new AI service
func NewAIService(registry *Registry, db *database.DB) AIService {
	return &aiService{
		registry: registry,
		db:       db,
	}
}

// Summarize generates a summary
func (s *aiService) Summarize(ctx context.Context, content string) (string, error) {
	// Check AI usage limit
	if !s.registry.AITracker().CanMakeRequest() {
		return "", fmt.Errorf("daily AI usage limit reached")
	}

	// Get AI settings
	apiKey, _ := s.db.GetEncryptedSetting("ai_api_key")
	endpoint, _ := s.db.GetSetting("ai_endpoint")
	model, _ := s.db.GetSetting("ai_model")

	// Use defaults if not set
	defaults := config.Get()
	if endpoint == "" {
		endpoint = defaults.AIEndpoint
	}
	if model == "" {
		model = defaults.AIModel
	}

	// Create AI client using the same transport as configuration tests.
	httpClient, err := s.createHTTPClientWithProxy()
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP client: %w", err)
	}
	clientConfig := ai.ClientConfig{
		APIKey:   apiKey,
		Endpoint: endpoint,
		Model:    model,
		Timeout:  30 * time.Second,
	}
	client := ai.NewClientWithHTTPClient(clientConfig, httpClient)

	// Generate summary
	response, err := client.Request(content, "Summarize this article")
	if err != nil {
		return "", err
	}

	// Track usage
	s.registry.AITracker().AddUsage(int64(len(content)))

	return response, nil
}

// Chat handles AI chat conversations
func (s *aiService) Chat(ctx context.Context, sessionID int64, message string) (string, error) {
	// Check AI usage limit
	if !s.registry.AITracker().CanMakeRequest() {
		return "", fmt.Errorf("daily AI usage limit reached")
	}

	// Get AI settings
	apiKey, _ := s.db.GetEncryptedSetting("ai_api_key")
	endpoint, _ := s.db.GetSetting("ai_endpoint")
	model, _ := s.db.GetSetting("ai_model")

	// Use defaults if not set
	defaults := config.Get()
	if endpoint == "" {
		endpoint = defaults.AIEndpoint
	}
	if model == "" {
		model = defaults.AIModel
	}

	// Create AI client using the same transport as configuration tests.
	httpClient, err := s.createHTTPClientWithProxy()
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP client: %w", err)
	}
	clientConfig := ai.ClientConfig{
		APIKey:   apiKey,
		Endpoint: endpoint,
		Model:    model,
		Timeout:  30 * time.Second,
	}
	client := ai.NewClientWithHTTPClient(clientConfig, httpClient)

	// Send chat message
	response, err := client.Request(message, "")
	if err != nil {
		return "", err
	}

	// Track usage
	s.registry.AITracker().AddUsage(int64(len(message)))

	return response, nil
}

// Search performs semantic search
func (s *aiService) Search(ctx context.Context, query string) ([]models.Article, error) {
	// This is a placeholder - actual implementation would use vector embeddings
	// For now, return empty results
	return []models.Article{}, nil
}

// TestConfig tests AI configuration
func (s *aiService) TestConfig(ctx context.Context) error {
	// Get AI settings
	apiKey, _ := s.db.GetEncryptedSetting("ai_api_key")
	endpoint, _ := s.db.GetSetting("ai_endpoint")
	model, _ := s.db.GetSetting("ai_model")

	// Use defaults if not set
	defaults := config.Get()
	if endpoint == "" {
		endpoint = defaults.AIEndpoint
	}
	if model == "" {
		model = defaults.AIModel
	}

	// Validate endpoint URL format
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("API endpoint must use HTTP or HTTPS")
	}

	// Create HTTP client with proxy support if configured
	httpClient, err := s.createHTTPClientWithProxy()
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}
	httpClient.Timeout = 30 * time.Second

	// Create AI client for testing
	clientConfig := ai.ClientConfig{
		APIKey:   apiKey,
		Endpoint: endpoint,
		Model:    model,
		Timeout:  30 * time.Second,
	}
	client := ai.NewClientWithHTTPClient(clientConfig, httpClient)

	// Try a simple test request
	_, err = client.Request("", "test")
	return err
}

// createHTTPClientWithProxy creates the canonical HTTP client with global proxy settings.
func (s *aiService) createHTTPClientWithProxy() (*http.Client, error) {
	return httputil.CreateHTTPClientWithProxySettings(s.db, 30*time.Second)
}

package service

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"MrRSS/internal/ai"
	"MrRSS/internal/config"
	"MrRSS/internal/database"
	"MrRSS/internal/models"
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

	// Create AI client
	clientConfig := ai.ClientConfig{
		APIKey:   apiKey,
		Endpoint: endpoint,
		Model:    model,
		Timeout:  30 * time.Second,
	}
	client := ai.NewClient(clientConfig)

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

	// Create AI client
	clientConfig := ai.ClientConfig{
		APIKey:   apiKey,
		Endpoint: endpoint,
		Model:    model,
		Timeout:  30 * time.Second,
	}
	client := ai.NewClient(clientConfig)

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

// createHTTPClientWithProxy creates an HTTP client with global proxy settings if enabled
func (s *aiService) createHTTPClientWithProxy() (*http.Client, error) {
	// Check if global proxy is enabled
	proxyEnabled, _ := s.db.GetSetting("proxy_enabled")
	if proxyEnabled != "true" {
		return &http.Client{}, nil
	}

	// Build proxy URL from global settings
	proxyType, _ := s.db.GetSetting("proxy_type")
	proxyHost, _ := s.db.GetSetting("proxy_host")
	proxyPort, _ := s.db.GetSetting("proxy_port")
	proxyUsername, _ := s.db.GetEncryptedSetting("proxy_username")
	proxyPassword, _ := s.db.GetEncryptedSetting("proxy_password")

	// Build proxy URL
	proxyURL := s.buildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)

	if proxyURL == "" {
		return &http.Client{}, nil
	}

	// Parse proxy URL
	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(u),
		},
	}, nil
}

// buildProxyURL builds a proxy URL from components
func (s *aiService) buildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword string) string {
	if proxyHost == "" || proxyPort == "" {
		return ""
	}

	var urlBuilder strings.Builder
	urlBuilder.WriteString(strings.ToLower(proxyType))
	urlBuilder.WriteString("://")

	if proxyUsername != "" && proxyPassword != "" {
		urlBuilder.WriteString(url.QueryEscape(proxyUsername))
		urlBuilder.WriteString(":")
		urlBuilder.WriteString(url.QueryEscape(proxyPassword))
		urlBuilder.WriteString("@")
	}

	urlBuilder.WriteString(proxyHost)
	urlBuilder.WriteString(":")
	urlBuilder.WriteString(proxyPort)

	return urlBuilder.String()
}

package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"MrRSS/internal/ai"
	"MrRSS/internal/config"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

// TestResult represents the result of AI configuration test
type TestResult struct {
	ConfigValid       bool   `json:"config_valid"`
	ConnectionSuccess bool   `json:"connection_success"`
	ModelAvailable    bool   `json:"model_available"`
	ResponseTimeMs    int64  `json:"response_time_ms"`
	TestTime          string `json:"test_time"`
	ErrorMessage      string `json:"error_message,omitempty"`
}

// HandleTestAIConfig handles POST /api/ai/test to test AI configuration
// @Summary      Test AI configuration
// @Description  Test AI service configuration (endpoint, API key, model availability)
// @Tags         ai
// @Accept       json
// @Produce      json
// @Success      200  {object}  handlers.TestResult  "Test result (config_valid, connection_success, model_available, response_time_ms, test_time, error_message)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /ai/test [post]
func HandleTestAIConfig(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	result := TestResult{
		TestTime: time.Now().Format(time.RFC3339),
	}

	// Get AI settings
	apiKey, _ := h.DB.GetEncryptedSetting("ai_api_key")
	endpoint, _ := h.DB.GetSetting("ai_endpoint")
	model, _ := h.DB.GetSetting("ai_model")

	// Use defaults if not set
	defaults := config.Get()
	if endpoint == "" {
		endpoint = defaults.AIEndpoint
	}
	if model == "" {
		model = defaults.AIModel
	}

	// Validate configuration
	result.ConfigValid = true
	validationErrors := []string{}

	if endpoint == "" {
		validationErrors = append(validationErrors, "endpoint is required")
		result.ConfigValid = false
	}

	if model == "" {
		validationErrors = append(validationErrors, "model is required")
		result.ConfigValid = false
	}

	if !result.ConfigValid {
		result.ErrorMessage = "Configuration incomplete: " + strings.Join(validationErrors, ", ")
		response.JSON(w, result)
		return
	}

	// Validate endpoint URL format
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		result.ConfigValid = false
		result.ErrorMessage = "Invalid endpoint URL: " + err.Error()
		response.JSON(w, result)
		return
	}

	// Both HTTP and HTTPS are allowed
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		result.ConfigValid = false
		result.ErrorMessage = "API endpoint must use HTTP or HTTPS"
		response.JSON(w, result)
		return
	}

	// Test connection with a simple request
	startTime := time.Now()

	// Create HTTP client with proxy support if configured
	httpClient, err := createHTTPClientWithProxy(h)
	if err != nil {
		result.ConnectionSuccess = false
		result.ModelAvailable = false
		result.ErrorMessage = fmt.Sprintf("Failed to create HTTP client: %v", err)
		result.ResponseTimeMs = time.Since(startTime).Milliseconds()
		response.JSON(w, result)
		return
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

	if err != nil {
		result.ConnectionSuccess = false
		result.ModelAvailable = false
		result.ErrorMessage = fmt.Sprintf("Connection failed: %v", err)
	} else {
		result.ConnectionSuccess = true
		result.ModelAvailable = true
	}

	result.ResponseTimeMs = time.Since(startTime).Milliseconds()

	response.JSON(w, result)
}

// HandleGetAITestInfo handles GET /api/ai/test/info to get last test result
// @Summary      Get AI test info
// @Description  Get the last AI configuration test result (returns empty/default result as tests are not stored persistently)
// @Tags         ai
// @Accept       json
// @Produce      json
// @Success      200  {object}  handlers.TestResult  "Empty test result (config_valid, connection_success, model_available, response_time_ms, test_time)"
// @Router       /ai/test/info [get]
func HandleGetAITestInfo(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Return default/empty result - tests are not stored persistently
	// The frontend will trigger a new test if needed
	result := TestResult{
		ConfigValid:       false,
		ConnectionSuccess: false,
		ModelAvailable:    false,
		ResponseTimeMs:    0,
		TestTime:          "",
	}

	response.JSON(w, result)
}

// createHTTPClientWithProxy creates an HTTP client with global proxy settings if enabled
func createHTTPClientWithProxy(h *core.Handler) (*http.Client, error) {
	// Check if global proxy is enabled
	proxyEnabled, _ := h.DB.GetSetting("proxy_enabled")
	if proxyEnabled != "true" {
		return &http.Client{}, nil
	}

	// Build proxy URL from global settings
	proxyType, _ := h.DB.GetSetting("proxy_type")
	proxyHost, _ := h.DB.GetSetting("proxy_host")
	proxyPort, _ := h.DB.GetSetting("proxy_port")
	proxyUsername, _ := h.DB.GetEncryptedSetting("proxy_username")
	proxyPassword, _ := h.DB.GetEncryptedSetting("proxy_password")

	// Build proxy URL
	proxyURL := buildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)

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
func buildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword string) string {
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

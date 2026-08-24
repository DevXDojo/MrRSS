package translation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CustomTranslator implements a fully customizable HTTP-based translation service
type CustomTranslator struct {
	config         *CustomTranslatorConfig
	client         *http.Client
	db             DBInterface
	cachedMappings map[string]string // Cached language code mappings
}

// CustomTranslatorConfig holds the configuration for custom translation API
type CustomTranslatorConfig struct {
	Name            string            // Configuration name
	Endpoint        string            // API endpoint URL
	Method          string            // HTTP method (GET/POST)
	Headers         map[string]string // Custom HTTP headers
	BodyTemplate    string            // Request body template with placeholders
	ResponsePath    string            // JSONPath to extract translation from response
	LangCodeMapping map[string]string // Custom language code mapping
	Timeout         int               // Request timeout in seconds
}

// NewCustomTranslator creates a new custom translator with the given configuration
// db is optional - if nil, no proxy will be used
func NewCustomTranslator(config *CustomTranslatorConfig) *CustomTranslator {
	client, err := CreateHTTPClientWithProxy(nil, time.Duration(config.Timeout)*time.Second)
	if err != nil {
		client = &http.Client{Timeout: time.Duration(config.Timeout) * time.Second}
	}

	return &CustomTranslator{
		config: config,
		client: client,
		db:     nil,
	}
}

// NewCustomTranslatorWithDB creates a new custom translator with database for proxy support
func NewCustomTranslatorWithDB(config *CustomTranslatorConfig, db DBInterface) *CustomTranslator {
	client, err := CreateHTTPClientWithProxy(db, time.Duration(config.Timeout)*time.Second)
	if err != nil {
		// Fallback to default client if proxy creation fails
		client = &http.Client{Timeout: time.Duration(config.Timeout) * time.Second}
	}
	return &CustomTranslator{
		config:         config,
		client:         client,
		db:             db,
		cachedMappings: config.LangCodeMapping,
	}
}

// Translate translates text using the configured custom API
func (t *CustomTranslator) Translate(text, targetLang string) (string, error) {
	if text == "" {
		return "", nil
	}

	// Validate configuration
	if err := t.validateConfig(); err != nil {
		return "", fmt.Errorf("invalid custom translator configuration: %w", err)
	}

	// Map target language if custom mapping is provided
	mappedLang := t.mapLanguageCode(targetLang)

	// Build request
	req, err := t.buildRequest(text, mappedLang)
	if err != nil {
		return "", fmt.Errorf("failed to build request: %w", err)
	}

	// Set custom headers
	t.setHeaders(req)

	// Execute request
	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("translation request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response and extract translation
	result, err := t.extractTranslation(resp)
	if err != nil {
		return "", fmt.Errorf("failed to extract translation: %w", err)
	}

	return result, nil
}

// validateConfig checks if the configuration is valid
func (t *CustomTranslator) validateConfig() error {
	if t.config.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}

	// Validate URL format
	if _, err := url.Parse(t.config.Endpoint); err != nil {
		return fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// Validate method
	method := strings.ToUpper(t.config.Method)
	if method != "GET" && method != "POST" {
		return fmt.Errorf("method must be GET or POST")
	}

	// For POST requests, body template is required
	if method == "POST" && t.config.BodyTemplate == "" {
		return fmt.Errorf("body template is required for POST requests")
	}

	// Response path is required
	if t.config.ResponsePath == "" {
		return fmt.Errorf("response path is required")
	}

	return nil
}

// buildRequest creates the HTTP request with proper body
func (t *CustomTranslator) buildRequest(text, targetLang string) (*http.Request, error) {
	method := strings.ToUpper(t.config.Method)

	var req *http.Request
	var err error

	if method == "GET" {
		// Build query parameters for GET request
		req, err = t.buildGetRequest(text, targetLang)
	} else {
		// Build body for POST request
		req, err = t.buildPostRequest(text, targetLang)
	}

	if err != nil {
		return nil, err
	}

	return req, nil
}

// buildGetRequest builds a GET request with query parameters
func (t *CustomTranslator) buildGetRequest(text, targetLang string) (*http.Request, error) {
	// For GET, we replace placeholders in the URL query string
	// The template should be a URL-encoded query string or full URL
	u, err := url.Parse(t.config.Endpoint)
	if err != nil {
		return nil, err
	}

	// Replace placeholders in query parameters
	query := u.Query()
	query.Set("text", text)
	query.Set("target_lang", targetLang)
	u.RawQuery = query.Encode()

	return http.NewRequest("GET", u.String(), nil)
}

// buildPostRequest builds a POST request with JSON body
func (t *CustomTranslator) buildPostRequest(text, targetLang string) (*http.Request, error) {
	// Replace placeholders in body template
	body := t.replacePlaceholders(t.config.BodyTemplate, text, targetLang)

	return http.NewRequest("POST", t.config.Endpoint, bytes.NewBufferString(body))
}

// setHeaders sets custom headers on the request
func (t *CustomTranslator) setHeaders(req *http.Request) {
	// Set default content type for POST requests
	if strings.ToUpper(req.Method) == "POST" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for key, value := range t.config.Headers {
		req.Header.Set(key, value)
	}
}

// extractTranslation parses the response and extracts the translated text
func (t *CustomTranslator) extractTranslation(resp *http.Response) (string, error) {
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse response as JSON
	var responseData interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return "", fmt.Errorf("response is not valid JSON: %w", err)
	}

	// Extract translation using JSONPath-like syntax
	result, err := t.extractByPath(responseData, t.config.ResponsePath)
	if err != nil {
		return "", err
	}

	// Convert result to string
	if str, ok := result.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("extracted value is not a string")
}

// extractByPath extracts a value from nested JSON using a dot-notation path
func (t *CustomTranslator) extractByPath(data interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("path '%s' not found in response", part)
			}
		case []interface{}:
			// Handle array indexing (e.g., "0" or "[0]")
			idx := 0
			if _, err := fmt.Sscanf(part, "[%d]", &idx); err == nil {
				if idx >= 0 && idx < len(v) {
					current = v[idx]
				} else {
					return nil, fmt.Errorf("array index %d out of bounds", idx)
				}
			} else {
				// Try to parse as plain number
				if _, err := fmt.Sscanf(part, "%d", &idx); err == nil {
					if idx >= 0 && idx < len(v) {
						current = v[idx]
					} else {
						return nil, fmt.Errorf("array index %d out of bounds", idx)
					}
				} else {
					return nil, fmt.Errorf("cannot use '%s' on array", part)
				}
			}
		default:
			return nil, fmt.Errorf("cannot access '%s' on non-object/array", part)
		}
	}

	return current, nil
}

// replacePlaceholders replaces template placeholders with actual values
// Supported placeholders: {{text}}, {{target_lang}}, {{source_lang}}
func (t *CustomTranslator) replacePlaceholders(template, text, targetLang string) string {
	result := template
	result = strings.ReplaceAll(result, `"{{text}}"`, escapeJSONString(text))
	result = strings.ReplaceAll(result, "{{text}}", escapeJSONString(text))
	result = strings.ReplaceAll(result, "{{target_lang}}", targetLang)
	result = strings.ReplaceAll(result, "{{source_lang}}", "auto")
	return result
}

// mapLanguageCode maps a language code using custom mapping
func (t *CustomTranslator) mapLanguageCode(lang string) string {
	if t.cachedMappings == nil {
		return lang
	}

	if mapped, ok := t.cachedMappings[lang]; ok {
		return mapped
	}

	return lang
}

// escapeJSONString escapes special characters in a JSON string
func escapeJSONString(s string) string {
	// Simple JSON escaping - in production, use json.Marshal
	b, _ := json.Marshal(s)
	return string(b)
}

// ParseConfigFromSettings creates a CustomTranslatorConfig from settings values
func ParseConfigFromSettings(
	name, endpoint, method, headers, bodyTemplate, responsePath, langMapping string,
	timeout int,
) (*CustomTranslatorConfig, error) {
	// Parse headers JSON
	headerMap := make(map[string]string)
	if headers != "" {
		if err := json.Unmarshal([]byte(headers), &headerMap); err != nil {
			return nil, fmt.Errorf("invalid headers JSON: %w", err)
		}
	}

	// Parse language mapping JSON
	langMap := make(map[string]string)
	if langMapping != "" {
		if err := json.Unmarshal([]byte(langMapping), &langMap); err != nil {
			return nil, fmt.Errorf("invalid language mapping JSON: %w", err)
		}
	}

	return &CustomTranslatorConfig{
		Name:            name,
		Endpoint:        endpoint,
		Method:          method,
		Headers:         headerMap,
		BodyTemplate:    bodyTemplate,
		ResponsePath:    responsePath,
		LangCodeMapping: langMap,
		Timeout:         timeout,
	}, nil
}

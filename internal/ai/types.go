// Package ai provides shared types and interfaces for AI client operations
package ai

import (
	"encoding/json"
	"strings"
)

// FormatType represents the type of API format
type FormatType string

const (
	FormatTypeGemini    FormatType = "gemini"
	FormatTypeOpenAI    FormatType = "openai"
	FormatTypeOllama    FormatType = "ollama"
	FormatTypeAnthropic FormatType = "anthropic"
	FormatTypeDeepSeek  FormatType = "deepseek"
)

// RequestConfig holds the configuration for an AI request
type RequestConfig struct {
	Model               string
	SystemPrompt        string
	UserPrompt          string
	Messages            []map[string]string    // Alternative to SystemPrompt+UserPrompt
	Temperature         float64                // Optional temperature override
	MaxTokens           int                    // Optional max tokens override (deprecated for OpenAI)
	MaxCompletionTokens int                    // OpenAI: new parameter for max completion tokens
	ReasoningEffort     string                 // OpenAI: reasoning effort for o-series models ("none", "minimal", "low", "medium", "high")
	ResponseFormat      map[string]interface{} // OpenAI/Ollama: JSON schema for structured outputs
	ThinkingConfig      map[string]interface{} // Gemini: thinking configuration
	PresencePenalty     float64                // OpenAI/Gemini: presence penalty
	FrequencyPenalty    float64                // OpenAI/Gemini: frequency penalty
	TopP                float64                // Top-p sampling
	TopK                int                    // Top-k sampling (Gemini/Ollama)
	Seed                int                    // Seed for reproducible outputs
}

// ResponseResult holds the result from an AI API call
type ResponseResult struct {
	Content    string     // The main response content
	Thinking   string     // Optional thinking/reasoning content (for models that support it)
	FormatUsed FormatType // Which format was successful
}

// FormatHandler defines the interface for handling different API formats
type FormatHandler interface {
	// BuildRequest builds the request body for this format
	BuildRequest(config RequestConfig) (map[string]interface{}, error)

	// ParseResponse parses the response body for this format
	ParseResponse(body []byte) (ResponseResult, error)

	// FormatEndpoint formats the endpoint URL if needed (can return as-is)
	FormatEndpoint(endpoint, model string) string

	// ValidateResponse checks if the HTTP response indicates success
	ValidateResponse(statusCode int, body []byte) error
}

// ParseCustomHeaders parses custom headers from JSON string to map
func ParseCustomHeaders(headersJSON string) map[string]string {
	headers := make(map[string]string)
	if headersJSON == "" {
		return headers
	}

	// Try to parse as JSON array of key-value pairs
	var pairs []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal([]byte(headersJSON), &pairs); err == nil {
		for _, pair := range pairs {
			if pair.Key != "" {
				headers[pair.Key] = pair.Value
			}
		}
		return headers
	}

	// Fallback: try to parse as simple key:value format (one per line)
	lines := strings.Split(headersJSON, "\n")
	for _, line := range lines {
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return headers
}

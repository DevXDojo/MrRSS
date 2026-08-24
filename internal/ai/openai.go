// Package ai provides OpenAI-compatible API format handlers
package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// OpenAIHandler implements FormatHandler for OpenAI-compatible APIs
type OpenAIHandler struct{}

// NewOpenAIHandler creates a new OpenAI format handler
func NewOpenAIHandler() *OpenAIHandler {
	return &OpenAIHandler{}
}

// BuildRequest builds an OpenAI-compatible API request
func (h *OpenAIHandler) BuildRequest(config RequestConfig) (map[string]interface{}, error) {
	request := map[string]interface{}{
		"model": config.Model,
	}

	// Determine messages format
	if len(config.Messages) > 0 {
		// Use provided messages
		request["messages"] = config.Messages
	} else {
		// Build messages from system and user prompts
		messages := []map[string]string{}
		if config.SystemPrompt != "" {
			messages = append(messages, map[string]string{
				"role":    "system",
				"content": config.SystemPrompt,
			})
		}
		if config.UserPrompt != "" {
			messages = append(messages, map[string]string{
				"role":    "user",
				"content": config.UserPrompt,
			})
		}
		request["messages"] = messages
	}

	// Add optional parameters
	if config.Temperature > 0 {
		request["temperature"] = config.Temperature
	}

	// Use max_completion_tokens if provided (new parameter)
	if config.MaxCompletionTokens > 0 {
		request["max_completion_tokens"] = config.MaxCompletionTokens
	} else if config.MaxTokens > 0 {
		// Fallback to max_tokens for backward compatibility
		request["max_tokens"] = config.MaxTokens
	}

	// Reasoning effort for o-series models
	if config.ReasoningEffort != "" {
		request["reasoning_effort"] = config.ReasoningEffort
	}

	// Response format for structured outputs
	if config.ResponseFormat != nil {
		request["response_format"] = config.ResponseFormat
	}

	// Presence and frequency penalties
	if config.PresencePenalty != 0 {
		request["presence_penalty"] = config.PresencePenalty
	}
	if config.FrequencyPenalty != 0 {
		request["frequency_penalty"] = config.FrequencyPenalty
	}

	// Top-p sampling
	if config.TopP > 0 {
		request["top_p"] = config.TopP
	}

	// Seed for reproducible outputs
	if config.Seed > 0 {
		request["seed"] = config.Seed
	}

	return request, nil
}

// ParseResponse parses an OpenAI-compatible API response
func (h *OpenAIHandler) ParseResponse(body []byte) (ResponseResult, error) {
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return ResponseResult{}, fmt.Errorf("failed to decode OpenAI response: %w", err)
	}

	// Check for API error
	if response.Error != nil {
		return ResponseResult{}, fmt.Errorf("OpenAI API error: %s (type: %s)", response.Error.Message, response.Error.Type)
	}

	// Extract content
	if len(response.Choices) == 0 {
		return ResponseResult{}, fmt.Errorf("no choices in OpenAI response")
	}

	content := strings.TrimSpace(response.Choices[0].Message.Content)
	if content == "" {
		return ResponseResult{}, fmt.Errorf("empty content in OpenAI response")
	}

	return ResponseResult{
		Content:    content,
		FormatUsed: FormatTypeOpenAI,
	}, nil
}

// ValidateResponse validates the HTTP response status
func (h *OpenAIHandler) ValidateResponse(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("authentication failed - check API key")
	case http.StatusNotFound:
		return fmt.Errorf("model not found")
	case http.StatusBadRequest:
		return fmt.Errorf("bad request - check parameters")
	default:
		return fmt.Errorf("OpenAI API returned status %d: %s", statusCode, string(body))
	}
}

// FormatEndpoint returns the endpoint as-is for OpenAI format
func (h *OpenAIHandler) FormatEndpoint(endpoint, model string) string {
	return strings.TrimSuffix(endpoint, "/")
}

// IsOpenAIError checks if an error message indicates an OpenAI API format
func IsOpenAIError(errorMessage string) bool {
	openAIErrorPatterns := []string{
		"incorrect API key provided",
		"invalid_api_key",
		"context_length_exceeded",
		"rate_limit_exceeded",
		"server_error",
		"openai",
	}

	for _, pattern := range openAIErrorPatterns {
		if strings.Contains(strings.ToLower(errorMessage), pattern) {
			return true
		}
	}

	return false
}

// Package ai provides Anthropic Claude API format handler
package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// AnthropicHandler implements the FormatHandler interface for Anthropic Claude API
type AnthropicHandler struct{}

// BuildRequest constructs a request body for Anthropic Claude API
func (h *AnthropicHandler) BuildRequest(config RequestConfig) (map[string]interface{}, error) {
	request := make(map[string]interface{})

	// Model (required)
	if config.Model != "" {
		request["model"] = config.Model
	} else {
		return nil, fmt.Errorf("model is required for Anthropic API")
	}

	// Max tokens (required for Anthropic)
	maxTokens := config.MaxTokens
	if config.MaxCompletionTokens > 0 {
		maxTokens = config.MaxCompletionTokens
	}
	if maxTokens == 0 {
		maxTokens = 4096 // Default max tokens
	}
	request["max_tokens"] = maxTokens

	// Messages
	messages := []map[string]interface{}{}

	if len(config.Messages) > 0 {
		// Use provided messages
		for _, msg := range config.Messages {
			messages = append(messages, map[string]interface{}{
				"role":    msg["role"],
				"content": msg["content"],
			})
		}
	} else {
		// Construct from system and user prompts
		// Note: Anthropic uses a separate "system" field, not in messages
		if config.UserPrompt != "" {
			messages = append(messages, map[string]interface{}{
				"role":    "user",
				"content": config.UserPrompt,
			})
		}
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("at least one message is required")
	}
	request["messages"] = messages

	// System prompt (separate field in Anthropic API)
	if config.SystemPrompt != "" {
		request["system"] = config.SystemPrompt
	}

	// Temperature
	if config.Temperature > 0 {
		request["temperature"] = config.Temperature
	}

	// Top P
	if config.TopP > 0 {
		request["top_p"] = config.TopP
	}

	// Top K (Anthropic supports this)
	if config.TopK > 0 {
		request["top_k"] = config.TopK
	}

	// Thinking (for Claude 3.5+ with extended thinking)
	// Anthropic uses "thinking" parameter for extended thinking mode
	if config.ThinkingConfig != nil {
		if budget, ok := config.ThinkingConfig["thinkingBudget"].(int); ok && budget > 0 {
			request["thinking"] = map[string]interface{}{
				"type":   "enabled",
				"budget": budget,
			}
		} else if enabled, ok := config.ThinkingConfig["includeThoughts"].(bool); ok && enabled {
			request["thinking"] = map[string]interface{}{
				"type": "enabled",
			}
		}
	}

	// Stop sequences
	if stopSeq, ok := config.ResponseFormat["stop_sequences"].([]string); ok && len(stopSeq) > 0 {
		request["stop_sequences"] = stopSeq
	}

	// Metadata (optional, for tracking)
	metadata := make(map[string]interface{})
	if config.Seed > 0 {
		// Anthropic doesn't have native seed support, but we can track it in metadata
		metadata["seed"] = config.Seed
	}
	if len(metadata) > 0 {
		request["metadata"] = metadata
	}

	return request, nil
}

// ParseResponse extracts the content from Anthropic API response
func (h *AnthropicHandler) ParseResponse(body []byte) (ResponseResult, error) {
	var response struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Model        string `json:"model"`
		StopReason   string `json:"stop_reason"`
		StopSequence string `json:"stop_sequence"`
		Usage        struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return ResponseResult{}, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

	// Check for API errors
	if response.Error.Message != "" {
		return ResponseResult{}, fmt.Errorf("anthropic API error (%s): %s", response.Error.Type, response.Error.Message)
	}

	// Extract content
	var contentBuilder strings.Builder
	var thinkingContent string

	for _, content := range response.Content {
		switch content.Type {
		case "text":
			contentBuilder.WriteString(content.Text)
		case "thinking":
			// Extended thinking content
			thinkingContent = content.Text
		}
	}

	result := ResponseResult{
		Content:    contentBuilder.String(),
		Thinking:   thinkingContent,
		FormatUsed: FormatTypeAnthropic,
	}

	return result, nil
}

// ValidateResponse checks if the response is valid
func (h *AnthropicHandler) ValidateResponse(statusCode int, body []byte) error {
	var response struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
		Content []struct {
			Type string `json:"type"`
		} `json:"content"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("invalid JSON response: %w", err)
	}

	if response.Error.Message != "" {
		return fmt.Errorf("API error (%s): %s", response.Error.Type, response.Error.Message)
	}

	if len(response.Content) == 0 {
		return fmt.Errorf("no content in response")
	}

	return nil
}

// FormatEndpoint formats the API endpoint URL
func (h *AnthropicHandler) FormatEndpoint(endpoint, model string) string {
	// Anthropic Messages API endpoint
	if endpoint == "" {
		return "https://api.anthropic.com/v1/messages"
	}

	// Use endpoint as-is (user should provide full API path)
	return strings.TrimSuffix(endpoint, "/")
}

// GetRequiredHeaders returns the required HTTP headers for Anthropic API
func (h *AnthropicHandler) GetRequiredHeaders(apiKey string) map[string]string {
	headers := make(map[string]string)

	// Anthropic requires specific headers
	headers["x-api-key"] = apiKey
	headers["anthropic-version"] = "2023-06-01" // API version
	headers["content-type"] = "application/json"

	return headers
}

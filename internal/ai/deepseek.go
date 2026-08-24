// Package ai provides DeepSeek API format handler
package ai

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// DeepSeekHandler implements the FormatHandler interface for DeepSeek API
// DeepSeek API is OpenAI-compatible but has some specific features
type DeepSeekHandler struct{}

// BuildRequest constructs a request body for DeepSeek API
func (h *DeepSeekHandler) BuildRequest(config RequestConfig) (map[string]interface{}, error) {
	request := make(map[string]interface{})

	// Model (required)
	if config.Model != "" {
		request["model"] = config.Model
	} else {
		return nil, fmt.Errorf("model is required for DeepSeek API")
	}

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
		if config.SystemPrompt != "" {
			messages = append(messages, map[string]interface{}{
				"role":    "system",
				"content": config.SystemPrompt,
			})
		}

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

	// Temperature
	if config.Temperature > 0 {
		request["temperature"] = config.Temperature
	}

	// Max tokens
	// DeepSeek supports max_tokens (not max_completion_tokens)
	maxTokens := config.MaxTokens
	if config.MaxCompletionTokens > 0 {
		maxTokens = config.MaxCompletionTokens
	}
	if maxTokens > 0 {
		request["max_tokens"] = maxTokens
	}

	// Top P
	if config.TopP > 0 {
		request["top_p"] = config.TopP
	}

	// Frequency penalty
	if config.FrequencyPenalty != 0 {
		request["frequency_penalty"] = config.FrequencyPenalty
	}

	// Presence penalty
	if config.PresencePenalty != 0 {
		request["presence_penalty"] = config.PresencePenalty
	}

	// Response format (JSON mode)
	if config.ResponseFormat != nil {
		request["response_format"] = config.ResponseFormat
	}

	// Stream (always false for now)
	request["stream"] = false

	return request, nil
}

// ParseResponse extracts the content from DeepSeek API response
func (h *DeepSeekHandler) ParseResponse(body []byte) (ResponseResult, error) {
	var response struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		Model   string `json:"model"`
		Choices []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens          int `json:"prompt_tokens"`
			CompletionTokens      int `json:"completion_tokens"`
			TotalTokens           int `json:"total_tokens"`
			PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens"`
			PromptCacheMissTokens int `json:"prompt_cache_miss_tokens"`
		} `json:"usage"`
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return ResponseResult{}, fmt.Errorf("failed to parse DeepSeek response: %w", err)
	}

	// Check for API errors
	if response.Error.Message != "" {
		return ResponseResult{}, fmt.Errorf("DeepSeek API error (%s): %s", response.Error.Type, response.Error.Message)
	}

	// Extract content from choices
	if len(response.Choices) == 0 {
		return ResponseResult{}, fmt.Errorf("no choices in response")
	}

	content := response.Choices[0].Message.Content

	result := ResponseResult{
		Content:    content,
		FormatUsed: FormatTypeDeepSeek,
	}

	// DeepSeek doesn't have separate thinking content in standard mode
	// If using reasoning models in the future, this could be extended

	return result, nil
}

// ValidateResponse checks if the response is valid
func (h *DeepSeekHandler) ValidateResponse(statusCode int, body []byte) error {
	var response struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("invalid JSON response: %w", err)
	}

	if response.Error.Message != "" {
		return fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return fmt.Errorf("no choices in response")
	}

	if response.Choices[0].Message.Content == "" {
		return fmt.Errorf("empty content in response")
	}

	return nil
}

// FormatEndpoint formats the API endpoint URL
func (h *DeepSeekHandler) FormatEndpoint(endpoint, model string) string {
	// DeepSeek Chat Completions endpoint
	if endpoint == "" {
		return "https://api.deepseek.com/v1/chat/completions"
	}

	endpoint = strings.TrimSuffix(endpoint, "/")
	parsedURL, err := url.Parse(endpoint)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return endpoint
	}

	switch strings.TrimSuffix(parsedURL.Path, "/") {
	case "":
		parsedURL.Path = "/v1/chat/completions"
		return parsedURL.String()
	case "/v1":
		parsedURL.Path = "/v1/chat/completions"
		return parsedURL.String()
	case "/v1/chat/completions":
		return parsedURL.String()
	default:
		// Custom DeepSeek-compatible gateways may expose their own route.
		return endpoint
	}
}

// GetRequiredHeaders returns the required HTTP headers for DeepSeek API
func (h *DeepSeekHandler) GetRequiredHeaders(apiKey string) map[string]string {
	headers := make(map[string]string)

	// DeepSeek uses standard Authorization header (like OpenAI)
	headers["Authorization"] = "Bearer " + apiKey
	headers["Content-Type"] = "application/json"

	return headers
}

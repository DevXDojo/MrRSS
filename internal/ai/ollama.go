// Package ai provides Ollama API format handlers
package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// OllamaHandler implements FormatHandler for Ollama API
type OllamaHandler struct{}

// NewOllamaHandler creates a new Ollama format handler
func NewOllamaHandler() *OllamaHandler {
	return &OllamaHandler{}
}

// BuildRequest builds an Ollama API request
// Supports both /api/generate (prompt-based) and /api/chat (messages-based)
func (h *OllamaHandler) BuildRequest(config RequestConfig) (map[string]interface{}, error) {
	request := map[string]interface{}{
		"model":  config.Model,
		"stream": false,
	}

	// Check if we should use chat endpoint (when messages are provided)
	useChat := len(config.Messages) > 0

	if useChat {
		// Use /api/chat format (OpenAI-compatible messages)
		request["messages"] = config.Messages
	} else {
		// Use /api/generate format (simple prompt)
		var fullPrompt strings.Builder

		if config.SystemPrompt != "" {
			fullPrompt.WriteString(config.SystemPrompt)
			fullPrompt.WriteString("\n\n")
		}

		if config.UserPrompt != "" {
			fullPrompt.WriteString(config.UserPrompt)
		}

		request["prompt"] = fullPrompt.String()
	}

	// Options for advanced parameters
	options := make(map[string]interface{})

	if config.Temperature > 0 {
		options["temperature"] = config.Temperature
	}
	if config.MaxTokens > 0 {
		options["num_predict"] = config.MaxTokens
	}
	if config.TopP > 0 {
		options["top_p"] = config.TopP
	}
	if config.TopK > 0 {
		options["top_k"] = config.TopK
	}
	if config.Seed > 0 {
		options["seed"] = config.Seed
	}
	if config.PresencePenalty != 0 {
		options["presence_penalty"] = config.PresencePenalty
	}
	if config.FrequencyPenalty != 0 {
		options["frequency_penalty"] = config.FrequencyPenalty
	}

	if len(options) > 0 {
		request["options"] = options
	}

	// Add format for structured outputs (JSON schema)
	if config.ResponseFormat != nil {
		request["format"] = config.ResponseFormat
	}

	return request, nil
}

// ParseResponse parses an Ollama API response
func (h *OllamaHandler) ParseResponse(body []byte) (ResponseResult, error) {
	// Try parsing as chat response first (new format)
	var chatResponse struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Done  bool   `json:"done"`
		Error string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &chatResponse); err == nil && chatResponse.Message.Content != "" {
		// Check for Ollama error
		if chatResponse.Error != "" {
			return ResponseResult{}, fmt.Errorf("ollama API error: %s", chatResponse.Error)
		}

		// Check if response is complete
		if !chatResponse.Done {
			return ResponseResult{}, fmt.Errorf("ollama response incomplete (done=false)")
		}

		content := strings.TrimSpace(chatResponse.Message.Content)
		if content == "" {
			return ResponseResult{}, fmt.Errorf("empty response from Ollama")
		}

		return ResponseResult{
			Content:    content,
			FormatUsed: FormatTypeOllama,
		}, nil
	}

	// Fallback to generate response format (old format)
	var generateResponse struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
		Error    string `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &generateResponse); err != nil {
		return ResponseResult{}, fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	// Check for Ollama error
	if generateResponse.Error != "" {
		return ResponseResult{}, fmt.Errorf("ollama API error: %s", generateResponse.Error)
	}

	// Check if response is complete
	if !generateResponse.Done {
		return ResponseResult{}, fmt.Errorf("ollama response incomplete (done=false)")
	}

	content := strings.TrimSpace(generateResponse.Response)
	if content == "" {
		return ResponseResult{}, fmt.Errorf("empty response from ollama")
	}

	return ResponseResult{
		Content:    content,
		FormatUsed: FormatTypeOllama,
	}, nil
}

// ValidateResponse validates the HTTP response status
func (h *OllamaHandler) ValidateResponse(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("ollama authentication failed")
	case http.StatusNotFound:
		return fmt.Errorf("ollama endpoint or model not found")
	default:
		return fmt.Errorf("ollama API returned status %d: %s", statusCode, string(body))
	}
}

// FormatEndpoint returns the appropriate endpoint based on request type
func (h *OllamaHandler) FormatEndpoint(endpoint, model string) string {
	// Ollama endpoint (user should provide full API path like http://localhost:11434/api/generate or http://localhost:11434/api/chat)
	if endpoint == "" {
		return "http://localhost:11434/api/generate"
	}

	// Use endpoint as-is (user should provide full API path)
	return strings.TrimSuffix(endpoint, "/")
}

// IsOllamaError checks if an error message indicates an Ollama API format
func IsOllamaError(errorMessage string) bool {
	ollamaErrorPatterns := []string{
		"ollama",
		"model not found",
		"pull model",
		"invalid model",
	}

	for _, pattern := range ollamaErrorPatterns {
		if strings.Contains(strings.ToLower(errorMessage), pattern) {
			return true
		}
	}

	return false
}

// Package ai provides Gemini API format handlers
package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// GeminiHandler implements FormatHandler for Gemini API
type GeminiHandler struct{}

// NewGeminiHandler creates a new Gemini format handler
func NewGeminiHandler() *GeminiHandler {
	return &GeminiHandler{}
}

// BuildRequest builds a Gemini API request
func (h *GeminiHandler) BuildRequest(config RequestConfig) (map[string]interface{}, error) {
	contents := []map[string]interface{}{}
	systemParts := []string{}
	if config.SystemPrompt != "" {
		systemParts = append(systemParts, config.SystemPrompt)
	}

	// If messages are provided, convert them to Gemini format
	if len(config.Messages) > 0 {
		for _, msg := range config.Messages {
			role := msg["role"]
			content := msg["content"]

			// Skip empty messages
			if content == "" {
				continue
			}

			if role == "system" {
				systemParts = append(systemParts, content)
				continue
			}

			// Map roles to Gemini format
			geminiRole := "user"
			switch role {
			case "user":
				geminiRole = "user"
			case "assistant":
				geminiRole = "model"
			}

			geminiContent := map[string]interface{}{
				"role": geminiRole,
				"parts": []map[string]string{
					{"text": content},
				},
			}
			contents = append(contents, geminiContent)
		}
	} else {
		// Build from system and user prompts
		// Add user message
		userContent := map[string]interface{}{
			"role": "user",
			"parts": []map[string]string{
				{"text": config.UserPrompt},
			},
		}
		contents = append(contents, userContent)
	}

	// Build request body
	request := map[string]interface{}{
		"contents": contents,
		"generationConfig": map[string]interface{}{
			"temperature":     0.3,
			"maxOutputTokens": 2048,
		},
	}

	// Override defaults if provided
	genConfig := request["generationConfig"].(map[string]interface{})
	if config.Temperature > 0 {
		genConfig["temperature"] = config.Temperature
	}
	if config.MaxTokens > 0 {
		genConfig["maxOutputTokens"] = config.MaxTokens
	}

	// Top-p and Top-k sampling
	if config.TopP > 0 {
		genConfig["topP"] = config.TopP
	}
	if config.TopK > 0 {
		genConfig["topK"] = config.TopK
	}

	// Presence and frequency penalties
	if config.PresencePenalty != 0 {
		genConfig["presencePenalty"] = config.PresencePenalty
	}
	if config.FrequencyPenalty != 0 {
		genConfig["frequencyPenalty"] = config.FrequencyPenalty
	}

	// Seed for reproducible outputs
	if config.Seed > 0 {
		genConfig["seed"] = config.Seed
	}

	// Add system instruction if provided (Gemini-specific)
	// Note: systemInstruction does NOT have a "role" field in Gemini API
	if len(systemParts) > 0 {
		request["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]string{
				{"text": strings.Join(systemParts, "\n\n")},
			},
		}
	}

	// Add thinking config if provided (for thinking models)
	if config.ThinkingConfig != nil {
		genConfig["thinkingConfig"] = config.ThinkingConfig
	}

	return request, nil
}

// ParseResponse parses a Gemini API response
func (h *GeminiHandler) ParseResponse(body []byte) (ResponseResult, error) {
	// First check if this is an error response
	var errorResponse struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &errorResponse); err == nil && errorResponse.Error.Code != 0 {
		return ResponseResult{}, fmt.Errorf("gemini API error (code %d): %s", errorResponse.Error.Code, errorResponse.Error.Message)
	}

	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text    string `json:"text"`
					Thought bool   `json:"thought"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		PromptFeedback struct {
			BlockReason string `json:"blockReason,omitempty"`
		} `json:"promptFeedback"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return ResponseResult{}, fmt.Errorf("failed to decode Gemini response: %w", err)
	}

	// Check if prompt was blocked
	if response.PromptFeedback.BlockReason != "" {
		return ResponseResult{}, fmt.Errorf("prompt blocked: %s", response.PromptFeedback.BlockReason)
	}

	// Check if we have candidates
	if len(response.Candidates) == 0 {
		return ResponseResult{}, fmt.Errorf("no candidates in Gemini response")
	}

	// Get the first candidate's content
	candidate := response.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return ResponseResult{}, fmt.Errorf("no parts in candidate content")
	}

	// Check finish reason
	if candidate.FinishReason == "SAFETY" {
		return ResponseResult{}, fmt.Errorf("response blocked for safety reasons")
	}
	if candidate.FinishReason == "RECITATION" {
		return ResponseResult{}, fmt.Errorf("response blocked for recitation reasons")
	}
	if candidate.FinishReason == "IMAGE_SAFETY" {
		return ResponseResult{}, fmt.Errorf("response blocked for image safety reasons")
	}

	var answer, thinking strings.Builder
	for _, part := range candidate.Content.Parts {
		if part.Thought {
			thinking.WriteString(part.Text)
		} else {
			answer.WriteString(part.Text)
		}
	}
	content := strings.TrimSpace(answer.String())
	if content == "" {
		return ResponseResult{}, fmt.Errorf("empty content in Gemini response")
	}
	return ResponseResult{
		Content:    content,
		Thinking:   strings.TrimSpace(thinking.String()),
		FormatUsed: FormatTypeGemini,
	}, nil
}

// ValidateResponse validates the HTTP response status
func (h *GeminiHandler) ValidateResponse(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("gemini authentication failed (HTTP %d)", statusCode)
	case http.StatusNotFound:
		return fmt.Errorf("gemini model not found (HTTP %d)", statusCode)
	default:
		return fmt.Errorf("gemini API returned status %d: %s", statusCode, string(body))
	}
}

// FormatEndpoint formats the endpoint URL for Gemini API
func (h *GeminiHandler) FormatEndpoint(endpoint, model string) string {
	return FormatGeminiEndpoint(endpoint, model)
}

// IsGeminiEndpoint checks if the given endpoint is a Gemini API endpoint
func IsGeminiEndpoint(endpoint string) bool {
	return DetectAPIProvider(endpoint) == "gemini"
}

// IsGeminiError checks if an error message indicates a Gemini API format mismatch
func IsGeminiError(errorMessage string) bool {
	// Check for common Gemini error messages
	geminiErrorPatterns := []string{
		"Unknown name \"prompt\"",
		"Unknown name \"messages\"",
		"Cannot find field",
		"INVALID_ARGUMENT",
		"generativelanguage.googleapis.com",
	}

	for _, pattern := range geminiErrorPatterns {
		if strings.Contains(errorMessage, pattern) {
			return true
		}
	}

	return false
}

// FormatGeminiEndpoint formats a Gemini endpoint with the given model
// If the endpoint already contains a model, it's replaced
func FormatGeminiEndpoint(baseEndpoint, model string) string {
	if baseEndpoint == "" {
		baseEndpoint = "https://generativelanguage.googleapis.com/v1beta"
	}
	parsed, err := url.Parse(strings.TrimSpace(baseEndpoint))
	if err != nil || parsed.Host == "" {
		return baseEndpoint
	}
	path := strings.TrimRight(parsed.Path, "/")
	model = strings.TrimPrefix(model, "models/")
	switch path {
	case "":
		path = "/v1beta/models/" + model + ":generateContent"
	case "/v1", "/v1beta":
		path += "/models/" + model + ":generateContent"
	default:
		// Preserve gateway prefixes, query options, and custom nonstandard routes.
		if index := strings.LastIndex(path, "/models/"); index >= 0 && strings.HasSuffix(path, ":generateContent") && model != "" {
			path = path[:index] + "/models/" + model + ":generateContent"
		}
	}
	parsed.Path = path
	parsed.RawPath = ""
	return parsed.String()
}

// Package ai provides Gemini API format handlers
package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
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

	// If messages are provided, convert them to Gemini format
	if len(config.Messages) > 0 {
		for _, msg := range config.Messages {
			role := msg["role"]
			content := msg["content"]

			// Skip empty messages
			if content == "" {
				continue
			}

			// Map roles to Gemini format
			geminiRole := "user"
			switch role {
			case "system", "user":
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
	if config.SystemPrompt != "" {
		request["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]string{
				{"text": config.SystemPrompt},
			},
		}
	}

	// Add thinking config if provided (for thinking models)
	if config.ThinkingConfig != nil {
		request["thinkingConfig"] = config.ThinkingConfig
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
					Text string `json:"text"`
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

	content := strings.TrimSpace(candidate.Content.Parts[0].Text)
	return ResponseResult{
		Content:    content,
		FormatUsed: FormatTypeGemini,
	}, nil
}

// ValidateResponse validates the HTTP response status
func (h *GeminiHandler) ValidateResponse(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("gemini authentication failed")
	case http.StatusNotFound:
		return fmt.Errorf("gemini model not found")
	default:
		return fmt.Errorf("gemini API returned status %d: %s", statusCode, string(body))
	}
}

// FormatEndpoint formats the endpoint URL for Gemini API
func (h *GeminiHandler) FormatEndpoint(endpoint, model string) string {
	return FormatGeminiEndpoint(endpoint, model)
}

// DetectAPIProvider detects the AI provider from the endpoint URL
// Returns "gemini", "openai", "anthropic", "deepseek", "ollama", or "unknown"
func DetectAPIProvider(endpoint string) string {
	endpoint = strings.ToLower(endpoint)

	// Gemini API endpoints
	if strings.Contains(endpoint, "googleapis.com") ||
		strings.Contains(endpoint, "generativelanguage.googleapis.com") ||
		strings.Contains(endpoint, "gemini") {
		return "gemini"
	}

	// Anthropic API endpoints
	if strings.Contains(endpoint, "anthropic.com") ||
		strings.Contains(endpoint, "claude") {
		return "anthropic"
	}

	// DeepSeek API endpoints
	if strings.Contains(endpoint, "deepseek.com") ||
		strings.Contains(endpoint, "deepseek") {
		return "deepseek"
	}

	// Ollama endpoints
	if strings.Contains(endpoint, "localhost") ||
		strings.Contains(endpoint, "127.0.0.1") ||
		strings.Contains(endpoint, "ollama") {
		return "ollama"
	}

	// OpenAI-compatible endpoints (default)
	if strings.Contains(endpoint, "openai.com") ||
		strings.Contains(endpoint, "api.openai.com") {
		return "openai"
	}

	return "unknown"
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

// ExtractModelFromEndpoint extracts the model name from a Gemini endpoint
// For example: "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent"
// returns "gemini-pro"
func ExtractModelFromEndpoint(endpoint string) string {
	// Gemini endpoint format: .../models/{model}:generateContent
	re := regexp.MustCompile(`/models/([^:]+)`)
	matches := re.FindStringSubmatch(endpoint)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// FormatGeminiEndpoint formats a Gemini endpoint with the given model
// If the endpoint already contains a model, it's replaced
func FormatGeminiEndpoint(baseEndpoint, model string) string {
	// If endpoint already contains :generateContent or :streamGenerateContent, replace the model
	if strings.Contains(baseEndpoint, ":generateContent") {
		// Use simple pattern to replace model name
		re := regexp.MustCompile(`/models/[^:]+:generateContent`)
		if re.MatchString(baseEndpoint) {
			return re.ReplaceAllString(baseEndpoint, "/models/"+model+":generateContent")
		}
		return baseEndpoint
	}

	// Trim trailing slash
	baseEndpoint = strings.TrimSuffix(baseEndpoint, "/")

	// If model is already in the endpoint, just add the method
	if strings.Contains(baseEndpoint, "/models/"+model) {
		return baseEndpoint + ":generateContent"
	}

	// Check if endpoint ends with "models" (trailing slash was trimmed from .../models/)
	if strings.HasSuffix(baseEndpoint, "models") {
		// Case: .../models (trailing slash was trimmed)
		return baseEndpoint + "/" + model + ":generateContent"
	}

	// If endpoint has /models/ in the middle (not just ending with "models")
	if strings.Contains(baseEndpoint, "/models/") {
		// Find where /models/ ends
		idx := strings.Index(baseEndpoint, "/models/") + len("/models/")
		afterModels := baseEndpoint[idx:]

		// If there's content after /models/ (not empty), replace it
		if afterModels != "" {
			// Remove everything after /models/ and add new model
			return baseEndpoint[:idx] + model + ":generateContent"
		}
		// Just has /models/ at the end (shouldn't happen due to trim, but handle it)
		return baseEndpoint + model + ":generateContent"
	}

	// Add model and method
	return baseEndpoint + "/models/" + model + ":generateContent"
}

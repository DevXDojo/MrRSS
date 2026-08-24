package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"MrRSS/internal/ai"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/utils/textutil"
)

// ChatMessage represents a message in the chat conversation
type ChatMessage struct {
	Role    string `json:"role"` // "system", "user", or "assistant"
	Content string `json:"content"`
}

// ChatRequest represents the incoming chat request
type ChatRequest struct {
	Messages       []ChatMessage `json:"messages"`
	ArticleTitle   string        `json:"article_title,omitempty"`
	ArticleURL     string        `json:"article_url,omitempty"`
	ArticleContent string        `json:"article_content,omitempty"`
	IsFirstMessage bool          `json:"is_first_message,omitempty"`
}

// ChatResponse represents the response from the AI chat
type ChatResponse struct {
	Response string `json:"response"`
	HTML     string `json:"html,omitempty"` // Rendered HTML version of markdown response
}

// HandleAIChat handles chat requests for article discussions
// @Summary      AI chat with article
// @Description  Send messages to AI for discussing article content (requires ai_chat_enabled setting)
// @Tags         chat
// @Accept       json
// @Produce      json
// @Param        request  body      chat.ChatRequest  true  "Chat request (messages, article info)"
// @Success      200  {object}  chat.ChatResponse  "AI response (response, html)"
// @Failure      400  {object}  map[string]string  "Bad request (missing messages)"
// @Failure      403  {object}  map[string]string  "AI chat is disabled or limit reached"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /chat [post]
func HandleAIChat(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		response.Error(w, nil, http.StatusBadRequest)
		return
	}

	// Check if AI chat is enabled
	chatEnabled, _ := h.DB.GetSetting("ai_chat_enabled")
	if chatEnabled != "true" {
		response.Error(w, nil, http.StatusForbidden)
		return
	}

	// Check if AI usage limit is reached
	if h.AITracker.IsLimitReached() {
		log.Printf("AI usage limit reached for chat")
		w.WriteHeader(http.StatusTooManyRequests)
		response.JSON(w, map[string]string{
			"error": "AI usage limit reached",
		})
		return
	}

	// Apply rate limiting for AI requests
	h.AITracker.WaitForRateLimit()

	// Get AI settings - try ProfileProvider first
	var apiKey, endpoint, model string
	if h.AIProfileProvider != nil {
		cfg, err := h.AIProfileProvider.GetConfigForFeature(ai.FeatureChat)
		if err == nil && cfg != nil && (cfg.APIKey != "" || cfg.Endpoint != "") {
			apiKey = cfg.APIKey
			endpoint = cfg.Endpoint
			model = cfg.Model
			log.Printf("Using AI profile for chat (endpoint: %s, model: %s)", endpoint, model)
		}
	}

	// Fallback to global settings if no profile configured
	if endpoint == "" {
		endpoint, _ = h.DB.GetSetting("ai_endpoint")
		model, _ = h.DB.GetSetting("ai_model")
		apiKey, _ = h.DB.GetEncryptedSetting("ai_api_key")

		// Set defaults if still empty
		if endpoint == "" {
			endpoint = "https://api.openai.com/v1/chat/completions"
		}
		if model == "" {
			model = "gpt-4o-mini"
		}
		log.Printf("Using global AI settings for chat (endpoint: %s, model: %s)", endpoint, model)
	}

	// Optimize context to reduce token usage
	optimizedMessages := optimizeChatContext(req.Messages, req.ArticleTitle, req.ArticleURL, req.ArticleContent, req.IsFirstMessage)

	// Convert messages to map format
	messagesMap := make([]map[string]string, len(optimizedMessages))
	for i, msg := range optimizedMessages {
		messagesMap[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	// Create HTTP client with proxy support if configured
	httpClient, err := createHTTPClientWithProxy(h)
	if err != nil {
		log.Printf("Failed to create HTTP client with proxy: %v", err)
		httpClient = &http.Client{Timeout: 60 * time.Second}
	} else {
		httpClient.Timeout = 60 * time.Second
	}

	// Create AI client
	clientConfig := ai.ClientConfig{
		APIKey:   apiKey,
		Endpoint: endpoint,
		Model:    model,
		Timeout:  60 * time.Second,
	}
	client := ai.NewClientWithHTTPClient(clientConfig, httpClient)

	// Send chat request using universal client
	result, err := client.RequestWithMessages(messagesMap)
	if err != nil {
		log.Printf("AI chat request failed: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Extract thinking content and remove tags
	respContent := result.Content
	thinking := ai.ExtractThinking(respContent)
	respContent = ai.RemoveThinkingTags(respContent)

	// Convert markdown response to HTML
	htmlResponse := textutil.ConvertMarkdownToHTML(respContent)

	// Log thinking if present (for debugging)
	if thinking != "" {
		log.Printf("AI chat thinking: %s", thinking)
	}

	// Track AI usage (estimate tokens from input and output)
	estimatedTokens := estimateChatTokens(optimizedMessages, respContent)
	if err := h.AITracker.AddUsage(int64(estimatedTokens)); err != nil {
		log.Printf("Warning: failed to track AI usage: %v", err)
	}

	// Track statistics
	_ = h.DB.IncrementStat("ai_chat")

	response.JSON(w, ChatResponse{Response: respContent, HTML: htmlResponse})
}

// optimizeChatContext reduces the chat context to save tokens while preserving important information
func optimizeChatContext(messages []ChatMessage, articleTitle, articleURL, articleContent string, isFirstMessage bool) []ChatMessage {
	// If this is the first message, include article content
	if isFirstMessage && articleContent != "" {
		// Add article context as a system message
		systemMsg := ChatMessage{
			Role: "system",
			Content: fmt.Sprintf("You are discussing an article titled: %s\nURL: %s\n\nArticle content:\n%s\n\nPlease help the user understand and discuss this article.",
				articleTitle, articleURL, articleContent),
		}
		return append([]ChatMessage{systemMsg}, messages...)
	}

	// For subsequent messages, only keep recent conversation history
	const maxHistoryLength = 10
	if len(messages) <= maxHistoryLength {
		return messages
	}

	// Keep only the most recent messages
	return messages[len(messages)-maxHistoryLength:]
}

// estimateChatTokens estimates the number of tokens used for a chat request/response
func estimateChatTokens(messages []ChatMessage, response string) int {
	// Rough estimation: ~4 characters per token
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Content)
	}
	totalChars += len(response)

	// Add some overhead for JSON formatting and API overhead
	totalChars = int(float64(totalChars) * 1.2)

	// Estimate tokens (roughly 4 characters per token for English)
	return totalChars / 4
}

// createHTTPClientWithProxy creates an HTTP client with global proxy settings if enabled
func createHTTPClientWithProxy(h *core.Handler) (*http.Client, error) {
	// Check if global proxy is enabled
	proxyEnabled, _ := h.DB.GetSetting("proxy_enabled")
	if proxyEnabled != "true" {
		return &http.Client{Timeout: 60 * time.Second}, nil
	}

	// Build proxy URL from global settings
	proxyType, _ := h.DB.GetSetting("proxy_type")
	proxyHost, _ := h.DB.GetSetting("proxy_host")
	proxyPort, _ := h.DB.GetSetting("proxy_port")
	proxyUsername, _ := h.DB.GetEncryptedSetting("proxy_username")
	proxyPassword, _ := h.DB.GetEncryptedSetting("proxy_password")

	// Build proxy URL
	proxyURL := buildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)

	// Create HTTP client with proxy
	return createHTTPClient(proxyURL, 60*time.Second)
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

// createHTTPClient creates an HTTP client with optional proxy
func createHTTPClient(proxyURL string, timeout time.Duration) (*http.Client, error) {
	client := &http.Client{Timeout: timeout}

	if proxyURL != "" {
		u, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(u),
		}
	}

	return client, nil
}

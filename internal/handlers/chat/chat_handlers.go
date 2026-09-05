package chat

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"MrRSS/internal/ai"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/utils/httputil"
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
	SessionID      int64         `json:"session_id,omitempty"`
	ArticleID      int64         `json:"article_id,omitempty"`
	ArticleTitle   string        `json:"article_title,omitempty"`
	ArticleURL     string        `json:"article_url,omitempty"`
	ArticleContent string        `json:"article_content,omitempty"`
	IsFirstMessage bool          `json:"is_first_message,omitempty"`
}

// ChatResponse represents the response from the AI chat
type ChatResponse struct {
	Response     string `json:"response"`
	HTML         string `json:"html,omitempty"` // Rendered HTML version of markdown response
	Thinking     string `json:"thinking,omitempty"`
	SessionID    int64  `json:"session_id,omitempty"`
	HistorySaved bool   `json:"history_saved"`
}

type chatErrorResponse struct {
	Error     string `json:"error"`
	SessionID int64  `json:"session_id,omitempty"`
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

	sessionID, historyEnabled, err := persistUserChatMessage(h, &req)
	if err != nil {
		log.Printf("AI chat history preparation failed session=%d", req.SessionID)
		writeChatError(w, "Failed to save chat history", http.StatusInternalServerError, sessionID)
		return
	}

	// Check if AI usage limit is reached
	if h.AITracker.IsLimitReached() {
		log.Printf("AI usage limit reached for chat")
		writeChatError(w, "AI usage limit reached", http.StatusTooManyRequests, sessionID)
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
			log.Printf("Using AI profile for chat (model: %s)", model)
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
		log.Printf("Using global AI settings for chat (model: %s)", model)
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
		log.Printf("Failed to create HTTP client with proxy")
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
		log.Printf("AI chat request failed")
		writeChatError(w, "Failed to get response from AI. Please try again.", http.StatusInternalServerError, sessionID)
		return
	}

	// Extract thinking content and remove tags
	respContent := result.Content
	thinking := ai.ExtractThinking(respContent)
	respContent = ai.RemoveThinkingTags(respContent)

	// Convert markdown response to HTML
	htmlResponse := textutil.ConvertMarkdownToHTML(respContent)

	// Track AI usage (estimate tokens from input and output)
	estimatedTokens := estimateChatTokens(optimizedMessages, respContent)
	if err := h.AITracker.AddUsage(int64(estimatedTokens)); err != nil {
		log.Printf("Warning: failed to track AI usage: %v", err)
	}

	// Track statistics
	_ = h.DB.IncrementStat("ai_chat")

	historySaved := historyEnabled
	if historyEnabled {
		if _, saveErr := h.DB.CreateChatMessage(sessionID, "assistant", respContent, thinking); saveErr != nil {
			log.Printf("AI chat assistant history save failed session=%d", sessionID)
			historySaved = false
		}
	}

	response.JSON(w, ChatResponse{
		Response: respContent, HTML: htmlResponse, Thinking: thinking,
		SessionID: sessionID, HistorySaved: historySaved,
	})
}

func persistUserChatMessage(h *core.Handler, req *ChatRequest) (int64, bool, error) {
	if req.ArticleID <= 0 {
		// Backward compatibility for callers that do not send article_id.
		return req.SessionID, false, nil
	}

	lastUserMessage := ""
	for index := len(req.Messages) - 1; index >= 0; index-- {
		if req.Messages[index].Role == "user" {
			lastUserMessage = strings.TrimSpace(req.Messages[index].Content)
			break
		}
	}
	if lastUserMessage == "" {
		return req.SessionID, false, fmt.Errorf("latest user message is missing")
	}

	sessionID := req.SessionID
	if sessionID > 0 {
		session, err := h.DB.GetChatSession(sessionID)
		if err != nil {
			return sessionID, false, err
		}
		if session == nil || session.ArticleID != req.ArticleID {
			return sessionID, false, fmt.Errorf("chat session does not belong to the article")
		}
	} else {
		title := []rune(lastUserMessage)
		if len(title) > 60 {
			title = title[:60]
		}
		var err error
		sessionID, err = h.DB.CreateChatSession(req.ArticleID, string(title))
		if err != nil {
			return 0, false, err
		}
	}

	if _, err := h.DB.CreateChatMessage(sessionID, "user", lastUserMessage, ""); err != nil {
		return sessionID, false, err
	}
	return sessionID, true, nil
}

func writeChatError(w http.ResponseWriter, message string, status int, sessionID int64) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(chatErrorResponse{Error: message, SessionID: sessionID})
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

// createHTTPClientWithProxy creates the canonical HTTP client with global proxy settings.
func createHTTPClientWithProxy(h *core.Handler) (*http.Client, error) {
	return httputil.CreateHTTPClientWithProxySettings(h.DB, 60*time.Second)
}

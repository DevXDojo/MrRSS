package summary

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"MrRSS/internal/ai"
	"MrRSS/internal/config"
	"MrRSS/internal/utils/httputil"
)

// AISummarizer implements summarization using OpenAI-compatible APIs (GPT, Claude, etc.).
type AISummarizer struct {
	APIKey        string
	Endpoint      string
	Model         string
	SystemPrompt  string
	CustomHeaders string
	Language      string // User's language setting (e.g., "en", "zh")
	client        *ai.Client
}

// DBInterface defines the minimal database interface needed for proxy settings
type DBInterface interface {
	GetSetting(key string) (string, error)
	GetEncryptedSetting(key string) (string, error)
}

// CreateHTTPClientWithProxy creates an HTTP client with global proxy settings if enabled
func CreateHTTPClientWithProxy(db DBInterface, timeout time.Duration) (*http.Client, error) {
	var proxyURL string

	// Check if global proxy is enabled
	proxyEnabled, _ := db.GetSetting("proxy_enabled")
	if proxyEnabled == "true" {
		// Build proxy URL from global settings
		proxyType, _ := db.GetSetting("proxy_type")
		proxyHost, _ := db.GetSetting("proxy_host")
		proxyPort, _ := db.GetSetting("proxy_port")
		proxyUsername, _ := db.GetEncryptedSetting("proxy_username")
		proxyPassword, _ := db.GetEncryptedSetting("proxy_password")
		proxyURL = httputil.BuildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)
	}

	// Create HTTP client with or without proxy
	return httputil.CreateHTTPClient(proxyURL, timeout)
}

// NewAISummarizer creates a new AI summarizer with the given credentials.
// endpoint should be the full API URL (e.g., "https://api.openai.com/v1/chat/completions" for OpenAI, "http://localhost:11434/api/generate" for Ollama)
// model should be the model name (e.g., "gpt-4o-mini", "claude-3-haiku-20240307")
// Uses global AI settings shared between translation and summarization.
func NewAISummarizer(apiKey, endpoint, model string) *AISummarizer {
	defaults := config.Get()
	// Use global AI endpoint and model
	if endpoint == "" {
		endpoint = defaults.AIEndpoint
	}
	if model == "" {
		model = defaults.AIModel
	}

	clientConfig := ai.ClientConfig{
		APIKey:   apiKey,
		Endpoint: strings.TrimSuffix(endpoint, "/"),
		Model:    model,
		Timeout:  30 * time.Second,
	}

	return &AISummarizer{
		APIKey:        apiKey,
		Endpoint:      strings.TrimSuffix(endpoint, "/"),
		Model:         model,
		SystemPrompt:  "",   // Will be set from settings when used
		CustomHeaders: "",   // Will be set from settings when used
		Language:      "en", // Default to English
		client:        ai.NewClient(clientConfig),
	}
}

// NewAISummarizerWithDB creates a new AI summarizer with database for proxy support
func NewAISummarizerWithDB(apiKey, endpoint, model string, db DBInterface) *AISummarizer {
	defaults := config.Get()
	if endpoint == "" {
		endpoint = defaults.AIEndpoint
	}
	if model == "" {
		model = defaults.AIModel
	}

	httpClient, err := CreateHTTPClientWithProxy(db, 30*time.Second)
	if err != nil {
		// Fallback to default client if proxy creation fails
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	clientConfig := ai.ClientConfig{
		APIKey:   apiKey,
		Endpoint: strings.TrimSuffix(endpoint, "/"),
		Model:    model,
		Timeout:  30 * time.Second,
	}

	return &AISummarizer{
		APIKey:        apiKey,
		Endpoint:      strings.TrimSuffix(endpoint, "/"),
		Model:         model,
		SystemPrompt:  "",
		CustomHeaders: "",   // Will be set from settings when used
		Language:      "en", // Default to English
		client:        ai.NewClientWithHTTPClient(clientConfig, httpClient),
	}
}

// SetSystemPrompt sets a custom system prompt for the summarizer.
func (s *AISummarizer) SetSystemPrompt(prompt string) {
	s.SystemPrompt = prompt
	// Re-create client with updated system prompt
	s.recreateClient()
}

// SetCustomHeaders sets custom headers for AI requests.
func (s *AISummarizer) SetCustomHeaders(headers string) {
	s.CustomHeaders = headers
	// Re-create client with updated custom headers
	s.recreateClient()
}

// SetLanguage sets the language for the summarizer.
// If language is empty, it keeps the current language setting.
func (s *AISummarizer) SetLanguage(language string) {
	if language != "" {
		s.Language = language
	}
}

// recreateClient re-creates the AI client with current configuration
func (s *AISummarizer) recreateClient() {
	clientConfig := ai.ClientConfig{
		APIKey:        s.APIKey,
		Endpoint:      s.Endpoint,
		Model:         s.Model,
		SystemPrompt:  s.SystemPrompt,
		CustomHeaders: s.CustomHeaders,
		Timeout:       30 * time.Second,
	}
	s.client = ai.NewClient(clientConfig)
}

// getDefaultSystemPrompt returns the default system prompt based on the configured language.
func (s *AISummarizer) getDefaultSystemPrompt() string {
	// Check if language starts with "zh" to handle locale codes like "zh", "zh-CN", "zh-TW", etc.
	if strings.HasPrefix(s.Language, "zh") {
		return "你是一个专业的文章摘要助手。请为给定的文章生成清晰、格式良好的摘要。在列出项目、特性或要点时，请优先使用项目符号或编号列表来组织内容。使摘要易于阅读和浏览。"
	}
	return "You are a helpful AI assistant that creates clear, well-formatted summaries. When listing items, features, or points, prefer using bullet points or numbered lists to organize the content. Make the summary scannable and easy to read."
}

// getUserPrompt generates a localized user prompt with target language specification.
func (s *AISummarizer) getUserPrompt(targetWords int, text string) string {
	// Check if language starts with "zh" to handle locale codes like "zh", "zh-CN", "zh-TW", etc.
	if strings.HasPrefix(s.Language, "zh") {
		return fmt.Sprintf("请用中文将以下内容总结为大约 %d 字：\n\n%s", targetWords, text)
	}
	return fmt.Sprintf("Summarize the following text in English in approximately %d words:\n\n%s", targetWords, text)
}

// Summarize generates a summary of the given text using an OpenAI-compatible API.
// Automatically detects and adapts to different API formats (Gemini, OpenAI, Ollama).
func (s *AISummarizer) Summarize(text string, length SummaryLength) (SummaryResult, error) {
	// Clean the text first
	cleanedText := cleanText(text)

	// Check if text is too short
	if len(cleanedText) < MinContentLength {
		return SummaryResult{
			Summary:    cleanedText,
			IsTooShort: true,
		}, nil
	}

	targetWords := getTargetWordCount(length)

	// Use custom system prompt if provided, otherwise use default
	systemPrompt := s.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = s.getDefaultSystemPrompt()
	}

	// Generate localized user prompt with target language specification
	userPrompt := s.getUserPrompt(targetWords, cleanedText)

	// Use the universal client which handles format detection automatically
	result, err := s.client.RequestWithThinking(systemPrompt, userPrompt)
	if err != nil {
		return SummaryResult{}, err
	}

	// Extract thinking content using shared utility
	thinking := ai.ExtractThinking(result.Content)
	summary := ai.RemoveThinkingTags(result.Content)

	// Count sentences in the summary
	sentences := splitSentences(summary)

	return SummaryResult{
		Summary:       summary,
		Thinking:      thinking,
		SentenceCount: len(sentences),
		IsTooShort:    false,
	}, nil
}

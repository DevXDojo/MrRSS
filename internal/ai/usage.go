// Package ai provides AI usage tracking and rate limiting functionality.
package ai

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

// SettingsProvider is an interface for retrieving and storing settings.
type SettingsProvider interface {
	GetSetting(key string) (string, error)
	SetSetting(key, value string) error
}

// UsageTracker tracks AI usage (tokens) and enforces rate limits.
type UsageTracker struct {
	settings    SettingsProvider
	mu          sync.RWMutex
	lastRequest time.Time
	minInterval time.Duration // Minimum interval between AI requests
}

// NewUsageTracker creates a new AI usage tracker.
func NewUsageTracker(settings SettingsProvider) *UsageTracker {
	return &UsageTracker{
		settings:    settings,
		minInterval: 500 * time.Millisecond, // Default: max 2 requests per second
	}
}

// SetMinInterval sets the minimum interval between AI requests.
func (t *UsageTracker) SetMinInterval(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.minInterval = d
}

// CanMakeRequest checks if a new AI request can be made (rate limiting).
// Returns true if allowed, false if rate limited.
func (t *UsageTracker) CanMakeRequest() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if now.Sub(t.lastRequest) < t.minInterval {
		return false
	}
	t.lastRequest = now
	return true
}

// WaitForRateLimit blocks until a request can be made.
func (t *UsageTracker) WaitForRateLimit() {
	t.mu.Lock()
	elapsed := time.Since(t.lastRequest)
	wait := t.minInterval - elapsed
	t.mu.Unlock()

	if wait > 0 {
		time.Sleep(wait)
	}

	t.mu.Lock()
	t.lastRequest = time.Now()
	t.mu.Unlock()
}

// GetCurrentUsage returns the current token usage.
func (t *UsageTracker) GetCurrentUsage() (int64, error) {
	usageStr, err := t.settings.GetSetting("ai_usage_tokens")
	if err != nil {
		return 0, err
	}
	if usageStr == "" {
		return 0, nil
	}
	return strconv.ParseInt(usageStr, 10, 64)
}

// GetUsageLimit returns the configured usage limit (0 = unlimited).
func (t *UsageTracker) GetUsageLimit() (int64, error) {
	limitStr, err := t.settings.GetSetting("ai_usage_limit")
	if err != nil {
		return 0, err
	}
	if limitStr == "" {
		return 0, nil
	}
	return strconv.ParseInt(limitStr, 10, 64)
}

// IsLimitReached checks if the usage limit has been reached.
func (t *UsageTracker) IsLimitReached() bool {
	usage, err := t.GetCurrentUsage()
	if err != nil {
		return false
	}

	limit, err := t.GetUsageLimit()
	if err != nil {
		return false
	}

	// 0 means unlimited
	if limit == 0 {
		return false
	}

	return usage >= limit
}

// AddUsage adds tokens to the usage counter.
func (t *UsageTracker) AddUsage(tokens int64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Get current usage inside the lock to prevent race condition
	usageStr, err := t.settings.GetSetting("ai_usage_tokens")
	var current int64
	if err == nil && usageStr != "" {
		current, _ = strconv.ParseInt(usageStr, 10, 64)
	}

	newUsage := current + tokens
	return t.settings.SetSetting("ai_usage_tokens", strconv.FormatInt(newUsage, 10))
}

// ResetUsage resets the usage counter to zero.
func (t *UsageTracker) ResetUsage() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.settings.SetSetting("ai_usage_tokens", "0")
}

// TrackTranslation tracks token usage for a translation operation.
func (t *UsageTracker) TrackTranslation(sourceText, translatedText string) {
	// Estimate tokens for both input and output
	inputTokens := EstimateTokens(sourceText)
	outputTokens := EstimateTokens(translatedText)

	totalTokens := inputTokens + outputTokens
	if err := t.AddUsage(totalTokens); err != nil {
		log.Printf("Warning: failed to track AI usage: %v", err)
	}
}

// TrackSummary tracks token usage for a summarization operation.
func (t *UsageTracker) TrackSummary(content, summary string) {
	// Estimate tokens for both input and output
	inputTokens := EstimateTokens(content)
	outputTokens := EstimateTokens(summary)

	totalTokens := inputTokens + outputTokens
	if err := t.AddUsage(totalTokens); err != nil {
		log.Printf("Warning: failed to track AI usage: %v", err)
	}
}

// EstimateTokens estimates the number of tokens in a text.
// Uses a simple heuristic: ~4 characters per token for English, ~1.5 characters per token for CJK.
func EstimateTokens(text string) int64 {
	if text == "" {
		return 0
	}

	// Count CJK characters
	cjkCount := 0
	nonCJKCount := 0

	for _, r := range text {
		if isCJK(r) {
			cjkCount++
		} else if r > 32 { // Non-whitespace, non-control characters
			nonCJKCount++
		}
	}

	// Rough estimation:
	// - CJK: roughly 1 token per 1.5 characters
	// - Non-CJK: roughly 1 token per 4 characters (words average ~4-5 chars + spaces)
	cjkTokens := float64(cjkCount) / 1.5
	nonCJKTokens := float64(nonCJKCount) / 4.0

	// Add some overhead for special tokens
	total := int64(cjkTokens + nonCJKTokens + 10)
	if total < 1 {
		total = 1
	}

	return total
}

// EstimateTokensWithSegmentation estimates tokens using word-level segmentation.
// For CJK text, counts words/characters. For other text, counts space-separated words.
func EstimateTokensWithSegmentation(text string) int64 {
	if text == "" {
		return 0
	}

	tokens := int64(0)

	// Split by whitespace to handle mixed content
	words := strings.Fields(text)

	for _, word := range words {
		// Check if word contains CJK characters
		hasCJK := false
		for _, r := range word {
			if isCJK(r) {
				hasCJK = true
				break
			}
		}

		if hasCJK {
			// For CJK, count characters as rough token estimate
			// (In reality, subword tokenization is more complex)
			for _, r := range word {
				if isCJK(r) {
					tokens++
				}
			}
		} else {
			// For non-CJK, each word is roughly 1-2 tokens
			wordLen := len(word)
			if wordLen <= 4 {
				tokens++
			} else if wordLen <= 8 {
				tokens += 2
			} else {
				tokens += int64(wordLen/4) + 1
			}
		}
	}

	// Minimum 1 token
	if tokens < 1 {
		tokens = 1
	}

	return tokens
}

// isCJK checks if a rune is a CJK (Chinese, Japanese, Korean) character.
func isCJK(r rune) bool {
	// CJK Unified Ideographs
	if r >= 0x4E00 && r <= 0x9FFF {
		return true
	}
	// CJK Unified Ideographs Extension A
	if r >= 0x3400 && r <= 0x4DBF {
		return true
	}
	// CJK Unified Ideographs Extension B
	if r >= 0x20000 && r <= 0x2A6DF {
		return true
	}
	// Hiragana
	if r >= 0x3040 && r <= 0x309F {
		return true
	}
	// Katakana
	if r >= 0x30A0 && r <= 0x30FF {
		return true
	}
	// Hangul Syllables
	if r >= 0xAC00 && r <= 0xD7AF {
		return true
	}
	return false
}

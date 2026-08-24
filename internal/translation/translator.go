package translation

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"MrRSS/internal/utils/httputil"
)

// Translator defines the interface for translation services
type Translator interface {
	Translate(text, targetLang string) (string, error)
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
	if db != nil {
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
	}

	// Create HTTP client with or without proxy
	return httputil.CreateHTTPClient(proxyURL, timeout)
}

// MockTranslator is a simple translator for demonstration
type MockTranslator struct{}

func NewMockTranslator() *MockTranslator {
	return &MockTranslator{}
}

func (t *MockTranslator) Translate(text, targetLang string) (string, error) {
	// In a real application, this would call an external API (Google, DeepL, etc.)
	// For now, we simulate translation by appending the language code.
	// We can also do some simple word replacements to make it look "translated"

	prefix := fmt.Sprintf("[%s] ", strings.ToUpper(targetLang))
	if strings.HasPrefix(text, prefix) {
		return text, nil
	}

	return prefix + text, nil
}

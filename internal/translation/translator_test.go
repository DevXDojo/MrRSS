package translation

import (
	"net/http"
	"testing"
	"time"

	"MrRSS/internal/config"
)

func TestMockTranslator(t *testing.T) {
	translator := NewMockTranslator()
	text := "Hello"
	targetLang := "es"

	translated, err := translator.Translate(text, targetLang)
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}

	expected := "[ES] Hello"
	if translated != expected {
		t.Errorf("Expected '%s', got '%s'", expected, translated)
	}

	// Test idempotency (mock implementation detail)
	translated2, err := translator.Translate(translated, targetLang)
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}
	if translated2 != expected {
		t.Errorf("Expected '%s', got '%s'", expected, translated2)
	}
}

func TestBaiduTranslator_EmptyText(t *testing.T) {
	translator := NewBaiduTranslator("app_id", "secret_key")
	result, err := translator.Translate("", "zh")
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

func TestBaiduLangMapping(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"en", "en"},
		{"zh", "zh"},
		{"ja", "jp"},
		{"es", "spa"},
		{"fr", "fra"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		result := mapToBaiduLang(tt.input)
		if result != tt.expected {
			t.Errorf("mapToBaiduLang(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}

func TestAITranslator_EmptyText(t *testing.T) {
	translator := NewAITranslator("api_key", "", "")
	result, err := translator.Translate("", "zh")
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

func TestAITranslator_Defaults(t *testing.T) {
	defaults := config.Get()
	translator := NewAITranslator("api_key", "", "")
	if translator.Endpoint != defaults.AIEndpoint {
		t.Errorf("Expected default endpoint %s, got '%s'", defaults.AIEndpoint, translator.Endpoint)
	}
	if translator.Model != defaults.AIModel {
		t.Errorf("Expected default model %s, got '%s'", defaults.AIModel, translator.Model)
	}
}

func TestGetLanguageName(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"en", "English"},
		{"zh", "Simplified Chinese"},
		{"zh-TW", "Traditional Chinese"},
		{"ja", "Japanese"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		result := getLanguageName(tt.code)
		if result != tt.expected {
			t.Errorf("getLanguageName(%s) = %s, want %s", tt.code, result, tt.expected)
		}
	}
}

// mockSettingsProvider implements SettingsProvider for testing
type mockSettingsProvider struct {
	settings map[string]string
}

func (m *mockSettingsProvider) GetSetting(key string) (string, error) {
	return m.settings[key], nil
}

func (m *mockSettingsProvider) GetEncryptedSetting(key string) (string, error) {
	// In tests, just return the plain value (mock doesn't encrypt)
	return m.settings[key], nil
}

func TestDynamicTranslator_DefaultsToGoogle(t *testing.T) {
	provider := &mockSettingsProvider{
		settings: map[string]string{},
	}
	translator := NewDynamicTranslator(provider)

	// Should return empty for empty text without error
	result, err := translator.Translate("", "zh")
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}
}

func TestDynamicTranslator_RequiresCredentials(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		settings map[string]string
		wantErr  bool
	}{
		{
			name:     "deepl_no_key",
			provider: "deepl",
			settings: map[string]string{"translation_provider": "deepl"},
			wantErr:  true, // No API key
		},
		{
			name:     "baidu_no_credentials",
			provider: "baidu",
			settings: map[string]string{"translation_provider": "baidu"},
			wantErr:  true, // No app ID or secret key
		},
		{
			name:     "baidu_partial",
			provider: "baidu",
			settings: map[string]string{"translation_provider": "baidu", "baidu_app_id": "id"},
			wantErr:  true, // No secret key
		},
		{
			name:     "ai_no_key",
			provider: "ai",
			settings: map[string]string{"translation_provider": "ai"},
			wantErr:  true, // No API key
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &mockSettingsProvider{settings: tt.settings}
			translator := NewDynamicTranslator(provider)
			_, err := translator.Translate("Hello", "zh")
			if (err != nil) != tt.wantErr {
				t.Errorf("Translate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateHTTPClientWithProxy_EnabledAndDisabled(t *testing.T) {
	// Disabled
	provider1 := &mockSettingsProvider{settings: map[string]string{"proxy_enabled": "false"}}
	c1, err := CreateHTTPClientWithProxy(provider1, 1*time.Second)
	if err != nil {
		t.Fatalf("CreateHTTPClientWithProxy error: %v", err)
	}
	tr1, ok := c1.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport type")
	}
	if tr1.Proxy != nil {
		t.Fatalf("expected no proxy when disabled")
	}

	// Enabled
	provider2 := &mockSettingsProvider{settings: map[string]string{
		"proxy_enabled":  "true",
		"proxy_type":     "http",
		"proxy_host":     "127.0.0.1",
		"proxy_port":     "3128",
		"proxy_username": "u",
		"proxy_password": "p",
	}}
	c2, err := CreateHTTPClientWithProxy(provider2, 1*time.Second)
	if err != nil {
		t.Fatalf("CreateHTTPClientWithProxy error: %v", err)
	}
	tr2, ok := c2.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport type")
	}
	if tr2.Proxy == nil {
		t.Fatalf("expected proxy to be configured when enabled")
	}
}

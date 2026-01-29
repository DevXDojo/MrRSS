package translation

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
)

// SettingsProvider is an interface for retrieving translation settings.
type SettingsProvider interface {
	GetSetting(key string) (string, error)
	GetEncryptedSetting(key string) (string, error)
}

// CacheProvider is an interface for translation caching
type CacheProvider interface {
	GetCachedTranslation(sourceTextHash, targetLang, provider string) (string, bool, error)
	SetCachedTranslation(sourceTextHash, sourceText, targetLang, translatedText, provider string) error
}

// DynamicTranslator is a translator that dynamically selects the translation provider
// based on user settings. It creates the appropriate translator at translation time.
type DynamicTranslator struct {
	settings SettingsProvider
	cache    CacheProvider
	mu       sync.RWMutex
	// Cache the current translator to avoid recreating it for every translation
	cachedTranslator    Translator
	cachedProvider      string
	cachedAPIKey        string
	cachedAppID         string
	cachedSecretKey     string
	cachedEndpoint      string
	cachedModel         string
	cachedPrompt        string
	cachedCustomHeaders string
	cachedBodyTemplate  string
	cachedResponsePath  string
	cachedLangMapping   string
	cachedTimeout       int
}

// NewDynamicTranslator creates a new dynamic translator that uses the given settings provider.
func NewDynamicTranslator(settings SettingsProvider) *DynamicTranslator {
	return &DynamicTranslator{
		settings: settings,
	}
}

// NewDynamicTranslatorWithCache creates a new dynamic translator with translation caching.
func NewDynamicTranslatorWithCache(settings SettingsProvider, cache CacheProvider) *DynamicTranslator {
	return &DynamicTranslator{
		settings: settings,
		cache:    cache,
	}
}

// Translate translates text using the currently configured translation provider.
func (t *DynamicTranslator) Translate(text, targetLang string) (string, error) {
	if text == "" {
		return "", nil
	}

	translator, provider, err := t.getTranslatorWithProvider()
	if err != nil {
		return "", err
	}

	// Wrap with caching if cache is available
	if t.cache != nil {
		cachedTranslator := NewCachedTranslator(translator, t.cache, provider)
		return cachedTranslator.Translate(text, targetLang)
	}

	return translator.Translate(text, targetLang)
}

// getTranslatorWithProvider returns the appropriate translator and provider name based on current settings.
// It caches the translator and only recreates it if settings have changed.
func (t *DynamicTranslator) getTranslatorWithProvider() (Translator, string, error) {
	provider, _ := t.settings.GetSetting("translation_provider")
	if provider == "" {
		provider = "google" // Default to Google Free
	}

	// Get provider-specific settings (use encrypted methods for sensitive credentials)
	var apiKey, appID, secretKey, endpoint, model, systemPrompt, customHeaders string
	var customName, customMethod, customBodyTemplate, customResponsePath, customLangMapping string
	var customTimeout int
	switch provider {
	case "deepl":
		apiKey, _ = t.settings.GetEncryptedSetting("deepl_api_key")
		endpoint, _ = t.settings.GetSetting("deepl_endpoint")
	case "baidu":
		appID, _ = t.settings.GetSetting("baidu_app_id")
		secretKey, _ = t.settings.GetEncryptedSetting("baidu_secret_key")
	case "ai":
		apiKey, _ = t.settings.GetEncryptedSetting("ai_api_key")
		endpoint, _ = t.settings.GetSetting("ai_endpoint")
		model, _ = t.settings.GetSetting("ai_model")
		systemPrompt, _ = t.settings.GetSetting("ai_translation_prompt")
		customHeaders, _ = t.settings.GetSetting("ai_custom_headers")
	case "custom":
		customName, _ = t.settings.GetSetting("custom_translation_name")
		endpoint, _ = t.settings.GetSetting("custom_translation_endpoint")
		customMethod, _ = t.settings.GetSetting("custom_translation_method")
		customHeaders, _ = t.settings.GetSetting("custom_translation_headers")
		customBodyTemplate, _ = t.settings.GetSetting("custom_translation_body_template")
		customResponsePath, _ = t.settings.GetSetting("custom_translation_response_path")
		customLangMapping, _ = t.settings.GetSetting("custom_translation_lang_mapping")
		timeoutStr, _ := t.settings.GetSetting("custom_translation_timeout")
		if timeoutStr != "" {
			fmt.Sscanf(timeoutStr, "%d", &customTimeout)
		}
	}

	// Check if we can reuse the cached translator
	t.mu.RLock()
	if t.cachedTranslator != nil &&
		t.cachedProvider == provider &&
		t.cachedAPIKey == apiKey &&
		t.cachedAppID == appID &&
		t.cachedSecretKey == secretKey &&
		t.cachedEndpoint == endpoint &&
		t.cachedModel == model &&
		t.cachedPrompt == systemPrompt &&
		t.cachedCustomHeaders == customHeaders &&
		t.cachedBodyTemplate == customBodyTemplate &&
		t.cachedResponsePath == customResponsePath &&
		t.cachedLangMapping == customLangMapping &&
		t.cachedTimeout == customTimeout {
		translator := t.cachedTranslator
		t.mu.RUnlock()
		return translator, provider, nil
	}
	t.mu.RUnlock()

	// Create new translator
	t.mu.Lock()
	defer t.mu.Unlock()

	var translator Translator
	switch provider {
	case "google":
		translator = NewGoogleFreeTranslator()
	case "deepl":
		// For deeplx self-hosted, endpoint is required but API key is optional
		if endpoint == "" && apiKey == "" {
			return nil, "", fmt.Errorf("deepL API key is required (or provide a custom endpoint for deeplx)")
		}
		if endpoint != "" {
			translator = NewDeepLTranslatorWithEndpoint(apiKey, endpoint)
		} else {
			translator = NewDeepLTranslator(apiKey)
		}
	case "baidu":
		if appID == "" || secretKey == "" {
			return nil, "", fmt.Errorf("baidu App ID and Secret Key are required")
		}
		translator = NewBaiduTranslator(appID, secretKey)
	case "ai":
		// Allow empty API key for local endpoints (e.g., Ollama)
		if apiKey == "" && !isLocalEndpoint(endpoint) {
			return nil, "", fmt.Errorf("ai API key is required for non-local endpoints")
		}
		aiTranslator := NewAITranslator(apiKey, endpoint, model)
		if systemPrompt != "" {
			aiTranslator.SetSystemPrompt(systemPrompt)
		}
		if customHeaders != "" {
			aiTranslator.SetCustomHeaders(customHeaders)
		}
		translator = aiTranslator
	case "custom":
		// Custom translator with user-defined configuration
		if endpoint == "" {
			return nil, "", fmt.Errorf("custom translation endpoint is required")
		}
		if customTimeout == 0 {
			customTimeout = 10 // Default timeout
		}
		customConfig, err := ParseConfigFromSettings(
			customName, endpoint, customMethod, customHeaders,
			customBodyTemplate, customResponsePath, customLangMapping,
			customTimeout,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse custom translator config: %w", err)
		}
		translator = NewCustomTranslatorWithDB(customConfig, t.settings)
	default:
		translator = NewGoogleFreeTranslator()
	}

	// Cache the translator
	t.cachedTranslator = translator
	t.cachedProvider = provider
	t.cachedAPIKey = apiKey
	t.cachedAppID = appID
	t.cachedSecretKey = secretKey
	t.cachedEndpoint = endpoint
	t.cachedModel = model
	t.cachedPrompt = systemPrompt
	t.cachedCustomHeaders = customHeaders
	t.cachedBodyTemplate = customBodyTemplate
	t.cachedResponsePath = customResponsePath
	t.cachedLangMapping = customLangMapping
	t.cachedTimeout = customTimeout

	return translator, provider, nil
}

// isLocalEndpoint checks if an endpoint URL points to a local service (localhost, 127.0.0.1, etc.)
// This allows using empty API keys for local LLM services like Ollama
func isLocalEndpoint(endpointURL string) bool {
	if endpointURL == "" {
		return false
	}

	// Parse the URL to extract the host
	parsedURL, err := url.Parse(endpointURL)
	if err != nil {
		return false
	}

	host := parsedURL.Host
	// Remove port if present
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		// Handle IPv6 addresses like [::1]:8080
		if !strings.Contains(host[idx:], "]") {
			host = host[:idx]
		}
	}
	// Remove brackets from IPv6 addresses
	host = strings.Trim(host, "[]")

	return host == "localhost" ||
		host == "127.0.0.1" ||
		host == "::1" ||
		strings.HasPrefix(host, "127.") ||
		host == "0.0.0.0"
}

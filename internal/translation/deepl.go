package translation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type DeepLTranslator struct {
	APIKey   string
	Endpoint string // Custom endpoint for deeplx self-hosted service
	client   *http.Client
	db       DBInterface
}

// NewDeepLTranslator creates a new DeepL Translator
// db is optional - if nil, no proxy will be used
func NewDeepLTranslator(apiKey string) *DeepLTranslator {
	return &DeepLTranslator{
		APIKey:   apiKey,
		Endpoint: "",
		client:   &http.Client{Timeout: 10 * time.Second},
		db:       nil,
	}
}

// NewDeepLTranslatorWithEndpoint creates a new DeepL Translator with custom endpoint (for deeplx)
func NewDeepLTranslatorWithEndpoint(apiKey, endpoint string) *DeepLTranslator {
	return &DeepLTranslator{
		APIKey:   apiKey,
		Endpoint: strings.TrimSuffix(endpoint, "/"),
		client:   &http.Client{Timeout: 10 * time.Second},
		db:       nil,
	}
}

// NewDeepLTranslatorWithDB creates a new DeepL Translator with database for proxy support
func NewDeepLTranslatorWithDB(apiKey string, db DBInterface) *DeepLTranslator {
	client, err := CreateHTTPClientWithProxy(db, 10*time.Second)
	if err != nil {
		// Fallback to default client if proxy creation fails
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &DeepLTranslator{
		APIKey:   apiKey,
		Endpoint: "",
		client:   client,
		db:       db,
	}
}

// NewDeepLTranslatorWithEndpointAndDB creates a DeepL Translator with custom endpoint and proxy support
func NewDeepLTranslatorWithEndpointAndDB(apiKey, endpoint string, db DBInterface) *DeepLTranslator {
	client, err := CreateHTTPClientWithProxy(db, 10*time.Second)
	if err != nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &DeepLTranslator{
		APIKey:   apiKey,
		Endpoint: strings.TrimSuffix(endpoint, "/"),
		client:   client,
		db:       db,
	}
}

func (t *DeepLTranslator) Translate(text, targetLang string) (string, error) {
	if text == "" {
		return "", nil
	}

	// Use custom endpoint if provided (for deeplx self-hosted service)
	if t.Endpoint != "" {
		return t.translateWithDeeplx(text, targetLang)
	}

	// Standard DeepL API
	apiURL := "https://api.deepl.com/v2/translate"
	if strings.HasSuffix(t.APIKey, ":fx") {
		apiURL = "https://api-free.deepl.com/v2/translate"
	}

	data := url.Values{}
	data.Set("auth_key", t.APIKey)
	data.Set("text", text)
	data.Set("target_lang", strings.ToUpper(targetLang))

	resp, err := t.client.PostForm(apiURL, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if err := rateLimitResponse(resp); err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("deepl api returned status: %d", resp.StatusCode)
	}

	var result struct {
		Translations []struct {
			Text string `json:"text"`
		} `json:"translations"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Translations) > 0 {
		return result.Translations[0].Text, nil
	}

	return "", fmt.Errorf("no translation found")
}

// translateWithDeeplx handles translation using deeplx self-hosted service
// deeplx API: POST /translate with JSON body {text, source_lang, target_lang}
func (t *DeepLTranslator) translateWithDeeplx(text, targetLang string) (string, error) {
	apiURL := t.Endpoint + "/translate"

	requestBody := map[string]string{
		"text":        text,
		"source_lang": "auto",
		"target_lang": strings.ToUpper(targetLang),
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal deeplx request: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create deeplx request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Add Authorization header if API key is provided (some deeplx deployments require it)
	if t.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+t.APIKey)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("deeplx request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := rateLimitResponse(resp); err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("deeplx returned status: %d", resp.StatusCode)
	}

	// deeplx response format: {code, message, data, source_lang, target_lang, alternatives}
	var result struct {
		Code    int      `json:"code"`
		Message string   `json:"message"`
		Data    string   `json:"data"`
		Alt     []string `json:"alternatives"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode deeplx response: %w", err)
	}

	if result.Code != 200 {
		return "", fmt.Errorf("deeplx error: %s", result.Message)
	}

	if result.Data != "" {
		return result.Data, nil
	}

	return "", fmt.Errorf("no translation found from deeplx")
}

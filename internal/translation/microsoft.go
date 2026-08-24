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

// MicrosoftTranslator implements translation using Microsoft Translator Text API v3.0.
// API Documentation: https://docs.microsoft.com/azure/cognitive-services/translator/reference/v3-0-translate
type MicrosoftTranslator struct {
	APIKey   string
	Region   string // Optional region for multi-service resources
	Endpoint string // Custom endpoint (defaults to api.cognitive.microsofttranslator.com)
	client   *http.Client
	db       DBInterface
}

// NewMicrosoftTranslator creates a new Microsoft Translator instance.
// db is optional - if nil, no proxy will be used
func NewMicrosoftTranslator(apiKey string) *MicrosoftTranslator {
	return &MicrosoftTranslator{
		APIKey:   apiKey,
		Endpoint: "",
		client:   &http.Client{Timeout: 10 * time.Second},
		db:       nil,
	}
}

// NewMicrosoftTranslatorWithRegion creates a new Microsoft Translator with region.
func NewMicrosoftTranslatorWithRegion(apiKey, region string) *MicrosoftTranslator {
	return &MicrosoftTranslator{
		APIKey:   apiKey,
		Region:   region,
		Endpoint: "",
		client:   &http.Client{Timeout: 10 * time.Second},
		db:       nil,
	}
}

// NewMicrosoftTranslatorWithEndpoint creates a new Microsoft Translator with custom endpoint.
func NewMicrosoftTranslatorWithEndpoint(apiKey, endpoint string) *MicrosoftTranslator {
	return &MicrosoftTranslator{
		APIKey:   apiKey,
		Endpoint: strings.TrimSuffix(endpoint, "/"),
		client:   &http.Client{Timeout: 10 * time.Second},
		db:       nil,
	}
}

// NewMicrosoftTranslatorWithDB creates a new Microsoft Translator with database for proxy support.
func NewMicrosoftTranslatorWithDB(apiKey string, db DBInterface) *MicrosoftTranslator {
	client, err := CreateHTTPClientWithProxy(db, 10*time.Second)
	if err != nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &MicrosoftTranslator{
		APIKey:   apiKey,
		Endpoint: "",
		client:   client,
		db:       db,
	}
}

// NewMicrosoftTranslatorWithAll creates a new Microsoft Translator with all options.
func NewMicrosoftTranslatorWithAll(apiKey, region, endpoint string, db DBInterface) *MicrosoftTranslator {
	client, err := CreateHTTPClientWithProxy(db, 10*time.Second)
	if err != nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &MicrosoftTranslator{
		APIKey:   apiKey,
		Region:   region,
		Endpoint: strings.TrimSuffix(endpoint, "/"),
		client:   client,
		db:       db,
	}
}

// Translate translates text to the target language using Microsoft Translator Text API v3.0.
func (t *MicrosoftTranslator) Translate(text, targetLang string) (string, error) {
	if text == "" {
		return "", nil
	}

	// Map language code to Microsoft format
	msLang := mapToMicrosoftLang(targetLang)

	// Determine endpoint
	endpoint := t.Endpoint
	if endpoint == "" {
		endpoint = "https://api.cognitive.microsofttranslator.com"
	}

	// Build request URL
	apiURL := fmt.Sprintf("%s/translate?api-version=3.0&to=%s", endpoint, url.QueryEscape(msLang))

	// Build request body - Microsoft API expects an array of objects
	requestBody := []map[string]string{{"Text": text}}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal microsoft request: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create microsoft request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Ocp-Apim-Subscription-Key", t.APIKey)

	// Add region header if provided (required for multi-service resources)
	if t.Region != "" {
		req.Header.Set("Ocp-Apim-Subscription-Region", t.Region)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("microsoft api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("microsoft api returned status: %d", resp.StatusCode)
	}

	// Parse response
	var result []struct {
		Translations []struct {
			Text string `json:"text"`
			To   string `json:"to"`
		} `json:"translations"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode microsoft response: %w", err)
	}

	if len(result) > 0 && len(result[0].Translations) > 0 {
		return result[0].Translations[0].Text, nil
	}

	return "", fmt.Errorf("no translation found in microsoft response")
}

// mapToMicrosoftLang maps standard language codes to Microsoft's format.
// Microsoft uses BCP 47 language tags (e.g., "zh-Hans", "en-US").
func mapToMicrosoftLang(lang string) string {
	langMap := map[string]string{
		"en":    "en",
		"zh":    "zh-Hans",
		"zh-TW": "zh-Hant",
		"es":    "es",
		"fr":    "fr",
		"de":    "de",
		"ja":    "ja",
		"ko":    "ko",
		"pt":    "pt",
		"ru":    "ru",
		"it":    "it",
		"ar":    "ar",
		"tr":    "tr",
		"pl":    "pl",
		"nl":    "nl",
		"sv":    "sv",
		"da":    "da",
		"fi":    "fi",
		"no":    "nb",
		"cs":    "cs",
		"el":    "el",
		"he":    "he",
		"id":    "id",
		"ms":    "ms",
		"th":    "th",
		"uk":    "uk",
		"vi":    "vi",
		"hi":    "hi",
		"bn":    "bn",
	}
	if msLang, ok := langMap[lang]; ok {
		return msLang
	}
	return lang
}

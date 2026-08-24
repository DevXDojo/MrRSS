package translation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
)

type GoogleFreeTranslator struct {
	client *http.Client
	db     DBInterface
}

// NewGoogleFreeTranslator creates a new Google Free Translator
// db is optional - if nil, no proxy will be used
func NewGoogleFreeTranslator() *GoogleFreeTranslator {
	return &GoogleFreeTranslator{
		client: &http.Client{Timeout: 10 * time.Second},
		db:     nil,
	}
}

// NewGoogleFreeTranslatorWithDB creates a new Google Free Translator with database for proxy support
func NewGoogleFreeTranslatorWithDB(db DBInterface) *GoogleFreeTranslator {
	client, err := CreateHTTPClientWithProxy(db, 10*time.Second)
	if err != nil {
		// Fallback to default client if proxy creation fails
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &GoogleFreeTranslator{
		client: client,
		db:     db,
	}
}

func (t *GoogleFreeTranslator) Translate(text, targetLang string) (string, error) {
	if text == "" {
		return "", nil
	}
	if textAlreadyMatchesTarget(text, targetLang) {
		return text, nil
	}

	translated, err := t.translateWithGoogle(text, targetLang)
	if err != nil {
		return "", fmt.Errorf("google translation service is unavailable: %s", sanitizedTranslationError(err))
	}
	return translated, nil
}

func (t *GoogleFreeTranslator) translateWithGoogle(text, targetLang string) (string, error) {
	chunks := splitTranslationText(text, 1_800)
	translatedChunks := make([]string, 0, len(chunks))
	for _, chunk := range chunks {
		translated, err := t.translateGoogleChunk(chunk, targetLang)
		if err != nil {
			return "", err
		}
		translatedChunks = append(translatedChunks, translated)
	}
	return strings.Join(translatedChunks, ""), nil
}

func (t *GoogleFreeTranslator) translateGoogleChunk(text, targetLang string) (string, error) {

	// Get the configured endpoint, default to translate.googleapis.com
	endpoint := "translate.googleapis.com"
	if t.db != nil {
		if configuredEndpoint, err := t.db.GetSetting("google_translate_endpoint"); err == nil && configuredEndpoint != "" {
			endpoint = configuredEndpoint
		}
	}

	// Determine which client parameter and path to use based on endpoint
	var baseURL string
	var clientParam string

	if endpoint == "clients5.google.com" {
		baseURL = "https://clients5.google.com/translate_a/t"
		clientParam = "dict-chrome-ex"
	} else {
		// Default to translate.googleapis.com or any other endpoint
		baseURL = "https://" + endpoint + "/translate_a/single"
		clientParam = "gtx"
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("client", clientParam)
	q.Set("sl", "auto")
	q.Set("tl", targetLang)
	q.Set("dt", "t")
	q.Set("q", text)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("translation api returned status: %d", resp.StatusCode)
	}

	// The response is a complex nested array structure
	// [[[ "translated", "original", ... ]], ...]
	var result []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result) > 0 {
		if inner, ok := result[0].([]interface{}); ok {
			var translatedText string
			for _, slice := range inner {
				if s, ok := slice.([]interface{}); ok && len(s) > 0 {
					if str, ok := s[0].(string); ok {
						translatedText += str
					}
				}
			}
			if translatedText != "" {
				return translatedText, nil
			}
		}
	}

	return "", fmt.Errorf("invalid response format")
}

func textAlreadyMatchesTarget(text, targetLang string) bool {
	var letters, cjk, latin int
	for _, r := range text {
		if unicode.IsLetter(r) {
			letters++
			if unicode.In(r, unicode.Han) {
				cjk++
			} else if unicode.In(r, unicode.Latin) {
				latin++
			}
		}
	}
	if letters == 0 {
		return true
	}
	switch strings.ToLower(targetLang) {
	case "zh", "zh-cn", "zh-tw", "zh-hans", "zh-hant":
		return cjk*100/letters >= 70
	case "en", "en-us", "en-gb":
		return latin*100/letters >= 95
	default:
		return false
	}
}

func splitTranslationText(text string, maxBytes int) []string {
	if len([]byte(text)) <= maxBytes {
		return []string{text}
	}

	chunks := make([]string, 0, len(text)/maxBytes+1)
	var builder strings.Builder
	currentBytes := 0
	for _, r := range text {
		runeBytes := len(string(r))
		if currentBytes > 0 && currentBytes+runeBytes > maxBytes {
			chunks = append(chunks, builder.String())
			builder.Reset()
			currentBytes = 0
		}
		builder.WriteRune(r)
		currentBytes += runeBytes
	}
	if builder.Len() > 0 {
		chunks = append(chunks, builder.String())
	}
	return chunks
}

func sanitizedTranslationError(err error) string {
	if err == nil {
		return "unknown error"
	}
	message := err.Error()
	if strings.Contains(message, "context deadline exceeded") || strings.Contains(message, "Client.Timeout") {
		return "request timed out; configure the proxy or select another translation provider"
	}
	if strings.Contains(message, "no such host") {
		return "service address could not be resolved; configure the proxy or select another translation provider"
	}
	if len(message) > 160 {
		return message[:160]
	}
	return message
}

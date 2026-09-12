package translation

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
)

const (
	edgeAuthURL      = "https://edge.microsoft.com/translate/auth"
	edgeTranslateURL = "https://api-edge.cognitive.microsofttranslator.com/translate"
	edgeUserAgent    = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36 Edg/113.0.1774.42"
)

// edgeProvider uses the consumer Edge translation protocol, independently of
// Azure subscription credentials. Tokens are kept only in memory.
type edgeProvider struct {
	client       *http.Client
	gate         chan struct{} // Bounds requests and protects the token cache.
	token        string
	tokenExpires time.Time
}

func newEdgeProvider(settings SettingsProvider) (*edgeProvider, error) {
	client, err := CreateHTTPClientWithProxy(settings, 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("configure Microsoft Edge translation proxy: %w", err)
	}
	// Never forward a bearer token or article text to a redirect destination.
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	return &edgeProvider{client: client, gate: make(chan struct{}, 1)}, nil
}

func (p *edgeProvider) Name() string                 { return string(ProviderMicrosoftEdge) }
func (p *edgeProvider) IsAvailable() bool            { return true }
func (p *edgeProvider) SupportedLanguages() []string { return nil }

func (p *edgeProvider) Translate(ctx context.Context, text, targetLang string) (*TranslationResult, error) {
	result := &TranslationResult{Original: text, FromLang: "auto", ToLang: targetLang, Provider: p.Name()}
	if strings.TrimSpace(text) == "" {
		result.Translated = text
		return result, nil
	}
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	select {
	case p.gate <- struct{}{}:
		defer func() { <-p.gate }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	var translated strings.Builder
	for _, chunk := range splitEdgeText(text) {
		part, err := p.translateChunk(ctx, chunk, targetLang)
		if err != nil {
			return nil, err // Do not cache a partially translated article.
		}
		translated.WriteString(part)
	}
	result.Translated = translated.String()
	return result, nil
}

func (p *edgeProvider) translateChunk(ctx context.Context, text, targetLang string) (string, error) {
	// Preserve whitespace at chunk boundaries even if the service trims it.
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return text, nil
	}
	prefix := text[:len(text)-len(strings.TrimLeftFunc(text, unicode.IsSpace))]
	suffix := text[len(strings.TrimRightFunc(text, unicode.IsSpace)):]
	body, err := json.Marshal([]map[string]string{{"Text": trimmed}})
	if err != nil {
		return "", fmt.Errorf("encode Edge translation request: %w", err)
	}
	endpoint := edgeTranslateURL + "?api-version=3.0&to=" + url.QueryEscape(mapToMicrosoftLang(targetLang))
	for attempt := 0; attempt < 2; attempt++ {
		if err := p.ensureToken(ctx); err != nil {
			return "", err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		req.Header.Set("User-Agent", edgeUserAgent)
		req.Header.Set("Authorization", "Bearer "+p.token)
		resp, err := p.client.Do(req)
		if err != nil {
			return "", fmt.Errorf("Microsoft Edge translation request: %w", err)
		}
		if resp.StatusCode == http.StatusUnauthorized {
			resp.Body.Close()
			p.token = ""
			if attempt == 0 {
				continue
			}
			return "", fmt.Errorf("Microsoft Edge translation authorization expired (HTTP 401)")
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return "", fmt.Errorf("Microsoft Edge translation returned HTTP %d", resp.StatusCode)
		}
		data, readErr := io.ReadAll(io.LimitReader(resp.Body, (2<<20)+1))
		resp.Body.Close()
		if readErr != nil {
			return "", fmt.Errorf("read Microsoft Edge translation: %w", readErr)
		}
		if len(data) > 2<<20 {
			return "", fmt.Errorf("Microsoft Edge translation response is too large")
		}
		var response []struct {
			Translations []struct {
				Text string `json:"text"`
			} `json:"translations"`
		}
		if err := json.Unmarshal(data, &response); err != nil {
			return "", fmt.Errorf("invalid Microsoft Edge translation response")
		}
		if len(response) != 1 || len(response[0].Translations) == 0 || strings.TrimSpace(response[0].Translations[0].Text) == "" {
			return "", fmt.Errorf("Microsoft Edge returned no translation")
		}
		return prefix + strings.TrimSpace(response[0].Translations[0].Text) + suffix, nil
	}
	return "", fmt.Errorf("Microsoft Edge translation authorization failed")
}

func (p *edgeProvider) ensureToken(ctx context.Context) error {
	if p.token != "" && time.Now().Before(p.tokenExpires) {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, edgeAuthURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", edgeUserAgent)
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("request Microsoft Edge translation token: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Microsoft Edge translation authorization returned HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, (16<<10)+1))
	if err != nil {
		return fmt.Errorf("read Microsoft Edge translation token: %w", err)
	}
	if len(data) > 16<<10 {
		return fmt.Errorf("Microsoft Edge translation token is too large")
	}
	token := strings.TrimSpace(string(data))
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid Microsoft Edge translation token")
	}
	for _, part := range parts {
		if _, err := base64.RawURLEncoding.DecodeString(part); err != nil || part == "" {
			return fmt.Errorf("invalid Microsoft Edge translation token")
		}
	}
	// The expiry claim is only a cache hint; authentication remains the service's
	// responsibility. Cap the local lifetime even if the server claims longer.
	expires := time.Now().Add(5 * time.Minute)
	claimsJSON, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if json.Unmarshal(claimsJSON, &claims) == nil && claims.Exp > 0 {
		if hinted := time.Unix(claims.Exp, 0).Add(-30 * time.Second); hinted.Before(expires) {
			expires = hinted
		}
	}
	p.token, p.tokenExpires = token, expires
	return nil
}

// Use modest request sizes, counting UTF-16 units so supplementary characters
// are never split or counted as one unit.
func splitEdgeText(text string) []string {
	const limit = 5000
	var chunks []string
	for len(text) > 0 {
		units, cut, boundary := 0, len(text), 0
		for index, r := range text {
			width := 1
			if r > 0xffff {
				width = 2
			}
			if units+width > limit {
				cut = index
				break
			}
			units += width
			if unicode.IsSpace(r) && units >= limit/2 {
				boundary = index
			}
		}
		if cut < len(text) && boundary > 0 {
			cut = boundary
		}
		chunks = append(chunks, text[:cut])
		text = text[cut:]
	}
	return chunks
}

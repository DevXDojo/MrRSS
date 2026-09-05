package translation

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"MrRSS/internal/ai"
	"MrRSS/internal/utils/httputil"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDeepLTranslate_SuccessAndEmpty(t *testing.T) {
	t1 := NewDeepLTranslator("apikey")

	// empty input should return empty without error
	out, err := t1.Translate("", "es")
	if err != nil || out != "" {
		t.Fatalf("expected empty translate for empty input, got %q err=%v", out, err)
	}

	// Mock client response for success
	t1.client = &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"translations":[{"text":"Hola"}]}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{"Content-Type": {"application/json"}}}, nil
	}), Timeout: 5 * time.Second}

	out2, err := t1.Translate("Hello", "es")
	if err != nil {
		t.Fatalf("DeepL translate failed: %v", err)
	}
	if out2 != "Hola" {
		t.Fatalf("expected Hola, got %s", out2)
	}
}

func TestBaiduTranslate_SuccessAndEmpty(t *testing.T) {
	t1 := NewBaiduTranslator("appid", "secret")

	out, err := t1.Translate("", "en")
	if err != nil || out != "" {
		t.Fatalf("expected empty translate for empty input, got %q err=%v", out, err)
	}

	// Mock baidu response
	t1.client = &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"trans_result":[{"src":"Hello","dst":"你好"}], "error_code":"52000"}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{"Content-Type": {"application/json"}}}, nil
	}), Timeout: 5 * time.Second}

	out2, err := t1.Translate("Hello", "zh")
	if err != nil {
		t.Fatalf("Baidu translate failed: %v", err)
	}
	if out2 != "你好" {
		t.Fatalf("expected 你好, got %s", out2)
	}
}

func TestAITranslate_SuccessAndEmpty(t *testing.T) {
	t1 := NewAITranslator("apikey", "https://api.test", "m1")

	out, err := t1.Translate("", "en")
	if err != nil || out != "" {
		t.Fatalf("expected empty translate for empty input, got %q err=%v", out, err)
	}

	// Mock AI response with custom HTTP client
	testHTTPClient := &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"choices":[{"message":{"content":"Bonjour"}}]}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{"Content-Type": {"application/json"}}}, nil
	}), Timeout: 5 * time.Second}

	// Re-create AI client with custom HTTP client
	t1.client = ai.NewClientWithHTTPClient(ai.ClientConfig{
		APIKey:   "apikey",
		Endpoint: "https://api.test",
		Model:    "m1",
		Timeout:  5 * time.Second,
	}, testHTTPClient)

	out2, err := t1.Translate("Hello", "fr")
	if err != nil {
		t.Fatalf("AI translate failed: %v", err)
	}
	if out2 != "Bonjour" {
		t.Fatalf("expected Bonjour, got %s", out2)
	}
}

func TestAITranslate_AutoDetectOllama(t *testing.T) {
	t1 := NewAITranslator("", "http://localhost:11434/api/generate", "llama3.2:1b")

	// Mock Ollama response (since endpoint is localhost, Ollama format is tried first and should succeed)
	callCount := 0
	testHTTPClient := &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		callCount++
		// Read request body to verify Ollama format
		bodyBytes, _ := io.ReadAll(req.Body)
		var requestBody map[string]interface{}
		json.Unmarshal(bodyBytes, &requestBody)

		// Ollama format should have "prompt" and "model" fields, not "messages"
		if _, hasPrompt := requestBody["prompt"]; !hasPrompt {
			t.Fatalf("expected Ollama format request with 'prompt' field, got: %s", string(bodyBytes))
		}

		// Return successful Ollama response
		body := `{"response":"Bonjour","done":true}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{"Content-Type": {"application/json"}}}, nil
	}), Timeout: 5 * time.Second}

	// Re-create AI client with custom HTTP client
	t1.client = ai.NewClientWithHTTPClient(ai.ClientConfig{
		APIKey:   "",
		Endpoint: "http://localhost:11434/api/generate",
		Model:    "llama3.2:1b",
		Timeout:  5 * time.Second,
	}, testHTTPClient)

	out, err := t1.Translate("Hello", "fr")
	if err != nil {
		t.Fatalf("AI translate auto-detect failed: %v", err)
	}
	if out != "Bonjour" {
		t.Fatalf("expected Bonjour, got %s", out)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 API call (Ollama format should succeed on first try), got %d", callCount)
	}
}

func TestAITranslate_RemovesThinkingBlocks(t *testing.T) {
	t1 := NewAITranslator("apikey", "https://api.test", "m1")

	testHTTPClient := &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"choices":[{"message":{"content":"<thinking>analyze first</thinking>\nBonjour"}}]}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{"Content-Type": {"application/json"}}}, nil
	}), Timeout: 5 * time.Second}

	t1.client = ai.NewClientWithHTTPClient(ai.ClientConfig{
		APIKey:   "apikey",
		Endpoint: "https://api.test",
		Model:    "m1",
		Timeout:  5 * time.Second,
	}, testHTTPClient)

	out, err := t1.Translate("Hello", "fr")
	if err != nil {
		t.Fatalf("AI translate failed: %v", err)
	}
	if out != "Bonjour" {
		t.Fatalf("expected thinking-free translation, got %q", out)
	}
}

func TestAITranslatorRetainsHTTPClientAfterConfigurationChanges(t *testing.T) {
	translator := NewAITranslator("apikey", "https://api.test", "m1")
	requestCount := 0
	injectedClient := &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		requestCount++
		if req.Header.Get("X-Gateway") != "enabled" {
			t.Fatalf("expected custom header to be retained, got %q", req.Header.Get("X-Gateway"))
		}
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}
		if !strings.Contains(string(bodyBytes), "Custom translation prompt") {
			t.Fatalf("expected custom system prompt in request: %s", string(bodyBytes))
		}
		body := `{"choices":[{"message":{"content":"Bonjour"}}]}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     http.Header{"Content-Type": {"application/json"}},
		}, nil
	}), Timeout: 5 * time.Second}

	translator.httpClient = injectedClient
	translator.SetSystemPrompt("Custom translation prompt")
	translator.SetCustomHeaders(`{"X-Gateway":"enabled"}`)

	translated, err := translator.Translate("Hello", "fr")
	if err != nil {
		t.Fatalf("Translate failed: %v", err)
	}
	if translated != "Bonjour" {
		t.Fatalf("expected Bonjour, got %q", translated)
	}
	if requestCount != 1 {
		t.Fatalf("expected injected transport to receive one request, got %d", requestCount)
	}
}

func TestAITranslatorNegotiatesHTTP2WithCompatibleGateway(t *testing.T) {
	t.Setenv(httputil.InsecureSkipTLSVerifyEnv, "true")

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 2 {
			t.Errorf("expected translation request over HTTP/2, got %s", r.Proto)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"response":"Bonjour","done":true}`)
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	translator := NewAITranslator("", server.URL+"/api/generate", "test-model")
	translated, err := translator.Translate("Hello", "fr")
	if err != nil {
		t.Fatalf("Translate failed over HTTP/2: %v", err)
	}
	if translated != "Bonjour" {
		t.Fatalf("expected Bonjour, got %q", translated)
	}
}

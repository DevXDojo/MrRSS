package translation

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
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

	// Mock AI response
	t1.client = &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"choices":[{"message":{"content":"Bonjour"}}]}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{"Content-Type": {"application/json"}}}, nil
	}), Timeout: 5 * time.Second}

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

	// Mock Ollama response (first try OpenAI format, which should fail, then try Ollama format)
	callCount := 0
	t1.client = &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			// First call - OpenAI format fails
			return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader(`{"error":{"message":"Invalid format"}}`)), Header: http.Header{"Content-Type": {"application/json"}}}, nil
		}
		// Second call - Ollama format succeeds
		body := `{"response":"Bonjour","done":true}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{"Content-Type": {"application/json"}}}, nil
	}), Timeout: 5 * time.Second}

	out, err := t1.Translate("Hello", "fr")
	if err != nil {
		t.Fatalf("AI translate auto-detect failed: %v", err)
	}
	if out != "Bonjour" {
		t.Fatalf("expected Bonjour, got %s", out)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 API calls (OpenAI then Ollama), got %d", callCount)
	}
}

func TestGoogleTranslateSkipsRequestWhenTextMatchesTargetLanguage(t *testing.T) {
	translator := NewGoogleFreeTranslator()
	requestCount := 0
	translator.client = &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		requestCount++
		return nil, context.DeadlineExceeded
	})}

	translated, err := translator.Translate("这里是中文内容", "zh")
	if err != nil {
		t.Fatalf("expected same-language translation to succeed: %v", err)
	}
	if translated != "这里是中文内容" {
		t.Fatalf("expected unchanged text, got %q", translated)
	}
	if requestCount != 0 {
		t.Fatalf("expected no network requests, got %d", requestCount)
	}
}

func TestGoogleTranslateReturnsSanitizedTimeout(t *testing.T) {
	translator := NewGoogleFreeTranslator()
	translator.client = &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		return nil, context.DeadlineExceeded
	})}

	_, err := translator.Translate("Hello", "zh")
	if err == nil {
		t.Fatal("expected translation to fail")
	}
	if strings.Contains(err.Error(), "q=Hello") || !strings.Contains(err.Error(), "configure the proxy") {
		t.Fatalf("expected a concise configuration hint without source text, got %q", err.Error())
	}
}

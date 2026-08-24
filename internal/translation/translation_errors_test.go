package translation

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"MrRSS/internal/ai"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDeepL_Non200AndTimeout(t *testing.T) {
	t1 := NewDeepLTranslator("key")

	// Non-200
	t1.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader("error")), Header: make(http.Header)}, nil
	}), Timeout: 2 * time.Second}

	if _, err := t1.Translate("hello", "en"); err == nil {
		t.Fatalf("expected error for non-200 response")
	}

	// Timeout simulated by returning an error
	t1.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, context.DeadlineExceeded
	}), Timeout: 20 * time.Millisecond}

	if _, err := t1.Translate("hello", "en"); err == nil {
		t.Fatalf("expected timeout error")
	}
}

func TestBaidu_Non200AndTimeout(t *testing.T) {
	b := NewBaiduTranslator("app", "secret")

	b.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 502, Body: io.NopCloser(strings.NewReader("bad")), Header: make(http.Header)}, nil
	}), Timeout: 2 * time.Second}

	if _, err := b.Translate("hello", "en"); err == nil {
		t.Fatalf("expected error for non-200 response")
	}

	b.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, context.DeadlineExceeded
	}), Timeout: 20 * time.Millisecond}

	if _, err := b.Translate("hello", "en"); err == nil {
		t.Fatalf("expected timeout error")
	}
}

func TestAI_Non200AndTimeout(t *testing.T) {
	a := NewAITranslator("k", "https://api.test", "m")

	// Create custom HTTP client for testing
	testHTTPClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader(`{"error":{"message":"srv"}}`)), Header: make(http.Header)}, nil
	}), Timeout: 2 * time.Second}

	// Re-create AI client with custom HTTP client
	a.client = ai.NewClientWithHTTPClient(ai.ClientConfig{
		APIKey:   "k",
		Endpoint: "https://api.test",
		Model:    "m",
		Timeout:  2 * time.Second,
	}, testHTTPClient)

	if _, err := a.Translate("hello", "en"); err == nil {
		t.Fatalf("expected error for non-200 response")
	}

	// Test timeout
	testHTTPClient2 := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, context.DeadlineExceeded
	}), Timeout: 20 * time.Millisecond}

	a.client = ai.NewClientWithHTTPClient(ai.ClientConfig{
		APIKey:   "k",
		Endpoint: "https://api.test",
		Model:    "m",
		Timeout:  20 * time.Millisecond,
	}, testHTTPClient2)

	if _, err := a.Translate("hello", "en"); err == nil {
		t.Fatalf("expected timeout error")
	}
}

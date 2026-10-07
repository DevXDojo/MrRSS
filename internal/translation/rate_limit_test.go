package translation

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"MrRSS/internal/ai"
)

func TestProvidersPreserveRateLimitDelay(t *testing.T) {
	tests := []struct {
		name      string
		translate func(*http.Client) error
	}{
		{"DeepL", func(client *http.Client) error {
			translator := NewDeepLTranslator("test-key")
			translator.client = client
			_, err := translator.Translate("Hello", "zh")
			return err
		}},
		{"DeepLX", func(client *http.Client) error {
			translator := NewDeepLTranslator("")
			translator.Endpoint = "https://example.org"
			translator.client = client
			_, err := translator.Translate("Hello", "zh")
			return err
		}},
		{"Baidu", func(client *http.Client) error {
			translator := NewBaiduTranslator("test-id", "test-key")
			translator.client = client
			_, err := translator.Translate("Hello", "zh")
			return err
		}},
		{"Tencent", func(client *http.Client) error {
			translator := NewTencentTranslator("test-id", "test-key")
			translator.client = client
			_, err := translator.Translate("Hello", "zh")
			return err
		}},
		{"Microsoft", func(client *http.Client) error {
			translator := NewMicrosoftTranslator("test-key")
			translator.client = client
			_, err := translator.Translate("Hello", "zh")
			return err
		}},
		{"Custom", func(client *http.Client) error {
			translator := NewCustomTranslator(&CustomTranslatorConfig{
				Endpoint: "https://example.org", Method: "POST", BodyTemplate: `{"text":{{text}}}`, ResponsePath: "text",
			})
			translator.client = client
			_, err := translator.Translate("Hello", "zh")
			return err
		}},
		{"Edge", func(client *http.Client) error {
			provider := &edgeProvider{client: client, gate: make(chan struct{}, 1), token: "test-token", tokenExpires: time.Now().Add(time.Hour)}
			_, err := provider.Translate(context.Background(), "Hello", "zh")
			return err
		}},
		{"Edge authorization", func(client *http.Client) error {
			provider := &edgeProvider{client: client, gate: make(chan struct{}, 1)}
			_, err := provider.Translate(context.Background(), "Hello", "zh")
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: 429, Header: http.Header{"Retry-After": {"90"}}, Body: io.NopCloser(strings.NewReader("private provider details"))}, nil
			})}
			err := tt.translate(client)
			var limited *RateLimitError
			if !errors.As(err, &limited) || limited.RetryAfter != 90*time.Second || calls != 1 {
				t.Fatalf("expected one request and a 90s delay, got %d requests, %v", calls, err)
			}
		})
	}
}

func TestAIRateLimitIsRetryable(t *testing.T) {
	translator := NewAITranslator("test-key", "https://example.org/v1/chat/completions", "test-model")
	translator.client = ai.NewClientWithHTTPClient(ai.ClientConfig{
		Endpoint: translator.Endpoint, Model: translator.Model,
	}, &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 429, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"error":{"message":"limited"}}`))}, nil
	})})
	_, err := translator.TranslateContext(context.Background(), "Hello", "zh")
	if !IsRateLimited(err) {
		t.Fatalf("AI limit was not retryable: %v", err)
	}
}

func TestAICancellationReachesProvider(t *testing.T) {
	translator := NewAITranslator("test-key", "https://example.org/v1/chat/completions", "test-model")
	ctx, cancel := context.WithCancel(context.Background())
	translator.client = ai.NewClientWithHTTPClient(ai.ClientConfig{
		Endpoint: translator.Endpoint, Model: translator.Model,
	}, &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		cancel()
		if r.Context().Err() == nil {
			t.Fatal("AI request did not receive cancellation")
		}
		return nil, r.Context().Err()
	})})
	defer cancel()
	_, err := translator.TranslateContext(ctx, "Hello", "zh")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}

func TestGoogleRateLimitCooldown(t *testing.T) {
	translator := NewGoogleFreeTranslator()
	calls := 0
	translator.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 429, Header: http.Header{"Retry-After": {"120"}}, Body: io.NopCloser(strings.NewReader("limited"))}, nil
	})
	for range 2 {
		_, err := translator.Translate("Hello", "zh")
		var limited *RateLimitError
		if !errors.As(err, &limited) || limited.RetryAfter < 119*time.Second {
			t.Fatalf("expected provider retry delay, got %v", err)
		}
	}
	if calls != 2 {
		t.Fatalf("provider was contacted %d times during cooldown", calls)
	}
	translator.retryAt["translate.googleapis.com"] = time.Now().Add(-time.Second)
	_, _ = translator.Translate("Hello", "zh")
	if calls != 3 {
		t.Fatal("translation did not resume after cooldown")
	}
}

func TestChineseVariantConcurrent(t *testing.T) {
	var workers sync.WaitGroup
	for range 16 {
		workers.Go(func() {
			for _, tc := range []struct{ text, want string }{
				{"这是一篇关于技术和编程的测试文章。", "zh"},
				{"這是一篇關於技術與程式設計的測試文章。", "zh-TW"},
			} {
				if got := detectChineseVariant(tc.text); got != tc.want {
					t.Errorf("got %s want %s", got, tc.want)
				}
			}
		})
	}
	workers.Wait()
}

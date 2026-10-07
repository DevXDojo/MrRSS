package translation

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGoogleFailoverKeepsPreferredEndpointAndIndependentCooldowns(t *testing.T) {
	for _, preferred := range []string{"translate.googleapis.com", "clients5.google.com"} {
		t.Run(preferred, func(t *testing.T) {
			settings := &mockSettingsProvider{settings: map[string]string{"google_translate_endpoint": preferred}}
			translator := NewGoogleFreeTranslatorWithDB(settings)
			calls := make(map[string]int)
			healthy := false
			translator.client.Transport = rtFunc(func(r *http.Request) (*http.Response, error) {
				calls[r.URL.Host]++
				if r.URL.Query().Get("q") != "Hello" || r.URL.Query().Get("tl") != "zh-CN" {
					t.Fatalf("request contract changed: %s", r.URL.RawQuery)
				}
				if r.URL.Host == preferred && !healthy {
					return &http.Response{StatusCode: 429, Header: http.Header{"Retry-After": {"120"}}, Body: io.NopCloser(strings.NewReader("limited"))}, nil
				}
				body := `[[["你好","Hello"]]]`
				if r.URL.Host == "clients5.google.com" {
					if r.URL.Path != "/translate_a/t" || r.URL.Query().Get("client") != "dict-chrome-ex" {
						t.Fatal("wrong dictionary endpoint")
					}
					body = `{"sentences":[{"trans":"你好"}]}`
				}
				return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			for range 2 {
				out, err := translator.Translate("Hello", "zh")
				if err != nil || out != "你好" {
					t.Fatalf("fallback failed: %q, %v", out, err)
				}
			}
			if calls[preferred] != 1 {
				t.Fatalf("retried limited endpoint %d times", calls[preferred])
			}
			if calls["translate.googleapis.com"]+calls["clients5.google.com"] != 3 {
				t.Fatalf("unexpected requests: %v", calls)
			}
			healthy = true
			translator.retryAt[preferred] = time.Now().Add(-time.Second)
			out, err := translator.Translate("Hello", "zh")
			if err != nil || out != "你好" || calls[preferred] != 2 {
				t.Fatalf("preferred endpoint not restored: %q, %v, %v", out, err, calls)
			}
		})
	}
}

func TestGoogleBothEndpointsLimitedRetryEarliestAvailable(t *testing.T) {
	translator := NewGoogleFreeTranslator()
	calls := make(map[string]int)
	translator.client.Transport = rtFunc(func(r *http.Request) (*http.Response, error) {
		calls[r.URL.Host]++
		retry := "120"
		if r.URL.Host == "clients5.google.com" {
			retry = "30"
		}
		return &http.Response{StatusCode: 429, Header: http.Header{"Retry-After": {retry}}, Body: io.NopCloser(strings.NewReader("limited"))}, nil
	})
	for range 2 {
		_, err := translator.Translate("Hello", "zh")
		var limited *RateLimitError
		if !errors.As(err, &limited) || limited.RetryAfter < 29*time.Second || limited.RetryAfter > 30*time.Second {
			t.Fatalf("did not use earliest recovery: %v", err)
		}
	}
	if calls["translate.googleapis.com"] != 1 || calls["clients5.google.com"] != 1 {
		t.Fatalf("repeated limited calls: %v", calls)
	}
	translator.retryAt["clients5.google.com"] = time.Now().Add(-time.Second)
	_, _ = translator.Translate("Hello", "zh")
	if calls["translate.googleapis.com"] != 1 || calls["clients5.google.com"] != 2 {
		t.Fatalf("retried unavailable endpoint: %v", calls)
	}
}

func TestGoogleCustomEndpointAndOtherFailuresDoNotChangeHosts(t *testing.T) {
	for _, tc := range []struct {
		endpoint string
		status   int
	}{
		{"translation.example.org", 429}, {"translate.googleapis.com", 500}, {"clients5.google.com", 401},
	} {
		t.Run(tc.endpoint, func(t *testing.T) {
			translator := NewGoogleFreeTranslatorWithDB(&mockSettingsProvider{settings: map[string]string{"google_translate_endpoint": tc.endpoint}})
			calls := 0
			translator.client.Transport = rtFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.URL.Host != tc.endpoint {
					t.Fatalf("changed configured host to %s", r.URL.Host)
				}
				return &http.Response{StatusCode: tc.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("error"))}, nil
			})
			_, err := translator.Translate("Hello", "zh")
			if err == nil || calls != 1 {
				t.Fatalf("unexpected fallback: %d calls, %v", calls, err)
			}
		})
	}
}

func TestGoogleCancellationPreventsFailover(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	translator := NewGoogleFreeTranslator()
	calls := 0
	translator.client.Transport = rtFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		cancel()
		return &http.Response{StatusCode: 429, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("limited"))}, nil
	})
	_, err := translator.TranslateContext(ctx, "Hello", "zh")
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("cancellation not respected: %d calls, %v", calls, err)
	}
}

func TestGoogleFailoverConcurrent(t *testing.T) {
	translator := NewGoogleFreeTranslator()
	var alternateCalls atomic.Int64
	translator.client.Transport = rtFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host == "translate.googleapis.com" {
			return &http.Response{StatusCode: 429, Header: http.Header{"Retry-After": {"120"}}, Body: io.NopCloser(strings.NewReader("limited"))}, nil
		}
		alternateCalls.Add(1)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`["你好"]`))}, nil
	})
	var workers sync.WaitGroup
	for range 16 {
		workers.Go(func() {
			result, err := translator.Translate("Hello", "zh")
			if err != nil || result != "你好" {
				t.Errorf("concurrent failover failed: %q, %v", result, err)
			}
		})
	}
	workers.Wait()
	if alternateCalls.Load() != 16 {
		t.Fatalf("lost translation requests: %d", alternateCalls.Load())
	}
}

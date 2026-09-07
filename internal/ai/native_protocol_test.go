package ai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type protocolTransport func(*http.Request) (*http.Response, error)

func (f protocolTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNativeProtocolsThroughClient(t *testing.T) {
	cases := []struct{ name, endpoint, path, protocol string }{
		{"Claude root", "https://api.anthropic.com", "/v1/messages", "anthropic"},
		{"Claude gateway", "https://gateway.example/prefix/v1/messages?option=1", "/prefix/v1/messages", "anthropic"},
		{"Gemini base", "https://generativelanguage.googleapis.com/v1beta", "/v1beta/models/current-model:generateContent", "gemini"},
		{"Gemini model replacement", "https://gateway.example/prefix/models/old-model:generateContent?option=1", "/prefix/models/current-model:generateContent", "gemini"},
		{"Gemini compatibility", "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions", "/v1beta/openai/chat/completions", "openai"},
		{"Claude-named compatible gateway", "https://claude.example/v1/chat/completions", "/v1/chat/completions", "openai"},
		{"Local compatible gateway", "http://127.0.0.1:9999/v1/chat/completions", "/v1/chat/completions", "openai"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			transport := protocolTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Method != "POST" || r.URL.Path != tc.path {
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				if strings.Contains(tc.endpoint, "option=1") && r.URL.Query().Get("option") != "1" {
					t.Fatal("lost query option")
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				response := `{"choices":[{"message":{"content":"Hello world"}}]}`
				switch tc.protocol {
				case "anthropic":
					if r.Header.Get("x-api-key") != "test-key" || r.Header.Get("anthropic-version") != "2023-06-01" {
						t.Fatal("missing native Claude headers")
					}
					if r.Header.Get("Authorization") != "" || r.URL.Query().Get("key") != "" {
						t.Fatal("wrong authentication protocol")
					}
					if body["system"] != "system context" || body["model"] != "current-model" || body["max_tokens"].(float64) <= 0 {
						t.Fatalf("invalid Claude request: %#v", body)
					}
					messages := body["messages"].([]any)
					if len(messages) != 3 || messages[0].(map[string]any)["role"] != "user" {
						t.Fatal("system leaked into messages or conversation lost")
					}
					response = `{"content":[{"type":"text","text":"Hello "},{"type":"text","text":"world"}]}`
				case "gemini":
					if r.URL.Query().Get("key") != "test-key" || r.Header.Get("Authorization") != "" {
						t.Fatal("wrong native Gemini authentication")
					}
					if body["systemInstruction"] == nil || len(body["contents"].([]any)) != 3 {
						t.Fatal("missing system instruction or conversation")
					}
					response = `{"candidates":[{"content":{"parts":[{"text":"private reasoning","thought":true},{"text":"Hello "},{"text":"world"}]},"finishReason":"STOP"}]}`
				case "openai":
					if r.Header.Get("Authorization") != "Bearer test-key" || r.URL.Query().Get("key") != "" {
						t.Fatal("compatible endpoint received native authentication")
					}
					if body["contents"] != nil || len(body["messages"].([]any)) != 4 {
						t.Fatal("wrong compatible request body")
					}
				}
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(response))}, nil
			})
			client := NewClientWithHTTPClient(ClientConfig{Endpoint: tc.endpoint, Model: "current-model", APIKey: "test-key"}, &http.Client{Transport: transport})
			result, err := client.RequestWithMessages([]map[string]string{
				{"role": "system", "content": "system context"}, {"role": "user", "content": "first"},
				{"role": "assistant", "content": "previous answer"}, {"role": "user", "content": "second"},
			})
			if err != nil || result.Content != "Hello world" || string(result.FormatUsed) != tc.protocol || calls != 1 {
				t.Fatalf("result=%+v, err=%v calls=%d", result, err, calls)
			}
			if tc.protocol == "gemini" && result.Thinking != "private reasoning" {
				t.Fatal("thinking mixed with answer")
			}
		})
	}
}

func TestKnownProtocolErrorsAreNotRetriedAsOtherFormats(t *testing.T) {
	for _, endpoint := range []string{"https://api.anthropic.com/v1/messages", "https://generativelanguage.googleapis.com/v1beta", "https://gateway.example/v1/chat/completions"} {
		for status, code := range map[int]string{401: ErrorCodeAuthenticationFailed, 429: ErrorCodeRateLimited, 413: ErrorCodeRequestTooLarge, 503: ErrorCodeProviderUnavailable} {
			calls := 0
			client := NewClientWithHTTPClient(ClientConfig{Endpoint: endpoint, Model: "model", APIKey: "secret"}, &http.Client{Transport: protocolTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"error":{"message":"private provider text"}}`))}, nil
			})})
			_, err := client.Request("system", "user")
			public := ClassifyUserFacingError(err)
			if err == nil || calls != 1 || public.Code != code || strings.Contains(public.Message, "private") {
				t.Fatalf("%s status=%d calls=%d code=%s error=%v", endpoint, status, calls, public.Code, err)
			}
		}
	}
}

func TestRequestCancellationReachesEveryProtocolWithoutFallback(t *testing.T) {
	for _, endpoint := range []string{"https://api.anthropic.com/v1/messages", "https://generativelanguage.googleapis.com/v1beta", "http://localhost:11434/api/chat", "https://example.com/v1/chat/completions", "https://example.com/custom"} {
		t.Run(endpoint, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			entered := make(chan struct{})
			calls := 0
			client := NewClientWithHTTPClient(ClientConfig{Endpoint: endpoint, Model: "model"}, &http.Client{Transport: protocolTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				close(entered)
				<-r.Context().Done()
				return nil, r.Context().Err()
			})})
			done := make(chan error, 1)
			go func() {
				_, err := client.RequestWithMessagesContext(ctx, []map[string]string{{"role": "user", "content": "question"}})
				done <- err
			}()
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("provider not called")
			}
			cancel()
			select {
			case err := <-done:
				if !errors.Is(err, context.Canceled) || calls != 1 {
					t.Fatalf("err=%v calls=%d", err, calls)
				}
			case <-time.After(time.Second):
				t.Fatal("provider did not cancel")
			}
		})
	}
}

func TestChatRequestRegistryCancellationOrder(t *testing.T) {
	var registry ChatRequestRegistry
	registry.Cancel(1, "early")
	ctx, finish := registry.Begin(context.Background(), 1, "early")
	defer finish()
	if ctx.Err() != context.Canceled {
		t.Fatal("early cancel lost")
	}
	active, done := registry.Begin(context.Background(), 1, "active")
	other, otherDone := registry.Begin(context.Background(), 2, "active")
	defer otherDone()
	registry.Cancel(1, "active")
	if active.Err() != context.Canceled || other.Err() != nil {
		t.Fatal("cancellation crossed session boundary")
	}
	if registry.RunIfActive(active, func() { t.Fatal("cancelled work executed") }) {
		t.Fatal("cancelled request accepted")
	}
	done()
	replay, replayDone := registry.Begin(context.Background(), 1, "active")
	defer replayDone()
	if replay.Err() != context.Canceled {
		t.Fatal("finished request replayed")
	}
	completed, complete := registry.Begin(context.Background(), 1, "complete")
	if !registry.RunIfActive(completed, func() {}) {
		t.Fatal("active completion rejected")
	}
	complete()
	registry.Cancel(1, "complete")
	// Expired tombstones are reclaimed, while active entries survive pruning.
	registry.mu.Lock()
	registry.requests[chatRequestKey{1, "early"}].expires = time.Now().Add(-time.Second)
	registry.mu.Unlock()
	fresh, freshDone := registry.Begin(context.Background(), 1, "early")
	defer freshDone()
	if fresh.Err() != nil || other.Err() != nil {
		t.Fatal("incorrect expiration cleanup")
	}
}

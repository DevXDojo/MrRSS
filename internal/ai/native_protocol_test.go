package ai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
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

// observeProtocolBody wraps the real transport body and signals only after bytes
// have reached the AI client's ReadAll, not merely after response headers arrive.
type observeProtocolBody struct {
	io.ReadCloser
	read chan struct{}
	once sync.Once
}

func (b *observeProtocolBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if n > 0 {
		b.once.Do(func() {
			select {
			case b.read <- struct{}{}:
			default:
			}
		})
	}
	return n, err
}

func TestRequestCancellationWhileConnecting(t *testing.T) {
	var providerCalls, dialCalls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerCalls.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(provider.Close)
	ctx, cancel := context.WithCancel(context.Background())
	entered, release := make(chan struct{}), make(chan struct{})
	var enteredOnce sync.Once
	var dialing sync.WaitGroup
	transport := &http.Transport{DialContext: func(dialCtx context.Context, network, address string) (net.Conn, error) {
		dialCalls.Add(1)
		dialing.Add(1)
		defer dialing.Done()
		enteredOnce.Do(func() { close(entered) })
		// A transport dial may outlive a cancelled request for connection reuse.
		// Keep an explicit cleanup gate so this test never leaves that dial blocked.
		select {
		case <-dialCtx.Done():
		case <-release:
		}
		return nil, context.Canceled
	}}
	client := NewClientWithHTTPClient(ClientConfig{Endpoint: provider.URL + "/v1/chat/completions", Model: "model"}, &http.Client{Transport: transport})
	result, finished := make(chan error, 1), make(chan struct{})
	go func() {
		defer close(finished)
		_, err := client.RequestWithMessagesContext(ctx, []map[string]string{{"role": "user", "content": "question"}})
		result <- err
	}()
	t.Cleanup(func() {
		cancel()
		close(release)
		transport.CloseIdleConnections()
		select {
		case <-finished:
		case <-time.After(3 * time.Second):
			t.Error("request goroutine did not exit")
		}
		// All gates are open before waiting, including on an assertion failure.
		dialing.Wait()
	})
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("connection attempt did not start")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("connection cancellation error=%v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("request did not cancel while connecting")
	}
	if dialCalls.Load() != 1 || providerCalls.Load() != 0 {
		t.Fatalf("dial calls=%d provider calls=%d", dialCalls.Load(), providerCalls.Load())
	}
}

func TestRequestCancellationWhileReadingResponseBody(t *testing.T) {
	endpoint := "http://provider.invalid/custom"
	if DetectAPIProvider(endpoint) != "unknown" {
		t.Fatal("test must exercise the fallback path")
	}
	var providerCalls, transportCalls atomic.Int32
	release, providerCancelled, bodyRead := make(chan struct{}), make(chan struct{}), make(chan struct{}, 1)
	var cancelledOnce sync.Once
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		providerCalls.Add(1)
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		// This is deliberately incomplete JSON: the reader must wait for more body.
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"partial`))
		w.(http.Flusher).Flush()
		select {
		case <-r.Context().Done():
			cancelledOnce.Do(func() { close(providerCancelled) })
		case <-release:
		}
	}))
	ctx, cancel := context.WithCancel(context.Background())
	dialer := &net.Dialer{}
	transport := &http.Transport{DialContext: func(dialCtx context.Context, network, address string) (net.Conn, error) {
		return dialer.DialContext(dialCtx, network, provider.Listener.Addr().String())
	}}
	// The synthetic hostname selects the unknown-protocol fallback path, while
	// DialContext sends every request only to this local httptest server.
	client := NewClientWithHTTPClient(ClientConfig{Endpoint: endpoint, Model: "model"}, &http.Client{Transport: protocolTransport(func(r *http.Request) (*http.Response, error) {
		transportCalls.Add(1)
		response, err := transport.RoundTrip(r)
		if err == nil {
			response.Body = &observeProtocolBody{ReadCloser: response.Body, read: bodyRead}
		}
		return response, err
	})})
	result, finished := make(chan error, 1), make(chan struct{})
	go func() {
		defer close(finished)
		_, err := client.RequestWithMessagesContext(ctx, []map[string]string{{"role": "user", "content": "question"}})
		result <- err
	}()
	t.Cleanup(func() {
		cancel()
		close(release)
		transport.CloseIdleConnections()
		provider.CloseClientConnections()
		provider.Close()
		select {
		case <-finished:
		case <-time.After(3 * time.Second):
			t.Error("body reader goroutine did not exit")
		}
	})
	select {
	case <-bodyRead:
	case <-time.After(3 * time.Second):
		t.Fatal("client did not read the partial response body")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("body cancellation error=%v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("response body read did not cancel")
	}
	select {
	case <-providerCancelled:
	case <-time.After(3 * time.Second):
		t.Fatal("provider did not observe cancellation during body streaming")
	}
	if transportCalls.Load() != 1 || providerCalls.Load() != 1 {
		t.Fatalf("cancelled body read retried: transport calls=%d provider calls=%d", transportCalls.Load(), providerCalls.Load())
	}
}

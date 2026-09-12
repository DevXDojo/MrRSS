package translation

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
	"unicode/utf8"
)

const testEdgeToken = "e30.e30.c2lnbmF0dXJl" // Synthetic JWT-shaped fixture, no credentials.

func edgeResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func TestEdgeTranslationWithoutAzureCredentials(t *testing.T) {
	settings := &mockSettingsProvider{settings: map[string]string{"translation_provider": "microsoft_edge"}}
	dynamic := NewDynamicTranslator(settings)
	provider, err := dynamic.getProvider()
	if err != nil {
		t.Fatal(err)
	}
	edge, ok := provider.(*edgeProvider)
	if !ok || !provider.IsAvailable() || provider.Name() != "microsoft_edge" {
		t.Fatalf("wrong provider: %T", provider)
	}
	if again, err := dynamic.getProvider(); err != nil || again != edge {
		t.Fatal("provider must be reused for its token cache")
	}
	authCalls, translateCalls := 0, 0
	edge.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("Ocp-Apim-Subscription-Key") != "" {
			t.Error("unexpected Azure key")
		}
		if req.URL.String() == edgeAuthURL {
			authCalls++
			if req.Method != http.MethodGet || req.Header.Get("Authorization") != "" {
				t.Error("invalid token request")
			}
			return edgeResponse(200, testEdgeToken), nil
		}
		translateCalls++
		if req.URL.Host != "api-edge.cognitive.microsofttranslator.com" || req.Method != http.MethodPost || req.URL.Query().Get("to") != "zh-Hans" {
			t.Errorf("invalid translation destination: %v", req.URL)
		}
		if req.Header.Get("Authorization") != "Bearer "+testEdgeToken {
			t.Error("missing temporary bearer token")
		}
		var payload []map[string]string
		if err := json.NewDecoder(req.Body).Decode(&payload); err != nil || len(payload) != 1 || payload[0]["Text"] != "Hello" {
			t.Errorf("wrong payload: %#v, %v", payload, err)
		}
		return edgeResponse(200, `[{"translations":[{"text":"你好","to":"zh-Hans"}]}]`), nil
	})
	for i := 0; i < 2; i++ {
		got, err := dynamic.TranslateContext(context.Background(), " Hello\n", "zh")
		if err != nil || got != " 你好\n" {
			t.Fatalf("translation = %q, %v", got, err)
		}
	}
	if authCalls != 1 || translateCalls != 2 {
		t.Fatalf("requests = %d auth, %d translations", authCalls, translateCalls)
	}
	edge.tokenExpires = time.Now().Add(-time.Second)
	if _, err := dynamic.Translate("Hello", "zh"); err != nil {
		t.Fatal(err)
	}
	if authCalls != 2 {
		t.Error("expired token was reused")
	}
}

func TestEdgeAuthenticationRetriesAreBounded(t *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusTooManyRequests} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			provider, err := newEdgeProvider(nil)
			if err != nil {
				t.Fatal(err)
			}
			authCalls, translateCalls := 0, 0
			provider.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				if req.URL.String() == edgeAuthURL {
					authCalls++
					return edgeResponse(200, testEdgeToken), nil
				}
				translateCalls++
				return edgeResponse(status, "private article text and token details"), nil
			})
			result, err := provider.Translate(context.Background(), "hello", "zh")
			if err == nil || result != nil || strings.Contains(err.Error(), "private") {
				t.Fatalf("unsafe failure: %v, %v", result, err)
			}
			want := 1
			if status == http.StatusUnauthorized {
				want = 2
			}
			if authCalls != want || translateCalls != want {
				t.Fatalf("unexpected retry count: %d/%d", authCalls, translateCalls)
			}
		})
	}
}

func TestEdgeRejectsMalformedTokenAndRedirects(t *testing.T) {
	for _, body := range []string{"<html>private error details</html>", strings.Repeat("x", (16<<10)+1)} {
		provider, err := newEdgeProvider(nil)
		if err != nil {
			t.Fatal(err)
		}
		calls := 0
		provider.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
			calls++
			return edgeResponse(200, body), nil
		})
		if _, err := provider.Translate(context.Background(), "hello", "zh"); err == nil || strings.Contains(err.Error(), body) {
			t.Fatalf("malformed token was accepted or leaked: %v", err)
		}
		if calls != 1 {
			t.Error("translation requested with invalid token")
		}
	}
	provider, err := newEdgeProvider(nil)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	provider.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if req.URL.String() == edgeAuthURL {
			return edgeResponse(200, testEdgeToken), nil
		}
		if req.URL.Host != "api-edge.cognitive.microsofttranslator.com" {
			t.Error("redirect followed")
		}
		response := edgeResponse(http.StatusTemporaryRedirect, "")
		response.Header.Set("Location", "https://unrelated.example/collect")
		return response, nil
	})
	if _, err := provider.Translate(context.Background(), "hello", "zh"); err == nil {
		t.Error("redirect accepted")
	}
	if calls != 2 {
		t.Fatalf("redirect leaked a request: %d", calls)
	}
}

func TestEdgeCancellationWhileWaitingAndInFlight(t *testing.T) {
	provider, err := newEdgeProvider(nil)
	if err != nil {
		t.Fatal(err)
	}
	provider.gate <- struct{}{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := provider.Translate(ctx, "hello", "zh"); !errors.Is(err, context.Canceled) {
		t.Fatalf("queued cancellation: %v", err)
	}
	<-provider.gate
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	provider.client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		cancel()
		return nil, req.Context().Err()
	})
	if _, err := provider.Translate(ctx, "hello", "zh"); !errors.Is(err, context.Canceled) {
		t.Fatalf("request cancellation: %v", err)
	}
}

func TestEdgeLongTextSplittingPreservesUnicodeAndWhitespace(t *testing.T) {
	for _, text := range []string{
		strings.Repeat("😀", 6000),
		strings.Repeat("word ", 3000),
		strings.Repeat("中", 5001),
		strings.Repeat(" ", 10001),
		strings.Repeat("a", 4999) + "😀 end",
	} {
		chunks := splitEdgeText(text)
		if strings.Join(chunks, "") != text {
			t.Fatal("text was lost during splitting")
		}
		for _, chunk := range chunks {
			if chunk == "" || !utf8.ValidString(chunk) || len(utf16.Encode([]rune(chunk))) > 5000 {
				t.Fatal("invalid chunk boundary")
			}
		}
	}
}

func TestContextMarkdownDoesNotReportPartialSuccess(t *testing.T) {
	calls := 0
	failure := errors.New("provider temporarily unavailable")
	translator := &TestTranslator{TranslateFunc: func(text, _ string) (string, error) {
		calls++
		if calls == 2 {
			return "", failure
		}
		return "translated " + text, nil
	}}
	result, err := TranslateMarkdownPreservingStructureContext(context.Background(), "- one\n- two\n- three", translator, "zh")
	if !errors.Is(err, failure) || result != "" || calls != 2 {
		t.Fatalf("partial success returned: %q, %v, calls=%d", result, err, calls)
	}
}

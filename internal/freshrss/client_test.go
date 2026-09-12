package freshrss

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClientForProviderUsesCorrectGoogleReaderRoot(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		serverURL string
		wantRoot  string
	}{
		{
			name:      "FreshRSS appends greader endpoint",
			provider:  string(ProviderFreshRSS),
			serverURL: "https://freshrss.example.com/",
			wantRoot:  "https://freshrss.example.com/api/greader.php",
		},
		{
			name:      "Miniflux uses server root",
			provider:  string(ProviderMiniflux),
			serverURL: "https://miniflux.example.com/miniflux/",
			wantRoot:  "https://miniflux.example.com/miniflux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClientForProvider(tt.serverURL, "user", "password", tt.provider)
			if client.baseURL != tt.wantRoot {
				t.Fatalf("baseURL = %q, want %q", client.baseURL, tt.wantRoot)
			}
		})
	}
}

func TestGoogleReaderUpdatedTimeSupportsSecondsAndMilliseconds(t *testing.T) {
	if got := googleReaderUpdatedTime(1_710_000_300).Unix(); got != 1_710_000_300 {
		t.Fatalf("seconds timestamp = %d", got)
	}
	if got := googleReaderUpdatedTime(1_710_000_300_000).Unix(); got != 1_710_000_300 {
		t.Fatalf("milliseconds timestamp = %d", got)
	}
}

func TestMinifluxGetStreamContentsUsesItemIDsThenContents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/reader/api/0/stream/items/ids":
			if r.Method != http.MethodGet || r.URL.Query().Get("s") != "feed/42" {
				t.Fatalf("unexpected item IDs request: %s %s", r.Method, r.URL)
			}
			if got := r.Header.Get("Authorization"); got != "GoogleLogin auth=auth-token" {
				t.Fatalf("authorization = %q", got)
			}
			_, _ = w.Write([]byte(`{"itemRefs":[{"id":"123"}],"continuation":"next-page"}`))
		case "/reader/api/0/token":
			_, _ = w.Write([]byte("auth-token"))
		case "/reader/api/0/stream/items/contents":
			if r.Method != http.MethodPost {
				t.Fatalf("contents method = %s", r.Method)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if r.Form.Get("T") != "auth-token" || r.Form.Get("output") != "json" || r.Form.Get("i") != "123" {
				t.Fatalf("unexpected contents form: %v", r.Form)
			}
			_, _ = w.Write([]byte(`{"updated":1710000000,"items":[{"id":"123","title":"Entry","canonical":[{"href":"https://example.com/article"}],"summary":{"content":"body"},"published":1710000000,"updated":1710000001,"categories":["user/1/state/com.google/read"],"origin":{"streamId":"feed/42"}}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewClientForProvider(server.URL, "user", "password", string(ProviderMiniflux))
	client.authToken = "auth-token"
	result, err := client.GetStreamContents(context.Background(), "feed/42", nil, 100, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 1 || result.Items[0].URL != "https://example.com/article" {
		t.Fatalf("unexpected items: %#v", result.Items)
	}
	if result.Items[0].Updated.Unix() != 1710000001 {
		t.Fatalf("updated = %d", result.Items[0].Updated.Unix())
	}
	if result.Continuation != "next-page" {
		t.Fatalf("continuation = %q", result.Continuation)
	}
}

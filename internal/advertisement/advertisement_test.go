package advertisement

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

const basicAd = `{"id":"partner","title":{"en":"AI APIs","zh":"AI 接口"},"description":"Partner discount","url":"https://example.com/offer","coupon_code":"MRRSS"}`

func TestExtensibleDocument(t *testing.T) {
	data := `{"version":{"future":42},"layout":{"complex":true},"ads":[
		{"id":"invalid","title":"Invalid","url":"javascript:alert(1)"},
		{"id":"rich","title":{"en":{"rich":"new layout"}},"fallback":{"title":"Compatible title","description":{"zh":"基础说明"},"url":"https://example.com/rich"},"blocks":[
			{"type":"interactive","script":"ignored"},
			{"type":"text","text":{"zh":"补充说明","metadata":{"font":"new"}}},
			{"type":"list","items":["One",{"en":"Two"}]}
		],"creative":{"future":"fields"}},` + basicAd + `]}`
	ads, err := Parse([]byte(data))
	if err != nil || len(ads) != 2 {
		t.Fatalf("complex fields suppressed supported ads: %#v, %v", ads, err)
	}
	if ads[0].Title["en"] != "Compatible title" || ads[0].Description["zh"] != "基础说明" || len(ads[0].Blocks) != 2 {
		t.Fatalf("fallback/blocks not parsed: %#v", ads[0])
	}
}

func TestUnsafeLinksAndMalformedEntriesAreIsolated(t *testing.T) {
	for _, link := range []string{"javascript:alert(1)", "data:text/html,test", "http://example.com", "https://user:secret@example.com", "https:///path", "file:///tmp/a"} {
		data := `{"ads":[null,7,{"id":"unsafe","title":"Unsafe","url":"` + link + `"},` + basicAd + `]}`
		ads, err := Parse([]byte(data))
		if err != nil || len(ads) != 1 || ads[0].ID != "partner" {
			t.Fatalf("link %s was accepted or suppressed safe ad: %#v, %v", link, ads, err)
		}
	}
	if _, err := Parse([]byte(strings.Repeat(" ", maxDocumentSize+1))); err == nil {
		t.Fatal("oversized document accepted")
	}
}

func TestRemoteCacheOfflineFallbackAndRemoval(t *testing.T) {
	remote := `{"version":100,"unknown":{"future":[1,2]},"ads":[` + basicAd + `]}`
	var payload atomic.Value
	payload.Store(remote)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.RawQuery != "" || r.Header.Get("Cookie") != "" || r.Header.Get("Authorization") != "" {
			t.Error("reader data/credentials sent to remote source")
		}
		w.Write([]byte(payload.Load().(string)))
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "ads.json")
	service := New([]byte(`{"ads":[]}`))
	service.sourceURL = server.URL
	if ads := service.Get(context.Background(), path, server.Client()); len(ads) != 1 {
		t.Fatalf("remote ad not loaded: %#v", ads)
	}
	service.Get(context.Background(), path, server.Client())
	if calls.Load() != 1 {
		t.Fatal("fetched configuration more than once within cache interval")
	}
	stored, _ := os.ReadFile(path)
	if string(stored) != remote {
		t.Fatal("unknown future fields were discarded from cache")
	}
	remote = "invalid JSON"
	payload.Store(remote)
	offline := New([]byte(`{"ads":[]}`))
	offline.sourceURL = server.URL
	if ads := offline.Get(context.Background(), path, server.Client()); len(ads) != 1 {
		t.Fatal("invalid response discarded local ad")
	}
	stored, _ = os.ReadFile(path)
	if string(stored) == remote {
		t.Fatal("invalid response overwrote cache")
	}
	remote = `{"ads":[]}`
	payload.Store(remote)
	removed := New(nil)
	removed.sourceURL = server.URL
	if ads := removed.Get(context.Background(), path, server.Client()); len(ads) != 0 {
		t.Fatal("remote removal did not clear ad")
	}
	remote = "broken"
	if ads := New([]byte(`{"ads":[`+basicAd+`]}`)).Get(context.Background(), path, nil); len(ads) != 0 {
		t.Fatal("removed advertisement restored from bundled fallback")
	}
}

func TestBundledFallbackAndCancellation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ads.json")
	service := New([]byte(`{"ads":[` + basicAd + `]}`))
	if len(service.Get(context.Background(), path, nil)) != 1 {
		t.Fatal("bundled fallback not loaded")
	}
	ctx, cancel := context.WithCancel(context.Background())
	service.gate <- struct{}{}
	cancel()
	if ads := service.Get(ctx, path, nil); len(ads) != 0 {
		t.Fatal("cancelled caller waited for configuration lock")
	}
	<-service.gate
	if ads := New(nil).Get(context.Background(), path, nil); len(ads) != 0 {
		t.Fatal("default should have no advertisements")
	}
}

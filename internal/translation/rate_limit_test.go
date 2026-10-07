package translation

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

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
	if calls != 1 {
		t.Fatalf("provider was contacted %d times during cooldown", calls)
	}
	translator.retryAt = time.Now().Add(-time.Second)
	_, _ = translator.Translate("Hello", "zh")
	if calls != 2 {
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

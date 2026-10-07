package translation

import (
	translator "MrRSS/internal/translation"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTranslationRateLimitResponse(t *testing.T) {
	w := httptest.NewRecorder()
	translationError(w, fmt.Errorf("provider: %w", &translator.RateLimitError{RetryAfter: 90 * time.Second}))
	if w.Code != 429 || w.Header().Get("Retry-After") != "90" {
		t.Fatalf("status=%d retry=%s", w.Code, w.Header().Get("Retry-After"))
	}
}

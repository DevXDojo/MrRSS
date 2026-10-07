package translation

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RateLimitError preserves a provider's retry delay across the HTTP API.
type RateLimitError struct{ RetryAfter time.Duration }

func (e *RateLimitError) Error() string { return "translation service rate limit reached; retry later" }

func translationRetryDelay(value string) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && seconds > 0 {
		return time.Duration(min(seconds, 86400)) * time.Second
	}
	if deadline, err := http.ParseTime(value); err == nil && time.Until(deadline) > 0 {
		return min(time.Until(deadline), 24*time.Hour)
	}
	return time.Minute
}

func (e *RateLimitError) RetryAfterHeader() string {
	return fmt.Sprint(max(1, int64(e.RetryAfter.Seconds()+0.999)))
}

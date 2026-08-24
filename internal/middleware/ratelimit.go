package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiterConfig holds configuration for rate limiting.
type RateLimiterConfig struct {
	// RequestsPerSecond is the maximum requests per second per IP.
	RequestsPerSecond int
	// BurstSize is the maximum burst size.
	BurstSize int
	// CleanupInterval is how often to clean up old entries.
	CleanupInterval time.Duration
}

// DefaultRateLimiterConfig returns default rate limiter configuration.
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerSecond: 10,
		BurstSize:         20,
		CleanupInterval:   time.Minute,
	}
}

type visitor struct {
	tokens    float64
	lastVisit time.Time
}

// RateLimiter returns a rate limiting middleware.
func RateLimiter(config RateLimiterConfig) Middleware {
	visitors := make(map[string]*visitor)
	var mu sync.RWMutex

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(config.CleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastVisit) > config.CleanupInterval {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)

			mu.Lock()
			v, exists := visitors[ip]
			now := time.Now()

			if !exists {
				visitors[ip] = &visitor{
					tokens:    float64(config.BurstSize - 1),
					lastVisit: now,
				}
				mu.Unlock()
				next.ServeHTTP(w, r)
				return
			}

			// Refill tokens based on time elapsed
			elapsed := now.Sub(v.lastVisit).Seconds()
			v.tokens += elapsed * float64(config.RequestsPerSecond)
			if v.tokens > float64(config.BurstSize) {
				v.tokens = float64(config.BurstSize)
			}
			v.lastVisit = now

			if v.tokens < 1 {
				mu.Unlock()
				w.Header().Set("Retry-After", "1")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			v.tokens--
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP from the request.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

package middleware

import (
	"net/http"
	"strings"
)

// CORSConfig holds the configuration for CORS middleware.
type CORSConfig struct {
	// AllowedOrigins is a list of allowed origins. Use "*" to allow all.
	AllowedOrigins []string
	// AllowedMethods is a list of allowed HTTP methods.
	AllowedMethods []string
	// AllowedHeaders is a list of allowed headers.
	AllowedHeaders []string
	// ExposedHeaders is a list of headers exposed to the client.
	ExposedHeaders []string
	// AllowCredentials indicates whether credentials are allowed.
	AllowCredentials bool
	// MaxAge is the max age for preflight cache in seconds.
	MaxAge int
}

// DefaultCORSConfig returns a default CORS configuration.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		MaxAge:         86400, // 24 hours
	}
}

// CORS returns a CORS middleware with default configuration.
func CORS() Middleware {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig returns a CORS middleware with custom configuration.
func CORSWithConfig(config CORSConfig) Middleware {
	allowedOrigins := strings.Join(config.AllowedOrigins, ", ")
	allowedMethods := strings.Join(config.AllowedMethods, ", ")
	allowedHeaders := strings.Join(config.AllowedHeaders, ", ")
	exposedHeaders := strings.Join(config.ExposedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			if origin != "" {
				allowed := false
				for _, o := range config.AllowedOrigins {
					if o == "*" || o == origin {
						allowed = true
						break
					}
				}

				if allowed {
					if config.AllowedOrigins[0] == "*" {
						w.Header().Set("Access-Control-Allow-Origin", "*")
					} else {
						w.Header().Set("Access-Control-Allow-Origin", origin)
					}

					if config.AllowCredentials {
						w.Header().Set("Access-Control-Allow-Credentials", "true")
					}
				}
			} else {
				// No origin header, set default
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
			}

			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)

			if exposedHeaders != "" {
				w.Header().Set("Access-Control-Expose-Headers", exposedHeaders)
			}

			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", string(rune(config.MaxAge)))
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

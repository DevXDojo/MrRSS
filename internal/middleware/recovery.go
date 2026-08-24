package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

// Recovery returns a middleware that recovers from panics.
func Recovery() Middleware {
	return RecoveryWithConfig(RecoveryConfig{})
}

// RecoveryConfig holds configuration for recovery middleware.
type RecoveryConfig struct {
	// LogFunc is a custom log function for panic messages.
	LogFunc func(format string, v ...interface{})
	// EnableStackTrace enables stack trace logging.
	EnableStackTrace bool
	// OnPanic is a callback function called when panic occurs.
	OnPanic func(w http.ResponseWriter, r *http.Request, err interface{})
}

// RecoveryWithConfig returns a recovery middleware with custom configuration.
func RecoveryWithConfig(config RecoveryConfig) Middleware {
	if config.LogFunc == nil {
		config.LogFunc = log.Printf
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic
					config.LogFunc("PANIC recovered: %v", err)

					if config.EnableStackTrace {
						config.LogFunc("Stack trace:\n%s", debug.Stack())
					}

					// Call custom panic handler if provided
					if config.OnPanic != nil {
						config.OnPanic(w, r, err)
						return
					}

					// Default response
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error"))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

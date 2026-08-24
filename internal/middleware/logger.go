package middleware

import (
	"log"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// Logger returns a middleware that logs HTTP requests.
func Logger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			log.Printf("[%s] %s %s %d %v",
				r.Method,
				r.URL.Path,
				r.RemoteAddr,
				rw.statusCode,
				duration,
			)
		})
	}
}

// LoggerWithConfig returns a logger middleware with custom configuration.
type LoggerConfig struct {
	// SkipPaths is a list of paths to skip logging for.
	SkipPaths []string
	// LogFunc is a custom log function. Defaults to log.Printf.
	LogFunc func(format string, v ...interface{})
}

func LoggerWithConfig(config LoggerConfig) Middleware {
	if config.LogFunc == nil {
		config.LogFunc = log.Printf
	}

	skipMap := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipMap[path] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip logging for certain paths
			if skipMap[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			rw := newResponseWriter(w)

			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			config.LogFunc("[%s] %s %s %d %v",
				r.Method,
				r.URL.Path,
				r.RemoteAddr,
				rw.statusCode,
				duration,
			)
		})
	}
}

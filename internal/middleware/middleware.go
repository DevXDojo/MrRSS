// Package middleware provides HTTP middleware components for the application.
package middleware

import (
	"net/http"
)

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain creates a new middleware chain from the given middlewares.
// Middlewares are applied in reverse order, so the first middleware
// in the list will be the outermost handler.
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Apply applies a chain of middlewares to an http.Handler.
func Apply(h http.Handler, middlewares ...Middleware) http.Handler {
	return Chain(middlewares...)(h)
}

// ApplyFunc applies a chain of middlewares to an http.HandlerFunc.
func ApplyFunc(h http.HandlerFunc, middlewares ...Middleware) http.Handler {
	return Apply(h, middlewares...)
}

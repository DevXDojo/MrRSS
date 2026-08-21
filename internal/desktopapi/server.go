// Package desktopapi exposes the desktop application's API on a loopback-only
// HTTP listener for local integrations such as the mrrss-assistant skill.
package desktopapi

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

const DefaultAddress = "127.0.0.1:1234"

// Server owns the desktop API listener and its lifecycle.
type Server struct {
	httpServer *http.Server
	listener   net.Listener
	errors     chan error
}

// Start starts an HTTP server on a loopback address. It returns after the
// listener has been created, so address conflicts are reported synchronously.
func Start(address string, handler http.Handler) (*Server, error) {
	if handler == nil {
		return nil, errors.New("desktop API handler is required")
	}
	if err := validateLoopbackAddress(address); err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listen for desktop API on %s: %w", address, err)
	}

	httpServer := &http.Server{
		Addr:              listener.Addr().String(),
		Handler:           protectLocalAPI(handler),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	server := &Server{
		httpServer: httpServer,
		listener:   listener,
		errors:     make(chan error, 1),
	}

	go func() {
		err := httpServer.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		server.errors <- err
		close(server.errors)
	}()

	return server, nil
}

// Address returns the listener's resolved address.
func (s *Server) Address() string {
	return s.listener.Addr().String()
}

// Errors reports the terminal serving error. A graceful shutdown reports nil.
func (s *Server) Errors() <-chan error {
	return s.errors
}

// Shutdown gracefully stops accepting desktop API requests.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func validateLoopbackAddress(address string) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid desktop API address %q: %w", address, err)
	}
	if port == "" {
		return fmt.Errorf("invalid desktop API address %q: port is required", address)
	}

	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("desktop API address %q must use a loopback IP", address)
	}
	return nil
}

func protectLocalAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Sec-Fetch-Site")

		if !isLoopbackHost(r.Host) {
			http.Error(w, "desktop API requires a loopback Host header", http.StatusForbidden)
			return
		}
		if r.Header.Get("Origin") != "" || strings.EqualFold(r.Header.Get("Sec-Fetch-Site"), "cross-site") {
			http.Error(w, "cross-origin browser requests are not allowed", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isLoopbackHost(hostPort string) bool {
	host := hostPort
	if parsedHost, _, err := net.SplitHostPort(hostPort); err == nil {
		host = parsedHost
	}
	host = strings.Trim(host, "[]")
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

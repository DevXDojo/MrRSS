package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

const waitTimeout = 90 * time.Second

func frontendClient(serverURL string) *http.Client {
	transport := &http.Transport{Proxy: nil}
	parsed, _ := url.Parse(serverURL)
	if parsed != nil && (parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1") {
		dialer := &net.Dialer{Timeout: time.Second}
		transport.DialContext = func(ctx context.Context, _ string, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp4", address)
		}
	}
	return &http.Client{Timeout: time.Second, Transport: transport}
}

func main() {
	serverURL := os.Getenv("FRONTEND_DEVSERVER_URL")
	if serverURL == "" {
		port := os.Getenv("WAILS_VITE_PORT")
		if port == "" {
			port = "5173"
		}
		serverURL = "http://localhost:" + port
	}

	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()

	// Probe the same address family used by Wails' frontend proxy. A Vite
	// listener on ::1 alone must not count as ready for that IPv4 proxy.
	client := frontendClient(serverURL)
	defer client.CloseIdleConnections()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid frontend development server URL: %v\n", err)
			os.Exit(1)
		}
		response, err := client.Do(request)
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode < http.StatusInternalServerError {
				fmt.Printf("Frontend development server is ready at %s\n", serverURL)
				return
			}
		}

		select {
		case <-ctx.Done():
			fmt.Fprintf(os.Stderr, "frontend development server did not become ready at %s within %s\n", serverURL, waitTimeout)
			os.Exit(1)
		case <-ticker.C:
		}
	}
}

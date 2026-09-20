package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const waitTimeout = 90 * time.Second

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

	client := &http.Client{
		Timeout:   time.Second,
		Transport: &http.Transport{Proxy: nil},
	}
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

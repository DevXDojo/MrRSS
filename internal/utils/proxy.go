package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// BuildProxyURL constructs a proxy URL from settings
func BuildProxyURL(proxyType, proxyHost, proxyPort, username, password string) string {
	if proxyHost == "" || proxyPort == "" {
		return ""
	}

	// Build auth string if username is provided
	auth := ""
	if username != "" {
		if password != "" {
			auth = username + ":" + password + "@"
		} else {
			auth = username + "@"
		}
	}

	return fmt.Sprintf("%s://%s%s:%s", proxyType, auth, proxyHost, proxyPort)
}

// CreateHTTPClient creates an HTTP client with optional proxy support
// This is the canonical implementation with proper TLS config and connection pooling
func CreateHTTPClient(proxyURL string, timeout time.Duration) (*http.Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:        50, // Reduced from 100 to prevent connection exhaustion
		MaxIdleConnsPerHost: 5,  // Reduced from 10 to limit connections per host
		IdleConnTimeout:     90 * time.Second,
		// Disable HTTP/2 for RSS feeds - it can cause performance issues
		// HTTP/1.1 is more reliable and faster for simple RSS feed fetching
		ForceAttemptHTTP2: false,
		// Write buffer size
		WriteBufferSize: 32 * 1024, // 32KB
		// Read buffer size
		ReadBufferSize: 32 * 1024, // 32KB
	}

	// Configure proxy if provided
	if proxyURL != "" {
		parsedProxy, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(parsedProxy)
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	return client, nil
}

// RoundTripFunc is an adapter to allow the use of ordinary functions as http.RoundTripper
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements http.RoundTripper
func (rt RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

// UserAgentTransport wraps an http.RoundTripper to add User-Agent headers
type UserAgentTransport struct {
	Original  http.RoundTripper
	userAgent string
}

// RoundTrip implements http.RoundTripper with automatic Cloudflare bypass
func (t *UserAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// First attempt: Use browser User-Agent
	return t.roundTripWithRetry(req, true)
}

// roundTripWithRetry performs the actual HTTP request with optional retry logic
func (t *UserAgentTransport) roundTripWithRetry(req *http.Request, useBrowserUA bool) (*http.Response, error) {
	if useBrowserUA {
		// Use browser-like headers
		req.Header.Set("User-Agent", t.userAgent)
		req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml, application/atom+xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
		req.Header.Set("DNT", "1")

		// Add Sec-Fetch headers to mimic modern browsers
		if req.Header.Get("Sec-Fetch-Dest") == "" {
			req.Header.Set("Sec-Fetch-Dest", "document")
		}
		if req.Header.Get("Sec-Fetch-Mode") == "" {
			req.Header.Set("Sec-Fetch-Mode", "navigate")
		}
		if req.Header.Get("Sec-Fetch-Site") == "" {
			req.Header.Set("Sec-Fetch-Site", "none")
		}
		if req.Header.Get("Sec-Fetch-User") == "" {
			req.Header.Set("Sec-Fetch-User", "?1")
		}

		if req.Header.Get("Cache-Control") == "" {
			req.Header.Set("Cache-Control", "max-age=0")
		}
	} else {
		// Use simple curl-like User-Agent to bypass Cloudflare
		req.Header.Set("User-Agent", "curl/8.11.1")
		req.Header.Set("Accept", "*/*")

		// Remove browser-specific headers that might trigger Cloudflare
		req.Header.Del("Sec-Fetch-Dest")
		req.Header.Del("Sec-Fetch-Mode")
		req.Header.Del("Sec-Fetch-Site")
		req.Header.Del("Sec-Fetch-User")
		req.Header.Del("Cache-Control")
		req.Header.Del("DNT")
		req.Header.Del("Accept-Language")
	}

	// Perform the request
	resp, err := t.Original.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// If this is the first attempt and we got a 403, check for Cloudflare challenge
	if useBrowserUA && resp.StatusCode == 403 {
		// Read the response body to check for Cloudflare challenge page
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			// If we can't read the body, just return the 403 response as-is
			return resp, fmt.Errorf("failed to read 403 response body: %w", err)
		}

		bodyStr := string(body)

		// Check for Cloudflare challenge page indicators
		isCloudflare := strings.Contains(bodyStr, "Checking your browser") ||
			strings.Contains(bodyStr, "Cloudflare") ||
			strings.Contains(bodyStr, "cf_chl_opt") ||
			strings.Contains(bodyStr, "challenge-platform") ||
			strings.Contains(bodyStr, "jschl-answer") ||
			strings.Contains(bodyStr, "cf-browser-verification")

		if isCloudflare {
			// Detected Cloudflare challenge page - retry with simple User-Agent
			// Create a new request for retry
			retryResp, retryErr := t.roundTripWithRetry(req, false)
			if retryErr != nil {
				// Retry failed, return original response (create new body reader)
				resp.Body = io.NopCloser(bytes.NewReader(body))
				return resp, retryErr
			}

			// Retry succeeded - return the new response
			return retryResp, nil
		}

		// Not a Cloudflare challenge - restore body and return original response
		resp.Body = io.NopCloser(bytes.NewReader(body))
	}

	return resp, nil
}

// CreateHTTPClientWithUserAgent creates an HTTP client with a custom User-Agent
// This is important because some RSS servers block requests without a proper User-Agent
func CreateHTTPClientWithUserAgent(proxyURL string, timeout time.Duration, userAgent string) (*http.Client, error) {
	baseClient, err := CreateHTTPClient(proxyURL, timeout)
	if err != nil {
		return nil, err
	}

	// Wrap the transport to add User-Agent to all requests
	baseClient.Transport = &UserAgentTransport{
		Original:  baseClient.Transport,
		userAgent: userAgent,
	}

	return baseClient, nil
}

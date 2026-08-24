// Package httputil provides HTTP client utilities including proxy support,
// custom transports, and Cloudflare bypass functionality.
package httputil

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// InsecureSkipTLSVerifyEnv enables TLS certificate verification bypass for
// private LAN services such as self-signed Ollama endpoints. It is intentionally
// opt-in and environment-scoped.
const InsecureSkipTLSVerifyEnv = "MRRSS_INSECURE_SKIP_TLS_VERIFY"

// BuildProxyURL constructs a proxy URL from settings.
func BuildProxyURL(proxyType, proxyHost, proxyPort, username, password string) string {
	if proxyHost == "" || proxyPort == "" {
		return ""
	}

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

// CreateHTTPClient creates an HTTP client with optional proxy support.
func CreateHTTPClient(proxyURL string, timeout time.Duration) (*http.Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: insecureSkipTLSVerifyEnabled(),
		},
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     90 * time.Second,
		ForceAttemptHTTP2:   false,
		WriteBufferSize:     32 * 1024,
		ReadBufferSize:      32 * 1024,
	}

	if proxyURL != "" {
		parsedProxy, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(parsedProxy)
	}

	return &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}, nil
}

func insecureSkipTLSVerifyEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(InsecureSkipTLSVerifyEnv))) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

// CreateHTTPClientWithUserAgent creates an HTTP client with custom User-Agent.
func CreateHTTPClientWithUserAgent(proxyURL string, timeout time.Duration, userAgent string) (*http.Client, error) {
	baseClient, err := CreateHTTPClient(proxyURL, timeout)
	if err != nil {
		return nil, err
	}

	baseClient.Transport = &UserAgentTransport{
		Original:  baseClient.Transport,
		userAgent: userAgent,
	}

	return baseClient, nil
}

// RoundTripFunc is an adapter for ordinary functions as http.RoundTripper.
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements http.RoundTripper.
func (rt RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

// UserAgentTransport wraps http.RoundTripper to add User-Agent headers.
type UserAgentTransport struct {
	Original  http.RoundTripper
	userAgent string
}

// RoundTrip implements http.RoundTripper with automatic Cloudflare bypass.
func (t *UserAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.roundTripWithRetry(req, true)
}

func (t *UserAgentTransport) roundTripWithRetry(req *http.Request, useBrowserUA bool) (*http.Response, error) {
	if useBrowserUA {
		req.Header.Set("User-Agent", t.userAgent)
		req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml, application/atom+xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
		req.Header.Set("DNT", "1")

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
		req.Header.Set("User-Agent", "curl/8.11.1")
		req.Header.Set("Accept", "*/*")
		req.Header.Del("Sec-Fetch-Dest")
		req.Header.Del("Sec-Fetch-Mode")
		req.Header.Del("Sec-Fetch-Site")
		req.Header.Del("Sec-Fetch-User")
		req.Header.Del("Cache-Control")
		req.Header.Del("DNT")
		req.Header.Del("Accept-Language")
	}

	resp, err := t.Original.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if useBrowserUA && resp.StatusCode == 403 {
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			return resp, fmt.Errorf("failed to read 403 response body: %w", err)
		}

		bodyStr := string(body)

		isCloudflare := strings.Contains(bodyStr, "Checking your browser") ||
			strings.Contains(bodyStr, "Cloudflare") ||
			strings.Contains(bodyStr, "cf_chl_opt") ||
			strings.Contains(bodyStr, "challenge-platform") ||
			strings.Contains(bodyStr, "jschl-answer") ||
			strings.Contains(bodyStr, "cf-browser-verification")

		if isCloudflare {
			retryResp, retryErr := t.roundTripWithRetry(req, false)
			if retryErr != nil {
				resp.Body = io.NopCloser(bytes.NewReader(body))
				return resp, retryErr
			}
			return retryResp, nil
		}

		resp.Body = io.NopCloser(bytes.NewReader(body))
	}

	return resp, nil
}

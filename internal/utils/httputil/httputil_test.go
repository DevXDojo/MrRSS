package httputil

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type refererFallbackBody struct {
	io.Reader
	closed bool
}

func (b *refererFallbackBody) Close() error {
	b.closed = true
	return nil
}

func TestDoWithRefererFallbackConditions(t *testing.T) {
	for _, tc := range []struct {
		name      string
		method    string
		referer   string
		status    int
		finalCode int
		wantRetry bool
	}{
		{"success", http.MethodGet, "https://article.example", 200, 200, false},
		{"forbidden", http.MethodGet, "https://article.example", 403, 200, true},
		{"still forbidden", http.MethodGet, "https://article.example", 403, 403, true},
		{"no referer", http.MethodGet, "", 403, 403, false},
		{"not GET", http.MethodHead, "https://article.example", 403, 403, false},
		{"not found", http.MethodGet, "https://article.example", 404, 404, false},
		{"rate limited", http.MethodGet, "https://article.example", 429, 429, false},
		{"server error", http.MethodGet, "https://article.example", 500, 500, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			firstBody := &refererFallbackBody{Reader: strings.NewReader("first")}
			calls := 0
			client := &http.Client{Transport: RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return &http.Response{StatusCode: tc.status, Body: firstBody, Header: make(http.Header), Request: req}, nil
				}
				if !firstBody.closed {
					t.Error("first response body was not closed before retry")
				}
				if req.Header.Get("Referer") != "" || req.Header.Get("User-Agent") != "media-test" {
					t.Errorf("unexpected retry headers: %v", req.Header)
				}
				return &http.Response{StatusCode: tc.finalCode, Body: io.NopCloser(strings.NewReader("image")), Header: make(http.Header), Request: req}, nil
			})}
			req, err := http.NewRequestWithContext(context.Background(), tc.method, "https://media.example/image", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Referer", tc.referer)
			req.Header.Set("User-Agent", "media-test")
			resp, retried, err := DoWithRefererFallback(client, req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			wantCalls := 1
			if tc.wantRetry {
				wantCalls++
			}
			if retried != tc.wantRetry || calls != wantCalls || resp.StatusCode != tc.finalCode {
				t.Fatalf("retry=%v calls=%d status=%d", retried, calls, resp.StatusCode)
			}
			if req.Header.Get("Referer") != tc.referer || client.CheckRedirect != nil {
				t.Fatal("helper changed original request or shared client")
			}
			if !tc.wantRetry && firstBody.closed {
				t.Fatal("helper closed response that belongs to caller")
			}
		})
	}
}

func TestDoWithRefererFallbackErrorsAndCancellation(t *testing.T) {
	for _, mode := range []string{"network", "cancelled before request", "cancelled after forbidden", "cancelled during retry"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			firstBody := &refererFallbackBody{Reader: strings.NewReader("forbidden")}
			networkErr := errors.New("network unavailable")
			client := &http.Client{Transport: RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				if mode == "network" {
					return nil, networkErr
				}
				if mode == "cancelled after forbidden" {
					cancel()
				}
				if calls == 2 {
					cancel()
					return nil, req.Context().Err()
				}
				return &http.Response{StatusCode: 403, Body: firstBody, Header: make(http.Header), Request: req}, nil
			})}
			if mode == "cancelled before request" {
				cancel()
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://media.example/image", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Referer", "https://article.example")
			resp, retried, err := DoWithRefererFallback(client, req)
			if resp != nil {
				resp.Body.Close()
			}
			wantErr := error(context.Canceled)
			wantCalls := 1
			if mode == "network" {
				wantErr = networkErr
			} else if mode == "cancelled before request" {
				wantCalls = 0
			} else if mode == "cancelled during retry" {
				wantCalls = 2
			}
			if !errors.Is(err, wantErr) || calls != wantCalls || retried != (wantCalls == 2) {
				t.Fatalf("err=%v calls=%d retried=%v", err, calls, retried)
			}
			if strings.HasPrefix(mode, "cancelled after") || strings.HasPrefix(mode, "cancelled during") {
				if !firstBody.closed {
					t.Fatal("forbidden body was not closed")
				}
			}
		})
	}
}

func TestDoWithRefererFallbackClearsRedirectReferer(t *testing.T) {
	var calls atomic.Int32
	var redirects atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path == "/image" && r.Header.Get("Referer") != "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if r.Header.Get("Referer") != "" {
			t.Errorf("retry or redirect retained Referer %q", r.Header.Get("Referer"))
		}
		if r.URL.Path == "/image" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		if r.Header.Get("X-Redirect-Policy") != "preserved" {
			t.Error("original redirect callback was not applied")
		}
		_, _ = io.WriteString(w, "image")
	}))
	defer server.Close()
	client := server.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		redirects.Add(1)
		req.Header.Set("Referer", "https://must-be-removed.example")
		req.Header.Set("X-Redirect-Policy", "preserved")
		return nil
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/image", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Referer", "https://article.example")
	resp, retried, err := DoWithRefererFallback(client, req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if !retried || resp.StatusCode != 200 || calls.Load() != 3 || redirects.Load() != 1 {
		t.Fatalf("retry=%v status=%d requests=%d redirects=%d", retried, resp.StatusCode, calls.Load(), redirects.Load())
	}
	probe, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	if err := client.CheckRedirect(probe, nil); err != nil || probe.Header.Get("Referer") == "" {
		t.Fatal("shared client's redirect callback was changed")
	}
}

func TestDoWithRefererFallbackPreservesRedirectLimits(t *testing.T) {
	for _, custom := range []bool{false, true} {
		t.Run(map[bool]string{false: "default", true: "custom"}[custom], func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Header.Get("Referer") != "" {
					w.WriteHeader(http.StatusForbidden)
					return
				}
				http.Redirect(w, r, "/loop", http.StatusFound)
			}))
			defer server.Close()
			client := server.Client()
			policyErr := errors.New("redirect rejected by policy")
			if custom {
				client.CheckRedirect = func(*http.Request, []*http.Request) error { return policyErr }
			}
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/image", nil)
			req.Header.Set("Referer", "https://article.example")
			resp, retried, err := DoWithRefererFallback(client, req)
			if resp != nil {
				resp.Body.Close()
			}
			wantCalls := int32(11)
			if custom {
				wantCalls = 2
				if !errors.Is(err, policyErr) {
					t.Fatalf("redirect policy error lost: %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), "stopped after 10 redirects") {
				t.Fatalf("default redirect limit lost: %v", err)
			}
			if !retried || calls.Load() != wantCalls {
				t.Fatalf("retry=%v calls=%d want=%d", retried, calls.Load(), wantCalls)
			}
		})
	}
}

func TestBuildProxyURLPreservesCredentialsAndIPv6(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "::1", "[::1]"} {
		raw := BuildProxyURL("http", host, "7890", "user@company", "p@ss:/?#%")
		parsed, err := url.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		password, _ := parsed.User.Password()
		if parsed.User.Username() != "user@company" || password != "p@ss:/?#%" {
			t.Fatal("proxy credentials changed during URL construction")
		}
		if parsed.Port() != "7890" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			t.Fatal("proxy credentials leaked into URL components")
		}
	}
}

func TestCreateHTTPClientHonorsInsecureTLSVerifyEnv(t *testing.T) {
	t.Setenv(InsecureSkipTLSVerifyEnv, "true")

	client, err := CreateHTTPClient("", time.Second)
	if err != nil {
		t.Fatalf("CreateHTTPClient returned error: %v", err)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport type %T", client.Transport)
	}
	if transport.TLSClientConfig == nil {
		t.Fatalf("TLSClientConfig is nil")
	}
	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatalf("expected InsecureSkipVerify to be true")
	}
	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Fatalf("expected TLS 1.2 minimum, got %d", transport.TLSClientConfig.MinVersion)
	}
}

func TestCreateHTTPClientKeepsTLSVerificationByDefault(t *testing.T) {
	t.Setenv(InsecureSkipTLSVerifyEnv, "")

	client, err := CreateHTTPClient("", time.Second)
	if err != nil {
		t.Fatalf("CreateHTTPClient returned error: %v", err)
	}

	transport := client.Transport.(*http.Transport)
	if transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatalf("expected InsecureSkipVerify to be false by default")
	}
}

func TestCreateHTTPClientNegotiatesHTTP2(t *testing.T) {
	t.Setenv(InsecureSkipTLSVerifyEnv, "true")

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 2 {
			t.Errorf("expected HTTP/2 request, got %s", r.Proto)
		}
		_, _ = io.WriteString(w, "ok")
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	client, err := CreateHTTPClient("", 5*time.Second)
	if err != nil {
		t.Fatalf("CreateHTTPClient returned error: %v", err)
	}
	response, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("HTTP/2 request failed: %v", err)
	}
	defer response.Body.Close()
}

func TestCreateHTTPClientFallsBackToHTTP1(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 1 {
			t.Errorf("expected HTTP/1.1 request, got %s", r.Proto)
		}
		_, _ = io.WriteString(w, "ok")
	}))
	defer server.Close()

	client, err := CreateHTTPClient("", 5*time.Second)
	if err != nil {
		t.Fatalf("CreateHTTPClient returned error: %v", err)
	}
	response, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("HTTP/1.1 request failed: %v", err)
	}
	defer response.Body.Close()
}

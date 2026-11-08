package httputil

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"testing/fstest"
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

func TestCompressStaticFileServer(t *testing.T) {
	modified := time.Date(2026, time.October, 30, 0, 0, 0, 0, time.UTC)
	files := fstest.MapFS{
		"index.html": {Data: []byte(strings.Repeat("<p>Reader</p>", 100)), ModTime: modified},
		"app.js":     {Data: []byte(strings.Repeat("console.log('reader');\n", 100)), ModTime: modified},
		"style.css":  {Data: []byte(strings.Repeat("body { color: black; }\n", 100)), ModTime: modified},
		"data.json":  {Data: []byte(`{"reader":true}`), ModTime: modified},
		"icon.svg":   {Data: []byte(`<svg xmlns="http://www.w3.org/2000/svg"><path d="M0 0"/></svg>`), ModTime: modified},
		"image.png":  {Data: append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 100)...), ModTime: modified},
	}
	plain := http.FileServer(http.FS(files))
	compressed := CompressStatic(plain)
	for _, tc := range []struct {
		name       string
		method     string
		path       string
		encodings  []string
		headers    map[string]string
		wantGzip   bool
		wantStatus int
	}{
		{"javascript", "GET", "/app.js", []string{"gzip"}, nil, true, 200},
		{"html", "GET", "/", []string{"gzip"}, nil, true, 200},
		{"css", "GET", "/style.css", []string{"gzip"}, nil, true, 200},
		{"json", "GET", "/data.json", []string{"gzip"}, nil, true, 200},
		{"svg", "GET", "/icon.svg", []string{"gzip"}, nil, true, 200},
		{"unsupported", "GET", "/app.js", []string{"br"}, nil, false, 200},
		{"identity", "GET", "/app.js", []string{"identity"}, nil, false, 200},
		{"no encoding", "GET", "/app.js", nil, nil, false, 200},
		{"explicit zero", "GET", "/app.js", []string{"gzip;q=0, *;q=1"}, nil, false, 200},
		{"zero wildcard", "GET", "/app.js", []string{"*;q=0"}, nil, false, 200},
		{"wildcard", "GET", "/app.js", []string{"br, *;q=0.5"}, nil, true, 200},
		{"multiline", "GET", "/app.js", []string{"br", "GZIP; Q=0.5"}, nil, true, 200},
		{"prefer identity", "GET", "/app.js", []string{"gzip;q=0.5, identity;q=1"}, nil, false, 200},
		{"invalid quality", "GET", "/app.js", []string{"gzip;q=invalid"}, nil, false, 200},
		{"nonfinite quality", "GET", "/app.js", []string{"gzip;q=NaN"}, nil, false, 200},
		{"quality outside range", "GET", "/app.js", []string{"gzip;q=1.5"}, nil, false, 200},
		{"head", "HEAD", "/app.js", []string{"gzip"}, nil, false, 200},
		{"range", "GET", "/app.js", []string{"gzip"}, map[string]string{"Range": "bytes=0-8"}, false, 206},
		{"not modified", "GET", "/app.js", []string{"gzip"}, map[string]string{"If-Modified-Since": modified.Format(http.TimeFormat)}, false, 304},
		{"not found", "GET", "/missing.js", []string{"gzip"}, nil, false, 404},
		{"binary", "GET", "/image.png", []string{"gzip"}, nil, false, 200},
		{"request no-transform", "GET", "/app.js", []string{"gzip"}, map[string]string{"Cache-Control": "max-age=60, No-Transform"}, false, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			for _, encoding := range tc.encodings {
				req.Header.Add("Accept-Encoding", encoding)
			}
			for key, value := range tc.headers {
				req.Header.Set(key, value)
			}
			original := httptest.NewRecorder()
			plain.ServeHTTP(original, req)
			result := httptest.NewRecorder()
			compressed.ServeHTTP(result, req)
			if result.Code != tc.wantStatus || result.Code != original.Code {
				t.Fatalf("status=%d, want %d", result.Code, tc.wantStatus)
			}
			for _, header := range []string{"Content-Type", "Content-Range", "Last-Modified"} {
				if got, want := result.Header().Get(header), original.Header().Get(header); got != want {
					t.Errorf("%s=%q, want %q", header, got, want)
				}
			}
			if !strings.Contains(strings.ToLower(strings.Join(result.Header().Values("Vary"), ",")), "accept-encoding") {
				t.Error("missing Accept-Encoding cache variation")
			}
			body := result.Body.Bytes()
			if tc.wantGzip {
				if result.Header().Get("Content-Encoding") != "gzip" || result.Header().Get("Content-Length") != "" {
					t.Fatalf("unexpected compressed headers: %v", result.Header())
				}
				reader, err := gzip.NewReader(bytes.NewReader(body))
				if err != nil {
					t.Fatal(err)
				}
				body, err = io.ReadAll(reader)
				if err != nil {
					t.Fatal(err)
				}
				if err := reader.Close(); err != nil {
					t.Fatal(err)
				}
			} else if result.Header().Get("Content-Encoding") != "" || result.Header().Get("Content-Length") != original.Header().Get("Content-Length") {
				t.Fatalf("identity headers changed: %v, original %v", result.Header(), original.Header())
			}
			if !bytes.Equal(body, original.Body.Bytes()) {
				t.Fatal("response bytes changed")
			}
		})
	}
}

func TestCompressStaticResponseGuards(t *testing.T) {
	body := []byte(strings.Repeat("reader text ", 100))
	for _, tc := range []struct {
		name     string
		status   int
		headers  map[string]string
		empty    bool
		implicit bool
	}{
		{"existing encoding", 200, map[string]string{"Content-Encoding": "br"}, false, false},
		{"content range", 200, map[string]string{"Content-Range": "bytes 0-10/200"}, false, false},
		{"no transform", 200, map[string]string{"Cache-Control": "private, no-transform"}, false, false},
		{"binary content type", 200, map[string]string{"Content-Type": "application/octet-stream"}, false, false},
		{"invalid content type", 200, map[string]string{"Content-Type": "invalid/type; broken"}, false, false},
		{"error", 500, nil, false, false},
		{"created", 201, nil, false, false},
		{"empty explicit", 200, nil, true, false},
		{"empty implicit", 200, nil, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				length := len(body)
				if tc.empty {
					length = 0
				}
				w.Header().Set("Content-Length", strconv.Itoa(length))
				for key, value := range tc.headers {
					w.Header().Set(key, value)
				}
				if !tc.implicit {
					w.WriteHeader(tc.status)
				}
				if !tc.empty {
					_, _ = w.Write(body)
				}
			})
			req := httptest.NewRequest("GET", "/asset", nil)
			req.Header.Set("Accept-Encoding", "gzip")
			original, result := httptest.NewRecorder(), httptest.NewRecorder()
			handler.ServeHTTP(original, req)
			CompressStatic(handler).ServeHTTP(result, req)
			if result.Code != original.Code || !bytes.Equal(result.Body.Bytes(), original.Body.Bytes()) {
				t.Fatal("status or original body changed")
			}
			for _, key := range []string{"Content-Type", "Content-Encoding", "Content-Length", "Content-Range", "Cache-Control"} {
				if result.Header().Get(key) != original.Header().Get(key) {
					t.Errorf("%s changed: result %q, original %q", key, result.Header().Get(key), original.Header().Get(key))
				}
			}
		})
	}
}

func TestCompressStaticPreservesVary(t *testing.T) {
	for _, values := range [][]string{{"Origin", "Accept-Language"}, {"Origin, accept-encoding"}, {"*"}} {
		for _, encoding := range []string{"gzip", "identity"} {
			t.Run(strings.Join(values, "/")+"/"+encoding, func(t *testing.T) {
				req := httptest.NewRequest("GET", "/asset.js", nil)
				req.Header.Set("Accept-Encoding", encoding)
				recorder := httptest.NewRecorder()
				CompressStatic(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					for _, value := range values {
						w.Header().Add("Vary", value)
					}
					w.Header().Set("Content-Type", "application/javascript")
					_, _ = w.Write([]byte("let reader = true;"))
				})).ServeHTTP(recorder, req)
				got := recorder.Header().Values("Vary")
				if len(got) < len(values) {
					t.Fatalf("Vary lost values: %v", got)
				}
				for i, value := range values {
					if got[i] != value {
						t.Fatalf("Vary changed: %v", got)
					}
				}
				if values[0] == "*" && len(got) != 1 {
					t.Errorf("redundant Vary: %v", got)
				}
				if strings.Contains(strings.ToLower(strings.Join(values, ",")), "accept-encoding") && len(got) != len(values) {
					t.Errorf("duplicate Vary: %v", got)
				}
			})
		}
	}
}

type failedCompressionWriter struct {
	header    http.Header
	remaining int
}

func (w *failedCompressionWriter) Header() http.Header { return w.header }
func (w *failedCompressionWriter) WriteHeader(int)     {}
func (w *failedCompressionWriter) Write(data []byte) (int, error) {
	if len(data) <= w.remaining {
		w.remaining -= len(data)
		return len(data), nil
	}
	written := w.remaining
	w.remaining = 0
	return written, io.ErrClosedPipe
}

func TestCompressStaticPropagatesWriteFailures(t *testing.T) {
	for _, allowed := range []int{0, 10} {
		underlying := &failedCompressionWriter{header: make(http.Header), remaining: allowed}
		underlying.Header().Set("Content-Type", "text/plain")
		writer := &staticGzipWriter{ResponseWriter: underlying, canCompress: true}
		_, writeErr := writer.Write([]byte("reader"))
		closeErr := writer.close()
		if !errors.Is(writeErr, io.ErrClosedPipe) && !errors.Is(closeErr, io.ErrClosedPipe) {
			t.Fatalf("write=%v close=%v, want closed pipe", writeErr, closeErr)
		}
	}
}

func TestCompressStaticEmptyWriteAndCommittedHeaders(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/javascript")
			if explicit {
				w.WriteHeader(http.StatusOK)
				w.Header().Set("Content-Type", "application/octet-stream")
				w.Header().Set("X-Late", "must-not-be-sent")
			}
			if n, err := w.Write(nil); n != 0 || err != nil {
				t.Errorf("empty write=%d, %v", n, err)
			}
			// Even an empty write commits an implicit 200 and its headers.
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("X-Late", "must-not-be-sent")
			_, _ = io.WriteString(w, "let reader = true;")
		})
		req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		recorder := httptest.NewRecorder()
		CompressStatic(handler).ServeHTTP(recorder, req)
		response := recorder.Result()
		if response.StatusCode != http.StatusOK || response.Header.Get("Content-Encoding") != "gzip" ||
			response.Header.Get("Content-Type") != "application/javascript" || response.Header.Get("X-Late") != "" {
			t.Fatalf("explicit=%v, unexpected response headers: %v", explicit, response.Header)
		}
		reader, err := gzip.NewReader(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(reader)
		if err != nil || string(body) != "let reader = true;" {
			t.Fatalf("body=%q err=%v", body, err)
		}
		_ = reader.Close()
		_ = response.Body.Close()
	}
}

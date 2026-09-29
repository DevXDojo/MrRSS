package cache

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"MrRSS/internal/utils/httputil"
)

func TestMediaCacheRefererFallback(t *testing.T) {
	for _, inMemory := range []bool{false, true} {
		for _, finalStatus := range []int{http.StatusOK, http.StatusForbidden, http.StatusNotFound} {
			name := fmt.Sprintf("memory=%v/status=%d", inMemory, finalStatus)
			t.Run(name, func(t *testing.T) {
				dir := t.TempDir()
				mc, err := NewMediaCache(dir)
				if err != nil {
					t.Fatal(err)
				}
				var calls atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					calls.Add(1)
					if r.Header.Get("Referer") != "" && finalStatus != http.StatusNotFound {
						w.WriteHeader(http.StatusForbidden)
						return
					}
					w.Header().Set("Content-Type", "image/png")
					w.WriteHeader(finalStatus)
					_, _ = io.WriteString(w, "image")
				}))
				defer server.Close()
				mediaURL := server.URL + "/image.png"
				if inMemory {
					_, _, err = mc.Get(context.Background(), server.Client(), mediaURL, "https://article.example")
				} else {
					_, _, err = mc.DownloadToFile(context.Background(), server.Client(), mediaURL, "https://article.example")
				}
				wantCalls := int32(2)
				if finalStatus == http.StatusNotFound {
					wantCalls = 1
				}
				if calls.Load() != wantCalls {
					t.Fatalf("requests=%d want=%d", calls.Load(), wantCalls)
				}
				if finalStatus == http.StatusOK {
					if err != nil {
						t.Fatal(err)
					}
					data, contentType, err := mc.Get(context.Background(), server.Client(), mediaURL, "https://article.example")
					if err != nil || string(data) != "image" || contentType != "image/png" || calls.Load() != wantCalls {
						t.Fatalf("cache hit: data=%q type=%q err=%v requests=%d", data, contentType, err, calls.Load())
					}
				} else {
					var downloadErr *MediaDownloadError
					if !errors.As(err, &downloadErr) || downloadErr.RefererFallbackAttempted != (wantCalls == 2) {
						t.Fatalf("download error lost retry state: %v", err)
					}
					entries, err := os.ReadDir(dir)
					if err != nil || len(entries) != 0 {
						t.Fatalf("failed download left files: %v, %v", entries, err)
					}
				}
			})
		}
	}
}

type failingMediaBody struct {
	data   string
	err    error
	cancel context.CancelFunc
	closed bool
}

func (b *failingMediaBody) Read(p []byte) (int, error) {
	if b.data != "" {
		n := copy(p, b.data)
		b.data = b.data[n:]
		return n, nil
	}
	if b.cancel != nil {
		b.cancel()
	}
	return 0, b.err
}

func (b *failingMediaBody) Close() error {
	b.closed = true
	return nil
}

func TestMediaCacheIncompleteOrCancelledBodyNeverCached(t *testing.T) {
	for _, inMemory := range []bool{false, true} {
		for _, mode := range []string{"short", "read failure", "cancelled", "empty"} {
			t.Run(fmt.Sprintf("memory=%v/%s", inMemory, mode), func(t *testing.T) {
				dir := t.TempDir()
				mc, err := NewMediaCache(dir)
				if err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				body := &failingMediaBody{data: "partial", err: io.EOF}
				length := int64(10)
				if mode == "read failure" {
					body.err = io.ErrUnexpectedEOF
				} else if mode == "cancelled" {
					body.cancel = cancel
					length = -1
				} else if mode == "empty" {
					body.data = ""
					length = 0
				}
				calls := 0
				client := &http.Client{Transport: httputil.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
					calls++
					if calls == 1 {
						return &http.Response{StatusCode: 403, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("forbidden")), Request: req}, nil
					}
					return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"image/png"}}, Body: body, ContentLength: length, Request: req}, nil
				})}
				if inMemory {
					_, _, err = mc.Get(ctx, client, "https://media.example/image.png", "https://article.example")
				} else {
					_, _, err = mc.DownloadToFile(ctx, client, "https://media.example/image.png", "https://article.example")
				}
				var downloadErr *MediaDownloadError
				if !errors.As(err, &downloadErr) || !downloadErr.RefererFallbackAttempted || calls != 2 {
					t.Fatalf("err=%v calls=%d", err, calls)
				}
				if mode == "cancelled" && !errors.Is(err, context.Canceled) {
					t.Fatalf("cancellation lost: %v", err)
				}
				if !body.closed {
					t.Fatal("failed response body was not closed")
				}
				entries, err := os.ReadDir(dir)
				if err != nil || len(entries) != 0 {
					t.Fatalf("incomplete download left files: %v, %v", entries, err)
				}
			})
		}
	}
}

func TestMediaCacheStorageErrorAllowsDirectFallback(t *testing.T) {
	for _, inMemory := range []bool{false, true} {
		t.Run(fmt.Sprintf("memory=%v", inMemory), func(t *testing.T) {
			// A file in place of the cache directory makes cache storage fail
			// deterministically, including when tests run as a privileged user.
			cachePath := filepath.Join(t.TempDir(), "not-a-directory")
			if err := os.WriteFile(cachePath, []byte("occupied"), 0600); err != nil {
				t.Fatal(err)
			}
			mc := &MediaCache{cacheDir: cachePath}
			calls := 0
			client := &http.Client{Transport: httputil.RoundTripFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				code := 200
				if calls == 1 {
					code = 403
				}
				return &http.Response{StatusCode: code, Header: http.Header{"Content-Type": {"image/png"}}, Body: io.NopCloser(strings.NewReader("image")), ContentLength: 5, Request: req}, nil
			})}
			var err error
			if inMemory {
				_, _, err = mc.Get(context.Background(), client, "https://media.example/image.png", "https://article.example")
			} else {
				_, _, err = mc.DownloadToFile(context.Background(), client, "https://media.example/image.png", "https://article.example")
			}
			var downloadErr *MediaDownloadError
			if err == nil || errors.As(err, &downloadErr) || calls != 2 {
				t.Fatalf("storage error incorrectly prevents direct fallback: err=%v calls=%d", err, calls)
			}
		})
	}
}

func TestMediaCacheCancelledDownloadDoesNotWriteFile(t *testing.T) {
	mc, err := NewMediaCache(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("cancelled request reached media host")
	}))
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err = mc.Get(ctx, server.Client(), server.URL+"/image.png", "")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if mc.Exists(server.URL + "/image.png") {
		t.Fatal("cancelled download was cached")
	}
}

func TestMediaCache_BasicOperations(t *testing.T) {
	dir := t.TempDir()

	mc, err := NewMediaCache(dir)
	if err != nil {
		t.Fatalf("NewMediaCache failed: %v", err)
	}

	url := "https://example.com/image.jpg?query=1"
	path := mc.GetCachedPath(url)
	if filepath.Dir(path) != dir {
		t.Fatalf("cached path in wrong dir: %s", path)
	}

	// Create a cached file to simulate existing cache
	if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
		t.Fatalf("write cached file: %v", err)
	}

	if !mc.Exists(url) {
		t.Fatalf("expected Exists to be true for cached file")
	}

	data, ctype, err := mc.Get(context.Background(), http.DefaultClient, url, "")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(data) != "data" {
		t.Fatalf("unexpected data")
	}
	if ctype != "image/jpeg" {
		t.Fatalf("unexpected content type: %s", ctype)
	}

	// Test CleanupOldFiles (create an old file)
	oldPath := filepath.Join(dir, "old.bin")
	if err := os.WriteFile(oldPath, []byte("x"), 0644); err != nil {
		t.Fatalf("write old file: %v", err)
	}
	// backdate modification time
	oldTime := time.Now().AddDate(0, 0, -10)
	_ = os.Chtimes(oldPath, oldTime, oldTime)

	removed, err := mc.CleanupOldFiles(1)
	if err != nil {
		t.Fatalf("CleanupOldFiles failed: %v", err)
	}
	if removed == 0 {
		t.Fatalf("expected CleanupOldFiles to remove old file")
	}
}

func TestGetExtensionAndContentTypeHelpers(t *testing.T) {
	if ext := getExtensionFromURL("https://x/y.png?v=1"); ext != ".png" {
		t.Fatalf("expected .png got %s", ext)
	}
	if ct := getContentTypeFromPath("file.jpg"); ct != "image/jpeg" {
		t.Fatalf("unexpected content type: %s", ct)
	}
	if ext := getExtensionFromContentType("image/png; charset=utf8"); ext != ".png" {
		t.Fatalf("unexpected ext: %s", ext)
	}
}

func TestMediaCacheDownloadToFileStreamsAndOpens(t *testing.T) {
	dir := t.TempDir()
	mc, err := NewMediaCache(dir)
	if err != nil {
		t.Fatal(err)
	}

	payload := make([]byte, 64*1024)
	for i := range payload {
		payload[i] = byte(i % 251)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	mediaURL := server.URL + "/image.png"
	path, contentType, err := mc.DownloadToFile(context.Background(), server.Client(), mediaURL, "")
	if err != nil {
		t.Fatalf("DownloadToFile failed: %v", err)
	}
	if contentType != "image/png" {
		t.Fatalf("content type = %q, want image/png", contentType)
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("cached file outside cache dir: %s", path)
	}

	stored, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cached file: %v", err)
	}
	if len(stored) != len(payload) {
		t.Fatalf("cached %d bytes, want %d", len(stored), len(payload))
	}

	file, openType, modTime, err := mc.Open(mediaURL)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer file.Close()
	if openType != "image/png" {
		t.Fatalf("Open content type = %q, want image/png", openType)
	}
	if modTime.IsZero() {
		t.Fatal("Open returned zero mod time")
	}
	streamed, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("read through Open: %v", err)
	}
	if len(streamed) != len(payload) {
		t.Fatalf("streamed %d bytes, want %d", len(streamed), len(payload))
	}

	if _, _, _, err := mc.Open(server.URL + "/missing.png"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Open for missing media = %v, want os.ErrNotExist", err)
	}
}

func TestMediaCacheDownloadToFileCancelledLeavesNoFile(t *testing.T) {
	dir := t.TempDir()
	mc, err := NewMediaCache(dir)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("cancelled request reached media host")
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	mediaURL := server.URL + "/image.png"
	if _, _, err := mc.DownloadToFile(ctx, server.Client(), mediaURL, ""); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if mc.Exists(mediaURL) {
		t.Fatal("cancelled download was cached")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read cache dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("cancelled download left %d file(s) behind", len(entries))
	}
}

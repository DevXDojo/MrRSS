package cache

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

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

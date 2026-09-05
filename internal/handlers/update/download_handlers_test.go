package update

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
)

func TestHandleDownloadUpdate_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/update/download", nil)
	rr := httptest.NewRecorder()

	HandleDownloadUpdate(&core.Handler{}, rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleDownloadUpdate_InvalidURLPrefix(t *testing.T) {
	body := bytes.NewReader([]byte(`{"download_url":"https://malicious.com/file","asset_name":"app.zip"}`))
	req := httptest.NewRequest(http.MethodPost, "/update/download", body)
	rr := httptest.NewRecorder()

	HandleDownloadUpdate(&core.Handler{}, rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid url prefix, got %d", rr.Code)
	}
}

func TestHandleDownloadUpdate_InvalidAssetName(t *testing.T) {
	body := bytes.NewReader([]byte(`{"download_url":"https://github.com/DevXDojo/MrRSS/releases/download/v1/app.zip","asset_name":"../secret"}`))
	req := httptest.NewRequest(http.MethodPost, "/update/download", body)
	rr := httptest.NewRecorder()

	HandleDownloadUpdate(&core.Handler{}, rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid asset name, got %d", rr.Code)
	}
}
func TestHandleDownloadUpdateProgress(t *testing.T) {
	requestID := "progress-test"
	setDownloadProgress(downloadProgress{
		RequestID: requestID, State: "downloading", BytesWritten: 50,
		TotalBytes: 100, Percentage: 50,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/download-update/progress?request_id="+requestID, nil)
	rr := httptest.NewRecorder()
	HandleDownloadUpdateProgress(nil, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var progress downloadProgress
	if err := json.NewDecoder(rr.Body).Decode(&progress); err != nil {
		t.Fatalf("decode progress: %v", err)
	}
	if progress.BytesWritten != 50 || progress.Percentage != 50 || progress.State != "downloading" {
		t.Fatalf("unexpected progress: %+v", progress)
	}
}

func TestDownloadUpdateFileRetriesWithRange(t *testing.T) {
	content := []byte(strings.Repeat("update-data-", 8192))
	half := len(content) / 2
	var attempts atomic.Int32
	var resumed atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := attempts.Add(1)
		if attempt == 1 {
			w.Header().Set("Content-Length", fmt.Sprint(len(content)))
			_, _ = w.Write(content[:half])
			return
		}
		expectedRange := fmt.Sprintf("bytes=%d-", half)
		if r.Header.Get("Range") != expectedRange {
			t.Errorf("expected Range %q, got %q", expectedRange, r.Header.Get("Range"))
		}
		resumed.Store(true)
		w.Header().Set("Content-Length", fmt.Sprint(len(content)-half))
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", half, len(content)-1, len(content)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(content[half:])
	}))
	defer server.Close()

	filePath := filepath.Join(t.TempDir(), "update.bin")
	written, total, err := downloadUpdateFile(context.Background(), server.Client(), server.URL, filePath, "range-test", 3, 0)
	if err != nil {
		t.Fatalf("downloadUpdateFile: %v", err)
	}
	if !resumed.Load() || attempts.Load() != 2 {
		t.Fatalf("expected one resumed request, attempts=%d resumed=%v", attempts.Load(), resumed.Load())
	}
	if written != int64(len(content)) || total != int64(len(content)) {
		t.Fatalf("unexpected sizes: written=%d total=%d", written, total)
	}
	got, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatal("downloaded content does not match")
	}
}

func TestDownloadUpdateFileRestartsWhenRangeUnsupported(t *testing.T) {
	content := []byte(strings.Repeat("no-range-", 4096))
	half := len(content) / 2
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := attempts.Add(1)
		w.Header().Set("Content-Length", fmt.Sprint(len(content)))
		if attempt == 1 {
			_, _ = w.Write(content[:half])
			return
		}
		if r.Header.Get("Range") == "" {
			t.Error("expected a resume request")
		}
		_, _ = w.Write(content)
	}))
	defer server.Close()

	filePath := filepath.Join(t.TempDir(), "update.bin")
	_, _, err := downloadUpdateFile(context.Background(), server.Client(), server.URL, filePath, "no-range-test", 3, 0)
	if err != nil {
		t.Fatalf("downloadUpdateFile: %v", err)
	}
	got, _ := os.ReadFile(filePath)
	if !bytes.Equal(got, content) {
		t.Fatal("fallback full download did not replace partial content")
	}
}

func TestDownloadUpdateFileRetriesServerErrors(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) < 3 {
			http.Error(w, "temporary", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("complete"))
	}))
	defer server.Close()

	filePath := filepath.Join(t.TempDir(), "update.bin")
	if _, _, err := downloadUpdateFile(context.Background(), server.Client(), server.URL, filePath, "server-retry", 3, 0); err != nil {
		t.Fatalf("downloadUpdateFile: %v", err)
	}
	if attempts.Load() != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestDownloadUpdateFileCleansPartialAfterFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "100")
		_, _ = w.Write([]byte("partial"))
	}))
	defer server.Close()

	filePath := filepath.Join(t.TempDir(), "update.bin")
	if _, _, err := downloadUpdateFile(context.Background(), server.Client(), server.URL, filePath, "cleanup-test", 2, 0); err == nil {
		t.Fatal("expected incomplete download error")
	}
	if _, err := os.Stat(filePath + ".part"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial file was not removed: %v", err)
	}
	if _, err := os.Stat(filePath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("incomplete final file exists: %v", err)
	}
}

func TestDownloadUpdateFileTracksUnknownLength(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("response writer does not support flushing")
		}
		_, _ = w.Write([]byte("chunk-one"))
		flusher.Flush()
		_, _ = w.Write([]byte("chunk-two"))
	}))
	defer server.Close()

	filePath := filepath.Join(t.TempDir(), "update.bin")
	written, total, err := downloadUpdateFile(context.Background(), server.Client(), server.URL, filePath, "unknown-size", 1, 0)
	if err != nil {
		t.Fatalf("downloadUpdateFile: %v", err)
	}
	if total != 0 || written == 0 {
		t.Fatalf("expected unknown total and downloaded bytes, total=%d written=%d", total, written)
	}
	updateDownloads.RLock()
	progress := updateDownloads.items["unknown-size"]
	updateDownloads.RUnlock()
	if !progress.Indeterminate || progress.BytesWritten != written {
		t.Fatalf("unexpected unknown-length progress: %+v", progress)
	}
}

func TestCreateUpdateHTTPClientUsesGlobalProxy(t *testing.T) {
	db, err := database.NewDB(filepath.Join(t.TempDir(), "settings.db"))
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	defer db.Close()
	if err := db.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}
	for key, value := range map[string]string{
		"proxy_enabled": "true", "proxy_type": "http", "proxy_host": "127.0.0.1", "proxy_port": "7890",
	} {
		if err := db.SetSetting(key, value); err != nil {
			t.Fatalf("SetSetting(%s): %v", key, err)
		}
	}

	client, err := createUpdateHTTPClient(&core.Handler{DB: db}, time.Second)
	if err != nil {
		t.Fatalf("createUpdateHTTPClient: %v", err)
	}
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport %T", client.Transport)
	}
	proxyURL, err := transport.Proxy(httptest.NewRequest(http.MethodGet, "https://github.com", nil))
	if err != nil {
		t.Fatalf("proxy lookup: %v", err)
	}
	if proxyURL == nil || proxyURL.String() != "http://127.0.0.1:7890" {
		t.Fatalf("unexpected proxy URL: %v", proxyURL)
	}
}

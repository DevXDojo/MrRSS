package update

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateDownloadsUseSeparateDirectories(t *testing.T) {
	first, err := createUpdateDownloadPath("app.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filepath.Dir(first))
	second, err := createUpdateDownloadPath("app.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filepath.Dir(second))
	if first == second || filepath.Base(first) != "app.zip" {
		t.Fatal("downloads must have isolated directories and preserve installer names")
	}
}

func TestDownloadRejectsInvalidPartialResponse(t *testing.T) {
	for _, value := range []string{"bytes 4-7/8", "bytes 0-99/8", "bytes 0-no/8", "bytes 0-7/0"} {
		t.Run(value, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Range", value)
				w.WriteHeader(http.StatusPartialContent)
				fmt.Fprint(w, "partial")
			}))
			defer server.Close()
			path := filepath.Join(t.TempDir(), "app.zip")
			_, _, err := downloadUpdateFile(context.Background(), server.Client(), server.URL, path, "invalid-partial", 1, 0)
			if err == nil {
				t.Fatal("invalid partial installer was accepted")
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatal("invalid installer was retained")
			}
			if _, err := os.Stat(path + ".part"); !os.IsNotExist(err) {
				t.Fatal("partial installer was retained")
			}
		})
	}
}

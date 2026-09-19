package media

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServeCachedMediaRangeAndConditional(t *testing.T) {
	path := filepath.Join(t.TempDir(), "media.png")
	if err := os.WriteFile(path, []byte("0123456789"), 0600); err != nil {
		t.Fatal(err)
	}
	modified := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		method, header, value string
		status                int
		body                  string
	}{
		{"GET", "Range", "bytes=2-5", 206, "2345"},
		{"GET", "If-Modified-Since", "Thu, 01 Jan 2026 00:00:00 GMT", 304, ""},
		{"HEAD", "", "", 200, ""},
	} {
		r := httptest.NewRequest(tc.method, "/media", nil)
		if tc.header != "" {
			r.Header.Set(tc.header, tc.value)
		}
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		serveCachedMedia(w, r, file, "image/png", modified, "media.png")
		if w.Code != tc.status || w.Body.String() != tc.body {
			t.Errorf("%s %s: status %d, body %q", tc.method, tc.header, w.Code, w.Body.String())
		}
		if _, err := file.Read(make([]byte, 1)); err == nil {
			t.Fatal("served file was not closed")
		}
	}
}

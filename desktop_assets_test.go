//go:build !server

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestDesktopAssetMiddlewarePreservesRuntimeAndFrontend(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/version", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("api"))
	})
	combined := &CombinedHandler{apiMux: mux}
	chain := APIMiddleware(combined)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("wails"))
	}))
	for _, path := range []string{"/", "/?ticker=1", "/?ticker=settings", "/src/main.ts", "/assets/App.js", "/wails/runtime.js", "/wails/runtime"} {
		response := httptest.NewRecorder()
		chain.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Body.String() != "wails" {
			t.Errorf("%s bypassed the Wails asset chain: %q", path, response.Body.String())
		}
	}
	response := httptest.NewRecorder()
	chain.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/version", nil))
	if response.Body.String() != "api" {
		t.Fatalf("API request reached the frontend: %q", response.Body.String())
	}
}

func TestDesktopBundledAssetsWithoutDevServer(t *testing.T) {
	t.Setenv("FRONTEND_DEVSERVER_URL", "")
	assets := application.BundledAssetFileServer(fstest.MapFS{
		"index.html":    &fstest.MapFile{Data: []byte("embedded index")},
		"assets/App.js": &fstest.MapFile{Data: []byte("export default 'embedded'")},
	})
	for path, want := range map[string]string{"/": "embedded index", "/assets/App.js": "export default 'embedded'"} {
		response := httptest.NewRecorder()
		assets.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK || response.Body.String() != want {
			t.Errorf("%s: status=%d body=%q", path, response.Code, response.Body.String())
		}
	}
	response := httptest.NewRecorder()
	assets.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/wails/runtime.js", nil))
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "application/javascript" || response.Body.Len() == 0 {
		t.Fatalf("Wails runtime was not served: status=%d type=%q", response.Code, response.Header().Get("Content-Type"))
	}
}

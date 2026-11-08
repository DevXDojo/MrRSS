//go:build !server && !production

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestDesktopDevAssetsUseTheSameFrontendForIndexAndModules(t *testing.T) {
	frontend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<script type="module" src="/src/main.ts"></script>`))
		case "/src/main.ts":
			w.Header().Set("Content-Type", "text/javascript")
			_, _ = w.Write([]byte(`import('/src/App.vue')`))
		case "/src/App.vue":
			w.Header().Set("Content-Type", "text/javascript")
			_, _ = w.Write([]byte(`export default { template: '<main>Ready</main>' }`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer frontend.Close()
	t.Setenv("FRONTEND_DEVSERVER_URL", frontend.URL)
	combined := &CombinedHandler{
		apiMux: http.NewServeMux(),
		fileServer: application.BundledAssetFileServer(fstest.MapFS{
			"index.html": &fstest.MapFile{Data: []byte("stale embedded index")},
		}),
	}
	chain := APIMiddleware(combined)(combined)
	for path, want := range map[string]string{
		"/":                 `<script type="module" src="/src/main.ts"></script>`,
		"/?ticker=1":        `<script type="module" src="/src/main.ts"></script>`,
		"/?ticker=settings": `<script type="module" src="/src/main.ts"></script>`,
		"/src/main.ts":      `import('/src/App.vue')`,
		"/src/App.vue":      `export default { template: '<main>Ready</main>' }`,
	} {
		response := httptest.NewRecorder()
		chain.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusOK || response.Body.String() != want {
			t.Errorf("%s: status=%d body=%q, want %q", path, response.Code, response.Body.String(), want)
		}
	}
}

package media

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"MrRSS/internal/database"
	corepkg "MrRSS/internal/handlers/core"
)

func setupHandler(t *testing.T) *corepkg.Handler {
	t.Helper()
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	return corepkg.NewHandler(db, nil, nil, nil)
}

func TestHandleMediaProxy_MethodNotAllowed(t *testing.T) {
	h := setupHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/media/proxy", nil)
	rr := httptest.NewRecorder()

	HandleMediaProxy(h, rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleMediaProxy_CacheDisabled(t *testing.T) {
	h := setupHandler(t)

	// Disable both cache and fallback to test disabled state
	_ = h.DB.SetSetting("media_cache_enabled", "false")
	_ = h.DB.SetSetting("media_proxy_fallback", "false")

	req := httptest.NewRequest(http.MethodGet, "/media/proxy?url=https://example.com/image.jpg", nil)
	rr := httptest.NewRecorder()

	HandleMediaProxy(h, rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Code)
	}
}

func TestHandleMediaProxy_MissingURL(t *testing.T) {
	h := setupHandler(t)
	// enable cache setting
	_ = h.DB.SetSetting("media_cache_enabled", "true")

	req := httptest.NewRequest(http.MethodGet, "/media/proxy", nil)
	rr := httptest.NewRecorder()

	HandleMediaProxy(h, rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleMediaProxy_InvalidURL(t *testing.T) {
	h := setupHandler(t)
	_ = h.DB.SetSetting("media_cache_enabled", "true")

	req := httptest.NewRequest(http.MethodGet, "/media/proxy?url=ftp://example.com/file.jpg", nil)
	rr := httptest.NewRecorder()

	HandleMediaProxy(h, rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestProxyImagesInHTML_RelativeURLs(t *testing.T) {
	referer := "https://example.com/blog/post-123"

	testCases := []struct {
		name           string
		html           string
		expectedSuffix string // The URL should end with this after proxying
		skipProxy      bool   // If true, URL should not be proxied
	}{
		{
			name:           "Absolute URL",
			html:           `<img src="https://cdn.example.com/image.jpg">`,
			expectedSuffix: "url_b64=",
		},
		{
			name:           "Relative path - no slash",
			html:           `<img src="images/photo.jpg">`,
			expectedSuffix: "url_b64=",
		},
		{
			name:           "Relative path - dot slash",
			html:           `<img src="./img.png">`,
			expectedSuffix: "url_b64=",
		},
		{
			name:           "Relative path - parent directory",
			html:           `<img src="../assets/image.gif">`,
			expectedSuffix: "url_b64=",
		},
		{
			name:           "Relative path - multiple parent directories",
			html:           `<img src="../../static/logo.png">`,
			expectedSuffix: "url_b64=",
		},
		{
			name:           "Absolute path - domain relative",
			html:           `<img src="/static/img.png">`,
			expectedSuffix: "url_b64=",
		},
		{
			name:      "Data URL",
			html:      `<img src="data:image/png;base64,iVBORw0KG">`,
			skipProxy: true,
		},
		{
			name:      "Blob URL",
			html:      `<img src="blob:http://localhost/abc-123">`,
			skipProxy: true,
		},
		{
			name:           "Single quoted URL",
			html:           `<img src='images/photo.jpg'>`,
			expectedSuffix: "url_b64=",
		},
		{
			name:           "Unquoted URL (no spaces)",
			html:           `<img src=images/photo.jpg>`,
			expectedSuffix: "url_b64=",
		},
		{
			name:           "URL with HTML entities &amp;",
			html:           `<img src="https://wechat2rss.dev/img-proxy?key=val&amp;other=test">`,
			expectedSuffix: "url_b64=",
		},
		{
			name:           "Relative URL with HTML entities &amp;",
			html:           `<img src="images.jpg?w=800&amp;h=600">`,
			expectedSuffix: "url_b64=",
		},
		{
			name:           "URL with multiple HTML entities",
			html:           `<img src="https://example.com/img?a=1&amp;b=2&amp;c=3">`,
			expectedSuffix: "url_b64=",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ProxyImagesInHTML(tc.html, referer)

			if tc.skipProxy {
				// URL should not be proxied
				if !contains(result, "src=\"") || contains(result, "/api/media/proxy") {
					t.Errorf("Expected URL to remain unchanged, got: %s", result)
				}
			} else {
				// URL should be proxied
				if !contains(result, tc.expectedSuffix) {
					t.Errorf("Expected URL to contain %q, got: %s", tc.expectedSuffix, result)
				}
				if !contains(result, "/api/media/proxy") {
					t.Errorf("Expected proxy URL in result, got: %s", result)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

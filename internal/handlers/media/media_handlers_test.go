package media

import (
	"encoding/base64"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
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

func TestRewriteHTMLContent_ResponsiveImageCandidates(t *testing.T) {
	baseURL := "https://example.com/news/article"
	htmlContent := `<picture>
<source srcSet="/_next/image?url=%2Fhero.jpg&amp;w=1280&amp;q=75 1x, https://cdn.example.com/hero.jpg 2x">
<img src="/fallback.jpg" data-srcset="images/small.jpg 320w, images/large.jpg 1280w">
</picture>`

	result := string(rewriteHTMLContent([]byte(htmlContent), baseURL))
	for _, descriptor := range []string{" 1x", " 2x", " 320w", " 1280w"} {
		if !strings.Contains(result, descriptor) {
			t.Errorf("missing srcset descriptor %q in %s", descriptor, result)
		}
	}
	if strings.Contains(result, `srcSet="/_next/image`) || strings.Contains(result, `data-srcset="images/`) {
		t.Fatalf("responsive image candidates were not proxied: %s", result)
	}

	encodedURLs := regexp.MustCompile(`url_b64=([A-Za-z0-9+/=]+)`).FindAllStringSubmatch(result, -1)
	var decodedURLs []string
	for _, match := range encodedURLs {
		decoded, err := base64.StdEncoding.DecodeString(match[1])
		if err != nil {
			t.Fatalf("decode proxied URL: %v", err)
		}
		decodedURLs = append(decodedURLs, string(decoded))
	}
	joinedURLs := strings.Join(decodedURLs, "\n")
	for _, expected := range []string{
		"https://example.com/_next/image?url=%2Fhero.jpg&w=1280&q=75",
		"https://cdn.example.com/hero.jpg",
		"https://example.com/news/images/small.jpg",
		"https://example.com/news/images/large.jpg",
	} {
		if !strings.Contains(joinedURLs, expected) {
			t.Errorf("missing decoded candidate %q in %s", expected, joinedURLs)
		}
	}
	if strings.Contains(joinedURLs, "&amp;") {
		t.Fatalf("HTML entities leaked into proxied URLs: %s", joinedURLs)
	}
}

func TestRewriteSrcsetAttribute_SkipsNonHTTPAndProxiedCandidates(t *testing.T) {
	content := `<img srcset="data:image/png;base64,AAAA 1x, blob:https://example.com/id 2x, #poster 320w, /api/webpage/resource?url_b64=abc 640w">`
	if got := rewriteSrcsetAttribute(content, "img", "srcset", "https://example.com/article"); got != content {
		t.Fatalf("special srcset candidates changed:\nwant: %s\n got: %s", content, got)
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

func TestRewriteHTMLContentDiscuzAttachments(t *testing.T) {
	base := "https://forum.example.org/thread-1.html"
	input := `<IMG id="aimg_1" src="static/image/common/none.gif" zoomfile="//cdn.example.org/large.jpg?a=1&amp;b=2" file="/thumb.jpg" onclick="zoom(this,this.getAttribute('zoomfile'))"><img src="/none.gif" file="javascript:alert(1)" zoomfile="/real.jpg"><img src="/none.gif" file="/other.jpg" data-src="/preferred.jpg"><img src="/none.gif" zoomfile="https:///missing.jpg" file="/valid.jpg">`
	result := rewriteHTMLContent([]byte(input), base)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(result)))
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"https://cdn.example.org/large.jpg?a=1&b=2", "https://forum.example.org/real.jpg", "https://forum.example.org/preferred.jpg", "https://forum.example.org/valid.jpg"}
	if doc.Find("img").Length() != len(expected) {
		t.Fatalf("unexpected image count: %s", result)
	}
	doc.Find("img").Each(func(i int, img *goquery.Selection) {
		for _, attr := range []string{"src", "zoomfile", "file"} {
			raw, exists := img.Attr(attr)
			if !exists {
				continue
			}
			parsed, err := url.Parse(raw)
			if err != nil {
				t.Fatal(err)
			}
			if parsed.Path != "/api/webpage/resource" {
				t.Errorf("%s is not proxied: %s", attr, raw)
				continue
			}
			decoded, err := base64.StdEncoding.DecodeString(parsed.Query().Get("url_b64"))
			if err != nil {
				t.Fatal(err)
			}
			if attr == "src" && string(decoded) != expected[i] {
				t.Errorf("image %d src=%s, want %s", i, decoded, expected[i])
			}
			if strings.Contains(string(decoded), "none.gif") || strings.Contains(string(decoded), "javascript:") || strings.Contains(string(decoded), "missing.jpg") {
				t.Errorf("invalid attachment source: %s", decoded)
			}
		}
	})
}

func TestConvertLazyImagesKeepsExistingProxies(t *testing.T) {
	proxy := "/api/webpage/resource?url_b64=" + base64.StdEncoding.EncodeToString([]byte("https://cdn.example.org/photo.jpg")) + "&referer_b64=" + base64.StdEncoding.EncodeToString([]byte("https://forum.example.org/thread"))
	for _, input := range []string{`<img src="` + proxy + `" zoomfile="` + proxy + `" file="` + proxy + `">`, `<img src="/none.gif" zoomfile="` + proxy + `">`} {
		output := convertLazyImages(input, "https://forum.example.org/thread")
		output = convertLazyImages(output, "https://forum.example.org/thread")
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(output))
		if err != nil {
			t.Fatal(err)
		}
		img := doc.Find("img").First()
		if img.AttrOr("src", "") != proxy || img.AttrOr("zoomfile", "") != proxy {
			t.Errorf("local proxy changed on repeat: %s", output)
		}
	}
}

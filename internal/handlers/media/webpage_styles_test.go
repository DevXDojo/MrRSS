package media

import (
	"encoding/base64"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

func decodedWebpageResources(t *testing.T, content string) []string {
	t.Helper()
	var resources []string
	for _, match := range regexp.MustCompile(`/api/webpage/resource\?[^\s"'<>\)]+`).FindAllString(html.UnescapeString(content), -1) {
		parsed, err := url.Parse(match)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := base64.StdEncoding.DecodeString(parsed.Query().Get("url_b64"))
		if err != nil {
			t.Fatal(err)
		}
		resources = append(resources, string(decoded))
	}
	return resources
}

func TestForumStylesheetURLsAndInlineStyles(t *testing.T) {
	base := "https://forum.example.org/thread/42"
	content := `<LINK REL="stylesheet" HREF="../assets/forum.css?v=1&amp;theme=light"><div style='background-image: url("../img/bg.png?v=1&amp;size=2")'></div>`
	result := rewriteInlineStyles(rewriteLinkHref(content, base), base)
	resources := strings.Join(decodedWebpageResources(t, result), "\n")
	for _, want := range []string{"https://forum.example.org/assets/forum.css?v=1&theme=light", "https://forum.example.org/img/bg.png?v=1&size=2"} {
		if !strings.Contains(resources, want) {
			t.Errorf("missing resource %s in %s", want, resources)
		}
	}
	// Standard base64 can contain '+', which query parsing otherwise turns into a space.
	original := "https://example.org/?q=~~~"
	proxied, ok := proxyWebpageResourceURL(original, base)
	if !ok {
		t.Fatal("resource was not proxied")
	}
	got := decodedWebpageResources(t, proxied)
	if len(got) != 1 || got[0] != original {
		t.Fatalf("resource changed during query round trip: %v", got)
	}
}

func TestStylesheetProxyUsesItsOwnBaseAndMIMEWithVersionQuery(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/assets/redirect.css" {
			http.Redirect(w, r, "/static/css/forum.css?v=2", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, `@import "./mobile.css" screen; @font-face { src: url('../fonts/forum.woff2'); } .post { background: URL(../../images/bg.png); mask: url(#mask); cursor: url(data:image/png;base64,AAAA); }`)
	}))
	defer upstream.Close()
	h := setupHandler(t)
	defer h.DB.Close()
	query := url.Values{"url": {upstream.URL + "/assets/redirect.css?v=1"}, "referer": {upstream.URL + "/forum.php?mod=viewthread"}}
	req := httptest.NewRequest(http.MethodGet, "/api/webpage/resource?"+query.Encode(), nil)
	w := httptest.NewRecorder()
	HandleWebpageResource(h, w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.HasPrefix(w.Header().Get("Content-Type"), "text/css") {
		t.Fatalf("wrong CSS MIME: %s", w.Header().Get("Content-Type"))
	}
	resources := strings.Join(decodedWebpageResources(t, w.Body.String()), "\n")
	for _, path := range []string{"/static/css/mobile.css", "/static/fonts/forum.woff2", "/images/bg.png"} {
		if !strings.Contains(resources, upstream.URL+path) {
			t.Errorf("incorrect stylesheet-relative path %s in %s", path, resources)
		}
	}
	for _, unchanged := range []string{"url(#mask)", "url(data:image/png;base64,AAAA)"} {
		if !strings.Contains(w.Body.String(), unchanged) {
			t.Errorf("changed local CSS reference %s", unchanged)
		}
	}
}

func TestWebpageProxyIsolatesActiveContent(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Content-Security-Policy", "script-src *")
		w.Header().Set("Refresh", "0; url=/api/settings")
		w.Header().Set("Set-Cookie", "injected=yes")
		fmt.Fprint(w, `<meta http-equiv="refresh" content="0;url=/api/settings"><script src="evil.js"></script><img src="/pic.png" onerror="evil()"><iframe srcdoc="evil"></iframe><a href="javascript:evil()">unsafe</a><a href="../news?a=1&amp;b=2">news</a>`)
	}))
	defer upstream.Close()
	h := setupHandler(t)
	defer h.DB.Close()
	query := url.Values{"url": {upstream.URL + "/articles/original"}, "referer": {upstream.URL}}
	for _, endpoint := range []struct {
		path   string
		handle http.HandlerFunc
	}{
		{"/api/webpage/proxy", func(w http.ResponseWriter, r *http.Request) { HandleWebpageProxy(h, w, r) }},
		{"/api/webpage/resource", func(w http.ResponseWriter, r *http.Request) { HandleWebpageResource(h, w, r) }},
	} {
		t.Run(endpoint.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			endpoint.handle(w, httptest.NewRequest(http.MethodGet, endpoint.path+"?"+query.Encode(), nil))
			if w.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			csp := w.Header().Get("Content-Security-Policy")
			if !strings.HasPrefix(csp, "sandbox ") || !strings.Contains(csp, "default-src 'none'") || strings.Contains(csp, "allow-scripts") || strings.Contains(csp, "allow-same-origin") {
				t.Errorf("unsafe response policy: %q", csp)
			}
			if w.Header().Get("Refresh") != "" || w.Header().Get("Set-Cookie") != "" {
				t.Errorf("unexpected upstream headers: %v", w.Header())
			}
			if endpoint.path != "/api/webpage/proxy" {
				return
			}
			body := w.Body.String()
			for _, active := range []string{"<script", "onerror", "<iframe", "http-equiv", "javascript:"} {
				if strings.Contains(body, active) {
					t.Errorf("preview retained %q", active)
				}
			}
			if !strings.Contains(body, `href="`+upstream.URL+`/news?a=1&amp;b=2" target="_blank" rel="noopener noreferrer"`) {
				t.Errorf("outbound link was lost: %s", body)
			}
			if got := decodedWebpageResources(t, body); len(got) != 1 || got[0] != upstream.URL+"/pic.png" {
				t.Errorf("wrong preview resources: %v", got)
			}
		})
	}
}

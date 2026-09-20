package media

import (
	"MrRSS/internal/adfilter"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
)

func TestWebpageAdResourceStopsBeforeNetworkAndHonorsSiteException(t *testing.T) {
	h := setupHandler(t)
	defer h.DB.Close()
	var calls atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("fixture resource"))
	}))
	defer proxy.Close()
	u, _ := url.Parse(proxy.URL)
	host, port, _ := net.SplitHostPort(u.Host)
	for key, value := range map[string]string{"proxy_enabled": "true", "proxy_type": "http", "proxy_host": host, "proxy_port": port} {
		h.DB.SetSetting(key, value)
	}
	options := adfilter.DefaultConfig()
	for _, allowed := range []bool{false, true} {
		if allowed {
			options.Allowlist = "example.org"
		}
		raw, _ := json.Marshal(options)
		h.DB.SetSetting("ad_filter_config", string(raw))
		w := httptest.NewRecorder()
		HandleWebpageResource(h, w, httptest.NewRequest("GET", "/api/webpage/resource?url="+url.QueryEscape("http://pagead2.googlesyndication.com/ad.js")+"&referer="+url.QueryEscape("https://news.example.org/story"), nil))
		if !allowed && (w.Code != 204 || calls.Load() != 0 || w.Header().Get("X-MrRSS-Filtered") == "") {
			t.Fatalf("unblocked resource %d calls=%d", w.Code, calls.Load())
		}
		if allowed && (w.Code != 200 || calls.Load() != 1) {
			t.Fatalf("exception not honored %d calls=%d", w.Code, calls.Load())
		}
	}
}

package media

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
)

func TestMediaProxyUsesApplicationProxy(t *testing.T) {
	for _, cacheEnabled := range []bool{false, true} {
		name := "fallback"
		if cacheEnabled {
			name = "cache"
		}
		t.Run(name, func(t *testing.T) {
			tmp := t.TempDir()
			t.Setenv("APPDATA", tmp)
			t.Setenv("HOME", tmp)
			t.Setenv("XDG_DATA_HOME", tmp)
			h := setupHandler(t)
			defer h.DB.Close()
			var requests atomic.Int32
			proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.URL.String() != "http://media.invalid/photo.png" {
					t.Errorf("unexpected proxy target: %s", r.URL)
				}
				wantAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("test-user:test-password"))
				if r.Header.Get("Proxy-Authorization") != wantAuth {
					t.Error("proxy did not receive configured credentials")
				}
				w.Header().Set("Content-Type", "image/png")
				_, _ = w.Write([]byte("image data"))
			}))
			defer proxy.Close()
			proxyURL, _ := url.Parse(proxy.URL)
			host, port, err := net.SplitHostPort(proxyURL.Host)
			if err != nil {
				t.Fatal(err)
			}
			settings := map[string]string{
				"proxy_enabled": "true", "proxy_type": "http", "proxy_host": host, "proxy_port": port,
				"media_cache_enabled": "false", "media_proxy_fallback": "true",
			}
			if cacheEnabled {
				settings["media_cache_enabled"] = "true"
				settings["media_proxy_fallback"] = "false"
			}
			for key, value := range settings {
				if err := h.DB.SetSetting(key, value); err != nil {
					t.Fatal(err)
				}
			}
			for key, value := range map[string]string{"proxy_username": "test-user", "proxy_password": "test-password"} {
				if err := h.DB.SetEncryptedSetting(key, value); err != nil {
					t.Fatal(err)
				}
			}
			for i := 0; i < 2; i++ {
				req := httptest.NewRequest(http.MethodGet, "/api/media/proxy?url="+url.QueryEscape("http://media.invalid/photo.png"), nil)
				rr := httptest.NewRecorder()
				HandleMediaProxy(h, rr, req)
				if rr.Code != http.StatusOK || rr.Body.String() != "image data" || rr.Header().Get("Content-Type") != "image/png" {
					t.Fatalf("unexpected media response: %d %s", rr.Code, rr.Body.String())
				}
				wantSource := "direct-proxy"
				if cacheEnabled {
					wantSource = "cache"
				}
				if rr.Header().Get("X-Media-Source") != wantSource {
					t.Errorf("unexpected source: %s", rr.Header().Get("X-Media-Source"))
				}
			}
			wantRequests := int32(2)
			if cacheEnabled {
				wantRequests = 1
			}
			if requests.Load() != wantRequests {
				t.Errorf("proxy requests = %d, want %d", requests.Load(), wantRequests)
			}
		})
	}
}

func TestMediaProxyDisabledApplicationProxy(t *testing.T) {
	h := setupHandler(t)
	defer h.DB.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Proxy-Authorization") != "" {
			t.Error("proxy credentials leaked to media host")
		}
		_, _ = w.Write([]byte("direct image"))
	}))
	defer server.Close()
	for key, value := range map[string]string{
		"proxy_enabled": "false", "proxy_host": "127.0.0.1", "proxy_port": "1",
		"media_cache_enabled": "false", "media_proxy_fallback": "true",
	} {
		if err := h.DB.SetSetting(key, value); err != nil {
			t.Fatal(err)
		}
	}
	rr := httptest.NewRecorder()
	HandleMediaProxy(h, rr, httptest.NewRequest(http.MethodGet, "/api/media/proxy?url="+url.QueryEscape(server.URL), nil))
	if rr.Code != http.StatusOK || rr.Body.String() != "direct image" {
		t.Fatalf("unexpected direct response: %d %s", rr.Code, rr.Body.String())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	rr = httptest.NewRecorder()
	HandleMediaProxy(h, rr, httptest.NewRequest(http.MethodGet, "/api/media/proxy?url="+url.QueryEscape(server.URL), nil).WithContext(ctx))
	if rr.Code == http.StatusOK {
		t.Fatal("cancelled download succeeded")
	}
}

func TestMediaProxyRefererFallbackAndExhaustion(t *testing.T) {
	for _, cacheEnabled := range []bool{false, true} {
		for _, outcome := range []string{"first-success", "retry-success", "forbidden"} {
			t.Run(fmt.Sprintf("cache=%t/%s", cacheEnabled, outcome), func(t *testing.T) {
				tmp := t.TempDir()
				t.Setenv("APPDATA", tmp)
				t.Setenv("HOME", tmp)
				t.Setenv("XDG_DATA_HOME", tmp)
				h := setupHandler(t)
				defer h.DB.Close()
				if err := h.DB.SetSetting("media_cache_enabled", fmt.Sprint(cacheEnabled)); err != nil {
					t.Fatal(err)
				}
				if err := h.DB.SetSetting("media_proxy_fallback", "true"); err != nil {
					t.Fatal(err)
				}
				var requests atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					n := requests.Add(1)
					if n == 1 && r.Header.Get("Referer") == "" {
						t.Error("first request lost original referer")
					}
					if outcome == "forbidden" || (outcome == "retry-success" && r.Header.Get("Referer") != "") {
						w.WriteHeader(http.StatusForbidden)
						return
					}
					w.Header().Set("Content-Type", "image/png")
					_, _ = w.Write([]byte("image data"))
				}))
				defer server.Close()
				address := "/api/media/proxy?url=" + url.QueryEscape(server.URL+"/photo.png") + "&referer=" + url.QueryEscape(server.URL+"/article")
				rec := httptest.NewRecorder()
				HandleMediaProxy(h, rec, httptest.NewRequest(http.MethodGet, address, nil))
				wantStatus := http.StatusOK
				wantRequests := int32(1)
				if outcome != "first-success" {
					wantRequests = 2
				}
				if outcome == "forbidden" {
					wantStatus = http.StatusInternalServerError
				}
				if rec.Code != wantStatus || requests.Load() != wantRequests {
					t.Fatalf("status=%d requests=%d, want status=%d requests=%d; %s", rec.Code, requests.Load(), wantStatus, wantRequests, rec.Body.String())
				}
				if wantStatus == http.StatusOK && (rec.Body.String() != "image data" || rec.Header().Get("Content-Type") != "image/png") {
					t.Fatalf("invalid image response: %s", rec.Body.String())
				}
				if cacheEnabled && wantStatus == http.StatusOK {
					cached := httptest.NewRecorder()
					HandleMediaProxy(h, cached, httptest.NewRequest(http.MethodGet, address, nil))
					if cached.Code != http.StatusOK || requests.Load() != wantRequests {
						t.Fatal("second load did not use successful cache")
					}
				}
			})
		}
	}
}

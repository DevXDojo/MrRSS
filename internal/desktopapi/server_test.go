package desktopapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestValidateLoopbackAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{name: "IPv4 loopback", address: "127.0.0.1:1234"},
		{name: "IPv6 loopback", address: "[::1]:1234"},
		{name: "ephemeral loopback port", address: "127.0.0.1:0"},
		{name: "all interfaces", address: "0.0.0.0:1234", wantErr: true},
		{name: "non-loopback", address: "192.0.2.10:1234", wantErr: true},
		{name: "missing port", address: "127.0.0.1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLoopbackAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateLoopbackAddress(%q) error = %v, wantErr %v", tt.address, err, tt.wantErr)
			}
		})
	}
}

func TestProtectLocalAPI(t *testing.T) {
	handler := protectLocalAPI(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name           string
		host           string
		origin         string
		secFetchSite   string
		wantStatusCode int
	}{
		{name: "local agent request", host: "127.0.0.1:1234", wantStatusCode: http.StatusNoContent},
		{name: "localhost agent request", host: "localhost:1234", wantStatusCode: http.StatusNoContent},
		{name: "foreign host", host: "attacker.example", wantStatusCode: http.StatusForbidden},
		{name: "browser origin", host: "127.0.0.1:1234", origin: "https://attacker.example", wantStatusCode: http.StatusForbidden},
		{name: "cross-site fetch", host: "127.0.0.1:1234", secFetchSite: "cross-site", wantStatusCode: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/version", nil)
			req.Host = tt.host
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.secFetchSite != "" {
				req.Header.Set("Sec-Fetch-Site", tt.secFetchSite)
			}
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestServerLifecycle(t *testing.T) {
	server, err := Start("127.0.0.1:0", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/version" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	response, err := http.Get("http://" + server.Address() + "/api/version")
	if err != nil {
		t.Fatalf("GET desktop API: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusNoContent)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := <-server.Errors(); err != nil {
		t.Fatalf("Serve() error after shutdown = %v", err)
	}
}

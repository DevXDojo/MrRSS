package httputil

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestBuildProxyURLPreservesCredentialsAndIPv6(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "::1", "[::1]"} {
		raw := BuildProxyURL("http", host, "7890", "user@company", "p@ss:/?#%")
		parsed, err := url.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		password, _ := parsed.User.Password()
		if parsed.User.Username() != "user@company" || password != "p@ss:/?#%" {
			t.Fatal("proxy credentials changed during URL construction")
		}
		if parsed.Port() != "7890" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			t.Fatal("proxy credentials leaked into URL components")
		}
	}
}

func TestCreateHTTPClientHonorsInsecureTLSVerifyEnv(t *testing.T) {
	t.Setenv(InsecureSkipTLSVerifyEnv, "true")

	client, err := CreateHTTPClient("", time.Second)
	if err != nil {
		t.Fatalf("CreateHTTPClient returned error: %v", err)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport type %T", client.Transport)
	}
	if transport.TLSClientConfig == nil {
		t.Fatalf("TLSClientConfig is nil")
	}
	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatalf("expected InsecureSkipVerify to be true")
	}
	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Fatalf("expected TLS 1.2 minimum, got %d", transport.TLSClientConfig.MinVersion)
	}
}

func TestCreateHTTPClientKeepsTLSVerificationByDefault(t *testing.T) {
	t.Setenv(InsecureSkipTLSVerifyEnv, "")

	client, err := CreateHTTPClient("", time.Second)
	if err != nil {
		t.Fatalf("CreateHTTPClient returned error: %v", err)
	}

	transport := client.Transport.(*http.Transport)
	if transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatalf("expected InsecureSkipVerify to be false by default")
	}
}

func TestCreateHTTPClientNegotiatesHTTP2(t *testing.T) {
	t.Setenv(InsecureSkipTLSVerifyEnv, "true")

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 2 {
			t.Errorf("expected HTTP/2 request, got %s", r.Proto)
		}
		_, _ = io.WriteString(w, "ok")
	}))
	server.EnableHTTP2 = true
	server.StartTLS()
	defer server.Close()

	client, err := CreateHTTPClient("", 5*time.Second)
	if err != nil {
		t.Fatalf("CreateHTTPClient returned error: %v", err)
	}
	response, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("HTTP/2 request failed: %v", err)
	}
	defer response.Body.Close()
}

func TestCreateHTTPClientFallsBackToHTTP1(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 1 {
			t.Errorf("expected HTTP/1.1 request, got %s", r.Proto)
		}
		_, _ = io.WriteString(w, "ok")
	}))
	defer server.Close()

	client, err := CreateHTTPClient("", 5*time.Second)
	if err != nil {
		t.Fatalf("CreateHTTPClient returned error: %v", err)
	}
	response, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("HTTP/1.1 request failed: %v", err)
	}
	defer response.Body.Close()
}

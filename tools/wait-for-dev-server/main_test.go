package main

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestFrontendClientConnectsToIPv4Localhost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer server.Close()
	address := strings.Replace(server.URL, "127.0.0.1", "localhost", 1)
	client := frontendClient(address)
	defer client.CloseIdleConnections()
	response, err := client.Get(address)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected readiness status: %d", response.StatusCode)
	}
}

func TestFrontendClientRejectsIPv6OnlyLocalhost(t *testing.T) {
	listener, err := net.Listen("tcp6", "[::1]:0")
	if err != nil {
		t.Skipf("IPv6 loopback unavailable: %v", err)
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	server.Listener.Close()
	server.Listener = listener
	server.Start()
	defer server.Close()
	address := "http://localhost:" + strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
	client := frontendClient(address)
	defer client.CloseIdleConnections()
	response, err := client.Get(address)
	if response != nil {
		response.Body.Close()
	}
	if err == nil {
		t.Fatal("an IPv6-only listener must not mark the IPv4 Wails proxy ready")
	}
}

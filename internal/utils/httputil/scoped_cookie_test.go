package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScopedCookieStaysOnConfiguredOriginAcrossRedirects(t *testing.T) {
	var targetCookie string
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { targetCookie = r.Header.Get("Cookie") }))
	defer target.Close()
	var firstCookie string
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		firstCookie = r.Header.Get("Cookie")
		http.Redirect(w, r, target.URL, 302)
	}))
	defer source.Close()
	client := WithScopedCookie(source.Client(), source.URL, "session=secret")
	res, err := client.Get(source.URL)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if firstCookie != "session=secret" || targetCookie != "" {
		t.Fatalf("cookies source=%q target=%q", firstCookie, targetCookie)
	}
}

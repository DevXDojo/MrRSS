package httputil

import (
	"net/http"
	"net/url"
	"strings"
)

type scopedCookieTransport struct {
	base   http.RoundTripper
	origin *url.URL
	cookie string
}

func (t scopedCookieTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	if strings.EqualFold(req.URL.Scheme, t.origin.Scheme) && strings.EqualFold(req.URL.Host, t.origin.Host) {
		clone.Header.Set("Cookie", t.cookie)
	}
	return t.base.RoundTrip(clone)
}

// WithScopedCookie attaches a Cookie only to the configured exact origin, including redirects.
func WithScopedCookie(client *http.Client, origin, cookie string) *http.Client {
	parsed, err := url.Parse(origin)
	if err != nil || cookie == "" || parsed.Host == "" {
		return client
	}
	copyClient := *client
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	copyClient.Transport = scopedCookieTransport{transport, parsed, cookie}
	return &copyClient
}

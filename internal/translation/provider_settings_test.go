package translation

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFactoryUsesProxySettingsForTraditionalProviders(t *testing.T) {
	settings := &mockSettingsProvider{settings: map[string]string{
		"proxy_enabled": "true", "proxy_type": "http", "proxy_host": "127.0.0.1", "proxy_port": "7890",
		"proxy_username": "proxy-user", "proxy_password": "proxy-password",
		"deepl_api_key": "test-key", "baidu_app_id": "test-id", "baidu_secret_key": "test-secret",
		"microsoft_api_key": "test-key", "microsoft_region": "eastasia", "microsoft_endpoint": "https://translator.example",
		"tencent_secret_id": "test-id", "tencent_secret_key": "test-secret", "tencent_region": "ap-shanghai",
	}}
	for _, providerType := range []ProviderType{ProviderGoogle, ProviderDeepL, ProviderBaidu, ProviderMicrosoft, ProviderTencent} {
		t.Run(providerType.String(), func(t *testing.T) {
			provider, err := NewFactory(settings).Create(providerType)
			if err != nil {
				t.Fatal(err)
			}
			var client *http.Client
			switch p := provider.(type) {
			case *googleProvider:
				client = p.translator.client
				if p.translator.db != settings {
					t.Error("Google endpoint settings were not connected")
				}
			case *deepLProvider:
				client = p.translator.client
			case *baiduProvider:
				client = p.translator.client
			case *microsoftProvider:
				client = p.translator.client
				if p.translator.Region != "eastasia" || p.translator.Endpoint != "https://translator.example" {
					t.Error("Microsoft region and endpoint must both be retained")
				}
			case *tencentProvider:
				client = p.translator.client
			}
			if client == nil {
				t.Fatal("missing provider client")
			}
			defer client.CloseIdleConnections()
			transport, ok := client.Transport.(*http.Transport)
			if !ok || transport.Proxy == nil {
				t.Fatal("missing configured proxy transport")
			}
			req, _ := http.NewRequest(http.MethodGet, "https://translation.example", nil)
			proxy, err := transport.Proxy(req)
			if err != nil || proxy == nil || proxy.Host != "127.0.0.1:7890" || proxy.User.Username() != "proxy-user" {
				t.Fatal("incorrect proxy configuration")
			}
			if password, _ := proxy.User.Password(); password != "proxy-password" {
				t.Error("proxy password not loaded")
			}
		})
	}
}

func TestGoogleAlternateEndpointResponses(t *testing.T) {
	for _, body := range []string{
		`{"sentences":[{"trans":"Bonjour"},{"trans":" le monde"}]}`,
		`[["Bonjour le monde","en"]]`,
		`["Bonjour le monde"]`,
	} {
		settings := &mockSettingsProvider{settings: map[string]string{"google_translate_endpoint": "clients5.google.com"}}
		provider := NewFactory(settings).createGoogleProvider(ProviderConfig{}).(*googleProvider)
		provider.translator.client = &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Host != "clients5.google.com" || req.URL.Path != "/translate_a/t" || req.URL.Query().Get("q") != "Hello world" {
				t.Errorf("unexpected request endpoint: %s", req.URL)
			}
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body))}, nil
		})}
		result, err := provider.Translate(context.Background(), "Hello world", "fr")
		if err != nil || result.Translated != "Bonjour le monde" {
			t.Fatalf("unexpected translation: %+v, %v", result, err)
		}
	}
	for _, body := range []string{`null`, `{}`, `[]`, `{"error":"unavailable"}`, `[[""]]`} {
		if _, err := decodeGoogleDictionaryResponse([]byte(body)); err == nil {
			t.Errorf("accepted invalid response: %s", body)
		}
	}
}

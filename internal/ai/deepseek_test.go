package ai

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
)

func TestDeepSeekFormatEndpointNormalizesBaseURLs(t *testing.T) {
	handler := &DeepSeekHandler{}

	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{
			name:     "empty endpoint uses default",
			endpoint: "",
			want:     "https://api.deepseek.com/v1/chat/completions",
		},
		{
			name:     "provider base URL",
			endpoint: "https://api.deepseek.com",
			want:     "https://api.deepseek.com/v1/chat/completions",
		},
		{
			name:     "v1 base URL",
			endpoint: "https://api.deepseek.com/v1",
			want:     "https://api.deepseek.com/v1/chat/completions",
		},
		{
			name:     "full chat completions URL",
			endpoint: "https://api.deepseek.com/v1/chat/completions",
			want:     "https://api.deepseek.com/v1/chat/completions",
		},
		{
			name:     "custom compatible route",
			endpoint: "https://gateway.example.com/deepseek/v1/chat/completions",
			want:     "https://gateway.example.com/deepseek/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := handler.FormatEndpoint(tt.endpoint, "deepseek-v4-flash"); got != tt.want {
				t.Fatalf("FormatEndpoint(%q) = %q, want %q", tt.endpoint, got, tt.want)
			}
		})
	}
}

func TestClassifyUserFacingErrorDoesNotExposeProviderResponse(t *testing.T) {
	const secret = `{"error":{"message":"provider detail sk-secret-value"}}`
	tests := []struct {
		name string
		err  error
		code string
	}{
		{name: "authentication", err: errors.New("OpenAI API returned status 401: " + secret), code: ErrorCodeAuthenticationFailed},
		{name: "payment", err: errors.New("OpenAI API returned status 402: " + secret), code: ErrorCodePaymentRequired},
		{name: "not found", err: errors.New("OpenAI API returned status 404: " + secret), code: ErrorCodeModelOrEndpointNotFound},
		{name: "too large", err: errors.New("OpenAI API returned status 413: " + secret), code: ErrorCodeRequestTooLarge},
		{name: "rate limited", err: errors.New("OpenAI API returned status 429: " + secret), code: ErrorCodeRateLimited},
		{name: "provider unavailable", err: errors.New("OpenAI API returned status 503: " + secret), code: ErrorCodeProviderUnavailable},
		{name: "timeout", err: context.DeadlineExceeded, code: ErrorCodeTimeout},
		{name: "network", err: &net.DNSError{Err: "no such host", Name: "private.example"}, code: ErrorCodeNetwork},
		{name: "invalid response", err: errors.New("invalid JSON response"), code: ErrorCodeInvalidResponse},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyUserFacingError(tt.err)
			if got.Code != tt.code {
				t.Fatalf("code = %q, want %q", got.Code, tt.code)
			}
			if got.Message == "" || strings.Contains(got.Message, "sk-secret-value") || strings.Contains(got.Message, "provider detail") {
				t.Fatalf("unsafe user-facing message: %q", got.Message)
			}
		})
	}

	if got := RedactEndpoint("https://user:secret@api.example.com/v1/chat?token=secret#fragment"); got != "https://api.example.com/v1/chat" {
		t.Fatalf("RedactEndpoint leaked or changed the endpoint: %q", got)
	}
}

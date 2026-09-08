package ai

import "testing"

func TestOpenAIFormatEndpointNormalizesCompatibleBaseURLs(t *testing.T) {
	handler := NewOpenAIHandler()
	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{name: "default", endpoint: "", want: "https://api.openai.com/v1/chat/completions"},
		{name: "OpenAI base", endpoint: "https://api.openai.com/v1", want: "https://api.openai.com/v1/chat/completions"},
		{name: "OpenRouter base", endpoint: "https://openrouter.ai/api/v1", want: "https://openrouter.ai/api/v1/chat/completions"},
		{name: "OpenRouter full URL", endpoint: "https://openrouter.ai/api/v1/chat/completions", want: "https://openrouter.ai/api/v1/chat/completions"},
		{name: "custom route", endpoint: "https://gateway.example.com/custom/chat", want: "https://gateway.example.com/custom/chat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := handler.FormatEndpoint(tt.endpoint, "model"); got != tt.want {
				t.Fatalf("FormatEndpoint(%q) = %q, want %q", tt.endpoint, got, tt.want)
			}
		})
	}
}

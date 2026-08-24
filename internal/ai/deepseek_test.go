package ai

import "testing"

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

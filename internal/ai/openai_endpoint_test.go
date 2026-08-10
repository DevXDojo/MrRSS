package ai

import "testing"

func TestOpenAIFormatEndpointNormalizesBaseURLs(t *testing.T) {
	handler := NewOpenAIHandler()

	tests := []struct {
		name     string
		endpoint string
		want     string
	}{
		{
			name:     "empty defaults to openai",
			endpoint: "",
			want:     "https://api.openai.com/v1/chat/completions",
		},
		{
			name:     "host only",
			endpoint: "http://10.0.1.4:8080",
			want:     "http://10.0.1.4:8080/v1/chat/completions",
		},
		{
			name:     "v1 base path",
			endpoint: "http://10.0.1.4:8080/v1",
			want:     "http://10.0.1.4:8080/v1/chat/completions",
		},
		{
			name:     "v1 with trailing slash",
			endpoint: "http://10.0.1.4:8080/v1/",
			want:     "http://10.0.1.4:8080/v1/chat/completions",
		},
		{
			name:     "full chat completions path",
			endpoint: "http://10.0.1.4:8080/v1/chat/completions",
			want:     "http://10.0.1.4:8080/v1/chat/completions",
		},
		{
			name:     "compatible mode v1",
			endpoint: "https://dashscope.aliyuncs.com/compatible-mode/v1",
			want:     "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions",
		},
		{
			name:     "custom gateway path kept",
			endpoint: "https://gateway.example.com/proxy/openai",
			want:     "https://gateway.example.com/proxy/openai",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := handler.FormatEndpoint(tt.endpoint, "qwen-local"); got != tt.want {
				t.Fatalf("FormatEndpoint(%q) = %q, want %q", tt.endpoint, got, tt.want)
			}
		})
	}
}

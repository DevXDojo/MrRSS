package rsshub

import "testing"

func TestClientBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		apiKey   string
		route    string
		want     string
	}{
		{
			name:     "simple route",
			endpoint: "https://rsshub.example.com/",
			route:    "weibo/user/123",
			want:     "https://rsshub.example.com/weibo/user/123",
		},
		{
			name:     "leading slash route",
			endpoint: "https://rsshub.example.com",
			route:    "/weibo/user/123",
			want:     "https://rsshub.example.com/weibo/user/123",
		},
		{
			name:     "api key is appended",
			endpoint: "https://rsshub.example.com",
			apiKey:   "secret",
			route:    "weibo/user/123",
			want:     "https://rsshub.example.com/weibo/user/123?key=secret",
		},
		{
			name:     "route query is preserved when api key is appended",
			endpoint: "https://rsshub.example.com",
			apiKey:   "secret",
			route:    "weibo/user/123?limit=20",
			want:     "https://rsshub.example.com/weibo/user/123?key=secret&limit=20",
		},
		{
			name:     "api key is escaped",
			endpoint: "https://rsshub.example.com",
			apiKey:   "a b&c",
			route:    "weibo/user/123",
			want:     "https://rsshub.example.com/weibo/user/123?key=a+b%26c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.endpoint, tt.apiKey)
			if got := client.BuildURL(tt.route); got != tt.want {
				t.Fatalf("BuildURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

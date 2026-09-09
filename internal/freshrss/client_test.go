package freshrss

import "testing"

func TestNewClientForProviderUsesCorrectGoogleReaderRoot(t *testing.T) {
	tests := []struct {
		name      string
		provider  string
		serverURL string
		wantRoot  string
	}{
		{
			name:      "FreshRSS appends greader endpoint",
			provider:  string(ProviderFreshRSS),
			serverURL: "https://freshrss.example.com/",
			wantRoot:  "https://freshrss.example.com/api/greader.php",
		},
		{
			name:      "Miniflux uses server root",
			provider:  string(ProviderMiniflux),
			serverURL: "https://miniflux.example.com/miniflux/",
			wantRoot:  "https://miniflux.example.com/miniflux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClientForProvider(tt.serverURL, "user", "password", tt.provider)
			if client.baseURL != tt.wantRoot {
				t.Fatalf("baseURL = %q, want %q", client.baseURL, tt.wantRoot)
			}
		})
	}
}

func TestGoogleReaderUpdatedTimeSupportsSecondsAndMilliseconds(t *testing.T) {
	if got := googleReaderUpdatedTime(1_710_000_300).Unix(); got != 1_710_000_300 {
		t.Fatalf("seconds timestamp = %d", got)
	}
	if got := googleReaderUpdatedTime(1_710_000_300_000).Unix(); got != 1_710_000_300 {
		t.Fatalf("milliseconds timestamp = %d", got)
	}
}

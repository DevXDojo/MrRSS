package ai

import (
	"net/url"
	"strings"
)

// DetectAPIProvider detects the wire protocol from an explicit API path first.
// Domain names describe a provider, which may expose several protocols.
func DetectAPIProvider(endpoint string) string {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return "unknown"
	}
	path := strings.ToLower(strings.TrimRight(parsed.Path, "/"))
	host := strings.ToLower(parsed.Hostname())
	switch {
	case strings.HasSuffix(path, "/messages"):
		return "anthropic"
	case strings.HasSuffix(path, "/chat/completions"):
		return "openai"
	case strings.HasSuffix(path, ":generatecontent"):
		return "gemini"
	case strings.HasSuffix(path, "/api/chat"), strings.HasSuffix(path, "/api/generate"):
		return "ollama"
	}
	switch {
	case host == "generativelanguage.googleapis.com", strings.Contains(host, "gemini"):
		return "gemini"
	case host == "api.anthropic.com", strings.Contains(host, "claude"):
		return "anthropic"
	case strings.Contains(host, "deepseek"):
		return "deepseek"
	case host == "localhost", host == "127.0.0.1", host == "::1", strings.Contains(host, "ollama"):
		return "ollama"
	case host == "api.openai.com":
		return "openai"
	}
	return "unknown"
}

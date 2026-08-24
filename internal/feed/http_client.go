package feed

import (
	"MrRSS/internal/utils/httputil"
)

// BuildProxyURL constructs a proxy URL from settings
// Wrapper around httputil.BuildProxyURL for backward compatibility
func BuildProxyURL(proxyType, proxyHost, proxyPort, username, password string) string {
	return httputil.BuildProxyURL(proxyType, proxyHost, proxyPort, username, password)
}

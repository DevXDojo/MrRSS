//go:build !server

package main

// DesktopCapabilities is bound only to the native WebView. The loopback API
// and server build cannot mint or retrieve this per-process update capability.
type DesktopCapabilities struct{ updateAuthorization string }

func (s *DesktopCapabilities) UpdateAuthorization() string { return s.updateAuthorization }

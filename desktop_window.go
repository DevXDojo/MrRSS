//go:build !server

package main

import "github.com/wailsapp/wails/v3/pkg/application"

// Showing an existing native window must not reset its normal bounds or call
// Restore(), which also unmaximizes a visible window.
func showExistingWindow(window application.Window, maximize bool) {
	window.Show()
	window.UnMinimise()
	if maximize {
		window.Maximise()
	}
	window.Focus()
}

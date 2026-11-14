//go:build !server

package desktopapi

import "github.com/wailsapp/wails/v3/pkg/application"

// ShowExistingWindow must not reset its normal bounds or call
// Restore(), which also unmaximizes a visible window.
func ShowExistingWindow(window application.Window, maximize bool) {
	window.Show()
	window.UnMinimise()
	if maximize {
		window.Maximise()
	}
	window.Focus()
}

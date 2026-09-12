//go:build !server

package main

import (
	"fmt"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Configure rendering before creating the application: GTK and WebKit read
// these environment variables when initializing their renderers.
func configureLinuxRendering(platform string, args []string) (application.LinuxWindow, error) {
	options := application.LinuxWindow{}
	if platform != "linux" {
		return options, nil
	}
	for _, arg := range args {
		if arg == "--" {
			break
		}
		if arg != "--software-rendering" {
			continue
		}
		// The explicit option overrides renderer preferences for this process.
		// Without it, preserve both the environment and Wails' defaults.
		if err := os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1"); err != nil {
			return options, fmt.Errorf("disable WebKit DMA-BUF rendering: %w", err)
		}
		if err := os.Setenv("GSK_RENDERER", "cairo"); err != nil {
			return options, fmt.Errorf("select GTK software renderer: %w", err)
		}
		options.WebviewGpuPolicy = application.WebviewGpuPolicyNever
		break
	}
	return options, nil
}

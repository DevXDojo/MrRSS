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
	software, enableDMABUF := false, false
	for _, arg := range args {
		if arg == "--" {
			break
		}
		switch arg {
		case "--software-rendering":
			software = true
		case "--enable-dmabuf-renderer":
			enableDMABUF = true
		}
	}
	if software && enableDMABUF {
		return options, fmt.Errorf("--software-rendering and --enable-dmabuf-renderer cannot be combined")
	}
	if enableDMABUF {
		// Wails may have set this variable during package initialization simply
		// because the NVIDIA module is loaded, even on hybrid GPU systems.
		if err := os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "0"); err != nil {
			return options, fmt.Errorf("enable WebKit DMA-BUF rendering: %w", err)
		}
	}
	if software {
		// The explicit option overrides renderer preferences for this process.
		// Without it, preserve both the environment and Wails' defaults.
		if err := os.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "1"); err != nil {
			return options, fmt.Errorf("disable WebKit DMA-BUF rendering: %w", err)
		}
		if err := os.Setenv("GSK_RENDERER", "cairo"); err != nil {
			return options, fmt.Errorf("select GTK software renderer: %w", err)
		}
		options.WebviewGpuPolicy = application.WebviewGpuPolicyNever
	}
	return options, nil
}

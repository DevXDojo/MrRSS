//go:build !server

package main

import (
	"os"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestLinuxSoftwareRenderingOptIn(t *testing.T) {
	for _, tc := range []struct {
		name     string
		platform string
		args     []string
		enabled  bool
	}{
		{"default", "linux", nil, false},
		{"explicit", "linux", []string{"--software-rendering"}, true},
		{"after other argument", "linux", []string{"--portable", "--software-rendering"}, true},
		{"end of options", "linux", []string{"--", "--software-rendering"}, false},
		{"similar argument", "linux", []string{"--software-rendering=false"}, false},
		{"windows", "windows", []string{"--software-rendering"}, false},
		{"macOS", "darwin", []string{"--software-rendering"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("WEBKIT_DISABLE_DMABUF_RENDERER", "0")
			t.Setenv("GSK_RENDERER", "gl")
			options, err := configureLinuxRendering(tc.platform, tc.args)
			if err != nil {
				t.Fatal(err)
			}
			wantPolicy := application.WebviewGpuPolicyAlways
			wantDMABUF, wantRenderer := "0", "gl"
			if tc.enabled {
				wantPolicy = application.WebviewGpuPolicyNever
				wantDMABUF, wantRenderer = "1", "cairo"
			}
			if options.WebviewGpuPolicy != wantPolicy {
				t.Errorf("GPU policy = %v, want %v", options.WebviewGpuPolicy, wantPolicy)
			}
			if got := os.Getenv("WEBKIT_DISABLE_DMABUF_RENDERER"); got != wantDMABUF {
				t.Errorf("DMA-BUF setting = %q, want %q", got, wantDMABUF)
			}
			if got := os.Getenv("GSK_RENDERER"); got != wantRenderer {
				t.Errorf("GTK renderer = %q, want %q", got, wantRenderer)
			}
		})
	}
}

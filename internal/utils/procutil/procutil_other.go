//go:build !windows

package procutil

import "os/exec"

// HideWindow is a no-op on platforms where child processes never open a console
// window of their own.
func HideWindow(cmd *exec.Cmd) {}

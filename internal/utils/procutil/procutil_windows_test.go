//go:build windows

package procutil

import (
	"os/exec"
	"syscall"
	"testing"
)

func TestHideWindowSetsHideFlag(t *testing.T) {
	cmd := exec.Command("cmd", "/c", "exit 0")

	HideWindow(cmd)

	if cmd.SysProcAttr == nil {
		t.Fatal("HideWindow() left SysProcAttr nil")
	}
	if !cmd.SysProcAttr.HideWindow {
		t.Error("HideWindow should set SysProcAttr.HideWindow")
	}
}

func TestHideWindowKeepsExistingAttr(t *testing.T) {
	cmd := exec.Command("cmd", "/c", "exit 0")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000004}

	HideWindow(cmd)

	if !cmd.SysProcAttr.HideWindow {
		t.Error("HideWindow should set SysProcAttr.HideWindow")
	}
	if cmd.SysProcAttr.CreationFlags != 0x00000004 {
		t.Errorf("CreationFlags = %#x, want 0x4 preserved", cmd.SysProcAttr.CreationFlags)
	}
}

func TestHideWindowNilCommand(t *testing.T) {
	HideWindow(nil) // must not panic
}

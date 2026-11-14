//go:build windows

package procutil

import (
	"os/exec"
	"syscall"
)

// HideWindow keeps cmd from showing a console window to the user.
//
// MrRSS is a GUI application, so it has no console of its own. Windows therefore
// allocates a fresh console for every console-subsystem child process, and that
// console is displayed by default — which makes helpers such as python, node or
// reg.exe flash a black window on screen every time a feed refreshes.
//
// Setting SysProcAttr.HideWindow makes Go pass STARTF_USESHOWWINDOW with
// SW_HIDE, so the console is created but never shown.
//
// Note: CREATE_NO_WINDOW is deliberately not used. Go creates child processes
// through STARTUPINFOEX with _EXTENDED_STARTUPINFO_PRESENT, and Windows ignores
// the flag in that path — verified by measuring GetConsoleCP and
// GetConsoleProcessList in the child process. DETACHED_PROCESS would remove the
// console entirely, but it also detaches the process from the caller's console
// group, which is a behavioural change for user scripts that is not worth the
// risk here.
func HideWindow(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
}

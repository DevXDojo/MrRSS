package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinuxStartupUsesPersistentAppImageAndXDGDirectory(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configDir)
	image := filepath.Join(t.TempDir(), "Mr RSS.AppImage")
	t.Setenv("APPIMAGE", image)
	if err := enableStartupLinux(filepath.Join(t.TempDir(), ".mount_MrRSS", "MrRSS")); err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(configDir, "autostart", "mrrss.desktop")
	body, err := os.ReadFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	want, err := desktopExec(image)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "Exec="+want+" --start-minimized\n") || strings.Contains(string(body), ".mount_") {
		t.Fatalf("unexpected entry: %s", body)
	}
	if err := disableStartupLinux(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(entry); !os.IsNotExist(err) {
		t.Fatalf("entry still exists: %v", err)
	}
	if err := disableStartupLinux(); err != nil {
		t.Fatal(err)
	}
}

func TestDesktopExecEscapingAndValidation(t *testing.T) {
	base := filepath.VolumeName(t.TempDir()) + string(filepath.Separator)
	for _, tc := range []struct{ name, escaped string }{
		{"Mr RSS", "Mr RSS"}, {"100%", "100%%"}, {`a"b`, `a\\"b`}, {"$HOME", `\\$HOME`}, {"a`b", "a\\\\`b"},
	} {
		got, err := desktopExec(base + tc.name)
		if err != nil || !strings.HasSuffix(got, tc.escaped+`"`) || !strings.HasPrefix(got, `"`) {
			t.Errorf("%q: %q, %v", tc.name, got, err)
		}
	}
	for _, invalid := range []string{"relative.AppImage", base + "bad\nExec=evil", base + "bad\rpath", base + "bad\x00path"} {
		if _, err := desktopExec(invalid); err == nil {
			t.Errorf("accepted invalid executable %q", invalid)
		}
	}
}

func TestLinuxStartupFallsBackToExecutable(t *testing.T) {
	t.Setenv("APPIMAGE", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	executable := filepath.Join(t.TempDir(), "MrRSS")
	if err := enableStartupLinux(executable); err != nil {
		t.Fatal(err)
	}
	dir, err := linuxAutostartDir()
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "mrrss.desktop"))
	if err != nil {
		t.Fatal(err)
	}
	want, _ := desktopExec(executable)
	if !strings.Contains(string(body), "Exec="+want+" --start-minimized\n") {
		t.Fatalf("unexpected entry: %s", body)
	}
}

func TestStartupDarwinArgumentsRequestsMinimizedLaunch(t *testing.T) {
	executable := filepath.Join(t.TempDir(), "MrRSS")
	args := startupDarwinArguments(executable)
	if !strings.Contains(args, "<string>--start-minimized</string>") {
		t.Fatalf("missing minimized launch argument: %s", args)
	}
}

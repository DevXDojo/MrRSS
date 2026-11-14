package fileutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestConfigureDataDirThroughJunction(t *testing.T) {
	previous := customDataDir
	t.Cleanup(func() { customDataDir = previous })
	dir := t.TempDir()
	target := filepath.Join(dir, "persist")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("cmd.exe", "/d", "/c", "mklink", "/J", "data", "persist")
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("create junction: %v: %s", err, output)
	}
	if err := ConfigureDataDir(target); err != nil {
		t.Fatal(err)
	}
	want := CustomDataDir()
	if err := ConfigureDataDir(filepath.Join(dir, "data")); err != nil {
		t.Fatal(err)
	}
	if got := CustomDataDir(); got != want {
		t.Errorf("junction resolved to %q, want %q", got, want)
	}
}

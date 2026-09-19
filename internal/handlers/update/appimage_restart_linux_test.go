package update

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestAppImageRestartEnvironment(t *testing.T) {
	got := appImageRestartEnv([]string{"APPIMAGE=/old", "APPDIR=/mount", "ARGV0=old", "PATH=/mount/usr/bin:/usr/bin", "LD_LIBRARY_PATH=/mount/usr/lib:/custom", "DISPLAY=:0"}, "/mount")
	want := []string{"PATH=/usr/bin", "LD_LIBRARY_PATH=/custom", "DISPLAY=:0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("env = %v", got)
	}
}

func TestAppImageRestartWaitsForExitAndPreservesArguments(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	old := exec.CommandContext(ctx, "/bin/sleep", "60")
	if err := old.Start(); err != nil {
		t.Fatal(err)
	}
	defer old.Process.Kill()
	output := filepath.Join(t.TempDir(), "result with spaces")
	value := "quoted ' ; $(touch should-not-exist)"
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", appImageRestartScript, "test", strconv.Itoa(old.Process.Pid), "/bin/sh", "-c", `printf '%s' "$1" > "$2"`, "test", value, output)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatal("new instance started before old exit")
	}
	_ = old.Process.Kill()
	_ = old.Wait()
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(output); err != nil || string(got) != value {
		t.Fatalf("arguments = %q, %v", got, err)
	}
}

//go:build linux

package singleinstance

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInstanceLockLifecycle(t *testing.T) {
	dir := t.TempDir()
	first, err := Acquire(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { first.Close() })
	if second, err := Acquire(dir); err == nil {
		second.Close()
		t.Fatal("duplicate instance acquired lock")
	}
	other, err := Acquire(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	other.Close()
	first.Close()
	if _, err := os.Stat(filepath.Join(dir, "mrrss.lock")); err != nil {
		t.Fatal("lock file must survive unlock", err)
	}
	third, err := Acquire(dir)
	if err != nil {
		t.Fatal("stale lock file must not block launch", err)
	}
	third.Close()
}

func TestInstanceLockProcessExit(t *testing.T) {
	if dir := os.Getenv("MRRSS_TEST_LOCK_DIR"); dir != "" {
		lock, err := Acquire(dir)
		if err != nil {
			os.Exit(2)
		}
		_ = lock
		os.Exit(0) // No Close: kernel must release the lock on exit.
	}
	dir := t.TempDir()
	held, err := Acquire(dir)
	if err != nil {
		t.Fatal(err)
	}
	blocked := exec.Command(os.Args[0], "-test.run=^TestInstanceLockProcessExit$")
	blocked.Env = append(os.Environ(), "MRRSS_TEST_LOCK_DIR="+dir)
	if err := blocked.Run(); err == nil {
		held.Close()
		t.Fatal("second process acquired held lock")
	}
	held.Close()
	cmd := exec.Command(os.Args[0], "-test.run=^TestInstanceLockProcessExit$")
	cmd.Env = append(os.Environ(), "MRRSS_TEST_LOCK_DIR="+dir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("child: %v %s", err, output)
	}
	lock, err := Acquire(dir)
	if err != nil {
		t.Fatal("exited process retained lock", err)
	}
	lock.Close()
}

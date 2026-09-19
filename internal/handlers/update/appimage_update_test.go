package update

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceAppImage(t *testing.T) {
	for _, scenario := range []string{"success", "invalid", "cancelled"} {
		t.Run(scenario, func(t *testing.T) {
			dir := t.TempDir()
			target := filepath.Join(dir, "installed with spaces.AppImage")
			download := filepath.Join(t.TempDir(), "new.AppImage")
			payload := []byte("\x7fELF\x02\x01\x01\x00AI\x02new-image")
			if scenario == "invalid" {
				payload = []byte("<html>download failed</html>")
			}
			if err := os.WriteFile(target, []byte("old-image"), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(download, payload, 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if scenario == "cancelled" {
				cancel()
			}
			got, err := replaceAppImage(ctx, download, target)
			want := "old-image"
			if scenario == "success" {
				if err != nil || got != target {
					t.Fatalf("replace = %q, %v", got, err)
				}
				want = string(payload)
			} else if err == nil {
				t.Fatal("expected replacement failure")
			}
			data, err := os.ReadFile(target)
			if err != nil || string(data) != want {
				t.Fatalf("installed image = %q, %v", data, err)
			}
			entries, err := os.ReadDir(dir)
			if err != nil || len(entries) != 1 {
				t.Fatalf("staging files left behind: %v, %v", entries, err)
			}
		})
	}
}

func TestUpdateDownloadContainmentAndCleanup(t *testing.T) {
	path, err := createUpdateDownloadPath("update.AppImage")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(path); _ = os.Remove(filepath.Dir(path)) })
	if err := os.WriteFile(path, []byte("image"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := validateUpdateDownload(path); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []string{"relative.AppImage", filepath.Join(os.TempDir(), "unrelated.AppImage"), filepath.Join(t.TempDir(), "update.AppImage")} {
		if _, err := validateUpdateDownload(invalid); err == nil {
			t.Fatalf("accepted %q", invalid)
		}
	}
	cleanupUpdateDownload(path)
	if _, err := os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
		t.Fatalf("download directory remains: %v", err)
	}
}

func TestUpdateDownloadRejectsSymlink(t *testing.T) {
	path, err := createUpdateDownloadPath("update.AppImage")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(path); _ = os.Remove(filepath.Dir(path)) })
	target := filepath.Join(t.TempDir(), "untouched")
	if err := os.WriteFile(target, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := validateUpdateDownload(path); err == nil {
		t.Fatal("accepted symlink")
	}
	cleanupUpdateDownload(path)
	if data, err := os.ReadFile(target); err != nil || string(data) != "keep" {
		t.Fatal("cleanup modified unrelated file")
	}
}

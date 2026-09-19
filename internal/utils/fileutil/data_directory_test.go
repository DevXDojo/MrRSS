package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCustomDataDirectory(t *testing.T) {
	previous, server := customDataDir, isServerMode
	t.Cleanup(func() { customDataDir, isServerMode = previous, server })
	envDir := filepath.Join(t.TempDir(), "环境 目录")
	t.Setenv("MRRSS_DATA_DIR", envDir)
	if err := ConfigureDataDir(""); err != nil {
		t.Fatal(err)
	}
	if got, _ := GetDataDir(); got != envDir {
		t.Fatalf("environment path = %q", got)
	}
	explicit := filepath.Join(t.TempDir(), "explicit space")
	if err := ConfigureDataDir(explicit); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []bool{false, true} {
		SetServerMode(mode)
		for _, get := range []func() (string, error){GetDBPath, GetLogPath, GetScriptsDir} {
			path, err := get()
			if err != nil || !strings.HasPrefix(path, explicit+string(filepath.Separator)) {
				t.Fatalf("path = %q, %v", path, err)
			}
		}
	}
	marker := filepath.Join(explicit, "keep")
	if err := os.WriteFile(marker, []byte("existing data"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ConfigureDataDir(explicit); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(marker); err != nil || string(data) != "existing data" {
		t.Fatal("existing data changed")
	}
	if err := ConfigureDataDir(marker); err == nil {
		t.Fatal("accepted a file as directory")
	}
	if customDataDir != explicit {
		t.Fatal("failed configuration changed active directory")
	}
}

func TestDataDirArgument(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
		fail bool
	}{
		{[]string{"--software-rendering", "--data-dir", "path with 空格"}, "path with 空格", false},
		{[]string{"--data-dir=folder"}, "folder", false},
		{[]string{"--", "--data-dir=ignored"}, "", false},
		{[]string{"--data-dir"}, "", true},
		{[]string{"--data-dir="}, "", true},
		{[]string{"--data-dir", "--software-rendering"}, "", true},
	} {
		got, err := DataDirArgument(tc.args)
		if got != tc.want || (err != nil) != tc.fail {
			t.Errorf("%v: %q, %v", tc.args, got, err)
		}
	}
}

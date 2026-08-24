package fileutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetDataDir(t *testing.T) {
	// Test that GetDataDir returns a valid path
	dir, err := GetDataDir()
	if err != nil {
		t.Fatalf("GetDataDir failed: %v", err)
	}

	if dir == "" {
		t.Error("GetDataDir returned empty string")
	}

	// Check that directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("Data directory does not exist: %s", dir)
	}

	// Check that it ends with MrRSS (in normal mode) or data (in portable mode)
	if !IsPortableMode() && !strings.HasSuffix(dir, "MrRSS") {
		t.Errorf("Expected path to end with MrRSS in normal mode, got: %s", dir)
	}
	if IsPortableMode() && !strings.HasSuffix(dir, "data") {
		t.Errorf("Expected path to end with data in portable mode, got: %s", dir)
	}
}

func TestGetDBPath(t *testing.T) {
	path, err := GetDBPath()
	if err != nil {
		t.Fatalf("GetDBPath failed: %v", err)
	}

	if path == "" {
		t.Error("GetDBPath returned empty string")
	}

	// Check that it ends with rss.db
	if !strings.HasSuffix(path, "rss.db") {
		t.Errorf("Expected path to end with rss.db, got: %s", path)
	}

	// Check that parent directory exists
	parentDir := filepath.Dir(path)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		t.Errorf("Parent directory does not exist: %s", parentDir)
	}
}

func TestGetLogPath(t *testing.T) {
	path, err := GetLogPath()
	if err != nil {
		t.Fatalf("GetLogPath failed: %v", err)
	}

	if path == "" {
		t.Error("GetLogPath returned empty string")
	}

	// Check that it ends with debug.log
	if !strings.HasSuffix(path, "debug.log") {
		t.Errorf("Expected path to end with debug.log, got: %s", path)
	}
}

func TestGetDataDir_PlatformSpecific(t *testing.T) {
	// Test platform-specific behavior
	switch runtime.GOOS {
	case "windows":
		// On Windows, should use APPDATA or USERPROFILE
		originalAppData := os.Getenv("APPDATA")
		originalUserProfile := os.Getenv("USERPROFILE")
		defer func() {
			os.Setenv("APPDATA", originalAppData)
			os.Setenv("USERPROFILE", originalUserProfile)
		}()

		// Test with APPDATA set
		os.Setenv("APPDATA", "C:\\TestAppData")
		dir, err := GetDataDir()
		if err != nil {
			t.Fatalf("GetDataDir failed: %v", err)
		}
		if !strings.Contains(dir, "TestAppData") {
			t.Errorf("Expected path to contain TestAppData, got: %s", dir)
		}

	case "darwin":
		// On macOS, should use HOME/Library/Application Support
		originalHome := os.Getenv("HOME")
		defer os.Setenv("HOME", originalHome)

		// Use tmp to avoid permission issues
		os.Setenv("HOME", "/tmp/Users/testuser")
		dir, err := GetDataDir()
		if err != nil {
			t.Fatalf("GetDataDir failed: %v", err)
		}
		expected := filepath.Join("/tmp/Users/testuser", "Library", "Application Support", "MrRSS")
		if dir != expected {
			t.Errorf("Expected %s, got %s", expected, dir)
		}

	case "linux":
		// On Linux, should use XDG_DATA_HOME or HOME/.local/share
		originalXDG := os.Getenv("XDG_DATA_HOME")
		originalHome := os.Getenv("HOME")
		defer func() {
			os.Setenv("XDG_DATA_HOME", originalXDG)
			os.Setenv("HOME", originalHome)
		}()

		os.Setenv("XDG_DATA_HOME", "/tmp/test-xdg")
		dir, err := GetDataDir()
		if err != nil {
			t.Fatalf("GetDataDir failed: %v", err)
		}
		if !strings.Contains(dir, "test-xdg") {
			t.Errorf("Expected path to contain test-xdg, got: %s", dir)
		}
	}
}

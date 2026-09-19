package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Set once at startup, before logging, locking, database access or goroutines.
var customDataDir string

func CustomDataDir() string { return customDataDir }

// Preserve the canonical directory when an updater restarts from a different
// working directory, including an originally relative --data-dir argument.
func DataDirRestartArgs(args []string) []string {
	if customDataDir == "" {
		return append([]string(nil), args...)
	}
	result := []string{}
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			result = append(result, "--data-dir", customDataDir)
			return append(result, args[i:]...)
		}
		if args[i] == "--data-dir" {
			i++
			continue
		}
		if strings.HasPrefix(args[i], "--data-dir=") {
			continue
		}
		result = append(result, args[i])
	}
	return append(result, "--data-dir", customDataDir)
}

func ConfigureDataDir(path string) error {
	if path == "" {
		path = os.Getenv("MRRSS_DATA_DIR")
	}
	if path == "" {
		customDataDir = ""
		return nil
	}
	if strings.ContainsAny(path, "\x00\r\n") {
		return fmt.Errorf("invalid data directory")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve data directory: %w", err)
	}
	if err := os.MkdirAll(abs, 0700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	// Freeze the canonical path so the instance lock and all data consumers use
	// the same directory, including launches through a symlink or relative path.
	abs, err = filepath.EvalSymlinks(abs)
	if err != nil {
		return fmt.Errorf("resolve data directory: %w", err)
	}
	probe, err := os.CreateTemp(abs, ".mrrss-write-check-*")
	if err != nil {
		return fmt.Errorf("data directory is not writable: %w", err)
	}
	if err := probe.Close(); err != nil {
		_ = os.Remove(probe.Name())
		return err
	}
	if err := os.Remove(probe.Name()); err != nil {
		return err
	}
	customDataDir = abs
	return nil
}

// Desktop uses Wails' own argument handling; consume only our option and leave
// rendering/platform arguments untouched.
func DataDirArgument(args []string) (string, error) {
	value := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		if arg == "--data-dir" {
			if i+1 == len(args) || strings.HasPrefix(args[i+1], "--") {
				return "", fmt.Errorf("--data-dir requires a path")
			}
			i++
			value = args[i]
		} else if strings.HasPrefix(arg, "--data-dir=") {
			value = strings.TrimPrefix(arg, "--data-dir=")
		} else {
			continue
		}
		if value == "" {
			return "", fmt.Errorf("--data-dir requires a nonempty path")
		}
	}
	return value, nil
}

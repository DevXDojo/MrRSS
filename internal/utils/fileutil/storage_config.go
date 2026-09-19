package fileutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// This bootstrap setting must live outside the database it selects.
// Written only by the explicit directory-change action, never by autosave.
type StorageConfig struct {
	DataDirectory    string `json:"data_directory"`
	PendingDirectory string `json:"pending_directory,omitempty"`
	LastError        string `json:"last_error,omitempty"`
}

var storageMu sync.Mutex
var desktopStorageManaged bool
var storageConfigFile string

func DesktopStorageManaged() bool { return desktopStorageManaged }

func storageConfigPath() (string, error) {
	if IsPortableMode() {
		exe, err := os.Executable()
		if err != nil {
			return "", err
		}
		return filepath.Join(filepath.Dir(exe), "mrrss-storage.json"), nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "MrRSS-bootstrap", "storage.json"), nil
}

func readStorageConfig(path string) (StorageConfig, error) {
	var cfg StorageConfig
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("read storage configuration: %w", err)
	}
	return cfg, nil
}

func writeStorageConfig(path string, cfg StorageConfig) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".storage-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(file.Name(), path)
}

// InitializeDesktopStorage runs before logging, SQLite or background work.
// The returned lock is held for the process lifetime to prevent migration while
// another desktop instance still uses the source. CLI/environment keep priority.
func InitializeDesktopStorage(option string) (io.Closer, error) {
	storageMu.Lock()
	defer storageMu.Unlock()
	desktopStorageManaged = option == "" && os.Getenv("MRRSS_DATA_DIR") == ""
	if !desktopStorageManaged {
		if err := ConfigureDataDir(option); err != nil {
			return nil, err
		}
		dir, err := GetDataDir()
		if err != nil {
			return nil, err
		}
		lock, err := acquireStorageUse(dir)
		if isStorageLockBusy(err) {
			return nil, nil
		}
		if err != nil {
			return nil, err
		}
		return lock, nil
	}
	path, err := storageConfigPath()
	if err != nil {
		return nil, err
	}
	storageConfigFile = path
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	lock, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := lockStorageFile(lock); err != nil {
		lock.Close()
		return nil, fmt.Errorf("MrRSS is already running; fully quit it before changing storage")
	}
	defer lock.Close()
	fail := func(err error) (io.Closer, error) { return nil, err }
	cfg, err := readStorageConfig(path)
	if err != nil {
		return fail(err)
	}
	if cfg.DataDirectory == "" {
		cfg.DataDirectory, err = DefaultDataDir()
		if err != nil {
			return fail(err)
		}
	} else if info, err := os.Stat(cfg.DataDirectory); err != nil || !info.IsDir() {
		return fail(fmt.Errorf("configured data directory is unavailable: %s", cfg.DataDirectory))
	}
	if err := ConfigureDataDir(cfg.DataDirectory); err != nil {
		return fail(err)
	}
	cfg.DataDirectory = customDataDir
	useLock, lockErr := acquireStorageUse(cfg.DataDirectory)
	// A second launch must still reach Wails so it can raise the existing window.
	// It must never migrate while the primary process owns this directory.
	if isStorageLockBusy(lockErr) {
		return nil, nil
	}
	if lockErr != nil {
		return fail(lockErr)
	}
	if cfg.PendingDirectory != "" {
		target := cfg.PendingDirectory
		if err := migrateStorage(cfg.DataDirectory, target); err != nil {
			cfg.LastError = err.Error()
		} else {
			cfg.DataDirectory = target
			cfg.LastError = ""
		}
		cfg.PendingDirectory = ""
		if err := writeStorageConfig(path, cfg); err != nil {
			useLock.Close()
			return fail(fmt.Errorf("save migrated storage location: %w", err))
		}
	}
	if cfg.DataDirectory != customDataDir {
		useLock.Close()
		if err := ConfigureDataDir(cfg.DataDirectory); err != nil {
			return fail(err)
		}
		useLock, err = acquireStorageUse(cfg.DataDirectory)
		if err != nil {
			return fail(err)
		}
	}
	return useLock, nil
}

func DesktopStorageStatus() (StorageConfig, error) {
	storageMu.Lock()
	defer storageMu.Unlock()
	if !desktopStorageManaged {
		current, err := GetDataDir()
		return StorageConfig{DataDirectory: current}, err
	}
	cfg, err := readStorageConfig(storageConfigFile)
	if err != nil {
		return cfg, err
	}
	cfg.DataDirectory, err = GetDataDir()
	return cfg, err
}

func ScheduleDataDirectory(path string) (StorageConfig, error) {
	storageMu.Lock()
	defer storageMu.Unlock()
	var cfg StorageConfig
	if !desktopStorageManaged {
		return cfg, fmt.Errorf("data directory is controlled by launch configuration")
	}
	current, err := GetDataDir()
	if err != nil {
		return cfg, err
	}
	target, err := validateStorageDestination(current, path)
	if err != nil {
		return cfg, err
	}
	cfg = StorageConfig{DataDirectory: current, PendingDirectory: target}
	return cfg, writeStorageConfig(storageConfigFile, cfg)
}

func CancelDataDirectoryChange() error {
	storageMu.Lock()
	defer storageMu.Unlock()
	if !desktopStorageManaged {
		return fmt.Errorf("data directory is controlled by launch configuration")
	}
	current, err := GetDataDir()
	if err != nil {
		return err
	}
	return writeStorageConfig(storageConfigFile, StorageConfig{DataDirectory: current})
}

func validateStorageDestination(source, destination string) (string, error) {
	if !filepath.IsAbs(destination) || strings.ContainsAny(destination, "\x00\r\n") {
		return "", fmt.Errorf("select an absolute directory path")
	}
	source, err := filepath.EvalSymlinks(source)
	if err != nil {
		return "", err
	}
	destination, err = filepath.EvalSymlinks(destination)
	if err != nil {
		return "", fmt.Errorf("destination must be an existing empty folder: %w", err)
	}
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return "", err
	}
	targetInfo, err := os.Stat(destination)
	if err != nil {
		return "", err
	}
	if !targetInfo.IsDir() || os.SameFile(sourceInfo, targetInfo) {
		return "", fmt.Errorf("choose a different empty directory")
	}
	within := func(parent, child string) bool {
		rel, err := filepath.Rel(parent, child)
		return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
	}
	if within(source, destination) || within(destination, source) {
		return "", fmt.Errorf("source and destination cannot contain each other")
	}
	entries, err := os.ReadDir(destination)
	if err != nil {
		return "", err
	}
	if len(entries) != 0 {
		return "", fmt.Errorf("destination must be empty; existing files will not be overwritten")
	}
	probe, err := os.CreateTemp(destination, ".mrrss-probe-*")
	if err != nil {
		return "", err
	}
	name := probe.Name()
	closeErr := probe.Close()
	removeErr := os.Remove(name)
	if closeErr != nil {
		return "", closeErr
	}
	if removeErr != nil {
		return "", removeErr
	}
	return destination, nil
}

func acquireStorageUse(dir string) (*os.File, error) {
	file, err := os.OpenFile(filepath.Join(dir, ".mrrss-storage.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := lockStorageFile(file); err != nil {
		file.Close()
		return nil, err
	}
	return file, nil
}

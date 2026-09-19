package fileutil

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Copy into a sibling staging directory, then publish it. Source remains an
// untouched backup. Refuse links/special files and nonempty targets. On failure
// the application keeps using the source and surfaces LastError in settings.
func migrateStorage(source, target string) error {
	canonical, err := validateStorageDestination(source, target)
	if err != nil {
		return err
	}
	if canonical != target {
		return fmt.Errorf("destination changed since selection")
	}
	staging, err := os.MkdirTemp(filepath.Dir(target), ".mrrss-migrate-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	err = filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		// Process-specific descriptors must be recreated, not migrated.
		if relative == ".mrrss-storage.lock" || relative == "mrrss.lock" {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("cannot migrate symbolic link: %s", relative)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		dest := filepath.Join(staging, relative)
		if entry.IsDir() {
			return os.Mkdir(dest, info.Mode().Perm()|0700)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("cannot migrate special file: %s", relative)
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		defer input.Close()
		output, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, info.Mode().Perm())
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(output, input)
		if copyErr == nil {
			copyErr = output.Sync()
		}
		closeErr := output.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
	if err != nil {
		return fmt.Errorf("copy data: %w", err)
	}
	// Revalidate immediately before publication; never merge with existing files.
	if _, err := validateStorageDestination(source, target); err != nil {
		return err
	}
	if err := os.Remove(target); err != nil {
		return err
	} // empty directory only
	if err := os.Rename(staging, target); err != nil {
		_ = os.Mkdir(target, 0700)
		return err
	}
	return nil
}

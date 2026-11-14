//go:build !windows

package fileutil

import "path/filepath"

func resolveStorageDirectory(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

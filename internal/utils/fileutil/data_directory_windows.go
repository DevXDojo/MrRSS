package fileutil

import (
	"os"
	"strings"

	"golang.org/x/sys/windows"
)

func resolveStorageDirectory(path string) (string, error) {
	// Opening the directory lets Windows follow junctions and symbolic links.
	// filepath.EvalSymlinks can leave junctions unresolved as irregular files.
	dir, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer dir.Close()

	buffer := make([]uint16, windows.MAX_PATH)
	for {
		n, err := windows.GetFinalPathNameByHandle(windows.Handle(dir.Fd()), &buffer[0], uint32(len(buffer)), 0)
		if err != nil {
			return "", &os.PathError{Op: "resolve directory", Path: path, Err: err}
		}
		if int(n) >= len(buffer) {
			buffer = make([]uint16, n+1)
			continue
		}
		resolved := windows.UTF16ToString(buffer[:n])
		if unc, ok := strings.CutPrefix(resolved, `\\?\UNC\`); ok {
			return `\\` + unc, nil
		}
		return strings.TrimPrefix(resolved, `\\?\`), nil
	}
}

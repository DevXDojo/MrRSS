//go:build linux

// Package singleinstance protects the Linux desktop database before startup.
package singleinstance

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// Acquire holds a kernel lock until Close or process exit. The file must never
// be unlinked: another process could otherwise lock a different inode.
func Acquire(dataDir string) (io.Closer, error) {
	f, err := os.OpenFile(filepath.Join(dataDir, "mrrss.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("open desktop instance lock: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, fmt.Errorf("MrRSS is already running for this data directory; open it from the system tray")
		}
		return nil, fmt.Errorf("lock desktop instance: %w", err)
	}
	return f, nil
}

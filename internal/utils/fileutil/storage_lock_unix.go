//go:build !windows

package fileutil

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
)

func isStorageLockBusy(err error) bool {
	return errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN)
}

func lockStorageFile(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
}

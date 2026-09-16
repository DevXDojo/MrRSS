//go:build !linux

package singleinstance

import "io"

// Acquire leaves Windows and macOS single-instance handling to Wails.
func Acquire(dataDir string) (io.Closer, error) { return nil, nil }

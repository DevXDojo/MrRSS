//go:build !linux

package update

import "fmt"

func startAppImageAfterExit(string, []string) error {
	return fmt.Errorf("AppImage restart is only supported on Linux")
}

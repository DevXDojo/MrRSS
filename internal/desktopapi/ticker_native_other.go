//go:build !windows && !server

package desktopapi

import (
	"errors"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type tickerNativeWindow struct{}

func nativeTickerPositions() []string { return nil }

func (*tickerNativeWindow) windowMessage(uintptr, uint32, uintptr, uintptr) (uintptr, bool) {
	return 0, false
}

func (*tickerNativeWindow) apply(*application.WebviewWindow, *application.Screen, string, int) error {
	return errors.New("native ticker modes are unavailable")
}

func (*tickerNativeWindow) release() {}

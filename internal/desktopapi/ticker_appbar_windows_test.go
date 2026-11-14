//go:build !server

package desktopapi

import (
	"testing"
	"unsafe"
)

func TestTickerAppbarLayout(t *testing.T) {
	wantStructSize := uintptr(36)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantStructSize = 48
	}
	if got := unsafe.Sizeof(tickerAppBarData{}); got != wantStructSize {
		t.Fatalf("APPBARDATA layout has size %d, want %d", got, wantStructSize)
	}
	monitor := tickerWinRect{left: -1920, top: -120, right: 0, bottom: 960}
	approved := tickerWinRect{left: -1880, top: 894, right: 0, bottom: 912}
	if got := tickerAppbarBounds(approved, monitor, 66); got != (tickerWinRect{-1880, 846, 0, 912}) {
		t.Errorf("ticker must preserve the shell-approved bottom and side edges: %+v", got)
	}
	approved.bottom = -90
	if got := tickerAppbarBounds(approved, monitor, 66); got.top != monitor.top {
		t.Errorf("ticker escaped the display: %+v", got)
	}
}

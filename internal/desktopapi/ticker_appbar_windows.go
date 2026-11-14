//go:build !server

package desktopapi

import (
	"fmt"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/sys/windows"
)

var shellAppBarMessage = windows.NewLazySystemDLL("shell32.dll").NewProc("SHAppBarMessage")

const (
	appbarNew              = 0
	appbarRemove           = 1
	appbarQueryPos         = 2
	appbarSetPos           = 3
	appbarActivate         = 6
	appbarWindowPosChanged = 9
	appbarBottomEdge       = 3
	appbarStateChange      = 0
	appbarPosChanged       = 1
	appbarFullscreen       = 2
)

type tickerWinRect struct{ left, top, right, bottom int32 }

type tickerMonitorInfo struct {
	size          uint32
	monitor, work tickerWinRect
	flags         uint32
}

type tickerAppBarData struct {
	size           uint32
	hwnd           uintptr
	callback, edge uint32
	rect           tickerWinRect
	param          uintptr
}

func rectForTickerScreen(r application.Rect) tickerWinRect {
	return tickerWinRect{int32(r.X), int32(r.Y), int32(r.X + r.Width), int32(r.Y + r.Height)}
}

func (data *tickerAppBarData) send(message uintptr) bool {
	data.size = uint32(unsafe.Sizeof(*data))
	result, _, _ := shellAppBarMessage.Call(message, uintptr(unsafe.Pointer(data)))
	return result != 0
}

func (n *tickerNativeWindow) placeAppbar(info tickerMonitorInfo, height int32) error {
	data := tickerAppBarData{hwnd: n.hwnd, callback: n.appbarMessage, edge: appbarBottomEdge}
	if !n.registered {
		if !data.send(appbarNew) {
			return fmt.Errorf("register ticker as a Windows appbar")
		}
		n.registered = true
	}
	data.rect = info.monitor
	if !data.send(appbarQueryPos) {
		return fmt.Errorf("query ticker appbar position")
	}
	// The shell may shrink or move the proposed rectangle to avoid the taskbar
	// and other appbars. Restore the requested thickness above its bottom edge.
	data.rect = tickerAppbarBounds(data.rect, info.monitor, height)
	if data.rect.right <= data.rect.left || data.rect.bottom <= data.rect.top {
		return fmt.Errorf("no space for ticker appbar on selected display")
	}
	if data.rect != n.requested {
		requested := data.rect
		if !data.send(appbarSetPos) {
			return fmt.Errorf("reserve space for ticker appbar")
		}
		if data.rect.right <= data.rect.left || data.rect.bottom <= data.rect.top {
			return fmt.Errorf("Windows returned an empty ticker appbar reservation")
		}
		n.requested, n.reservation = requested, data.rect
	}
	n.setBounds(n.reservation)
	n.setAppbarOrder()
	return nil
}

func tickerAppbarBounds(approved, monitor tickerWinRect, height int32) tickerWinRect {
	approved.top = max(monitor.top, approved.bottom-height)
	return approved
}

func (n *tickerNativeWindow) setAppbarOrder() {
	order := tickerHWNDTopmost
	if n.fullscreen {
		order = 1 // HWND_BOTTOM keeps exclusive fullscreen windows above the bar.
	}
	setTickerWindowPos.Call(n.hwnd, order, 0, 0, 0, 0, tickerNoMove|tickerNoSize|tickerNoActivate)
}

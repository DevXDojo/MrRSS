//go:build !server

package desktopapi

import (
	"fmt"
	"log"
	"math"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/sys/windows"
)

var (
	tickerUser32       = windows.NewLazySystemDLL("user32.dll")
	setTickerWindowPos = tickerUser32.NewProc("SetWindowPos")
	getTickerStyle     = tickerUser32.NewProc("GetWindowLongW")
	setTickerStyle     = tickerUser32.NewProc("SetWindowLongW")
	registerTickerMsg  = tickerUser32.NewProc("RegisterWindowMessageW")
	postTickerMsg      = tickerUser32.NewProc("PostMessageW")
	getTickerWindow    = tickerUser32.NewProc("GetWindow")
	findTickerWindow   = tickerUser32.NewProc("FindWindowExW")
	monitorFromRect    = tickerUser32.NewProc("MonitorFromRect")
	getMonitorInfo     = tickerUser32.NewProc("GetMonitorInfoW")
	setTickerTimer     = tickerUser32.NewProc("SetTimer")
	killTickerTimer    = tickerUser32.NewProc("KillTimer")
	showTickerWindow   = tickerUser32.NewProc("ShowWindow")
	tickerIsIconic     = tickerUser32.NewProc("IsIconic")
)

const (
	tickerHWNDTop      = 0
	tickerHWNDNotTop   = ^uintptr(1)
	tickerHWNDTopmost  = ^uintptr(0)
	tickerNoMove       = 0x0002
	tickerNoSize       = 0x0001
	tickerNoActivate   = 0x0010
	tickerStyleIndex   = ^uintptr(15) // GWL_STYLE
	tickerExStyleIndex = ^uintptr(19) // GWL_EXSTYLE
	tickerMinimizeBox  = 0x00020000
	tickerExTopmost    = 0x00000008
	tickerExNoActivate = 0x08000000
	tickerTimerID      = 0x4d525353
)

// Accessed on Wails' UI thread, including the native message callback.
type tickerNativeWindow struct {
	window                        *application.WebviewWindow
	hwnd                          uintptr
	mode, screenID                string
	height                        int
	appbarMessage, layoutMessage  uint32
	taskbarCreated                uint32
	registered, fullscreen        bool
	layoutPending, placing        bool
	desktopHost                   uintptr
	requested, reservation        tickerWinRect
	hadMinimizeBox, hadNoActivate bool
}

func nativeTickerPositions() []string { return []string{tickerReserved, tickerDesktop} }

func tickerMessage(name string) uint32 {
	ptr, _ := windows.UTF16PtrFromString(name)
	result, _, _ := registerTickerMsg.Call(uintptr(unsafe.Pointer(ptr)))
	return uint32(result)
}

func (n *tickerNativeWindow) apply(window *application.WebviewWindow, screen *application.Screen, mode string, height int) error {
	hwnd := uintptr(window.NativeWindow())
	if hwnd == 0 {
		return fmt.Errorf("ticker window unavailable")
	}
	if n.layoutMessage == 0 || n.appbarMessage == 0 || n.taskbarCreated == 0 {
		n.layoutMessage = tickerMessage("MrRSS.Ticker.Layout")
		n.appbarMessage = tickerMessage("MrRSS.Ticker.AppBar")
		n.taskbarCreated = tickerMessage("TaskbarCreated")
		if n.layoutMessage == 0 || n.appbarMessage == 0 || n.taskbarCreated == 0 {
			return fmt.Errorf("register ticker window messages")
		}
	}
	n.window, n.hwnd, n.mode, n.screenID, n.height = window, hwnd, mode, screen.ID, height
	style, _, _ := getTickerStyle.Call(hwnd, tickerStyleIndex)
	n.hadMinimizeBox = style&tickerMinimizeBox != 0
	// System panels stay present while Explorer shows the desktop or minimizes
	// ordinary windows. A minimized appbar would leave an empty reservation.
	setTickerStyle.Call(hwnd, tickerStyleIndex, style&^tickerMinimizeBox)
	if mode == tickerDesktop {
		exStyle, _, _ := getTickerStyle.Call(hwnd, tickerExStyleIndex)
		n.hadNoActivate = exStyle&tickerExNoActivate != 0
		setTickerStyle.Call(hwnd, tickerExStyleIndex, exStyle|tickerExNoActivate)
	}
	if err := n.place(screen); err != nil {
		n.release()
		return err
	}
	if mode == tickerDesktop {
		if result, _, err := setTickerTimer.Call(hwnd, tickerTimerID, 200, 0); result == 0 {
			n.release()
			return fmt.Errorf("start desktop ticker timer: %w", err)
		}
	}
	return nil
}

func (n *tickerNativeWindow) release() {
	if n.hwnd == 0 {
		return
	}
	hwnd, mode := n.hwnd, n.mode
	registered := n.registered
	n.window, n.hwnd, n.mode = nil, 0, ""
	n.registered, n.fullscreen, n.layoutPending = false, false, false
	n.desktopHost = 0
	n.requested, n.reservation = tickerWinRect{}, tickerWinRect{}
	if registered {
		data := tickerAppBarData{hwnd: hwnd}
		data.send(appbarRemove)
	}
	if mode == tickerDesktop {
		killTickerTimer.Call(hwnd, tickerTimerID)
	}
	if !windows.IsWindow(windows.HWND(hwnd)) {
		return
	}
	if n.hadMinimizeBox {
		style, _, _ := getTickerStyle.Call(hwnd, tickerStyleIndex)
		setTickerStyle.Call(hwnd, tickerStyleIndex, style|tickerMinimizeBox)
	}
	if mode == tickerDesktop && !n.hadNoActivate {
		exStyle, _, _ := getTickerStyle.Call(hwnd, tickerExStyleIndex)
		setTickerStyle.Call(hwnd, tickerExStyleIndex, exStyle&^tickerExNoActivate)
	}
}

func (n *tickerNativeWindow) windowMessage(hwnd uintptr, message uint32, wParam, lParam uintptr) (uintptr, bool) {
	if hwnd != n.hwnd || n.hwnd == 0 {
		return 0, false
	}
	switch message {
	case n.layoutMessage:
		n.layoutPending = false
		if err := n.place(nil); err != nil {
			log.Printf("Could not reposition ticker: %v", err)
			if n.registered {
				n.registered = false
				n.requested, n.reservation = tickerWinRect{}, tickerWinRect{}
				data := tickerAppBarData{hwnd: hwnd}
				data.send(appbarRemove)
			}
		}
		return 0, true
	case n.appbarMessage:
		if n.mode == tickerReserved {
			switch wParam {
			case appbarStateChange, appbarPosChanged:
				n.queueLayout()
			case appbarFullscreen:
				n.fullscreen = lParam != 0
				n.setAppbarOrder()
			}
		}
		return 0, true
	case n.taskbarCreated:
		// Explorer drops all AppBar registrations when it restarts.
		n.registered = false
		n.desktopHost = 0
		n.requested, n.reservation = tickerWinRect{}, tickerWinRect{}
		n.queueLayout()
	case 0x001a, 0x007e, 0x02e0: // WM_SETTINGCHANGE, WM_DISPLAYCHANGE, WM_DPICHANGED
		n.queueLayout()
	case 0x0021: // WM_MOUSEACTIVATE
		if n.mode == tickerDesktop {
			return 3, true // MA_NOACTIVATE: deliver the click without raising the ticker
		}
	case 0x0006: // WM_ACTIVATE
		if n.registered {
			data := tickerAppBarData{hwnd: hwnd, param: wParam & 0xffff}
			data.send(appbarActivate)
		}
	case 0x0047: // WM_WINDOWPOSCHANGED
		if n.registered && !n.placing {
			data := tickerAppBarData{hwnd: hwnd}
			data.send(appbarWindowPosChanged)
		}
	case 0x0113: // WM_TIMER
		if wParam == tickerTimerID && n.mode == tickerDesktop {
			n.followDesktop()
			return 0, true
		}
	case 0x0082: // WM_NCDESTROY
		n.release()
	}
	return 0, false
}

func (n *tickerNativeWindow) queueLayout() {
	if n.layoutPending {
		return
	}
	posted, _, _ := postTickerMsg.Call(n.hwnd, uintptr(n.layoutMessage), 0, 0)
	n.layoutPending = posted != 0
}

func (n *tickerNativeWindow) place(screen *application.Screen) error {
	if n.placing {
		return nil
	}
	n.placing = true
	defer func() { n.placing = false }()
	if screen == nil {
		screens := application.Get().Screen
		screen = screens.GetByID(n.screenID)
		if screen == nil {
			screen = screens.GetPrimary()
		}
	}
	if screen == nil {
		return fmt.Errorf("ticker display unavailable")
	}
	rect := rectForTickerScreen(screen.PhysicalBounds)
	monitor, _, _ := monitorFromRect.Call(uintptr(unsafe.Pointer(&rect)), 2) // MONITOR_DEFAULTTONEAREST
	info := tickerMonitorInfo{size: uint32(unsafe.Sizeof(tickerMonitorInfo{}))}
	if result, _, err := getMonitorInfo.Call(monitor, uintptr(unsafe.Pointer(&info))); result == 0 {
		return fmt.Errorf("get ticker monitor work area: %w", err)
	}
	height := int32(math.Round(float64(n.height) * max(1, float64(screen.ScaleFactor))))
	if n.mode == tickerReserved {
		return n.placeAppbar(info, height)
	}
	area := info.work
	area.top = max(area.top, area.bottom-height)
	n.setBounds(area)
	n.followDesktop()
	return nil
}

func (n *tickerNativeWindow) setBounds(rect tickerWinRect) {
	bounds := application.Rect{X: int(rect.left), Y: int(rect.top), Width: int(rect.right - rect.left), Height: int(rect.bottom - rect.top)}
	if n.window.PhysicalBounds() != bounds {
		n.window.SetPhysicalBounds(bounds)
	}
}

var desktopViewClass, _ = windows.UTF16PtrFromString("SHELLDLL_DefView")
var desktopWorkerClass, _ = windows.UTF16PtrFromString("WorkerW")

func desktopIconsHost() uintptr {
	shell := windows.GetShellWindow()
	if shell == 0 {
		return 0
	}
	findView := func(parent uintptr) uintptr {
		child, _, _ := findTickerWindow.Call(parent, 0, uintptr(unsafe.Pointer(desktopViewClass)), 0)
		return child
	}
	if findView(uintptr(shell)) != 0 {
		return uintptr(shell)
	}
	// Older Explorer versions host the icons in a separate top-level WorkerW.
	var shellPID uint32
	_, _ = windows.GetWindowThreadProcessId(shell, &shellPID)
	for after := uintptr(0); ; {
		worker, _, _ := findTickerWindow.Call(0, after, uintptr(unsafe.Pointer(desktopWorkerClass)), 0)
		if worker == 0 {
			return uintptr(shell)
		}
		var pid uint32
		_, _ = windows.GetWindowThreadProcessId(windows.HWND(worker), &pid)
		if pid == shellPID && findView(worker) != 0 {
			return worker
		}
		after = worker
	}
}

func (n *tickerNativeWindow) followDesktop() {
	if minimized, _, _ := tickerIsIconic.Call(n.hwnd); minimized != 0 {
		showTickerWindow.Call(n.hwnd, 4) // SW_SHOWNOACTIVATE
	}
	if n.desktopHost == 0 || !windows.IsWindow(windows.HWND(n.desktopHost)) {
		n.desktopHost = desktopIconsHost()
	}
	host := n.desktopHost
	if host == 0 || !windows.IsWindowVisible(windows.HWND(host)) {
		n.desktopHost = 0
		return
	}
	// Win+D raises the desktop above ordinary windows. Track the desktop host's
	// z-order so the ticker moves with it, without reparenting to Explorer.
	above, _, _ := getTickerWindow.Call(host, 3) // GW_HWNDPREV
	hostStyle, _, _ := getTickerStyle.Call(host, tickerExStyleIndex)
	tickerStyle, _, _ := getTickerStyle.Call(n.hwnd, tickerExStyleIndex)
	if above == n.hwnd && (hostStyle^tickerStyle)&tickerExTopmost == 0 {
		return
	}
	flags := uintptr(tickerNoActivate | tickerNoMove | tickerNoSize)
	if hostStyle&tickerExTopmost == 0 {
		if tickerStyle&tickerExTopmost != 0 {
			setTickerWindowPos.Call(n.hwnd, tickerHWNDNotTop, 0, 0, 0, 0, flags)
		}
		if above == n.hwnd {
			above, _, _ = getTickerWindow.Call(host, 3)
		}
		if above == n.hwnd {
			return
		}
		if above != 0 {
			if style, _, _ := getTickerStyle.Call(above, tickerExStyleIndex); style&tickerExTopmost != 0 {
				above = tickerHWNDTop
			}
		}
	} else if above == 0 {
		above = tickerHWNDTopmost
	}
	setTickerWindowPos.Call(n.hwnd, above, 0, 0, 0, 0, flags)
}

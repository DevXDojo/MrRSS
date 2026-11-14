//go:build !server

package desktopapi

import (
	"sync"
	"time"

	"MrRSS/internal/database"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// DesktopCapabilities exposes native window operations and a per-process update
// capability. Article data continues to use the HTTP API.
type DesktopCapabilities struct {
	updateAuthorization string
	app                 *application.App
	db                  *database.DB
	showMainWindow      func()
	mu                  sync.Mutex
	ticker              *application.WebviewWindow
	tickerNative        tickerNativeWindow
	tickerGeometry      tickerGeometry
	tickerConfigured    bool
	tickerBoundsTimer   *time.Timer
	pendingTickerBounds *TickerBounds
	tickerClosed        bool
}

func (s *DesktopCapabilities) UpdateAuthorization() string { return s.updateAuthorization }

func NewDesktopCapabilities(db *database.DB, updateAuthorization string, showMainWindow func()) *DesktopCapabilities {
	return &DesktopCapabilities{db: db, updateAuthorization: updateAuthorization, showMainWindow: showMainWindow}
}

func WindowOptions(s *DesktopCapabilities) application.WindowsOptions {
	return application.WindowsOptions{WndProcInterceptor: s.tickerNative.windowMessage}
}

func Attach(s *DesktopCapabilities, app *application.App) {
	s.app = app
	app.OnShutdown(s.shutdownTicker)
}

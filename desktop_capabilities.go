//go:build !server

package main

import (
	"sync"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/desktopapi"
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
	tickerGeometry      tickerGeometry
	tickerConfigured    bool
	tickerBoundsTimer   *time.Timer
	pendingTickerBounds *desktopapi.TickerBounds
	tickerClosed        bool
}

func (s *DesktopCapabilities) UpdateAuthorization() string { return s.updateAuthorization }

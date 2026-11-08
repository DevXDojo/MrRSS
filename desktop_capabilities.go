//go:build !server

package main

import (
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// DesktopCapabilities is bound only to the native WebView. The loopback API
// and server build cannot mint or retrieve this per-process update capability.
type DesktopCapabilities struct {
	updateAuthorization string
	app                 *application.App
	mu                  sync.Mutex
	ticker              *application.WebviewWindow
}

func (s *DesktopCapabilities) UpdateAuthorization() string { return s.updateAuthorization }

func (s *DesktopCapabilities) ShowTicker() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.app == nil {
		return fmt.Errorf("desktop application unavailable")
	}
	if s.ticker == nil {
		s.ticker = s.app.Window.NewWithOptions(application.WebviewWindowOptions{
			Name: "MrRSS-ticker", Title: "MrRSS", URL: "/?ticker=1",
			Width: 900, Height: 72, MinWidth: 320, MinHeight: 56,
			Frameless: true, AlwaysOnTop: true,
			BackgroundColour: application.NewRGB(30, 41, 59),
		})
		s.ticker.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
			event.Cancel()
			s.ticker.Hide()
		})
		s.ticker.Center()
	}
	s.ticker.Show()
	return nil
}

func (s *DesktopCapabilities) HideTicker() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ticker != nil {
		s.ticker.Hide()
	}
}

func (s *DesktopCapabilities) SetTickerExpanded(expanded bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ticker == nil {
		return
	}
	height := 72
	if expanded {
		height = 200
	}
	s.ticker.SetSize(s.ticker.Width(), height)
}

func (s *DesktopCapabilities) PositionTicker(top bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ticker == nil {
		return fmt.Errorf("ticker is not open")
	}
	screen, err := s.ticker.GetScreen()
	if err != nil || screen == nil {
		screen = s.app.Screen.GetPrimary()
	}
	if screen == nil {
		return fmt.Errorf("screen unavailable")
	}
	area := screen.WorkArea
	height := 72
	if area.Height < height {
		height = area.Height
	}
	s.ticker.SetSize(area.Width, height)
	y := area.Y
	if !top {
		y += area.Height - height
	}
	s.ticker.SetPosition(area.X, y)
	return nil
}

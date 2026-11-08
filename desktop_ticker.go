//go:build !server

package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"MrRSS/internal/config"
	"MrRSS/internal/desktopapi"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

type tickerGeometry struct {
	position, screen      string
	width, fontSize, x, y int
}
type TickerDisplay struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Primary bool   `json:"primary"`
}

func (s *DesktopCapabilities) tickerSetting(key string) string {
	if s.db != nil {
		if value, err := s.db.GetSetting(key); err == nil {
			return value
		}
	}
	return config.GetString(key)
}
func (s *DesktopCapabilities) tickerInteger(key string) int {
	value, err := strconv.Atoi(s.tickerSetting(key))
	if err != nil {
		value, _ = strconv.Atoi(config.GetString(key))
	}
	return value
}

// Native window calls and shutdown hooks share the UI thread. Acquire the
// state lock on that thread so a shutdown cannot block a pending UI call.
func (s *DesktopCapabilities) runTickerUI(action func() error) error {
	if s.app == nil || s.app.Context().Err() != nil {
		return fmt.Errorf("desktop application unavailable")
	}
	var result error
	application.InvokeSync(func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.tickerClosed {
			result = fmt.Errorf("desktop application unavailable")
			return
		}
		result = action()
	})
	return result
}
func (s *DesktopCapabilities) ShowTicker() error { return s.runTickerUI(s.showTickerLocked) }
func (s *DesktopCapabilities) showTickerLocked() error {
	if s.app == nil || s.tickerClosed {
		return fmt.Errorf("desktop application unavailable")
	}
	if s.ticker == nil {
		s.ticker = s.app.Window.NewWithOptions(application.WebviewWindowOptions{
			Name: "MrRSS-ticker", Title: "MrRSS", URL: "/?ticker=1", Width: 960, Height: 44, MinWidth: 360, MinHeight: 44,
			Frameless: true, DisableResize: true, AlwaysOnTop: true, Hidden: true, BackgroundColour: application.NewRGB(15, 23, 42),
			Windows: application.WindowsWindow{HiddenOnTaskbar: true},
		})
		s.ticker.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
			_ = s.runTickerUI(func() error {
				event.Cancel()
				s.ticker.EmitEvent("ticker:request-disable")
				s.hideTickerLocked()
				return nil
			})
		})
		s.ticker.OnWindowEvent(events.Common.WindowDidMove, func(_ *application.WindowEvent) { s.rememberTickerBounds() })
		s.ticker.OnWindowEvent(events.Common.WindowDidResize, func(_ *application.WindowEvent) { s.rememberTickerBounds() })
	}
	if err := s.applyTickerSettingsLocked(); err != nil {
		return err
	}
	if !s.ticker.IsVisible() {
		s.ticker.Show()
		s.ticker.EmitEvent("ticker:visibility", true)
	}
	return nil
}
func (s *DesktopCapabilities) hideTickerLocked() {
	if s.ticker != nil {
		s.ticker.EmitEvent("ticker:visibility", false)
		s.ticker.Hide()
	}
}
func (s *DesktopCapabilities) HideTicker() {
	_ = s.runTickerUI(func() error { s.hideTickerLocked(); return nil })
}

// ShowTickerSettings focuses the main window and opens its nested ticker settings.
func (s *DesktopCapabilities) ShowTickerSettings() error {
	if s.showMainWindow == nil {
		return fmt.Errorf("main window unavailable")
	}
	return s.runTickerUI(func() error {
		s.showMainWindow()
		s.app.Event.Emit("ticker:open-settings")
		return nil
	})
}

// ApplyTickerSettings is called after the HTTP settings request has succeeded.
func (s *DesktopCapabilities) ApplyTickerSettings() error {
	return s.runTickerUI(s.applyTickerPreferencesLocked)
}
func (s *DesktopCapabilities) applyTickerPreferencesLocked() error {
	if s.app == nil || s.tickerClosed {
		return fmt.Errorf("desktop application unavailable")
	}
	if s.tickerSetting("ticker_enabled") == "true" {
		if err := s.showTickerLocked(); err != nil {
			return err
		}
	} else {
		s.hideTickerLocked()
	}
	s.app.Event.Emit("ticker:settings-changed")
	return nil
}
func (s *DesktopCapabilities) requestedTickerGeometry() tickerGeometry {
	geometry := tickerGeometry{position: s.tickerSetting("ticker_position"), screen: s.tickerSetting("ticker_screen_id"), width: s.tickerInteger("ticker_width"), fontSize: s.tickerInteger("ticker_font_size"), x: s.tickerInteger("ticker_window_x"), y: s.tickerInteger("ticker_window_y")}
	// A quick toolbar action must not restore stale database bounds while a
	// drag or resize is still waiting for the debounced write.
	if pending := s.pendingTickerBounds; geometry.position == "floating" && pending != nil {
		geometry.width, geometry.x, geometry.y = pending.Width, pending.X, pending.Y
	}
	return geometry
}
func (s *DesktopCapabilities) applyTickerSettingsLocked() error {
	geometry := s.requestedTickerGeometry()
	s.ticker.SetAlwaysOnTop(s.tickerSetting("ticker_always_on_top") == "true")
	if s.tickerConfigured && geometry == s.tickerGeometry {
		return nil
	}
	screen := s.app.Screen.GetByID(geometry.screen)
	if screen == nil && s.tickerConfigured {
		screen, _ = s.ticker.GetScreen()
	}
	if screen == nil && geometry.position == "floating" && geometry.x != -10000 && geometry.y != -10000 {
		for _, candidate := range s.app.Screen.GetAll() {
			if candidate.WorkArea.Contains(application.Point{X: geometry.x, Y: geometry.y}) {
				screen = candidate
				break
			}
		}
	}
	if screen == nil {
		screen = s.app.Screen.GetPrimary()
	}
	if screen == nil {
		return fmt.Errorf("screen unavailable")
	}
	area := screen.WorkArea
	bounds := desktopapi.CalculateTickerBounds(desktopapi.TickerBounds{X: area.X, Y: area.Y, Width: area.Width, Height: area.Height}, geometry.position, geometry.width, geometry.fontSize, geometry.x, geometry.y)
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return fmt.Errorf("screen has no available work area")
	}
	s.tickerGeometry = geometry
	s.tickerConfigured = true
	s.ticker.SetResizable(geometry.position == "floating")
	s.ticker.SetMaxSize(0, 0)
	s.ticker.SetMinSize(min(360, bounds.Width), bounds.Height)
	maxWidth := bounds.Width
	if geometry.position == "floating" {
		maxWidth = min(2400, area.Width)
	}
	s.ticker.SetMaxSize(maxWidth, bounds.Height)
	s.ticker.SetBounds(application.Rect{X: bounds.X, Y: bounds.Y, Width: bounds.Width, Height: bounds.Height})
	return nil
}
func (s *DesktopCapabilities) GetTickerDisplays() []TickerDisplay {
	if s.app == nil {
		return nil
	}
	var displays []TickerDisplay
	for _, screen := range s.app.Screen.GetAll() {
		displays = append(displays, TickerDisplay{ID: screen.ID, Name: screen.Name, Primary: screen.IsPrimary})
	}
	return displays
}
func (s *DesktopCapabilities) OpenTickerArticle(articleID, feedID int64) error {
	if articleID <= 0 || feedID <= 0 || s.app == nil || s.showMainWindow == nil {
		return fmt.Errorf("reader unavailable")
	}
	return s.runTickerUI(func() error {
		s.showMainWindow()
		s.app.Event.Emit("ticker:open-article", map[string]int64{"article_id": articleID, "feed_id": feedID})
		return nil
	})
}
func (s *DesktopCapabilities) rememberTickerBounds() {
	_ = s.runTickerUI(func() error {
		if s.ticker == nil || s.tickerGeometry.position != "floating" {
			return nil
		}
		x, y := s.ticker.Position()
		width, height := s.ticker.Size()
		if width <= 0 {
			return nil
		}
		s.tickerGeometry.width, s.tickerGeometry.x, s.tickerGeometry.y = width, x, y
		s.pendingTickerBounds = &desktopapi.TickerBounds{X: x, Y: y, Width: width, Height: height}
		if s.tickerBoundsTimer != nil {
			s.tickerBoundsTimer.Stop()
		}
		s.tickerBoundsTimer = time.AfterFunc(300*time.Millisecond, func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			if !s.tickerClosed {
				s.saveTickerBoundsLocked()
			}
		})
		return nil
	})
}
func (s *DesktopCapabilities) saveTickerBoundsLocked() {
	if s.db == nil || s.pendingTickerBounds == nil {
		return
	}
	position := s.pendingTickerBounds
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	tx, err := s.db.BeginTx(ctx, nil)
	if err == nil {
		defer tx.Rollback()
		for key, value := range map[string]int{"ticker_window_x": position.X, "ticker_window_y": position.Y, "ticker_width": position.Width} {
			if _, err = tx.ExecContext(ctx, "INSERT OR REPLACE INTO settings(key,value) VALUES(?,?)", key, strconv.Itoa(value)); err != nil {
				break
			}
		}
		if err == nil {
			err = tx.Commit()
		}
	}
	if err != nil {
		log.Printf("Could not save ticker bounds: %v", err)
		return
	}
	s.pendingTickerBounds = nil
}
func (s *DesktopCapabilities) shutdownTicker() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickerClosed = true
	if s.tickerBoundsTimer != nil {
		s.tickerBoundsTimer.Stop()
	}
	s.saveTickerBoundsLocked()
}

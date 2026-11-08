//go:build !server

package main

import (
	"path/filepath"
	"testing"

	"MrRSS/internal/database"
	"MrRSS/internal/desktopapi"
)

func tickerTestDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.NewDB(filepath.Join(t.TempDir(), "ticker.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestTickerShutdownSavesFloatingPositionAndWidth(t *testing.T) {
	db := tickerTestDB(t)
	caps := &DesktopCapabilities{db: db, pendingTickerBounds: &desktopapi.TickerBounds{X: -1400, Y: 250, Width: 880}}
	caps.shutdownTicker()
	for key, want := range map[string]string{"ticker_window_x": "-1400", "ticker_window_y": "250", "ticker_width": "880"} {
		value, err := db.GetSetting(key)
		if err != nil || value != want {
			t.Errorf("%s = %q, err=%v; want %q", key, value, err, want)
		}
	}
	if caps.pendingTickerBounds != nil || !caps.tickerClosed {
		t.Fatal("shutdown did not flush the pending geometry")
	}
}

func TestTickerPendingResizeWinsOverStoredGeometry(t *testing.T) {
	db := tickerTestDB(t)
	for key, value := range map[string]string{"ticker_position": "floating", "ticker_width": "640", "ticker_window_x": "100", "ticker_window_y": "100"} {
		if err := db.SetSetting(key, value); err != nil {
			t.Fatal(err)
		}
	}
	caps := &DesktopCapabilities{db: db, pendingTickerBounds: &desktopapi.TickerBounds{X: 220, Y: 280, Width: 1040}}
	geometry := caps.requestedTickerGeometry()
	if geometry.width != 1040 || geometry.x != 220 || geometry.y != 280 {
		t.Fatalf("quick preferences would restore stale geometry: %+v", geometry)
	}
	_ = db.Close()
	caps.saveTickerBoundsLocked()
	if caps.pendingTickerBounds == nil {
		t.Fatal("failed persistence discarded the pending resize")
	}
}

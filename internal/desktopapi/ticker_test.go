//go:build !server

package desktopapi

import "testing"

func TestTickerPositionFallsBackForUnsupportedShell(t *testing.T) {
	for _, position := range []string{tickerReserved, tickerDesktop} {
		if got := effectiveTickerPosition(position, []string{"floating", "bottom"}); got != "bottom" {
			t.Errorf("unsupported %s resolved to %s", position, got)
		}
		if got := effectiveTickerPosition(position, []string{"floating", "bottom", position}); got != position {
			t.Errorf("supported %s resolved to %s", position, got)
		}
	}
}

func TestTickerVisibilityOverrideLastsUntilPreferenceChanges(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		visible := !enabled
		capabilities := DesktopCapabilities{tickerVisibleOverride: &visible}
		if got := capabilities.resolveTickerVisibility(enabled); got != visible {
			t.Errorf("runtime override resolved to %t, want %t", got, visible)
		}
		if got := capabilities.resolveTickerVisibility(visible); got != visible {
			t.Errorf("new preference resolved to %t, want %t", got, visible)
		}
		if got := capabilities.resolveTickerVisibility(enabled); got != enabled {
			t.Errorf("previous override still applied: got %t, want %t", got, enabled)
		}
	}
}

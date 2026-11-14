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

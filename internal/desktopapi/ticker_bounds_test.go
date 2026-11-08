package desktopapi

import "testing"

func TestTickerBoundsRespectWorkArea(t *testing.T) {
	area := TickerBounds{X: -1920, Y: -100, Width: 1920, Height: 1040}
	for _, tc := range []struct {
		position string
		y        int
	}{{"top", -100}, {"bottom", 896}, {"unknown", 896}} {
		got := CalculateTickerBounds(area, tc.position, 960, 18, -10000, -10000)
		if got.X != area.X || got.Width != area.Width || got.Y != tc.y || got.Height != 44 {
			t.Errorf("%s: %+v", tc.position, got)
		}
	}
	floating := CalculateTickerBounds(area, "floating", 960, 18, -10000, -10000)
	if floating.X != -1440 || floating.Y != 872 || floating.Width != 960 {
		t.Fatalf("new floating window: %+v", floating)
	}
	restored := CalculateTickerBounds(area, "floating", 960, 18, 5000, -5000)
	if restored.X != -960 || restored.Y != -100 {
		t.Fatalf("offscreen position not restored inside display: %+v", restored)
	}
}

func TestTickerBoundsFitSmallScreensAndLargeFonts(t *testing.T) {
	area := TickerBounds{X: 20, Y: 30, Width: 300, Height: 50}
	for _, position := range []string{"top", "bottom", "floating"} {
		got := CalculateTickerBounds(area, position, 99999, 99, -10000, -10000)
		if got.X < area.X || got.Y < area.Y || got.X+got.Width > area.X+area.Width || got.Y+got.Height > area.Y+area.Height {
			t.Errorf("%s escaped area: %+v", position, got)
		}
	}
	if got := CalculateTickerBounds(TickerBounds{}, "floating", 960, 18, 0, 0); got != (TickerBounds{}) {
		t.Fatalf("invalid screen: %+v", got)
	}
	if got := CalculateTickerBounds(TickerBounds{Width: 1200, Height: 800}, "bottom", 960, 40, 0, 0); got.Height != 60 || got.Y != 740 {
		t.Fatalf("large type clipped: %+v", got)
	}
}

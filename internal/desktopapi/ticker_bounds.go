package desktopapi

// TickerBounds uses logical coordinates, including negative display origins.
type TickerBounds struct{ X, Y, Width, Height int }

// CalculateTickerBounds keeps restored windows inside the available work area.
// The desktop mode uses the same footprint as a bottom-edge ticker. Native
// work-area reservations negotiate their position separately with the shell.
func CalculateTickerBounds(area TickerBounds, position string, width, fontSize, savedX, savedY int) TickerBounds {
	if area.Width <= 0 || area.Height <= 0 {
		return TickerBounds{}
	}
	height := min(area.Height, max(44, min(40, max(12, fontSize))+20))
	bounds := TickerBounds{X: area.X, Y: area.Y, Width: area.Width, Height: height}
	if position == "top" {
		return bounds
	}
	if position == "floating" {
		bounds.Width = min(area.Width, max(360, min(2400, width)))
		if savedX == -10000 || savedY == -10000 {
			bounds.X = area.X + (area.Width-bounds.Width)/2
			bounds.Y = area.Y + max(0, area.Height-height-24)
		} else {
			bounds.X = max(area.X, min(savedX, area.X+area.Width-bounds.Width))
			bounds.Y = max(area.Y, min(savedY, area.Y+area.Height-height))
		}
		return bounds
	}
	bounds.Y += area.Height - height
	return bounds
}

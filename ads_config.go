package main

import (
	_ "embed"

	"MrRSS/internal/advertisement"
	"MrRSS/internal/handlers/core"
)

//go:embed ads.json
var bundledAds []byte

func configureAdvertisements(h *core.Handler) {
	h.Advertisements = advertisement.New(bundledAds)
}

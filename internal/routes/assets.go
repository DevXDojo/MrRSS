package routes

import (
	"net/http"
	"strings"
)

// AssetHandler dispatches application API requests and serves reader assets.
type AssetHandler struct {
	API      http.Handler
	Frontend http.Handler
}

func (h *AssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		h.API.ServeHTTP(w, r)
		return
	}
	h.Frontend.ServeHTTP(w, r)
}

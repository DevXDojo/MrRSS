package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/models"
)

func HandleXPathPreview(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		URL          string `json:"url"`
		ProxyEnabled bool   `json:"proxy_enabled"`
		ProxyURL     string `json:"proxy_url"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&input); err != nil {
		response.Error(w, fmt.Errorf("invalid preview request"), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	preview, err := h.Fetcher.PreviewXPathPage(ctx, &models.Feed{URL: input.URL, ProxyEnabled: input.ProxyEnabled, ProxyURL: input.ProxyURL})
	if err != nil {
		response.Error(w, fmt.Errorf("could not load page preview"), http.StatusBadGateway)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	response.JSON(w, preview)
}

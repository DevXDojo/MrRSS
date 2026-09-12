package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/models"
	"MrRSS/internal/rsshub"
	"MrRSS/internal/utils/urlutil"
)

func HandlePreviewFeed(h *core.Handler, w http.ResponseWriter, r *http.Request) {
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
	input.URL = urlutil.NormalizeFeedURL(strings.TrimSpace(input.URL))
	parsed, err := url.Parse(input.URL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https" && !rsshub.IsRSSHubURL(input.URL)) {
		response.Error(w, fmt.Errorf("preview requires an HTTP, HTTPS, or RSSHub feed URL"), http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	preview, err := h.Fetcher.PreviewSubscription(ctx, &models.Feed{URL: input.URL, ProxyEnabled: input.ProxyEnabled, ProxyURL: input.ProxyURL})
	if err != nil {
		response.Error(w, fmt.Errorf("could not load feed preview"), http.StatusBadGateway)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	response.JSON(w, preview)
}

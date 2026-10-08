package advertisement

import (
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/utils/fileutil"
	"MrRSS/internal/utils/httputil"
)

func HandleAds(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path, err := fileutil.GetDataDir()
	if err != nil {
		http.Error(w, "Advertisement storage unavailable", http.StatusInternalServerError)
		return
	}
	client, _ := httputil.CreateHTTPClientWithProxySettings(h.DB, 8*time.Second)
	if client != nil {
		defer client.CloseIdleConnections()
	}
	ads := h.Advertisements.Get(r.Context(), filepath.Join(path, "ads.json"), client)
	snoozedUntil, _ := h.DB.GetSettingContext(r.Context(), "ads_snoozed_until")
	until, _ := strconv.ParseInt(snoozedUntil, 10, 64)
	w.Header().Set("Cache-Control", "no-store")
	response.JSON(w, map[string]interface{}{"ads": ads, "snoozed_until": until})
}

func HandleSnooze(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	until := time.Now().Add(7 * 24 * time.Hour).Unix()
	_, err := h.DB.ExecContext(r.Context(), "INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)", "ads_snoozed_until", strconv.FormatInt(until, 10))
	if err != nil {
		http.Error(w, "Could not save advertisement preference", http.StatusInternalServerError)
		return
	}
	response.JSON(w, map[string]int64{"snoozed_until": until})
}

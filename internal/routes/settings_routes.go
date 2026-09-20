package routes

import (
	"net/http"

	"MrRSS/internal/handlers/core"
	notificationhandlers "MrRSS/internal/handlers/notification"
	settings "MrRSS/internal/handlers/settings"
	stathandlers "MrRSS/internal/handlers/statistics"
)

// registerSettingsRoutes registers all settings-related routes
func registerSettingsRoutes(mux *http.ServeMux, h *core.Handler) {
	mux.HandleFunc("/api/ad-filter/analyze", func(w http.ResponseWriter, r *http.Request) { settings.HandleAdFilterAnalyze(h, w, r) })
	mux.HandleFunc("/api/ad-filter/config", func(w http.ResponseWriter, r *http.Request) { settings.HandleAdFilterConfig(h, w, r) })
	mux.HandleFunc("/api/ad-filter/preview", func(w http.ResponseWriter, r *http.Request) { settings.HandleAdFilterPreview(h, w, r) })
	mux.HandleFunc("/api/notifications/config", func(w http.ResponseWriter, r *http.Request) { notificationhandlers.HandleConfig(h, w, r) })
	mux.HandleFunc("/api/notifications/test", func(w http.ResponseWriter, r *http.Request) { notificationhandlers.HandleTest(h, w, r) })
	mux.HandleFunc("/api/notifications/preview", func(w http.ResponseWriter, r *http.Request) { notificationhandlers.HandlePreview(h, w, r) })
	mux.HandleFunc("/api/notifications/history", func(w http.ResponseWriter, r *http.Request) { notificationhandlers.HandleHistory(h, w, r) })
	// Settings
	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) { settings.HandleSettings(h, w, r) })

	mux.HandleFunc("/api/settings/data-directory", func(w http.ResponseWriter, r *http.Request) { settings.HandleDataDirectory(h, w, r) })
	mux.HandleFunc("/api/settings/data-directory/select", func(w http.ResponseWriter, r *http.Request) { settings.HandleSelectDataDirectory(h, w, r) })

	// Statistics
	mux.HandleFunc("/api/statistics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			stathandlers.HandleResetStatistics(h, w, r)
		} else {
			stathandlers.HandleGetStatistics(h, w, r)
		}
	})
	mux.HandleFunc("/api/statistics/all-time", func(w http.ResponseWriter, r *http.Request) { stathandlers.HandleGetAllTimeStatistics(h, w, r) })
	mux.HandleFunc("/api/statistics/available-months", func(w http.ResponseWriter, r *http.Request) { stathandlers.HandleGetAvailableMonths(h, w, r) })
}

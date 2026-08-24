package routes

import (
	"net/http"

	"MrRSS/internal/handlers/core"
	settings "MrRSS/internal/handlers/settings"
	stathandlers "MrRSS/internal/handlers/statistics"
)

// registerSettingsRoutes registers all settings-related routes
func registerSettingsRoutes(mux *http.ServeMux, h *core.Handler) {
	// Settings
	mux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) { settings.HandleSettings(h, w, r) })

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

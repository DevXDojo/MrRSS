// Package routes provides centralized route registration for the MrRSS API.
package routes

import (
	"net/http"

	"MrRSS/internal/handlers/core"
	discovery "MrRSS/internal/handlers/discovery"
	feedhandlers "MrRSS/internal/handlers/feed"
	filter_category "MrRSS/internal/handlers/filter_category"
	rsshubHandler "MrRSS/internal/handlers/rsshub"
	taghandlers "MrRSS/internal/handlers/tags"
)

// registerFeedRoutes registers all feed-related routes
func registerFeedRoutes(mux *http.ServeMux, h *core.Handler) {
	mux.HandleFunc("/api/feeds", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleFeeds(h, w, r) })
	mux.HandleFunc("/api/feeds/add", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleAddFeed(h, w, r) })
	mux.HandleFunc("/api/feeds/delete", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleDeleteFeed(h, w, r) })
	mux.HandleFunc("/api/feeds/update", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleUpdateFeed(h, w, r) })
	mux.HandleFunc("/api/feeds/refresh", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleRefreshFeed(h, w, r) })
	mux.HandleFunc("/api/feeds/reorder", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleReorderFeed(h, w, r) })
	mux.HandleFunc("/api/feeds/test-imap", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleTestIMAPConnection(h, w, r) })

	// Discovery routes
	mux.HandleFunc("/api/feeds/discover", func(w http.ResponseWriter, r *http.Request) { discovery.HandleDiscoverBlogs(h, w, r) })
	mux.HandleFunc("/api/feeds/discover-all", func(w http.ResponseWriter, r *http.Request) { discovery.HandleDiscoverAllFeeds(h, w, r) })
	mux.HandleFunc("/api/feeds/discover/start", func(w http.ResponseWriter, r *http.Request) { discovery.HandleStartSingleDiscovery(h, w, r) })
	mux.HandleFunc("/api/feeds/discover/progress", func(w http.ResponseWriter, r *http.Request) { discovery.HandleGetSingleDiscoveryProgress(h, w, r) })
	mux.HandleFunc("/api/feeds/discover/clear", func(w http.ResponseWriter, r *http.Request) { discovery.HandleClearSingleDiscovery(h, w, r) })
	mux.HandleFunc("/api/feeds/discover-all/start", func(w http.ResponseWriter, r *http.Request) { discovery.HandleStartBatchDiscovery(h, w, r) })
	mux.HandleFunc("/api/feeds/discover-all/progress", func(w http.ResponseWriter, r *http.Request) { discovery.HandleGetBatchDiscoveryProgress(h, w, r) })
	mux.HandleFunc("/api/feeds/discover-all/clear", func(w http.ResponseWriter, r *http.Request) { discovery.HandleClearBatchDiscovery(h, w, r) })

	// Tag routes
	mux.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) { taghandlers.HandleTags(h, w, r) })
	mux.HandleFunc("/api/tags/update", func(w http.ResponseWriter, r *http.Request) { taghandlers.HandleTagUpdate(h, w, r) })
	mux.HandleFunc("/api/tags/delete", func(w http.ResponseWriter, r *http.Request) { taghandlers.HandleTagDelete(h, w, r) })
	mux.HandleFunc("/api/tags/reorder", func(w http.ResponseWriter, r *http.Request) { taghandlers.HandleTagReorder(h, w, r) })

	// Saved filters routes
	mux.HandleFunc("/api/saved-filters", func(w http.ResponseWriter, r *http.Request) {
		filter_category.HandleSavedFilters(h, w, r)
	})
	mux.HandleFunc("/api/saved-filters/reorder", func(w http.ResponseWriter, r *http.Request) {
		filter_category.HandleReorderSavedFilters(h, w, r)
	})
	mux.HandleFunc("/api/saved-filters/filter", func(w http.ResponseWriter, r *http.Request) {
		filter_category.HandleSavedFilter(h, w, r)
	})

	// RSSHub routes
	mux.HandleFunc("/api/rsshub/add", func(w http.ResponseWriter, r *http.Request) { rsshubHandler.HandleAddFeed(h, w, r) })
	mux.HandleFunc("/api/rsshub/test-connection", func(w http.ResponseWriter, r *http.Request) { rsshubHandler.HandleTestConnection(h, w, r) })
	mux.HandleFunc("/api/rsshub/validate-route", func(w http.ResponseWriter, r *http.Request) { rsshubHandler.HandleValidateRoute(h, w, r) })
	mux.HandleFunc("/api/rsshub/transform-url", func(w http.ResponseWriter, r *http.Request) { rsshubHandler.HandleTransformURL(h, w, r) })
}

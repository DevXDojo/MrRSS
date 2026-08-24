package routes

import (
	"MrRSS/internal/handlers/article"
	browser "MrRSS/internal/handlers/browser"
	"MrRSS/internal/handlers/core"
	customcss "MrRSS/internal/handlers/custom_css"
	freshrssHandler "MrRSS/internal/handlers/freshrss"
	media "MrRSS/internal/handlers/media"
	networkhandlers "MrRSS/internal/handlers/network"
	opml "MrRSS/internal/handlers/opml"
	rules "MrRSS/internal/handlers/rules"
	script "MrRSS/internal/handlers/script"
	update "MrRSS/internal/handlers/update"
	window "MrRSS/internal/handlers/window"
	"net/http"
)

// registerOtherRoutes registers all other miscellaneous routes
func registerOtherRoutes(mux *http.ServeMux, h *core.Handler) {
	// Refresh and progress
	mux.HandleFunc("/api/refresh", func(w http.ResponseWriter, r *http.Request) { article.HandleRefresh(h, w, r) })
	mux.HandleFunc("/api/progress", func(w http.ResponseWriter, r *http.Request) { article.HandleProgress(h, w, r) })
	mux.HandleFunc("/api/progress/task-details", func(w http.ResponseWriter, r *http.Request) { article.HandleTaskDetails(h, w, r) })

	// OPML
	mux.HandleFunc("/api/opml/import", func(w http.ResponseWriter, r *http.Request) { opml.HandleOPMLImport(h, w, r) })
	mux.HandleFunc("/api/opml/export", func(w http.ResponseWriter, r *http.Request) { opml.HandleOPMLExport(h, w, r) })
	mux.HandleFunc("/api/opml/import-dialog", func(w http.ResponseWriter, r *http.Request) { opml.HandleOPMLImportDialog(h, w, r) })
	mux.HandleFunc("/api/opml/export-dialog", func(w http.ResponseWriter, r *http.Request) { opml.HandleOPMLExportDialog(h, w, r) })

	// Update
	mux.HandleFunc("/api/check-updates", func(w http.ResponseWriter, r *http.Request) { update.HandleCheckUpdates(h, w, r) })
	mux.HandleFunc("/api/download-update", func(w http.ResponseWriter, r *http.Request) { update.HandleDownloadUpdate(h, w, r) })
	mux.HandleFunc("/api/install-update", func(w http.ResponseWriter, r *http.Request) { update.HandleInstallUpdate(h, w, r) })
	mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) { update.HandleVersion(h, w, r) })

	// Rules
	mux.HandleFunc("/api/rules/apply", func(w http.ResponseWriter, r *http.Request) { rules.HandleApplyRule(h, w, r) })

	// Scripts
	mux.HandleFunc("/api/scripts/dir", func(w http.ResponseWriter, r *http.Request) { script.HandleGetScriptsDir(h, w, r) })
	mux.HandleFunc("/api/scripts/open", func(w http.ResponseWriter, r *http.Request) { script.HandleOpenScriptsDir(h, w, r) })
	mux.HandleFunc("/api/scripts/list", func(w http.ResponseWriter, r *http.Request) { script.HandleListScripts(h, w, r) })

	// Media
	mux.HandleFunc("/api/media/proxy", func(w http.ResponseWriter, r *http.Request) { media.HandleMediaProxy(h, w, r) })
	mux.HandleFunc("/api/media/cleanup", func(w http.ResponseWriter, r *http.Request) { media.HandleMediaCacheCleanup(h, w, r) })
	mux.HandleFunc("/api/media/info", func(w http.ResponseWriter, r *http.Request) { media.HandleMediaCacheInfo(h, w, r) })
	mux.HandleFunc("/api/webpage/proxy", func(w http.ResponseWriter, r *http.Request) { media.HandleWebpageProxy(h, w, r) })
	mux.HandleFunc("/api/webpage/resource", func(w http.ResponseWriter, r *http.Request) { media.HandleWebpageResource(h, w, r) })

	// Network
	mux.HandleFunc("/api/network/detect", func(w http.ResponseWriter, r *http.Request) { networkhandlers.HandleDetectNetwork(h, w, r) })
	mux.HandleFunc("/api/network/info", func(w http.ResponseWriter, r *http.Request) { networkhandlers.HandleGetNetworkInfo(h, w, r) })

	// Browser
	mux.HandleFunc("/api/browser/open", func(w http.ResponseWriter, r *http.Request) { browser.HandleOpenURL(h, w, r) })

	// Custom CSS
	mux.HandleFunc("/api/custom-css/upload-dialog", func(w http.ResponseWriter, r *http.Request) { customcss.HandleUploadCSSDialog(h, w, r) })
	mux.HandleFunc("/api/custom-css/upload", func(w http.ResponseWriter, r *http.Request) { customcss.HandleUploadCSS(h, w, r) })
	mux.HandleFunc("/api/custom-css", func(w http.ResponseWriter, r *http.Request) { customcss.HandleGetCSS(h, w, r) })
	mux.HandleFunc("/api/custom-css/delete", func(w http.ResponseWriter, r *http.Request) { customcss.HandleDeleteCSS(h, w, r) })

	// Window
	mux.HandleFunc("/api/window/state", func(w http.ResponseWriter, r *http.Request) { window.HandleGetWindowState(h, w, r) })
	mux.HandleFunc("/api/window/save", func(w http.ResponseWriter, r *http.Request) { window.HandleSaveWindowState(h, w, r) })

	// FreshRSS
	mux.HandleFunc("/api/freshrss/sync", func(w http.ResponseWriter, r *http.Request) { freshrssHandler.HandleSync(h, w, r) })
	mux.HandleFunc("/api/freshrss/sync-feed", func(w http.ResponseWriter, r *http.Request) { freshrssHandler.HandleSyncFeed(h, w, r) })
	mux.HandleFunc("/api/freshrss/status", func(w http.ResponseWriter, r *http.Request) { freshrssHandler.HandleSyncStatus(h, w, r) })
}

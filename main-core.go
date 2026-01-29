//go:build server

package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/feed"
	aihandlers "MrRSS/internal/handlers/ai"
	article "MrRSS/internal/handlers/article"
	browser "MrRSS/internal/handlers/browser"
	chat "MrRSS/internal/handlers/chat"
	handlers "MrRSS/internal/handlers/core"
	customcss "MrRSS/internal/handlers/custom_css"
	discovery "MrRSS/internal/handlers/discovery"
	feedhandlers "MrRSS/internal/handlers/feed"
	filter_category "MrRSS/internal/handlers/filter_category"
	freshrssHandler "MrRSS/internal/handlers/freshrss"
	media "MrRSS/internal/handlers/media"
	networkhandlers "MrRSS/internal/handlers/network"
	opml "MrRSS/internal/handlers/opml"
	rsshubHandler "MrRSS/internal/handlers/rsshub"
	rules "MrRSS/internal/handlers/rules"
	script "MrRSS/internal/handlers/script"
	settings "MrRSS/internal/handlers/settings"
	stathandlers "MrRSS/internal/handlers/statistics"
	summary "MrRSS/internal/handlers/summary"
	taghandlers "MrRSS/internal/handlers/tags"
	translationhandlers "MrRSS/internal/handlers/translation"
	update "MrRSS/internal/handlers/update"
	window "MrRSS/internal/handlers/window"
	"MrRSS/internal/network"
	"MrRSS/internal/translation"
	"MrRSS/internal/utils"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           MrRSS API
// @version         1.3.19
// @description     MrRSS is a modern, cross-platform desktop RSS reader with auto-translation, smart feed discovery, and AI-powered summarization.

// @contact.name   API Support
// @contact.url    https://github.com/WCY-dt/MrRSS
// @contact.email  mail@ch3nyang.top

// @license.name  GPL-3.0
// @license.url   https://www.gnu.org/licenses/gpl-3.0.en.html

// @host      localhost:1234
// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

var debugLogging = os.Getenv("MRRSS_DEBUG") != ""

func debugLog(format string, args ...interface{}) {
	if debugLogging {
		log.Printf(format, args...)
	}
}

//go:embed frontend/dist
var frontendFiles embed.FS

type CombinedHandler struct {
	apiMux     *http.ServeMux
	fileServer http.Handler
}

func (h *CombinedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		h.apiMux.ServeHTTP(w, r)
		return
	}
	h.fileServer.ServeHTTP(w, r)
}

func main() {
	// Parse flags
	flag.BoolFunc("server", "Run in headless server mode", func(s string) error {
		v, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		utils.SetServerMode(v)
		return nil
	})
	host := flag.String("host", "0.0.0.0", "Host to listen on in server mode")
	port := flag.String("port", "1234", "Port to listen on in server mode")
	flag.Parse()

	// Force server mode for this build
	utils.SetServerMode(true)

	// Get proper paths for data files
	logPath, err := utils.GetLogPath()
	if err != nil {
		log.Printf("Warning: Could not get log path: %v. Using current directory.", err)
		logPath = "debug.log"
	}

	// In server mode, log to both stdout and file
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.SetOutput(os.Stdout) // Fallback
	} else {
		// Note: we don't close f here as it needs to stay open for logging
		// It will be closed by OS on process exit
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}

	log.Println("Starting application in server mode...")

	// Log portable mode status
	if utils.IsPortableMode() {
		log.Println("Running in PORTABLE mode")
	} else {
		log.Println("Running in NORMAL mode")
	}

	log.Printf("Log file: %s", logPath)

	// Get database path
	dbPath, err := utils.GetDBPath()
	if err != nil {
		log.Printf("Error getting database path: %v", err)
		log.Fatal(err)
	}
	debugLog("Database path: %s", dbPath)

	// Initialize database
	log.Println("Initializing Database...")
	db, err := database.NewDB(dbPath)
	if err != nil {
		log.Printf("Error initializing database: %v", err)
		log.Fatal(err)
	}

	// Run database schema initialization synchronously to ensure it's ready
	log.Println("Running DB migrations...")
	if err := db.Init(); err != nil {
		log.Printf("Error initializing database schema: %v", err)
		log.Fatal(err)
	}
	log.Println("Database initialized successfully")

	translator := translation.NewDynamicTranslatorWithCache(db, db)
	fetcher := feed.NewFetcher(db)
	h := handlers.NewHandler(db, fetcher, translator)

	// API Routes
	log.Println("Setting up API routes...")
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/feeds", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleFeeds(h, w, r) })
	apiMux.HandleFunc("/api/feeds/add", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleAddFeed(h, w, r) })
	apiMux.HandleFunc("/api/feeds/delete", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleDeleteFeed(h, w, r) })
	apiMux.HandleFunc("/api/feeds/update", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleUpdateFeed(h, w, r) })
	apiMux.HandleFunc("/api/feeds/refresh", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleRefreshFeed(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover", func(w http.ResponseWriter, r *http.Request) { discovery.HandleDiscoverBlogs(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover-all", func(w http.ResponseWriter, r *http.Request) { discovery.HandleDiscoverAllFeeds(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover/start", func(w http.ResponseWriter, r *http.Request) { discovery.HandleStartSingleDiscovery(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover/progress", func(w http.ResponseWriter, r *http.Request) { discovery.HandleGetSingleDiscoveryProgress(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover/clear", func(w http.ResponseWriter, r *http.Request) { discovery.HandleClearSingleDiscovery(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover-all/start", func(w http.ResponseWriter, r *http.Request) { discovery.HandleStartBatchDiscovery(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover-all/progress", func(w http.ResponseWriter, r *http.Request) { discovery.HandleGetBatchDiscoveryProgress(h, w, r) })
	apiMux.HandleFunc("/api/feeds/discover-all/clear", func(w http.ResponseWriter, r *http.Request) { discovery.HandleClearBatchDiscovery(h, w, r) })
	apiMux.HandleFunc("/api/feeds/reorder", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleReorderFeed(h, w, r) })
	apiMux.HandleFunc("/api/feeds/test-imap", func(w http.ResponseWriter, r *http.Request) { feedhandlers.HandleTestIMAPConnection(h, w, r) })
	// Tag routes
	apiMux.HandleFunc("/api/tags", func(w http.ResponseWriter, r *http.Request) { taghandlers.HandleTags(h, w, r) })
	apiMux.HandleFunc("/api/tags/update", func(w http.ResponseWriter, r *http.Request) { taghandlers.HandleTagUpdate(h, w, r) })
	apiMux.HandleFunc("/api/tags/delete", func(w http.ResponseWriter, r *http.Request) { taghandlers.HandleTagDelete(h, w, r) })
	apiMux.HandleFunc("/api/tags/reorder", func(w http.ResponseWriter, r *http.Request) { taghandlers.HandleTagReorder(h, w, r) })
	// Saved filters routes
	apiMux.HandleFunc("/api/saved-filters", func(w http.ResponseWriter, r *http.Request) {
		filter_category.HandleSavedFilters(h, w, r)
	})
	apiMux.HandleFunc("/api/saved-filters/reorder", func(w http.ResponseWriter, r *http.Request) {
		filter_category.HandleReorderSavedFilters(h, w, r)
	})
	apiMux.HandleFunc("/api/saved-filters/filter", func(w http.ResponseWriter, r *http.Request) {
		filter_category.HandleSavedFilter(h, w, r)
	})
	apiMux.HandleFunc("/api/articles", func(w http.ResponseWriter, r *http.Request) { article.HandleArticles(h, w, r) })
	apiMux.HandleFunc("/api/articles/images", func(w http.ResponseWriter, r *http.Request) { article.HandleImageGalleryArticles(h, w, r) })
	apiMux.HandleFunc("/api/articles/filter", func(w http.ResponseWriter, r *http.Request) { article.HandleFilteredArticles(h, w, r) })
	apiMux.HandleFunc("/api/articles/read", func(w http.ResponseWriter, r *http.Request) { article.HandleMarkReadWithImmediateSync(h, w, r) })
	apiMux.HandleFunc("/api/articles/favorite", func(w http.ResponseWriter, r *http.Request) { article.HandleToggleFavoriteWithImmediateSync(h, w, r) })
	apiMux.HandleFunc("/api/articles/mark-relative", func(w http.ResponseWriter, r *http.Request) { article.HandleMarkRelativeToArticle(h, w, r) })
	apiMux.HandleFunc("/api/articles/cleanup", func(w http.ResponseWriter, r *http.Request) { article.HandleCleanupArticles(h, w, r) })
	apiMux.HandleFunc("/api/articles/cleanup-content", func(w http.ResponseWriter, r *http.Request) { article.HandleCleanupArticleContent(h, w, r) })
	apiMux.HandleFunc("/api/articles/content-cache-info", func(w http.ResponseWriter, r *http.Request) { article.HandleGetArticleContentCacheInfo(h, w, r) })
	apiMux.HandleFunc("/api/articles/translate", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleTranslateArticle(h, w, r) })
	apiMux.HandleFunc("/api/articles/translate-text", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleTranslateText(h, w, r) })
	apiMux.HandleFunc("/api/articles/clear-translations", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleClearTranslations(h, w, r) })
	apiMux.HandleFunc("/api/ai-usage", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleGetAIUsage(h, w, r) })
	apiMux.HandleFunc("/api/ai-usage/reset", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleResetAIUsage(h, w, r) })
	apiMux.HandleFunc("/api/translation/test-custom", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleTestCustomTranslation(h, w, r) })
	apiMux.HandleFunc("/api/ai-chat", func(w http.ResponseWriter, r *http.Request) { chat.HandleAIChat(h, w, r) })
	apiMux.HandleFunc("/api/ai/chat/sessions/delete-all", func(w http.ResponseWriter, r *http.Request) { chat.HandleDeleteAllSessions(h, w, r) })
	apiMux.HandleFunc("/api/ai/chat/sessions", func(w http.ResponseWriter, r *http.Request) { chat.HandleListSessions(h, w, r) })
	apiMux.HandleFunc("/api/ai/chat/session/create", func(w http.ResponseWriter, r *http.Request) { chat.HandleCreateSession(h, w, r) })
	apiMux.HandleFunc("/api/ai/chat/session", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			chat.HandleGetSession(h, w, r)
		case http.MethodPut, http.MethodPatch:
			chat.HandleUpdateSession(h, w, r)
		case http.MethodDelete:
			chat.HandleDeleteSession(h, w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	apiMux.HandleFunc("/api/ai/chat/messages", func(w http.ResponseWriter, r *http.Request) { chat.HandleListMessages(h, w, r) })
	apiMux.HandleFunc("/api/ai/chat/message/delete", func(w http.ResponseWriter, r *http.Request) { chat.HandleDeleteMessage(h, w, r) })
	apiMux.HandleFunc("/api/ai/test", func(w http.ResponseWriter, r *http.Request) { aihandlers.HandleTestAIConfig(h, w, r) })
	apiMux.HandleFunc("/api/ai/test/info", func(w http.ResponseWriter, r *http.Request) { aihandlers.HandleGetAITestInfo(h, w, r) })
	apiMux.HandleFunc("/api/ai/search", func(w http.ResponseWriter, r *http.Request) { aihandlers.HandleAISearch(h, w, r) })
	apiMux.HandleFunc("/api/articles/toggle-hide", func(w http.ResponseWriter, r *http.Request) { article.HandleToggleHideArticle(h, w, r) })
	apiMux.HandleFunc("/api/articles/toggle-read-later", func(w http.ResponseWriter, r *http.Request) { article.HandleToggleReadLater(h, w, r) })
	apiMux.HandleFunc("/api/articles/content", func(w http.ResponseWriter, r *http.Request) { article.HandleGetArticleContent(h, w, r) })
	apiMux.HandleFunc("/api/articles/fetch-full", func(w http.ResponseWriter, r *http.Request) { article.HandleFetchFullArticle(h, w, r) })
	apiMux.HandleFunc("/api/articles/extract-images", func(w http.ResponseWriter, r *http.Request) { article.HandleExtractAllImages(h, w, r) })
	apiMux.HandleFunc("/api/articles/unread-counts", func(w http.ResponseWriter, r *http.Request) { article.HandleGetUnreadCounts(h, w, r) })
	apiMux.HandleFunc("/api/articles/filter-counts", func(w http.ResponseWriter, r *http.Request) { article.HandleGetFilterCounts(h, w, r) })
	apiMux.HandleFunc("/api/articles/mark-all-read", func(w http.ResponseWriter, r *http.Request) { article.HandleMarkAllAsRead(h, w, r) })
	apiMux.HandleFunc("/api/articles/clear-read-later", func(w http.ResponseWriter, r *http.Request) { article.HandleClearReadLater(h, w, r) })
	apiMux.HandleFunc("/api/articles/summarize", func(w http.ResponseWriter, r *http.Request) { summary.HandleSummarizeArticle(h, w, r) })
	apiMux.HandleFunc("/api/articles/clear-summaries", func(w http.ResponseWriter, r *http.Request) { summary.HandleClearSummaries(h, w, r) })
	apiMux.HandleFunc("/api/articles/export/obsidian", func(w http.ResponseWriter, r *http.Request) { article.HandleExportToObsidian(h, w, r) })
	apiMux.HandleFunc("/api/settings", func(w http.ResponseWriter, r *http.Request) { settings.HandleSettings(h, w, r) })
	apiMux.HandleFunc("/api/refresh", func(w http.ResponseWriter, r *http.Request) { article.HandleRefresh(h, w, r) })
	apiMux.HandleFunc("/api/progress", func(w http.ResponseWriter, r *http.Request) { article.HandleProgress(h, w, r) })
	apiMux.HandleFunc("/api/progress/task-details", func(w http.ResponseWriter, r *http.Request) { article.HandleTaskDetails(h, w, r) })
	apiMux.HandleFunc("/api/opml/import", func(w http.ResponseWriter, r *http.Request) { opml.HandleOPMLImport(h, w, r) })
	apiMux.HandleFunc("/api/opml/export", func(w http.ResponseWriter, r *http.Request) { opml.HandleOPMLExport(h, w, r) })
	apiMux.HandleFunc("/api/opml/import-dialog", func(w http.ResponseWriter, r *http.Request) { opml.HandleOPMLImportDialog(h, w, r) })
	apiMux.HandleFunc("/api/opml/export-dialog", func(w http.ResponseWriter, r *http.Request) { opml.HandleOPMLExportDialog(h, w, r) })
	apiMux.HandleFunc("/api/check-updates", func(w http.ResponseWriter, r *http.Request) { update.HandleCheckUpdates(h, w, r) })
	apiMux.HandleFunc("/api/download-update", func(w http.ResponseWriter, r *http.Request) { update.HandleDownloadUpdate(h, w, r) })
	apiMux.HandleFunc("/api/install-update", func(w http.ResponseWriter, r *http.Request) { update.HandleInstallUpdate(h, w, r) })
	apiMux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) { update.HandleVersion(h, w, r) })
	apiMux.HandleFunc("/api/rules/apply", func(w http.ResponseWriter, r *http.Request) { rules.HandleApplyRule(h, w, r) })
	apiMux.HandleFunc("/api/scripts/dir", func(w http.ResponseWriter, r *http.Request) { script.HandleGetScriptsDir(h, w, r) })
	apiMux.HandleFunc("/api/scripts/open", func(w http.ResponseWriter, r *http.Request) { script.HandleOpenScriptsDir(h, w, r) })
	apiMux.HandleFunc("/api/scripts/list", func(w http.ResponseWriter, r *http.Request) { script.HandleListScripts(h, w, r) })
	apiMux.HandleFunc("/api/media/proxy", func(w http.ResponseWriter, r *http.Request) { media.HandleMediaProxy(h, w, r) })
	apiMux.HandleFunc("/api/media/cleanup", func(w http.ResponseWriter, r *http.Request) { media.HandleMediaCacheCleanup(h, w, r) })
	apiMux.HandleFunc("/api/media/info", func(w http.ResponseWriter, r *http.Request) { media.HandleMediaCacheInfo(h, w, r) })
	apiMux.HandleFunc("/api/webpage/proxy", func(w http.ResponseWriter, r *http.Request) { media.HandleWebpageProxy(h, w, r) })
	apiMux.HandleFunc("/api/webpage/resource", func(w http.ResponseWriter, r *http.Request) { media.HandleWebpageResource(h, w, r) })
	apiMux.HandleFunc("/api/window/state", func(w http.ResponseWriter, r *http.Request) { window.HandleGetWindowState(h, w, r) })
	apiMux.HandleFunc("/api/window/save", func(w http.ResponseWriter, r *http.Request) { window.HandleSaveWindowState(h, w, r) })
	apiMux.HandleFunc("/api/network/detect", func(w http.ResponseWriter, r *http.Request) { networkhandlers.HandleDetectNetwork(h, w, r) })
	apiMux.HandleFunc("/api/network/info", func(w http.ResponseWriter, r *http.Request) { networkhandlers.HandleGetNetworkInfo(h, w, r) })
	apiMux.HandleFunc("/api/browser/open", func(w http.ResponseWriter, r *http.Request) { browser.HandleOpenURL(h, w, r) })
	apiMux.HandleFunc("/api/custom-css/upload-dialog", func(w http.ResponseWriter, r *http.Request) { customcss.HandleUploadCSSDialog(h, w, r) })
	apiMux.HandleFunc("/api/custom-css/upload", func(w http.ResponseWriter, r *http.Request) { customcss.HandleUploadCSS(h, w, r) })
	apiMux.HandleFunc("/api/custom-css", func(w http.ResponseWriter, r *http.Request) { customcss.HandleGetCSS(h, w, r) })
	apiMux.HandleFunc("/api/custom-css/delete", func(w http.ResponseWriter, r *http.Request) { customcss.HandleDeleteCSS(h, w, r) })
	apiMux.HandleFunc("/api/freshrss/sync", func(w http.ResponseWriter, r *http.Request) { freshrssHandler.HandleSync(h, w, r) })
	apiMux.HandleFunc("/api/freshrss/sync-feed", func(w http.ResponseWriter, r *http.Request) { freshrssHandler.HandleSyncFeed(h, w, r) })
	apiMux.HandleFunc("/api/freshrss/status", func(w http.ResponseWriter, r *http.Request) { freshrssHandler.HandleSyncStatus(h, w, r) })
	// RSSHub routes
	apiMux.HandleFunc("/api/rsshub/add", func(w http.ResponseWriter, r *http.Request) { rsshubHandler.HandleAddFeed(h, w, r) })
	apiMux.HandleFunc("/api/rsshub/test-connection", func(w http.ResponseWriter, r *http.Request) { rsshubHandler.HandleTestConnection(h, w, r) })
	apiMux.HandleFunc("/api/rsshub/validate-route", func(w http.ResponseWriter, r *http.Request) { rsshubHandler.HandleValidateRoute(h, w, r) })
	apiMux.HandleFunc("/api/rsshub/transform-url", func(w http.ResponseWriter, r *http.Request) { rsshubHandler.HandleTransformURL(h, w, r) })
	// Statistics routes
	apiMux.HandleFunc("/api/statistics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			stathandlers.HandleResetStatistics(h, w, r)
		} else {
			stathandlers.HandleGetStatistics(h, w, r)
		}
	})
	apiMux.HandleFunc("/api/statistics/all-time", func(w http.ResponseWriter, r *http.Request) { stathandlers.HandleGetAllTimeStatistics(h, w, r) })
	apiMux.HandleFunc("/api/statistics/available-months", func(w http.ResponseWriter, r *http.Request) { stathandlers.HandleGetAvailableMonths(h, w, r) })

	// Swagger Documentation - Serve swagger.json file
	apiMux.HandleFunc("/docs/SERVER_MODE/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/SERVER_MODE/swagger.json")
	})

	apiMux.HandleFunc("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/docs/SERVER_MODE/swagger.json"),
	))

	// Static Files
	log.Println("Setting up static files...")
	frontendFS, err := fs.Sub(frontendFiles, "frontend/dist")
	if err != nil {
		log.Fatal(err)
	}

	fileServer := http.FileServer(http.FS(frontendFS))

	combinedHandler := &CombinedHandler{
		apiMux:     apiMux,
		fileServer: fileServer,
	}

	log.Printf("Starting in headless server mode on http://%s:%s", *host, *port)

	// Start background scheduler
	// Use a context that we can cancel on shutdown
	bgCtx, bgCancel := context.WithCancel(context.Background())

	log.Println("Starting background scheduler...")
	go h.StartBackgroundScheduler(bgCtx)

	// Start Network Speed Detection (optional but good to have)
	go func() {
		log.Println("Detecting network speed...")
		detector := network.NewDetector(&http.Client{Timeout: 10 * time.Second})
		detectCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		result := detector.DetectSpeed(detectCtx)
		if result.DetectionSuccess {
			db.SetSetting("network_speed", string(result.SpeedLevel))
			db.SetSetting("network_bandwidth_mbps", fmt.Sprintf("%.2f", result.BandwidthMbps))
			log.Printf("Network detection complete: %s", result.SpeedLevel)
		}
	}()

	// Start HTTP Server
	srv := &http.Server{
		Addr:    *host + ":" + *port,
		Handler: combinedHandler,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for interrupt
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	bgCancel()

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close Database
	if err := db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	} else {
		log.Println("Database closed")
	}

	log.Println("Server exited")
}

//go:build !server

package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

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
	translationhandlers "MrRSS/internal/handlers/translation"
	update "MrRSS/internal/handlers/update"
	window "MrRSS/internal/handlers/window"
	"MrRSS/internal/network"
	"MrRSS/internal/translation"
	"MrRSS/internal/utils"
)

var debugLogging = os.Getenv("MRRSS_DEBUG") != ""

func debugLog(format string, args ...interface{}) {
	if debugLogging {
		log.Printf(format, args...)
	}
}

//go:embed frontend/dist
var frontendFiles embed.FS

// Platform-specific icon embedding
// Windows and macOS both use PNG format for system tray
// Windows .ico is only used for executable icon (via syso)
//
//go:embed build/windows/icon.png
var appIconWindows []byte

//go:embed build/appicon.png
var appIconMacOS []byte

// getAppIcon returns the appropriate icon for the current platform
func getAppIcon() []byte {
	if runtime.GOOS == "windows" {
		return appIconWindows
	}
	return appIconMacOS
}

type windowState struct {
	width  int
	height int
	x      int
	y      int
	valid  atomic.Bool
}

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

// APIMiddleware routes API requests to the API handler, and lets Wails handle the rest
func APIMiddleware(combinedHandler *CombinedHandler) application.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Let the /wails route be handled by Wails runtime
			if strings.HasPrefix(r.URL.Path, "/wails") {
				next.ServeHTTP(w, r)
				return
			}
			// Handle API routes and serve static files
			combinedHandler.ServeHTTP(w, r)
		})
	}
}

func setupApplicationMenu(app *application.App, db *database.DB, h *handlers.Handler, bgCtx context.Context) {
	// Only for macOS
	if runtime.GOOS != "darwin" {
		return
	}

	menu := app.NewMenu()
	menu.AddRole(application.AppMenu) // About, Hide, Quit, etc

	// Get language for labels
	lang := "en"
	if l, err := db.GetSetting("language"); err == nil && l != "" {
		lang = l
	}

	var fileLabel, refreshLabel, addFeedLabel, settingsLabel string
	switch lang {
	case "zh-CN", "zh", "zh-cn":
		fileLabel = "文件"
		refreshLabel = "立即刷新"
		addFeedLabel = "添加订阅"
		settingsLabel = "设置"
	default:
		fileLabel = "File"
		refreshLabel = "Refresh Now"
		addFeedLabel = "Add Feed"
		settingsLabel = "Settings"
	}

	// File Menu
	fileMenu := menu.AddSubmenu(fileLabel)
	fileMenu.Add(refreshLabel).OnClick(func(ctx *application.Context) {
		if h.Fetcher != nil {
			go h.Fetcher.FetchAll(bgCtx)
		}
	}).SetAccelerator("CmdOrCtrl+R")

	fileMenu.Add(addFeedLabel).OnClick(func(ctx *application.Context) {
		app.Event.Emit("show-add-feed")
	}).SetAccelerator("CmdOrCtrl+N")

	fileMenu.AddSeparator()
	fileMenu.Add(settingsLabel).OnClick(func(ctx *application.Context) {
		app.Event.Emit("show-settings")
	}).SetAccelerator("CmdOrCtrl+,")

	menu.AddRole(application.EditMenu)
	menu.AddRole(application.WindowMenu)

	app.Menu.SetApplicationMenu(menu)
}

func main() {
	// Get proper paths for data files
	logPath, err := utils.GetLogPath()
	if err != nil {
		log.Printf("Warning: Could not get log path: %v. Using current directory.", err)
		logPath = "debug.log"
	}

	// Clear previous log by opening in truncate mode
	f, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		// Fallback to stdout
		f = nil
	}
	if f != nil {
		defer f.Close()
		log.SetOutput(f)
	}

	log.Println("Starting application...")

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

	var quitRequested atomic.Bool
	var lastWindowState windowState

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

	shouldCloseToTray := func() bool {
		val, err := db.GetSetting("close_to_tray")
		return err == nil && val == "true"
	}

	// Start background scheduler
	log.Println("Starting background scheduler...")
	bgCtx, bgCancel := context.WithCancel(context.Background())

	// Encryption key for single instance communication (IPC between app instances).
	// This key is used to encrypt/decrypt messages between first and subsequent instances.
	// Note: This is not for sensitive data encryption - it only carries launch arguments.
	// The key is hardcoded per Wails v3 examples since the data exchanged is not sensitive
	// (just signals to bring window to front).
	var encryptionKey = [32]byte{
		0x1e, 0x1f, 0x1c, 0x1d, 0x1a, 0x1b, 0x18, 0x19,
		0x16, 0x17, 0x14, 0x15, 0x12, 0x13, 0x10, 0x11,
		0x0e, 0x0f, 0x0c, 0x0d, 0x0a, 0x0b, 0x08, 0x09,
		0x06, 0x07, 0x04, 0x05, 0x02, 0x03, 0x00, 0x01,
	}

	// Variable to store the main window reference
	var mainWindow application.Window

	log.Println("Starting Wails v3...")

	// Create new Wails v3 application
	app := application.New(application.Options{
		Name:        "MrRSS",
		Description: "A modern, privacy-focused RSS reader",
		LogLevel:    slog.LevelError,
		Assets: application.AssetOptions{
			Handler:    combinedHandler,
			Middleware: APIMiddleware(combinedHandler),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		SingleInstance: func() *application.SingleInstanceOptions {
			// Disable single instance on Linux due to potential D-Bus issues
			if runtime.GOOS == "linux" {
				return nil
			}
			return &application.SingleInstanceOptions{
				UniqueID:      "com.mrrss.app",
				EncryptionKey: encryptionKey,
				OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
					log.Printf("Second instance detected, bringing window to front")
					if mainWindow != nil {
						// Restore window state if it was stored (minimized to tray)
						if lastWindowState.valid.Load() {
							width := lastWindowState.width
							height := lastWindowState.height
							x := lastWindowState.x
							y := lastWindowState.y

							// Ensure minimum window size
							if width < 400 {
								width = 1024
							}
							if height < 300 {
								height = 768
							}

							// Ensure window is at least partially on screen
							if x < -1000 || x > 3000 {
								x = 100
							}
							if y < -1000 || y > 3000 {
								y = 100
							}

							mainWindow.SetSize(width, height)
							mainWindow.SetPosition(x, y)
						}
						// Show and unminimize the window
						mainWindow.Show()
						mainWindow.Restore()
					}
				},
			}
		}(),
	})

	// Set app instance to handler for browser integration
	h.SetApp(app)
	log.Println("Browser integration enabled")

	// Get window dimensions from stored state or defaults
	windowWidth := 1024
	windowHeight := 768
	windowX := 0
	windowY := 0
	restoredFromDB := false

	// Try to restore window state from database
	if x, err := db.GetSetting("window_x"); err == nil && x != "" {
		if y, err := db.GetSetting("window_y"); err == nil && y != "" {
			if width, err := db.GetSetting("window_width"); err == nil && width != "" {
				if height, err := db.GetSetting("window_height"); err == nil && height != "" {
					// Parse values
					var xInt, yInt, widthInt, heightInt int
					if _, err := fmt.Sscanf(x, "%d", &xInt); err == nil {
						if _, err := fmt.Sscanf(y, "%d", &yInt); err == nil {
							if _, err := fmt.Sscanf(width, "%d", &widthInt); err == nil {
								if _, err := fmt.Sscanf(height, "%d", &heightInt); err == nil {
									// Validate values
									if widthInt >= 400 && heightInt >= 300 && widthInt <= 4000 && heightInt <= 3000 {
										if xInt > -1000 && xInt < 3000 && yInt > -1000 && yInt < 3000 {
											windowWidth = widthInt
											windowHeight = heightInt
											windowX = xInt
											windowY = yInt
											restoredFromDB = true
											// Store in memory for minimize-restore
											lastWindowState.width = widthInt
											lastWindowState.height = heightInt
											lastWindowState.x = xInt
											lastWindowState.y = yInt
											lastWindowState.valid.Store(true)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Determine background color based on theme setting
	// Default to dark gray to prevent white flash on startup/close
	// This matches the CSS dark mode background color (#1e1e1e = rgb(30, 30, 30))
	// For "dark" or "auto", use dark background

	// Create main window options
	windowOptions := application.WebviewWindowOptions{
		Name:             "MrRSS-main-window",
		Title:            "MrRSS",
		Width:            windowWidth,
		Height:           windowHeight,
		URL:              "/",
		Frameless:        true,
		BackgroundType:   application.BackgroundTypeTransparent,
		Windows:          application.WindowsWindow{},
		Linux:            application.LinuxWindow{},
		BackgroundColour: application.NewRGBA(0, 0, 0, 0),
	}

	// Set position if restored from DB
	if restoredFromDB {
		windowOptions.X = windowX
		windowOptions.Y = windowY
	}

	// Create main window
	mainWindow = app.Window.NewWithOptions(windowOptions)

	// Setup application menu for macOS
	setupApplicationMenu(app, db, h, bgCtx)

	if !restoredFromDB {
		mainWindow.Center()
	}

	// Helper function to store window state
	storeWindowState := func() {
		if mainWindow == nil {
			return
		}

		w, h := mainWindow.Size()
		x, y := mainWindow.Position()

		// Only store state if it's valid (reasonable size and position)
		if w >= 400 && h >= 300 && w <= 4000 && h <= 3000 {
			if x > -1000 && x < 3000 && y > -1000 && y < 3000 {
				lastWindowState.width = w
				lastWindowState.height = h
				lastWindowState.x = x
				lastWindowState.y = y
				lastWindowState.valid.Store(true)
			}
		}
	}

	// Create system tray if close_to_tray is enabled
	var systemTray *application.SystemTray

	setupSystemTray := func() {
		if systemTray != nil {
			return // Already set up
		}

		systemTray = app.SystemTray.New()
		systemTray.SetIcon(getAppIcon())

		// Create tray menu
		trayMenu := app.NewMenu()

		// Get language for labels
		lang := "en"
		if l, err := db.GetSetting("language"); err == nil && l != "" {
			lang = l
		}

		var showLabel, refreshLabel, quitLabel string
		switch lang {
		case "zh-CN", "zh", "zh-cn":
			showLabel = "显示 MrRSS"
			refreshLabel = "立即刷新"
			quitLabel = "退出"
		default:
			showLabel = "Show MrRSS"
			refreshLabel = "Refresh now"
			quitLabel = "Quit"
		}

		trayMenu.Add(showLabel).OnClick(func(ctx *application.Context) {
			if mainWindow != nil {
				// Restore window state if it was stored
				if lastWindowState.valid.Load() {
					width := lastWindowState.width
					height := lastWindowState.height
					x := lastWindowState.x
					y := lastWindowState.y

					if width < 400 {
						width = 1024
					}
					if height < 300 {
						height = 768
					}
					if x < -1000 || x > 3000 {
						x = 100
					}
					if y < -1000 || y > 3000 {
						y = 100
					}

					mainWindow.SetSize(width, height)
					mainWindow.SetPosition(x, y)
				}
				mainWindow.Show()
				mainWindow.Restore()
			}
		})

		trayMenu.Add(refreshLabel).OnClick(func(ctx *application.Context) {
			if h.Fetcher != nil {
				go h.Fetcher.FetchAll(bgCtx)
			}
		})

		trayMenu.AddSeparator()

		trayMenu.Add(quitLabel).OnClick(func(ctx *application.Context) {
			quitRequested.Store(true)
			app.Quit()
		})

		systemTray.SetMenu(trayMenu)

		// Handle clicks on tray icon to show window
		systemTray.OnClick(func() {
			if mainWindow != nil {
				mainWindow.Show()
				mainWindow.Restore()
			}
		})
	}

	// Track last window close attempt to handle macOS fullscreen properly
	var lastCloseAttempt atomic.Int64

	// Register hook for window closing event
	mainWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		if quitRequested.Load() {
			return // Allow close
		}

		if shouldCloseToTray() {
			// On macOS, handle fullscreen exit gracefully
			if runtime.GOOS == "darwin" {
				now := time.Now().UnixMilli()
				last := lastCloseAttempt.Load()

				// If last close was within 500ms, user clicked close twice quickly
				// This means fullscreen exit completed, proceed with hiding
				if last > 0 && (now-last) < 500 {
					lastCloseAttempt.Store(0) // Reset
					storeWindowState()
					setupSystemTray()
					mainWindow.Hide()
					e.Cancel()
					return
				}

				// First close attempt - try to exit fullscreen
				lastCloseAttempt.Store(now)
				mainWindow.Restore()
				// Cancel this close event
				// If window was fullscreen, user needs to click close again
				// If not fullscreen, Restore() does nothing and next close will proceed
				e.Cancel()
				return
			}

			// Non-macOS platforms: directly hide to tray
			storeWindowState()
			setupSystemTray()
			mainWindow.Hide()
			e.Cancel()
		}
	})

	// Register move and resize handlers to save window state
	mainWindow.RegisterHook(events.Common.WindowDidMove, func(e *application.WindowEvent) {
		storeWindowState()
	})

	mainWindow.RegisterHook(events.Common.WindowDidResize, func(e *application.WindowEvent) {
		storeWindowState()
	})

	// Setup tray on startup if close_to_tray is enabled
	if shouldCloseToTray() {
		setupSystemTray()
	}

	// On macOS, handle dock icon click to show the window
	if runtime.GOOS == "darwin" {
		app.Event.OnApplicationEvent(events.Mac.ApplicationShouldHandleReopen, func(event *application.ApplicationEvent) {
			log.Println("Dock icon clicked, showing window")
			if mainWindow != nil {
				mainWindow.Show()
				mainWindow.Restore()
			}
		})
	}

	// Detect network speed on startup in background
	go func() {
		time.Sleep(2 * time.Second) // Small delay to allow app to start
		log.Println("Detecting network speed...")

		// Get proxy settings
		proxyEnabled, _ := db.GetSetting("proxy_enabled")
		proxyType, _ := db.GetSetting("proxy_type")
		proxyHost, _ := db.GetSetting("proxy_host")
		proxyPort, _ := db.GetSetting("proxy_port")
		proxyUsername, _ := db.GetSetting("proxy_username")
		proxyPassword, _ := db.GetSetting("proxy_password")

		// Create HTTP client with proxy if enabled
		var httpClient *http.Client
		if proxyEnabled == "true" {
			proxyURL := utils.BuildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)
			if proxyURL != "" {
				client, err := utils.CreateHTTPClient(proxyURL, 10*time.Second)
				if err != nil {
					log.Printf("Failed to create HTTP client with proxy: %v", err)
					// Fall back to default client
					httpClient = &http.Client{Timeout: 10 * time.Second}
				} else {
					httpClient = client
				}
			} else {
				httpClient = &http.Client{Timeout: 10 * time.Second}
			}
		} else {
			httpClient = &http.Client{Timeout: 10 * time.Second}
		}

		detector := network.NewDetector(httpClient)
		detectCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		result := detector.DetectSpeed(detectCtx)
		if result.DetectionSuccess {
			db.SetSetting("network_speed", string(result.SpeedLevel))
			db.SetSetting("network_bandwidth_mbps", fmt.Sprintf("%.2f", result.BandwidthMbps))
			db.SetSetting("network_latency_ms", strconv.FormatInt(result.LatencyMs, 10))
			db.SetSetting("max_concurrent_refreshes", strconv.Itoa(result.MaxConcurrency))
			db.SetSetting("last_network_test", result.DetectionTime.Format(time.RFC3339))
			log.Printf("Network detection complete: %s (max concurrency: %d)", result.SpeedLevel, result.MaxConcurrency)
		} else {
			log.Printf("Network detection failed: %s", result.ErrorMessage)
		}
	}()

	// Start background scheduler after a delay to allow UI to show first
	go func() {
		time.Sleep(5 * time.Second)
		log.Println("Starting background scheduler...")
		h.StartBackgroundScheduler(bgCtx)
	}()

	log.Println("Window initialized, running app...")

	// Run the application
	err = app.Run()

	// Cleanup when app exits
	log.Println("Shutting down...")

	// Stop background tasks first
	bgCancel()
	// Give some time for tasks to finish
	time.Sleep(500 * time.Millisecond)

	// Close DB with timeout
	done := make(chan struct{})
	go func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
		close(done)
	}()

	select {
	case <-done:
		log.Println("Database closed")
	case <-time.After(2 * time.Second):
		log.Println("Database close timed out")
	}

	if err != nil {
		log.Printf("Error running Wails: %v", err)
		log.Fatal(err)
	}
	log.Println("Application finished")
}

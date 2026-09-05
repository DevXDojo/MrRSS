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

	"MrRSS/internal/ai"
	"MrRSS/internal/database"
	"MrRSS/internal/desktopapi"
	"MrRSS/internal/feed"
	handlers "MrRSS/internal/handlers/core"
	"MrRSS/internal/monitor"
	"MrRSS/internal/network"
	"MrRSS/internal/routes"
	"MrRSS/internal/translation"
	"MrRSS/internal/utils/fileutil"
	"MrRSS/internal/utils/httputil"
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

func main() {
	// Get proper paths for data files
	logPath, err := fileutil.GetLogPath()
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
	if fileutil.IsPortableMode() {
		log.Println("Running in PORTABLE mode")
	} else {
		log.Println("Running in NORMAL mode")
	}

	log.Printf("Log file: %s", logPath)

	// Get database path
	dbPath, err := fileutil.GetDBPath()
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

	// Initialize AI profile provider
	profileProvider := ai.NewProfileProvider(db)
	translator := translation.NewDynamicTranslatorWithCache(db, db)
	translator.SetProfileProvider(profileProvider)

	fetcher := feed.NewFetcher(db)
	h := handlers.NewHandler(db, fetcher, translator, profileProvider)

	var quitRequested atomic.Bool
	var lastMaximized atomic.Bool
	var hiddenToTray atomic.Bool
	var hideAfterFullscreen atomic.Bool

	// API Routes
	log.Println("Setting up API routes...")
	apiMux := http.NewServeMux()
	routes.RegisterAPIRoutes(apiMux, h)

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
	showMainWindow := func() {
		if mainWindow == nil {
			return
		}
		hideAfterFullscreen.Store(false)
		// Read the snapshot before Show/UnMinimise can emit native state events.
		restoreMaximized := hiddenToTray.Load() && lastMaximized.Load()
		showExistingWindow(mainWindow, restoreMaximized)
		hiddenToTray.Store(false)
	}

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
					showMainWindow()
				},
			}
		}(),
	})

	// Set app instance to handler for browser integration
	h.SetApp(app)
	log.Println("Browser integration enabled")

	// Expose the API to local integrations such as the mrrss-assistant skill.
	// The listener is loopback-only and deliberately does not serve frontend assets.
	desktopAPIServer, apiErr := desktopapi.Start(
		desktopapi.DefaultAddress,
		routes.WrapWithMiddleware(apiMux, routes.DefaultConfig()),
	)
	if apiErr != nil {
		log.Printf("Local desktop API unavailable: %v", apiErr)
	} else {
		log.Printf("Local desktop API listening on http://%s/api", desktopAPIServer.Address())
		go func() {
			if serveErr := <-desktopAPIServer.Errors(); serveErr != nil {
				log.Printf("Local desktop API stopped unexpectedly: %v", serveErr)
			}
		}()
	}

	// Get window dimensions from stored state or defaults
	windowWidth := 1024
	windowHeight := 768
	windowX := 0
	windowY := 0
	restoredFromDB := false
	restoredMaximized := false

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

	if maximized, err := db.GetSetting("window_maximized"); err == nil && maximized == "true" {
		restoredMaximized = true
		lastMaximized.Store(true)
	}

	// Determine background color based on theme setting
	// Default to dark gray to prevent white flash on startup/close
	// This matches the CSS dark mode background color (#1e1e1e = rgb(30, 30, 30))
	backgroundColour := application.NewRGB(30, 30, 30)
	if theme, err := db.GetSetting("theme"); err == nil {
		if theme == "light" {
			// Use white for light theme
			backgroundColour = application.NewRGB(255, 255, 255)
		}
		// For "dark" or "auto", use dark background
	}

	// Create main window options
	windowOptions := application.WebviewWindowOptions{
		Name:             "MrRSS-main-window",
		Title:            "MrRSS",
		Width:            windowWidth,
		Height:           windowHeight,
		URL:              "/",
		Mac:              application.MacWindow{},
		Windows:          application.WindowsWindow{},
		Linux:            application.LinuxWindow{},
		BackgroundColour: backgroundColour,
	}

	// Set position if restored from DB
	if restoredFromDB {
		windowOptions.X = windowX
		windowOptions.Y = windowY
	}

	// Create main window
	mainWindow = app.Window.NewWithOptions(windowOptions)

	if !restoredFromDB {
		mainWindow.Center()
	}
	if restoredMaximized {
		mainWindow.Maximise()
	}

	// Capture only at an explicit hide. Native windows keep their own normal bounds;
	// replaying Size/Position on restore can unmaximize the window and emit resize races.
	storeWindowState := func() {
		if mainWindow != nil && !hiddenToTray.Load() && !mainWindow.IsMinimised() {
			lastMaximized.Store(mainWindow.IsMaximised())
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
			showMainWindow()
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
			showMainWindow()
		})
	}

	// Fullscreen exits asynchronously on macOS. Hide on its completion event,
	// rather than requiring a second close within an arbitrary time window.
	mainWindow.RegisterHook(events.Common.WindowUnFullscreen, func(e *application.WindowEvent) {
		if hideAfterFullscreen.Swap(false) && !quitRequested.Load() {
			hiddenToTray.Store(true)
			mainWindow.Hide()
		}
	})

	mainWindow.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		if quitRequested.Load() || !shouldCloseToTray() {
			return
		}
		e.Cancel()
		if hideAfterFullscreen.Load() {
			return
		}
		storeWindowState()
		setupSystemTray()
		if runtime.GOOS == "darwin" && mainWindow.IsFullscreen() {
			hideAfterFullscreen.Store(true)
			mainWindow.UnFullscreen()
			return
		}
		hiddenToTray.Store(true)
		mainWindow.Hide()
	})

	mainWindow.RegisterHook(events.Common.WindowMaximise, func(e *application.WindowEvent) {
		if hiddenToTray.Load() {
			return
		}
		lastMaximized.Store(true)
		if err := db.SetSetting("window_maximized", "true"); err != nil {
			log.Printf("Failed to save maximized window state: %v", err)
		}
	})

	mainWindow.RegisterHook(events.Common.WindowUnMaximise, func(e *application.WindowEvent) {
		if hiddenToTray.Load() {
			return
		}
		lastMaximized.Store(false)
		if err := db.SetSetting("window_maximized", "false"); err != nil {
			log.Printf("Failed to save unmaximized window state: %v", err)
		}
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
			showMainWindow()
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
			proxyURL := httputil.BuildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)
			if proxyURL != "" {
				client, err := httputil.CreateHTTPClient(proxyURL, 10*time.Second)
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

	// Report app startup to analytics (non-blocking)
	go func() {
		time.Sleep(2 * time.Second) // Small delay to ensure app starts smoothly
		monitorClient := monitor.NewMonitorClient("https://cf-monitor-api.ch3nyang.workers.dev", "mrrss")
		_ = monitorClient.ReportAppStart(context.Background())
	}()

	log.Println("Window initialized, running app...")

	// Run the application
	err = app.Run()

	// Cleanup when app exits
	log.Println("Shutting down...")
	if desktopAPIServer != nil {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		if shutdownErr := desktopAPIServer.Shutdown(shutdownCtx); shutdownErr != nil {
			log.Printf("Local desktop API shutdown failed: %v", shutdownErr)
		}
		shutdownCancel()
	}

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

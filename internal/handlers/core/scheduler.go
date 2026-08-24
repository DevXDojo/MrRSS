package core

import (
	"context"
	"log"
	"strconv"
	"time"

	"MrRSS/internal/cache"
	"MrRSS/internal/models"
	"MrRSS/internal/utils/fileutil"
)

// StartBackgroundScheduler starts the background scheduler for auto-updates and cleanup.
func (h *Handler) StartBackgroundScheduler(ctx context.Context) {
	// Trigger initial cleanup on startup
	go func() {
		log.Println("Triggering initial cleanup on startup")
		h.Fetcher.GetCleanupManager().RequestCleanup()
	}()

	// Run initial cleanup only if auto_cleanup is enabled
	// This is now handled by the cleanup manager
	go func() {
		autoCleanup, _ := h.DB.GetSetting("auto_cleanup_enabled")
		if autoCleanup == "true" {
			log.Println("Auto cleanup enabled, will run after tasks complete")
		}

		// Run initial media cache cleanup if enabled
		mediaCacheEnabled, _ := h.DB.GetSetting("media_cache_enabled")
		if mediaCacheEnabled == "true" {
			log.Println("Running initial media cache cleanup...")
			h.cleanupMediaCache()
		}
	}()

	// Start the scheduler based on refresh mode
	refreshMode, _ := h.DB.GetSetting("refresh_mode")

	switch refreshMode {
	case "never":
		// Never auto-refresh - only allow manual refresh
		log.Println("Auto-refresh disabled (never mode)")
		return
	case "intelligent":
		// Use intelligent refresh mode
		h.startScheduler(ctx, true)
	default:
		// Use fixed interval mode (default)
		h.startScheduler(ctx, false)
	}
}

// startScheduler is the unified scheduler that handles both fixed and intelligent modes
// Logic:
// 1. Individual feeds with custom intervals (RefreshInterval != 0) are scheduled individually
// 2. Global refresh is triggered at the global interval for feeds using global setting (RefreshInterval == 0)
func (h *Handler) startScheduler(ctx context.Context, intelligentMode bool) {
	modeName := "fixed"
	if intelligentMode {
		modeName = "intelligent"
	}
	log.Printf("Starting %s interval scheduler", modeName)

	// Track last global refresh time
	// Try to load from settings, or initialize to current time if not exists
	lastGlobalRefreshStr, err := h.DB.GetSetting("last_global_refresh")
	var lastGlobalRefresh time.Time
	if err != nil {
		// First time running - set to current time and save
		lastGlobalRefresh = time.Now()
		lastGlobalRefreshStr = lastGlobalRefresh.Format(time.RFC3339)
		h.DB.SetSetting("last_global_refresh", lastGlobalRefreshStr)
		log.Printf("First run - initialized last_global_refresh to %v", lastGlobalRefresh)
		// Trigger initial global refresh for first-time users
		go h.triggerGlobalRefresh(ctx, intelligentMode, &lastGlobalRefresh)
	} else {
		// Parse stored time
		lastGlobalRefresh, err = time.Parse(time.RFC3339, lastGlobalRefreshStr)
		if err != nil {
			log.Printf("Failed to parse last_global_refresh (%s): %v, resetting to now", lastGlobalRefreshStr, err)
			lastGlobalRefresh = time.Now()
		} else {
			log.Printf("Loaded last_global_refresh from settings: %v", lastGlobalRefresh)
		}
	}

	// Get global interval
	getGlobalInterval := func() time.Duration {
		intervalStr, _ := h.DB.GetSetting("update_interval")
		interval := 30
		if i, err := strconv.Atoi(intervalStr); err == nil && i > 0 {
			interval = i
		}
		return time.Duration(interval) * time.Minute
	}

	globalInterval := getGlobalInterval()
	log.Printf("Global refresh interval: %v", globalInterval)

	// Use a ticker to check every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping scheduler")
			return
		case <-ticker.C:
			// Check if we need to trigger global refresh
			timeSinceLastGlobal := time.Since(lastGlobalRefresh)
			if timeSinceLastGlobal >= globalInterval {
				log.Printf("Time since last global refresh (%v) >= global interval (%v), triggering global refresh",
					timeSinceLastGlobal, globalInterval)
				// Trigger global refresh for feeds using global setting
				// Note: lastGlobalRefresh will be updated inside triggerGlobalRefresh
				go h.triggerGlobalRefresh(ctx, intelligentMode, &lastGlobalRefresh)
			}

			// Schedule individual feeds with custom intervals
			go h.scheduleIndividualFeeds(ctx, intelligentMode)
		}
	}
}

// triggerGlobalRefresh triggers a global refresh for all feeds with RefreshInterval == 0
// In intelligent mode, this calculates intervals per feed
// In fixed mode, all feeds refresh together at the global interval
func (h *Handler) triggerGlobalRefresh(ctx context.Context, intelligentMode bool, lastGlobalRefresh *time.Time) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		log.Printf("Error getting feeds for global refresh: %v", err)
		return
	}

	// Filter feeds that use global setting (RefreshInterval == 0)
	// Skip feeds with RefreshInterval == -2 (never refresh)
	globalFeeds := make([]models.Feed, 0)
	for _, feed := range feeds {
		if feed.RefreshInterval == 0 {
			globalFeeds = append(globalFeeds, feed)
		}
	}

	// Check if there are any refreshable feeds (excluding FreshRSS feeds)
	refreshableFeeds := make([]models.Feed, 0)
	for _, feed := range globalFeeds {
		if !feed.IsFreshRSSSource {
			refreshableFeeds = append(refreshableFeeds, feed)
		}
	}

	// If no refreshable feeds, skip updating last_global_refresh
	// This allows the next refresh to be triggered when feeds are added
	if len(refreshableFeeds) == 0 {
		log.Printf("No refreshable feeds found (all %d feeds are FreshRSS sources), skipping global refresh", len(globalFeeds))
		return
	}

	// Update in-memory timestamp and database setting
	*lastGlobalRefresh = time.Now()
	log.Printf("Global refresh triggered at %v", *lastGlobalRefresh)

	// Save to settings for persistence across restarts
	lastGlobalRefreshStr := lastGlobalRefresh.Format(time.RFC3339)
	if err := h.DB.SetSetting("last_global_refresh", lastGlobalRefreshStr); err != nil {
		log.Printf("Failed to save last_global_refresh to settings: %v", err)
	}

	log.Printf("Triggering global refresh for %d refreshable feeds (skipped %d FreshRSS feeds, intelligent mode: %v)",
		len(refreshableFeeds), len(globalFeeds)-len(refreshableFeeds), intelligentMode)

	if intelligentMode {
		// In intelligent mode, schedule each feed individually with calculated intervals
		calculator := h.Fetcher.GetIntelligentRefreshCalculator()
		for _, feed := range refreshableFeeds {
			interval := calculator.CalculateInterval(feed)
			staggerDelay := h.Fetcher.GetStaggeredDelay(feed.ID, len(refreshableFeeds))

			go func(f models.Feed, delay time.Duration, calculatedInterval time.Duration) {
				time.Sleep(delay)
				select {
				case <-ctx.Done():
					return
				default:
					log.Printf("Auto-refreshing feed %s (intelligent mode, interval: %v)", f.Title, calculatedInterval)
					h.Fetcher.FetchSingleFeed(ctx, f, false)
				}
			}(feed, staggerDelay, interval)
		}
	} else {
		// In fixed mode, refresh all feeds together
		h.Fetcher.FetchAll(ctx)
	}

	// Run media cache cleanup if enabled
	mediaCacheEnabled, _ := h.DB.GetSetting("media_cache_enabled")
	if mediaCacheEnabled == "true" {
		h.cleanupMediaCache()
	}
}

// scheduleIndividualFeeds schedules feeds with custom intervals (RefreshInterval != 0)
// These feeds are refreshed independently of the global refresh cycle
func (h *Handler) scheduleIndividualFeeds(ctx context.Context, intelligentMode bool) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		log.Printf("Error getting feeds for individual scheduling: %v", err)
		return
	}

	calculator := h.Fetcher.GetIntelligentRefreshCalculator()

	for _, feed := range feeds {
		// Skip feeds using global setting (RefreshInterval == 0)
		if feed.RefreshInterval == 0 {
			continue
		}

		// Skip feeds with never refresh mode (RefreshInterval == -2)
		if feed.RefreshInterval == -2 {
			continue
		}

		// Skip FreshRSS feeds - they are refreshed via sync, not standard refresh
		if feed.IsFreshRSSSource {
			continue
		}

		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Determine refresh interval
		var refreshInterval time.Duration
		if feed.RefreshInterval > 0 {
			// Use custom fixed interval
			refreshInterval = time.Duration(feed.RefreshInterval) * time.Minute
		} else if feed.RefreshInterval == -1 {
			// Use intelligent interval
			refreshInterval = calculator.CalculateInterval(feed)
		}

		// Check if feed needs refresh based on last_updated time
		timeSinceUpdate := time.Since(feed.LastUpdated)
		if timeSinceUpdate >= refreshInterval {
			// Apply staggered delay
			staggerDelay := h.Fetcher.GetStaggeredDelay(feed.ID, len(feeds))

			// Schedule feed refresh
			feedCopy := feed
			go func(f models.Feed, delay time.Duration, interval time.Duration) {
				time.Sleep(delay)
				select {
				case <-ctx.Done():
					return
				default:
					log.Printf("Auto-refreshing feed %s (custom interval: %v)", f.Title, interval)
					h.Fetcher.FetchSingleFeed(ctx, f, false)
				}
			}(feedCopy, staggerDelay, refreshInterval)
		}
	}
}

// cleanupMediaCache performs media cache cleanup based on settings
func (h *Handler) cleanupMediaCache() {
	cacheDir, err := fileutil.GetMediaCacheDir()
	if err != nil {
		log.Printf("Failed to get media cache directory: %v", err)
		return
	}

	mediaCache, err := cache.NewMediaCache(cacheDir)
	if err != nil {
		log.Printf("Failed to initialize media cache: %v", err)
		return
	}

	// Get settings
	maxAgeDaysStr, _ := h.DB.GetSetting("media_cache_max_age_days")
	maxSizeMBStr, _ := h.DB.GetSetting("media_cache_max_size_mb")

	maxAgeDays, err := strconv.Atoi(maxAgeDaysStr)
	if err != nil || maxAgeDays <= 0 {
		maxAgeDays = 7 // Default
	}

	maxSizeMB, err := strconv.Atoi(maxSizeMBStr)
	if err != nil || maxSizeMB <= 0 {
		maxSizeMB = 100 // Default
	}

	// Cleanup by age
	ageCount, err := mediaCache.CleanupOldFiles(maxAgeDays)
	if err != nil {
		log.Printf("Failed to cleanup old media files: %v", err)
	} else if ageCount > 0 {
		log.Printf("Media cache cleanup: removed %d old files", ageCount)
	}

	// Cleanup by size
	sizeCount, err := mediaCache.CleanupBySize(maxSizeMB)
	if err != nil {
		log.Printf("Failed to cleanup media files by size: %v", err)
	} else if sizeCount > 0 {
		log.Printf("Media cache cleanup: removed %d files to stay under size limit", sizeCount)
	}
}

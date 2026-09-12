package freshrss

import (
	"MrRSS/internal/database"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"MrRSS/internal/freshrss"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

var activeSyncs sync.Map

func requestProvider(r *http.Request) string {
	if strings.HasPrefix(r.URL.Path, "/api/miniflux/") {
		return "miniflux"
	}
	return "freshrss"
}

type syncKey struct {
	db       *database.DB
	provider string
}

// HandleSyncFeed syncs articles for a single FreshRSS feed
// @Summary      Sync single FreshRSS feed
// @Description  Synchronize articles for a specific FreshRSS feed/stream
// @Tags         freshrss
// @Accept       json
// @Produce      json
// @Param        stream_id  query     string  true  "FreshRSS stream ID"
// @Success      200  {object}  map[string]interface{}  "Sync started status (status, message)"
// @Failure      400  {object}  map[string]string  "Bad request (FreshRSS disabled or stream_id missing)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /freshrss/sync-feed [post]
func HandleSyncFeed(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get stream_id from query parameter
	streamID := r.URL.Query().Get("stream_id")
	if streamID == "" {
		response.Error(w, fmt.Errorf("stream_id is required"), http.StatusBadRequest)
		return
	}

	provider := requestProvider(r)
	enabled, err := h.DB.GetSetting(provider + "_enabled")
	if err != nil || enabled != "true" {
		response.Error(w, fmt.Errorf("reader sync is disabled"), http.StatusBadRequest)
		return
	}
	serverURL, username, password, err := h.DB.GetReaderConfig(provider)
	if err != nil || serverURL == "" || username == "" || password == "" {
		response.Error(w, fmt.Errorf("reader settings incomplete"), http.StatusBadRequest)
		return
	}
	key := syncKey{h.DB, provider}
	if _, running := activeSyncs.LoadOrStore(key, true); running {
		response.Error(w, fmt.Errorf("reader sync is already running"), http.StatusConflict)
		return
	}

	var count int
	if err := h.DB.QueryRow("SELECT COUNT(*) FROM feeds WHERE is_freshrss_source = 1 AND sync_provider = ? AND freshrss_stream_id = ?", provider, streamID).Scan(&count); err != nil || count == 0 {
		activeSyncs.Delete(key)
		response.Error(w, fmt.Errorf("stream does not belong to reader"), http.StatusBadRequest)
		return
	}

	// Create bidirectional sync service
	syncService := freshrss.NewBidirectionalSyncServiceForProvider(serverURL, username, password, provider, h.DB)
	log.Printf("[HandleSyncFeed] Syncing stream: %s", streamID)

	// Perform sync in background
	go func() {
		defer activeSyncs.Delete(key)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		count, err := syncService.SyncFeed(ctx, streamID)
		_ = h.DB.SetSetting(provider+"_last_sync_time", time.Now().Format(time.RFC3339Nano))

		if err != nil {
			log.Printf("FreshRSS feed sync failed for stream %s: %v", streamID, err)
		} else {
			log.Printf("FreshRSS feed sync completed for stream %s: %d articles", streamID, count)
		}
	}()

	// Return success response immediately
	response.JSON(w, map[string]interface{}{
		"status":  "sync_started",
		"message": "Feed synchronization started",
	})
}

// HandleSync performs bidirectional synchronization with FreshRSS server
// @Summary      Sync with FreshRSS
// @Description  Perform bidirectional synchronization with FreshRSS server (pull and push changes)
// @Tags         freshrss
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Sync started status (status, message)"
// @Failure      400  {object}  map[string]string  "Bad request (FreshRSS disabled or incomplete settings)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /freshrss/sync [post]
func HandleSync(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	log.Printf("[HandleSync] Sync request received")
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	provider := requestProvider(r)
	enabled, err := h.DB.GetSetting(provider + "_enabled")
	if err != nil || enabled != "true" {
		response.Error(w, fmt.Errorf("reader sync is disabled"), http.StatusBadRequest)
		return
	}
	serverURL, username, password, err := h.DB.GetReaderConfig(provider)
	if err != nil || serverURL == "" || username == "" || password == "" {
		response.Error(w, fmt.Errorf("reader settings incomplete"), http.StatusBadRequest)
		return
	}
	key := syncKey{h.DB, provider}
	if _, running := activeSyncs.LoadOrStore(key, true); running {
		response.Error(w, fmt.Errorf("reader sync is already running"), http.StatusConflict)
		return
	}

	// Create bidirectional sync service
	syncService := freshrss.NewBidirectionalSyncServiceForProvider(serverURL, username, password, provider, h.DB)
	log.Printf("[HandleSync] Sync service created, starting sync")

	// Perform sync in background
	go func() {
		defer activeSyncs.Delete(key)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		result, err := syncService.Sync(ctx)

		// Update last sync time
		lastSyncTime := time.Now().Format(time.RFC3339Nano)
		_ = h.DB.SetSetting(provider+"_last_sync_time", lastSyncTime)

		if err != nil {
			log.Printf("FreshRSS sync failed: %v", err)
		} else {
			log.Printf("FreshRSS sync completed: pull=%d changes, push=%d changes, duration=%s",
				result.PullChangesCount, result.PushChangesCount, result.Duration)
		}
	}()

	// Return success response immediately
	response.JSON(w, map[string]interface{}{
		"status":  "sync_started",
		"message": "FreshRSS synchronization started",
	})
}

// HandleSyncStatus returns the current sync status
// @Summary      Get FreshRSS sync status
// @Description  Get the current synchronization status with FreshRSS
// @Tags         freshrss
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Sync status (pending_changes, failed_items, last_sync_time)"
// @Router       /freshrss/status [get]
func HandleSyncStatus(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get pending count
	pendingCount, err := h.DB.GetPendingSyncCount(requestProvider(r))
	if err != nil {
		log.Printf("Error getting pending sync count: %v", err)
		pendingCount = 0
	}

	// Get failed items
	failedItems, err := h.DB.GetFailedSyncItems(10, requestProvider(r))
	if err != nil {
		log.Printf("Error getting failed sync items: %v", err)
		failedItems = nil
	}

	// Get last sync time from settings
	lastSyncStr, _ := h.DB.GetSetting(requestProvider(r) + "_last_sync_time")
	var lastSyncTime *time.Time
	if lastSyncStr != "" {
		if ts, err := time.Parse(time.RFC3339, lastSyncStr); err == nil {
			lastSyncTime = &ts
		}
	}

	response.JSON(w, map[string]interface{}{
		"pending_changes": pendingCount,
		"failed_items":    len(failedItems),
		"last_sync_time":  lastSyncTime,
	})
}

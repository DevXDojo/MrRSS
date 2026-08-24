package freshrss

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"MrRSS/internal/freshrss"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

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

	// Get FreshRSS settings
	enabled, err := h.DB.GetSetting("freshrss_enabled")
	if err != nil {
		log.Printf("Error getting freshrss_enabled: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	if enabled != "true" {
		response.Error(w, fmt.Errorf("FreshRSS sync is disabled"), http.StatusBadRequest)
		return
	}

	serverURL, _ := h.DB.GetSetting("freshrss_server_url")
	username, _ := h.DB.GetSetting("freshrss_username")
	password, _ := h.DB.GetEncryptedSetting("freshrss_api_password")

	if serverURL == "" || username == "" || password == "" {
		response.Error(w, fmt.Errorf("FreshRSS settings incomplete"), http.StatusBadRequest)
		return
	}

	// Create bidirectional sync service
	syncService := freshrss.NewBidirectionalSyncService(serverURL, username, password, h.DB)
	log.Printf("[HandleSyncFeed] Syncing stream: %s", streamID)

	// Perform sync in background
	go func() {
		ctx := context.Background()
		count, err := syncService.SyncFeed(ctx, streamID)

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

	// Get FreshRSS settings
	enabled, err := h.DB.GetSetting("freshrss_enabled")
	log.Printf("[HandleSync] FreshRSS enabled: %s", enabled)
	if err != nil {
		log.Printf("Error getting freshrss_enabled: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	if enabled != "true" {
		response.Error(w, fmt.Errorf("FreshRSS sync is disabled"), http.StatusBadRequest)
		return
	}

	serverURL, _ := h.DB.GetSetting("freshrss_server_url")
	username, _ := h.DB.GetSetting("freshrss_username")
	password, _ := h.DB.GetEncryptedSetting("freshrss_api_password")

	if serverURL == "" || username == "" || password == "" {
		response.Error(w, fmt.Errorf("FreshRSS settings incomplete"), http.StatusBadRequest)
		return
	}

	// Create bidirectional sync service
	syncService := freshrss.NewBidirectionalSyncService(serverURL, username, password, h.DB)
	log.Printf("[HandleSync] Sync service created, starting sync")

	// Perform sync in background
	go func() {
		ctx := context.Background()
		result, err := syncService.Sync(ctx)

		// Update last sync time
		lastSyncTime := time.Now().Format(time.RFC3339)
		_ = h.DB.SetSetting("freshrss_last_sync_time", lastSyncTime)

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
	pendingCount, err := h.DB.GetPendingSyncCount()
	if err != nil {
		log.Printf("Error getting pending sync count: %v", err)
		pendingCount = 0
	}

	// Get failed items
	failedItems, err := h.DB.GetFailedSyncItems(10)
	if err != nil {
		log.Printf("Error getting failed sync items: %v", err)
		failedItems = nil
	}

	// Get last sync time from settings
	lastSyncStr, _ := h.DB.GetSetting("freshrss_last_sync_time")
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

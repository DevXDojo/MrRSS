package freshrss

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"MrRSS/internal/freshrss"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

func requestProvider(r *http.Request) string {
	if strings.HasPrefix(r.URL.Path, "/api/miniflux/") {
		return "miniflux"
	}
	return "freshrss"
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
	if err := freshrss.StartSync(h.BackgroundContext, h.DB, provider, streamID); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, freshrss.ErrSyncRunning) {
			status = http.StatusConflict
		}
		response.Error(w, err, status)
		return
	}

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
	if err := freshrss.StartSync(h.BackgroundContext, h.DB, provider, ""); err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, freshrss.ErrSyncRunning) {
			status = http.StatusConflict
		}
		response.Error(w, err, status)
		return
	}

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
// @Success      200  {object}  map[string]interface{}  "Sync status (pending_changes, failed_items, last_sync_time, last_error, is_syncing)"
// @Router       /freshrss/status [get]
func HandleSyncStatus(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	provider := requestProvider(r)
	// Get pending count
	pendingCount, err := h.DB.GetPendingSyncCount(provider)
	if err != nil {
		log.Printf("Error getting pending sync count: %v", err)
		pendingCount = 0
	}

	// Get failed items
	failedItems, err := h.DB.GetFailedSyncItems(10, provider)
	if err != nil {
		log.Printf("Error getting failed sync items: %v", err)
		failedItems = nil
	}

	// Get last sync time from settings
	lastSyncStr, _ := h.DB.GetSetting(provider + "_last_sync_time")
	lastError, _ := h.DB.GetSetting(provider + "_last_sync_error")
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
		"last_error":      lastError,
		"is_syncing":      freshrss.IsSyncRunning(h.DB, provider),
	})
}

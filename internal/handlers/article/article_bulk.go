package article

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"MrRSS/internal/database"
	"MrRSS/internal/freshrss"
	"MrRSS/internal/handlers/core"
)

// HandleGetUnreadCounts returns unread counts for all feeds.
// @Summary      Get unread counts
// @Description  Get total unread count and per-feed unread counts
// @Tags         articles
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Unread counts (total + feed_counts map)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/unread-counts [get]
func HandleGetUnreadCounts(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	// Get total unread count
	totalCount, err := h.DB.GetTotalUnreadCount()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get unread counts per feed
	feedCounts, err := h.DB.GetUnreadCountsForAllFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"total":       totalCount,
		"feed_counts": feedCounts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[HandleGetUnreadCounts] ERROR encoding response: %v", err)
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

// HandleGetFilterCounts returns article counts for different filters (unread, favorites, read_later, images).
// @Summary      Get filter-specific feed counts
// @Description  Get per-feed counts for different filter types (unread, favorites, read_later, images)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Filter counts for all filter types"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/filter-counts [get]
func HandleGetFilterCounts(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get unread counts per feed
	unreadCounts, err := h.DB.GetUnreadCountsForAllFeeds()
	if err != nil {
		log.Printf("[HandleGetFilterCounts] ERROR getting unread counts: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get favorite counts per feed
	favoriteCounts, err := h.DB.GetFavoriteCountsForAllFeeds()
	if err != nil {
		log.Printf("[HandleGetFilterCounts] ERROR getting favorite counts: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get favorite AND unread counts per feed
	favoriteUnreadCounts, err := h.DB.GetFavoriteUnreadCountsForAllFeeds()
	if err != nil {
		log.Printf("[HandleGetFilterCounts] ERROR getting favorite unread counts: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get read_later counts per feed
	readLaterCounts, err := h.DB.GetReadLaterCountsForAllFeeds()
	if err != nil {
		log.Printf("[HandleGetFilterCounts] ERROR getting read_later counts: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get read_later AND unread counts per feed
	readLaterUnreadCounts, err := h.DB.GetReadLaterUnreadCountsForAllFeeds()
	if err != nil {
		log.Printf("[HandleGetFilterCounts] ERROR getting read_later unread counts: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get image mode counts per feed
	imageCounts, err := h.DB.GetImageModeCountsForAllFeeds()
	if err != nil {
		log.Printf("[HandleGetFilterCounts] ERROR getting image counts: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get image unread counts per feed
	imageUnreadCounts, err := h.DB.GetImageUnreadCountsForAllFeeds()
	if err != nil {
		log.Printf("[HandleGetFilterCounts] ERROR getting image unread counts: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"unread":            unreadCounts,
		"favorites":         favoriteCounts,
		"favorites_unread":  favoriteUnreadCounts,
		"read_later":        readLaterCounts,
		"read_later_unread": readLaterUnreadCounts,
		"images":            imageCounts,
		"images_unread":     imageUnreadCounts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[HandleGetFilterCounts] ERROR encoding response: %v", err)
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
	}
}

// HandleMarkAllAsRead marks all articles as read.
// @Summary      Mark all articles as read
// @Description  Mark all articles as read globally, by feed, or by category
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        feed_id   query     int64   false  "Mark all as read for specific feed ID"
// @Param        category  query     string  false  "Mark all as read for specific category"
// @Success      200  {string}  string  "Articles marked as read successfully"
// @Failure      400  {object}  map[string]string  "Bad request (invalid feed_id)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/mark-all-read [post]
func HandleMarkAllAsRead(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	feedIDStr := r.URL.Query().Get("feed_id")
	category := r.URL.Query().Get("category")

	var syncReqs []database.SyncRequest
	var err error

	if feedIDStr != "" {
		// Mark all as read for a specific feed
		feedID, parseErr := strconv.ParseInt(feedIDStr, 10, 64)
		if parseErr != nil {
			http.Error(w, "Invalid feed_id parameter", http.StatusBadRequest)
			return
		}
		syncReqs, err = h.DB.MarkAllAsReadForFeedWithSync(feedID)
	} else if category != "" {
		// Mark all as read for a specific category
		syncReqs, err = h.DB.MarkAllAsReadForCategoryWithSync(category)
	} else {
		// Mark all as read globally
		syncReqs, err = h.DB.MarkAllAsReadWithSync()
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Perform immediate sync to FreshRSS if needed
	if len(syncReqs) > 0 {
		go performImmediateBulkSync(h, syncReqs)
	}

	w.WriteHeader(http.StatusOK)
}

// HandleClearReadLater removes all articles from the read later list.
// @Summary      Clear read-later list
// @Description  Remove all articles from the read-later list
// @Tags         articles
// @Accept       json
// @Produce      json
// @Success      200  {string}  string  "Read-later list cleared successfully"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/clear-read-later [post]
func HandleClearReadLater(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := h.DB.ClearReadLater()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandleRefresh triggers a refresh of all feeds.
// @Summary      Refresh all feeds
// @Description  Trigger a background refresh of all feeds
// @Tags         articles
// @Accept       json
// @Produce      json
// @Success      200  {string}  string  "Refresh started successfully"
// @Router       /articles/refresh [post]
func HandleRefresh(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	// Mark progress as running before starting goroutine
	// This ensures the frontend immediately sees is_running=true
	taskManager := h.Fetcher.GetTaskManager()
	taskManager.MarkRunning()

	// Manual refresh - fetches all feeds in background
	go h.Fetcher.FetchAll(context.Background())

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "refreshing"})
}

// HandleCleanupArticles triggers manual cleanup of articles.
// This clears ALL articles and article contents, but keeps feeds and settings.
// @Summary      Cleanup all articles
// @Description  Delete all articles and article contents (keeps feeds and settings)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Cleanup statistics (deleted, articles, contents, type)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/cleanup [post]
func HandleCleanupArticles(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Manual cleanup: clear ALL articles and article contents, but keep feeds
	// Step 1: Delete all article contents
	contentCount, err := h.DB.CleanupAllArticleContents()
	if err != nil {
		log.Printf("Error cleaning up article contents: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Step 2: Delete all articles (but keep feeds and settings)
	articleCount, err := h.DB.DeleteAllArticles()
	if err != nil {
		log.Printf("Error deleting all articles: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Manual cleanup: cleared %d article contents and %d articles", contentCount, articleCount)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"deleted":  contentCount + articleCount,
		"articles": articleCount,
		"contents": contentCount,
		"type":     "all",
	})
}

// HandleCleanupArticleContent triggers manual cleanup of article content cache.
// @Summary      Cleanup article content cache
// @Description  Clear all cached article content (articles remain, only content cache is cleared)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Cleanup result (success, entries_cleaned)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/cleanup-content [post]
func HandleCleanupArticleContent(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := h.DB.CleanupAllArticleContents()
	if err != nil {
		log.Printf("Error cleaning up article content cache: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Cleaned up %d article content entries", count)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":         true,
		"entries_cleaned": count,
	})
}

// HandleGetArticleContentCacheInfo returns information about article content cache.
// @Summary      Get article content cache info
// @Description  Get statistics about the article content cache
// @Tags         articles
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Cache info (cached_articles count)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/content-cache-info [get]
func HandleGetArticleContentCacheInfo(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := h.DB.GetArticleContentCount()
	if err != nil {
		log.Printf("Error getting article content cache info: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"cached_articles": count,
	})
}

// HandleMarkRelativeToArticle marks articles as read relative to a reference article's published time.
// @Summary      Mark articles relative to reference article
// @Description  Marks articles as read based on their published time relative to a reference article (above = newer, below = older)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id        query     int64   true   "Reference article ID"
// @Param        direction query     string true   "Direction: 'above' for newer articles, 'below' for older articles"  Enums(above, below)
// @Param        feed_id   query     int64   false  "Optional: only mark articles from this feed"
// @Param        category  query     string false  "Optional: only mark articles from this category"
// @Success      200  {object}  map[string]interface{}  "Number of articles marked as read"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      404  {object}  map[string]string  "Article not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/mark-relative [post]
func HandleMarkRelativeToArticle(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get reference article ID
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	// Get direction
	direction := r.URL.Query().Get("direction")
	if direction != "above" && direction != "below" {
		http.Error(w, "Invalid direction. Must be 'above' or 'below'", http.StatusBadRequest)
		return
	}

	// Get optional feed_id and category
	var feedID int64
	if feedIDStr := r.URL.Query().Get("feed_id"); feedIDStr != "" {
		feedID, err = strconv.ParseInt(feedIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid feed_id parameter", http.StatusBadRequest)
			return
		}
	}

	category := r.URL.Query().Get("category")

	// Get the reference article to find its published_at time
	article, err := h.DB.GetArticleByID(id)
	if err != nil {
		log.Printf("[HandleMarkRelativeToArticle] Error getting article: %v", err)
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	if article == nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	// Mark articles relative to this article's published time
	count, syncReqs, err := h.DB.MarkArticlesRelativeToPublishedTimeWithSync(article.PublishedAt, direction, feedID, category)
	if err != nil {
		log.Printf("[HandleMarkRelativeToArticle] Error marking articles: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Perform immediate sync to FreshRSS if needed
	if len(syncReqs) > 0 {
		go performImmediateBulkSync(h, syncReqs)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"count":   count,
	})
}

// performImmediateBulkSync performs immediate sync for multiple articles to FreshRSS in a background goroutine
func performImmediateBulkSync(h *core.Handler, syncReqs []database.SyncRequest) {
	// Check if FreshRSS is enabled and configured
	enabled, _ := h.DB.GetSetting("freshrss_enabled")
	if enabled != "true" {
		return
	}

	serverURL, username, password, err := h.DB.GetFreshRSSConfig()
	if err != nil || serverURL == "" || username == "" || password == "" {
		log.Printf("[Bulk Sync] FreshRSS not configured, skipping sync")
		return
	}

	// Create sync service
	syncService := freshrss.NewBidirectionalSyncService(serverURL, username, password, h.DB)

	// Perform immediate sync for each article
	ctx := context.Background()
	successCount := 0
	for _, syncReq := range syncReqs {
		err = syncService.SyncArticleStatus(ctx, syncReq.ArticleID, syncReq.ArticleURL, syncReq.Action)
		if err != nil {
			log.Printf("[Bulk Sync] Failed for article %d: %v", syncReq.ArticleID, err)
			// Enqueue for retry during next global sync
			_ = h.DB.EnqueueSyncChange(syncReq.ArticleID, syncReq.ArticleURL, syncReq.Action)
		} else {
			successCount++
			log.Printf("[Bulk Sync] Success for article %d: %s", syncReq.ArticleID, syncReq.Action)
		}
	}
	log.Printf("[Bulk Sync] Completed: %d/%d articles synced successfully", successCount, len(syncReqs))
}

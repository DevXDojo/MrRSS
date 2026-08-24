package discovery

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"MrRSS/internal/discovery"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

// HandleDiscoverBlogs discovers blogs from a feed's friend links.
// @Summary      Discover blogs from feed
// @Description  Discover new blogs by analyzing friend links from a specific feed's RSS content
// @Tags         discovery
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Discovery request (feed_id)"
// @Success      200  {array}   discovery.DiscoveredBlog  "List of discovered blogs"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      404  {object}  map[string]string  "Feed not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /discovery/blogs [post]
func HandleDiscoverBlogs(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FeedID int64 `json:"feed_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Get the specific feed by ID
	targetFeed, err := h.DB.GetFeedByID(req.FeedID)
	if err != nil {
		if err == sql.ErrNoRows {
			response.Error(w, nil, http.StatusNotFound)
		} else {
			response.Error(w, err, http.StatusInternalServerError)
		}
		return
	}

	// Get all existing feed URLs for deduplication
	subscribedURLs, err := h.DB.GetAllFeedURLs()
	if err != nil {
		log.Printf("Error getting subscribed URLs: %v", err)
		subscribedURLs = make(map[string]bool) // Continue with empty set
	}

	// Discover blogs with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Printf("Starting blog discovery for feed: %s (%s)", targetFeed.Title, targetFeed.URL)
	discovered, err := h.DiscoveryService.DiscoverFromFeed(ctx, targetFeed.URL)
	if err != nil {
		log.Printf("Error discovering blogs: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Filter out already-subscribed feeds
	filtered := make([]discovery.DiscoveredBlog, 0)
	for _, blog := range discovered {
		if !subscribedURLs[blog.RSSFeed] {
			filtered = append(filtered, blog)
		} else {
			log.Printf("Filtering out already-subscribed feed: %s (%s)", blog.Name, blog.RSSFeed)
		}
	}

	// Mark the feed as discovered
	if err := h.DB.MarkFeedDiscovered(req.FeedID); err != nil {
		log.Printf("Error marking feed as discovered: %v", err)
	}

	log.Printf("Discovered %d blogs, %d after filtering", len(discovered), len(filtered))
	response.JSON(w, filtered)
}

// HandleStartSingleDiscovery starts a single feed discovery in the background.
// @Summary      Start single feed discovery
// @Description  Start an asynchronous blog discovery process for a specific feed
// @Tags         discovery
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Discovery request (feed_id)"
// @Success      202  {object}  map[string]string  "Discovery started (status)"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      409  {object}  map[string]string  "Discovery already in progress"
// @Failure      404  {object}  map[string]string  "Feed not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /discovery/single/start [post]
func HandleStartSingleDiscovery(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FeedID int64 `json:"feed_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Check if a discovery is already running
	h.DiscoveryMu.Lock()
	if h.SingleDiscoveryState != nil && h.SingleDiscoveryState.IsRunning {
		h.DiscoveryMu.Unlock()
		response.Error(w, nil, http.StatusConflict)
		return
	}

	// Initialize state
	h.SingleDiscoveryState = &core.DiscoveryState{
		IsRunning:  true,
		IsComplete: false,
		Progress: discovery.Progress{
			Stage:   "starting",
			Message: "Starting discovery",
		},
	}
	h.DiscoveryMu.Unlock()

	// Get the specific feed by ID
	targetFeed, err := h.DB.GetFeedByID(req.FeedID)
	if err != nil {
		h.DiscoveryMu.Lock()
		h.SingleDiscoveryState.IsRunning = false
		h.SingleDiscoveryState.IsComplete = true
		h.SingleDiscoveryState.Error = "Feed not found"
		h.DiscoveryMu.Unlock()
		response.Error(w, nil, http.StatusNotFound)
		return
	}

	// Get all existing feed URLs for deduplication
	subscribedURLs, err := h.DB.GetAllFeedURLs()
	if err != nil {
		log.Printf("Error getting subscribed URLs: %v", err)
		subscribedURLs = make(map[string]bool)
	}

	// Start discovery in background
	go func() {
		// Create a progress callback that updates the state
		progressCb := func(progress discovery.Progress) {
			h.DiscoveryMu.Lock()
			if h.SingleDiscoveryState != nil {
				h.SingleDiscoveryState.Progress = progress
			}
			h.DiscoveryMu.Unlock()
		}

		ctx, cancel := context.WithTimeout(context.Background(), core.SingleFeedDiscoveryTimeout)
		defer cancel()

		log.Printf("Starting background discovery for feed: %s (%s)", targetFeed.Title, targetFeed.URL)
		discovered, err := h.DiscoveryService.DiscoverFromFeedWithProgress(ctx, targetFeed.URL, progressCb)

		h.DiscoveryMu.Lock()
		defer h.DiscoveryMu.Unlock()

		if h.SingleDiscoveryState == nil {
			return
		}

		h.SingleDiscoveryState.IsRunning = false
		h.SingleDiscoveryState.IsComplete = true

		if err != nil {
			log.Printf("Error discovering blogs: %v", err)
			h.SingleDiscoveryState.Error = err.Error()
			return
		}

		// Filter out already-subscribed feeds
		filtered := make([]discovery.DiscoveredBlog, 0)
		for _, blog := range discovered {
			if !subscribedURLs[blog.RSSFeed] {
				filtered = append(filtered, blog)
			}
		}

		h.SingleDiscoveryState.Feeds = filtered

		// Mark the feed as discovered
		if err := h.DB.MarkFeedDiscovered(req.FeedID); err != nil {
			log.Printf("Error marking feed as discovered: %v", err)
		}

		log.Printf("Discovery complete: found %d blogs", len(filtered))
	}()

	w.WriteHeader(http.StatusAccepted)
	response.JSON(w, map[string]string{"status": "started"})
}

// HandleGetSingleDiscoveryProgress returns the current progress of single feed discovery.
// @Summary      Get single discovery progress
// @Description  Get the current progress and status of the single feed discovery operation
// @Tags         discovery
// @Accept       json
// @Produce      json
// @Success      200  {object}  core.DiscoveryState  "Discovery state (is_running, is_complete, progress, feeds, error)"
// @Router       /discovery/single/progress [get]
func HandleGetSingleDiscoveryProgress(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	h.DiscoveryMu.RLock()
	state := h.SingleDiscoveryState
	h.DiscoveryMu.RUnlock()

	if state == nil {
		response.JSON(w, &core.DiscoveryState{
			IsRunning:  false,
			IsComplete: false,
		})
		return
	}

	response.JSON(w, state)
}

// HandleClearSingleDiscovery clears the single feed discovery state.
// @Summary      Clear single discovery state
// @Description  Clear the current single feed discovery state and results
// @Tags         discovery
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string  "Clear status (status)"
// @Router       /discovery/single/clear [post]
func HandleClearSingleDiscovery(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	h.DiscoveryMu.Lock()
	h.SingleDiscoveryState = nil
	h.DiscoveryMu.Unlock()

	w.WriteHeader(http.StatusOK)
	response.JSON(w, map[string]string{"status": "cleared"})
}

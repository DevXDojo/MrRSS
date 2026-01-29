package feed

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/rsshub"
	"MrRSS/internal/utils"
)

// HandleFeeds returns all feeds.
// @Summary      Get all feeds
// @Description  Retrieve all RSS feed subscriptions (passwords are cleared)
// @Tags         feeds
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.Feed  "List of feeds"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /feeds [get]
func HandleFeeds(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Populate tags for each feed
	for i := range feeds {
		tags, _ := h.DB.GetFeedTags(feeds[i].ID)
		feeds[i].Tags = tags
		// Clear sensitive password fields before sending to frontend
		feeds[i].EmailPassword = ""
	}

	json.NewEncoder(w).Encode(feeds)
}

// HandleAddFeed adds a new feed subscription and immediately fetches its articles.
// @Summary      Add a new feed
// @Description  Add a new RSS/Atom/Email/Script/XPath feed subscription
// @Tags         feeds
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Feed details"
// @Success      200  {string}  string  "Feed added successfully"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      409  {object}  map[string]string  "Feed URL already exists"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /feeds/add [post]
func HandleAddFeed(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL              string `json:"url"`
		Category         string `json:"category"`
		Title            string `json:"title"`
		ScriptPath       string `json:"script_path"`
		HideFromTimeline bool   `json:"hide_from_timeline"`
		ProxyURL         string `json:"proxy_url"`
		ProxyEnabled     bool   `json:"proxy_enabled"`
		RefreshInterval  int    `json:"refresh_interval"`
		IsImageMode      bool   `json:"is_image_mode"`
		// XPath fields
		Type                string `json:"type"`
		XPathItem           string `json:"xpath_item"`
		XPathItemTitle      string `json:"xpath_item_title"`
		XPathItemContent    string `json:"xpath_item_content"`
		XPathItemUri        string `json:"xpath_item_uri"`
		XPathItemAuthor     string `json:"xpath_item_author"`
		XPathItemTimestamp  string `json:"xpath_item_timestamp"`
		XPathItemTimeFormat string `json:"xpath_item_time_format"`
		XPathItemThumbnail  string `json:"xpath_item_thumbnail"`
		XPathItemCategories string `json:"xpath_item_categories"`
		XPathItemUid        string `json:"xpath_item_uid"`
		ArticleViewMode     string `json:"article_view_mode"`
		AutoExpandContent   string `json:"auto_expand_content"`
		// Email/Newsletter fields
		EmailAddress    string `json:"email_address"`
		EmailIMAPServer string `json:"email_imap_server"`
		EmailIMAPPort   int    `json:"email_imap_port"`
		EmailUsername   string `json:"email_username"`
		EmailPassword   string `json:"email_password"`
		EmailFolder     string `json:"email_folder"`
		// Tags
		Tags []int64 `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Normalize the URL to ensure it has a protocol
	req.URL = utils.NormalizeFeedURL(req.URL)

	// Determine the feed URL to check for duplicates
	feedURL := req.URL
	if req.ScriptPath != "" {
		feedURL = "script://" + req.ScriptPath
	} else if req.Type == "email" {
		feedURL = "email://" + req.EmailAddress
	}

	// Check if feed with this URL already exists (excluding FreshRSS feeds)
	var existingID int64
	var existingIsFreshRSS bool
	err := h.DB.QueryRow("SELECT id, is_freshrss_source FROM feeds WHERE url = ?", feedURL).Scan(&existingID, &existingIsFreshRSS)
	if err == nil && !existingIsFreshRSS {
		// Feed exists and is not a FreshRSS feed - return conflict error
		http.Error(w, "feed with this URL already exists", http.StatusConflict)
		return
	}

	var feedID int64
	if req.ScriptPath != "" {
		// Add feed using custom script
		feedID, err = h.Fetcher.AddScriptSubscription(req.ScriptPath, req.Category, req.Title)
	} else if req.XPathItem != "" {
		// Add feed using XPath
		feedID, err = h.Fetcher.AddXPathSubscription(req.URL, req.Category, req.Title, req.Type, req.XPathItem, req.XPathItemTitle, req.XPathItemContent, req.XPathItemUri, req.XPathItemAuthor, req.XPathItemTimestamp, req.XPathItemTimeFormat, req.XPathItemThumbnail, req.XPathItemCategories, req.XPathItemUid)
	} else if req.Type == "email" {
		// Add feed as email newsletter subscription
		feedID, err = h.Fetcher.AddEmailSubscription(req.EmailAddress, req.EmailIMAPServer, req.EmailUsername, req.EmailPassword, req.Category, req.Title, req.EmailFolder, req.EmailIMAPPort)
	} else {
		// Add feed using URL
		feedID, err = h.Fetcher.AddSubscription(req.URL, req.Category, req.Title)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update all feed settings
	feed, err := h.DB.GetFeedByID(feedID)
	if err != nil {
		// Log the error but don't fail the request - feed was created successfully
		// The settings can be set later via edit
		http.Error(w, "feed created but failed to update settings: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err := h.DB.UpdateFeed(feed.ID, feed.Title, feed.URL, feed.Category, feed.ScriptPath, req.HideFromTimeline, req.ProxyURL, req.ProxyEnabled, req.RefreshInterval, req.IsImageMode, feed.Type, feed.XPathItem, feed.XPathItemTitle, feed.XPathItemContent, feed.XPathItemUri, feed.XPathItemAuthor, feed.XPathItemTimestamp, feed.XPathItemTimeFormat, feed.XPathItemThumbnail, feed.XPathItemCategories, feed.XPathItemUid, req.ArticleViewMode, req.AutoExpandContent, feed.EmailAddress, feed.EmailIMAPServer, feed.EmailUsername, feed.EmailPassword, feed.EmailFolder, feed.EmailIMAPPort); err != nil {
		http.Error(w, "feed created but failed to update settings: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set tags for the feed
	if len(req.Tags) > 0 {
		if err := h.DB.SetFeedTags(feedID, req.Tags); err != nil {
			// Log error but don't fail - feed was created successfully
			// Tags can be set later via edit
		}
	}

	// Immediately fetch articles for the newly added feed in background
	go func() {
		feed, err := h.DB.GetFeedByID(feedID)
		if err != nil {
			return
		}
		// Use manual refresh (queue head) for newly added feed
		h.Fetcher.FetchSingleFeed(context.Background(), *feed, true)
	}()

	w.WriteHeader(http.StatusOK)
}

// HandleDeleteFeed deletes a feed subscription.
// @Summary      Delete a feed
// @Description  Delete a feed subscription by ID
// @Tags         feeds
// @Accept       json
// @Produce      json
// @Param        id   query      int64  true  "Feed ID"
// @Success      200  {string}  string  "Feed deleted successfully"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /feeds/delete [post]
func HandleDeleteFeed(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := h.DB.DeleteFeed(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// HandleUpdateFeed updates a feed's properties.
// @Summary      Update a feed
// @Description  Update properties of an existing feed subscription
// @Tags         feeds
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Feed update details"
// @Success      200  {string}  string  "Feed updated successfully"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      409  {object}  map[string]string  "Feed URL already exists"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /feeds/update [post]
func HandleUpdateFeed(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID               int64  `json:"id"`
		Title            string `json:"title"`
		URL              string `json:"url"`
		Category         string `json:"category"`
		ScriptPath       string `json:"script_path"`
		HideFromTimeline bool   `json:"hide_from_timeline"`
		ProxyURL         string `json:"proxy_url"`
		ProxyEnabled     bool   `json:"proxy_enabled"`
		RefreshInterval  int    `json:"refresh_interval"`
		IsImageMode      bool   `json:"is_image_mode"`
		// XPath fields
		Type                string `json:"type"`
		XPathItem           string `json:"xpath_item"`
		XPathItemTitle      string `json:"xpath_item_title"`
		XPathItemContent    string `json:"xpath_item_content"`
		XPathItemUri        string `json:"xpath_item_uri"`
		XPathItemAuthor     string `json:"xpath_item_author"`
		XPathItemTimestamp  string `json:"xpath_item_timestamp"`
		XPathItemTimeFormat string `json:"xpath_item_time_format"`
		XPathItemThumbnail  string `json:"xpath_item_thumbnail"`
		XPathItemCategories string `json:"xpath_item_categories"`
		XPathItemUid        string `json:"xpath_item_uid"`
		ArticleViewMode     string `json:"article_view_mode"`
		AutoExpandContent   string `json:"auto_expand_content"`
		// Email/Newsletter fields
		EmailAddress    string `json:"email_address"`
		EmailIMAPServer string `json:"email_imap_server"`
		EmailIMAPPort   int    `json:"email_imap_port"`
		EmailUsername   string `json:"email_username"`
		EmailPassword   string `json:"email_password"`
		EmailFolder     string `json:"email_folder"`
		// Tags
		Tags []int64 `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Normalize the URL to ensure it has a protocol
	req.URL = utils.NormalizeFeedURL(req.URL)

	// Validate RSSHub URL if provided
	if req.URL != "" && rsshub.IsRSSHubURL(req.URL) {
		// Check if RSSHub is enabled
		enabledStr, _ := h.DB.GetSetting("rsshub_enabled")
		if enabledStr != "true" {
			http.Error(w, "RSSHub integration is disabled. Please enable it in settings", http.StatusBadRequest)
			return
		}

		endpoint, _ := h.DB.GetSetting("rsshub_endpoint")
		if endpoint == "" {
			endpoint = "https://rsshub.app"
		}
		apiKey, _ := h.DB.GetEncryptedSetting("rsshub_api_key")

		// Skip validation if API key is empty (public rsshub.app instance with Cloudflare protection)
		if apiKey != "" {
			route := rsshub.ExtractRoute(req.URL)
			client := rsshub.NewClient(endpoint, apiKey)
			if err := client.ValidateRoute(route); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	}

	// Determine the feed URL to check for duplicates
	feedURL := req.URL
	if req.ScriptPath != "" {
		feedURL = "script://" + req.ScriptPath
	} else if req.Type == "email" {
		feedURL = "email://" + req.EmailAddress
	}

	// Check if another feed with this URL already exists (excluding FreshRSS feeds and current feed)
	var existingID int64
	var existingIsFreshRSS bool
	err := h.DB.QueryRow("SELECT id, is_freshrss_source FROM feeds WHERE url = ? AND id != ?", feedURL, req.ID).Scan(&existingID, &existingIsFreshRSS)
	if err == nil && !existingIsFreshRSS {
		// Another feed exists with this URL and is not a FreshRSS feed - return conflict error
		http.Error(w, "feed with this URL already exists", http.StatusConflict)
		return
	}

	// If title is empty, fetch the default title from the feed
	finalTitle := req.Title
	if finalTitle == "" {
		// Get current feed to determine its type
		currentFeed, err := h.DB.GetFeedByID(req.ID)
		if err != nil {
			http.Error(w, "failed to get feed info: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Fetch default title based on feed type
		if req.ScriptPath != "" || (currentFeed != nil && currentFeed.ScriptPath != "") {
			// Script-based feed
			scriptPathToUse := req.ScriptPath
			if scriptPathToUse == "" {
				scriptPathToUse = currentFeed.ScriptPath
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			parsedFeed, err := h.Fetcher.ParseFeedWithScript(ctx, "", scriptPathToUse, false)
			if err == nil && parsedFeed.Title != "" {
				finalTitle = parsedFeed.Title
			} else {
				finalTitle = scriptPathToUse
			}
		} else if req.XPathItem != "" || (currentFeed != nil && currentFeed.XPathItem != "") {
			// XPath-based feed: use default title
			finalTitle = "XPath Feed"
		} else if req.Type == "email" || (currentFeed != nil && currentFeed.Type == "email") {
			// Email-based feed: use email address as title
			emailAddr := req.EmailAddress
			if emailAddr == "" && currentFeed != nil {
				emailAddr = currentFeed.EmailAddress
			}
			finalTitle = emailAddr
		} else {
			// URL-based feed: fetch and parse to get title
			urlToFetch := req.URL
			if urlToFetch == "" && currentFeed != nil {
				urlToFetch = currentFeed.URL
			}
			if urlToFetch != "" {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				parsedFeed, err := h.Fetcher.ParseFeed(ctx, urlToFetch)
				if err == nil && parsedFeed.Title != "" {
					finalTitle = parsedFeed.Title
				} else {
					// Fallback to URL as title
					finalTitle = urlToFetch
				}
			}
		}
	}

	if err := h.DB.UpdateFeed(req.ID, finalTitle, req.URL, req.Category, req.ScriptPath, req.HideFromTimeline, req.ProxyURL, req.ProxyEnabled, req.RefreshInterval, req.IsImageMode, req.Type, req.XPathItem, req.XPathItemTitle, req.XPathItemContent, req.XPathItemUri, req.XPathItemAuthor, req.XPathItemTimestamp, req.XPathItemTimeFormat, req.XPathItemThumbnail, req.XPathItemCategories, req.XPathItemUid, req.ArticleViewMode, req.AutoExpandContent, req.EmailAddress, req.EmailIMAPServer, req.EmailUsername, req.EmailPassword, req.EmailFolder, req.EmailIMAPPort); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update tags for the feed
	if req.Tags != nil {
		if err := h.DB.SetFeedTags(req.ID, req.Tags); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// HandleRefreshFeed refreshes a single feed by ID with progress tracking.
// @Summary      Refresh a single feed
// @Description  Trigger a refresh for a specific feed (runs in background with progress tracking)
// @Tags         feeds
// @Accept       json
// @Produce      json
// @Param        id   query     int64   true  "Feed ID"
// @Success      200  {string}  string  "Feed refresh started successfully"
// @Failure      400  {object}  map[string]string  "Bad request (invalid feed ID)"
// @Failure      404  {object}  map[string]string  "Feed not found"
// @Router       /feeds/refresh [post]
func HandleRefreshFeed(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid feed ID", http.StatusBadRequest)
		return
	}

	feed, err := h.DB.GetFeedByID(id)
	if err != nil {
		http.Error(w, "Feed not found", http.StatusNotFound)
		return
	}

	// Refresh the feed in background with progress tracking (manual = queue head)
	go h.Fetcher.FetchSingleFeed(context.Background(), *feed, true)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "refreshing"})
}

// HandleReorderFeed reorders a feed within or across categories.
// @Summary      Reorder a feed
// @Description  Change the position and optionally the category of a feed
// @Tags         feeds
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Reorder details (feed_id, category, position)"
// @Success      200  {object}  map[string]string  "Reorder status"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /feeds/reorder [post]
func HandleReorderFeed(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		FeedID   int64  `json:"feed_id"`
		Category string `json:"category"`
		Position int    `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.DB.ReorderFeed(req.FeedID, req.Category, req.Position); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

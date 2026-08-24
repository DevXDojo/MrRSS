package article

import (
	"log"
	"net/http"
	"strconv"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

// HandleArticles returns articles with filtering and pagination.
// @Summary      Get articles with filtering
// @Description  Retrieve articles with optional filtering by feed, category, status, and pagination
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        filter    query     string  false  "Filter: 'all', 'unread', 'favorite', 'read_later'"  Enums(all, unread, favorite, read_later)
// @Param        feed_id   query     int64   false  "Filter by feed ID"
// @Param        category  query     string  false  "Filter by category name"
// @Param        only_unread query   bool    false  "Filter for only unread articles"
// @Param        page      query     int     false  "Page number (default: 1)"  minimum(1)
// @Param        limit     query     int     false  "Items per page (default: 50, max: 500)"  minimum(1)  maximum(500)
// @Success      200  {array}   models.Article  "List of articles"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles [get]
func HandleArticles(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	feedIDStr := r.URL.Query().Get("feed_id")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	onlyUnread := r.URL.Query().Get("only_unread") == "true"

	// Check if category parameter exists (even if empty string)
	// We need to distinguish between "no category parameter" and "category='' for uncategorized"
	// Use a special value "\x00" to represent explicit uncategorized filtering
	var category string
	if _, exists := r.URL.Query()["category"]; exists {
		category = r.URL.Query().Get("category")
		// If category is empty string, convert to special value for database layer
		if category == "" {
			category = "\x00" // Special value for uncategorized
		}
	}

	var feedID int64
	if feedIDStr != "" {
		feedID, _ = strconv.ParseInt(feedIDStr, 10, 64)
	}

	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := (page - 1) * limit

	// Get show_hidden_articles setting
	showHiddenStr, _ := h.DB.GetSetting("show_hidden_articles")
	showHidden := showHiddenStr == "true"

	articles, err := h.DB.GetArticlesWithUnreadFilter(filter, feedID, category, showHidden, onlyUnread, limit, offset)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	response.JSON(w, articles)
}

// HandleToggleHideArticle toggles the hidden status of an article.
// @Summary      Toggle article hidden status
// @Description  Toggle the hidden status of an article (hidden articles are filtered out by default)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id   query     int64   true  "Article ID"
// @Success      200  {object}  map[string]bool  "Success status"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/toggle-hide [post]
func HandleToggleHideArticle(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	if err := h.DB.ToggleArticleHidden(id); err != nil {
		log.Printf("Error toggling article hidden status: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]bool{"success": true})
}

// HandleToggleReadLater toggles the read later status of an article.
// @Summary      Toggle article read-later status
// @Description  Toggle the read-later status of an article (add to/remove from reading list)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id   query     int64   true  "Article ID"
// @Success      200  {object}  map[string]bool  "Success status"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/toggle-read-later [post]
func HandleToggleReadLater(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	if err := h.DB.ToggleReadLater(id); err != nil {
		log.Printf("Error toggling article read later status: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]bool{"success": true})
}

// HandleImageGalleryArticles returns articles from image mode feeds with pagination.
// @Summary      Get image gallery articles
// @Description  Retrieve articles from image-mode feeds (visual/rss-gallery feeds) with pagination
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        feed_id     query     int64   false  "Filter by feed ID"
// @Param        category    query     string  false  "Filter by category name"
// @Param        only_unread query     bool    false  "Filter for only unread articles"
// @Param        page        query     int     false  "Page number (default: 1)"  minimum(1)
// @Param        limit       query     int     false  "Items per page (default: 50)"  minimum(1)
// @Success      200  {array}   models.Article  "List of image gallery articles"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/image-gallery [get]
func HandleImageGalleryArticles(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	feedIDStr := r.URL.Query().Get("feed_id")
	onlyUnreadStr := r.URL.Query().Get("only_unread")

	// Check if category parameter exists (even if empty string)
	// We need to distinguish between "no category parameter" and "category='' for uncategorized"
	// Use a special value "\x00" to represent explicit uncategorized filtering
	var category string
	if _, exists := r.URL.Query()["category"]; exists {
		category = r.URL.Query().Get("category")
		// If category is empty string, convert to special value for database layer
		if category == "" {
			category = "\x00" // Special value for uncategorized
		}
	}

	var feedID int64
	if feedIDStr != "" {
		feedID, _ = strconv.ParseInt(feedIDStr, 10, 64)
	}

	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	offset := (page - 1) * limit

	// Get show_hidden_articles setting
	showHiddenStr, _ := h.DB.GetSetting("show_hidden_articles")
	showHidden := showHiddenStr == "true"

	// Parse only_unread parameter
	onlyUnread := onlyUnreadStr == "true"

	articles, err := h.DB.GetImageGalleryArticles(feedID, category, showHidden, onlyUnread, limit, offset)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	response.JSON(w, articles)
}

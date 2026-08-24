package article

import (
	"log"
	"net/http"
	"strconv"

	"MrRSS/internal/feed"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

// HandleGetArticleContent fetches the article content from RSS feed dynamically.
// @Summary      Get article content
// @Description  Fetch the full HTML content of an article (uses cache if available)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id   query     int64   true  "Article ID"
// @Success      200  {object}  map[string]string  "Article content (content, feed_url)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid article ID)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/content [get]
func HandleGetArticleContent(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	articleIDStr := r.URL.Query().Get("id")
	articleID, err := strconv.ParseInt(articleIDStr, 10, 64)
	if err != nil {
		response.Error(w, nil, http.StatusBadRequest)
		return
	}

	// Get the article from database to access feed_id
	article, err := h.DB.GetArticleByID(articleID)
	if err != nil {
		log.Printf("Error getting article: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Use the cached content fetching method
	content, wasCached, err := h.GetArticleContent(articleID)
	if err != nil {
		log.Printf("Error getting article content: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Track article view
	_ = h.DB.IncrementStat("article_view")

	// Get feed URL to use as referer for image proxying
	feed, err := h.DB.GetFeedByID(article.FeedID)
	var feedURL string
	if err == nil && feed != nil {
		feedURL = feed.URL
	}

	response.JSON(w, map[string]interface{}{
		"content":  content,
		"feed_url": feedURL,
		"cached":   wasCached,
	})
}

// HandleReloadArticleContent clears cached content for a single article.
// @Summary      Reload article content
// @Description  Clear cached RSS content for an article so the next content request reloads it
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id   query     int64   true  "Article ID"
// @Success      200  {object}  map[string]string  "Reload status"
// @Failure      400  {object}  map[string]string  "Bad request (invalid article ID)"
// @Failure      404  {object}  map[string]string  "Article not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/reload-content [post]
func HandleReloadArticleContent(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	articleIDStr := r.URL.Query().Get("id")
	articleID, err := strconv.ParseInt(articleIDStr, 10, 64)
	if err != nil {
		response.Error(w, nil, http.StatusBadRequest)
		return
	}

	if _, err := h.DB.GetArticleByID(articleID); err != nil {
		response.Error(w, err, http.StatusNotFound)
		return
	}

	if err := h.DB.DeleteArticleContent(articleID); err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]string{"status": "reloading"})
}

// HandleFetchFullArticle fetches the full article content from the original URL using readability.
// @Summary      Fetch full article content
// @Description  Fetch the full article content from the original URL using readability extraction (requires full_text_fetch_enabled setting)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id   query     int64   true  "Article ID"
// @Success      200  {object}  map[string]string  "Full article content (content, feed_url)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid ID or missing URL)"
// @Failure      403  {object}  map[string]string  "Full-text fetching disabled"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/fetch-full [post]
func HandleFetchFullArticle(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	articleIDStr := r.URL.Query().Get("id")
	articleID, err := strconv.ParseInt(articleIDStr, 10, 64)
	if err != nil {
		response.Error(w, nil, http.StatusBadRequest)
		return
	}

	// Get the article from database
	article, err := h.DB.GetArticleByID(articleID)
	if err != nil {
		log.Printf("Error getting article: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	if article.URL == "" {
		response.Error(w, nil, http.StatusBadRequest)
		return
	}

	// Check if full-text fetching is enabled (global setting only)
	// auto_expand_content only affects auto-expansion behavior, not manual button clicks
	fullTextEnabledStr, _ := h.DB.GetSetting("full_text_fetch_enabled")
	if fullTextEnabledStr != "true" {
		response.Error(w, nil, http.StatusForbidden)
		return
	}

	// Get feed URL to use as referer for image proxying and proxy settings for fetching
	feed, err := h.DB.GetFeedByID(article.FeedID)
	var feedURL string
	if err == nil && feed != nil {
		feedURL = feed.URL
	}

	// Fetch full content
	fullContent, err := h.FetchFullArticleContentWithFeed(article.URL, feed)
	if err != nil {
		log.Printf("Error fetching full article content: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]string{
		"content":  fullContent,
		"feed_url": feedURL,
	})
}

// HandleExtractAllImages extracts all image URLs from article content
// @Summary      Extract all images from article
// @Description  Extract all image URLs from article content (including relative URLs resolved to absolute)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id   query     int64   true  "Article ID"
// @Success      200  {object}  map[string]interface{}  "List of image URLs (images, feed_url)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid article ID)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/extract-images [get]
func HandleExtractAllImages(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	articleIDStr := r.URL.Query().Get("id")
	articleID, err := strconv.ParseInt(articleIDStr, 10, 64)
	if err != nil {
		response.Error(w, nil, http.StatusBadRequest)
		return
	}

	// Get the article from database
	article, err := h.DB.GetArticleByID(articleID)
	if err != nil {
		log.Printf("Error getting article: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Get feed URL to use as base for resolving relative URLs
	feedObj, err := h.DB.GetFeedByID(article.FeedID)
	var feedURL string
	if err == nil && feedObj != nil {
		feedURL = feedObj.URL
	}

	// Get article content
	content, _, err := h.GetArticleContent(articleID)
	if err != nil {
		log.Printf("Error getting article content: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Extract all images from content
	rawImageURLs := feed.ExtractAllImageURLsFromHTML(content)

	// Resolve all relative URLs to absolute
	var resolvedImageURLs []string
	for _, imgURL := range rawImageURLs {
		resolvedURL := feed.ResolveRelativeURL(imgURL, feedURL)
		if resolvedURL != "" {
			resolvedImageURLs = append(resolvedImageURLs, resolvedURL)
		}
	}

	response.JSON(w, map[string]interface{}{
		"images":   resolvedImageURLs,
		"feed_url": feedURL,
	})
}

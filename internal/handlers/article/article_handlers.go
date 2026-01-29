package article

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
	"MrRSS/internal/rsshub"
)

// GetFeedType returns the type code of a feed
// Possible values: "regular", "freshrss", "rsshub", "script", "xpath", "email"
func GetFeedType(feed *models.Feed) string {
	// Check FreshRSS
	if feed.IsFreshRSSSource {
		return "freshrss"
	}

	// Check RSSHub
	if rsshub.IsRSSHubURL(feed.URL) {
		return "rsshub"
	}

	// Check custom script
	if feed.ScriptPath != "" {
		return "script"
	}

	// Check email
	if feed.Type == "email" {
		return "email"
	}

	// Check XPath
	if feed.Type == "HTML+XPath" || feed.Type == "XML+XPath" {
		return "xpath"
	}

	// Default: regular RSS/Atom feed
	return "regular"
}

// HandleProgress returns the current fetch progress with statistics.
// @Summary      Get fetch progress
// @Description  Get the current feed fetching progress with statistics
// @Tags         articles
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Progress information"
// @Router       /progress [get]
func HandleProgress(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	progress := h.Fetcher.GetProgressWithStats()

	// Log for debugging
	log.Printf("[HandleProgress] Returning progress: is_running=%v, pool=%d, queue=%d",
		progress.IsRunning, progress.PoolTaskCount, progress.QueueTaskCount)

	if err := json.NewEncoder(w).Encode(progress); err != nil {
		log.Printf("[HandleProgress] ERROR encoding progress: %v", err)
		http.Error(w, fmt.Sprintf("Failed to encode progress: %v", err), http.StatusInternalServerError)
	}
}

// TaskDetailsResponse contains detailed task information
type TaskDetailsResponse struct {
	PoolTasks  []PoolTaskInfo  `json:"pool_tasks"`
	QueueTasks []QueueTaskInfo `json:"queue_tasks"`
}

// PoolTaskInfo contains information about a task in the pool
type PoolTaskInfo struct {
	FeedID    int64  `json:"feed_id"`
	FeedTitle string `json:"feed_title"`
	Reason    int    `json:"reason"`
	CreatedAt string `json:"created_at"`
}

// QueueTaskInfo contains information about a task in the queue
type QueueTaskInfo struct {
	FeedID    int64  `json:"feed_id"`
	FeedTitle string `json:"feed_title"`
	Position  int    `json:"position"`
}

// HandleTaskDetails returns detailed information about tasks in pool and queue
// @Summary      Get task details
// @Description  Get detailed information about tasks in pool and queue
// @Tags         articles
// @Accept       json
// @Produce      json
// @Success      200  {object}  TaskDetailsResponse  "Task details"
// @Router       /progress/task-details [get]
func HandleTaskDetails(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	tm := h.Fetcher.GetTaskManager()

	// Get pool tasks
	poolTasksRaw := tm.GetPoolTasks()
	poolTasks := make([]PoolTaskInfo, len(poolTasksRaw))
	for i, task := range poolTasksRaw {
		poolTasks[i] = PoolTaskInfo{
			FeedID:    task.FeedID,
			FeedTitle: task.FeedTitle,
			Reason:    int(task.Reason),
			CreatedAt: task.CreatedAt.Format(time.RFC3339),
		}
	}

	// Sort pool tasks alphabetically by feed title
	sort.Slice(poolTasks, func(i, j int) bool {
		return poolTasks[i].FeedTitle < poolTasks[j].FeedTitle
	})

	// Get queue tasks (limit to first 3)
	queueTasksRaw := tm.GetQueueTasks(3)
	queueTasks := make([]QueueTaskInfo, len(queueTasksRaw))
	for i, task := range queueTasksRaw {
		queueTasks[i] = QueueTaskInfo{
			FeedID:    task.FeedID,
			FeedTitle: task.FeedTitle,
			Position:  task.Position,
		}
	}

	response := TaskDetailsResponse{
		PoolTasks:  poolTasks,
		QueueTasks: queueTasks,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode task details: %v", err), http.StatusInternalServerError)
	}
}

// HandleFilteredArticles returns articles filtered by advanced conditions from the database.
// @Summary      Get filtered articles
// @Description  Retrieve articles with advanced filtering conditions
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        request  body      FilterRequest  true  "Filter criteria"
// @Success      200  {object}  FilterResponse  "Filtered articles"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/filter [post]
func HandleFilteredArticles(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set default pagination values
	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 {
		limit = 50
	}

	// Get show_hidden_articles setting
	showHiddenStr, _ := h.DB.GetSetting("show_hidden_articles")
	showHidden := showHiddenStr == "true"

	// Get all articles from database
	// Note: Using a high limit to fetch all articles for filtering
	// For very large datasets, consider implementing database-level filtering
	articles, err := h.DB.GetArticles("", 0, "", showHidden, 50000, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get feeds for category lookup
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create maps of feed ID to feed data
	feedCategories := make(map[int64]string)
	feedTypes := make(map[int64]string)
	feedIsImageMode := make(map[int64]bool)
	feedTags := make(map[int64][]string)

	for _, feed := range feeds {
		feedCategories[feed.ID] = feed.Category
		feedTypes[feed.ID] = GetFeedType(&feed)
		feedIsImageMode[feed.ID] = feed.IsImageMode

		// Build tag names list for this feed
		tags, _ := h.DB.GetFeedTags(feed.ID)
		tagNames := make([]string, len(tags))
		for i, tag := range tags {
			tagNames[i] = tag.Name
		}
		feedTags[feed.ID] = tagNames
	}

	// Apply filter conditions
	if len(req.Conditions) > 0 {
		var filteredArticles []models.Article
		for _, article := range articles {
			if evaluateArticleConditions(article, req.Conditions, feedCategories, feedTypes, feedIsImageMode, feedTags) {
				filteredArticles = append(filteredArticles, article)
			}
		}
		articles = filteredArticles
	}

	// Apply pagination
	total := len(articles)
	offset := (page - 1) * limit
	end := offset + limit

	// Handle edge cases for pagination
	var paginatedArticles []models.Article
	if offset >= total {
		// No more articles to show
		paginatedArticles = []models.Article{}
	} else {
		if end > total {
			end = total
		}
		paginatedArticles = articles[offset:end]
	}

	hasMore := end < total

	response := FilterResponse{
		Articles: paginatedArticles,
		Total:    total,
		Page:     page,
		Limit:    limit,
		HasMore:  hasMore,
	}

	json.NewEncoder(w).Encode(response)
}

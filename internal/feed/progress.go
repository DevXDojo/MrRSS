package feed

// Progress tracks the state of feed fetching operations
type Progress struct {
	IsRunning bool             `json:"is_running"`
	Errors    map[int64]string `json:"errors,omitempty"` // Map of feed ID to error message
}

// ProgressWithStats extends Progress with runtime statistics
type ProgressWithStats struct {
	Progress
	PoolTaskCount     int `json:"pool_task_count"`     // Tasks in pool
	ArticleClickCount int `json:"article_click_count"` // Article click triggered tasks
	QueueTaskCount    int `json:"queue_task_count"`    // Tasks in queue
}

// GetProgress returns the current progress of the feed fetching operation
func (f *Fetcher) GetProgress() Progress {
	// Delegate to task manager
	return f.taskManager.GetProgress()
}

// GetProgressWithStats returns the current progress with statistics
func (f *Fetcher) GetProgressWithStats() ProgressWithStats {
	progress := f.GetProgress()
	stats := f.taskManager.GetStats()

	return ProgressWithStats{
		Progress:          progress,
		PoolTaskCount:     stats.PoolTaskCount,
		ArticleClickCount: stats.ArticleClickCount,
		QueueTaskCount:    stats.QueueTaskCount,
	}
}

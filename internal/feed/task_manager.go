package feed

import (
	"MrRSS/internal/models"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TaskReason represents the reason why a task was created
type TaskReason int

const (
	TaskReasonManualAdd       TaskReason = iota // Manually added/edited feed
	TaskReasonManualRefresh                     // Right-click refresh
	TaskReasonScheduledCustom                   // Scheduled refresh with custom interval
	TaskReasonScheduledGlobal                   // Global refresh
	TaskReasonArticleClick                      // Article content missing
)

// RefreshTask represents a single feed refresh task
type RefreshTask struct {
	Feed      models.Feed
	Reason    TaskReason
	CreatedAt time.Time
}

// TaskManager manages the task queue and pool for feed refreshing
type TaskManager struct {
	fetcher *Fetcher

	// Double-ended queue for pending tasks
	queue      []int64 // Feed IDs only for efficient storage
	queueMutex sync.RWMutex

	// Task pool for active tasks (limited capacity)
	pool      map[int64]*RefreshTask
	poolMutex sync.RWMutex

	// Pool configuration
	poolCapacity int
	poolSem      chan struct{} // Semaphore for pool capacity

	// State tracking
	isRunning  bool
	isStopped  bool
	stateMutex sync.RWMutex
	wg         sync.WaitGroup
	stopChan   chan struct{}

	// Progress and statistics
	progress      Progress
	progressMutex sync.Mutex
	stats         TaskStats
	statsMutex    sync.RWMutex

	// Task logging
	logFile    *os.File
	logMutex   sync.Mutex
	logEnabled bool
}

// TaskStats represents runtime statistics
type TaskStats struct {
	PoolTaskCount     int // Tasks currently in pool
	ArticleClickCount int // Article click triggered tasks
	QueueTaskCount    int // Tasks in queue
}

// NewTaskManager creates a new task manager
func NewTaskManager(fetcher *Fetcher, poolCapacity int) *TaskManager {
	if poolCapacity < 1 {
		poolCapacity = 5 // Default capacity
	}

	tm := &TaskManager{
		fetcher:      fetcher,
		queue:        make([]int64, 0),
		pool:         make(map[int64]*RefreshTask),
		poolCapacity: poolCapacity,
		poolSem:      make(chan struct{}, poolCapacity),
		stopChan:     make(chan struct{}),
	}

	// Initialize task log file
	tm.initTaskLog()

	return tm
}

// SetPoolCapacity updates the pool capacity and adjusts the semaphore channel
func (tm *TaskManager) SetPoolCapacity(capacity int) {
	if capacity < 1 {
		capacity = 5
	}

	tm.poolMutex.Lock()
	tm.poolCapacity = capacity
	tm.poolMutex.Unlock()

	log.Printf("Task manager pool capacity updated to %d", capacity)
}

// Start starts the task manager
func (tm *TaskManager) Start() {
	tm.stateMutex.Lock()
	defer tm.stateMutex.Unlock()

	if tm.isRunning {
		return
	}

	tm.isRunning = true
	tm.isStopped = false

	log.Printf("Task manager started with pool capacity: %d", tm.poolCapacity)
}

// Stop stops the task manager and waits for all tasks to complete
func (tm *TaskManager) Stop() {
	tm.stateMutex.Lock()
	if tm.isStopped {
		tm.stateMutex.Unlock()
		return
	}
	tm.isStopped = true
	isRunning := tm.isRunning
	tm.stateMutex.Unlock()

	if !isRunning {
		return
	}

	log.Println("Stopping task manager...")

	// Signal stop
	close(tm.stopChan)

	// Wait for all workers to complete
	tm.wg.Wait()

	// Clear state
	tm.queueMutex.Lock()
	tm.queue = make([]int64, 0)
	tm.queueMutex.Unlock()

	log.Println("Task manager stopped")
}

// MarkRunning marks the progress as running
func (tm *TaskManager) MarkRunning() {
	tm.progressMutex.Lock()
	defer tm.progressMutex.Unlock()

	if !tm.progress.IsRunning {
		tm.progress.IsRunning = true
		tm.progress.Errors = make(map[int64]string)
	}
}

// MarkCompleted marks the progress as completed
func (tm *TaskManager) MarkCompleted() {
	tm.progressMutex.Lock()
	defer tm.progressMutex.Unlock()

	tm.progress.IsRunning = false
	log.Println("Progress marked as completed")
}

// AddToQueueHead adds a task to the queue head (highest priority)
// Used for: manual add, manual refresh
func (tm *TaskManager) AddToQueueHead(ctx context.Context, feed models.Feed, reason TaskReason) {
	// Skip FreshRSS feeds - they are refreshed via sync, not standard refresh
	if feed.IsFreshRSSSource {
		log.Printf("Skipping FreshRSS feed %s (refreshed via sync only)", feed.Title)
		return
	}

	tm.stateMutex.RLock()
	isStopped := tm.isStopped
	tm.stateMutex.RUnlock()

	if isStopped {
		log.Println("Task manager is stopped, ignoring task")
		return
	}

	// Mark progress as running
	tm.progressMutex.Lock()
	if !tm.progress.IsRunning {
		tm.progress.IsRunning = true
		tm.progress.Errors = make(map[int64]string)
	}
	tm.progressMutex.Unlock()

	// Remove existing task from queue if present
	tm.queueMutex.Lock()
	removed := removeFromQueue(&tm.queue, feed.ID)

	// Check if already in pool
	tm.poolMutex.RLock()
	inPool := tm.pool[feed.ID] != nil
	tm.poolMutex.RUnlock()

	// Only add if not in pool
	var added bool
	if !inPool {
		// Add to queue head
		tm.queue = append([]int64{feed.ID}, tm.queue...)
		added = true
	}

	tm.queueMutex.Unlock()

	// Log operation after releasing lock to avoid deadlock
	if added {
		if removed {
			log.Printf("Moved feed %s to queue head (reason: %d)", feed.Title, reason)
		} else {
			log.Printf("Added feed %s to queue head (reason: %d)", feed.Title, reason)
		}
		tm.logOperation("AF", feed.Title)
	} else {
		log.Printf("Feed %s already in pool, ignoring (reason: %d)", feed.Title, reason)
		return
	}

	// Update stats
	tm.updateStats()

	// Trigger processing
	go tm.processQueue(ctx)
}

// AddToQueueTail adds a task to the queue tail (lowest priority)
// Used for: scheduled refresh with custom interval
func (tm *TaskManager) AddToQueueTail(ctx context.Context, feed models.Feed, reason TaskReason) {
	// Skip FreshRSS feeds - they are refreshed via sync, not standard refresh
	if feed.IsFreshRSSSource {
		log.Printf("Skipping FreshRSS feed %s (refreshed via sync only)", feed.Title)
		return
	}

	tm.stateMutex.RLock()
	isStopped := tm.isStopped
	tm.stateMutex.RUnlock()

	if isStopped {
		log.Println("Task manager is stopped, ignoring task")
		return
	}

	// Mark progress as running
	tm.progressMutex.Lock()
	if !tm.progress.IsRunning {
		tm.progress.IsRunning = true
		tm.progress.Errors = make(map[int64]string)
	}
	tm.progressMutex.Unlock()

	// Check if already in queue or pool
	tm.queueMutex.Lock()
	tm.poolMutex.RLock()

	inQueue := containsInQueue(tm.queue, feed.ID)
	inPool := tm.pool[feed.ID] != nil

	tm.poolMutex.RUnlock()

	// Only add if not in queue and not in pool
	var added bool
	if !inQueue && !inPool {
		tm.queue = append(tm.queue, feed.ID)
		added = true
	}

	tm.queueMutex.Unlock()

	// Log operation after releasing lock to avoid deadlock
	if added {
		log.Printf("Added feed %s to queue tail (reason: %d)", feed.Title, reason)
		tm.logOperation("AR", feed.Title)
	} else {
		if inQueue {
			log.Printf("Feed %s already in queue, ignoring (reason: %d)", feed.Title, reason)
		} else {
			log.Printf("Feed %s already in pool, ignoring (reason: %d)", feed.Title, reason)
		}
		return
	}

	// Update stats
	tm.updateStats()

	// Trigger processing
	go tm.processQueue(ctx)
}

// AddGlobalRefresh adds multiple feeds to the queue tail for global refresh
// Used for: scheduled global refresh
func (tm *TaskManager) AddGlobalRefresh(ctx context.Context, feeds []models.Feed) {
	tm.stateMutex.RLock()
	isStopped := tm.isStopped
	tm.stateMutex.RUnlock()

	if isStopped {
		return
	}

	if len(feeds) == 0 {
		return
	}

	// Filter out FreshRSS feeds - they are refreshed via sync, not standard refresh
	filteredFeeds := make([]models.Feed, 0, len(feeds))
	skippedCount := 0
	for _, feed := range feeds {
		if feed.IsFreshRSSSource {
			skippedCount++
		} else {
			filteredFeeds = append(filteredFeeds, feed)
		}
	}
	if skippedCount > 0 {
		log.Printf("Filtered out %d FreshRSS feeds from global refresh (refreshed via sync only)", skippedCount)
	}
	feeds = filteredFeeds

	if len(feeds) == 0 {
		return
	}

	// Shuffle feeds to randomize refresh order
	rand.Shuffle(len(feeds), func(i, j int) {
		feeds[i], feeds[j] = feeds[j], feeds[i]
	})

	// Mark progress as running and clear previous errors
	tm.progressMutex.Lock()
	if !tm.progress.IsRunning {
		tm.progress.IsRunning = true
	}
	// Clear previous errors on new global refresh
	tm.progress.Errors = make(map[int64]string)
	tm.progressMutex.Unlock()

	// Update last global refresh time when global refresh starts
	newUpdateTime := time.Now().Format(time.RFC3339)
	log.Printf("Global refresh started, updating last_global_refresh to: %s", newUpdateTime)
	if err := tm.fetcher.db.SetSetting("last_global_refresh", newUpdateTime); err != nil {
		log.Printf("ERROR: Failed to update last_global_refresh: %v", err)
	}

	// Track global refresh in statistics
	if err := tm.fetcher.db.IncrementStat("feed_refresh"); err != nil {
		log.Printf("ERROR: Failed to track feed refresh: %v", err)
	}

	// Clear all feed error marks in database
	if err := tm.fetcher.db.ClearAllFeedErrors(); err != nil {
		log.Printf("Failed to clear all feed errors: %v", err)
	}

	// Add feeds to queue tail with deduplication
	tm.queueMutex.Lock()
	tm.poolMutex.RLock()

	existingFeedIDs := make(map[int64]bool)
	for _, feedID := range tm.queue {
		existingFeedIDs[feedID] = true
	}
	for feedID := range tm.pool {
		existingFeedIDs[feedID] = true
	}

	tm.poolMutex.RUnlock()

	addedCount := 0
	addedFeeds := make([]models.Feed, 0, len(feeds))

	for _, feed := range feeds {
		if !existingFeedIDs[feed.ID] {
			tm.queue = append(tm.queue, feed.ID)
			existingFeedIDs[feed.ID] = true
			addedCount++
			addedFeeds = append(addedFeeds, feed)
		}
	}

	tm.queueMutex.Unlock()

	// Log operations after releasing locks to avoid deadlock
	for _, feed := range addedFeeds {
		log.Printf("Added feed %s to queue tail (global refresh)", feed.Title)
		tm.logOperation("AR", feed.Title)
	}

	log.Printf("Added %d feeds to queue tail for global refresh", addedCount)

	// Update stats
	tm.updateStats()

	// Trigger processing
	go tm.processQueue(ctx)
}

// ExecuteImmediately executes a task immediately, bypassing queue and pool
// Used for: article click triggered refresh
// Returns a function that should be called when the task completes
func (tm *TaskManager) ExecuteImmediately(ctx context.Context, feed models.Feed) func() {
	tm.stateMutex.RLock()
	isStopped := tm.isStopped
	tm.stateMutex.RUnlock()

	if isStopped {
		log.Println("Task manager is stopped, ignoring immediate task")
		return func() {}
	}

	// Remove from queue if present
	tm.queueMutex.Lock()
	removedFromQueue := removeFromQueue(&tm.queue, feed.ID)
	tm.queueMutex.Unlock()

	// Remove from pool if present
	var removedTask *RefreshTask
	tm.poolMutex.Lock()
	if task := tm.pool[feed.ID]; task != nil {
		removedTask = task
		delete(tm.pool, feed.ID)
	}
	tm.poolMutex.Unlock()

	if removedFromQueue {
		log.Printf("Removed feed %s from queue for immediate execution", feed.Title)
	}
	if removedTask != nil {
		log.Printf("Removed feed %s from pool for immediate execution", feed.Title)
	}

	// Create task
	task := &RefreshTask{
		Feed:      feed,
		Reason:    TaskReasonArticleClick,
		CreatedAt: time.Now(),
	}

	// Update stats (increment article click count)
	tm.statsMutex.Lock()
	tm.stats.ArticleClickCount++
	tm.statsMutex.Unlock()

	log.Printf("Executing feed %s immediately (article click)", feed.Title)

	// Start worker goroutine
	tm.wg.Add(1)
	go func() {
		defer func() {
			tm.wg.Done()

			// Update stats (decrement article click count)
			tm.statsMutex.Lock()
			tm.stats.ArticleClickCount--
			tm.statsMutex.Unlock()

			// Check if all tasks completed
			tm.checkCompletion()
		}()

		// Execute with timeout and retry
		var err error
		var success bool

		// Get retry timeout from settings
		retryTimeout := tm.getRetryTimeout()

		// First attempt: 10 second timeout
		ctx1, cancel1 := context.WithTimeout(ctx, 10*time.Second)
		defer cancel1()

		err = tm.fetcher.fetchFeedWithContext(ctx1, task.Feed)
		if err == nil {
			success = true
			log.Printf("Successfully fetched feed: %s (immediate, first attempt)", task.Feed.Title)
		}

		// Second attempt: use configured retry timeout if first attempt failed
		if !success && err != nil {
			log.Printf("First attempt failed for %s: %v, retrying with %v timeout", task.Feed.Title, err, retryTimeout)

			ctx2, cancel2 := context.WithTimeout(ctx, retryTimeout)
			defer cancel2()

			err = tm.fetcher.fetchFeedWithContext(ctx2, task.Feed)
			if err == nil {
				success = true
				log.Printf("Successfully fetched feed: %s (immediate, second attempt)", task.Feed.Title)
			}
		}

		// Handle result
		if err != nil {
			log.Printf("Failed to fetch feed %s (immediate): %v", task.Feed.Title, err)
			tm.fetcher.db.UpdateFeedError(task.Feed.ID, err.Error())
			tm.fetcher.db.UpdateFeedLastUpdated(task.Feed.ID)

			tm.progressMutex.Lock()
			if tm.progress.Errors == nil {
				tm.progress.Errors = make(map[int64]string)
			}
			tm.progress.Errors[task.Feed.ID] = err.Error()
			tm.progressMutex.Unlock()
		} else {
			tm.fetcher.db.UpdateFeedError(task.Feed.ID, "")
			tm.fetcher.db.UpdateFeedLastUpdated(task.Feed.ID)
		}
	}()

	// Return completion callback
	return func() {
		// Task already handled in defer
	}
}

// processQueue processes tasks from the queue
func (tm *TaskManager) processQueue(ctx context.Context) {
	for {
		// Check if stopped
		select {
		case <-tm.stopChan:
			return
		case <-ctx.Done():
			return
		default:
		}

		// Check if we can start a new task
		tm.queueMutex.Lock()
		tm.poolMutex.Lock()

		// Get next task from queue
		var feedID int64
		if len(tm.queue) > 0 && len(tm.pool) < tm.poolCapacity {
			feedID = tm.queue[0]
			tm.queue = tm.queue[1:]
		}

		tm.poolMutex.Unlock()
		tm.queueMutex.Unlock()

		if feedID == 0 {
			// No task available or pool is full
			tm.checkCompletion()
			return
		}

		// Get feed from database
		feed, err := tm.fetcher.db.GetFeedByID(feedID)
		if err != nil {
			log.Printf("Error getting feed %d: %v", feedID, err)
			continue
		}

		// Create task
		task := &RefreshTask{
			Feed:      *feed,
			Reason:    TaskReasonScheduledGlobal, // Default reason
			CreatedAt: time.Now(),
		}

		// Acquire semaphore FIRST (this will block if pool is at capacity)
		// This prevents tasks from being added to pool without a worker
		tm.poolSem <- struct{}{}

		// Add to pool AFTER acquiring semaphore
		tm.poolMutex.Lock()
		tm.pool[feedID] = task
		tm.poolMutex.Unlock()

		// Log move to pool
		tm.logOperation("MV", task.Feed.Title)

		// Update stats
		tm.updateStats()

		// Start worker goroutine
		tm.wg.Add(1)
		go tm.processTask(ctx, task)
	}
}

// processTask processes a single task with timeout and retry logic
func (tm *TaskManager) processTask(ctx context.Context, task *RefreshTask) {
	defer func() {
		// Release semaphore
		<-tm.poolSem
		tm.wg.Done()

		// Remove from pool
		tm.poolMutex.Lock()
		delete(tm.pool, task.Feed.ID)
		tm.poolMutex.Unlock()

		// Update stats
		tm.updateStats()

		// Check if this was the last task
		tm.checkCompletion()

		// Try to process next task
		tm.processQueue(ctx)
	}()

	log.Printf("Processing feed: %s (reason: %d)", task.Feed.Title, task.Reason)

	// Try fetching with timeout and retry
	var err error
	var success bool

	// Get retry timeout from settings
	retryTimeout := tm.getRetryTimeout()

	// First attempt: 60 second timeout (increased from 10s for large feeds)
	// Many feeds have 100+ articles, and processing can take time
	ctx1, cancel1 := context.WithTimeout(ctx, 60*time.Second)
	defer cancel1()

	log.Printf("Starting first attempt to fetch feed: %s (timeout: 60s)", task.Feed.Title)
	err = tm.fetcher.fetchFeedWithContext(ctx1, task.Feed)
	if err == nil {
		success = true
		log.Printf("Successfully fetched feed: %s (first attempt)", task.Feed.Title)
	}

	// Second attempt: use configured retry timeout if first attempt failed
	if !success && err != nil {
		log.Printf("First attempt failed for %s: %v, retrying with %v timeout", task.Feed.Title, err, retryTimeout)
		tm.logOperation("RT", task.Feed.Title)

		ctx2, cancel2 := context.WithTimeout(ctx, retryTimeout)
		defer cancel2()

		err = tm.fetcher.fetchFeedWithContext(ctx2, task.Feed)
		if err == nil {
			success = true
			log.Printf("Successfully fetched feed: %s (second attempt)", task.Feed.Title)
		}
	}

	// Handle result
	if err != nil {
		log.Printf("Failed to fetch feed %s after retry: %v", task.Feed.Title, err)
		tm.logOperation("FL", task.Feed.Title)

		// Update feed error and last_updated in database
		tm.fetcher.db.UpdateFeedError(task.Feed.ID, err.Error())
		tm.fetcher.db.UpdateFeedLastUpdated(task.Feed.ID)

		// Add to progress errors
		tm.progressMutex.Lock()
		if tm.progress.Errors == nil {
			tm.progress.Errors = make(map[int64]string)
		}
		tm.progress.Errors[task.Feed.ID] = err.Error()
		tm.progressMutex.Unlock()
	} else {
		tm.logOperation("SC", task.Feed.Title)
		// Clear error on success and update last_updated
		tm.fetcher.db.UpdateFeedError(task.Feed.ID, "")
		tm.fetcher.db.UpdateFeedLastUpdated(task.Feed.ID)
	}
}

// checkCompletion checks if all tasks are completed and triggers cleanup if needed
func (tm *TaskManager) checkCompletion() {
	tm.queueMutex.RLock()
	queueLen := len(tm.queue)
	tm.queueMutex.RUnlock()

	tm.poolMutex.RLock()
	poolLen := len(tm.pool)
	tm.poolMutex.RUnlock()

	tm.statsMutex.RLock()
	articleClickCount := tm.stats.ArticleClickCount
	tm.statsMutex.RUnlock()

	tm.progressMutex.Lock()
	defer tm.progressMutex.Unlock()

	if queueLen == 0 && poolLen == 0 && articleClickCount == 0 && tm.progress.IsRunning {
		// All tasks completed
		tm.progress.IsRunning = false

		log.Println("All tasks completed")

		// Trigger cleanup through cleanup manager
		tm.fetcher.cleanupManager.RequestCleanup()
	}
}

// GetProgress returns the current progress
func (tm *TaskManager) GetProgress() Progress {
	tm.progressMutex.Lock()
	defer tm.progressMutex.Unlock()

	return Progress{
		IsRunning: tm.progress.IsRunning,
		Errors:    tm.progress.Errors,
	}
}

// GetStats returns the current statistics
func (tm *TaskManager) GetStats() TaskStats {
	tm.statsMutex.RLock()
	defer tm.statsMutex.RUnlock()

	// Also get current pool and queue counts
	tm.poolMutex.RLock()
	poolLen := len(tm.pool)
	tm.poolMutex.RUnlock()

	tm.queueMutex.RLock()
	queueLen := len(tm.queue)
	tm.queueMutex.RUnlock()

	stats := TaskStats{
		PoolTaskCount:     poolLen,
		ArticleClickCount: tm.stats.ArticleClickCount,
		QueueTaskCount:    queueLen,
	}

	return stats
}

// GetActiveFeedNames returns the names of feeds currently being processed, sorted alphabetically
func (tm *TaskManager) GetActiveFeedNames() []string {
	tm.poolMutex.RLock()
	defer tm.poolMutex.RUnlock()

	names := make([]string, 0, len(tm.pool))
	for _, task := range tm.pool {
		names = append(names, task.Feed.Title)
	}

	// Sort alphabetically
	sortStrings(names)

	return names
}

// GetQueuedFeedNames returns the names of feeds in the queue, sorted alphabetically
func (tm *TaskManager) GetQueuedFeedNames() []string {
	tm.queueMutex.RLock()
	defer tm.queueMutex.RUnlock()

	// Need to fetch feed titles from database
	feedIDs := make([]int64, len(tm.queue))
	copy(feedIDs, tm.queue)

	names := make([]string, 0, len(feedIDs))
	for _, feedID := range feedIDs {
		feed, err := tm.fetcher.db.GetFeedByID(feedID)
		if err == nil {
			names = append(names, feed.Title)
		}
	}

	// Sort alphabetically
	sortStrings(names)

	return names
}

// GetPoolTasks returns detailed information about tasks currently in the pool
func (tm *TaskManager) GetPoolTasks() []PoolTaskInfo {
	tm.poolMutex.RLock()
	defer tm.poolMutex.RUnlock()

	tasks := make([]PoolTaskInfo, 0, len(tm.pool))
	for _, task := range tm.pool {
		tasks = append(tasks, PoolTaskInfo{
			FeedID:    task.Feed.ID,
			FeedTitle: task.Feed.Title,
			Reason:    task.Reason,
			CreatedAt: task.CreatedAt,
		})
	}

	// Sort by creation time (oldest first)
	for i := 0; i < len(tasks)-1; i++ {
		for j := i + 1; j < len(tasks); j++ {
			if tasks[i].CreatedAt.After(tasks[j].CreatedAt) {
				tasks[i], tasks[j] = tasks[j], tasks[i]
			}
		}
	}

	return tasks
}

// GetQueueTasks returns detailed information about tasks in the queue (up to limit)
// Returns tasks in queue order (head first)
func (tm *TaskManager) GetQueueTasks(limit int) []QueueTaskInfo {
	tm.queueMutex.RLock()
	defer tm.queueMutex.RUnlock()

	// Determine how many tasks to return
	count := len(tm.queue)
	if limit > 0 && count > limit {
		count = limit
	}

	tasks := make([]QueueTaskInfo, 0, count)
	for i := 0; i < count; i++ {
		feedID := tm.queue[i]
		feed, err := tm.fetcher.db.GetFeedByID(feedID)
		if err == nil {
			tasks = append(tasks, QueueTaskInfo{
				FeedID:    feed.ID,
				FeedTitle: feed.Title,
				Position:  i,
			})
		}
	}

	return tasks
}

// PoolTaskInfo contains information about a task in the pool
type PoolTaskInfo struct {
	FeedID    int64      `json:"feed_id"`
	FeedTitle string     `json:"feed_title"`
	Reason    TaskReason `json:"reason"`
	CreatedAt time.Time  `json:"created_at"`
}

// QueueTaskInfo contains information about a task in the queue
type QueueTaskInfo struct {
	FeedID    int64  `json:"feed_id"`
	FeedTitle string `json:"feed_title"`
	Position  int    `json:"position"`
}

// IsRunning returns true if the task manager is running
func (tm *TaskManager) IsRunning() bool {
	tm.progressMutex.Lock()
	defer tm.progressMutex.Unlock()
	return tm.progress.IsRunning
}

// Wait waits for all tasks to complete
func (tm *TaskManager) Wait(timeout time.Duration) bool {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			return false // Timeout
		case <-ticker.C:
			if !tm.IsRunning() {
				return true
			}
		}
	}
}

// ClearQueue clears all pending tasks from the queue
func (tm *TaskManager) ClearQueue() {
	tm.queueMutex.Lock()
	defer tm.queueMutex.Unlock()

	tm.queue = make([]int64, 0)

	log.Println("Queue cleared")
}

// updateStats updates the statistics
func (tm *TaskManager) updateStats() {
	tm.poolMutex.RLock()
	poolLen := len(tm.pool)
	tm.poolMutex.RUnlock()

	tm.queueMutex.RLock()
	queueLen := len(tm.queue)
	tm.queueMutex.RUnlock()

	tm.statsMutex.Lock()
	tm.stats.PoolTaskCount = poolLen
	tm.stats.QueueTaskCount = queueLen
	tm.statsMutex.Unlock()
}

// Helper functions

// removeFromQueue removes a feed ID from the queue and returns true if it was present
func removeFromQueue(queue *[]int64, feedID int64) bool {
	for i, id := range *queue {
		if id == feedID {
			*queue = append((*queue)[:i], (*queue)[i+1:]...)
			return true
		}
	}
	return false
}

// containsInQueue checks if a feed ID is in the queue
func containsInQueue(queue []int64, feedID int64) bool {
	for _, id := range queue {
		if id == feedID {
			return true
		}
	}
	return false
}

// sortStrings sorts a slice of strings alphabetically
func sortStrings(slice []string) {
	for i := 0; i < len(slice)-1; i++ {
		for j := i + 1; j < len(slice); j++ {
			if slice[i] > slice[j] {
				slice[i], slice[j] = slice[j], slice[i]
			}
		}
	}
}

// parseInt parses a string to int
func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

// getRetryTimeout retrieves the retry timeout from settings
// Returns the configured timeout in seconds (default 60 seconds)
func (tm *TaskManager) getRetryTimeout() time.Duration {
	retryTimeoutStr, err := tm.fetcher.db.GetSetting("retry_timeout_seconds")
	if err != nil || retryTimeoutStr == "" {
		return 60 * time.Second // Default to 60 seconds
	}

	retryTimeout, err := parseInt(retryTimeoutStr)
	if err != nil || retryTimeout <= 0 {
		return 60 * time.Second // Default to 60 seconds if invalid
	}

	return time.Duration(retryTimeout) * time.Second
}

// initTaskLog initializes the task log file
func (tm *TaskManager) initTaskLog() {
	// Get data directory
	dataDir, err := tm.fetcher.getDataDir()
	if err != nil {
		log.Printf("Failed to get data directory for task log: %v", err)
		return
	}

	// Create logs directory
	logDir := filepath.Join(dataDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create logs directory: %v", err)
		return
	}

	// Open log file with truncate flag to clear previous logs
	logPath := filepath.Join(logDir, "tasks.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Printf("Failed to open task log file: %v", err)
		return
	}

	tm.logFile = logFile
	tm.logEnabled = true

	log.Printf("Task log initialized: %s", logPath)
}

// logOperation logs a task operation with the specified format
// Format: AF/AR/MV/RT/SC/FL n/m name
// AF = Add to Front (queue head), AR = Add to Rear (queue tail)
// MV = Move to Pool, RT = Retry, SC = Success, FL = Failure
// n = pool task count, m = queue task count
func (tm *TaskManager) logOperation(operation string, feedName string) {
	if !tm.logEnabled || tm.logFile == nil {
		return
	}

	tm.poolMutex.RLock()
	poolLen := len(tm.pool)
	tm.poolMutex.RUnlock()

	tm.queueMutex.RLock()
	queueLen := len(tm.queue)
	tm.queueMutex.RUnlock()

	logEntry := fmt.Sprintf("%s %d/%d %s\n", operation, poolLen, queueLen, feedName)

	tm.logMutex.Lock()
	defer tm.logMutex.Unlock()

	if _, err := tm.logFile.WriteString(logEntry); err != nil {
		log.Printf("Failed to write to task log: %v", err)
	}
}

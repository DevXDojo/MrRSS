package feed

import (
	"log"
	"sync"
	"time"
)

// CleanupManager manages automatic cleanup with retry mechanism
type CleanupManager struct {
	fetcher *Fetcher

	// State tracking
	isRunning      bool
	cleanupRunning bool
	mu             sync.RWMutex

	// Cleanup request tracking
	pendingCleanup   bool
	pendingCleanupMu sync.Mutex

	// Retry mechanism
	retryInterval time.Duration // 10 minutes
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewCleanupManager creates a new cleanup manager
func NewCleanupManager(fetcher *Fetcher) *CleanupManager {
	return &CleanupManager{
		fetcher:        fetcher,
		retryInterval:  10 * time.Minute,
		stopChan:       make(chan struct{}),
		pendingCleanup: false,
	}
}

// Start starts the cleanup manager
func (cm *CleanupManager) Start() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.isRunning {
		return
	}

	cm.stopChan = make(chan struct{})
	cm.isRunning = true

	// Start retry goroutine
	cm.wg.Add(1)
	go cm.retryLoop()

	log.Println("Cleanup manager started")
}

// Stop stops the cleanup manager
func (cm *CleanupManager) Stop() {
	cm.mu.Lock()
	if !cm.isRunning {
		cm.mu.Unlock()
		return
	}
	cm.isRunning = false
	close(cm.stopChan)
	cm.mu.Unlock()
	cm.wg.Wait()
	log.Println("Cleanup manager stopped")
}

// RequestCleanup requests a cleanup operation
// If cleanup is blocked (tasks running), it will be retried every 10 minutes
func (cm *CleanupManager) RequestCleanup() {
	cm.pendingCleanupMu.Lock()
	cm.pendingCleanup = true
	cm.pendingCleanupMu.Unlock()

	// Try to execute immediately
	cm.tryCleanup()
}

// RequestManualCleanup clears all article contents immediately
// This is for manual cleanup triggered by user
func (cm *CleanupManager) RequestManualCleanup() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if !cm.isRunning {
		return
	}
	// Manual cleanup clears all content regardless of tasks
	cm.wg.Add(1)
	go func() {
		defer cm.wg.Done()

		log.Println("Executing manual cleanup (clearing all article contents)")

		count, err := cm.fetcher.db.CleanupAllArticleContents()
		if err != nil {
			log.Printf("Manual cleanup error: %v", err)
		} else {
			log.Printf("Manual cleanup completed: cleared %d article contents", count)
		}
	}()
}

// tryCleanup attempts to execute cleanup if conditions are met
func (cm *CleanupManager) tryCleanup() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if !cm.isRunning || cm.cleanupRunning {
		return
	}
	// Check if we can cleanup (no tasks running)
	if !cm.canCleanup() {
		log.Println("Cleanup blocked: tasks are running, will retry later")
		return
	}

	cm.pendingCleanupMu.Lock()
	if !cm.pendingCleanup {
		cm.pendingCleanupMu.Unlock()
		return
	}
	cm.pendingCleanup = false
	cm.pendingCleanupMu.Unlock()

	// Only one cleanup may run at a time.
	cm.cleanupRunning = true
	cm.wg.Add(1)
	go func() {
		defer cm.wg.Done()
		defer func() { cm.mu.Lock(); cm.cleanupRunning = false; cm.mu.Unlock() }()
		cm.executeCleanup()
	}()
}

// canCleanup checks if cleanup can be executed (no tasks running)
func (cm *CleanupManager) canCleanup() bool {
	stats := cm.fetcher.taskManager.GetStats()

	// Check if queue, pool, or article click tasks are running
	if stats.QueueTaskCount > 0 || stats.PoolTaskCount > 0 || stats.ArticleClickCount > 0 {
		return false
	}

	return true
}

// executeCleanup executes the layered cleanup
func (cm *CleanupManager) executeCleanup() {
	log.Println("Starting automatic cleanup...")

	// Use the bounded oldest-first policy shared by other automatic cleanup.
	totalRemoved, err := cm.fetcher.db.CleanupBySize()
	if err != nil {
		log.Printf("Automatic cleanup failed: %v", err)
		return
	}

	if totalRemoved > 0 {
		log.Printf("Automatic cleanup completed: removed %d items", totalRemoved)
	} else {
		log.Println("Automatic cleanup completed: nothing to clean")
	}
}

// getTargetSize returns the target database size in MB
func (cm *CleanupManager) getTargetSize() float64 {
	maxSizeMBStr, _ := cm.fetcher.db.GetSetting("max_cache_size_mb")
	maxSizeMB := 500 // Default
	if maxSizeMBStr != "" {
		if size, err := parseInt(maxSizeMBStr); err == nil && size > 0 {
			maxSizeMB = size
		}
	}
	return float64(maxSizeMB)
}

// retryLoop checks every 10 minutes if pending cleanup can be executed
func (cm *CleanupManager) retryLoop() {
	defer cm.wg.Done()

	ticker := time.NewTicker(cm.retryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cm.stopChan:
			return
		case <-ticker.C:
			cm.pendingCleanupMu.Lock()
			hasPending := cm.pendingCleanup
			cm.pendingCleanupMu.Unlock()

			if hasPending {
				log.Println("Retry: attempting pending cleanup")
				cm.tryCleanup()
			}
		}
	}
}

// CheckSizeAndCleanup checks database size and triggers cleanup if needed
func (cm *CleanupManager) CheckSizeAndCleanup() {
	maxSizeMB := cm.getTargetSize()

	currentSizeMB, err := cm.fetcher.db.GetDatabaseSizeMB()
	if err != nil {
		log.Printf("Error checking database size: %v", err)
		return
	}

	if currentSizeMB > maxSizeMB {
		log.Printf("Database size %.2f MB exceeds limit %.2f MB, triggering cleanup", currentSizeMB, maxSizeMB)
		cm.RequestCleanup()
	}
}

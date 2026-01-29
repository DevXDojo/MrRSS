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
	isRunning bool
	mu        sync.RWMutex

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

	cm.isRunning = true

	// Start retry goroutine
	cm.wg.Add(1)
	go cm.retryLoop()

	log.Println("Cleanup manager started")
}

// Stop stops the cleanup manager
func (cm *CleanupManager) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.isRunning {
		return
	}

	close(cm.stopChan)
	cm.wg.Wait()

	cm.isRunning = false
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

	// Execute cleanup
	cm.wg.Add(1)
	go func() {
		defer cm.wg.Done()
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

	maxSizeMB := cm.getTargetSize()

	// Execute layered cleanup with 80% target
	totalRemoved := cm.layeredCleanup(maxSizeMB * 0.8)

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

// layeredCleanup executes cleanup in layers until target size is reached
// Cleanup order:
// 1. Old article contents
// 2. Medium article contents
// 3. Old article metadata
// 4. New article contents
// 5. Latest article contents
// 6. Medium article metadata
// Note: New and latest article metadata are never cleaned
func (cm *CleanupManager) layeredCleanup(targetSizeMB float64) int64 {
	totalRemoved := int64(0)

	// Get current size
	currentSizeMB, _ := cm.fetcher.db.GetDatabaseSizeMB()

	if currentSizeMB <= targetSizeMB {
		return 0
	}

	log.Printf("Current size: %.2f MB, Target: %.2f MB", currentSizeMB, targetSizeMB)

	// Layer 1: Old article contents (7+ days old)
	if currentSizeMB > targetSizeMB {
		count, err := cm.fetcher.db.CleanupArticleContentsByAge(7)
		if err != nil {
			log.Printf("Layer 1 error: %v", err)
		} else {
			log.Printf("Layer 1: Removed %d old article contents", count)
			totalRemoved += count
			currentSizeMB, _ = cm.fetcher.db.GetDatabaseSizeMB()
		}
	}

	// Layer 2: Medium article contents (3+ days old)
	if currentSizeMB > targetSizeMB {
		count, err := cm.fetcher.db.CleanupArticleContentsByAge(3)
		if err != nil {
			log.Printf("Layer 2 error: %v", err)
		} else {
			log.Printf("Layer 2: Removed %d medium article contents", count)
			totalRemoved += count
			currentSizeMB, _ = cm.fetcher.db.GetDatabaseSizeMB()
		}
	}

	// Layer 3: Old article metadata (read, 30+ days old)
	if currentSizeMB > targetSizeMB {
		count, err := cm.fetcher.db.CleanupOldReadArticles(30)
		if err != nil {
			log.Printf("Layer 3 error: %v", err)
		} else {
			log.Printf("Layer 3: Removed %d old article metadata", count)
			totalRemoved += count
			currentSizeMB, _ = cm.fetcher.db.GetDatabaseSizeMB()
		}
	}

	// Layer 4: New article contents (1+ days old)
	if currentSizeMB > targetSizeMB {
		count, err := cm.fetcher.db.CleanupArticleContentsByAge(1)
		if err != nil {
			log.Printf("Layer 4 error: %v", err)
		} else {
			log.Printf("Layer 4: Removed %d new article contents", count)
			totalRemoved += count
			currentSizeMB, _ = cm.fetcher.db.GetDatabaseSizeMB()
		}
	}

	// Layer 5: Latest article contents (all)
	if currentSizeMB > targetSizeMB {
		count, err := cm.fetcher.db.CleanupAllArticleContents()
		if err != nil {
			log.Printf("Layer 5 error: %v", err)
		} else {
			log.Printf("Layer 5: Removed %d latest article contents", count)
			totalRemoved += count
			currentSizeMB, _ = cm.fetcher.db.GetDatabaseSizeMB()
		}
	}

	// Layer 6: Medium article metadata (unread, 60+ days old, not favorite/read-later)
	if currentSizeMB > targetSizeMB {
		count, err := cm.fetcher.db.CleanupOldUnreadArticles(60)
		if err != nil {
			log.Printf("Layer 6 error: %v", err)
		} else {
			log.Printf("Layer 6: Removed %d medium article metadata", count)
			totalRemoved += count
			_, _ = cm.fetcher.db.GetDatabaseSizeMB()
		}
	}

	// Final size check
	finalSizeMB, _ := cm.fetcher.db.GetDatabaseSizeMB()
	log.Printf("Final size: %.2f MB (target was %.2f MB)", finalSizeMB, targetSizeMB)

	return totalRemoved
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

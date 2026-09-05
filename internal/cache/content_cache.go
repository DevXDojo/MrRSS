// Package cache provides content caching functionality for article content.
package cache

import (
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

// ContentCacheItem represents a cached content item with expiration
type ContentCacheItem struct {
	Content   string
	ExpiresAt time.Time
	SetAt     time.Time // When the item was set
}

// FeedCacheItem represents a cached feed with expiration
type FeedCacheItem struct {
	Feed      *gofeed.Feed
	ExpiresAt time.Time
	SetAt     time.Time // When the item was set
}

// ContentCache provides LRU-style caching for article content
type ContentCache struct {
	mu      sync.RWMutex
	content map[int64]*ContentCacheItem
	feeds   map[int64]*FeedCacheItem // Cache feeds by feedID
	maxSize int
	ttl     time.Duration
}

// NewContentCache creates a new content cache
func NewContentCache(maxSize int, ttl time.Duration) *ContentCache {
	return &ContentCache{
		content: make(map[int64]*ContentCacheItem),
		feeds:   make(map[int64]*FeedCacheItem),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// Get retrieves content from cache if it exists and hasn't expired
func (cc *ContentCache) Get(articleID int64) (string, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	item, exists := cc.content[articleID]
	if !exists {
		return "", false
	}

	// Check if expired
	if time.Now().After(item.ExpiresAt) {
		// Item expired, remove it
		go func() {
			cc.mu.Lock()
			delete(cc.content, articleID)
			cc.mu.Unlock()
		}()
		return "", false
	}

	return item.Content, true
}

// GetFeed retrieves feed from cache if it exists and hasn't expired
func (cc *ContentCache) GetFeed(feedID int64) (*gofeed.Feed, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	item, exists := cc.feeds[feedID]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.ExpiresAt) {
		// Item expired, remove it
		go func() {
			cc.mu.Lock()
			delete(cc.feeds, feedID)
			cc.mu.Unlock()
		}()
		return nil, false
	}

	return item.Feed, true
}

// Set stores content in cache
func (cc *ContentCache) Set(articleID int64, content string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	now := time.Now()

	// If cache is at max capacity, remove oldest item before adding new one
	if len(cc.content) >= cc.maxSize {
		// Find oldest item by set time
		var oldestID int64
		var oldestTime = time.Now() // Initialize to current time

		for id, item := range cc.content {
			if item.SetAt.Before(oldestTime) {
				oldestTime = item.SetAt
				oldestID = id
			}
		}

		if oldestID != 0 {
			delete(cc.content, oldestID)
		}
	}

	cc.content[articleID] = &ContentCacheItem{
		Content:   content,
		ExpiresAt: now.Add(cc.ttl),
		SetAt:     now,
	}
}

// SetFeed stores feed in cache
func (cc *ContentCache) SetFeed(feedID int64, feed *gofeed.Feed) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	now := time.Now()

	// If cache is at max capacity, remove oldest item before adding new one
	if len(cc.feeds) >= cc.maxSize {
		// Find oldest item by set time
		var oldestID int64
		var oldestTime = time.Now() // Initialize to current time

		for id, item := range cc.feeds {
			if item.SetAt.Before(oldestTime) {
				oldestTime = item.SetAt
				oldestID = id
			}
		}

		if oldestID != 0 {
			delete(cc.feeds, oldestID)
		}
	}

	cc.feeds[feedID] = &FeedCacheItem{
		Feed:      feed,
		ExpiresAt: now.Add(cc.ttl),
		SetAt:     now,
	}
}

// Clear removes all cached content
func (cc *ContentCache) Clear() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.content = make(map[int64]*ContentCacheItem)
	cc.feeds = make(map[int64]*FeedCacheItem)
}

// Size returns the current number of cached items
func (cc *ContentCache) Size() int {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return len(cc.content) + len(cc.feeds)
}

// Delete invalidates an article after an explicit reload.
func (cc *ContentCache) Delete(articleID int64) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	delete(cc.content, articleID)
}

/**
 * Global image cache to persist loaded images across component refreshes
 * This prevents images from disappearing when the article list refreshes
 * Simplified to eliminate retry overhead and blocking operations
 */

interface ImageCacheEntry {
  url: string;
  timestamp: number;
}

interface LoadState {
  status: 'loading' | 'loaded' | 'error';
}

class ImageCacheManager {
  private cache = new Map<string, ImageCacheEntry>();
  private loadStates = new Map<string, LoadState>();
  private readonly CACHE_TTL = 24 * 60 * 60 * 1000; // 24 hours

  /**
   * Get the current URL for an image, returning cached version if available
   * Optimized to reduce Map lookups
   */
  getImageUrl(originalUrl: string): string {
    const state = this.loadStates.get(originalUrl);

    // If current load failed but we have a cache, return the cached URL
    if (state?.status === 'error') {
      const cached = this.cache.get(originalUrl);
      if (cached) return cached.url;
    }

    return originalUrl;
  }

  /**
   * Mark an image as successfully loaded and cache it
   */
  markAsLoaded(url: string): void {
    // Cache the successful URL
    this.cache.set(url, {
      url,
      timestamp: Date.now(),
    });

    // Update load state
    this.loadStates.set(url, {
      status: 'loaded',
    });

    // Clean up old entries (debounced)
    this.scheduleCleanup();
  }

  private cleanupScheduled = false;
  private scheduleCleanup(): void {
    if (!this.cleanupScheduled) {
      this.cleanupScheduled = true;
      // Use requestIdleCallback for non-blocking cleanup
      if (typeof window.requestIdleCallback !== 'undefined') {
        window.requestIdleCallback(() => {
          this.cleanUp();
          this.cleanupScheduled = false;
        });
      } else {
        // Fallback to setTimeout
        setTimeout(() => {
          this.cleanUp();
          this.cleanupScheduled = false;
        }, 5000); // Longer delay for cleanup
      }
    }
  }

  /**
   * Handle image load error - marks as permanently failed, no retries
   * Simplified to eliminate all retry overhead
   */
  handleLoadError(url: string): { shouldRetry: boolean } {
    // If we have a cached version, restore from cache immediately
    if (this.cache.has(url)) {
      this.loadStates.set(url, {
        status: 'loaded',
      });
      return { shouldRetry: false };
    }

    // Mark as permanently failed - no retries
    this.loadStates.set(url, {
      status: 'error',
    });

    return { shouldRetry: false };
  }

  /**
   * Get the current load state of an image
   */
  getLoadState(url: string): LoadState | undefined {
    return this.loadStates.get(url);
  }

  /**
   * Check if an image has been cached
   */
  hasCached(url: string): boolean {
    return this.cache.has(url);
  }

  /**
   * Clean up old cache entries
   */
  private cleanUp(): void {
    const now = Date.now();
    const urlsToDelete: string[] = [];

    // Collect expired entries first (faster than deleting during iteration)
    for (const [url, entry] of this.cache.entries()) {
      if (now - entry.timestamp > this.CACHE_TTL) {
        urlsToDelete.push(url);
      }
    }

    // Batch delete
    for (const url of urlsToDelete) {
      this.cache.delete(url);
      this.loadStates.delete(url);
    }
  }

  /**
   * Clear all cache (useful for testing or manual refresh)
   */
  clearAll(): void {
    this.cache.clear();
    this.loadStates.clear();
  }
}

// Export a singleton instance
export const imageCache = new ImageCacheManager();

package database

import (
	"log"
	"strconv"
	"time"
)

const defaultMaxArticlesPerFeed = 15000

// CleanupOldArticles removes articles based on age and status.
// - Articles older than configured days: delete except favorited or read later
// - Read article metadata beyond the per-feed retention limit
// - Also checks database size against max_cache_size_mb setting
func (db *DB) CleanupOldArticles() (int64, error) {
	db.WaitForReady()

	totalDeleted := int64(0)

	// Step 1: Clean up by age (existing logic)
	maxAgeDaysStr, err := db.GetSetting("max_article_age_days")
	maxAgeDays := 30
	if err == nil {
		if days, err := strconv.Atoi(maxAgeDaysStr); err == nil && days > 0 {
			maxAgeDays = days
		}
	}

	cutoffDate := time.Now().AddDate(0, 0, -maxAgeDays)

	// Delete articles older than configured age that are not favorited or in read later
	result, err := db.Exec(`
		DELETE FROM articles
		WHERE published_at < ?
		AND is_favorite = 0
		AND is_read_later = 0
	`, cutoffDate)
	if err != nil {
		return 0, err
	}

	count, _ := result.RowsAffected()
	totalDeleted += count

	// Step 2: Apply per-feed article retention so high-volume feeds do not
	// force low-volume feeds to lose their local history.
	perFeedDeleted, err := db.CleanupReadArticlesOverPerFeedLimit(defaultMaxArticlesPerFeed)
	if err != nil {
		log.Printf("Error during per-feed retention cleanup: %v", err)
	} else {
		totalDeleted += perFeedDeleted
	}

	// Step 3: Check database size and clean up if over limit
	sizeDeleted, err := db.CleanupBySize()
	if err != nil {
		log.Printf("Error during size-based cleanup: %v", err)
	} else {
		totalDeleted += sizeDeleted
	}

	// Also cleanup related caches with the same age limit
	_, _ = db.CleanupTranslationCache(maxAgeDays)
	_, _ = db.CleanupOldArticleContents(maxAgeDays)

	// Reclaim freelist pages
	if totalDeleted > 0 {
		_, _ = db.IncrementalVacuum()
	}

	return totalDeleted, nil
}

// CleanupAllArticleContents removes all cached article contents
func (db *DB) CleanupAllArticleContents() (int64, error) {
	db.WaitForReady()
	result, err := db.Exec(`DELETE FROM article_contents`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// DeleteAllArticles removes ALL articles from the database
// This keeps feeds, settings, and other metadata intact.
// With foreign_keys enabled, ON DELETE CASCADE automatically removes
// associated article_contents, chat_sessions, and chat_messages rows.
func (db *DB) DeleteAllArticles() (int64, error) {
	db.WaitForReady()
	result, err := db.Exec(`DELETE FROM articles`)
	if err != nil {
		return 0, err
	}

	count, _ := result.RowsAffected()
	// Reclaim freelist pages after bulk delete
	if count > 0 {
		_, _ = db.IncrementalVacuum()
	}
	return count, nil
}

// CleanupUnimportantArticles removes all articles except read, favorited, and read later ones.
func (db *DB) CleanupUnimportantArticles() (int64, error) {
	db.WaitForReady()

	result, err := db.Exec(`
		DELETE FROM articles
		WHERE is_read = 0
		AND is_favorite = 0
		AND is_read_later = 0
	`)
	if err != nil {
		return 0, err
	}

	count, _ := result.RowsAffected()

	// Reclaim freelist pages
	if count > 0 {
		_, _ = db.IncrementalVacuum()
	}

	return count, nil
}

// CleanupReadArticlesOverPerFeedLimit removes old read articles above the
// per-feed retention limit while preserving favorites, read-later items, and
// unread metadata. Protected articles may cause a feed to exceed the limit.
func (db *DB) CleanupReadArticlesOverPerFeedLimit(maxArticlesPerFeed int) (int64, error) {
	db.WaitForReady()

	if maxArticlesPerFeed <= 0 {
		return 0, nil
	}

	result, err := db.Exec(`
		WITH ranked_articles AS (
			SELECT
				id,
				ROW_NUMBER() OVER (
					PARTITION BY feed_id
					ORDER BY published_at DESC, id DESC
				) AS feed_rank
			FROM articles
		)
		DELETE FROM articles
		WHERE id IN (
			SELECT articles.id
			FROM articles
			JOIN ranked_articles ON ranked_articles.id = articles.id
			WHERE ranked_articles.feed_rank > ?
			AND articles.is_read = 1
			AND articles.is_favorite = 0
			AND articles.is_read_later = 0
		)
	`, maxArticlesPerFeed)
	if err != nil {
		return 0, err
	}

	count, _ := result.RowsAffected()
	if count > 0 {
		_, _ = db.IncrementalVacuum()
	}
	return count, nil
}

// GetDatabaseSizeMB returns the current database ACTUAL data size in megabytes.
// This excludes freelist pages (pages freed by DELETE but not yet reclaimed),
// so it reflects the real data footprint rather than the on-disk file size.
// Use GetDatabaseFileSizeMB if you need the physical file size instead.
func (db *DB) GetDatabaseSizeMB() (float64, error) {
	db.WaitForReady()

	var pageCount, pageSize, freelistCount int64
	err := db.QueryRow("PRAGMA page_count").Scan(&pageCount)
	if err != nil {
		return 0, err
	}

	err = db.QueryRow("PRAGMA page_size").Scan(&pageSize)
	if err != nil {
		return 0, err
	}

	err = db.QueryRow("PRAGMA freelist_count").Scan(&freelistCount)
	if err != nil {
		return 0, err
	}

	// Actual used pages = total pages - freelist pages
	usedPages := pageCount - freelistCount
	if usedPages < 0 {
		usedPages = 0
	}
	sizeBytes := usedPages * pageSize
	sizeMB := float64(sizeBytes) / (1024 * 1024)

	return sizeMB, nil
}

// GetDatabaseFileSizeMB returns the physical database file size in megabytes,
// including freelist pages that have not yet been reclaimed.
func (db *DB) GetDatabaseFileSizeMB() (float64, error) {
	db.WaitForReady()

	var pageCount, pageSize int64
	err := db.QueryRow("PRAGMA page_count").Scan(&pageCount)
	if err != nil {
		return 0, err
	}

	err = db.QueryRow("PRAGMA page_size").Scan(&pageSize)
	if err != nil {
		return 0, err
	}

	sizeBytes := pageCount * pageSize
	sizeMB := float64(sizeBytes) / (1024 * 1024)

	return sizeMB, nil
}

// IncrementalVacuum reclaims freelist pages, returning the number of pages freed.
// Requires auto_vacuum=INCREMENTAL mode (set during migration).
// If auto_vacuum is not INCREMENTAL, this is a no-op.
func (db *DB) IncrementalVacuum() (int64, error) {
	db.WaitForReady()

	// Check if auto_vacuum is in INCREMENTAL mode (value 2)
	var autoVacuum int64
	err := db.QueryRow("PRAGMA auto_vacuum").Scan(&autoVacuum)
	if err != nil {
		return 0, err
	}
	if autoVacuum != 2 {
		// Not in incremental mode; fall back to full VACUUM if there are many freelist pages
		var freelistCount int64
		if err := db.QueryRow("PRAGMA freelist_count").Scan(&freelistCount); err != nil {
			return 0, err
		}
		if freelistCount > 1000 {
			if _, err := db.Exec("VACUUM"); err != nil {
				return 0, err
			}
			return freelistCount, nil
		}
		return 0, nil
	}

	// In INCREMENTAL mode: reclaim up to N pages (0 = all possible)
	var beforeFreelist int64
	_ = db.QueryRow("PRAGMA freelist_count").Scan(&beforeFreelist)

	// incremental_vacuum(0) reclaims all freelist pages
	if _, err := db.Exec("PRAGMA incremental_vacuum"); err != nil {
		return 0, err
	}

	var afterFreelist int64
	_ = db.QueryRow("PRAGMA freelist_count").Scan(&afterFreelist)

	reclaimed := beforeFreelist - afterFreelist
	if reclaimed < 0 {
		reclaimed = 0
	}
	return reclaimed, nil
}

// ShouldCleanupBeforeSave checks if database is approaching the size limit.
// Returns true if database size is over 80% of max_cache_size_mb.
func (db *DB) ShouldCleanupBeforeSave() (bool, error) {
	db.WaitForReady()

	// Get max cache size from settings (default 500 MB)
	maxSizeMBStr, err := db.GetSetting("max_cache_size_mb")
	maxSizeMB := 500
	if err == nil {
		if size, err := strconv.Atoi(maxSizeMBStr); err == nil && size > 0 {
			maxSizeMB = size
		}
	}

	// Get current database size
	currentSizeMB, err := db.GetDatabaseSizeMB()
	if err != nil {
		return false, err
	}

	// Trigger cleanup if over 80% of limit
	threshold := float64(maxSizeMB) * 0.8
	return currentSizeMB >= threshold, nil
}

// CleanupBySize reduces cached content first to keep database under max_cache_size_mb.
// Article metadata is preserved whenever possible so refreshed feeds do not
// reinsert old read items as new articles after cleanup.
func (db *DB) CleanupBySize() (int64, error) {
	db.WaitForReady()

	// Get max cache size from settings (default 500 MB)
	maxSizeMBStr, err := db.GetSetting("max_cache_size_mb")
	maxSizeMB := 500
	if err == nil {
		if size, err := strconv.Atoi(maxSizeMBStr); err == nil && size > 0 {
			maxSizeMB = size
		}
	}

	// Get current database size
	currentSizeMB, err := db.GetDatabaseSizeMB()
	if err != nil {
		return 0, err
	}

	// If under limit, no cleanup needed
	if currentSizeMB <= float64(maxSizeMB) {
		return 0, nil
	}

	log.Printf("Database size (%.2f MB) exceeds limit (%d MB), starting cleanup...", currentSizeMB, maxSizeMB)

	totalDeleted := int64(0)
	targetSizeMB := float64(maxSizeMB) * 0.95 // Aim for 95% of limit

	// Step 1: Delete oldest cached article contents. This saves most space while
	// keeping article metadata, read status, favorites, and dedupe keys intact.
	for currentSizeMB > targetSizeMB {
		result, err := db.Exec(`
			DELETE FROM article_contents
			WHERE article_id IN (
				SELECT article_id FROM article_contents
				ORDER BY fetched_at ASC
				LIMIT 100
			)
		`)
		if err != nil {
			break
		}

		count, _ := result.RowsAffected()
		if count == 0 {
			break // No more cached content to delete
		}

		totalDeleted += count
		currentSizeMB, _ = db.GetDatabaseSizeMB()
		log.Printf("Deleted %d cached article contents, current size: %.2f MB", count, currentSizeMB)
	}

	// Step 2: If still over limit, delete oldest read article metadata as a last resort.
	if currentSizeMB > targetSizeMB {
		count, err := db.CleanupReadArticlesOverPerFeedLimit(defaultMaxArticlesPerFeed)
		if err != nil {
			log.Printf("Per-feed article retention cleanup failed: %v", err)
		} else if count > 0 {
			totalDeleted += count
			currentSizeMB, _ = db.GetDatabaseSizeMB()
			log.Printf("Deleted %d read article metadata rows over per-feed limit, current size: %.2f MB", count, currentSizeMB)
		}
	}

	// Step 3: If per-feed retention is not enough, delete oldest read article metadata.
	for currentSizeMB > targetSizeMB {
		result, err := db.Exec(`
			DELETE FROM articles
			WHERE id IN (
				SELECT id FROM articles
				WHERE is_read = 1
				AND is_favorite = 0
				AND is_read_later = 0
				ORDER BY published_at ASC
				LIMIT 100
			)
		`)
		if err != nil {
			break
		}

		count, _ := result.RowsAffected()
		if count == 0 {
			break // No more read articles to delete
		}

		totalDeleted += count
		currentSizeMB, _ = db.GetDatabaseSizeMB()
		log.Printf("Deleted %d read article metadata rows, current size: %.2f MB", count, currentSizeMB)
	}

	if totalDeleted > 0 {
		log.Printf("Size-based cleanup completed: removed %d articles, final size: %.2f MB", totalDeleted, currentSizeMB)
		// Reclaim freelist pages to keep the file size in sync with actual data.
		// Without this, deleted rows leave freelist pages that inflate the file size
		// and cause ShouldCleanupBeforeSave to trigger on every SaveArticles call.
		if reclaimed, err := db.IncrementalVacuum(); err != nil {
			log.Printf("Warning: incremental vacuum after cleanup failed: %v", err)
		} else if reclaimed > 0 {
			log.Printf("Incremental vacuum reclaimed %d pages", reclaimed)
		}
	}

	return totalDeleted, nil
}

// CleanupArticleContentsByAge removes article content cache entries older than maxAgeDays
// This only deletes content, not article metadata
func (db *DB) CleanupArticleContentsByAge(maxAgeDays int) (int64, error) {
	db.WaitForReady()
	result, err := db.Exec(
		`DELETE FROM article_contents WHERE fetched_at < datetime('now', '-' || ? || ' days')`,
		maxAgeDays,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// CleanupArticleContentsBySize removes oldest article contents to reduce database size
// This only deletes content, not article metadata
func (db *DB) CleanupArticleContentsBySize() (int64, error) {
	db.WaitForReady()

	// Get max cache size from settings (default 500 MB)
	maxSizeMBStr, err := db.GetSetting("max_cache_size_mb")
	maxSizeMB := 500
	if err == nil {
		if size, err := strconv.Atoi(maxSizeMBStr); err == nil && size > 0 {
			maxSizeMB = size
		}
	}

	// Get current database size
	currentSizeMB, err := db.GetDatabaseSizeMB()
	if err != nil {
		return 0, err
	}

	// If under limit, no cleanup needed
	if currentSizeMB <= float64(maxSizeMB)*0.9 {
		return 0, nil
	}

	totalDeleted := int64(0)
	targetSizeMB := float64(maxSizeMB) * 0.85

	// Delete oldest contents in batches
	for currentSizeMB > targetSizeMB {
		result, err := db.Exec(`
			DELETE FROM article_contents
			WHERE article_id IN (
				SELECT article_id FROM article_contents
				ORDER BY fetched_at ASC
				LIMIT 100
			)
		`)
		if err != nil {
			break
		}

		count, _ := result.RowsAffected()
		if count == 0 {
			break
		}

		totalDeleted += count
		currentSizeMB, _ = db.GetDatabaseSizeMB()
	}

	if totalDeleted > 0 {
		_, _ = db.IncrementalVacuum()
	}

	return totalDeleted, nil
}

// CleanupOldArticlesLayered removes articles in layers:
// Layer 1: Read articles older than 30 days (not favorited/read later)
// Layer 2: Read articles older than 14 days (not favorited/read later)
// Layer 3: Unread articles older than 90 days (not favorited/read later)
// Layer 4: Unread articles older than 60 days (not favorited/read later)
func (db *DB) CleanupOldArticlesLayered() (int64, error) {
	db.WaitForReady()

	totalDeleted := int64(0)

	// Get max article age from settings
	maxAgeDaysStr, err := db.GetSetting("max_article_age_days")
	maxAgeDays := 30
	if err == nil {
		if days, err := strconv.Atoi(maxAgeDaysStr); err == nil && days > 0 {
			maxAgeDays = days
		}
	}

	// Layer 1: Delete very old read articles (maxAgeDays)
	cutoffDate := time.Now().AddDate(0, 0, -maxAgeDays)
	result, err := db.Exec(`
		DELETE FROM articles
		WHERE published_at < ?
		AND is_read = 1
		AND is_favorite = 0
		AND is_read_later = 0
	`, cutoffDate)
	if err == nil {
		count, _ := result.RowsAffected()
		totalDeleted += count
		if count > 0 {
			log.Printf("Layer 1: Deleted %d read articles older than %d days", count, maxAgeDays)
		}
	}

	// Layer 2: Delete old read articles (14 days)
	cutoffDate = time.Now().AddDate(0, 0, -14)
	result, err = db.Exec(`
		DELETE FROM articles
		WHERE published_at < ?
		AND is_read = 1
		AND is_favorite = 0
		AND is_read_later = 0
	`, cutoffDate)
	if err == nil {
		count, _ := result.RowsAffected()
		totalDeleted += count
		if count > 0 {
			log.Printf("Layer 2: Deleted %d read articles older than 14 days", count)
		}
	}

	// Layer 3: Delete very old unread articles (90 days)
	cutoffDate = time.Now().AddDate(0, 0, -90)
	result, err = db.Exec(`
		DELETE FROM articles
		WHERE published_at < ?
		AND is_read = 0
		AND is_favorite = 0
		AND is_read_later = 0
	`, cutoffDate)
	if err == nil {
		count, _ := result.RowsAffected()
		totalDeleted += count
		if count > 0 {
			log.Printf("Layer 3: Deleted %d unread articles older than 90 days", count)
		}
	}

	// Layer 4: Delete old unread articles (60 days)
	cutoffDate = time.Now().AddDate(0, 0, -60)
	result, err = db.Exec(`
		DELETE FROM articles
		WHERE published_at < ?
		AND is_read = 0
		AND is_favorite = 0
		AND is_read_later = 0
	`, cutoffDate)
	if err == nil {
		count, _ := result.RowsAffected()
		totalDeleted += count
		if count > 0 {
			log.Printf("Layer 4: Deleted %d unread articles older than 60 days", count)
		}
	}

	// Reclaim freelist pages if we deleted anything
	if totalDeleted > 0 {
		if reclaimed, err := db.IncrementalVacuum(); err != nil {
			log.Printf("Warning: incremental vacuum after layered cleanup failed: %v", err)
		} else if reclaimed > 0 {
			log.Printf("Incremental vacuum reclaimed %d pages after layered cleanup", reclaimed)
		}
	}

	return totalDeleted, nil
}

// CleanupOldReadArticles removes read articles older than specified days
// Protects favorited and read later articles
// With foreign_keys enabled, ON DELETE CASCADE automatically removes
// associated article_contents, chat_sessions, and chat_messages rows.
func (db *DB) CleanupOldReadArticles(maxAgeDays int) (int64, error) {
	db.WaitForReady()

	cutoffDate := time.Now().AddDate(0, 0, -maxAgeDays)
	result, err := db.Exec(`
		DELETE FROM articles
		WHERE published_at < ?
		AND is_read = 1
		AND is_favorite = 0
		AND is_read_later = 0
	`, cutoffDate)
	if err != nil {
		return 0, err
	}

	count, _ := result.RowsAffected()
	return count, nil
}

// CleanupOldUnreadArticles removes unread articles older than specified days
// Protects favorited and read later articles
// With foreign_keys enabled, ON DELETE CASCADE automatically removes
// associated article_contents, chat_sessions, and chat_messages rows.
func (db *DB) CleanupOldUnreadArticles(maxAgeDays int) (int64, error) {
	db.WaitForReady()

	cutoffDate := time.Now().AddDate(0, 0, -maxAgeDays)
	result, err := db.Exec(`
		DELETE FROM articles
		WHERE published_at < ?
		AND is_read = 0
		AND is_favorite = 0
		AND is_read_later = 0
	`, cutoffDate)
	if err != nil {
		return 0, err
	}

	count, _ := result.RowsAffected()
	return count, nil
}

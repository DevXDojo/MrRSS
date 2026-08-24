package freshrss

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/models"
)

// SyncResult represents the result of a sync operation
type SyncResult struct {
	PullSuccess      bool
	PullChangesCount int
	PushSuccess      bool
	PushChangesCount int
	Errors           []string
	Duration         time.Duration
	LastSyncTime     time.Time
}

// BidirectionalSyncService handles bidirectional synchronization
type BidirectionalSyncService struct {
	client *Client
	db     *database.DB
}

// NewBidirectionalSyncService creates a new bidirectional sync service
func NewBidirectionalSyncService(serverURL, username, password string, db *database.DB) *BidirectionalSyncService {
	return &BidirectionalSyncService{
		client: NewClient(serverURL, username, password),
		db:     db,
	}
}

// Sync performs a full bidirectional sync
// This is called for manual/scheduled sync
// Logic: Pull remote changes first, then push local changes
func (s *BidirectionalSyncService) Sync(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{
		LastSyncTime: time.Now(),
	}
	startTime := time.Now()
	defer func() { result.Duration = time.Since(startTime) }()

	// Stage 1: Login to FreshRSS
	if err := s.client.Login(ctx); err != nil {
		return result, fmt.Errorf("login failed: %w", err)
	}

	// Stage 2: Pull from server (feeds, articles, starred status, read status)
	log.Printf("Stage 1: Pull from server")
	pullChanges, err := s.pullFromServer(ctx)
	if err != nil {
		log.Printf("Stage 1 ERROR: pull failed: %v", err)
		result.Errors = append(result.Errors, fmt.Sprintf("pull failed: %v", err))
		result.PullSuccess = false
		result.PushSuccess = false
		return result, err
	}
	log.Printf("Stage 1 SUCCESS: %d changes pulled", pullChanges)
	result.PullSuccess = true
	result.PullChangesCount = pullChanges

	// Stage 3: Push local changes to server
	log.Printf("Stage 2: Push to server")
	pushChanges, err := s.pushToServer(ctx)
	if err != nil {
		log.Printf("Stage 2 ERROR: push failed: %v", err)
		result.Errors = append(result.Errors, fmt.Sprintf("push failed: %v", err))
		result.PushSuccess = false
	} else {
		log.Printf("Stage 2 SUCCESS: %d changes pushed", pushChanges)
		result.PushSuccess = true
		result.PushChangesCount = pushChanges
	}

	return result, nil
}

// SyncFeed syncs articles for a single FreshRSS feed/stream
// This is called when user right-clicks a FreshRSS feed and selects "Sync Feed"
// Now fetches ALL articles using pagination, not just a limited number
func (s *BidirectionalSyncService) SyncFeed(ctx context.Context, streamID string) (int, error) {
	// Login to FreshRSS
	if err := s.client.Login(ctx); err != nil {
		return 0, fmt.Errorf("login failed: %w", err)
	}

	log.Printf("[SyncFeed] Syncing stream: %s", streamID)

	// Fetch all articles from this stream using pagination
	allArticles, err := s.fetchAllArticles(ctx, streamID, 10000)
	if err != nil {
		return 0, fmt.Errorf("fetch articles from stream: %w", err)
	}

	if len(allArticles) == 0 {
		log.Printf("[SyncFeed] No articles in stream: %s", streamID)
		return 0, nil
	}

	// Save articles to database
	count, err := s.saveArticlesFromServer(ctx, allArticles)
	if err != nil {
		return 0, fmt.Errorf("save articles: %w", err)
	}

	log.Printf("[SyncFeed] Synced %d articles for stream: %s", count, streamID)
	return count, nil
}

// SyncArticleStatus syncs a single article's status immediately
// This is called when user manually marks an article as read/unread or starred/unstarred
// Logic: Immediately push local status to server, overwriting remote
// If sync fails, the change is added to the queue for later retry
func (s *BidirectionalSyncService) SyncArticleStatus(ctx context.Context, articleID int64, articleURL string, action database.SyncAction) error {
	// Login to FreshRSS
	if err := s.client.Login(ctx); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Get the article to check if we have FreshRSS Item ID
	article, err := s.db.GetArticleByID(articleID)
	if err != nil {
		log.Printf("[Immediate Sync] Failed to get article: %v", err)
		return err
	}

	// Use FreshRSS Item ID if available, otherwise fall back to URL
	identifier := articleURL
	if article.FreshRSSItemID != "" {
		identifier = article.FreshRSSItemID
		log.Printf("[Immediate Sync] Using FreshRSS Item ID: %s for article %d", identifier, articleID)
	} else {
		log.Printf("[Immediate Sync] No FreshRSS Item ID, using URL: %s for article %d", articleURL, articleID)
	}

	// Perform the action immediately
	log.Printf("[Immediate Sync] Syncing article status: %s (%s) -> %s", articleURL, identifier, action)

	var syncErr error
	switch action {
	case database.SyncActionMarkRead:
		syncErr = s.client.MarkAsReadBatch(ctx, []string{identifier})
	case database.SyncActionMarkUnread:
		syncErr = s.client.MarkAsUnreadBatch(ctx, []string{identifier})
	case database.SyncActionStar:
		syncErr = s.client.StarBatch(ctx, []string{identifier})
	case database.SyncActionUnstar:
		syncErr = s.client.UnstarBatch(ctx, []string{identifier})
	}

	if syncErr != nil {
		log.Printf("[Immediate Sync] ERROR: %v", syncErr)
		// Add to queue for retry
		if queueErr := s.db.EnqueueSyncChange(articleID, articleURL, action); queueErr != nil {
			log.Printf("[Immediate Sync] Failed to enqueue for retry: %v", queueErr)
		} else {
			log.Printf("[Immediate Sync] Enqueued for retry due to sync failure")
		}
		return syncErr
	}

	log.Printf("[Immediate Sync] SUCCESS: %s -> %s", articleURL, action)
	return nil
}

// pullFromServer pulls changes from FreshRSS server
func (s *BidirectionalSyncService) pullFromServer(ctx context.Context) (int, error) {
	totalChanges := 0
	log.Printf("pullFromServer: Starting pull from server")

	// Step 1: Get subscriptions and create feeds
	subscriptions, err := s.client.GetSubscriptions(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get subscriptions: %v", err)
		if subscriptions == nil {
			subscriptions = []Subscription{}
		}
	}

	// Only proceed if we have subscriptions
	if len(subscriptions) > 0 {
		feedsCreated, err := s.createFeedsFromSubscriptions(ctx, subscriptions)
		if err != nil {
			log.Printf("Warning: Failed to create feeds: %v", err)
		} else {
			totalChanges += feedsCreated
			log.Printf("Created/updated %d feeds from FreshRSS", feedsCreated)
		}

		// Step 2: Get articles from each subscription (no limit - fetch all articles)
		articlesPerFeed := 10000 // Increased from 100 to fetch more articles per feed
		totalArticles := 0

		for _, sub := range subscriptions {
			feedURL := strings.TrimPrefix(sub.ID, "feed/")

			// Fetch all articles using pagination
			continuation := ""
			feedArticleCount := 0
			for {
				result, err := s.client.GetStreamContents(ctx, sub.ID, nil, articlesPerFeed, continuation)
				if err != nil {
					log.Printf("Warning: Failed to get articles for feed %s: %v", feedURL, err)
					break
				}

				if len(result.Items) > 0 {
					saved, err := s.saveArticlesFromServer(ctx, result.Items)
					if err != nil {
						log.Printf("Warning: Failed to save articles for feed %s: %v", feedURL, err)
					} else {
						totalArticles += saved
						feedArticleCount += saved
					}
				}

				// Check if there are more articles to fetch
				if result.Continuation == "" {
					break
				}
				continuation = result.Continuation
				log.Printf("Fetched %d articles for feed %s, continuing with pagination...", feedArticleCount, feedURL)
				time.Sleep(50 * time.Millisecond)
			}

			log.Printf("Total articles fetched for feed %s: %d", feedURL, feedArticleCount)
			time.Sleep(50 * time.Millisecond)
		}

		totalChanges += totalArticles
		log.Printf("Saved %d total articles from %d feeds", totalArticles, len(subscriptions))
	} else {
		log.Printf("No subscriptions to sync from FreshRSS")
	}

	// Step 3: Apply starred status from server (fetch all with pagination)
	log.Printf("pullFromServer: Step 3 - Applying starred status")
	starredArticles, err := s.fetchAllArticles(ctx, "user/-/state/com.google/starred", 10000)
	if err != nil {
		log.Printf("Warning: Failed to get starred articles: %v", err)
	} else {
		for _, article := range starredArticles {
			_, err := s.applyServerStatus(article.URL, true, "is_favorite")
			if err != nil {
				log.Printf("Warning: Failed to apply starred status for %s: %v", article.URL, err)
			}
		}
		log.Printf("Applied starred status to %d articles from server", len(starredArticles))
	}

	// Step 4: Apply read status from server using unread count API (fetch all with pagination)
	log.Printf("pullFromServer: Step 4 - Applying read status using unread count API")

	// First, get unread counts from server to verify sync
	unreadCounts, err := s.client.GetUnreadCount(ctx)
	if err != nil {
		log.Printf("Warning: Failed to get unread counts: %v", err)
	} else {
		// Log unread counts for debugging
		for streamID, count := range unreadCounts {
			if count > 0 {
				log.Printf("Server unread count: %s = %d", streamID, count)
			}
		}
	}

	// Fetch all read articles using pagination
	readArticles, err := s.fetchAllArticles(ctx, "user/-/state/com.google/read", 10000)
	if err != nil {
		log.Printf("Warning: Failed to get read articles: %v", err)
	} else {
		for _, article := range readArticles {
			_, err := s.applyServerStatus(article.URL, true, "is_read")
			if err != nil {
				log.Printf("Warning: Failed to apply read status for %s: %v", article.URL, err)
			}
		}
		log.Printf("Applied read status to %d articles from server", len(readArticles))
	}

	log.Printf("Pull from server: %d total changes applied", totalChanges)
	log.Printf("pullFromServer: Completed")

	return totalChanges, nil
}

// fetchAllArticles fetches all articles from a stream using pagination
// This helper function automatically handles pagination to fetch all articles
func (s *BidirectionalSyncService) fetchAllArticles(ctx context.Context, streamID string, itemsPerPage int) ([]Article, error) {
	allArticles := make([]Article, 0)
	continuation := ""
	pageCount := 0

	for {
		result, err := s.client.GetStreamContents(ctx, streamID, nil, itemsPerPage, continuation)
		if err != nil {
			return allArticles, fmt.Errorf("fetch page %d: %w", pageCount, err)
		}

		if len(result.Items) == 0 {
			break
		}

		allArticles = append(allArticles, result.Items...)
		pageCount++

		log.Printf("Fetched page %d (%d articles) from stream %s (total: %d)",
			pageCount, len(result.Items), streamID, len(allArticles))

		// Check if there are more articles to fetch
		if result.Continuation == "" {
			break
		}
		continuation = result.Continuation

		// Small delay to avoid overwhelming the server
		time.Sleep(50 * time.Millisecond)
	}

	log.Printf("Finished fetching %d total articles from stream %s (%d pages)",
		len(allArticles), streamID, pageCount)
	return allArticles, nil
}

// applyServerStatus applies server status to local article
// If local has a pending sync for the same action type, clear it to avoid conflicts
func (s *BidirectionalSyncService) applyServerStatus(articleURL string, status bool, column string) (int, error) {
	// Check if article exists locally
	localArticle, err := s.db.GetArticleByURL(articleURL)
	if err != nil {
		// Article doesn't exist locally, skip
		return 0, nil
	}

	// Check if there's a pending sync change for the SAME action type
	// Only clear conflicting pending syncs
	pendingItems, err := s.db.GetPendingSyncChanges(1000)
	hasConflictingSync := false
	for _, item := range pendingItems {
		if item.ArticleURL == articleURL {
			isConflictingAction := false
			switch column {
			case "is_favorite":
				// Star/unstar actions conflict with favorite status changes
				isConflictingAction = (item.Action == database.SyncActionStar || item.Action == database.SyncActionUnstar)
			case "is_read":
				// Mark read/unread actions conflict with read status changes
				isConflictingAction = (item.Action == database.SyncActionMarkRead || item.Action == database.SyncActionMarkUnread)
			}

			if isConflictingAction {
				hasConflictingSync = true
				log.Printf("[Conflict detected] Article %s has pending %s, server says %s=%v - clearing pending sync",
					articleURL, item.Action, column, status)
				break
			}
		}
	}

	// Apply server status
	switch column {
	case "is_favorite":
		err = s.db.SetArticleFavorite(localArticle.ID, status)
	case "is_read":
		err = s.db.MarkArticleRead(localArticle.ID, status)
	}
	if err != nil {
		return 0, err
	}

	// Clear conflicting pending sync if any
	if hasConflictingSync {
		_ = s.db.ClearPendingSyncForArticle(localArticle.ID)
	}

	return 1, nil
}

// createFeedsFromSubscriptions creates local feeds from FreshRSS subscriptions
func (s *BidirectionalSyncService) createFeedsFromSubscriptions(ctx context.Context, subscriptions []Subscription) (int, error) {
	feedsCreated := 0

	// Get all existing feeds to check for duplicates
	existingFeeds, err := s.db.GetFeeds()
	if err != nil {
		log.Printf("Warning: Failed to get existing feeds: %v", err)
	}

	// Create a map of feed URLs to existing feeds
	// Since we now allow same URL from different sources, use composite key: url + is_freshrss_source
	type feedKey struct {
		URL              string
		IsFreshRSSSource bool
	}
	feedMap := make(map[feedKey]*models.Feed)
	titleMap := make(map[string]int64)
	// Track categories to check if they contain non-FreshRSS feeds
	categoryMap := make(map[string][]*models.Feed) // category -> feeds in this category
	for i := range existingFeeds {
		key := feedKey{
			URL:              existingFeeds[i].URL,
			IsFreshRSSSource: existingFeeds[i].IsFreshRSSSource,
		}
		feedMap[key] = &existingFeeds[i]
		titleMap[existingFeeds[i].Title] = existingFeeds[i].ID
		category := existingFeeds[i].Category
		if category != "" {
			categoryMap[category] = append(categoryMap[category], &existingFeeds[i])
		}
	}

	// Helper function to generate unique category name for FreshRSS
	generateFreshRSSCategoryName := func(originalCategory string) string {
		// If category doesn't exist or has only FreshRSS feeds, use as-is
		feeds := categoryMap[originalCategory]
		if len(feeds) == 0 {
			return originalCategory
		}
		allFreshRSS := true
		for _, feed := range feeds {
			if !feed.IsFreshRSSSource {
				allFreshRSS = false
				break
			}
		}
		if allFreshRSS {
			return originalCategory
		}

		// Category has mixed or non-FreshRSS feeds, need to rename
		newCategory := originalCategory + " (FreshRSS)"
		counter := 1
		for {
			if _, exists := categoryMap[newCategory]; !exists {
				break
			}
			newCategory = fmt.Sprintf("%s (FreshRSS %d)", originalCategory, counter)
			counter++
		}
		log.Printf("[Category Conflict] Renaming FreshRSS category '%s' to '%s' to avoid mixing with local feeds",
			originalCategory, newCategory)
		return newCategory
	}

	for _, sub := range subscriptions {
		feedURL := sub.URL

		// Extract category from subscription categories
		category := ""
		if len(sub.Categories) > 0 {
			for _, cat := range sub.Categories {
				if strings.HasPrefix(cat.ID, "user/-/label/") {
					originalCategory := cat.Label
					// Check if this category would conflict with local feeds
					category = generateFreshRSSCategoryName(originalCategory)
					break
				}
			}
		}

		// Adjust title if there's a conflict with an existing feed with same title but different URL
		feedTitle := sub.Title
		if existingID, exists := titleMap[feedTitle]; exists {
			// Check if there's a feed with this title but different URL (not the same feed)
			titleConflict := true
			for _, feed := range existingFeeds {
				if feed.ID == existingID && feed.URL == feedURL {
					// Same title and same URL - this is actually the same feed (possibly different source)
					titleConflict = false
					break
				}
			}
			if titleConflict {
				feedTitle = feedTitle + " (FreshRSS)"
				log.Printf("Title conflict detected for '%s', using adjusted title '%s'", sub.Title, feedTitle)
			}
		}

		// Check if feed already exists (by URL + FreshRSS source combination)
		key := feedKey{
			URL:              feedURL,
			IsFreshRSSSource: true, // We're syncing FreshRSS feeds
		}

		if existingFeed, exists := feedMap[key]; exists {
			// Feed exists with same URL and same source type, check if we need to update it
			needsUpdate := false

			if existingFeed.Title != feedTitle {
				needsUpdate = true
			}

			if category != "" && existingFeed.Category != category {
				needsUpdate = true
			}

			if needsUpdate {
				updateCategory := existingFeed.Category
				if category != "" {
					updateCategory = category
				}

				err := s.db.UpdateFeed(
					existingFeed.ID,
					feedTitle,
					existingFeed.URL,
					updateCategory,
					existingFeed.ScriptPath,
					existingFeed.HideFromTimeline,
					existingFeed.ProxyURL,
					existingFeed.ProxyEnabled,
					existingFeed.RefreshInterval,
					existingFeed.IsImageMode,
					existingFeed.Type,
					existingFeed.XPathItem,
					existingFeed.XPathItemTitle,
					existingFeed.XPathItemContent,
					existingFeed.XPathItemUri,
					existingFeed.XPathItemAuthor,
					existingFeed.XPathItemTimestamp,
					existingFeed.XPathItemTimeFormat,
					existingFeed.XPathItemThumbnail,
					existingFeed.XPathItemCategories,
					existingFeed.XPathItemUid,
					existingFeed.ArticleViewMode,
					existingFeed.AutoExpandContent,
					existingFeed.EmailAddress,
					existingFeed.EmailIMAPServer,
					existingFeed.EmailUsername,
					existingFeed.EmailPassword,
					existingFeed.EmailFolder,
					existingFeed.EmailIMAPPort,
				)
				if err != nil {
					log.Printf("Warning: Failed to update feed %s: %v", feedURL, err)
				} else {
					feedsCreated++
					log.Printf("Updated feed '%s' with category '%s'", feedTitle, updateCategory)
				}
			}
			continue
		}

		// Check if there's a local feed with the same URL
		localKey := feedKey{
			URL:              feedURL,
			IsFreshRSSSource: false,
		}
		if _, exists := feedMap[localKey]; exists {
			log.Printf("[URL Conflict] Local feed with URL '%s' already exists, creating separate FreshRSS feed with title '%s'", feedURL, feedTitle)
			// The title should already have been adjusted by the title conflict logic above
			// Just continue to create the new feed below
		}

		// Create new feed
		newFeed := &models.Feed{
			URL:              feedURL,
			Title:            feedTitle,
			Link:             feedURL,
			Description:      "",
			Category:         category,
			IsFreshRSSSource: true,
			FreshRSSStreamID: sub.ID,
		}

		_, err = s.db.AddFeed(newFeed)
		if err != nil {
			log.Printf("Warning: Failed to create feed %s: %v", feedURL, err)
		} else {
			feedsCreated++
			if category != "" {
				log.Printf("Created feed '%s' in category '%s'", feedTitle, category)
			} else {
				log.Printf("Created feed '%s' (no category)", feedTitle)
			}
		}
	}

	// Delete local FreshRSS feeds that no longer exist on the server
	remoteFeedURLs := make(map[string]bool)
	for _, sub := range subscriptions {
		remoteFeedURLs[sub.URL] = true
	}

	for _, feed := range existingFeeds {
		if feed.IsFreshRSSSource {
			if !remoteFeedURLs[feed.URL] {
				log.Printf("Deleting local FreshRSS feed '%s' (removed from server)", feed.Title)
				err := s.db.DeleteFeed(feed.ID)
				if err != nil {
					log.Printf("Warning: Failed to delete feed '%s': %v", feed.Title, err)
				} else {
					feedsCreated++
				}
			}
		}
	}

	return feedsCreated, nil
}

// saveArticlesFromServer saves articles from FreshRSS to local database
func (s *BidirectionalSyncService) saveArticlesFromServer(ctx context.Context, articles []Article) (int, error) {
	if len(articles) == 0 {
		return 0, nil
	}

	// Get all existing feeds to map stream IDs to feed IDs
	existingFeeds, err := s.db.GetFeeds()
	if err != nil {
		return 0, fmt.Errorf("get existing feeds: %w", err)
	}

	feedStreamIDMap := make(map[string]int64)
	feedURLMap := make(map[string]int64)
	for i := range existingFeeds {
		var streamID string
		if existingFeeds[i].IsFreshRSSSource && existingFeeds[i].FreshRSSStreamID != "" {
			streamID = existingFeeds[i].FreshRSSStreamID
		} else {
			streamID = "feed/" + existingFeeds[i].URL
		}
		feedStreamIDMap[streamID] = existingFeeds[i].ID
		feedURLMap[existingFeeds[i].URL] = existingFeeds[i].ID
	}

	// Convert FreshRSS articles to models.Article
	mrssArticles := make([]*models.Article, 0, len(articles))
	articleContentMap := make(map[string]string)
	skippedCount := 0

	for _, article := range articles {
		var feedID int64

		// Try to find feed ID using article's Origin stream ID first
		if article.OriginStreamID != "" {
			if id, exists := feedStreamIDMap[article.OriginStreamID]; exists {
				feedID = id
			}
		}

		// Fallback: try to find feed ID using categories
		if feedID == 0 && len(article.Categories) > 0 {
			for _, cat := range article.Categories {
				if strings.HasPrefix(cat, "user/-/label/") {
					feedURL := strings.TrimPrefix(cat, "user/-/label/")
					feedURL = strings.TrimPrefix(feedURL, "feed/")
					if id, exists := feedURLMap[feedURL]; exists {
						feedID = id
						break
					}
				}
			}
		}

		if feedID == 0 {
			skippedCount++
			if skippedCount <= 5 {
				log.Printf("Warning: Could not find feed for article '%s' (stream ID: %s)",
					article.Title, article.OriginStreamID)
			}
			continue
		}

		// Check article status from categories
		isRead := false
		isStarred := false
		for _, cat := range article.Categories {
			if cat == "user/-/state/com.google/read" {
				isRead = true
			}
			if cat == "user/-/state/com.google/starred" {
				isStarred = true
			}
		}

		// Check if article already exists (by URL)
		existingArticle, err := s.db.GetArticleByURL(article.URL)

		if err == nil && existingArticle != nil {
			// Article already exists - this is the deduplication logic
			// Update the article with FreshRSS data, preserving FreshRSS ID
			updated := false

			// ALWAYS update FreshRSS Item ID if provided by FreshRSS
			// This ensures that even if the article came from a non-FreshRSS source,
			// it will be linked to FreshRSS for future sync operations
			if article.ID != "" && existingArticle.FreshRSSItemID != article.ID {
				err := s.db.UpdateFreshRSSItemID(existingArticle.ID, article.ID)
				if err != nil {
					log.Printf("Warning: Failed to update FreshRSS Item ID for article %s: %v", article.URL, err)
				} else {
					log.Printf("Updated FreshRSS Item ID for existing article %s: %s (was: %s)",
						article.URL, article.ID, existingArticle.FreshRSSItemID)
					updated = true
				}
			}

			// Update read status from FreshRSS (server is authoritative)
			// Only update if status differs to avoid unnecessary writes
			if isRead != existingArticle.IsRead {
				err := s.db.MarkArticleRead(existingArticle.ID, isRead)
				if err != nil {
					log.Printf("Warning: Failed to update read status for article %s: %v", article.URL, err)
				} else {
					log.Printf("Updated read status for article %s: %v (from FreshRSS)", article.URL, isRead)
					updated = true
				}
			}

			// Update favorite status from FreshRSS (server is authoritative)
			if isStarred != existingArticle.IsFavorite {
				err := s.db.SetArticleFavorite(existingArticle.ID, isStarred)
				if err != nil {
					log.Printf("Warning: Failed to update favorite status for article %s: %v", article.URL, err)
				} else {
					log.Printf("Updated favorite status for article %s: %v (from FreshRSS)", article.URL, isStarred)
					updated = true
				}
			}

			// Extract and update thumbnail if article doesn't have one but has content
			if article.Content != "" {
				// Check if article already has an image URL
				var existingImageURL string
				err := s.db.QueryRow("SELECT image_url FROM articles WHERE id = ?", existingArticle.ID).Scan(&existingImageURL)
				hasImage := err == nil && existingImageURL != ""

				if !hasImage {
					imageURL := extractImageURLFromHTML(article.Content)
					if imageURL != "" {
						// Update the article with the extracted image URL
						_, err := s.db.Exec("UPDATE articles SET image_url = ? WHERE id = ?", imageURL, existingArticle.ID)
						if err != nil {
							log.Printf("Warning: Failed to update image URL for article %s: %v", article.URL, err)
						} else {
							log.Printf("Updated image URL for article %s: %s (from FreshRSS)", article.URL, imageURL)
							updated = true
						}
					}
				}
			}

			// If the existing article is NOT from FreshRSS but we just updated it,
			// we should mark it as coming from FreshRSS if it's in a FreshRSS feed
			if existingArticle.FreshRSSItemID == "" && article.ID != "" {
				// This article now has a FreshRSS ID
				updated = true
			}

			if updated {
				log.Printf("Merged FreshRSS data into existing article: %s", article.URL)
			}
			continue
		}

		// Extract thumbnail from article content before creating the article
		imageURL := extractImageURLFromHTML(article.Content)

		// Create new article
		mrssArticle := &models.Article{
			FeedID:         feedID,
			Title:          article.Title,
			URL:            article.URL,
			ImageURL:       imageURL,
			Summary:        "",
			PublishedAt:    article.Published,
			IsRead:         isRead,
			IsFavorite:     isStarred,
			FreshRSSItemID: article.ID, // Save FreshRSS/Google Reader item ID
		}

		mrssArticles = append(mrssArticles, mrssArticle)

		// Store content for later insertion
		if article.Content != "" {
			articleContentMap[article.URL] = article.Content
		}
	}

	if skippedCount > 5 {
		log.Printf("Warning: Skipped %d articles due to missing feed mapping", skippedCount)
	}

	if len(mrssArticles) == 0 {
		return 0, nil
	}

	// Save articles to database
	err = s.db.SaveArticles(ctx, mrssArticles)
	if err != nil {
		return 0, fmt.Errorf("save articles: %w", err)
	}

	// Save article contents
	contentSavedCount := 0
	for url, content := range articleContentMap {
		savedArticle, err := s.db.GetArticleByURL(url)
		if err != nil {
			log.Printf("Warning: Could not find saved article with URL %s to set content", url)
			continue
		}

		if err := s.db.SetArticleContent(savedArticle.ID, content); err != nil {
			log.Printf("Warning: Failed to save content for article ID %d: %v", savedArticle.ID, err)
		} else {
			contentSavedCount++
		}
	}

	if contentSavedCount > 0 {
		log.Printf("Saved content for %d articles", contentSavedCount)
	}

	return len(mrssArticles), nil
}

// pushToServer pushes local changes to FreshRSS server
// This compares local vs remote state and immediately syncs any differences
func (s *BidirectionalSyncService) pushToServer(ctx context.Context) (int, error) {
	totalChanges := 0

	// First, process any failed items from the queue (retry mechanism)
	pendingChanges, err := s.db.GetPendingSyncChanges(500)
	if err != nil {
		log.Printf("Warning: Failed to get pending changes: %v", err)
	} else if len(pendingChanges) > 0 {
		log.Printf("[Push] Retrying %d failed sync items from queue", len(pendingChanges))
		for i, item := range pendingChanges {
			log.Printf("  [%d] ArticleID=%d URL=%s Action=%s", i, item.ArticleID, item.ArticleURL, item.Action)
		}
		changes, err := s.pushPendingItems(ctx, pendingChanges)
		if err != nil {
			log.Printf("[Push] Error pushing pending items: %v", err)
			// Continue anyway, don't fail entire sync
		} else {
			totalChanges += changes
		}
	}

	// Get all FreshRSS feeds
	feeds, err := s.db.GetFeeds()
	if err != nil {
		return totalChanges, fmt.Errorf("get feeds: %w", err)
	}

	// Filter to only FreshRSS feeds
	freshRSSFeeds := make([]models.Feed, 0)
	for _, feed := range feeds {
		if feed.IsFreshRSSSource {
			freshRSSFeeds = append(freshRSSFeeds, feed)
		}
	}

	if len(freshRSSFeeds) == 0 {
		log.Printf("[Push] No FreshRSS feeds to sync")
		return totalChanges, nil
	}

	log.Printf("[Push] Checking %d FreshRSS feeds for local changes to push", len(freshRSSFeeds))

	// Get remote read and starred articles
	remoteReadArticles := make(map[string]bool)    // URL -> is read
	remoteStarredArticles := make(map[string]bool) // URL -> is starred

	// Fetch remote read status (fetch all with pagination)
	readArticles, err := s.fetchAllArticles(ctx, "user/-/state/com.google/read", 10000)
	if err != nil {
		log.Printf("[Push] Warning: Failed to get remote read articles: %v", err)
	} else {
		for _, article := range readArticles {
			remoteReadArticles[article.URL] = true
		}
		log.Printf("[Push] Fetched %d read articles from remote", len(remoteReadArticles))
	}

	// Fetch remote starred status (fetch all with pagination)
	starredArticles, err := s.fetchAllArticles(ctx, "user/-/state/com.google/starred", 10000)
	if err != nil {
		log.Printf("[Push] Warning: Failed to get remote starred articles: %v", err)
	} else {
		for _, article := range starredArticles {
			remoteStarredArticles[article.URL] = true
		}
		log.Printf("[Push] Fetched %d starred articles from remote", len(remoteStarredArticles))
	}

	// For each FreshRSS feed, check articles and sync differences
	readIDs := make([]string, 0)
	unreadIDs := make([]string, 0)
	starIDs := make([]string, 0)
	unstarIDs := make([]string, 0)

	for _, feed := range freshRSSFeeds {
		// Get all articles for this feed (increased limit from 1000 to 10000)
		// Use pagination to get ALL articles if needed
		allArticles := make([]models.Article, 0)
		offset := 0
		limit := 10000

		for {
			articles, err := s.db.GetArticles("all", feed.ID, "", false, limit, offset)
			if err != nil {
				log.Printf("[Push] Warning: Failed to get articles for feed %s: %v", feed.Title, err)
				break
			}

			if len(articles) == 0 {
				break
			}

			allArticles = append(allArticles, articles...)

			// If we got fewer articles than the limit, we've reached the end
			if len(articles) < limit {
				break
			}

			offset += limit
			log.Printf("[Push] Fetched batch of %d articles for feed %s (total: %d)",
				len(articles), feed.Title, len(allArticles))
		}

		log.Printf("[Push] Total %d articles to check for feed %s", len(allArticles), feed.Title)

		for _, article := range allArticles {
			// Use FreshRSS Item ID if available, otherwise fall back to URL
			identifier := article.URL
			if article.FreshRSSItemID != "" {
				identifier = article.FreshRSSItemID
			}

			// Check read status differences
			remoteIsRead := remoteReadArticles[article.URL]
			if article.IsRead && !remoteIsRead {
				// Local is read, remote is not - push read status
				readIDs = append(readIDs, identifier)
			} else if !article.IsRead && remoteIsRead {
				// Local is unread, remote is read - push unread status
				unreadIDs = append(unreadIDs, identifier)
			}

			// Check starred status differences
			remoteIsStarred := remoteStarredArticles[article.URL]
			if article.IsFavorite && !remoteIsStarred {
				// Local is starred, remote is not - push star status
				starIDs = append(starIDs, identifier)
			} else if !article.IsFavorite && remoteIsStarred {
				// Local is unstarred, remote is starred - push unstar status
				unstarIDs = append(unstarIDs, identifier)
			}
		}
	}

	// Execute batch operations
	if len(readIDs) > 0 {
		log.Printf("[Push] Marking %d articles as read (local has read, remote doesn't)", len(readIDs))
		if err := s.client.MarkAsReadBatch(ctx, readIDs); err != nil {
			log.Printf("[Push] ERROR marking as read: %v", err)
			return totalChanges, fmt.Errorf("mark read batch: %w", err)
		}
		log.Printf("[Push] Successfully marked %d articles as read", len(readIDs))
		totalChanges += len(readIDs)
	}

	if len(unreadIDs) > 0 {
		log.Printf("[Push] Marking %d articles as unread (local has unread, remote doesn't)", len(unreadIDs))
		if err := s.client.MarkAsUnreadBatch(ctx, unreadIDs); err != nil {
			log.Printf("[Push] ERROR marking as unread: %v", err)
			return totalChanges, fmt.Errorf("mark unread batch: %w", err)
		}
		log.Printf("[Push] Successfully marked %d articles as unread", len(unreadIDs))
		totalChanges += len(unreadIDs)
	}

	if len(starIDs) > 0 {
		log.Printf("[Push] Starring %d articles (local has starred, remote doesn't)", len(starIDs))
		if err := s.client.StarBatch(ctx, starIDs); err != nil {
			log.Printf("[Push] ERROR starring: %v", err)
			return totalChanges, fmt.Errorf("star batch: %w", err)
		}
		log.Printf("[Push] Successfully starred %d articles", len(starIDs))
		totalChanges += len(starIDs)
	}

	if len(unstarIDs) > 0 {
		log.Printf("[Push] Unstarring %d articles (local has unstarred, remote doesn't)", len(unstarIDs))
		if err := s.client.UnstarBatch(ctx, unstarIDs); err != nil {
			log.Printf("[Push] ERROR unstarring: %v", err)
			return totalChanges, fmt.Errorf("unstar batch: %w", err)
		}
		log.Printf("[Push] Successfully unstarred %d articles", len(unstarIDs))
		totalChanges += len(unstarIDs)
	}

	log.Printf("[Push] Total %d changes synced to server", totalChanges)
	return totalChanges, nil
}

// pushPendingItems pushes items that failed previously (from the queue)
func (s *BidirectionalSyncService) pushPendingItems(ctx context.Context, pendingChanges []database.SyncQueueItem) (int, error) {
	totalChanges := 0

	// Group changes by action type
	readIDs := make([]string, 0)
	unreadIDs := make([]string, 0)
	starIDs := make([]string, 0)
	unstarIDs := make([]string, 0)
	itemIDs := make([]int64, 0)

	// Get article IDs to fetch FreshRSS item IDs
	articleIDs := make([]int64, len(pendingChanges))
	for i, item := range pendingChanges {
		articleIDs[i] = item.ArticleID
	}

	// Fetch articles to get their FreshRSS item IDs
	articles, err := s.db.GetArticlesByIDs(articleIDs)
	if err != nil {
		log.Printf("[PushPending] Failed to fetch articles: %v", err)
		return totalChanges, err
	}

	// Create a map of articles by their ID
	articleByID := make(map[int64]models.Article)
	for _, article := range articles {
		articleByID[article.ID] = article
	}

	// Use FreshRSS item ID if available, otherwise fall back to URL
	for _, item := range pendingChanges {
		itemIDs = append(itemIDs, item.ID)

		article, exists := articleByID[item.ArticleID]
		identifier := item.ArticleURL // Default fallback

		if exists && article.FreshRSSItemID != "" {
			identifier = article.FreshRSSItemID
			log.Printf("  Using FreshRSS Item ID: %s for article %d", identifier, item.ArticleID)
		} else {
			log.Printf("  Warning: No FreshRSS Item ID for article %d, using URL: %s", item.ArticleID, item.ArticleURL)
		}

		switch item.Action {
		case database.SyncActionMarkRead:
			readIDs = append(readIDs, identifier)
		case database.SyncActionMarkUnread:
			unreadIDs = append(unreadIDs, identifier)
		case database.SyncActionStar:
			starIDs = append(starIDs, identifier)
		case database.SyncActionUnstar:
			unstarIDs = append(unstarIDs, identifier)
		}
	}

	// Execute batch operations
	if len(readIDs) > 0 {
		log.Printf("[PushPending] Marking %d articles as read", len(readIDs))
		if err := s.client.MarkAsReadBatch(ctx, readIDs); err != nil {
			log.Printf("[PushPending] ERROR marking as read: %v", err)
			return totalChanges, fmt.Errorf("mark read batch: %w", err)
		}
		log.Printf("[PushPending] Successfully marked %d articles as read", len(readIDs))
		totalChanges += len(readIDs)
	}

	if len(unreadIDs) > 0 {
		if err := s.client.MarkAsUnreadBatch(ctx, unreadIDs); err != nil {
			return totalChanges, fmt.Errorf("mark unread batch: %w", err)
		}
		totalChanges += len(unreadIDs)
	}

	if len(starIDs) > 0 {
		if err := s.client.StarBatch(ctx, starIDs); err != nil {
			return totalChanges, fmt.Errorf("star batch: %w", err)
		}
		totalChanges += len(starIDs)
	}

	if len(unstarIDs) > 0 {
		if err := s.client.UnstarBatch(ctx, unstarIDs); err != nil {
			return totalChanges, fmt.Errorf("unstar batch: %w", err)
		}
		totalChanges += len(unstarIDs)
	}

	// Mark all as synced
	if err := s.db.MarkSynced(itemIDs); err != nil {
		log.Printf("Warning: Failed to mark items as synced: %v", err)
	}

	log.Printf("[PushPending] Successfully synced %d items from queue", totalChanges)

	// Clean up old synced items
	_ = s.db.DeleteOldSyncedItems(7 * 24 * time.Hour)

	return totalChanges, nil
}

// GetPendingCount returns the number of pending sync changes
func (s *BidirectionalSyncService) GetPendingCount() (int, error) {
	return s.db.GetPendingSyncCount()
}

// extractImageURLFromHTML extracts the first image URL from HTML content
// This is used as a fallback for FreshRSS articles that don't have image metadata
func extractImageURLFromHTML(htmlContent string) string {
	if htmlContent == "" {
		return ""
	}

	re := regexp.MustCompile(`<img[^>]+src="([^">]+)"`)
	matches := re.FindStringSubmatch(htmlContent)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// GetFailedItems returns items that failed to sync
func (s *BidirectionalSyncService) GetFailedItems(limit int) ([]database.SyncQueueItem, error) {
	return s.db.GetFailedSyncItems(limit)
}

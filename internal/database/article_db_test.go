package database_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	dbpkg "MrRSS/internal/database"
	"MrRSS/internal/models"
)

func setupDBWithFeed(t *testing.T) *dbpkg.DB {
	t.Helper()
	db := setupTestDB(t)

	// Insert a feed to satisfy foreign key joins
	res, err := db.Exec(`INSERT INTO feeds (title, url, category, is_image_mode, hide_from_timeline) VALUES (?, ?, ?, ?, ?)`, "Test Feed", "https://example.com/feed", "news", 0, 0)
	if err != nil {
		t.Fatalf("insert feed error: %v", err)
	}
	_, _ = res.LastInsertId()
	return db
}

func TestCleanupBySizePreservesUnreadMetadataAndDeletesContentFirst(t *testing.T) {
	db := setupDBWithFeed(t)
	if err := db.SetSetting("max_cache_size_mb", "1"); err != nil {
		t.Fatalf("SetSetting error: %v", err)
	}

	var feedID int64
	if err := db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, "https://example.com/feed").Scan(&feedID); err != nil {
		t.Fatalf("scan feed id: %v", err)
	}

	res, err := db.Exec(
		`INSERT INTO articles (feed_id, title, url, published_at, is_read, is_favorite, is_read_later, unique_id) VALUES (?, ?, ?, ?, 0, 0, 0, ?)`,
		feedID,
		"Old unread article",
		"https://example.com/old",
		time.Now().AddDate(-10, 0, 0),
		"cleanup-preserve-unread",
	)
	if err != nil {
		t.Fatalf("insert article: %v", err)
	}
	articleID, _ := res.LastInsertId()

	if err := db.SetArticleContent(articleID, strings.Repeat("content ", 200000)); err != nil {
		t.Fatalf("SetArticleContent error: %v", err)
	}

	deleted, err := db.CleanupBySize()
	if err != nil {
		t.Fatalf("CleanupBySize error: %v", err)
	}
	if deleted == 0 {
		t.Fatalf("expected cleanup to delete cached content")
	}

	var isRead bool
	if err := db.QueryRow(`SELECT is_read FROM articles WHERE id = ?`, articleID).Scan(&isRead); err != nil {
		t.Fatalf("expected article metadata to remain: %v", err)
	}
	if isRead {
		t.Fatalf("expected unread state to be preserved")
	}

	_, found, err := db.GetArticleContent(articleID)
	if err != nil {
		t.Fatalf("GetArticleContent error: %v", err)
	}
	if found {
		t.Fatalf("expected cached article content to be removed")
	}
}

func TestCleanupReadArticlesOverPerFeedLimitKeepsFeedsIndependent(t *testing.T) {
	db := setupDBWithFeed(t)

	var busyFeedID int64
	if err := db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, "https://example.com/feed").Scan(&busyFeedID); err != nil {
		t.Fatalf("scan busy feed id: %v", err)
	}
	res, err := db.Exec(`INSERT INTO feeds (title, url, category) VALUES (?, ?, ?)`, "Slow Feed", "https://example.com/slow-feed", "blogs")
	if err != nil {
		t.Fatalf("insert slow feed: %v", err)
	}
	slowFeedID, _ := res.LastInsertId()

	now := time.Now()
	busyRows := []struct {
		title       string
		isRead      int
		isFavorite  int
		isReadLater int
	}{
		{"busy-newest", 1, 0, 0},
		{"busy-middle", 1, 0, 0},
		{"busy-unread-protected", 0, 0, 0},
		{"busy-readlater-protected", 1, 0, 1},
		{"busy-favorite-protected", 1, 1, 0},
		{"busy-oldest-read", 1, 0, 0},
	}
	for i, row := range busyRows {
		_, err := db.Exec(
			`INSERT INTO articles (feed_id, title, url, published_at, is_read, is_favorite, is_read_later, unique_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			busyFeedID,
			row.title,
			"https://example.com/"+row.title,
			now.Add(-time.Duration(i)*time.Hour),
			row.isRead,
			row.isFavorite,
			row.isReadLater,
			row.title,
		)
		if err != nil {
			t.Fatalf("insert busy article %q: %v", row.title, err)
		}
	}

	for i := 0; i < 2; i++ {
		title := fmt.Sprintf("slow-%d", i)
		_, err := db.Exec(
			`INSERT INTO articles (feed_id, title, url, published_at, is_read, is_favorite, is_read_later, unique_id) VALUES (?, ?, ?, ?, 1, 0, 0, ?)`,
			slowFeedID,
			title,
			"https://example.com/"+title,
			now.Add(-time.Duration(i)*time.Hour),
			title,
		)
		if err != nil {
			t.Fatalf("insert slow article %q: %v", title, err)
		}
	}

	deleted, err := db.CleanupReadArticlesOverPerFeedLimit(3)
	if err != nil {
		t.Fatalf("CleanupReadArticlesOverPerFeedLimit error: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 old read article deleted, got %d", deleted)
	}

	var deletedCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM articles WHERE title = ?`, "busy-oldest-read").Scan(&deletedCount); err != nil {
		t.Fatalf("count deleted article: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("expected oldest unprotected busy article to be deleted")
	}

	protectedTitles := []string{"busy-unread-protected", "busy-readlater-protected", "busy-favorite-protected"}
	for _, title := range protectedTitles {
		var count int
		if err := db.QueryRow(`SELECT COUNT(*) FROM articles WHERE title = ?`, title).Scan(&count); err != nil {
			t.Fatalf("count protected article %q: %v", title, err)
		}
		if count != 1 {
			t.Fatalf("expected protected article %q to remain", title)
		}
	}

	var slowCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM articles WHERE feed_id = ?`, slowFeedID).Scan(&slowCount); err != nil {
		t.Fatalf("count slow feed articles: %v", err)
	}
	if slowCount != 2 {
		t.Fatalf("expected slow feed to remain untouched, got %d articles", slowCount)
	}
}

func TestGetArticlesWithUnreadFilterCombinesWithFavorites(t *testing.T) {
	db := setupDBWithFeed(t)

	var feedID int64
	if err := db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, "https://example.com/feed").Scan(&feedID); err != nil {
		t.Fatalf("scan feed id: %v", err)
	}

	rows := []struct {
		title      string
		url        string
		isRead     int
		isFavorite int
	}{
		{"Unread favorite", "https://example.com/unread-favorite", 0, 1},
		{"Read favorite", "https://example.com/read-favorite", 1, 1},
		{"Unread normal", "https://example.com/unread-normal", 0, 0},
	}
	for _, row := range rows {
		if _, err := db.Exec(
			`INSERT INTO articles (feed_id, title, url, published_at, is_read, is_favorite, is_read_later, unique_id) VALUES (?, ?, ?, ?, ?, ?, 0, ?)`,
			feedID,
			row.title,
			row.url,
			time.Now(),
			row.isRead,
			row.isFavorite,
			row.url,
		); err != nil {
			t.Fatalf("insert article %q: %v", row.title, err)
		}
	}

	articles, err := db.GetArticlesWithUnreadFilter("favorites", 0, "", false, true, 10, 0)
	if err != nil {
		t.Fatalf("GetArticlesWithUnreadFilter error: %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("expected 1 unread favorite, got %d", len(articles))
	}
	if articles[0].Title != "Unread favorite" || articles[0].IsRead || !articles[0].IsFavorite {
		t.Fatalf("unexpected article returned: %+v", articles[0])
	}
}

func TestSaveAndGetArticle(t *testing.T) {
	db := setupDBWithFeed(t)

	// Get feed id
	var feedID int64
	row := db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, "https://example.com/feed")
	if err := row.Scan(&feedID); err != nil {
		t.Fatalf("scan feed id: %v", err)
	}

	a := &models.Article{
		FeedID:      feedID,
		Title:       "Hello",
		URL:         "https://example.com/article/1",
		ImageURL:    "https://example.com/img.jpg",
		PublishedAt: time.Now(),
	}

	if err := db.SaveArticle(a); err != nil {
		t.Fatalf("SaveArticle error: %v", err)
	}

	// Retrieve by GetArticles
	list, err := db.GetArticles("all", 0, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles error: %v", err)
	}
	if len(list) == 0 {
		t.Fatalf("expected at least one article, got 0")
	}

	// Get by ID
	got, err := db.GetArticleByID(list[0].ID)
	if err != nil {
		t.Fatalf("GetArticleByID error: %v", err)
	}
	if got.URL != a.URL || got.Title != a.Title {
		t.Fatalf("retrieved article mismatch: %+v vs %+v", got, a)
	}
}

func TestMarkReadAndReadLaterAndFavorites(t *testing.T) {
	db := setupDBWithFeed(t)

	// Get feed id
	var feedID int64
	_ = db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, "https://example.com/feed").Scan(&feedID)

	// Insert article
	res, err := db.Exec(`INSERT INTO articles (feed_id, title, url, published_at, is_read, is_favorite, is_read_later) VALUES (?, ?, ?, ?, 0, 0, 0)`, feedID, "A", "u1", time.Now())
	if err != nil {
		t.Fatalf("insert article: %v", err)
	}
	id, _ := res.LastInsertId()

	// Mark read
	if err := db.MarkArticleRead(id, true); err != nil {
		t.Fatalf("MarkArticleRead error: %v", err)
	}

	// Should be marked read and not read later
	var isRead, isReadLater int
	_ = db.QueryRow("SELECT is_read, is_read_later FROM articles WHERE id = ?", id).Scan(&isRead, &isReadLater)
	if isRead != 1 || isReadLater != 0 {
		t.Fatalf("unexpected read/readlater state: %d/%d", isRead, isReadLater)
	}

	// Toggle favorite
	if err := db.ToggleFavorite(id); err != nil {
		t.Fatalf("ToggleFavorite error: %v", err)
	}
	var isFav int
	_ = db.QueryRow("SELECT is_favorite FROM articles WHERE id = ?", id).Scan(&isFav)
	if isFav != 1 {
		t.Fatalf("expected favorite set, got %d", isFav)
	}

	// Toggle read later (will unset since currently 0 -> toggled to 0? ensure it works)
	if err := db.ToggleReadLater(id); err != nil {
		t.Fatalf("ToggleReadLater error: %v", err)
	}
}

func TestUnreadCountsAndMarkAll(t *testing.T) {
	db := setupDBWithFeed(t)

	// Insert feed id
	var feedID int64
	_ = db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, "https://example.com/feed").Scan(&feedID)

	// Insert multiple articles
	for i := 0; i < 5; i++ {
		_, err := db.Exec(`INSERT INTO articles (feed_id, title, url, published_at, is_read, is_hidden) VALUES (?, ?, ?, ?, 0, 0)`, feedID, fmt.Sprintf("t%d", i), fmt.Sprintf("u%d", i), time.Now())
		if err != nil {
			t.Fatalf("insert article: %v", err)
		}
	}

	total, err := db.GetTotalUnreadCount()
	if err != nil {
		t.Fatalf("GetTotalUnreadCount error: %v", err)
	}
	if total < 5 {
		t.Fatalf("expected at least 5 unread, got %d", total)
	}

	byFeed, err := db.GetUnreadCountByFeed(feedID)
	if err != nil {
		t.Fatalf("GetUnreadCountByFeed error: %v", err)
	}
	if byFeed < 1 {
		t.Fatalf("expected unread for feed, got %d", byFeed)
	}

	counts, err := db.GetUnreadCountsForAllFeeds()
	if err != nil {
		t.Fatalf("GetUnreadCountsForAllFeeds error: %v", err)
	}
	if counts[feedID] < 1 {
		t.Fatalf("expected counts map to include feed %d", feedID)
	}

	// Mark all as read for feed
	if err := db.MarkAllAsReadForFeed(feedID); err != nil {
		t.Fatalf("MarkAllAsReadForFeed error: %v", err)
	}
	totalAfter, _ := db.GetTotalUnreadCount()
	if totalAfter != 0 {
		t.Fatalf("expected 0 unread after marking all read, got %d", totalAfter)
	}
}

func TestCleanupOldAndUnimportantAndDBSize(t *testing.T) {
	db := setupDBWithFeed(t)

	// Insert old article (older than default 30 days)
	oldTime := time.Now().AddDate(0, 0, -100)
	var feedID int64
	_ = db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, "https://example.com/feed").Scan(&feedID)

	_, err := db.Exec(`INSERT INTO articles (feed_id, title, url, published_at, is_favorite, is_read_later) VALUES (?, ?, ?, ?, 0, 0)`, feedID, "old", "oldurl", oldTime)
	if err != nil {
		t.Fatalf("insert old article: %v", err)
	}

	// Insert unimportant article (unread, not favorite/readlater)
	_, err = db.Exec(`INSERT INTO articles (feed_id, title, url, published_at, is_read, is_favorite, is_read_later) VALUES (?, ?, ?, ?, 0, 0, 0)`, feedID, "tmp", "u2", time.Now())
	if err != nil {
		t.Fatalf("insert tmp article: %v", err)
	}

	// Cleanup old articles
	deleted, err := db.CleanupOldArticles()
	if err != nil {
		t.Fatalf("CleanupOldArticles error: %v", err)
	}
	if deleted < 1 {
		t.Fatalf("expected at least 1 deleted old article, got %d", deleted)
	}

	// Cleanup unimportant
	del2, err := db.CleanupUnimportantArticles()
	if err != nil {
		t.Fatalf("CleanupUnimportantArticles error: %v", err)
	}
	if del2 < 0 {
		t.Fatalf("unexpected deleted count: %d", del2)
	}

	// DB size
	sz, err := db.GetDatabaseSizeMB()
	if err != nil {
		t.Fatalf("GetDatabaseSizeMB error: %v", err)
	}
	if sz < 0 {
		t.Fatalf("unexpected db size: %f", sz)
	}
}

func TestSaveArticlesBatchContextCancel(t *testing.T) {
	db := setupDBWithFeed(t)

	// Prepare articles
	// determine feed id
	var feedID2 int64
	_ = db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, "https://example.com/feed").Scan(&feedID2)

	articles := []*models.Article{}
	for i := 0; i < 10; i++ {
		articles = append(articles, &models.Article{FeedID: feedID2, Title: "b", URL: "u" + string(rune(i))})
	}

	// Cancel context immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := db.SaveArticles(ctx, articles); err == nil {
		t.Fatalf("expected error due to canceled context")
	}
}

func TestSaveArticlesUpdatePreservesRelatedData(t *testing.T) {
	db := setupDBWithFeed(t)

	var feedID int64
	if err := db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, "https://example.com/feed").Scan(&feedID); err != nil {
		t.Fatalf("scan feed id: %v", err)
	}

	publishedAt := time.Date(2026, time.July, 13, 12, 0, 0, 0, time.UTC)
	article := &models.Article{
		FeedID:                feedID,
		Title:                 "Article with related data",
		URL:                   "https://example.com/article/original",
		PublishedAt:           publishedAt,
		HasValidPublishedTime: true,
	}
	if err := db.SaveArticles(context.Background(), []*models.Article{article}); err != nil {
		t.Fatalf("initial SaveArticles error: %v", err)
	}

	var articleID int64
	if err := db.QueryRow(`SELECT id FROM articles WHERE feed_id = ?`, feedID).Scan(&articleID); err != nil {
		t.Fatalf("scan article id: %v", err)
	}
	if err := db.SetArticleContent(articleID, "cached content"); err != nil {
		t.Fatalf("SetArticleContent error: %v", err)
	}
	sessionID, err := db.CreateChatSession(articleID, "Existing chat")
	if err != nil {
		t.Fatalf("CreateChatSession error: %v", err)
	}
	if _, err := db.CreateChatMessage(sessionID, "user", "Keep this message", ""); err != nil {
		t.Fatalf("CreateChatMessage error: %v", err)
	}

	article.URL = "https://example.com/article/updated"
	if err := db.SaveArticles(context.Background(), []*models.Article{article}); err != nil {
		t.Fatalf("update SaveArticles error: %v", err)
	}

	var updatedArticleID int64
	var updatedURL string
	if err := db.QueryRow(`SELECT id, url FROM articles WHERE feed_id = ?`, feedID).Scan(&updatedArticleID, &updatedURL); err != nil {
		t.Fatalf("scan updated article: %v", err)
	}
	if updatedArticleID != articleID {
		t.Fatalf("article id changed from %d to %d", articleID, updatedArticleID)
	}
	if updatedURL != article.URL {
		t.Fatalf("article URL = %q, want %q", updatedURL, article.URL)
	}

	content, found, err := db.GetArticleContent(articleID)
	if err != nil {
		t.Fatalf("GetArticleContent error: %v", err)
	}
	if !found || content != "cached content" {
		t.Fatalf("article content was not preserved: found=%v content=%q", found, content)
	}

	session, err := db.GetChatSession(sessionID)
	if err != nil {
		t.Fatalf("GetChatSession error: %v", err)
	}
	if session == nil || session.MessageCount != 1 {
		t.Fatalf("chat data was not preserved: session=%+v", session)
	}
}

func TestArticleDeduplicationByUniqueID(t *testing.T) {
	db := setupDBWithFeed(t)

	// Get feed id
	var feedID int64
	row := db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, "https://example.com/feed")
	if err := row.Scan(&feedID); err != nil {
		t.Fatalf("scan feed id: %v", err)
	}

	publishedAt := time.Now()

	// Save same article multiple times with different URLs (should be deduplicated by unique_id)
	article1 := &models.Article{
		FeedID:      feedID,
		Title:       "Test Article",
		URL:         "https://example.com/article/1",
		PublishedAt: publishedAt,
	}

	article2 := &models.Article{
		FeedID:      feedID,
		Title:       "Test Article",                                  // Same title
		URL:         "https://example.com/article/1?utm_source=test", // Different URL
		PublishedAt: publishedAt,                                     // Same time
	}

	// Save first article
	if err := db.SaveArticle(article1); err != nil {
		t.Fatalf("SaveArticle error: %v", err)
	}

	// Try to save the same article again (should be ignored due to unique_id)
	if err := db.SaveArticle(article2); err != nil {
		t.Fatalf("SaveArticle error: %v", err)
	}

	// Verify only one article exists
	articles, err := db.GetArticles("all", feedID, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles error: %v", err)
	}

	if len(articles) != 1 {
		t.Fatalf("expected 1 article after deduplication, got %d", len(articles))
	}

	// Verify the article has the correct unique_id
	if articles[0].Title != "Test Article" {
		t.Fatalf("expected title 'Test Article', got '%s'", articles[0].Title)
	}
}

func TestArticleDifferentTitlesNotDeduplicated(t *testing.T) {
	db := setupDBWithFeed(t)

	// Get feed id
	var feedID int64
	row := db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, "https://example.com/feed")
	if err := row.Scan(&feedID); err != nil {
		t.Fatalf("scan feed id: %v", err)
	}

	publishedAt := time.Now()

	// Save different articles with same feed and time (should NOT be deduplicated)
	article1 := &models.Article{
		FeedID:      feedID,
		Title:       "Article One",
		URL:         "https://example.com/article/1",
		PublishedAt: publishedAt,
	}

	article2 := &models.Article{
		FeedID:      feedID,
		Title:       "Article Two", // Different title
		URL:         "https://example.com/article/2",
		PublishedAt: publishedAt, // Same time
	}

	// Save both articles
	if err := db.SaveArticle(article1); err != nil {
		t.Fatalf("SaveArticle error: %v", err)
	}

	if err := db.SaveArticle(article2); err != nil {
		t.Fatalf("SaveArticle error: %v", err)
	}

	// Verify both articles exist
	articles, err := db.GetArticles("all", feedID, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles error: %v", err)
	}

	if len(articles) != 2 {
		t.Fatalf("expected 2 articles with different titles, got %d", len(articles))
	}
}

// TestSaveArticlesNoPubDateKeepsFirstSeenTime verifies that articles without a
// valid pubDate keep their originally stored published_at on refresh when the
// article is unchanged, instead of being bumped to the current time every save.
func TestSaveArticlesNoPubDateKeepsFirstSeenTime(t *testing.T) {
	db := setupDBWithFeed(t)

	var feedID int64
	if err := db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, "https://example.com/feed").Scan(&feedID); err != nil {
		t.Fatalf("scan feed id: %v", err)
	}

	firstSeen := time.Now()
	article := &models.Article{
		FeedID:                feedID,
		Title:                 "No pubDate article",
		URL:                   "https://example.com/no-pubdate",
		PublishedAt:           firstSeen,
		HasValidPublishedTime: false, // feed item had no pubDate
	}

	// First save: article is inserted with the first-seen time
	if err := db.SaveArticles(context.Background(), []*models.Article{article}); err != nil {
		t.Fatalf("initial SaveArticles error: %v", err)
	}

	// Simulate the next feed refresh: processArticles stamps a NEW time.Now()
	// but the article content is unchanged.
	later := firstSeen.Add(2 * time.Hour)
	article.PublishedAt = later
	if err := db.SaveArticles(context.Background(), []*models.Article{article}); err != nil {
		t.Fatalf("refresh SaveArticles error: %v", err)
	}

	// Only one article must exist (dedup by title+feedID, no date part)
	articles, err := db.GetArticles("all", feedID, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles error: %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}

	stored := articles[0].PublishedAt
	// Must stay at the first-seen time, NOT be bumped to the refresh time
	if diff := stored.Sub(firstSeen); diff < 0 || diff > time.Minute {
		t.Fatalf("published_at bumped to refresh time: got %v, want ≈ %v (diff %v)", stored, firstSeen, diff)
	}
}

// TestSaveArticlesNoPubDateBumpsOnContentChange verifies that an article without
// a valid pubDate DOES move to the new time when its content actually changed
// (e.g. the source entry was edited).
func TestSaveArticlesNoPubDateBumpsOnContentChange(t *testing.T) {
	db := setupDBWithFeed(t)

	var feedID int64
	if err := db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, "https://example.com/feed").Scan(&feedID); err != nil {
		t.Fatalf("scan feed id: %v", err)
	}

	firstSeen := time.Now()
	article := &models.Article{
		FeedID:                feedID,
		Title:                 "Edited no-pubDate article",
		URL:                   "https://example.com/edited/1",
		PublishedAt:           firstSeen,
		OriginalSummary:       "old summary",
		HasValidPublishedTime: false,
	}

	if err := db.SaveArticles(context.Background(), []*models.Article{article}); err != nil {
		t.Fatalf("initial SaveArticles error: %v", err)
	}

	// Next refresh: content changed (summary updated) and processArticles stamped
	// a new time — published_at should move to the new time.
	editedAt := firstSeen.Add(2 * time.Hour)
	article.PublishedAt = editedAt
	article.OriginalSummary = "updated summary"
	if err := db.SaveArticles(context.Background(), []*models.Article{article}); err != nil {
		t.Fatalf("refresh SaveArticles error: %v", err)
	}

	articles, err := db.GetArticles("all", feedID, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles error: %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}

	stored := articles[0].PublishedAt
	if diff := stored.Sub(editedAt); diff < 0 || diff > time.Minute {
		t.Fatalf("changed article time not bumped: got %v, want ≈ %v (diff %v)", stored, editedAt, diff)
	}
}

// TestSaveArticlesValidPubDateStillUpdatesTime guards the CASE expression:
// articles WITH a valid pubDate must still get their real published time on refresh.
func TestSaveArticlesValidPubDateStillUpdatesTime(t *testing.T) {
	db := setupDBWithFeed(t)

	var feedID int64
	if err := db.QueryRow(`SELECT id FROM feeds WHERE url = ?`, "https://example.com/feed").Scan(&feedID); err != nil {
		t.Fatalf("scan feed id: %v", err)
	}

	// Same calendar date so the unique_id (title+feedID+date) matches and the
	// second save hits the ON CONFLICT update path.
	t1 := time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 7, 13, 20, 0, 0, 0, time.UTC)
	article := &models.Article{
		FeedID:                feedID,
		Title:                 "With pubDate article",
		URL:                   "https://example.com/with-pubdate",
		PublishedAt:           t1,
		HasValidPublishedTime: true,
	}

	if err := db.SaveArticles(context.Background(), []*models.Article{article}); err != nil {
		t.Fatalf("initial SaveArticles error: %v", err)
	}

	// Refresh with a new (valid) published time on the same day
	article.PublishedAt = t2
	if err := db.SaveArticles(context.Background(), []*models.Article{article}); err != nil {
		t.Fatalf("refresh SaveArticles error: %v", err)
	}

	articles, err := db.GetArticles("all", feedID, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles error: %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}

	stored := articles[0].PublishedAt
	if diff := stored.Sub(t2); diff < 0 || diff > time.Minute {
		t.Fatalf("valid pubDate not honored on refresh: got %v, want %v (diff %v)", stored, t2, diff)
	}
}

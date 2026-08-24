package article_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"MrRSS/internal/database"
	ff "MrRSS/internal/feed"
	"MrRSS/internal/handlers/article"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
)

func setupHandler(t *testing.T) *core.Handler {
	t.Helper()
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	f := ff.NewFetcher(db)
	return core.NewHandler(db, f, nil, nil)
}

func TestHandleArticles_ListAndImageGallery(t *testing.T) {
	h := setupHandler(t)

	// Add a feed and articles
	feedID, err := h.DB.AddFeed(&models.Feed{Title: "F", URL: "http://x"})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	articles := []*models.Article{
		{FeedID: feedID, Title: "a1", URL: "u1", PublishedAt: time.Now()},
		{FeedID: feedID, Title: "a2", URL: "u2", PublishedAt: time.Now()},
	}
	if err := h.DB.SaveArticles(context.Background(), articles); err != nil {
		t.Fatalf("SaveArticles: %v", err)
	}

	// Call HandleArticles
	req := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
	w := httptest.NewRecorder()
	article.HandleArticles(h, w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Result().StatusCode)
	}
	var got []models.Article
	if err := json.NewDecoder(w.Result().Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) < 2 {
		t.Fatalf("expected >=2 articles, got %d", len(got))
	}

	// Image gallery: mark feed as image mode and add image article
	if err := h.DB.UpdateFeed(feedID, "F", "http://x", "", "", false, "", false, 0, true, "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", 0); err != nil {
		t.Fatalf("UpdateFeed: %v", err)
	}
	imgArticle := &models.Article{FeedID: feedID, Title: "img", URL: "iu", ImageURL: "http://img", PublishedAt: time.Now()}
	if err := h.DB.SaveArticles(context.Background(), []*models.Article{imgArticle}); err != nil {
		t.Fatalf("SaveArticles img: %v", err)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/articles/image_gallery", nil)
	w2 := httptest.NewRecorder()
	article.HandleImageGalleryArticles(h, w2, req2)
	if w2.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 image gallery, got %d", w2.Result().StatusCode)
	}
	var imgs []models.Article
	if err := json.NewDecoder(w2.Result().Body).Decode(&imgs); err != nil {
		t.Fatalf("decode imgs: %v", err)
	}
	if len(imgs) == 0 {
		t.Fatalf("expected image articles, got 0")
	}
}

func TestHandleMarkAllAsRead_EmptyCategoryScopesUncategorized(t *testing.T) {
	h := setupHandler(t)

	uncategorizedFeedID, err := h.DB.AddFeed(&models.Feed{Title: "Uncategorized", URL: "http://uncategorized"})
	if err != nil {
		t.Fatalf("AddFeed uncategorized: %v", err)
	}
	categorizedFeedID, err := h.DB.AddFeed(&models.Feed{Title: "Tech", URL: "http://tech", Category: "Tech"})
	if err != nil {
		t.Fatalf("AddFeed categorized: %v", err)
	}

	uncategorized := &models.Article{
		FeedID:      uncategorizedFeedID,
		Title:       "uncategorized",
		URL:         "http://uncategorized/article",
		PublishedAt: time.Now(),
	}
	categorized := &models.Article{
		FeedID:      categorizedFeedID,
		Title:       "categorized",
		URL:         "http://tech/article",
		PublishedAt: time.Now(),
	}
	if err := h.DB.SaveArticles(context.Background(), []*models.Article{uncategorized, categorized}); err != nil {
		t.Fatalf("SaveArticles: %v", err)
	}

	uncategorizedArticles, err := h.DB.GetArticles("", uncategorizedFeedID, "", true, 10, 0)
	if err != nil || len(uncategorizedArticles) != 1 {
		t.Fatalf("Get uncategorized article: %v", err)
	}
	categorizedArticles, err := h.DB.GetArticles("", categorizedFeedID, "", true, 10, 0)
	if err != nil || len(categorizedArticles) != 1 {
		t.Fatalf("Get categorized article: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/articles/mark-all-read?category=", nil)
	w := httptest.NewRecorder()
	article.HandleMarkAllAsRead(h, w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Result().StatusCode)
	}

	uncategorizedAfter, err := h.DB.GetArticleByID(uncategorizedArticles[0].ID)
	if err != nil {
		t.Fatalf("Get uncategorized article after mark: %v", err)
	}
	categorizedAfter, err := h.DB.GetArticleByID(categorizedArticles[0].ID)
	if err != nil {
		t.Fatalf("Get categorized article after mark: %v", err)
	}
	if !uncategorizedAfter.IsRead {
		t.Fatal("uncategorized article should be marked as read")
	}
	if categorizedAfter.IsRead {
		t.Fatal("categorized article should remain unread")
	}
}

func TestArticleActions_MarkRead_Favorite_Hide_ReadLater(t *testing.T) {
	h := setupHandler(t)
	feedID, _ := h.DB.AddFeed(&models.Feed{Title: "F2", URL: "http://y"})

	a := &models.Article{FeedID: feedID, Title: "act", URL: "u", PublishedAt: time.Now()}
	if err := h.DB.SaveArticles(context.Background(), []*models.Article{a}); err != nil {
		t.Fatalf("SaveArticles: %v", err)
	}
	// fetch saved article id
	arts, err := h.DB.GetArticles("", feedID, "", true, 10, 0)
	if err != nil || len(arts) == 0 {
		t.Fatalf("GetArticles: %v", err)
	}
	id := arts[0].ID

	// Mark unread -> read
	req := httptest.NewRequest(http.MethodPost, "/api/articles/mark-read-sync?id="+fmt.Sprint(id)+"&read=true", nil)
	w := httptest.NewRecorder()
	article.HandleMarkReadWithImmediateSync(h, w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("mark read failed: %d", w.Result().StatusCode)
	}

	// Toggle favorite
	req2 := httptest.NewRequest(http.MethodPost, "/api/articles/toggle-favorite-sync?id="+fmt.Sprint(id), nil)
	w2 := httptest.NewRecorder()
	article.HandleToggleFavoriteWithImmediateSync(h, w2, req2)
	if w2.Result().StatusCode != http.StatusOK {
		t.Fatalf("toggle fav failed: %d", w2.Result().StatusCode)
	}

	// Toggle hide (invalid method GET -> 405)
	req3 := httptest.NewRequest(http.MethodGet, "/api/articles/toggle_hide?id="+fmt.Sprint(id), nil)
	w3 := httptest.NewRecorder()
	article.HandleToggleHideArticle(h, w3, req3)
	if w3.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET hide, got %d", w3.Result().StatusCode)
	}

	// Proper POST hide
	req4 := httptest.NewRequest(http.MethodPost, "/api/articles/toggle_hide?id="+fmt.Sprint(id), nil)
	w4 := httptest.NewRecorder()
	article.HandleToggleHideArticle(h, w4, req4)
	if w4.Result().StatusCode != http.StatusOK {
		t.Fatalf("toggle hide failed: %d", w4.Result().StatusCode)
	}

	// Toggle read later (POST)
	req5 := httptest.NewRequest(http.MethodPost, "/api/articles/toggle_read_later?id="+fmt.Sprint(id), nil)
	w5 := httptest.NewRecorder()
	article.HandleToggleReadLater(h, w5, req5)
	if w5.Result().StatusCode != http.StatusOK {
		t.Fatalf("toggle read later failed: %d", w5.Result().StatusCode)
	}
}

func TestHandleReloadArticleContentClearsOnlyArticleContent(t *testing.T) {
	h := setupHandler(t)
	feedID, err := h.DB.AddFeed(&models.Feed{Title: "Reload Feed", URL: "http://example.com/feed"})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	articleModel := &models.Article{
		FeedID:      feedID,
		Title:       "Reload Article",
		URL:         "http://example.com/article",
		PublishedAt: time.Now(),
	}
	if err := h.DB.SaveArticles(context.Background(), []*models.Article{articleModel}); err != nil {
		t.Fatalf("SaveArticles: %v", err)
	}

	articles, err := h.DB.GetArticles("", feedID, "", true, 10, 0)
	if err != nil || len(articles) == 0 {
		t.Fatalf("GetArticles: %v", err)
	}
	articleID := articles[0].ID
	if err := h.DB.SetArticleContent(articleID, "<p>cached</p>"); err != nil {
		t.Fatalf("SetArticleContent: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/articles/reload-content?id="+fmt.Sprint(articleID), nil)
	w := httptest.NewRecorder()
	article.HandleReloadArticleContent(h, w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Result().StatusCode, w.Body.String())
	}

	if _, found, err := h.DB.GetArticleContent(articleID); err != nil {
		t.Fatalf("GetArticleContent: %v", err)
	} else if found {
		t.Fatalf("expected article content cache to be cleared")
	}

	if _, err := h.DB.GetArticleByID(articleID); err != nil {
		t.Fatalf("article should remain after content reload: %v", err)
	}
}

func TestHandleCleanupArticlesPreservesProtectedArticles(t *testing.T) {
	h := setupHandler(t)
	feedID, err := h.DB.AddFeed(&models.Feed{Title: "Cleanup Feed", URL: "http://example.com/cleanup"})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	articles := []*models.Article{
		{FeedID: feedID, Title: "Delete", URL: "http://example.com/delete", PublishedAt: time.Now()},
		{FeedID: feedID, Title: "Read", URL: "http://example.com/read", PublishedAt: time.Now(), IsRead: true},
		{FeedID: feedID, Title: "Favorite", URL: "http://example.com/favorite", PublishedAt: time.Now(), IsFavorite: true},
		{FeedID: feedID, Title: "Read Later", URL: "http://example.com/read-later", PublishedAt: time.Now(), IsReadLater: true},
	}
	if err := h.DB.SaveArticles(context.Background(), articles); err != nil {
		t.Fatalf("SaveArticles: %v", err)
	}

	savedArticles, err := h.DB.GetArticles("", feedID, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles: %v", err)
	}
	for _, savedArticle := range savedArticles {
		if err := h.DB.SetArticleContent(savedArticle.ID, "cached "+savedArticle.Title); err != nil {
			t.Fatalf("SetArticleContent: %v", err)
		}
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/articles/cleanup", nil)
	getRecorder := httptest.NewRecorder()
	article.HandleCleanupArticles(h, getRecorder, getReq)
	if getRecorder.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected GET cleanup to return 405, got %d", getRecorder.Result().StatusCode)
	}

	postReq := httptest.NewRequest(http.MethodPost, "/api/articles/cleanup", nil)
	postRecorder := httptest.NewRecorder()
	article.HandleCleanupArticles(h, postRecorder, postReq)
	if postRecorder.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected cleanup to return 200, got %d: %s", postRecorder.Result().StatusCode, postRecorder.Body.String())
	}

	var result struct {
		Deleted  int64  `json:"deleted"`
		Articles int64  `json:"articles"`
		Contents int64  `json:"contents"`
		Type     string `json:"type"`
	}
	if err := json.NewDecoder(postRecorder.Result().Body).Decode(&result); err != nil {
		t.Fatalf("decode cleanup response: %v", err)
	}
	if result.Deleted != 1 || result.Articles != 1 || result.Contents != 0 || result.Type != "unimportant" {
		t.Fatalf("unexpected cleanup response: %+v", result)
	}

	remaining, err := h.DB.GetArticles("", feedID, "", false, 10, 0)
	if err != nil {
		t.Fatalf("GetArticles after cleanup: %v", err)
	}
	if len(remaining) != 3 {
		t.Fatalf("expected 3 protected articles, got %d", len(remaining))
	}
	for _, remainingArticle := range remaining {
		if _, found, err := h.DB.GetArticleContent(remainingArticle.ID); err != nil || !found {
			t.Fatalf("expected cached content for %q to remain, found=%v, err=%v", remainingArticle.Title, found, err)
		}
	}
}

func TestHandleExportToObsidian(t *testing.T) {
	h := setupHandler(t)

	// Enable Obsidian integration
	if err := h.DB.SetSetting("obsidian_enabled", "true"); err != nil {
		t.Fatalf("SetSetting obsidian_enabled: %v", err)
	}
	if err := h.DB.SetSetting("obsidian_vault_path", t.TempDir()); err != nil {
		t.Fatalf("SetSetting obsidian_vault_path: %v", err)
	}

	// Add a feed and article
	feedID, err := h.DB.AddFeed(&models.Feed{Title: "Test Feed", URL: "http://example.com"})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	articleModel := &models.Article{
		FeedID:      feedID,
		Title:       "Test Article",
		URL:         "http://example.com/article",
		PublishedAt: time.Now(),
	}
	if err := h.DB.SaveArticles(context.Background(), []*models.Article{articleModel}); err != nil {
		t.Fatalf("SaveArticles: %v", err)
	}

	// Get the article ID
	articles, err := h.DB.GetArticles("", feedID, "", false, 10, 0)
	if err != nil || len(articles) == 0 {
		t.Fatalf("GetArticles: %v", err)
	}
	articleID := articles[0].ID

	// Test export request
	reqBody := fmt.Sprintf(`{"article_id": %d}`, articleID)
	req := httptest.NewRequest(http.MethodPost, "/api/articles/export/obsidian", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	article.HandleExportToObsidian(h, w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("Export failed: %d, body: %s", w.Result().StatusCode, w.Body.String())
	}

	// Verify response
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["success"] != "true" {
		t.Fatalf("Export not successful: %v", response)
	}
}

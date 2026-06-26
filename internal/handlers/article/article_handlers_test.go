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

func TestHandleArticles_JSONContractForFlutterClient(t *testing.T) {
	h := setupHandler(t)

	feedID, err := h.DB.AddFeed(&models.Feed{
		Title:    "Contract Feed",
		URL:      "https://example.com/feed.xml",
		Category: "Tech",
	})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	publishedAt := time.Date(2026, 6, 25, 9, 45, 0, 0, time.UTC)
	articleModel := &models.Article{
		FeedID:          feedID,
		Title:           "Contract Article",
		URL:             "https://example.com/article",
		ImageURL:        "https://example.com/image.png",
		AudioURL:        "https://example.com/audio.mp3",
		VideoURL:        "https://example.com/video.mp4",
		PublishedAt:     publishedAt,
		IsRead:          true,
		IsFavorite:      true,
		IsReadLater:     true,
		Author:          "Writer",
		TranslatedTitle: "Translated Contract Article",
		Summary:         "Short summary",
	}
	if err := h.DB.SaveArticles(context.Background(), []*models.Article{articleModel}); err != nil {
		t.Fatalf("SaveArticles: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/articles?filter=all&feed_id="+fmt.Sprint(feedID)+"&page=1&limit=20", nil)
	w := httptest.NewRecorder()

	article.HandleArticles(h, w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var payload []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected one article, got %d", len(payload))
	}

	got := payload[0]
	assertJSONNumber(t, got, "feed_id", float64(feedID))
	assertJSONString(t, got, "title", "Contract Article")
	assertJSONString(t, got, "url", "https://example.com/article")
	assertJSONString(t, got, "image_url", "https://example.com/image.png")
	assertJSONString(t, got, "audio_url", "https://example.com/audio.mp3")
	assertJSONString(t, got, "video_url", "https://example.com/video.mp4")
	assertJSONBool(t, got, "is_read", true)
	assertJSONBool(t, got, "is_favorite", true)
	assertJSONBool(t, got, "is_hidden", false)
	assertJSONBool(t, got, "is_read_later", true)
	assertJSONString(t, got, "feed_title", "Contract Feed")
	assertJSONString(t, got, "author", "Writer")
	assertJSONString(t, got, "translated_title", "Translated Contract Article")
	assertJSONString(t, got, "summary", "Short summary")
	assertJSONString(t, got, "freshrss_item_id", "")

	published, ok := got["published_at"].(string)
	if !ok || published == "" {
		t.Fatalf("published_at must be a non-empty string, got %#v", got["published_at"])
	}
	parsedPublished, err := time.Parse(time.RFC3339Nano, published)
	if err != nil {
		t.Fatalf("published_at must be RFC3339-compatible for Flutter DateTime parsing, got %q: %v", published, err)
	}
	if !parsedPublished.Equal(publishedAt) {
		t.Fatalf("published_at = %s, want %s", parsedPublished, publishedAt)
	}
}

func TestHandleGetArticleContent_JSONContractForFlutterClient(t *testing.T) {
	h := setupHandler(t)

	feedID, err := h.DB.AddFeed(&models.Feed{
		Title: "Content Feed",
		URL:   "https://example.com/feed.xml",
	})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	articleModel := &models.Article{
		FeedID:      feedID,
		Title:       "Content Article",
		URL:         "https://example.com/article",
		PublishedAt: time.Date(2026, 6, 25, 10, 0, 0, 0, time.UTC),
	}
	if err := h.DB.SaveArticles(context.Background(), []*models.Article{articleModel}); err != nil {
		t.Fatalf("SaveArticles: %v", err)
	}

	articles, err := h.DB.GetArticles("all", feedID, "", false, 10, 0)
	if err != nil || len(articles) != 1 {
		t.Fatalf("GetArticles: len=%d err=%v", len(articles), err)
	}

	content := `<p>Hello <strong>Flutter</strong> reader.</p>`
	if err := h.DB.SetArticleContent(articles[0].ID, content); err != nil {
		t.Fatalf("SetArticleContent: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/articles/content?id="+fmt.Sprint(articles[0].ID), nil)
	w := httptest.NewRecorder()

	article.HandleGetArticleContent(h, w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	assertJSONString(t, payload, "content", content)
	assertJSONString(t, payload, "feed_url", "https://example.com/feed.xml")
	assertJSONBool(t, payload, "cached", true)
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

func TestArticleReadFavorite_JSONContractForFlutterClient(t *testing.T) {
	h := setupHandler(t)
	feedID, err := h.DB.AddFeed(&models.Feed{Title: "Status Feed", URL: "https://example.com/feed.xml"})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}

	articleModel := &models.Article{
		FeedID:      feedID,
		Title:       "Status Article",
		URL:         "https://example.com/status",
		PublishedAt: time.Date(2026, 6, 25, 11, 0, 0, 0, time.UTC),
	}
	if err := h.DB.SaveArticles(context.Background(), []*models.Article{articleModel}); err != nil {
		t.Fatalf("SaveArticles: %v", err)
	}

	articles, err := h.DB.GetArticles("all", feedID, "", true, 10, 0)
	if err != nil || len(articles) != 1 {
		t.Fatalf("GetArticles: len=%d err=%v", len(articles), err)
	}
	articleID := articles[0].ID

	req := httptest.NewRequest(http.MethodPost, "/api/articles/read?id="+fmt.Sprint(articleID)+"&read=true", nil)
	w := httptest.NewRecorder()
	article.HandleMarkReadWithImmediateSync(h, w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected read status 200 OK, got %d", w.Result().StatusCode)
	}

	updated, err := h.DB.GetArticleByID(articleID)
	if err != nil {
		t.Fatalf("GetArticleByID after read: %v", err)
	}
	if !updated.IsRead {
		t.Fatalf("expected article to be marked read")
	}

	req = httptest.NewRequest(http.MethodPost, "/api/articles/read?id="+fmt.Sprint(articleID)+"&read=false", nil)
	w = httptest.NewRecorder()
	article.HandleMarkReadWithImmediateSync(h, w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected unread status 200 OK, got %d", w.Result().StatusCode)
	}

	updated, err = h.DB.GetArticleByID(articleID)
	if err != nil {
		t.Fatalf("GetArticleByID after unread: %v", err)
	}
	if updated.IsRead {
		t.Fatalf("expected article to be marked unread")
	}

	req = httptest.NewRequest(http.MethodPost, "/api/articles/favorite?id="+fmt.Sprint(articleID), nil)
	w = httptest.NewRecorder()
	article.HandleToggleFavoriteWithImmediateSync(h, w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected favorite status 200 OK, got %d", w.Result().StatusCode)
	}

	updated, err = h.DB.GetArticleByID(articleID)
	if err != nil {
		t.Fatalf("GetArticleByID after favorite: %v", err)
	}
	if !updated.IsFavorite {
		t.Fatalf("expected article favorite to be toggled on")
	}
}

func assertJSONString(t *testing.T, object map[string]any, key, want string) {
	t.Helper()

	got, ok := object[key].(string)
	if !ok {
		t.Fatalf("%s must be a string, got %#v", key, object[key])
	}
	if got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}

func assertJSONNumber(t *testing.T, object map[string]any, key string, want float64) {
	t.Helper()

	got, ok := object[key].(float64)
	if !ok {
		t.Fatalf("%s must be a number, got %#v", key, object[key])
	}
	if got != want {
		t.Fatalf("%s = %v, want %v", key, got, want)
	}
}

func assertJSONBool(t *testing.T, object map[string]any, key string, want bool) {
	t.Helper()

	got, ok := object[key].(bool)
	if !ok {
		t.Fatalf("%s must be a bool, got %#v", key, object[key])
	}
	if got != want {
		t.Fatalf("%s = %v, want %v", key, got, want)
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

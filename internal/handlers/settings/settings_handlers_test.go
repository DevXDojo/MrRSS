package settings

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
)

func setupHandlerWithDB(t *testing.T) *core.Handler {
	t.Helper()
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	return core.NewHandler(db, nil, nil, nil)
}

func TestHandleSettings_GET(t *testing.T) {
	h := setupHandlerWithDB(t)

	// Set a custom value
	h.DB.SetSetting("language", "xx-YY")

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()

	HandleSettings(h, w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var data map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if data["language"] != "xx-YY" {
		t.Fatalf("expected language xx-YY, got %s", data["language"])
	}
}

func TestHandleSettings_POST(t *testing.T) {
	h := setupHandlerWithDB(t)

	payload := map[string]string{
		"update_interval":     "15",
		"translation_enabled": "true",
		"deepl_api_key":       "deadbeef",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	HandleSettings(h, w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	// Verify settings saved
	v, _ := h.DB.GetSetting("update_interval")
	if v != "15" {
		t.Fatalf("expected update_interval 15, got %s", v)
	}

	v2, _ := h.DB.GetSetting("translation_enabled")
	if v2 != "true" {
		t.Fatalf("expected translation_enabled true, got %s", v2)
	}

	// Encrypted key should be retrievable via GetEncryptedSetting
	dec, err := h.DB.GetEncryptedSetting("deepl_api_key")
	if err != nil {
		t.Fatalf("GetEncryptedSetting error: %v", err)
	}
	if dec != "deadbeef" {
		t.Fatalf("expected deepl_api_key decrypted to be deadbeef, got %s", dec)
	}
}

func TestHandleSettings_POSTDisablingFreshRSSCleansSyncedData(t *testing.T) {
	h := setupHandlerWithDB(t)

	if err := h.DB.SetSetting("freshrss_enabled", "true"); err != nil {
		t.Fatalf("SetSetting freshrss_enabled: %v", err)
	}

	res, err := h.DB.Exec(`
		INSERT INTO feeds (title, url, is_freshrss_source, freshrss_stream_id)
		VALUES (?, ?, 1, ?)
	`, "FreshRSS Feed", "https://example.com/freshrss.xml", "feed/1")
	if err != nil {
		t.Fatalf("insert FreshRSS feed: %v", err)
	}
	feedID, _ := res.LastInsertId()

	res, err = h.DB.Exec(`
		INSERT INTO articles (feed_id, title, url, published_at, unique_id)
		VALUES (?, ?, ?, datetime('now'), ?)
	`, feedID, "FreshRSS Article", "https://example.com/article", "fresh-article")
	if err != nil {
		t.Fatalf("insert FreshRSS article: %v", err)
	}
	articleID, _ := res.LastInsertId()

	if err := h.DB.SetArticleContent(articleID, "<p>cached</p>"); err != nil {
		t.Fatalf("SetArticleContent: %v", err)
	}
	if err := h.DB.EnqueueSyncChange(articleID, "https://example.com/article", database.SyncActionMarkRead); err != nil {
		t.Fatalf("EnqueueSyncChange: %v", err)
	}

	payload := map[string]string{"freshrss_enabled": "false"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/settings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	HandleSettings(h, w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", resp.StatusCode, w.Body.String())
	}

	assertCount := func(query string, want int, args ...any) {
		t.Helper()
		var got int
		if err := h.DB.QueryRow(query, args...).Scan(&got); err != nil {
			t.Fatalf("count query failed %q: %v", query, err)
		}
		if got != want {
			t.Fatalf("query %q got %d, want %d", query, got, want)
		}
	}

	assertCount("SELECT COUNT(*) FROM feeds WHERE is_freshrss_source = 1", 0)
	assertCount("SELECT COUNT(*) FROM articles WHERE feed_id = ?", 0, feedID)
	assertCount("SELECT COUNT(*) FROM article_contents WHERE article_id = ?", 0, articleID)
	assertCount("SELECT COUNT(*) FROM freshrss_sync_queue", 0)

	enabled, err := h.DB.GetSetting("freshrss_enabled")
	if err != nil {
		t.Fatalf("GetSetting freshrss_enabled: %v", err)
	}
	if enabled != "false" {
		t.Fatalf("expected freshrss_enabled false, got %q", enabled)
	}
}

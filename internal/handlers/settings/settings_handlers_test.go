package settings

import (
	"bytes"
	"encoding/json"
	"errors"
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

type invalidatingTranslator struct{ invalidations int }

func (t *invalidatingTranslator) Translate(text, targetLang string) (string, error) {
	return text, nil
}

func (t *invalidatingTranslator) InvalidateCache() { t.invalidations++ }

func TestTranslationSettingsInvalidateProvider(t *testing.T) {
	for _, tc := range []struct {
		key, value string
		want       int
	}{
		{"proxy_enabled", "true", 1},
		{"google_translate_endpoint", "clients5.google.com", 1},
		{"microsoft_api_key", "test-key", 1},
		{"ai_translation_prompt", "Translate clearly", 1},
		{"theme", "dark", 0},
	} {
		t.Run(tc.key, func(t *testing.T) {
			h := setupHandlerWithDB(t)
			defer h.DB.Close()
			translator := &invalidatingTranslator{}
			h.Translator = translator
			body, _ := json.Marshal(map[string]string{tc.key: tc.value})
			w := httptest.NewRecorder()
			HandleSettings(h, w, httptest.NewRequest(http.MethodPost, "/api/settings", bytes.NewReader(body)))
			if w.Code != http.StatusOK || translator.invalidations != tc.want {
				t.Fatalf("status = %d, invalidations = %d, want %d", w.Code, translator.invalidations, tc.want)
			}
		})
	}
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

func TestHandleSettings_POSTUpdatesSystemStartupIntegration(t *testing.T) {
	h := setupHandlerWithDB(t)

	var enabled bool
	var calls int
	h.SetStartupOnBoot = func(value bool) error {
		enabled = value
		calls++
		return nil
	}

	body, _ := json.Marshal(map[string]string{"startup_on_boot": "true"})
	req := httptest.NewRequest(http.MethodPost, "/api/settings", bytes.NewReader(body))
	w := httptest.NewRecorder()

	HandleSettings(h, w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}
	if calls != 1 || !enabled {
		t.Fatalf("expected startup integration to be enabled once, calls=%d enabled=%v", calls, enabled)
	}
	value, err := h.DB.GetSetting("startup_on_boot")
	if err != nil || value != "true" {
		t.Fatalf("expected startup preference to be saved, value=%q err=%v", value, err)
	}
}

func TestHandleSettings_POSTDoesNotSaveStartupPreferenceWhenIntegrationFails(t *testing.T) {
	h := setupHandlerWithDB(t)
	h.SetStartupOnBoot = func(bool) error { return errors.New("registry unavailable") }

	body, _ := json.Marshal(map[string]string{"startup_on_boot": "true"})
	req := httptest.NewRequest(http.MethodPost, "/api/settings", bytes.NewReader(body))
	w := httptest.NewRecorder()

	HandleSettings(h, w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 Internal Server Error, got %d: %s", w.Code, w.Body.String())
	}
	value, _ := h.DB.GetSetting("startup_on_boot")
	if value == "true" {
		t.Fatal("startup preference was saved even though OS integration failed")
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

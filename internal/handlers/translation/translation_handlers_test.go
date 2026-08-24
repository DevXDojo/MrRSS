package translation

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"MrRSS/internal/database"
	corepkg "MrRSS/internal/handlers/core"
	transpkg "MrRSS/internal/translation"
)

func setupDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	})
	return db
}

func insertTestFeed(t *testing.T, db *database.DB) int64 {
	t.Helper()
	res, err := db.Exec("INSERT INTO feeds (title, url) VALUES (?, ?)", "Test Feed", "https://example.com/feed.xml")
	if err != nil {
		t.Fatalf("insert feed failed: %v", err)
	}
	feedID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("get feed ID failed: %v", err)
	}
	return feedID
}

func TestHandleTranslateText_Success(t *testing.T) {
	db := setupDB(t)
	h := &corepkg.Handler{DB: db, Translator: transpkg.NewMockTranslator()}

	// Use a longer English text that will be reliably detected as English
	body := map[string]string{"text": "This is a test article in English", "target_language": "fr"}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/translate/text", bytes.NewReader(b))
	rr := httptest.NewRecorder()

	HandleTranslateText(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	expected := "[FR] This is a test article in English"
	if resp["translated_text"] != expected {
		t.Fatalf("unexpected translation: got %v, want %v", resp["translated_text"], expected)
	}
}

func TestHandleTranslateArticle_SuccessAndDBUpdate(t *testing.T) {
	db := setupDB(t)
	feedID := insertTestFeed(t, db)

	// insert an article
	res, err := db.Exec("INSERT INTO articles (feed_id, title, url, published_at) VALUES (?, 't', 'u', datetime('now'))", feedID)
	if err != nil {
		t.Fatalf("insert article failed: %v", err)
	}
	id, _ := res.LastInsertId()

	h := &corepkg.Handler{DB: db, Translator: transpkg.NewMockTranslator()}

	// Use a longer English text that will be reliably detected as English
	title := "This is an article title in English"
	body := map[string]interface{}{"article_id": id, "title": title, "target_language": "es"}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/translate/article", bytes.NewReader(b))
	rr := httptest.NewRecorder()

	HandleTranslateArticle(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	expected := "[ES] This is an article title in English"
	if resp["translated_title"] != expected {
		t.Fatalf("unexpected translation: got %v, want %v", resp["translated_title"], expected)
	}

	// verify in DB
	var stored string
	if err := db.QueryRow("SELECT translated_title FROM articles WHERE id = ?", id).Scan(&stored); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if stored != expected {
		t.Fatalf("db value mismatch: got %v, want %v", stored, expected)
	}
}

func TestHandleClearTranslations(t *testing.T) {
	db := setupDB(t)
	feedID := insertTestFeed(t, db)

	// insert an article with translated title
	res, err := db.Exec("INSERT INTO articles (feed_id, title, url, translated_title, published_at) VALUES (?, 't', 'u', 'x', datetime('now'))", feedID)
	if err != nil {
		t.Fatalf("insert article failed: %v", err)
	}
	_, _ = res.LastInsertId()

	h := &corepkg.Handler{DB: db, Translator: transpkg.NewMockTranslator()}

	req := httptest.NewRequest(http.MethodPost, "/translate/clear", nil)
	rr := httptest.NewRecorder()

	HandleClearTranslations(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}

	// verify cleared
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM articles WHERE translated_title != ''").Scan(&count); err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 translations remaining, got %d", count)
	}
}

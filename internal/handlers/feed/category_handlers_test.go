package feed_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	fh "MrRSS/internal/handlers/feed"
	"MrRSS/internal/models"
)

// reuse setupHandler from feed_handlers_test.go

func TestHandleUpdateFeedCategory(t *testing.T) {
	h := setupHandler(t)

	id, err := h.DB.AddFeed(&models.Feed{
		Title:    "example",
		URL:      "http://example.com/feed",
		Category: "old",
		ProxyURL: "http://127.0.0.1:8080",
	})
	if err != nil {
		t.Fatalf("AddFeed error: %v", err)
	}

	body, _ := json.Marshal(map[string]any{"id": id, "category": "  Tech  "})
	req := httptest.NewRequest("POST", "/api/feeds/category", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	fh.HandleUpdateFeedCategory(h, w, req)
	if w.Result().StatusCode != 200 {
		t.Fatalf("expected 200 OK, got %d", w.Result().StatusCode)
	}

	feed, err := h.DB.GetFeedByID(id)
	if err != nil {
		t.Fatalf("GetFeedByID error: %v", err)
	}
	if feed.Category != "Tech" {
		t.Fatalf("expected the category to be trimmed to Tech, got %q", feed.Category)
	}
	// Moving a feed between categories must not disturb its other settings.
	if feed.ProxyURL != "http://127.0.0.1:8080" {
		t.Fatalf("expected the proxy setting to survive, got %q", feed.ProxyURL)
	}
}

func TestHandleUpdateFeedCategoryRejectsBadRequests(t *testing.T) {
	h := setupHandler(t)

	req := httptest.NewRequest("POST", "/api/feeds/category", bytes.NewReader([]byte("notjson")))
	w := httptest.NewRecorder()
	fh.HandleUpdateFeedCategory(h, w, req)
	if w.Result().StatusCode != 400 {
		t.Fatalf("expected 400 for an invalid payload, got %d", w.Result().StatusCode)
	}

	body, _ := json.Marshal(map[string]any{"category": "Tech"})
	missingID := httptest.NewRequest("POST", "/api/feeds/category", bytes.NewReader(body))
	w2 := httptest.NewRecorder()
	fh.HandleUpdateFeedCategory(h, w2, missingID)
	if w2.Result().StatusCode != 400 {
		t.Fatalf("expected 400 when the feed identifier is missing, got %d", w2.Result().StatusCode)
	}
}

package feed_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	fh "MrRSS/internal/handlers/feed"
	"MrRSS/internal/models"
)

func TestHandleAddFeedDuplicateReturnsExistingFeedID(t *testing.T) {
	h := setupHandler(t)

	existingID, err := h.DB.AddFeed(&models.Feed{
		Title: "Existing",
		URL:   "https://example.com/feed.xml",
	})
	if err != nil {
		t.Fatalf("AddFeed error: %v", err)
	}

	payload := map[string]interface{}{
		"url": "https://example.com/feed.xml",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/feeds/add", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	fh.HandleAddFeed(h, w, req)

	if w.Result().StatusCode != http.StatusConflict {
		t.Fatalf("expected 409 Conflict, got %d", w.Result().StatusCode)
	}

	var resp struct {
		Success        bool   `json:"success"`
		Error          string `json:"error"`
		ExistingFeedID int64  `json:"existing_feed_id"`
	}
	if err := json.NewDecoder(w.Result().Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Success {
		t.Fatalf("expected success=false")
	}
	if resp.ExistingFeedID != existingID {
		t.Fatalf("expected existing_feed_id %d, got %d", existingID, resp.ExistingFeedID)
	}
	if resp.Error == "" {
		t.Fatalf("expected non-empty error message")
	}
}

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

func TestHandleUpdateFeed_ValidAndInvalid(t *testing.T) {
	h := setupHandler(t)

	// create a feed
	id, err := h.DB.AddFeed(&models.Feed{Title: "old", URL: "http://example.com/feed"})
	if err != nil {
		t.Fatalf("AddFeed error: %v", err)
	}

	// valid update
	payload := map[string]interface{}{"id": id, "title": "new", "url": "http://example.com/feed", "category": "c"}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/feeds/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	fh.HandleUpdateFeed(h, w, req)
	if w.Result().StatusCode != 200 {
		t.Fatalf("expected 200 OK for valid update, got %d", w.Result().StatusCode)
	}

	// invalid payload
	badReq := httptest.NewRequest("POST", "/api/feeds/update", bytes.NewReader([]byte("notjson")))
	w2 := httptest.NewRecorder()
	fh.HandleUpdateFeed(h, w2, badReq)
	if w2.Result().StatusCode != 400 {
		t.Fatalf("expected 400 for invalid payload, got %d", w2.Result().StatusCode)
	}
}

// The feed listing never sends the IMAP password back, so a client editing a
// newsletter feed has nothing to put in that field. Saving must therefore leave
// the stored password alone rather than clearing it.
func TestHandleUpdateFeed_KeepsIMAPPasswordWhenNoneIsSent(t *testing.T) {
	h := setupHandler(t)

	id, err := h.DB.AddFeed(&models.Feed{
		Title:           "Newsletter",
		URL:             "email://newsletter",
		EmailIMAPServer: "imap.example.com",
		EmailIMAPPort:   993,
		EmailUsername:   "reader",
		EmailPassword:   "secret",
		EmailFolder:     "INBOX",
	})
	if err != nil {
		t.Fatalf("AddFeed error: %v", err)
	}

	payload := map[string]interface{}{
		"id":                id,
		"title":             "Newsletter",
		"url":               "email://newsletter",
		"email_imap_server": "imap.example.com",
		"email_imap_port":   993,
		"email_username":    "reader",
		"email_password":    "",
		"email_folder":      "INBOX",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/feeds/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	fh.HandleUpdateFeed(h, w, req)
	if w.Result().StatusCode != 200 {
		t.Fatalf("expected 200 OK, got %d", w.Result().StatusCode)
	}

	updated, err := h.DB.GetFeedByID(id)
	if err != nil {
		t.Fatalf("GetFeedByID error: %v", err)
	}
	if updated.EmailPassword != "secret" {
		t.Fatalf("expected the stored password to survive, got %q", updated.EmailPassword)
	}
}

func TestHandleUpdateFeed_ReplacesIMAPPasswordWhenOneIsSent(t *testing.T) {
	h := setupHandler(t)

	id, err := h.DB.AddFeed(&models.Feed{
		Title:           "Newsletter",
		URL:             "email://newsletter",
		EmailIMAPServer: "imap.example.com",
		EmailUsername:   "reader",
		EmailPassword:   "secret",
	})
	if err != nil {
		t.Fatalf("AddFeed error: %v", err)
	}

	payload := map[string]interface{}{
		"id":                id,
		"title":             "Newsletter",
		"url":               "email://newsletter",
		"email_imap_server": "imap.example.com",
		"email_username":    "reader",
		"email_password":    "rotated",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/feeds/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	fh.HandleUpdateFeed(h, w, req)
	if w.Result().StatusCode != 200 {
		t.Fatalf("expected 200 OK, got %d", w.Result().StatusCode)
	}

	updated, err := h.DB.GetFeedByID(id)
	if err != nil {
		t.Fatalf("GetFeedByID error: %v", err)
	}
	if updated.EmailPassword != "rotated" {
		t.Fatalf("expected the new password to be stored, got %q", updated.EmailPassword)
	}
}

package feed_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	feedhandlers "MrRSS/internal/handlers/feed"
	"MrRSS/internal/models"
)

func TestCategoryHandlerValidationAndProtection(t *testing.T) {
	h := setupHandler(t)
	_, err := h.DB.AddFeed(&models.Feed{Title: "Synced", URL: "https://example.com/synced", Category: "Synced", IsFreshRSSSource: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		method, body string
		status       int
	}{
		{"GET", `{}`, 405},
		{"POST", `{}`, 400},
		{"POST", `{"action":"dissolve","category":""}`, 400},
		{"POST", `{"action":"other","category":"News"}`, 400},
		{"POST", `{"action":"unsubscribe","category":"Synced"}`, 403},
		{"POST", `{"action":"unsubscribe","category":"Missing"}`, 200},
	} {
		req := httptest.NewRequest(test.method, "/api/feeds/category", strings.NewReader(test.body))
		writer := httptest.NewRecorder()
		feedhandlers.HandleCategory(h, writer, req)
		if writer.Code != test.status {
			t.Errorf("%s %s: got %d, want %d", test.method, test.body, writer.Code, test.status)
		}
	}
}

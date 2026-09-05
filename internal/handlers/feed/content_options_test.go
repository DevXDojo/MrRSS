package feed

import (
	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContentOptionsAPIHidesCookieAndRejectsInvalidSelectors(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	id, err := db.AddFeed(&models.Feed{Title: "Synced feed", URL: "https://example.org/rss", IsFreshRSSSource: true})
	if err != nil {
		t.Fatal(err)
	}
	h := &core.Handler{DB: db}
	endpoint := fmt.Sprintf("/api/feeds/content-options?id=%d", id)
	body := `{"content_selector":"article","remove_selector":".ad","cookie_origin":"https://example.org","cookie":"session=private"}`
	rr := httptest.NewRecorder()
	HandleContentOptions(h, rr, httptest.NewRequest("POST", endpoint, strings.NewReader(body)))
	if rr.Code != 200 || bytes.Contains(rr.Body.Bytes(), []byte("private")) {
		t.Fatalf("save status=%d body=%s", rr.Code, rr.Body.String())
	}
	var saved map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &saved); err != nil {
		t.Fatal(err)
	}
	if saved["has_cookie"] != true || saved["cookie"] != nil {
		t.Fatal("Cookie leaked or missing save indication")
	}
	rr = httptest.NewRecorder()
	HandleContentOptions(h, rr, httptest.NewRequest("POST", endpoint, strings.NewReader(`{"content_selector":"["}`)))
	if rr.Code != 400 {
		t.Fatal("invalid selector accepted", rr.Code)
	}
	options, err := db.GetFeedContentOptions(context.Background(), id)
	if err != nil || options.ContentSelector != "article" {
		t.Fatal("invalid update changed saved settings", err)
	}
	rr = httptest.NewRecorder()
	HandleContentOptions(h, rr, httptest.NewRequest("GET", endpoint, nil))
	if rr.Code != 200 || bytes.Contains(rr.Body.Bytes(), []byte("private")) {
		t.Fatal("read exposed Cookie")
	}
}

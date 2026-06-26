package feed_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	fh "MrRSS/internal/handlers/feed"
	"MrRSS/internal/models"
)

func TestHandleFeeds_ReturnsList(t *testing.T) {
	h := setupHandler(t)

	// add two feeds
	if _, err := h.DB.AddFeed(&models.Feed{Title: "a", URL: "http://x/1"}); err != nil {
		t.Fatalf("add feed: %v", err)
	}
	if _, err := h.DB.AddFeed(&models.Feed{Title: "b", URL: "http://x/2"}); err != nil {
		t.Fatalf("add feed: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/feeds", nil)
	w := httptest.NewRecorder()

	fh.HandleFeeds(h, w, req)
	res := w.Result()
	if res.StatusCode != 200 {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}

	var feeds []models.Feed
	if err := json.NewDecoder(res.Body).Decode(&feeds); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(feeds) < 2 {
		t.Fatalf("expected at least 2 feeds, got %d", len(feeds))
	}
}

func TestHandleFeeds_JSONContractForFlutterClient(t *testing.T) {
	h := setupHandler(t)

	feedID, err := h.DB.AddFeed(&models.Feed{
		Title:            "Contract Feed",
		URL:              "https://example.com/feed.xml",
		Link:             "https://example.com",
		Description:      "Feed used by Flutter contract tests",
		Category:         "Tech",
		ImageURL:         "https://example.com/icon.png",
		Position:         7,
		HideFromTimeline: true,
		ProxyEnabled:     true,
		RefreshInterval:  30,
		IsImageMode:      true,
		EmailAddress:     "newsletter@example.com",
		EmailPassword:    "secret-password",
	})
	if err != nil {
		t.Fatalf("add feed: %v", err)
	}

	tagID, err := h.DB.AddTag(&models.Tag{Name: "Important", Color: "#2563eb", Position: 1})
	if err != nil {
		t.Fatalf("add tag: %v", err)
	}
	if err := h.DB.SetFeedTags(feedID, []int64{tagID}); err != nil {
		t.Fatalf("set feed tags: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/feeds", nil)
	w := httptest.NewRecorder()

	fh.HandleFeeds(h, w, req)
	res := w.Result()
	if res.StatusCode != 200 {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}

	var payload []map[string]any
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if len(payload) != 1 {
		t.Fatalf("expected one feed, got %d", len(payload))
	}

	feed := payload[0]
	assertJSONNumber(t, feed, "id", float64(feedID))
	assertJSONString(t, feed, "title", "Contract Feed")
	assertJSONString(t, feed, "url", "https://example.com/feed.xml")
	assertJSONString(t, feed, "link", "https://example.com")
	assertJSONString(t, feed, "description", "Feed used by Flutter contract tests")
	assertJSONString(t, feed, "category", "Tech")
	assertJSONString(t, feed, "image_url", "https://example.com/icon.png")
	assertJSONNumber(t, feed, "position", 7)
	assertJSONBool(t, feed, "hide_from_timeline", true)
	assertJSONBool(t, feed, "proxy_enabled", true)
	assertJSONNumber(t, feed, "refresh_interval", 30)
	assertJSONBool(t, feed, "is_image_mode", true)
	assertJSONString(t, feed, "email_address", "newsletter@example.com")
	assertJSONMissingOrEmptyString(t, feed, "email_password")
	assertJSONString(t, feed, "last_update_status", "success")

	lastUpdated, ok := feed["last_updated"].(string)
	if !ok || lastUpdated == "" {
		t.Fatalf("last_updated must be a non-empty string, got %#v", feed["last_updated"])
	}
	if _, err := time.Parse(time.RFC3339Nano, lastUpdated); err != nil {
		t.Fatalf("last_updated must be RFC3339-compatible for Flutter DateTime parsing, got %q: %v", lastUpdated, err)
	}

	tags, ok := feed["tags"].([]any)
	if !ok || len(tags) != 1 {
		t.Fatalf("tags must contain one tag, got %#v", feed["tags"])
	}
	tag, ok := tags[0].(map[string]any)
	if !ok {
		t.Fatalf("tag must be an object, got %#v", tags[0])
	}
	assertJSONNumber(t, tag, "id", float64(tagID))
	assertJSONString(t, tag, "name", "Important")
	assertJSONString(t, tag, "color", "#2563eb")
}

func TestHandleRefreshFeed_JSONContractForFlutterClient(t *testing.T) {
	h := setupHandler(t)

	feedID, err := h.DB.AddFeed(&models.Feed{
		Title: "Refresh Feed",
		URL:   "https://example.com/feed.xml",
	})
	if err != nil {
		t.Fatalf("add feed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/feeds/refresh?id="+strconv.FormatInt(feedID, 10), nil)
	w := httptest.NewRecorder()

	fh.HandleRefreshFeed(h, w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	var payload map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload["status"] != "refreshing" {
		t.Fatalf("status = %q, want refreshing", payload["status"])
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

func assertJSONMissingOrEmptyString(t *testing.T, object map[string]any, key string) {
	t.Helper()

	value, ok := object[key]
	if !ok {
		return
	}
	got, ok := value.(string)
	if !ok {
		t.Fatalf("%s must be absent or an empty string, got %#v", key, value)
	}
	if got != "" {
		t.Fatalf("%s must not expose sensitive data, got %q", key, got)
	}
}

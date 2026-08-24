package discovery

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"MrRSS/internal/database"
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
	return core.NewHandler(db, nil, nil, nil)
}

func TestHandleDiscoverBlogs_NotFound(t *testing.T) {
	h := setupHandler(t)

	payload := map[string]int64{"feed_id": 9999}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/discovery/single", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	HandleDiscoverBlogs(h, w, req)
	if w.Result().StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for missing feed, got %d", w.Result().StatusCode)
	}
}

func TestHandleDiscoverBlogs_Success(t *testing.T) {
	h := setupHandler(t)

	// External server that will host the friend's homepage and RSS
	friendRSS := `<?xml version="1.0"?><rss><channel><title>Friend</title><link>/</link><item><title>F1</title><link>/1</link><guid>1</guid></item></channel></rss>`
	var friendSrv *httptest.Server
	friendSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/friend1/rss" {
			w.Header().Set("Content-Type", "application/rss+xml")
			w.Write([]byte(friendRSS))
			return
		}
		// friend homepage: include rel alternate to rss
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><link rel="alternate" type="application/rss+xml" href="` + friendSrv.URL + `/friend1/rss"></head><body>friend</body></html>`))
	}))
	defer friendSrv.Close()

	// Main feed server: provides feed -> homepage -> links page that links to friendSrv
	var mainSrv *httptest.Server
	mainSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/feed":
			// feed with link to homepage
			w.Header().Set("Content-Type", "application/rss+xml")
			w.Write([]byte(`<?xml version="1.0"?><rss><channel><title>Main</title><link>` + mainSrv.URL + `/home</link></channel></rss>`))
		case "/home":
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><body><a href="` + mainSrv.URL + `/links.html">Links</a></body></html>`))
		case "/links.html":
			w.Header().Set("Content-Type", "text/html")
			// include external link to friend server so extractExternalLinks picks it up
			w.Write([]byte(`<html><body><a href="` + friendSrv.URL + `/friend1">Friend Site</a></body></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer mainSrv.Close()

	// Add feed pointing to mainSrv /feed
	feed := &models.Feed{Title: "main", URL: mainSrv.URL + "/feed"}
	feedID, err := h.DB.AddFeed(feed)
	if err != nil {
		t.Fatalf("AddFeed error: %v", err)
	}

	payload := map[string]int64{"feed_id": feedID}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/discovery/single", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	HandleDiscoverBlogs(h, w, req)
	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from discover blogs, got %d", w.Result().StatusCode)
	}

	var discovered []interface{}
	if err := json.NewDecoder(w.Result().Body).Decode(&discovered); err != nil {
		t.Fatalf("failed to decode discovered response: %v", err)
	}
	if len(discovered) == 0 {
		t.Fatalf("expected at least one discovered blog, got 0")
	}

	// Start single discovery (background) and poll progress endpoints
	startReq := httptest.NewRequest(http.MethodPost, "/api/discovery/start", bytes.NewReader(body))
	startReq.Header.Set("Content-Type", "application/json")
	sw := httptest.NewRecorder()
	HandleStartSingleDiscovery(h, sw, startReq)
	if sw.Result().StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted from start, got %d", sw.Result().StatusCode)
	}

	// Give background goroutine a moment
	time.Sleep(100 * time.Millisecond)

	// Query progress
	progReq := httptest.NewRequest(http.MethodGet, "/api/discovery/progress", nil)
	pw := httptest.NewRecorder()
	HandleGetSingleDiscoveryProgress(h, pw, progReq)
	if pw.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from progress, got %d", pw.Result().StatusCode)
	}

	// Clear discovery
	clearReq := httptest.NewRequest(http.MethodPost, "/api/discovery/clear", nil)
	cw := httptest.NewRecorder()
	HandleClearSingleDiscovery(h, cw, clearReq)
	if cw.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from clear, got %d", cw.Result().StatusCode)
	}
}

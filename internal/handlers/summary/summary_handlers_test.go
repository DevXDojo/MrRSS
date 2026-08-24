package summary

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"MrRSS/internal/database"
	"MrRSS/internal/feed"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"

	"github.com/mmcdole/gofeed"
)

func TestHandleSummarizeArticle_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/summary/article", nil)
	rr := httptest.NewRecorder()

	HandleSummarizeArticle(nil, rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleSummarizeArticle_InvalidLength(t *testing.T) {
	payload := []byte(`{"article_id": 1, "length": "bad"}`)
	req := httptest.NewRequest(http.MethodPost, "/summary/article", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	// Use a nil handler pointer; length validation happens before DB access
	HandleSummarizeArticle(nil, rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, rr.Code)
	}
}

// Test successful summarization using the local summarizer and a mocked feed parser.
func TestHandleSummarizeArticle_Success(t *testing.T) {
	// Setup in-memory DB
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db init failed: %v", err)
	}

	// Add feed
	feedID, err := db.AddFeed(&models.Feed{Title: "T", URL: "http://example.com/feed"})
	if err != nil {
		t.Fatalf("AddFeed failed: %v", err)
	}

	// Add article
	art := &models.Article{
		FeedID:      feedID,
		Title:       "A",
		URL:         "http://example.com/article/1",
		PublishedAt: time.Now(),
	}
	if err := db.SaveArticle(art); err != nil {
		t.Fatalf("SaveArticle failed: %v", err)
	}

	// Get the inserted article ID
	var articleID int64
	if err := db.QueryRow("SELECT id FROM articles WHERE url = ?", art.URL).Scan(&articleID); err != nil {
		t.Fatalf("failed to query article id: %v", err)
	}

	// Create a fetcher and replace its parser with a mock that returns the article content
	f := feed.NewFetcher(db)
	// fp is unexported; inject via reflection+unsafe for testing
	mock := &mockParser{items: []*gofeed.Item{{Link: art.URL, Content: "This is a test content. It has multiple sentences. Useful for summarization."}}}
	rv := reflect.ValueOf(f).Elem()
	fpField := rv.FieldByName("fp")
	ptr := reflect.NewAt(fpField.Type(), unsafe.Pointer(fpField.UnsafeAddr())).Elem()
	ptr.Set(reflect.ValueOf(mock))

	h := core.NewHandler(db, f, nil, nil)

	payload := []byte(`{"article_id": ` + fmt.Sprintf("%d", articleID) + `, "length": "short"}`)
	req := httptest.NewRequest(http.MethodPost, "/summary/article", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	HandleSummarizeArticle(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d; body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte("summary")) {
		t.Fatalf("expected response to contain summary, got: %s", rr.Body.String())
	}
}

func TestHandleSummarizeArticle_RSSProviderUsesOriginalSummary(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("failed to create db: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db init failed: %v", err)
	}
	if err := db.SetSetting("summary_provider", "rss"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}

	feedID, err := db.AddFeed(&models.Feed{Title: "T", URL: "http://example.com/feed"})
	if err != nil {
		t.Fatalf("AddFeed failed: %v", err)
	}

	art := &models.Article{
		FeedID:          feedID,
		Title:           "A",
		URL:             "http://example.com/article/1",
		PublishedAt:     time.Now(),
		Summary:         "generated summary cache",
		OriginalSummary: `<p>RSS provided <strong>summary</strong>.</p><script>alert(1)</script>`,
	}
	if err := db.SaveArticle(art); err != nil {
		t.Fatalf("SaveArticle failed: %v", err)
	}

	var articleID int64
	if err := db.QueryRow("SELECT id FROM articles WHERE url = ?", art.URL).Scan(&articleID); err != nil {
		t.Fatalf("failed to query article id: %v", err)
	}

	h := core.NewHandler(db, nil, nil, nil)
	payload := []byte(`{"article_id": ` + fmt.Sprintf("%d", articleID) + `, "length": "short"}`)
	req := httptest.NewRequest(http.MethodPost, "/summary/article", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	HandleSummarizeArticle(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d got %d; body: %s", http.StatusOK, rr.Code, rr.Body.String())
	}
	if !bytes.Contains(rr.Body.Bytes(), []byte("RSS provided")) {
		t.Fatalf("expected RSS original summary, got: %s", rr.Body.String())
	}
	if bytes.Contains(rr.Body.Bytes(), []byte("generated summary cache")) {
		t.Fatalf("expected generated summary cache to be ignored, got: %s", rr.Body.String())
	}
	if bytes.Contains(rr.Body.Bytes(), []byte("<script>")) {
		t.Fatalf("expected unsafe script content to be removed, got: %s", rr.Body.String())
	}
}

// mockParser implements feed.FeedParser
type mockParser struct {
	items []*gofeed.Item
}

func (m *mockParser) ParseURL(url string) (*gofeed.Feed, error) {
	return &gofeed.Feed{Items: m.items}, nil
}

func (m *mockParser) ParseURLWithContext(url string, ctx context.Context) (*gofeed.Feed, error) {
	return &gofeed.Feed{Items: m.items}, nil
}

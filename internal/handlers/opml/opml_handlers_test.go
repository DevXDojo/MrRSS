package opml

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"MrRSS/internal/database"
	"MrRSS/internal/feed"
	corepkg "MrRSS/internal/handlers/core"
)

func TestHandleOPMLImport_RawBody(t *testing.T) {
	xmlData := `<?xml version="1.0"?>
<opml version="1.0">
  <head><title>Test</title></head>
  <body>
    <outline text="Tech" title="Tech">
      <outline type="rss" text="Hacker News" title="Hacker News" xmlUrl="https://news.ycombinator.com/rss" />
    </outline>
    <outline type="rss" text="Go Blog" title="Go Blog" xmlUrl="https://blog.golang.org/feed.atom" />
  </body>
</opml>`

	// Use a real fetcher that writes to an in-memory DB (ImportSubscription uses DB.AddFeed)
	// Use unique shared cache database name to avoid test interference
	db := func() *database.DB {
		db, err := database.NewDB("file:TestHandleOPMLImport_RawBody?mode=memory&cache=shared")
		if err != nil {
			t.Fatalf("failed to create db: %v", err)
		}
		if err := db.Init(); err != nil {
			t.Fatalf("failed to init db: %v", err)
		}
		return db
	}()

	f := feed.NewFetcher(db)
	h := &corepkg.Handler{DB: db, Fetcher: f}

	req := httptest.NewRequest(http.MethodPost, "/opml/import", strings.NewReader(xmlData))
	rr := httptest.NewRecorder()

	HandleOPMLImport(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}

	// Verify feeds were added
	feeds, err := db.GetFeeds()
	if err != nil {
		t.Fatalf("GetFeeds failed: %v", err)
	}
	if len(feeds) != 2 {
		t.Fatalf("expected 2 feeds in DB, got %d", len(feeds))
	}

	// Stop background tasks to prevent interference with other tests
	f.GetTaskManager().Stop()
	f.GetCleanupManager().Stop()
}

func TestHandleOPMLImport_XPathFeed(t *testing.T) {
	xmlData := `<?xml version="1.0"?>
<opml version="1.0">
  <head><title>Test XPath</title></head>
  <body>
    <outline text="CH3NYANG&#39;S BLOG" title="CH3NYANG&#39;S BLOG" type="HTML+XPath" xmlUrl="https://blog.ch3nyang.top/" htmlUrl="" description="" category="" frss:xPathItem="//a[contains(@class, &#39;pagination__item-wrapper&#39;)]" frss:xPathItemTitle=".//div[contains(@class, &#39;pagination__item-title-text&#39;)]" frss:xPathItemContent=".//div[contains(@class, &#39;pagination__item-summary&#39;)]" frss:xPathItemUri="./@href" frss:xPathItemAuthor="" frss:xPathItemTimestamp=".//div[contains(@class, &#39;pagination__item-time&#39;)]" frss:xPathItemTimeFormat="2006-01" frss:xPathItemThumbnail="" frss:xPathItemCategories="" frss:xPathItemUid=""></outline>
  </body>
</opml>`

	// Use a real fetcher that writes to an in-memory DB
	// Use unique shared cache database name to avoid test interference
	db := func() *database.DB {
		db, err := database.NewDB("file:TestHandleOPMLImport_XPathFeed?mode=memory&cache=shared")
		if err != nil {
			t.Fatalf("failed to create db: %v", err)
		}
		if err := db.Init(); err != nil {
			t.Fatalf("failed to init db: %v", err)
		}
		return db
	}()

	f := feed.NewFetcher(db)
	h := &corepkg.Handler{DB: db, Fetcher: f}

	req := httptest.NewRequest(http.MethodPost, "/opml/import", strings.NewReader(xmlData))
	rr := httptest.NewRecorder()

	HandleOPMLImport(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}

	// Verify XPath feed was added
	feeds, err := db.GetFeeds()
	if err != nil {
		t.Fatalf("GetFeeds failed: %v", err)
	}
	if len(feeds) != 1 {
		t.Fatalf("expected 1 feed in DB, got %d", len(feeds))
	}

	feed := feeds[0]
	if feed.Type != "HTML+XPath" {
		t.Errorf("expected feed type 'HTML+XPath', got '%s'", feed.Type)
	}
	if feed.XPathItem != "//a[contains(@class, 'pagination__item-wrapper')]" {
		t.Errorf("expected XPathItem to be set, got '%s'", feed.XPathItem)
	}
	if feed.XPathItemTitle != ".//div[contains(@class, 'pagination__item-title-text')]" {
		t.Errorf("expected XPathItemTitle to be set, got '%s'", feed.XPathItemTitle)
	}
	if feed.XPathItemTimeFormat != "2006-01" {
		t.Errorf("expected XPathItemTimeFormat '2006-01', got '%s'", feed.XPathItemTimeFormat)
	}

	// Stop background tasks to prevent interference with other tests
	f.GetTaskManager().Stop()
	f.GetCleanupManager().Stop()
}

func TestHandleOPMLExport(t *testing.T) {
	// Use unique shared cache database name to avoid test interference
	db := func() *database.DB {
		db, err := database.NewDB("file:TestHandleOPMLExport?mode=memory&cache=shared")
		if err != nil {
			t.Fatalf("failed to create db: %v", err)
		}
		if err := db.Init(); err != nil {
			t.Fatalf("failed to init db: %v", err)
		}
		return db
	}()

	// insert a feed via SQL to keep test simple (provide non-null description and last_updated)
	_, _ = db.Exec("INSERT INTO feeds (title, url, description, last_updated) VALUES (?, ?, ?, datetime('now'))", "F1", "http://f1", "")
	// Sanity-check DB: try GetFeeds before calling handler
	if feeds, err := db.GetFeeds(); err != nil {
		t.Fatalf("GetFeeds before handler failed: %v", err)
	} else if len(feeds) == 0 {
		// continue — handler should still return data
	}

	h := &corepkg.Handler{DB: db}

	req := httptest.NewRequest(http.MethodGet, "/opml/export", nil)
	rr := httptest.NewRecorder()

	HandleOPMLExport(h, rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}

	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/xml") {
		t.Fatalf("expected text/xml content type, got %s", ct)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "http://f1") {
		t.Fatalf("exported OPML missing feed URL: %s", body)
	}
}

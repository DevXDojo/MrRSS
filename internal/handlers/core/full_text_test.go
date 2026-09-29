package core

import (
	"MrRSS/internal/database"
	"MrRSS/internal/models"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func fullTextHandler(t *testing.T) *Handler {
	t.Helper()
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return &Handler{DB: db}
}

func TestFullTextSelectorsLazyImagesAndRedirectBase(t *testing.T) {
	h := fullTextHandler(t)
	source := &models.Feed{Title: "test", URL: "https://example.org/feed"}
	id, err := h.DB.AddFeed(source)
	if err != nil {
		t.Fatal(err)
	}
	source.ID = id
	if err := h.DB.SetFeedContentOptions(context.Background(), id, database.FeedContentOptions{ContentSelector: ".story", RemoveSelector: ".advert"}); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/news/article", 302)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<html><body><nav>navigation</nav><section class="story"><h1>中文标题</h1><p>Selected article.</p><div class="story">Nested selection</div><div class="advert">REMOVE ME</div><img data-src="../photo.jpg" src="data:image/gif;base64,AAA"><img data-srcset="small.jpg 320w, large.jpg 1000w"></section><section class="story">Second selection</section></body></html>`)
	}))
	defer server.Close()
	content, err := h.FetchFullArticleContentContext(context.Background(), server.URL+"/start", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"中文标题", "Selected article", "Second selection", server.URL + "/photo.jpg", server.URL + "/news/large.jpg"} {
		if !strings.Contains(content, want) {
			t.Errorf("missing %q in %s", want, content)
		}
	}
	if strings.Contains(content, "REMOVE ME") || strings.Contains(content, "navigation") || strings.Count(content, "Nested selection") != 1 {
		t.Fatal(content)
	}
}

func TestFullTextReadabilityAndInvalidResponses(t *testing.T) {
	h := fullTextHandler(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/fail" {
			w.WriteHeader(403)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<html><head><title>Article</title></head><body><article><h1>Article</h1><p>"+strings.Repeat("A long article sentence, with useful detail and punctuation. ", 30)+"</p><img data-original='/photo.jpg'></article></body></html>")
	}))
	defer server.Close()
	got, err := h.FetchFullArticleContentContext(context.Background(), server.URL, nil)
	if err != nil || !strings.Contains(got, "useful detail") || !strings.Contains(got, server.URL+"/photo.jpg") {
		t.Fatalf("%s %v", got, err)
	}
	for _, address := range []string{"file:///etc/passwd", server.URL + "/fail"} {
		if _, err := h.FetchFullArticleContentContext(context.Background(), address, nil); err == nil {
			t.Errorf("expected error for %s", address)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := h.FetchFullArticleContentContext(ctx, server.URL, nil); err == nil {
		t.Fatal("canceled fetch succeeded")
	}
}

func TestFullTextNoSelectorMatchIsAnError(t *testing.T) {
	h := fullTextHandler(t)
	id, err := h.DB.AddFeed(&models.Feed{Title: "f", URL: "https://example.org"})
	if err != nil {
		t.Fatal(err)
	}
	if err := h.DB.SetFeedContentOptions(context.Background(), id, database.FeedContentOptions{ContentSelector: ".missing"}); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "<article>Unexpected fallback</article>") }))
	defer server.Close()
	if _, err := h.FetchFullArticleContentContext(context.Background(), server.URL, &models.Feed{ID: id}); err == nil {
		t.Fatal("selector mismatch silently fell back")
	}
}

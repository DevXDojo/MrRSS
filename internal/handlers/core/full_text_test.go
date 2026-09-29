package core

import (
	"MrRSS/internal/database"
	"MrRSS/internal/models"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
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

func TestFullTextReadabilityRestoresLeadImageWithoutDuplicates(t *testing.T) {
	h := fullTextHandler(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		body := strings.Repeat("A long article sentence, with useful detail and punctuation. ", 30)
		if r.URL.Path == "/existing" {
			fmt.Fprintf(w, `<html><head><meta property="og:image" content="/og.jpg"></head><body><article><h1>Article</h1><p>%s</p><img src="/body.jpg"></article></body></html>`, body)
			return
		}
		if r.URL.Path == "/no-lead" {
			fmt.Fprintf(w, `<html><body><article><h1>Article</h1><p>%s</p></article></body></html>`, body)
			return
		}
		fmt.Fprintf(w, `<html><head><meta property="og:image" content="/lead.jpg"></head><body><article><h1>Article</h1><p>%s</p></article></body></html>`, body)
	}))
	defer server.Close()

	content, err := h.FetchFullArticleContentContext(context.Background(), server.URL+"/lead", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(content, `src="`+server.URL+`/lead.jpg"`) || !strings.Contains(content, `referrerpolicy="no-referrer"`) {
		t.Fatalf("lead image was not restored safely: %s", content)
	}

	content, err = h.FetchFullArticleContentContext(context.Background(), server.URL+"/existing", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(content, server.URL+"/body.jpg") || strings.Contains(content, server.URL+"/og.jpg") {
		t.Fatalf("existing article image was duplicated or replaced: %s", content)
	}
	content, err = h.FetchFullArticleContentContext(context.Background(), server.URL+"/no-lead", nil)
	if err != nil || strings.Contains(content, "<img") {
		t.Fatalf("article without a lead image must not use its own URL as an image: %s (%v)", content, err)
	}
}

func TestFullTextReadabilityPromotesDiscuzImages(t *testing.T) {
	h := fullTextHandler(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<html><head><title>Forum article</title><base href="/news/"></head><body><article><h1>Forum article</h1><p>%s</p><ignore_js_op><img class="lazy" style="cursor:pointer" id="aimg_5366234" src="static/image/common/none.gif" onclick="zoom(this, this.getAttribute('zoomfile'))" zoomfile="https://att.huarenjie.com/attachment/forum/202609/29/004010ve8fzabf4ee88kua.jpg" file="https://att.huarenjie.com/attachment/forum/202609/29/004010ve8fzabf4ee88kua.jpg.thumb.jpg"></ignore_js_op><ignore_js_op><img id="aimg_5366235" src="static/image/common/none.gif" data-src="java&#x73;cript:alert(1)" zoomfile="../full.jpg?size=large&amp;type=image" file="/thumb.jpg"></ignore_js_op></article></body></html>`, strings.Repeat("An article with useful details and punctuation, describing a local event. ", 30))
	}))
	defer server.Close()
	content, err := h.FetchFullArticleContentContext(context.Background(), server.URL+"/article", nil)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"https://att.huarenjie.com/attachment/forum/202609/29/004010ve8fzabf4ee88kua.jpg",
		server.URL + "/full.jpg?size=large&type=image",
	}
	if doc.Find("img").Length() != len(want) {
		t.Fatalf("expected both Discuz attachments: %s", content)
	}
	doc.Find("img").Each(func(i int, image *goquery.Selection) {
		if image.AttrOr("src", "") != want[i] || image.AttrOr("referrerpolicy", "") != "no-referrer" {
			t.Errorf("unexpected image %d: %s", i, content)
		}
	})
	for _, unwanted := range []string{"none.gif", "thumb.jpg", "zoomfile", "javascript:", "onclick"} {
		if strings.Contains(content, unwanted) {
			t.Errorf("unexpected %q in full text: %s", unwanted, content)
		}
	}
}

func TestNormalizeArticleImagesPreservesOrdinaryAndSrcsetSources(t *testing.T) {
	base, _ := url.Parse("https://example.org/news/article")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<img src="/ordinary.jpg" srcset="/larger.jpg 2x"><img src="data:image/png;base64,aGVsbG8="><img data-src="javascript:alert(1)" srcset="small.jpg 320w, large.jpg 1000w"><img data-srcset="small.jpg 1x, large.jpg 2x"><img data-src="/data.jpg" zoomfile="/full.jpg" file="/thumb.jpg" class="lazy">`))
	if err != nil {
		t.Fatal(err)
	}
	normalizeArticleImages(doc, base)
	want := []string{
		"https://example.org/ordinary.jpg",
		"data:image/png;base64,aGVsbG8=",
		"https://example.org/news/large.jpg",
		"https://example.org/news/large.jpg",
		"https://example.org/data.jpg",
	}
	doc.Find("img").Each(func(i int, image *goquery.Selection) {
		if got := image.AttrOr("src", ""); got != want[i] {
			t.Errorf("image %d: got %q, want %q", i, got, want[i])
		}
	})
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

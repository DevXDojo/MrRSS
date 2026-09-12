package siyuan

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"MrRSS/internal/models"
)

const testNotebook = "20210817205410-2kvfpfn"
const testDocument = "20210914223645-oj2vnx2"

func TestCreateDocumentProtocol(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/prefix/api/filetree/createDocWithMd" || r.Header.Get("Authorization") != "Token test-token" {
			t.Error("incorrect SiYuan request")
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if payload["notebook"] != testNotebook || payload["path"] != "/MrRSS/Article [1]" || payload["markdown"] != "# Content" {
			t.Error("incorrect document payload")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "data": testDocument})
	}))
	defer server.Close()
	endpoint, err := ParseEndpoint(server.URL + "/prefix/")
	if err != nil {
		t.Fatal(err)
	}
	id, err := CreateDocument(context.Background(), server.Client(), endpoint, "test-token", testNotebook, "/MrRSS/Article [1]", "# Content")
	if err != nil || id != testDocument {
		t.Fatalf("document = %q, error = %v", id, err)
	}
}

func TestCreateDocumentRejectsFailuresAndRedirects(t *testing.T) {
	for _, body := range []string{`{"code":1,"msg":"private details"}`, `{"code":0,"data":""}`, `<html>Sign in</html>`, `{}`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
		endpoint, _ := ParseEndpoint(server.URL)
		_, err := CreateDocument(context.Background(), server.Client(), endpoint, "", testNotebook, "/test", "body")
		server.Close()
		if err == nil || strings.Contains(err.Error(), "private details") {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	forwarded := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { forwarded = true }))
	defer target.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()
	endpoint, _ := ParseEndpoint(redirect.URL)
	if _, err := CreateDocument(context.Background(), redirect.Client(), endpoint, "secret", testNotebook, "/test", "private article"); err == nil || forwarded {
		t.Fatal("followed export redirect")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := CreateDocument(ctx, target.Client(), endpoint, "", testNotebook, "/test", "body"); err == nil {
		t.Fatal("ignored cancellation")
	}
}

func TestExportInputBoundaries(t *testing.T) {
	for _, endpoint := range []string{"", "file:///tmp/notes", "https://user:secret@notes.example", "https://notes.example?token=secret", "https://notes.example/#fragment"} {
		if _, err := ParseEndpoint(endpoint); err == nil {
			t.Errorf("accepted %s", endpoint)
		}
	}
	for _, folder := range []string{"notes", "/notes/../private", "/notes/./private", "/notes\\private", "/notes\nprivate"} {
		if _, err := DocumentPath(folder, "Title", 1); err == nil {
			t.Errorf("accepted %q", folder)
		}
	}
	path, err := DocumentPath("/MrRSS/", "../../One\\Two\nTitle", 42)
	if err != nil || path != "/MrRSS/One Two Title [42]" {
		t.Fatalf("unsafe title path %q: %v", path, err)
	}
	for _, host := range []string{"http://localhost:6806", "http://127.0.0.1:6806", "http://[::1]:6806"} {
		endpoint, _ := ParseEndpoint(host)
		if !IsLoopback(endpoint) {
			t.Errorf("did not bypass proxy for %s", host)
		}
	}
}

func TestArticleMarkdownPreservesContentAndSafeSource(t *testing.T) {
	article := &models.Article{Title: "Article <script>unsafe</script>", URL: "https://example.com/posts/1", FeedTitle: "Feed", PublishedAt: time.Now()}
	markdown, err := ArticleMarkdown(article, "## Heading\n\n**Bold** text")
	if err != nil || !strings.Contains(markdown, "## Heading") || !strings.Contains(markdown, "**Bold**") || !strings.Contains(markdown, article.URL) {
		t.Fatalf("lost Markdown content: %s, %v", markdown, err)
	}
	markdown, err = ArticleMarkdown(article, `<p>Content<img src="/image.jpg" onerror="run()"></p><script>alert(1)</script><a href="javascript:run()">unsafe link</a>`)
	if err != nil || !strings.Contains(markdown, "https://example.com/image.jpg") || strings.Contains(markdown, "javascript:") || strings.Contains(markdown, "onerror=") || strings.Contains(markdown, "alert(1)") {
		t.Fatalf("unsafe export: %s, %v", markdown, err)
	}
}

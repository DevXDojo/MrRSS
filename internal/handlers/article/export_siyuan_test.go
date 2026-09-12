package article_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"MrRSS/internal/handlers/article"
	"MrRSS/internal/models"
)

func TestSiYuanExportUsesEncryptedTokenAndCachedContent(t *testing.T) {
	h := setupHandler(t)
	defer h.DB.Close()
	feedID, err := h.DB.AddFeed(&models.Feed{Title: "Feed", URL: "https://example.com/feed"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := h.DB.Exec(`INSERT INTO articles (feed_id, title, url, published_at) VALUES (?, 'Title', 'https://example.com/post', ?)`, feedID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if err := h.DB.SetArticleContent(id, "<p>Cached body</p>"); err != nil {
		t.Fatal(err)
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "Token test-token" {
			t.Error("token not decrypted")
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if !strings.Contains(payload["markdown"], "Cached body") || !strings.Contains(payload["markdown"], "https://example.com/post") {
			t.Error("article content or source missing")
		}
		_, _ = w.Write([]byte(`{"code":0,"data":"20210914223645-oj2vnx2"}`))
	}))
	defer server.Close()
	for key, value := range map[string]string{"siyuan_enabled": "true", "siyuan_endpoint": server.URL, "siyuan_notebook_id": "20210817205410-2kvfpfn", "proxy_enabled": "true", "proxy_host": "127.0.0.1", "proxy_port": "1"} {
		if err := h.DB.SetSetting(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.DB.SetEncryptedSetting("siyuan_api_token", "test-token"); err != nil {
		t.Fatal(err)
	}
	send := func() *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		article.HandleExportToSiYuan(h, w, httptest.NewRequest(http.MethodPost, "/api/articles/export/siyuan", strings.NewReader(fmt.Sprintf(`{"article_id":%d}`, id))))
		return w
	}
	if w := send(); w.Code != http.StatusOK || calls != 1 {
		t.Fatalf("status %d, calls %d", w.Code, calls)
	}
	if err := h.DB.SetSetting("siyuan_enabled", "false"); err != nil {
		t.Fatal(err)
	}
	if w := send(); w.Code != http.StatusBadRequest || calls != 1 {
		t.Fatalf("disabled export: status %d, calls %d", w.Code, calls)
	}
}

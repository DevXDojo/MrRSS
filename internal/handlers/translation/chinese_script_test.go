package translation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"MrRSS/internal/handlers/core"
	transpkg "MrRSS/internal/translation"
)

func TestChineseTranslationWithoutUpstreamAccess(t *testing.T) {
	var calls atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}))
	defer proxy.Close()
	proxyURL, err := url.Parse(proxy.URL)
	if err != nil {
		t.Fatal(err)
	}
	db := setupDB(t)
	for key, value := range map[string]string{
		"translation_provider": "google", "proxy_enabled": "true", "proxy_type": "http",
		"proxy_host": proxyURL.Hostname(), "proxy_port": proxyURL.Port(),
	} {
		if err := db.SetSetting(key, value); err != nil {
			t.Fatal(err)
		}
	}
	feedID := insertTestFeed(t, db)
	h := &core.Handler{DB: db, Translator: transpkg.NewDynamicTranslatorWithCache(db, db)}

	for _, tc := range []struct {
		text, target, want string
		force              bool
	}{
		{"軟體", "zh", "软体", false},
		{"小众软件：使用技巧", "zh", "小众软件：使用技巧", false},
		{"这是简体中文的第一段。\n这是简体中文的第二段。\n这是简体中文的第三段。\n這是繁體中文的第四段。", "zh", "这是简体中文的第一段。\n这是简体中文的第二段。\n这是简体中文的第三段。\n这是繁体中文的第四段。", false},
		{"⟪0閱讀連結0⟫，保留圖片⟦0⟧。", "zh", "⟪0阅读连结0⟫，保留图片⟦0⟧。", false},
		{"強制轉換繁體文章。", "zh", "强制转换繁体文章。", true},
	} {
		for _, title := range []bool{true, false} {
			for _, cacheOnly := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/title=%v/cache=%v", tc.text, title, cacheOnly), func(t *testing.T) {
					res, err := db.Exec("INSERT INTO articles (feed_id, title, url, published_at) VALUES (?, ?, ?, datetime('now'))", feedID, tc.text, "https://example.org/article")
					if err != nil {
						t.Fatal(err)
					}
					id, err := res.LastInsertId()
					if err != nil {
						t.Fatal(err)
					}
					body, err := json.Marshal(map[string]interface{}{
						"article_id": id, "title": tc.text, "text": tc.text,
						"target_language": tc.target, "cache_only": cacheOnly, "force": tc.force,
					})
					if err != nil {
						t.Fatal(err)
					}
					request := httptest.NewRequest(http.MethodPost, "/translate", bytes.NewReader(body))
					response := httptest.NewRecorder()
					if title {
						HandleTranslateArticle(h, response, request)
					} else {
						HandleTranslateText(h, response, request)
					}
					if response.Code != http.StatusOK {
						t.Fatalf("title=%v, cache_only=%v: status=%d", title, cacheOnly, response.Code)
					}
					var result struct {
						Title string `json:"translated_title"`
						Text  string `json:"translated_text"`
					}
					if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
						t.Fatal(err)
					}
					got := result.Text
					if title {
						got = result.Title
						stored, err := db.GetArticleByID(id)
						if err != nil || stored == nil || stored.TranslatedTitle != tc.want {
							t.Fatalf("title was not persisted correctly: %+v, %v", stored, err)
						}
					}
					if got != tc.want || calls.Load() != 0 {
						t.Fatalf("got %q, upstream requests=%d; want %q offline", got, calls.Load(), tc.want)
					}
				})
			}
		}
	}
}

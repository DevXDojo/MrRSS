package translation

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"MrRSS/internal/handlers/core"
	transpkg "MrRSS/internal/translation"
)

type cacheGuardTranslator struct {
	*transpkg.DynamicTranslator
	calls int
}

func (t *cacheGuardTranslator) Translate(string, string) (string, error) {
	t.calls++
	return "", fmt.Errorf("unexpected upstream translation")
}

func TestCacheOnlyTranslationNeverContactsProvider(t *testing.T) {
	for _, titleRequest := range []bool{false, true} {
		for _, scenario := range []string{"cached", "missing", "other-language", "other-provider", "already-target"} {
			t.Run(fmt.Sprintf("title=%v/%s", titleRequest, scenario), func(t *testing.T) {
				db := setupDB(t)
				feedID := insertTestFeed(t, db)
				text, result := "This is a previously translated article in English", "这是一篇已翻译的文章"
				lang, provider := "zh", "google"
				if scenario == "other-language" {
					lang = "fr"
				}
				if scenario == "other-provider" {
					provider = "microsoft_edge"
				}
				if scenario == "already-target" {
					text = "这是一篇中文文章不需要请求外部翻译服务"
					result = text
				}
				if scenario != "missing" {
					if err := db.SetCachedTranslation(fmt.Sprintf("%x", sha256.Sum256([]byte(text))), text, lang, result, provider); err != nil {
						t.Fatal(err)
					}
				}
				insert, err := db.Exec("INSERT INTO articles (feed_id, title, url, published_at) VALUES (?, ?, 'https://example.org/a', datetime('now'))", feedID, text)
				if err != nil {
					t.Fatal(err)
				}
				id, err := insert.LastInsertId()
				if err != nil {
					t.Fatal(err)
				}
				translator := &cacheGuardTranslator{DynamicTranslator: transpkg.NewDynamicTranslatorWithCache(db, db)}
				h := &core.Handler{DB: db, Translator: translator}
				body, err := json.Marshal(map[string]interface{}{
					"text": text, "title": text, "article_id": id, "target_language": "zh", "cache_only": true,
				})
				if err != nil {
					t.Fatal(err)
				}
				recorder := httptest.NewRecorder()
				request := httptest.NewRequest(http.MethodPost, "/translate", bytes.NewReader(body))
				if titleRequest {
					HandleTranslateArticle(h, recorder, request)
				} else {
					HandleTranslateText(h, recorder, request)
				}
				var response map[string]interface{}
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				if scenario == "cached" || scenario == "already-target" {
					key := "translated_text"
					if titleRequest {
						key = "translated_title"
					}
					if recorder.Code != http.StatusOK || response[key] != result {
						t.Fatalf("unexpected result: %d %+v", recorder.Code, response)
					}
					if titleRequest {
						article, err := db.GetArticleByID(id)
						if err != nil || article.TranslatedTitle != result {
							t.Fatalf("title not stored: %+v, %v", article, err)
						}
					}
				} else if recorder.Code != http.StatusAccepted || response["cache_miss"] != true {
					t.Fatalf("cache miss reported as success: %d %+v", recorder.Code, response)
				}
				if translator.calls != 0 {
					t.Fatalf("contacted upstream %d times", translator.calls)
				}
			})
		}
	}
}

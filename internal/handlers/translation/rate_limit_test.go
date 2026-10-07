package translation

import (
	"MrRSS/internal/ai"
	"MrRSS/internal/handlers/core"
	translator "MrRSS/internal/translation"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type limitedListTranslator struct{ calls int }

func (t *limitedListTranslator) Translate(text, target string) (string, error) {
	t.calls++
	if t.calls == 1 {
		return "translated first item", nil
	}
	return "", &translator.RateLimitError{RetryAfter: 90 * time.Second}
}

func TestLimitedListsDoNotReturnPartialSuccess(t *testing.T) {
	for _, provider := range []string{"google", "ai"} {
		for _, title := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/title=%v", provider, title), func(t *testing.T) {
				db := setupDB(t)
				if err := db.SetSetting("translation_provider", provider); err != nil {
					t.Fatal(err)
				}
				feedID := insertTestFeed(t, db)
				text := "- First English list item\n- Second English list item\n- Third English list item"
				result, err := db.Exec("INSERT INTO articles (feed_id, title, url, published_at) VALUES (?, ?, 'https://example.org/article', datetime('now'))", feedID, text)
				if err != nil {
					t.Fatal(err)
				}
				id, _ := result.LastInsertId()
				limited := &limitedListTranslator{}
				h := &core.Handler{DB: db, Translator: limited, AITracker: ai.NewUsageTracker(db)}
				body, _ := json.Marshal(map[string]interface{}{
					"article_id": id, "title": text, "text": text, "target_language": "zh", "force": true,
				})
				request := httptest.NewRequest(http.MethodPost, "/translate", bytes.NewReader(body))
				response := httptest.NewRecorder()
				if title {
					HandleTranslateArticle(h, response, request)
				} else {
					HandleTranslateText(h, response, request)
				}
				if response.Code != 429 || response.Header().Get("Retry-After") != "90" || limited.calls != 2 {
					t.Fatalf("expected retryable limit, got status=%d delay=%s calls=%d", response.Code, response.Header().Get("Retry-After"), limited.calls)
				}
				article, err := db.GetArticleByID(id)
				if err != nil || article.TranslatedTitle != "" {
					t.Fatalf("partial translation was persisted: %v, %v", article, err)
				}
			})
		}
	}
}

func TestTranslationRateLimitResponse(t *testing.T) {
	w := httptest.NewRecorder()
	translationError(w, fmt.Errorf("provider: %w", &translator.RateLimitError{RetryAfter: 90 * time.Second}))
	if w.Code != 429 || w.Header().Get("Retry-After") != "90" {
		t.Fatalf("status=%d retry=%s", w.Code, w.Header().Get("Retry-After"))
	}
}

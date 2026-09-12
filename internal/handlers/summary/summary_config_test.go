package summary

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"MrRSS/internal/ai"
	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
)

func TestSummaryUsesSelectedConfigurationAndDoesNotCacheFallback(t *testing.T) {
	for _, tc := range []struct {
		name          string
		profile, fail bool
	}{
		{"profile headers", true, false},
		{"legacy headers", false, false},
		{"transient failure", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := database.NewDB(":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if err := db.Init(); err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-Summary-Key") != "selected" {
					t.Error("summary did not use the selected configuration headers")
				}
				if tc.fail {
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte(`{"error":{"message":"unavailable"}}`))
					return
				}
				_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"AI generated summary."}}]}`))
			}))
			defer server.Close()
			headers := `{"X-Summary-Key":"selected"}`
			globalHeaders := headers
			if tc.profile {
				globalHeaders = `{"X-Summary-Key":"wrong-global-value"}`
				_, err := db.CreateAIProfile(&models.AIProfile{Name: "Summary", Endpoint: server.URL + "/v1/chat/completions", Model: "test", CustomHeaders: headers, IsDefault: true})
				if err != nil {
					t.Fatal(err)
				}
			}
			for key, value := range map[string]string{"summary_provider": "ai", "ai_endpoint": server.URL + "/v1/chat/completions", "ai_model": "test", "ai_custom_headers": globalHeaders} {
				if err := db.SetSetting(key, value); err != nil {
					t.Fatal(err)
				}
			}
			feedID, err := db.AddFeed(&models.Feed{Title: "Feed", URL: "https://feed.example/rss"})
			if err != nil {
				t.Fatal(err)
			}
			article := &models.Article{FeedID: feedID, Title: "Article", URL: "https://feed.example/article", PublishedAt: time.Now()}
			if err := db.SaveArticle(article); err != nil {
				t.Fatal(err)
			}
			var articleID int64
			if err := db.QueryRow("SELECT id FROM articles WHERE url = ?", article.URL).Scan(&articleID); err != nil {
				t.Fatal(err)
			}
			h := core.NewHandler(db, nil, nil, ai.NewProfileProvider(db))
			body, _ := json.Marshal(map[string]interface{}{"article_id": articleID, "content": strings.Repeat("This article describes an important scientific discovery with useful background and detailed explanations. ", 5)})
			w := httptest.NewRecorder()
			HandleSummarizeArticle(h, w, httptest.NewRequest(http.MethodPost, "/api/articles/summarize", bytes.NewReader(body)))
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d: %s", w.Code, w.Body.String())
			}
			stored, err := db.GetArticleByID(articleID)
			if err != nil {
				t.Fatal(err)
			}
			if tc.fail {
				if stored.Summary != "" || !strings.Contains(w.Body.String(), `"used_fallback":true`) {
					t.Fatalf("fallback should be temporary: summary = %q, response = %s", stored.Summary, w.Body.String())
				}
			} else if stored.Summary != "AI generated summary." {
				t.Fatalf("successful summary not saved: %q", stored.Summary)
			}
		})
	}
}

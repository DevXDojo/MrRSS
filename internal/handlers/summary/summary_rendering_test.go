package summary

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
)

func TestSummarySanitizesStoredAndRSSHTML(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	feedID, err := db.AddFeed(&models.Feed{Title: "Feed", URL: "https://example.org/feed"})
	if err != nil {
		t.Fatal(err)
	}
	content := `<p>Useful <strong>fact</strong>.</p><img src="https://example.org/image" onerror=alert(1)><a href="javascript:alert(1)">bad link</a><iframe src="https://evil.example/"></iframe>`
	if err := db.SaveArticle(&models.Article{FeedID: feedID, Title: "Article", URL: "https://example.org/article", PublishedAt: time.Now(), OriginalSummary: content}); err != nil {
		t.Fatal(err)
	}
	var id int64
	if err := db.QueryRow("SELECT id FROM articles WHERE url = ?", "https://example.org/article").Scan(&id); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateArticleSummary(id, content); err != nil {
		t.Fatal(err)
	}
	h := core.NewHandler(db, nil, nil, nil)
	for _, provider := range []string{"local", "rss"} {
		t.Run(provider, func(t *testing.T) {
			if err := db.SetSetting("summary_provider", provider); err != nil {
				t.Fatal(err)
			}
			body, _ := json.Marshal(map[string]any{"article_id": id})
			w := httptest.NewRecorder()
			HandleSummarizeArticle(h, w, httptest.NewRequest(http.MethodPost, "/api/summarize", bytes.NewReader(body)))
			var result struct {
				HTML string `json:"html"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || w.Code != http.StatusOK {
				t.Fatalf("response: %d %s", w.Code, w.Body.String())
			}
			if !strings.Contains(result.HTML, "<strong>fact</strong>") {
				t.Fatalf("lost useful formatting: %s", result.HTML)
			}
			for _, unsafe := range []string{"onerror", "javascript:", "<iframe"} {
				if strings.Contains(result.HTML, unsafe) {
					t.Fatalf("unsafe %s in %s", unsafe, result.HTML)
				}
			}
		})
	}
}

package handlers

import (
	"MrRSS/internal/ai"
	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestReadingReportPreviewGenerationAndFailures(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	feedID, err := db.AddFeed(&models.Feed{Title: "News", URL: "https://example.org/rss"})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveArticle(&models.Article{FeedID: feedID, Title: "Launch", URL: "https://example.org/article", OriginalSummary: "RSS fallback", PublishedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	var articleID int64
	if err := db.QueryRow("SELECT id FROM articles WHERE url = ?", "https://example.org/article").Scan(&articleID); err != nil {
		t.Fatal(err)
	}
	if err := db.SetArticleContent(articleID, "<p>New release cuts latency by 20%.</p>"); err != nil {
		t.Fatal(err)
	}
	calls := 0
	status := 200
	badCitation := false
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("X-Report") != "configured" {
			t.Error("lost profile headers")
		}
		var payload struct {
			Messages []map[string]string `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if len(payload.Messages) != 2 || !strings.Contains(payload.Messages[1]["content"], "20%") || !strings.Contains(payload.Messages[1]["content"], "latency") {
			t.Error("missing article or focus")
		}
		w.WriteHeader(status)
		if status != 200 {
			fmt.Fprint(w, "private provider error")
			return
		}
		id := 1
		if badCitation {
			id = 999
		}
		text := fmt.Sprintf(`{"overview":"A faster release","topics":[{"title":"Performance","summary":"Latency fell 20%%.","source_ids":[%d]}],"reading_order":[],"caveats":[]}`, id)
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": text}}}})
	}))
	defer provider.Close()
	profileID, err := db.CreateAIProfile(&models.AIProfile{Name: "Report", Endpoint: provider.URL + "/v1/chat/completions", Model: "fixture", CustomHeaders: `{"X-Report":"configured"}`})
	if err != nil {
		t.Fatal(err)
	}
	h := core.NewHandler(db, nil, nil, ai.NewProfileProvider(db))
	h.AITracker.SetMinInterval(0)
	request := func(preview bool, ids []int64, profile int64, ctx context.Context) *httptest.ResponseRecorder {
		body, _ := json.Marshal(ReadingReportRequest{ArticleIDs: ids, ProfileID: profile, Focus: "latency"})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/ai/reading-report", bytes.NewReader(body)).WithContext(ctx)
		HandleReadingReport(h, w, r, preview)
		return w
	}
	w := request(true, []int64{articleID}, profileID, context.Background())
	if w.Code != 200 || calls != 0 || !strings.Contains(w.Body.String(), `"kind":"cached"`) {
		t.Fatalf("preview: %d %s calls=%d", w.Code, w.Body.String(), calls)
	}
	w = request(false, []int64{articleID}, profileID, context.Background())
	if w.Code != 200 || calls != 1 || !strings.Contains(w.Body.String(), `"source_ids":[1]`) {
		t.Fatalf("generate: %d %s", w.Code, w.Body.String())
	}
	article, err := db.GetArticleByID(articleID)
	if err != nil || article.IsRead || article.Summary != "" {
		t.Fatal("report changed article state")
	}
	badCitation = true
	w = request(false, []int64{articleID}, profileID, context.Background())
	if w.Code != 502 || !strings.Contains(w.Body.String(), "invalid_response") {
		t.Fatalf("invented reference accepted: %s", w.Body.String())
	}
	badCitation = false
	status = 429
	w = request(false, []int64{articleID}, profileID, context.Background())
	if w.Code != 429 || !strings.Contains(w.Body.String(), "rate_limited") || strings.Contains(w.Body.String(), "private") {
		t.Fatalf("unsafe error: %s", w.Body.String())
	}
	before := calls
	for _, ids := range [][]int64{nil, {articleID, articleID}, {9999}} {
		if request(false, ids, profileID, context.Background()).Code != 400 {
			t.Fatal("invalid selection accepted")
		}
	}
	if request(false, []int64{articleID}, 9999, context.Background()).Code != 400 {
		t.Fatal("unknown profile accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request(false, []int64{articleID}, profileID, ctx)
	if calls != before {
		t.Fatal("invalid or cancelled request reached provider")
	}
}

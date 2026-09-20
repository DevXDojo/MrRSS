package core

import (
	"MrRSS/internal/adfilter"
	"MrRSS/internal/ai"
	"MrRSS/internal/models"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestAdFilterUsesExistingSelectedProfileAndKeepsSafeSource(t *testing.T) {
	h := fullTextHandler(t)
	h.AITracker = ai.NewUsageTracker(h.DB)
	h.AITracker.SetMinInterval(0)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["model"] != "configured-model" || r.Header.Get("Authorization") != "Bearer fixture-key" {
			t.Error("existing profile not used")
		}
		answer := `{"decisions":[{"id":"b2","category":"advertisement","confidence":0.99,"evidence":"Buy Acme Cloud now and save 40%"}]}`
		json.NewEncoder(w).Encode(map[string]interface{}{"choices": []interface{}{map[string]interface{}{"message": map[string]string{"content": answer}}}})
	}))
	defer server.Close()
	profile := &models.AIProfile{Name: "Ad review fixture", Endpoint: server.URL + "/v1/chat/completions", Model: "configured-model", APIKey: "fixture-key"}
	id, err := h.DB.CreateAIProfile(profile)
	if err != nil {
		t.Fatal(err)
	}
	h.DB.SetSetting("ai_ad_filter_profile_id", strconv.FormatInt(id, 10))
	source := `<h1>Database planning</h1><p>` + strings.Repeat("Measure query performance with representative workloads before adding indexes. ", 6) + `</p><p>Sponsored: Buy Acme Cloud now and save 40%.</p><script>alert(1)</script>`
	result, err := h.AnalyzeArticleAds(context.Background(), source, "https://example.org/article", adfilter.Options{Enabled: true, AIEnabled: true})
	if err != nil || calls != 2 || len(result.Decisions) != 1 || strings.Contains(result.Content, "Buy Acme") || strings.Contains(result.Content, "<script") {
		t.Fatalf("calls=%d result=%+v err=%v", calls, result, err)
	}
	_, err = h.AnalyzeArticleAds(context.Background(), source, "https://example.org/article", adfilter.Options{Enabled: true, AIEnabled: true})
	if err != nil || calls != 2 {
		t.Fatal("cached analysis made network request", err)
	}
	usage, err := h.AITracker.GetCurrentUsage()
	if err != nil || usage <= 0 {
		t.Fatal("AI analysis did not use shared accounting", err)
	}
	h.DB.SetSetting("ai_usage_limit", "1")
	_, err = h.AnalyzeArticleAds(context.Background(), source+"<p>Changed article</p>", "https://example.org/article", adfilter.Options{Enabled: true, AIEnabled: true})
	if err == nil || err.Error() != "ai_limit" || calls != 2 {
		t.Fatal("shared limit not honored", err, calls)
	}
}

package summary

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAISummaryUsesReadableEvidenceAndRejectsEmptyAnswers(t *testing.T) {
	answer := "Key result: latency fell by 20%."
	custom := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Messages []map[string]string `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Error(err)
		}
		if len(payload.Messages) != 2 {
			t.Error("missing prompt")
			return
		}
		if strings.Contains(payload.Messages[1]["content"], "doEvil") || !strings.Contains(payload.Messages[1]["content"], "20% &") {
			t.Error("HTML noise or entity not handled")
		}
		if custom && payload.Messages[0]["content"] != "Custom format" {
			t.Error("overwrote custom prompt")
		}
		if !custom && !strings.Contains(payload.Messages[0]["content"], "uncertainty") {
			t.Error("legacy default prompt not upgraded")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": answer}}}})
	}))
	defer server.Close()
	s := NewAISummarizer("", server.URL+"/v1/chat/completions", "fixture")
	defer s.Close()
	s.SetSystemPrompt("You are a summarizer. Generate a concise summary of the given text. Output ONLY the summary, nothing else.")
	content := "<script>doEvil()</script><style>doEvil</style><p>" + strings.Repeat("Latency fell 20% &amp; reliability improved. ", 10) + "</p>"
	if result, err := s.SummarizeContext(context.Background(), content, Medium); err != nil || result.Summary != answer {
		t.Fatal(result, err)
	}
	custom = true
	s.SetSystemPrompt("Custom format")
	answer = "<think>Only reasoning</think>"
	if _, err := s.SummarizeContext(context.Background(), content, Medium); err == nil {
		t.Fatal("thinking-only answer accepted")
	}
}

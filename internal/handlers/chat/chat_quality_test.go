package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"MrRSS/internal/ai"
	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
)

func TestChatContextRetainsArticleOnFollowupAndResume(t *testing.T) {
	messages := []ChatMessage{}
	for i := 0; i < 13; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		messages = append(messages, ChatMessage{Role: role, Content: fmt.Sprint(i)})
	}
	original := append([]ChatMessage(nil), messages...)
	for _, first := range []bool{true, false} {
		got := optimizeChatContext(messages, "标题", "https://example.org/article", "原文依据", first)
		if len(got) > 11 || got[0].Role != "system" || !strings.Contains(got[0].Content, "原文依据") || !strings.Contains(got[0].Content, "https://example.org/article") {
			t.Fatalf("missing article context: %+v", got)
		}
		if got[1].Role != "user" || got[len(got)-1].Content != "12" {
			t.Fatalf("lost latest user question or orphan reply: %+v", got)
		}
	}
	if !reflect.DeepEqual(messages, original) {
		t.Fatal("mutated conversation history")
	}
	got := optimizeChatContext(messages[3:], "Title", "", "Evidence", false)
	if got[1].Role != "user" {
		t.Fatal("frontend-trimmed history starts with orphan assistant")
	}
}

func TestChatUsesConfiguredHeadersAndSafeErrors(t *testing.T) {
	for _, tc := range []struct {
		name    string
		profile bool
		status  int
		content string
		code    string
	}{
		{"legacy", false, 200, "Answer", ""},
		{"selected profile", true, 200, "Answer", ""},
		{"authentication", true, 401, "private provider detail", ai.ErrorCodeAuthenticationFailed},
		{"rate limit", false, 429, "private provider detail", ai.ErrorCodeRateLimited},
		{"thinking without answer", true, 200, "<think>private reasoning</think>", ai.ErrorCodeInvalidResponse},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := database.NewDB(":memory:")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { db.Close() })
			if err := db.Init(); err != nil {
				t.Fatal(err)
			}
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Header.Get("X-Chat-Key") != "selected" || r.Header.Get("Authorization") != "Bearer test-key" {
					t.Error("lost selected headers or credentials")
				}
				var payload struct {
					Messages []ChatMessage `json:"messages"`
				}
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Error(err)
				}
				if len(payload.Messages) == 0 || !strings.Contains(payload.Messages[0].Content, "Article evidence") {
					t.Error("follow-up lost article")
				}
				w.WriteHeader(tc.status)
				if tc.status != 200 {
					fmt.Fprint(w, tc.content)
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": tc.content}}}})
			}))
			defer server.Close()
			for key, value := range map[string]string{
				"ai_chat_enabled": "true", "ai_chat_save_history": "false", "ai_endpoint": server.URL + "/v1/chat/completions", "ai_model": "test", "ai_custom_headers": `{"X-Chat-Key":"selected"}`,
			} {
				if err := db.SetSetting(key, value); err != nil {
					t.Fatal(err)
				}
			}
			if err := db.SetEncryptedSetting("ai_api_key", "test-key"); err != nil {
				t.Fatal(err)
			}
			var profileID int64
			if tc.profile {
				profileID, err = db.CreateAIProfile(&models.AIProfile{Name: "Chat", Endpoint: server.URL + "/v1/chat/completions", Model: "test", APIKey: "test-key", CustomHeaders: `{"X-Chat-Key":"selected"}`})
				if err != nil {
					t.Fatal(err)
				}
				if err := db.SetSetting("ai_custom_headers", `{"X-Chat-Key":"wrong"}`); err != nil {
					t.Fatal(err)
				}
			}
			h := core.NewHandler(db, nil, nil, ai.NewProfileProvider(db))
			body, _ := json.Marshal(ChatRequest{ProfileID: profileID, ArticleContent: "Article evidence", Messages: []ChatMessage{{Role: "user", Content: "Why?"}}})
			w := httptest.NewRecorder()
			HandleAIChat(h, w, httptest.NewRequest(http.MethodPost, "/api/ai-chat", bytes.NewReader(body)))
			if calls != 1 {
				t.Fatalf("provider calls = %d", calls)
			}
			if tc.code == "" {
				if w.Code != 200 || !strings.Contains(w.Body.String(), `"response":"Answer"`) {
					t.Fatalf("unexpected response: %d %s", w.Code, w.Body.String())
				}
			} else {
				var payload chatErrorResponse
				if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
					t.Fatal(err)
				}
				if payload.ErrorCode != tc.code || w.Code != ai.UserFacingErrorForCode(tc.code).HTTPStatus || strings.Contains(w.Body.String(), "private") {
					t.Fatalf("unexpected error: %d %s", w.Code, w.Body.String())
				}
			}
		})
	}
}

func TestChatHTMLRemovesActiveContent(t *testing.T) {
	got := renderChatHTML("## Evidence\n\n**Important**\n\n<img src=https://example.org/image onerror=alert(1)>\n<script>\nalert(1)\n</script>\n<a href=javascript:alert(1)>bad</a>\n")
	for _, bad := range []string{"onerror", "<script", "javascript:"} {
		if strings.Contains(got, bad) {
			t.Errorf("unsafe %s: %s", bad, got)
		}
	}
	for _, want := range []string{"<h2", "<strong>Important</strong>"} {
		if !strings.Contains(got, want) {
			t.Errorf("lost formatting %s: %s", want, got)
		}
	}
}

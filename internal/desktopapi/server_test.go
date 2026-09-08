package desktopapi

import (
	"MrRSS/internal/ai"
	"MrRSS/internal/database"
	"MrRSS/internal/handlers/chat"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidateLoopbackAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{name: "IPv4 loopback", address: "127.0.0.1:1234"},
		{name: "IPv6 loopback", address: "[::1]:1234"},
		{name: "ephemeral loopback port", address: "127.0.0.1:0"},
		{name: "all interfaces", address: "0.0.0.0:1234", wantErr: true},
		{name: "non-loopback", address: "192.0.2.10:1234", wantErr: true},
		{name: "missing port", address: "127.0.0.1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLoopbackAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateLoopbackAddress(%q) error = %v, wantErr %v", tt.address, err, tt.wantErr)
			}
		})
	}
}

func TestProtectLocalAPI(t *testing.T) {
	handler := protectLocalAPI(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name           string
		host           string
		origin         string
		secFetchSite   string
		wantStatusCode int
	}{
		{name: "local agent request", host: "127.0.0.1:1234", wantStatusCode: http.StatusNoContent},
		{name: "localhost agent request", host: "localhost:1234", wantStatusCode: http.StatusNoContent},
		{name: "foreign host", host: "attacker.example", wantStatusCode: http.StatusForbidden},
		{name: "browser origin", host: "127.0.0.1:1234", origin: "https://attacker.example", wantStatusCode: http.StatusForbidden},
		{name: "cross-site fetch", host: "127.0.0.1:1234", secFetchSite: "cross-site", wantStatusCode: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/api/version", nil)
			req.Host = tt.host
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.secFetchSite != "" {
				req.Header.Set("Sec-Fetch-Site", tt.secFetchSite)
			}
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatusCode {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatusCode)
			}
		})
	}
}

func TestServerLifecycle(t *testing.T) {
	server, err := Start("127.0.0.1:0", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/version" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	response, err := http.Get("http://" + server.Address() + "/api/version")
	if err != nil {
		t.Fatalf("GET desktop API: %v", err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusNoContent)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := <-server.Errors(); err != nil {
		t.Fatalf("Serve() error after shutdown = %v", err)
	}
}

// Exercise the chat handler through each wire protocol, rather than only testing
// settings serialization. Resumed chats and selected profiles use the same value.
func TestChatResponsePreferencesAcrossModelsAndProtocols(t *testing.T) {
	db, err := database.NewDB(filepath.Join(t.TempDir(), "preferences.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	for key, value := range map[string]string{"ai_chat_enabled": "true", "ai_usage_limit": "0", "ai_chat_response_preferences": "用中文回答，保持简洁。"} {
		if err := db.SetSetting(key, value); err != nil {
			t.Fatal(err)
		}
	}
	tracker := ai.NewUsageTracker(db)
	tracker.SetMinInterval(0)
	h := &core.Handler{DB: db, AITracker: tracker, AIProfileProvider: ai.NewProfileProvider(db)}
	for _, protocol := range []string{"openai", "anthropic", "gemini", "ollama"} {
		t.Run(protocol, func(t *testing.T) {
			requests := make(chan map[string]any, 1)
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				requests <- body
				responses := map[string]string{
					"openai":    `{"choices":[{"message":{"content":"answer"}}]}`,
					"anthropic": `{"content":[{"type":"text","text":"answer"}]}`,
					"gemini":    `{"candidates":[{"content":{"parts":[{"text":"answer"}]},"finishReason":"STOP"}]}`,
					"ollama":    `{"message":{"content":"answer"},"done":true}`,
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(responses[protocol]))
			}))
			defer provider.Close()
			paths := map[string]string{"openai": "/v1/chat/completions", "anthropic": "/v1/messages", "gemini": "/v1beta/models/old:generateContent", "ollama": "/api/chat"}
			for index, preference := range []string{"用中文回答，保持简洁。", "用中文回答，保持简洁。", ""} {
				if err := db.SetSetting("ai_chat_response_preferences", preference); err != nil {
					t.Fatal(err)
				}
				profileID, err := db.CreateAIProfile(&models.AIProfile{Name: protocol, Endpoint: provider.URL + paths[protocol], Model: []string{"first-model", "other-model", "default-model"}[index]})
				if err != nil {
					t.Fatal(err)
				}
				messages := make([]chat.ChatMessage, 12)
				for i := range messages {
					messages[i] = chat.ChatMessage{Role: "user", Content: "question"}
				}
				body, _ := json.Marshal(chat.ChatRequest{Messages: messages, ProfileID: profileID, IsFirstMessage: index == 0, ArticleContent: "article evidence"})
				recorder := httptest.NewRecorder()
				chat.HandleAIChat(h, recorder, httptest.NewRequest(http.MethodPost, "/api/ai-chat", bytes.NewReader(body)))
				if recorder.Code != http.StatusOK {
					t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
				}
				request := <-requests
				system := ""
				switch protocol {
				case "openai", "ollama":
					for _, item := range request["messages"].([]any) {
						msg := item.(map[string]any)
						if msg["role"] == "system" {
							system += msg["content"].(string)
						}
					}
				case "anthropic":
					system, _ = request["system"].(string)
				case "gemini":
					if instruction, ok := request["systemInstruction"].(map[string]any); ok {
						for _, item := range instruction["parts"].([]any) {
							system += item.(map[string]any)["text"].(string)
						}
					}
				}
				if preference != "" && strings.Count(system, preference) != 1 {
					t.Fatalf("preference missing or duplicated: %q", system)
				}
				if preference == "" && system != "" {
					t.Fatalf("cleared preference still sent: %q", system)
				}
				if index == 0 && !strings.Contains(system, "article evidence") {
					t.Fatal("article context lost")
				}
			}
		})
	}
}

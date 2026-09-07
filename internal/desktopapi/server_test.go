package desktopapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"MrRSS/internal/ai"
	"MrRSS/internal/database"
	"MrRSS/internal/handlers/chat"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/routes"
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

func TestChatCancellationStopsLocalProviderAndPreservesHistory(t *testing.T) {
	db, err := database.NewDB(filepath.Join(t.TempDir(), "cancel.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	feed, err := db.Exec(`INSERT INTO feeds(title,url) VALUES('test','https://example.com/feed')`)
	if err != nil {
		t.Fatal(err)
	}
	feedID, _ := feed.LastInsertId()
	article, err := db.Exec(`INSERT INTO articles(feed_id,title,url) VALUES(?,'test','https://example.com/article')`, feedID)
	if err != nil {
		t.Fatal(err)
	}
	articleID, _ := article.LastInsertId()
	entered, cancelled, release := make(chan struct{}, 2), make(chan struct{}, 2), make(chan struct{})
	var providerCalls atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		providerCalls.Add(1)
		entered <- struct{}{}
		select {
		case <-r.Context().Done():
			cancelled <- struct{}{}
		case <-release:
			_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"completed"}}]}`))
		}
	}))
	defer provider.Close()
	defer provider.CloseClientConnections()
	for key, value := range map[string]string{"ai_chat_enabled": "true", "ai_usage_limit": "0", "ai_endpoint": provider.URL + "/v1/chat/completions"} {
		if err := db.SetSetting(key, value); err != nil {
			t.Fatal(err)
		}
	}
	tracker := ai.NewUsageTracker(db)
	tracker.SetMinInterval(0)
	h := &core.Handler{DB: db, AITracker: tracker}
	mux := http.NewServeMux()
	routes.RegisterAPIRoutes(mux, h)
	api := httptest.NewServer(mux)
	defer api.Close()
	defer api.CloseClientConnections()
	post := func(path string, body any) (int, []byte) {
		t.Helper()
		data, _ := json.Marshal(body)
		response, err := http.Post(api.URL+path, "application/json", bytes.NewReader(data))
		if err != nil {
			t.Error(err)
			return 0, nil
		}
		defer response.Body.Close()
		result, _ := io.ReadAll(response.Body)
		return response.StatusCode, result
	}
	status, body := post("/api/ai/chat/session/create", map[string]any{"article_id": articleID, "title": "New Chat"})
	var session database.ChatSession
	if err := json.Unmarshal(body, &session); err != nil || status != 200 || session.ID <= 0 {
		t.Fatalf("create status=%d body=%s err=%v", status, body, err)
	}
	request := func(id string) chat.ChatRequest {
		return chat.ChatRequest{SessionID: session.ID, ArticleID: articleID, RequestID: id, Messages: []chat.ChatMessage{{Role: "user", Content: "question " + id}}}
	}
	cancel := func(id string) {
		t.Helper()
		status, body := post("/api/ai-chat/cancel", chat.CancelChatRequest{SessionID: session.ID, RequestID: id})
		if status != 200 || string(bytes.TrimSpace(body)) != `{"success":true}` {
			t.Fatalf("cancel status=%d body=%s", status, body)
		}
	}
	await := func(ch <-chan struct{}, label string) {
		t.Helper()
		select {
		case <-ch:
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out: %s", label)
		}
	}
	done := make(chan int, 1)
	go func() { status, _ := post("/api/ai-chat", request("first")); done <- status }()
	await(entered, "provider start")
	cancel("first")
	cancel("first")
	select {
	case status := <-done:
		if status != http.StatusRequestTimeout {
			t.Fatalf("cancelled status=%d", status)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("generation did not stop")
	}
	await(cancelled, "provider context cancellation")
	history, err := db.GetChatMessages(session.ID)
	if err != nil || len(history) != 1 || history[0].Role != "user" {
		t.Fatalf("cancel history=%+v err=%v", history, err)
	}
	cancel("early")
	status, _ = post("/api/ai-chat", request("early"))
	if status != http.StatusRequestTimeout || providerCalls.Load() != 1 {
		t.Fatal("early cancellation allowed provider request")
	}
	history, _ = db.GetChatMessages(session.ID)
	if len(history) != 1 {
		t.Fatal("early cancellation saved another user message")
	}

	// A delayed cancel for the old request must not touch a new request in the same session.
	go func() { status, _ := post("/api/ai-chat", request("next")); done <- status }()
	await(entered, "next provider start")
	cancel("first")
	select {
	case <-cancelled:
		t.Fatal("old cancellation stopped new request")
	default:
	}
	close(release)
	select {
	case status := <-done:
		if status != 200 {
			t.Fatalf("new request status=%d", status)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("new generation did not finish")
	}
	cancel("next")
	status, _ = post("/api/ai-chat", request("next"))
	if status != http.StatusRequestTimeout || providerCalls.Load() != 2 {
		t.Fatal("completed ID was replayed")
	}
	history, _ = db.GetChatMessages(session.ID)
	if len(history) != 3 || history[2].Role != "assistant" {
		t.Fatalf("completed history=%+v", history)
	}

	tracker.SetMinInterval(time.Hour)
	go func() { status, _ := post("/api/ai-chat", request("rate-wait")); done <- status }()
	deadline := time.Now().Add(2 * time.Second)
	for {
		history, _ = db.GetChatMessages(session.ID)
		if len(history) == 4 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("request did not enter rate limit wait")
		}
		time.Sleep(time.Millisecond)
	}
	cancel("rate-wait")
	select {
	case status := <-done:
		if status != http.StatusRequestTimeout {
			t.Fatalf("rate wait status=%d", status)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("rate limit wait did not cancel")
	}
	if providerCalls.Load() != 2 {
		t.Fatal("cancelled rate wait called provider")
	}
}

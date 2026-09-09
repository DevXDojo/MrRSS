package chat

import (
	"testing"

	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
)

func TestPersistUserChatMessageRebindsSessionToCurrentArticle(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Init(); err != nil {
		t.Fatalf("Init: %v", err)
	}

	feedID, err := db.AddFeed(&models.Feed{Title: "Feed", URL: "https://example.com/feed"})
	if err != nil {
		t.Fatalf("AddFeed: %v", err)
	}
	firstResult, err := db.Exec(
		`INSERT INTO articles (feed_id, title, url, unique_id) VALUES (?, ?, ?, ?)`,
		feedID, "First", "https://example.com/1", "1",
	)
	if err != nil {
		t.Fatalf("insert first article: %v", err)
	}
	firstID, err := firstResult.LastInsertId()
	if err != nil {
		t.Fatalf("first article ID: %v", err)
	}
	secondResult, err := db.Exec(
		`INSERT INTO articles (feed_id, title, url, unique_id) VALUES (?, ?, ?, ?)`,
		feedID, "Second", "https://example.com/2", "2",
	)
	if err != nil {
		t.Fatalf("insert second article: %v", err)
	}
	secondID, err := secondResult.LastInsertId()
	if err != nil {
		t.Fatalf("second article ID: %v", err)
	}
	sessionID, err := db.CreateChatSession(firstID, "Discussion")
	if err != nil {
		t.Fatalf("CreateChatSession: %v", err)
	}

	h := core.NewHandler(db, nil, nil, nil)
	_, enabled, err := persistUserChatMessage(h, &ChatRequest{
		SessionID: sessionID,
		ArticleID: secondID,
		Messages:  []ChatMessage{{Role: "user", Content: "Accidental stale request."}},
	})
	if err == nil || enabled {
		t.Fatalf("expected mismatched session without explicit rebind to fail, enabled=%v err=%v", enabled, err)
	}

	gotSessionID, enabled, err := persistUserChatMessage(h, &ChatRequest{
		SessionID:     sessionID,
		ArticleID:     secondID,
		RebindSession: true,
		Messages:      []ChatMessage{{Role: "user", Content: "Compare this article."}},
	})
	if err != nil {
		t.Fatalf("persistUserChatMessage: %v", err)
	}
	if gotSessionID != sessionID || !enabled {
		t.Fatalf("session=%d enabled=%v, want session=%d enabled=true", gotSessionID, enabled, sessionID)
	}
	session, err := db.GetChatSession(sessionID)
	if err != nil {
		t.Fatalf("GetChatSession: %v", err)
	}
	if session == nil || session.ArticleID != secondID {
		t.Fatalf("session article=%v, want %d", session, secondID)
	}
}

func TestPersistUserChatMessageSkipsStorageWhenHistoryDisabled(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	if err := db.SetSetting("ai_chat_save_history", "false"); err != nil {
		t.Fatal(err)
	}

	h := core.NewHandler(db, nil, nil, nil)
	sessionID, enabled, err := persistUserChatMessage(h, &ChatRequest{
		ArticleID: 1,
		Messages:  []ChatMessage{{Role: "user", Content: "Do not save this."}},
	})
	if err != nil || enabled || sessionID != 0 {
		t.Fatalf("session=%d enabled=%v err=%v, want disabled transient chat", sessionID, enabled, err)
	}
}

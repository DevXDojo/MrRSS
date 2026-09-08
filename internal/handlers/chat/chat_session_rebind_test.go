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
	firstID, err := db.AddArticle(&models.Article{FeedID: feedID, Title: "First", URL: "https://example.com/1", UniqueID: "1"})
	if err != nil {
		t.Fatalf("AddArticle first: %v", err)
	}
	secondID, err := db.AddArticle(&models.Article{FeedID: feedID, Title: "Second", URL: "https://example.com/2", UniqueID: "2"})
	if err != nil {
		t.Fatalf("AddArticle second: %v", err)
	}
	sessionID, err := db.CreateChatSession(firstID, "Discussion")
	if err != nil {
		t.Fatalf("CreateChatSession: %v", err)
	}

	h := core.NewHandler(db, nil, nil, nil)
	gotSessionID, enabled, err := persistUserChatMessage(h, &ChatRequest{
		SessionID: sessionID,
		ArticleID: secondID,
		Messages:  []ChatMessage{{Role: "user", Content: "Compare this article."}},
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

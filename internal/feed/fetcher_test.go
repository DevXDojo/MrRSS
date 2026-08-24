package feed

import (
	"MrRSS/internal/database"
	"testing"
)

func TestNewFetcherSanity(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}

	f := NewFetcher(db)
	if f == nil {
		t.Fatal("NewFetcher returned nil")
	}
}

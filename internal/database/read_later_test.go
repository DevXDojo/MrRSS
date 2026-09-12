package database_test

import (
	"testing"
	"time"
)

func TestReadLaterMembershipIsIndependentOfReading(t *testing.T) {
	db := setupDBWithFeed(t)
	t.Cleanup(func() { _ = db.Close() })
	var feedID int64
	if err := db.QueryRow(`SELECT id FROM feeds LIMIT 1`).Scan(&feedID); err != nil {
		t.Fatal(err)
	}
	result, err := db.Exec(`INSERT INTO articles (feed_id, title, url, published_at, is_read, is_read_later) VALUES (?, 'Saved', 'https://example.com/saved', ?, 1, 0)`, feedID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	assertState := func(t *testing.T, read, later bool) {
		t.Helper()
		article, err := db.GetArticleByID(id)
		if err != nil {
			t.Fatal(err)
		}
		if article.IsRead != read || article.IsReadLater != later {
			t.Fatalf("got read/later %v/%v, want %v/%v", article.IsRead, article.IsReadLater, read, later)
		}
	}
	for _, step := range []struct {
		name        string
		run         func() error
		read, later bool
	}{
		{"save read article", func() error { return db.ToggleReadLater(id) }, true, true},
		{"mark unread", func() error { return db.MarkArticleRead(id, false) }, false, true},
		{"read saved article", func() error { return db.MarkArticleRead(id, true) }, true, true},
		{"remove explicitly", func() error { return db.ToggleReadLater(id) }, true, false},
		{"set membership", func() error { return db.SetArticleReadLater(id, true) }, true, true},
		{"clear membership", func() error { return db.SetArticleReadLater(id, false) }, true, false},
	} {
		t.Run(step.name, func(t *testing.T) {
			if err := step.run(); err != nil {
				t.Fatal(err)
			}
			assertState(t, step.read, step.later)
		})
	}
}

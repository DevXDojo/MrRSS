package database_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"MrRSS/internal/database"
	"MrRSS/internal/models"
)

func TestCategoryActions(t *testing.T) {
	for _, action := range []string{"dissolve", "unsubscribe"} {
		t.Run(action, func(t *testing.T) {
			db := setupTestDB(t)
			ids := make([]int64, 0)
			for i, category := range []string{"A_%", "A_%/Child", "ABx", ""} {
				id, err := db.AddFeed(&models.Feed{Title: category, URL: fmt.Sprintf("https://example.com/%d", i), Category: category})
				if err != nil {
					t.Fatal(err)
				}
				ids = append(ids, id)
				if _, err := db.Exec(`INSERT INTO articles (feed_id, title, url, unique_id, is_favorite, is_read) VALUES (?, 'saved', ?, ?, 1, 1)`, id, fmt.Sprint(i), fmt.Sprint(i)); err != nil {
					t.Fatal(err)
				}
			}
			n, err := db.ChangeFeedCategory(context.Background(), "A_%", action)
			if err != nil || n != 2 {
				t.Fatalf("changed %d: %v", n, err)
			}
			var feedCount, articleCount int
			db.QueryRow(`SELECT COUNT(*) FROM feeds`).Scan(&feedCount)
			db.QueryRow(`SELECT COUNT(*) FROM articles`).Scan(&articleCount)
			want := 4
			if action == "unsubscribe" {
				want = 2
			}
			if feedCount != want || articleCount != want {
				t.Fatalf("feeds=%d articles=%d", feedCount, articleCount)
			}
			if action == "dissolve" {
				var category string
				db.QueryRow(`SELECT category FROM feeds WHERE id = ?`, ids[1]).Scan(&category)
				if category != "" {
					t.Fatalf("nested feed not dissolved: %q", category)
				}
			}
		})
	}
}

func TestCategoryUnsubscribeRollsBackAndProtectsSyncedFeeds(t *testing.T) {
	db := setupTestDB(t)
	id, err := db.AddFeed(&models.Feed{Title: "Local", URL: "https://example.com/local", Category: "News"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`INSERT INTO articles (feed_id, title, unique_id) VALUES (?, 'Keep me', 'keep')`, id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`CREATE TRIGGER block_feed_delete BEFORE DELETE ON feeds BEGIN SELECT RAISE(ABORT, 'blocked'); END`); err != nil {
		t.Fatal(err)
	}
	if _, err = db.ChangeFeedCategory(context.Background(), "News", "unsubscribe"); err == nil {
		t.Fatal("expected transaction failure")
	}
	var count int
	db.QueryRow(`SELECT COUNT(*) FROM articles`).Scan(&count)
	if count != 1 {
		t.Fatal("articles were deleted despite rollback")
	}
	db.Exec(`DROP TRIGGER block_feed_delete`)
	_, err = db.AddFeed(&models.Feed{Title: "Remote", URL: "https://example.com/remote", Category: "News/Remote", IsFreshRSSSource: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"dissolve", "unsubscribe"} {
		if _, err := db.ChangeFeedCategory(context.Background(), "News", action); !errors.Is(err, database.ErrSyncedCategory) {
			t.Fatalf("expected synced protection: %v", err)
		}
	}
	db.QueryRow(`SELECT COUNT(*) FROM articles`).Scan(&count)
	if count != 1 {
		t.Fatal("synced category operation changed data")
	}
}

func TestFavoritesIgnoreUnreadOnly(t *testing.T) {
	db := setupDBWithFeed(t)
	var feedID int64
	db.QueryRow(`SELECT id FROM feeds LIMIT 1`).Scan(&feedID)
	for i, favorite := range []int{1, 0} {
		_, err := db.Exec(`INSERT INTO articles (feed_id, title, url, unique_id, is_read, is_favorite, published_at) VALUES (?, 'Title', ?, ?, 1, ?, CURRENT_TIMESTAMP)`, feedID, fmt.Sprint(i), fmt.Sprint(i), favorite)
		if err != nil {
			t.Fatal(err)
		}
	}
	articles, err := db.GetArticlesWithUnreadFilter("favorites", 0, "", false, true, 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 1 || !articles[0].IsRead || !articles[0].IsFavorite {
		t.Fatalf("read favorite missing: %+v", articles)
	}
}

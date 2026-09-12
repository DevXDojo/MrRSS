package database_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestArticleFeedGroupingBeforePagination(t *testing.T) {
	db := setupTestDB(t)
	t.Cleanup(func() { _ = db.Close() })
	for id := 1; id <= 2; id++ {
		if _, err := db.Exec(`INSERT INTO feeds (id, title, url, category) VALUES (?, ?, ?, ?)`, id, "Same title", fmt.Sprintf("https://example.com/feed/%d", id), "news"); err != nil {
			t.Fatal(err)
		}
	}
	// Newest-first without grouping would alternate between feeds.
	for id, feedID := range []int{2, 1, 2, 1} {
		if _, err := db.Exec(`INSERT INTO articles (id, feed_id, title, url, unique_id, published_at) VALUES (?, ?, ?, ?, ?, ?)`, id+1, feedID, "Article", fmt.Sprintf("https://example.com/%d", id), fmt.Sprintf("group-%d", id), time.Date(2026, 9, 12, id, 0, 0, 0, time.UTC)); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct {
		order string
		group string
		want  []int64
	}{
		{"newest", "feed", []int64{4, 2, 3, 1}},
		{"oldest", "feed", []int64{2, 4, 1, 3}},
		{"newest", "date", []int64{4, 3, 2, 1}},
		{"newest", "feed; DROP TABLE articles", []int64{4, 3, 2, 1}},
	} {
		t.Run(tc.order+"/"+tc.group, func(t *testing.T) {
			var ids []int64
			for offset := 0; offset < 4; offset += 3 {
				articles, err := db.GetArticlesWithUnreadFilterSorted("all", 0, "news", false, false, tc.order, 3, offset, tc.group)
				if err != nil {
					t.Fatal(err)
				}
				for _, article := range articles {
					ids = append(ids, article.ID)
				}
			}
			if !reflect.DeepEqual(ids, tc.want) {
				t.Fatalf("article IDs = %v, want %v", ids, tc.want)
			}
		})
	}
}

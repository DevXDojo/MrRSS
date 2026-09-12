package freshrss

import (
	"context"
	"testing"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/models"
)

func TestReaderProvidersKeepIdenticalFeedsAndArticlesIndependent(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	localID, err := db.AddFeed(&models.Feed{Title: "Local", URL: "https://example.org/rss", Category: "News"})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.SaveArticles(ctx, []*models.Article{{FeedID: localID, Title: "Article", URL: "https://example.org/1", PublishedAt: time.Now()}}); err != nil {
		t.Fatal(err)
	}
	services := map[string]*BidirectionalSyncService{}
	ids := map[string]int64{}
	for _, provider := range []string{"freshrss", "miniflux"} {
		services[provider] = NewBidirectionalSyncServiceForProvider("https://"+provider+".example", "reader", "password", provider, db)
		for key, value := range map[string]string{"enabled": "true", "server_url": "https://" + provider + ".example", "username": "reader"} {
			if err := db.SetSetting(provider+"_"+key, value); err != nil {
				t.Fatal(err)
			}
		}
		if err := db.SetEncryptedSetting(provider+"_api_password", "password"); err != nil {
			t.Fatal(err)
		}
		subs := []Subscription{{ID: "feed/1", URL: "https://example.org/rss", Title: "Shared feed", Categories: []Category{{ID: "user/-/label/News", Label: "News"}}}}
		if _, err := services[provider].createFeedsFromSubscriptions(ctx, subs); err != nil {
			t.Fatal(err)
		}
		article := Article{ID: provider + "-item", Title: "Article", URL: "https://example.org/1", Content: "<p>" + provider + "</p>", OriginStreamID: "feed/1", Published: time.Now()}
		if _, err := services[provider].saveArticlesFromServer(ctx, []Article{article}); err != nil {
			t.Fatal(err)
		}
		// A second sync must reuse the same subscription and article.
		if _, err := services[provider].createFeedsFromSubscriptions(ctx, subs); err != nil {
			t.Fatal(err)
		}
		if _, err := services[provider].saveArticlesFromServer(ctx, []Article{article}); err != nil {
			t.Fatal(err)
		}
		saved, err := db.GetArticleByURL(article.URL, provider)
		if err != nil {
			t.Fatal(err)
		}
		if saved.FreshRSSItemID != article.ID {
			t.Fatal("remote item identifier not persisted")
		}
		ids[provider] = saved.ID
		if err := db.EnqueueSyncChange(saved.ID, article.URL, database.SyncActionMarkRead); err != nil {
			t.Fatal(err)
		}
	}
	feeds, err := db.GetFeeds()
	if err != nil || len(feeds) != 3 {
		t.Fatalf("feeds=%d err=%v", len(feeds), err)
	}
	for _, feed := range feeds {
		if feed.SyncProvider == "miniflux" && feed.IsFreshRSSSource && feed.Category != "News (Miniflux)" {
			t.Fatalf("unstable Miniflux category: %s", feed.Category)
		}
	}
	if ids["freshrss"] == ids["miniflux"] {
		t.Fatal("articles merged across providers")
	}
	for _, provider := range []string{"freshrss", "miniflux"} {
		queue, err := db.GetPendingSyncChanges(1, provider)
		if err != nil || len(queue) != 1 || queue[0].ArticleID != ids[provider] {
			t.Fatalf("queue crossed provider boundary: %s", provider)
		}
	}
	// Disabling FreshRSS must not suppress Miniflux bulk or immediate updates.
	if err := db.SetSetting("freshrss_enabled", "false"); err != nil {
		t.Fatal(err)
	}
	reqs, err := db.MarkArticlesReadWithSync([]int64{ids["freshrss"], ids["miniflux"]}, true)
	if err != nil || len(reqs) != 1 || reqs[0].ArticleID != ids["miniflux"] {
		t.Fatal("bulk sync routed incorrectly")
	}
	if err := services["freshrss"].SyncArticleStatus(ctx, ids["miniflux"], "https://example.org/1", database.SyncActionStar); err == nil {
		t.Fatal("foreign article accepted")
	}
	if _, err := services["miniflux"].applyServerStatus("https://example.org/1", true, "is_favorite"); err != nil {
		t.Fatal(err)
	}
	other, err := db.GetArticleByID(ids["freshrss"])
	if err != nil || other.IsFavorite {
		t.Fatal("server state leaked to other reader")
	}
	if err := db.CleanupReaderData("freshrss"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetArticleByID(ids["miniflux"]); err != nil {
		t.Fatal("cleanup deleted other provider")
	}
	if _, err := db.GetFeedByID(localID); err != nil {
		t.Fatal("cleanup deleted local subscription")
	}
	// Remote deletion must also be restricted to the selected reader.
	if _, err := services["freshrss"].createFeedsFromSubscriptions(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetArticleByID(ids["miniflux"]); err != nil {
		t.Fatal("remote deletion crossed provider boundary")
	}
}

package freshrss

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/models"
)

func TestReaderSyncRecordsFailureAndRecovery(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	var loggedIn atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimPrefix(r.URL.Path, "/api/greader.php") {
		case "/accounts/ClientLogin":
			if !loggedIn.Load() {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			fmt.Fprintln(w, "Auth=token")
		case "/reader/api/0/subscription/list":
			fmt.Fprintln(w, `{"subscriptions":[]}`)
		default:
			fmt.Fprintln(w, `{"items":[]}`)
		}
	}))
	t.Cleanup(server.Close)
	const lastGood = "2026-09-12T02:00:00Z"
	for key, value := range map[string]string{"enabled": "true", "server_url": server.URL, "username": "reader", "last_sync_time": lastGood} {
		if err := db.SetSetting("freshrss_"+key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.SetEncryptedSetting("freshrss_api_password", "test-password"); err != nil {
		t.Fatal(err)
	}
	run := func(wantError string, wantNewTime bool) {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := StartSync(ctx, db, "freshrss", ""); err != nil {
			t.Fatal(err)
		}
		for IsSyncRunning(db, "freshrss") && ctx.Err() == nil {
			time.Sleep(time.Millisecond)
		}
		if ctx.Err() != nil {
			t.Fatal(ctx.Err())
		}
		failure, _ := db.GetSetting("freshrss_last_sync_error")
		success, _ := db.GetSetting("freshrss_last_sync_time")
		attempt, _ := db.GetSetting("freshrss_last_sync_attempt")
		if failure != wantError || (success != lastGood) != wantNewTime || attempt == "" {
			t.Fatalf("sync outcome: failure=%q, success=%q, attempt=%q", failure, success, attempt)
		}
	}
	run("login", false)
	loggedIn.Store(true)
	run("", true)
}

func TestReaderPushesQueuedEditAfterPullTimeout(t *testing.T) {
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Init(); err != nil {
		t.Fatal(err)
	}
	feedID, err := db.AddFeed(&models.Feed{Title: "News", URL: "https://example.test/feed", IsFreshRSSSource: true, SyncProvider: "freshrss"})
	if err != nil {
		t.Fatal(err)
	}
	const articleURL = "https://example.test/story"
	if err := db.SaveArticles(t.Context(), []*models.Article{{FeedID: feedID, Title: "News", URL: articleURL, PublishedAt: time.Now()}}); err != nil {
		t.Fatal(err)
	}
	item, err := db.GetArticleByURL(articleURL, "freshrss")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.EnqueueSyncChange(item.ID, articleURL, database.SyncActionMarkRead); err != nil {
		t.Fatal(err)
	}
	var edited atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimPrefix(r.URL.Path, "/api/greader.php") {
		case "/accounts/ClientLogin":
			fmt.Fprintln(w, "Auth=token")
		case "/reader/api/0/subscription/list":
			<-r.Context().Done()
		case "/reader/api/0/token":
			fmt.Fprint(w, "token")
		case "/reader/api/0/edit-tag":
			edited.Store(true)
			fmt.Fprint(w, "OK")
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	service := NewBidirectionalSyncService(server.URL, "reader", "test-password", db)
	result, err := service.syncWithTimeout(t.Context(), 500*time.Millisecond, 5*time.Second)
	if err == nil || result.PullSuccess || !result.PushSuccess || !edited.Load() {
		t.Fatalf("expected timed-out pull and completed push: %+v, %v", result, err)
	}
	if pending, err := db.GetPendingSyncCount("freshrss"); err != nil || pending != 0 {
		t.Fatalf("pending=%d, err=%v", pending, err)
	}
}

func TestReaderSyncScheduling(t *testing.T) {
	now := time.Now()
	for _, test := range []struct {
		interval, last string
		startup, want  bool
	}{
		{"0", "", false, false}, {"-1", "", false, false}, {"invalid", "", false, false},
		{"0", "", true, true}, {"30", "", false, true},
		{"30", now.Add(-29 * time.Minute).Format(time.RFC3339Nano), false, false},
		{"30", now.Add(-30 * time.Minute).Format(time.RFC3339Nano), false, true},
	} {
		if got := readerSyncDue(now, test.interval, test.last, test.startup); got != test.want {
			t.Errorf("%+v got %v", test, got)
		}
	}
}

func TestReaderSyncGuardAndCancellation(t *testing.T) {
	db := &database.DB{}
	release, ok := AcquireSync(db, "freshrss")
	if !ok {
		t.Fatal("first sync blocked")
	}
	defer release()
	if _, ok := AcquireSync(db, "freshrss"); ok {
		t.Fatal("overlapping provider sync allowed")
	}
	other, ok := AcquireSync(db, "miniflux")
	if !ok {
		t.Fatal("different provider blocked")
	}
	other()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := StartSync(ctx, db, "freshrss", ""); err != context.Canceled {
		t.Fatalf("cancelled sync: %v", err)
	}
}

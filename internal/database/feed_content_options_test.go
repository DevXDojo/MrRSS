package database_test

import (
	dbpkg "MrRSS/internal/database"
	"MrRSS/internal/models"
	"context"
	"strings"
	"testing"
)

func TestFeedContentOptionsCookieEncryptionAndPreservation(t *testing.T) {
	db := setupTestDB(t)
	id, err := db.AddFeed(&models.Feed{Title: "f", URL: "https://example.org"})
	if err != nil {
		t.Fatal(err)
	}
	cookie := "session=private; account=test"
	options := dbpkg.FeedContentOptions{ContentSelector: "article", RemoveSelector: ".ad", Cookie: &cookie, CookieOrigin: "https://example.org"}
	if err := db.SetFeedContentOptions(context.Background(), id, options); err != nil {
		t.Fatal(err)
	}
	var stored string
	if err := db.QueryRow("SELECT cookie FROM feed_content_options WHERE feed_id=?", id).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stored, "private") || stored == cookie {
		t.Fatal("plaintext Cookie persisted")
	}
	options.Cookie = nil
	options.ContentSelector = "main"
	if err := db.SetFeedContentOptions(context.Background(), id, options); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetFeedContentOptions(context.Background(), id)
	if err != nil || got.Cookie == nil || *got.Cookie != cookie || got.ContentSelector != "main" {
		t.Fatalf("roundtrip: %+v %v", got, err)
	}
	empty := ""
	options.Cookie = &empty
	if err := db.SetFeedContentOptions(context.Background(), id, options); err != nil {
		t.Fatal(err)
	}
	got, err = db.GetFeedContentOptions(context.Background(), id)
	if err != nil || got.HasCookie {
		t.Fatal("Cookie was not removed")
	}
	if err := db.DeleteFeed(id); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM feed_content_options").Scan(&count); err != nil || count != 0 {
		t.Fatal("orphaned content options", err, count)
	}
}
func TestFeedContentOptionsValidation(t *testing.T) {
	for _, options := range []dbpkg.FeedContentOptions{{ContentSelector: "["}, {RemoveSelector: "["}, {CookieOrigin: "file:///test"}, {CookieOrigin: "https://example.org/private"}, {Cookie: ptrCookie("session=x\r\nInjected: yes"), CookieOrigin: "https://example.org"}, {Cookie: ptrCookie("x=y")}} {
		if options.Validate() == nil {
			t.Errorf("accepted invalid options: %+v", options)
		}
	}
}
func ptrCookie(value string) *string { return &value }

package notification

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"MrRSS/internal/crypto"
	"MrRSS/internal/database"
	"MrRSS/internal/models"
	"MrRSS/internal/utils/httputil"
)

func testChannel(provider string) Channel {
	c := Channel{ID: provider, Name: provider, Provider: provider, Enabled: true}
	switch provider {
	case "telegram":
		c.Token = "123:test_token"
		c.ChatID = "-1001"
	case "discord":
		c.Webhook = "https://discord.com/api/webhooks/123/test_token"
	case "feishu":
		c.Webhook = "https://open.feishu.cn/open-apis/bot/v2/hook/test-token"
		c.Secret = "test-secret"
	}
	return c
}
func testRule() Rule {
	return Rule{ID: "rule", Name: "Technology", Enabled: true, Mode: "alert", ChannelIDs: []string{"telegram"}, Match: "all", IntervalMinutes: 1, Time: "09:00", MaxItems: 5, IncludeSummary: true, IncludeLink: true}
}
func testConfig() Config {
	c := DefaultConfig()
	c.Enabled = true
	c.Timezone = "UTC"
	c.Channels = []Channel{testChannel("telegram")}
	c.Rules = []Rule{testRule()}
	return c
}
func testService(t *testing.T) (*Service, *database.DB) {
	t.Helper()
	db, err := database.NewDB(filepath.Join(t.TempDir(), "notifications.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if _, err = db.Exec(`INSERT INTO feeds(id,title,url,category) VALUES(1,'Engineering','https://example.test/feed','Tech')`); err != nil {
		t.Fatal(err)
	}
	return New(db), db
}
func addArticle(t *testing.T, db *database.DB, title string) int64 {
	t.Helper()
	res, err := db.Exec(`INSERT INTO articles(feed_id,title,url,original_summary,author,is_hidden,is_read,is_favorite,is_read_later) VALUES(1,?,'https://example.test/article','<p>Useful summary</p>','Writer',0,0,0,0)`, title)
	if err != nil {
		t.Fatal(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}
func mockClient(fn func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{Transport: httputil.RoundTripFunc(fn)}
}
func reply(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}

func TestProviders(t *testing.T) {
	now := time.Unix(1700000000, 0)
	for _, provider := range []string{"telegram", "discord", "feishu"} {
		t.Run(provider, func(t *testing.T) {
			ch := testChannel(provider)
			ch.ThreadID = 42
			client := mockClient(func(req *http.Request) (*http.Response, error) {
				if req.Method != "POST" || req.Header.Get("Content-Type") != "application/json" {
					t.Fatal("invalid request")
				}
				var body map[string]interface{}
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				switch provider {
				case "telegram":
					if body["text"] != "MrRSS @everyone <b>plain</b>" || body["message_thread_id"] != float64(42) || body["parse_mode"] != nil {
						t.Fatalf("unsafe or incorrect Telegram payload: %#v", body)
					}
					return reply(200, `{"ok":true}`), nil
				case "discord":
					if req.URL.Query().Get("wait") != "true" {
						t.Fatal("must wait for confirmation")
					}
					mentions := body["allowed_mentions"].(map[string]interface{})["parse"].([]interface{})
					if len(mentions) != 0 {
						t.Fatal("mentions enabled")
					}
					return reply(200, `{"id":"1"}`), nil
				case "feishu":
					mac := hmac.New(sha256.New, []byte(strconv.FormatInt(now.Unix(), 10)+"\n"+ch.Secret))
					if body["sign"] != base64.StdEncoding.EncodeToString(mac.Sum(nil)) {
						t.Fatal("invalid signature")
					}
					return reply(200, `{"code":0}`), nil
				}
				panic("unknown provider")
			})
			if err := Send(context.Background(), client, ch, "MrRSS @everyone <b>plain</b>", now); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestProviderFailuresDoNotExposeSecrets(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		code   string
		retry  bool
	}{
		{"auth", 401, "test_token", "unauthorized", false}, {"transient", 503, "test_token", "rejected", true}, {"rate", 429, `{"parameters":{"retry_after":120}}`, "rate_limit", true}, {"false-success", 200, `{"ok":false,"description":"test_token"}`, "rejected", false}, {"malformed", 200, "test_token", "rejected", false}, {"redirect", 302, "test_token", "rejected", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := mockClient(func(*http.Request) (*http.Response, error) { return reply(tc.status, tc.body), nil })
			err := Send(context.Background(), client, testChannel("telegram"), "test", time.Now())
			code, retry, after := deliveryFailure(err)
			if code != tc.code || retry != tc.retry {
				t.Fatalf("%v", err)
			}
			if strings.Contains(err.Error(), "test_token") {
				t.Fatal("secret exposed")
			}
			if tc.name == "rate" && after != 2*time.Minute {
				t.Fatal(after)
			}
		})
	}
	client := mockClient(func(*http.Request) (*http.Response, error) { return nil, errors.New("https://secret-token") })
	err := Send(context.Background(), client, testChannel("telegram"), "test", time.Now())
	if err.Error() != "network" {
		t.Fatal(err)
	}
	for _, body := range []string{`{}`, `{"code":19021}`, `{"StatusCode":1}`} {
		client := mockClient(func(*http.Request) (*http.Response, error) { return reply(200, body), nil })
		if Send(context.Background(), client, testChannel("feishu"), "test", time.Now()) == nil {
			t.Fatal("accepted rejected Feishu response")
		}
	}
}

func TestValidationAndSecretMask(t *testing.T) {
	for _, url := range []string{"http://discord.com/api/webhooks/123/a", "https://evil.test/api/webhooks/123/a", "https://discord.com.evil.test/api/webhooks/123/a", "https://discord.com@evil.test/api/webhooks/123/a", "https://discord.com/api/webhooks/123/a?wait=true", "https://discord.com:443/api/webhooks/123/a", "https://discord.com/api/webhooks/123/a/../../users"} {
		c := testChannel("discord")
		c.Webhook = url
		if ValidateChannel(c) == nil {
			t.Fatalf("accepted %s", url)
		}
	}
	c := testConfig()
	redacted := c.Redacted()
	if redacted.Channels[0].Token != SecretMask || c.Channels[0].Token == SecretMask {
		t.Fatal("redaction modified config")
	}
	if err := redacted.RestoreSecrets(c); err != nil || redacted.Channels[0].Token != c.Channels[0].Token {
		t.Fatal("mask round trip failed")
	}
	redacted = c.Redacted()
	redacted.Channels[0].ID = "other"
	if redacted.RestoreSecrets(c) == nil {
		t.Fatal("copied secret across IDs")
	}
	c = testConfig()
	c.Rules[0].Conditions = []Condition{{Field: "favorite", Operator: "contains", Value: "yes"}}
	if Validate(c) == nil {
		t.Fatal("invalid boolean condition accepted")
	}
}

func TestQuietHoursAndDailySchedule(t *testing.T) {
	c := testConfig()
	c.Timezone = "Asia/Shanghai"
	c.QuietEnabled = true
	for _, tc := range []struct {
		utc   string
		quiet bool
	}{{"2026-09-20T13:59:00Z", false}, {"2026-09-20T14:00:00Z", true}, {"2026-09-20T23:59:00Z", true}, {"2026-09-21T00:00:00Z", false}} {
		now, _ := time.Parse(time.RFC3339, tc.utc)
		if c.IsQuiet(now) != tc.quiet {
			t.Fatal(tc)
		}
	}
	r := testRule()
	r.Mode = "digest"
	r.Time = "09:00"
	before, _ := time.Parse(time.RFC3339, "2026-09-20T00:30:00Z")
	now := before.Add(30 * time.Minute)
	if !due(r, c, before, now) {
		t.Fatal("a rule created before today's scheduled time should run today")
	}
	if due(r, c, now, now.Add(time.Hour)) {
		t.Fatal("daily repeat")
	}
	r.Weekdays = []int{1}
	if due(r, c, before, now) {
		t.Fatal("wrong weekday")
	}
}

func TestConditionsAndMessages(t *testing.T) {
	a := database.NotificationArticle{Article: models.Article{Title: "Go Release", FeedTitle: "Engineering", OriginalSummary: "<script>secret()</script><p>安全 &amp; fast</p>", URL: "https://example.test/go", IsFavorite: true}, Category: "Tech"}
	r := testRule()
	r.Conditions = []Condition{{Field: "title", Operator: "contains", Value: "go"}, {Field: "category", Operator: "equals", Value: "other"}}
	if matches(a, r) {
		t.Fatal("AND failed")
	}
	r.Match = "any"
	if !matches(a, r) {
		t.Fatal("OR failed")
	}
	r.Match = "all"
	r.Conditions = []Condition{{Field: "favorite", Operator: "equals", Value: "true"}, {Field: "title", Operator: "not_contains", Value: "rust"}}
	if !matches(a, r) {
		t.Fatal("boolean or exclusion failed")
	}
	a.IsHidden = true
	if matches(a, r) {
		t.Fatal("hidden article matched")
	}
	a.IsHidden = false
	r.UnreadOnly = true
	a.IsRead = true
	if matches(a, r) {
		t.Fatal("read article matched")
	}
	r = testRule()
	a.Title = strings.Repeat("😀", 300)
	a.URL = "javascript:alert(1)"
	a.Summary = strings.Repeat("摘要😀", 400)
	articles := make([]database.NotificationArticle, 20)
	for i := range articles {
		articles[i] = a
	}
	parts := render(r, articles)
	if len(parts) < 2 {
		t.Fatal("long message not split")
	}
	for _, part := range parts {
		if units(part) > 1900 || strings.Contains(part, "javascript:") || strings.Contains(part, "<script>") {
			t.Fatal("unsafe or oversized message")
		}
	}
}

func TestPersistenceBaselineRetriesAndRestart(t *testing.T) {
	s, db := testService(t)
	ctx := context.Background()
	addArticle(t, db, "Old archive")
	c, err := s.Save(ctx, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	stored, err := db.GetSetting("notification_config")
	if err != nil || !crypto.IsEncrypted(stored) || strings.Contains(stored, "test_token") {
		t.Fatal("credentials not encrypted")
	}
	if c.Channels[0].Token != SecretMask {
		t.Fatal("API exposes secret")
	}
	if _, err = s.Save(ctx, testConfig()); !errors.Is(err, ErrConflict) {
		t.Fatal("stale update allowed")
	}
	newID := addArticle(t, db, "New article")
	attempts := 0
	factory := func() (*http.Client, error) {
		return mockClient(func(req *http.Request) (*http.Response, error) {
			attempts++
			raw, _ := io.ReadAll(req.Body)
			if strings.Contains(string(raw), "Old archive") {
				t.Fatal("sent old article")
			}
			if attempts == 1 {
				return reply(429, `{"retry_after":120}`), nil
			}
			return reply(200, `{"ok":true}`), nil
		}), nil
	}
	s.client = factory
	now := time.Now().Add(2 * time.Minute)
	if err = s.Tick(ctx, now); err != nil {
		t.Fatal(err)
	}
	state, err := db.GetNotificationState(ctx, "rule:telegram")
	if err != nil || state.Cursor != newID {
		t.Fatal(state, err)
	}
	history, err := s.History(ctx)
	if err != nil || len(history) != 1 || history[0].Status != "pending" || history[0].Attempts != 1 {
		t.Fatal(history, err)
	}
	// A new service instance represents an application restart.
	s = New(db)
	s.client = factory
	if err = s.Tick(ctx, now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if attempts != 1 {
		t.Fatal("retry ignored backoff")
	}
	if err = s.Tick(ctx, now.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 {
		t.Fatal("retry missing")
	}
	if err = s.Tick(ctx, now.Add(4*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if attempts != 2 {
		t.Fatal("duplicate delivery after restart")
	}
	history, _ = s.History(ctx)
	if history[0].Status != "sent" || history[0].Body != "" {
		t.Fatal("completed payload retained")
	}
	c.Enabled = false
	if _, err = s.Save(ctx, c); err != nil {
		t.Fatal(err)
	}
}

func TestQuietAndDisabledDoNotSend(t *testing.T) {
	s, db := testService(t)
	ctx := context.Background()
	c := testConfig()
	c.QuietEnabled = true
	c.QuietStart = "00:00"
	c.QuietEnd = "23:59"
	if _, err := s.Save(ctx, c); err != nil {
		t.Fatal(err)
	}
	addArticle(t, db, "Queued")
	s.client = func() (*http.Client, error) { t.Fatal("network during quiet hours"); return nil, nil }
	now := time.Date(2030, 1, 1, 12, 0, 0, 0, time.UTC)
	if err := s.Tick(ctx, now); err != nil {
		t.Fatal(err)
	}
	history, _ := s.History(ctx)
	if len(history) != 0 {
		t.Fatal("quiet hours should defer scanning")
	}
}

func TestChangedRuleCancelsPendingAndStartsAtCurrentID(t *testing.T) {
	s, db := testService(t)
	ctx := context.Background()
	c, err := s.Save(ctx, testConfig())
	if err != nil {
		t.Fatal(err)
	}
	id := addArticle(t, db, "Old destination")
	if err = db.QueueNotification(ctx, "rule:telegram", "telegram", "rule", "test", id, time.Now(), []string{"private text"}); err != nil {
		t.Fatal(err)
	}
	c.Channels[0].ChatID = "another-recipient"
	if _, err = s.Save(ctx, c); err != nil {
		t.Fatal(err)
	}
	history, _ := s.History(ctx)
	if len(history) != 1 || history[0].Status != "cancelled" || history[0].Body != "" {
		t.Fatal(history)
	}
	state, err := db.GetNotificationState(ctx, "rule:telegram")
	if err != nil || state.Cursor != id {
		t.Fatal(state, err)
	}
}

func TestPreviewNoSideEffectsAndBatchLimit(t *testing.T) {
	s, db := testService(t)
	ctx := context.Background()
	c := testConfig()
	c.Rules[0].MaxItems = 1
	if _, err := s.Save(ctx, c); err != nil {
		t.Fatal(err)
	}
	addArticle(t, db, "First")
	addArticle(t, db, "Second")
	p, err := s.Preview(ctx, c.Rules[0])
	if err != nil || p.Matched != 2 || len(p.Messages) != 1 {
		t.Fatal(p, err)
	}
	history, _ := s.History(ctx)
	if len(history) != 0 {
		t.Fatal("preview queued messages")
	}
	s.client = func() (*http.Client, error) {
		return mockClient(func(*http.Request) (*http.Response, error) { return reply(200, `{"ok":true}`), nil }), nil
	}
	now := time.Now().Add(2 * time.Minute)
	if err = s.Tick(ctx, now); err != nil {
		t.Fatal(err)
	}
	if err = s.Tick(ctx, now.Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	history, _ = s.History(ctx)
	if len(history) != 2 {
		t.Fatal("article above batch limit was lost")
	}
}

func TestSelectiveDigestScansBeyondFirstPage(t *testing.T) {
	s, db := testService(t)
	ctx := context.Background()
	c := testConfig()
	c.Rules[0].Mode = "digest"
	c.Rules[0].Conditions = []Condition{{Field: "title", Operator: "contains", Value: "important"}}
	if _, err := s.Save(ctx, c); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 510; i++ {
		addArticle(t, db, "Unrelated")
	}
	addArticle(t, db, "Important release")
	calls := 0
	s.client = func() (*http.Client, error) {
		return mockClient(func(req *http.Request) (*http.Response, error) {
			calls++
			body, _ := io.ReadAll(req.Body)
			if !strings.Contains(string(body), "Important release") {
				t.Fatal("wrong article")
			}
			return reply(200, `{"ok":true}`), nil
		}), nil
	}
	if err := s.Tick(ctx, time.Date(2030, 1, 1, 10, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("digest did not scan past first page")
	}
}

func TestRetriesStopAfterThreeAttempts(t *testing.T) {
	s, db := testService(t)
	ctx := context.Background()
	if _, err := s.Save(ctx, testConfig()); err != nil {
		t.Fatal(err)
	}
	addArticle(t, db, "Release")
	calls := 0
	s.client = func() (*http.Client, error) {
		return mockClient(func(*http.Request) (*http.Response, error) { calls++; return reply(503, "unavailable"), nil }), nil
	}
	now := time.Now().Add(time.Minute * 2)
	for i := 0; i < 5; i++ {
		if err := s.Tick(ctx, now.Add(time.Duration(i)*10*time.Minute)); err != nil {
			t.Fatal(err)
		}
	}
	history, err := s.History(ctx)
	if err != nil || calls != 3 || len(history) != 1 || history[0].Status != "failed" || history[0].Body != "" {
		t.Fatal(calls, history, err)
	}
}

func TestConcurrentTicksDoNotDuplicate(t *testing.T) {
	s, db := testService(t)
	ctx := context.Background()
	if _, err := s.Save(ctx, testConfig()); err != nil {
		t.Fatal(err)
	}
	addArticle(t, db, "Concurrent release")
	s.client = func() (*http.Client, error) {
		return mockClient(func(*http.Request) (*http.Response, error) { return reply(200, `{"ok":true}`), nil }), nil
	}
	now := time.Now().Add(2 * time.Minute)
	results := make(chan error, 8)
	for i := 0; i < 8; i++ {
		go func() { results <- s.Tick(ctx, now) }()
	}
	for i := 0; i < 8; i++ {
		if err := <-results; err != nil {
			t.Fatal(err)
		}
	}
	history, err := s.History(ctx)
	if err != nil || len(history) != 1 || history[0].Attempts != 1 {
		t.Fatal("concurrent duplicate", history, err)
	}
}

func TestOneBrokenChannelDoesNotBlockOtherDestinations(t *testing.T) {
	s, db := testService(t)
	ctx := context.Background()
	c := testConfig()
	c.Channels = []Channel{testChannel("telegram"), testChannel("discord"), testChannel("feishu")}
	c.Rules[0].ChannelIDs = []string{"telegram", "discord", "feishu"}
	if _, err := s.Save(ctx, c); err != nil {
		t.Fatal(err)
	}
	addArticle(t, db, "Release for all destinations")
	s.client = func() (*http.Client, error) {
		return mockClient(func(req *http.Request) (*http.Response, error) {
			switch req.URL.Host {
			case "api.telegram.org":
				return reply(401, `{}`), nil
			case "discord.com":
				return reply(200, `{"id":"message"}`), nil
			default:
				return reply(200, `{"code":0}`), nil
			}
		}), nil
	}
	if err := s.Tick(ctx, time.Now().Add(2*time.Minute)); err != nil {
		t.Fatal(err)
	}
	history, err := s.History(ctx)
	if err != nil || len(history) != 3 {
		t.Fatal(history, err)
	}
	for _, d := range history {
		want := "sent"
		if d.ChannelID == "telegram" {
			want = "failed"
		}
		if d.Status != want {
			t.Fatal(d)
		}
	}
}

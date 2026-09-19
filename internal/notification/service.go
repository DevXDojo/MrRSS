package notification

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"reflect"
	"time"

	"MrRSS/internal/crypto"
	"MrRSS/internal/database"
	"MrRSS/internal/utils/httputil"
)

var ErrConflict = errors.New("conflict")

type Service struct {
	db   *database.DB
	gate chan struct{}
	// Factory keeps delivery testable without live credentials or services.
	client func() (*http.Client, error)
}

func New(db *database.DB) *Service {
	return &Service{db: db, gate: make(chan struct{}, 1), client: func() (*http.Client, error) { return httputil.CreateHTTPClientWithProxySettings(db, 20*time.Second) }}
}

func (s *Service) lock(ctx context.Context) error {
	select {
	case s.gate <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (s *Service) unlock() { <-s.gate }

func (s *Service) Load(ctx context.Context) (Config, error) {
	raw, err := s.db.GetSettingContext(ctx, "notification_config")
	if err != nil {
		return Config{}, err
	}
	c := DefaultConfig()
	if raw == "" {
		return c, nil
	}
	if crypto.IsEncrypted(raw) {
		raw, err = crypto.Decrypt(raw)
		if err != nil {
			return Config{}, errors.New("credentials_unavailable")
		}
	}
	if err = json.Unmarshal([]byte(raw), &c); err != nil {
		return Config{}, errors.New("config_invalid")
	}
	c.normalize()
	return c, nil
}

func activePairs(c Config) map[string]bool {
	active := map[string]bool{}
	if !c.Enabled {
		return active
	}
	channels := map[string]bool{}
	for _, ch := range c.Channels {
		channels[ch.ID] = ch.Enabled
	}
	for _, r := range c.Rules {
		if r.Enabled {
			for _, id := range r.ChannelIDs {
				if channels[id] {
					active[r.ID+":"+id] = true
				}
			}
		}
	}
	return active
}

func (s *Service) Save(ctx context.Context, c Config) (Config, error) {
	c.normalize()
	if err := s.lock(ctx); err != nil {
		return Config{}, err
	}
	defer s.unlock()
	old, err := s.Load(ctx)
	if err != nil {
		return Config{}, err
	}
	if c.Revision != old.Revision {
		return Config{}, ErrConflict
	}
	if err = c.RestoreSecrets(old); err != nil {
		return Config{}, err
	}
	if err = Validate(c); err != nil {
		return Config{}, err
	}
	reset := map[string]bool{}
	for _, r := range c.Rules {
		for _, p := range old.Rules {
			if r.ID == p.ID && !reflect.DeepEqual(r, p) {
				for _, id := range r.ChannelIDs {
					reset[r.ID+":"+id] = true
				}
			}
		}
	}
	for _, ch := range c.Channels {
		for _, p := range old.Channels {
			// Changing the destination must not send already queued articles to a
			// different recipient. Name-only changes retain the baseline.
			ch.Name = p.Name
			if ch.ID == p.ID && !reflect.DeepEqual(ch, p) {
				for _, r := range c.Rules {
					reset[r.ID+":"+ch.ID] = true
				}
			}
		}
	}
	c.Revision = old.Revision + 1
	raw, err := json.Marshal(c)
	if err != nil {
		return Config{}, err
	}
	encrypted, err := crypto.Encrypt(string(raw))
	if err != nil {
		return Config{}, errors.New("credentials_unavailable")
	}
	if err = s.db.SaveNotificationConfig(ctx, encrypted, activePairs(c), reset, time.Now()); err != nil {
		return Config{}, err
	}
	return c.Redacted(), nil
}

func (s *Service) Test(ctx context.Context, ch Channel) error {
	if err := s.lock(ctx); err != nil {
		return err
	}
	defer s.unlock()
	old, err := s.Load(ctx)
	if err != nil {
		return err
	}
	c := Config{Channels: []Channel{ch}}
	if err = c.RestoreSecrets(old); err != nil {
		return err
	}
	client, err := s.client()
	if err != nil {
		return errors.New("proxy")
	}
	defer client.CloseIdleConnections()
	return Send(ctx, client, c.Channels[0], "MrRSS · Test notification / 测试推送\nYour notification channel is connected. 消息推送渠道已连通。", time.Now())
}

type Preview struct {
	Matched  int      `json:"matched"`
	Scanned  int      `json:"scanned"`
	Messages []string `json:"messages"`
}

func (s *Service) Preview(ctx context.Context, r Rule) (Preview, error) {
	// Use the same validation as a saved rule, without requiring real secrets.
	c := DefaultConfig()
	c.Channels = []Channel{{ID: "preview", Name: "preview", Provider: "telegram", Token: "1:test", ChatID: "1"}}
	r.ChannelIDs = []string{"preview"}
	c.Rules = []Rule{r}
	if err := Validate(c); err != nil {
		return Preview{}, err
	}
	articles, err := s.db.NotificationArticles(ctx, 0, true)
	if err != nil {
		return Preview{}, err
	}
	p := Preview{Scanned: len(articles), Messages: []string{}}
	selected := []database.NotificationArticle{}
	for _, a := range articles {
		if matches(a, r) {
			p.Matched++
			if len(selected) < r.MaxItems {
				selected = append(selected, a)
			}
		}
	}
	if messages := render(r, selected); messages != nil {
		p.Messages = messages
	}
	return p, nil
}

func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			work, cancel := context.WithTimeout(ctx, 50*time.Second)
			if err := s.Tick(work, time.Now()); err != nil && work.Err() == nil {
				log.Print("Notification scheduler could not complete this cycle")
			}
			cancel()
		}
	}
}

func (s *Service) Tick(ctx context.Context, now time.Time) error {
	if err := s.lock(ctx); err != nil {
		return err
	}
	defer s.unlock()
	c, err := s.Load(ctx)
	if err != nil {
		return err
	}
	if !c.Enabled {
		return nil
	}
	if err = Validate(c); err != nil {
		return err
	}
	if c.IsQuiet(now) {
		return nil
	}
	channels := map[string]Channel{}
	for _, ch := range c.Channels {
		if ch.Enabled {
			channels[ch.ID] = ch
		}
	}
	for _, r := range c.Rules {
		if !r.Enabled {
			continue
		}
		for _, id := range r.ChannelIDs {
			if _, ok := channels[id]; !ok {
				continue
			}
			key := r.ID + ":" + id
			state, e := s.db.GetNotificationState(ctx, key)
			if e != nil {
				return e
			}
			if !due(r, c, state.LastRun, now) {
				continue
			}
			pending, e := s.db.NotificationPendingCount(ctx, key)
			if e != nil {
				return e
			}
			if pending > 0 {
				continue
			}
			selected := []database.NotificationArticle{}
			cursor := state.Cursor
			lastRun := now
			// Scan in bounded pages so selective daily rules can see beyond the
			// first 500 incoming articles. Checkpoint long unmatched backlogs
			// without consuming today's scheduled run.
			for page := 0; page < 20; page++ {
				articles, e := s.db.NotificationArticles(ctx, cursor, false)
				if e != nil {
					return e
				}
				full := false
				for _, a := range articles {
					// Leave additional matching articles for the next run; nothing is
					// silently dropped when the user sets a small message limit.
					if matches(a, r) {
						if len(selected) >= r.MaxItems {
							full = true
							break
						}
						selected = append(selected, a)
					}
					cursor = a.ID
				}
				if full || len(articles) < 500 {
					break
				}
				if page == 19 && len(selected) == 0 {
					lastRun = state.LastRun
				}
			}
			if e = s.db.QueueNotification(ctx, key, id, r.ID, r.Name, cursor, lastRun, render(r, selected)); e != nil {
				return e
			}
		}
	}
	deliveries, err := s.db.NotificationDeliveries(ctx, true, now)
	if err != nil {
		return err
	}
	client, err := s.client()
	if err != nil {
		return errors.New("proxy")
	}
	defer client.CloseIdleConnections()
	for _, d := range deliveries {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		ch, ok := channels[d.ChannelID]
		if !ok {
			continue
		}
		status, code, next := "sent", "", now
		if sendErr := Send(ctx, client, ch, d.Body, now); sendErr != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			var retry bool
			var after time.Duration
			code, retry, after = deliveryFailure(sendErr)
			status = "failed"
			if retry && d.Attempts < 2 {
				status = "pending"
				backoff := time.Duration(1<<d.Attempts) * time.Minute
				if after < backoff {
					after = backoff
				}
				next = now.Add(after)
			}
		}
		if err = s.db.CompleteNotification(ctx, d.ID, status, code, next); err != nil {
			return err
		}
	}
	return s.db.PruneNotifications(ctx, now)
}

func (s *Service) History(ctx context.Context) ([]database.NotificationDelivery, error) {
	return s.db.NotificationDeliveries(ctx, false, time.Now())
}

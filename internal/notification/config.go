// Package notification delivers opt-in article alerts and scheduled briefings.
package notification

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
	_ "time/tzdata" // IANA zones are also available in Windows desktop builds.
	"unicode/utf8"
)

const SecretMask = "••••••••"

type Config struct {
	Revision     int64     `json:"revision"`
	Enabled      bool      `json:"enabled"`
	Timezone     string    `json:"timezone"`
	QuietEnabled bool      `json:"quiet_enabled"`
	QuietStart   string    `json:"quiet_start"`
	QuietEnd     string    `json:"quiet_end"`
	Channels     []Channel `json:"channels"`
	Rules        []Rule    `json:"rules"`
}

type Channel struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Enabled  bool   `json:"enabled"`
	Token    string `json:"token"`
	ChatID   string `json:"chat_id"`
	ThreadID int64  `json:"thread_id"`
	Webhook  string `json:"webhook"`
	Secret   string `json:"secret"`
}

type Rule struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	Enabled         bool        `json:"enabled"`
	Mode            string      `json:"mode"` // alert: interval, digest: daily local time
	ChannelIDs      []string    `json:"channel_ids"`
	Match           string      `json:"match"` // all or any
	Conditions      []Condition `json:"conditions"`
	IntervalMinutes int         `json:"interval_minutes"`
	Time            string      `json:"time"`
	Weekdays        []int       `json:"weekdays"` // 0=Sunday; empty means every day
	MaxItems        int         `json:"max_items"`
	IncludeSummary  bool        `json:"include_summary"`
	IncludeLink     bool        `json:"include_link"`
	UnreadOnly      bool        `json:"unread_only"`
}

type Condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

func DefaultConfig() Config {
	return Config{Timezone: "Local", QuietStart: "22:00", QuietEnd: "08:00", Channels: []Channel{}, Rules: []Rule{}}
}

// Keep API collection shapes stable for clients, including configurations sent
// by scripts that omit optional condition or weekday lists.
func (c *Config) normalize() {
	if c.Channels == nil {
		c.Channels = []Channel{}
	}
	if c.Rules == nil {
		c.Rules = []Rule{}
	}
	for i := range c.Rules {
		if c.Rules[i].Conditions == nil {
			c.Rules[i].Conditions = []Condition{}
		}
		if c.Rules[i].Weekdays == nil {
			c.Rules[i].Weekdays = []int{}
		}
	}
}

var identifier = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,80}$`)
var botToken = regexp.MustCompile(`^[0-9]+:[A-Za-z0-9_-]+$`)
var discordPath = regexp.MustCompile(`^/api(?:/v[0-9]+)?/webhooks/[0-9]+/[A-Za-z0-9._-]+$`)
var feishuPath = regexp.MustCompile(`^/open-apis/bot/v2/hook/[A-Za-z0-9_-]+$`)

func ValidateChannel(c Channel) error {
	if !identifier.MatchString(c.ID) || strings.TrimSpace(c.Name) == "" || utf8.RuneCountInString(c.Name) > 120 {
		return errors.New("channel_name")
	}
	switch c.Provider {
	case "telegram":
		if !botToken.MatchString(c.Token) || strings.TrimSpace(c.ChatID) == "" || len(c.ChatID) > 120 || c.ThreadID < 0 {
			return errors.New("telegram_credentials")
		}
	case "discord", "feishu":
		u, err := url.Parse(c.Webhook)
		if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || u.RawQuery != "" || u.Fragment != "" {
			return errors.New("webhook_url")
		}
		valid := false
		if c.Provider == "discord" {
			valid = (u.Host == "discord.com" || u.Host == "discordapp.com" || u.Host == "canary.discord.com" || u.Host == "ptb.discord.com") && discordPath.MatchString(u.Path)
		}
		if c.Provider == "feishu" {
			valid = (u.Host == "open.feishu.cn" || u.Host == "open.larksuite.com") && feishuPath.MatchString(u.Path)
		}
		if !valid || len(c.Secret) > 256 {
			return errors.New("webhook_url")
		}
	default:
		return errors.New("provider")
	}
	return nil
}

func Validate(c Config) error {
	if _, err := time.LoadLocation(c.Timezone); err != nil || c.Timezone == "" {
		return errors.New("timezone")
	}
	if !validTime(c.QuietStart) || !validTime(c.QuietEnd) || (c.QuietEnabled && c.QuietStart == c.QuietEnd) {
		return errors.New("quiet_time")
	}
	if len(c.Channels) > 20 || len(c.Rules) > 50 {
		return errors.New("limit")
	}
	channels := map[string]bool{}
	for _, ch := range c.Channels {
		if err := ValidateChannel(ch); err != nil {
			return err
		}
		if channels[ch.ID] {
			return errors.New("duplicate_id")
		}
		channels[ch.ID] = true
	}
	rules := map[string]bool{}
	for _, r := range c.Rules {
		if !identifier.MatchString(r.ID) || rules[r.ID] {
			return errors.New("duplicate_id")
		}
		rules[r.ID] = true
		if strings.TrimSpace(r.Name) == "" || utf8.RuneCountInString(r.Name) > 120 {
			return errors.New("rule_name")
		}
		if r.Mode != "alert" && r.Mode != "digest" {
			return errors.New("mode")
		}
		if r.Match != "all" && r.Match != "any" {
			return errors.New("match")
		}
		if r.IntervalMinutes < 1 || r.IntervalMinutes > 1440 || !validTime(r.Time) || r.MaxItems < 1 || r.MaxItems > 20 {
			return errors.New("schedule")
		}
		for _, d := range r.Weekdays {
			if d < 0 || d > 6 {
				return errors.New("schedule")
			}
		}
		if len(r.ChannelIDs) == 0 || len(r.ChannelIDs) > 20 || len(r.Conditions) > 10 {
			return errors.New("rule_channels")
		}
		seen := map[string]bool{}
		for _, id := range r.ChannelIDs {
			if !channels[id] || seen[id] {
				return errors.New("rule_channels")
			}
			seen[id] = true
		}
		for _, v := range r.Conditions {
			switch v.Field {
			case "title", "summary", "feed", "category", "author", "url", "favorite", "read_later":
			default:
				return errors.New("condition")
			}
			switch v.Operator {
			case "contains", "not_contains", "equals", "not_equals":
			default:
				return errors.New("condition")
			}
			if strings.TrimSpace(v.Value) == "" || utf8.RuneCountInString(v.Value) > 500 {
				return errors.New("condition")
			}
			if (v.Field == "favorite" || v.Field == "read_later") && ((v.Value != "true" && v.Value != "false") || (v.Operator != "equals" && v.Operator != "not_equals")) {
				return errors.New("condition")
			}
		}
	}
	return nil
}

func validTime(s string) bool {
	t, e := time.Parse("15:04", s)
	return e == nil && t.Format("15:04") == s
}

func (c Config) Redacted() Config {
	c.Channels = append([]Channel{}, c.Channels...)
	for i := range c.Channels {
		ch := &c.Channels[i]
		for _, s := range []*string{&ch.Token, &ch.Webhook, &ch.Secret} {
			if *s != "" {
				*s = SecretMask
			}
		}
	}
	return c
}

func (c *Config) RestoreSecrets(old Config) error {
	for i := range c.Channels {
		n := &c.Channels[i]
		var previous Channel
		for _, p := range old.Channels {
			if p.ID == n.ID && p.Provider == n.Provider {
				previous = p
				break
			}
		}
		for _, pair := range [][2]*string{{&n.Token, &previous.Token}, {&n.Webhook, &previous.Webhook}, {&n.Secret, &previous.Secret}} {
			if *pair[0] == SecretMask {
				if *pair[1] == "" {
					return errors.New("credentials_missing")
				}
				*pair[0] = *pair[1]
			}
		}
	}
	return nil
}

func (c Config) IsQuiet(now time.Time) bool {
	if !c.QuietEnabled {
		return false
	}
	loc, err := time.LoadLocation(c.Timezone)
	if err != nil {
		return true
	}
	t := now.In(loc).Format("15:04")
	if c.QuietStart < c.QuietEnd {
		return t >= c.QuietStart && t < c.QuietEnd
	}
	return t >= c.QuietStart || t < c.QuietEnd
}

func due(r Rule, c Config, last, now time.Time) bool {
	if r.Mode == "alert" {
		return now.Sub(last) >= time.Duration(r.IntervalMinutes)*time.Minute
	}
	loc, err := time.LoadLocation(c.Timezone)
	if err != nil {
		return false
	}
	local := now.In(loc)
	if len(r.Weekdays) > 0 {
		found := false
		for _, d := range r.Weekdays {
			if d == int(local.Weekday()) {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	clock, err := time.Parse("15:04", r.Time)
	if err != nil {
		return false
	}
	scheduled := time.Date(local.Year(), local.Month(), local.Day(), clock.Hour(), clock.Minute(), 0, 0, loc)
	return !now.Before(scheduled) && last.Before(scheduled)
}

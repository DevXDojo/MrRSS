package database

import (
	"MrRSS/internal/crypto"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/andybalholm/cascadia"
	"net/url"
	"strings"
)

type FeedContentOptions struct {
	Cookie          *string `json:"cookie,omitempty"`
	CookieOrigin    string  `json:"cookie_origin"`
	HasCookie       bool    `json:"has_cookie"`
	ContentSelector string  `json:"content_selector"`
	RemoveSelector  string  `json:"remove_selector"`
}

func (o FeedContentOptions) Validate() error {
	if o.Cookie != nil && (len(*o.Cookie) > 8192 || strings.IndexFunc(*o.Cookie, func(r rune) bool { return r < 32 || r == 127 }) >= 0) {
		return fmt.Errorf("invalid Cookie header")
	}
	if o.CookieOrigin != "" || (o.Cookie != nil && *o.Cookie != "") {
		origin, err := url.Parse(o.CookieOrigin)
		if err != nil || origin.Host == "" || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" || (origin.Path != "" && origin.Path != "/") || (origin.Scheme != "http" && origin.Scheme != "https") {
			return fmt.Errorf("Cookie website must be an HTTP(S) origin without a path")
		}
	}

	for _, selector := range []string{o.ContentSelector, o.RemoveSelector} {
		if len(selector) > 4096 {
			return fmt.Errorf("selector exceeds 4096 bytes")
		}
		if strings.TrimSpace(selector) != "" {
			if _, err := cascadia.Compile(selector); err != nil {
				return fmt.Errorf("invalid CSS selector: %w", err)
			}
		}
	}
	return nil
}

func (db *DB) GetFeedContentOptions(ctx context.Context, feedID int64) (FeedContentOptions, error) {
	db.WaitForReady()
	var options FeedContentOptions
	stmt, err := db.PrepareContext(ctx, `SELECT content_selector, remove_selector, cookie, cookie_origin FROM feed_content_options WHERE feed_id = ?`)
	if err != nil {
		return options, fmt.Errorf("prepare content options: %w", err)
	}
	defer stmt.Close()
	var encrypted string
	err = stmt.QueryRowContext(ctx, feedID).Scan(&options.ContentSelector, &options.RemoveSelector, &encrypted, &options.CookieOrigin)
	if errors.Is(err, sql.ErrNoRows) {
		return options, nil
	}
	if err != nil {
		return options, err
	}
	cookie, err := crypto.Decrypt(encrypted)
	if err != nil {
		return options, fmt.Errorf("decrypt feed Cookie: %w", err)
	}
	options.Cookie = &cookie
	options.HasCookie = cookie != ""
	return options, nil
}

func (db *DB) SetFeedContentOptions(ctx context.Context, feedID int64, options FeedContentOptions) error {
	db.WaitForReady()
	if options.Cookie == nil {
		existing, err := db.GetFeedContentOptions(ctx, feedID)
		if err != nil {
			return err
		}
		options.Cookie = existing.Cookie
	}
	if err := options.Validate(); err != nil {
		return err
	}
	cookie := ""
	if options.Cookie != nil {
		cookie = *options.Cookie
	}
	encrypted, err := crypto.Encrypt(cookie)
	if err != nil {
		return fmt.Errorf("encrypt feed Cookie: %w", err)
	}
	stmt, err := db.PrepareContext(ctx, `INSERT INTO feed_content_options (feed_id, content_selector, remove_selector, cookie, cookie_origin) VALUES (?, ?, ?, ?, ?)
 ON CONFLICT(feed_id) DO UPDATE SET content_selector = excluded.content_selector, remove_selector = excluded.remove_selector, cookie = excluded.cookie, cookie_origin = excluded.cookie_origin`)
	if err != nil {
		return fmt.Errorf("prepare content options: %w", err)
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, feedID, options.ContentSelector, options.RemoveSelector, encrypted, options.CookieOrigin)
	return err
}

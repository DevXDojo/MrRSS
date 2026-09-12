package tray

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

type unreadSource struct {
	setting              string
	count                int
	settingErr, countErr error
	countReads           int
}

func (s *unreadSource) GetSettingContext(context.Context, string) (string, error) {
	return s.setting, s.settingErr
}

func (s *unreadSource) GetTotalUnreadCountContext(context.Context) (int, error) {
	s.countReads++
	return s.count, s.countErr
}

func TestUnreadLabel(t *testing.T) {
	failure := errors.New("database unavailable")
	for _, tc := range []struct {
		name    string
		source  unreadSource
		want    string
		wantErr error
	}{
		{name: "unread", source: unreadSource{setting: "true", count: 42}, want: "42"},
		{name: "zero", source: unreadSource{setting: "true"}},
		{name: "hidden", source: unreadSource{setting: "false", count: 42}},
		{name: "legacy default", source: unreadSource{settingErr: sql.ErrNoRows, count: 8}, want: "8"},
		{name: "setting failure", source: unreadSource{settingErr: failure}, wantErr: failure},
		{name: "count failure", source: unreadSource{countErr: failure}, wantErr: failure},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := UnreadLabel(context.Background(), &tc.source)
			if got != tc.want || !errors.Is(err, tc.wantErr) {
				t.Fatalf("UnreadLabel = %q, %v; want %q, %v", got, err, tc.want, tc.wantErr)
			}
			if tc.source.setting == "false" && tc.source.countReads != 0 {
				t.Fatal("hidden indicator queried article counts")
			}
		})
	}
}

func TestWatchUnreadStopsOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	source := &unreadSource{count: 7}
	calls := 0
	WatchUnread(ctx, source, func(label string) {
		calls++
		if label != "7" {
			t.Errorf("label = %q", label)
		}
		cancel()
	})
	if calls != 1 || source.countReads != 1 {
		t.Fatalf("unexpected updates: %d, reads: %d", calls, source.countReads)
	}
}

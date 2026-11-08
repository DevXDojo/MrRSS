package freshrss

import (
	"context"
	"testing"
	"time"

	"MrRSS/internal/database"
)

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

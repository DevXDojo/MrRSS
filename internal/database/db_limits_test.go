package database_test

import (
	"context"
	"testing"

	dbpkg "MrRSS/internal/database"
)

// TestNewDBMemoryLimits pins the limits that bound how much memory the database
// layer can hold: the connection-pool size and the per-connection SQLite page
// cache (25 * 8000 KiB = 200 MiB worst case, against 25 * 32000 KiB = 800 MiB
// before). It asserts MaxOpenConnections and the PRAGMA that reaches the
// connection; the idle limit has no direct sql.DBStats field and is covered by
// the constant plus code review rather than by this test. Raising any of these
// values raises that ceiling; update this test only together with a deliberate
// decision about the new budget.
func TestNewDBMemoryLimits(t *testing.T) {
	db, err := dbpkg.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB() error = %v", err)
	}
	defer db.Close()

	if got := db.Stats().MaxOpenConnections; got != 25 {
		t.Errorf("MaxOpenConnections = %d, want 25", got)
	}

	// Pin a single connection so the PRAGMA reports the value applied from the
	// DSN instead of a coincidental default.
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn() error = %v", err)
	}
	defer conn.Close()

	var cacheSize int64
	if err := conn.QueryRowContext(ctx, "PRAGMA cache_size").Scan(&cacheSize); err != nil {
		t.Fatalf("PRAGMA cache_size error = %v", err)
	}
	if cacheSize != -8000 {
		t.Errorf("PRAGMA cache_size = %d, want -8000 (KiB)", cacheSize)
	}
}

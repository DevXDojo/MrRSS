package database

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// DB wraps sql.DB with initialization state tracking.
type DB struct {
	*sql.DB
	ready chan struct{}
	once  sync.Once
}

// NewDB creates a new database connection with optimized settings.
func NewDB(dataSourceName string) (*DB, error) {
	// Add busy_timeout to prevent "database is locked" errors
	// Also enable WAL mode for better concurrency
	// Add performance optimizations: increase cache size, set synchronous=NORMAL
	// Enable foreign_keys to make ON DELETE CASCADE work (disabled by default in SQLite)
	if !strings.Contains(dataSourceName, "?") {
		dataSourceName += "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=cache_size(-32000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(1)"
	} else {
		dataSourceName += "&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=cache_size(-32000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(1)"
	}

	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Set connection pool limits for better performance
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &DB{
		DB:    db,
		ready: make(chan struct{}),
	}, nil
}

// WaitForReady blocks until the database is initialized.
func (db *DB) WaitForReady() {
	<-db.ready
}

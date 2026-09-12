package database

import (
	"errors"
	"fmt"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// explainInitializationError distinguishes storage access failures from damage.
// It only annotates the error; it never removes or rebuilds user data.
func explainInitializationError(err error) error {
	var sqliteErr *sqlite.Error
	if !errors.As(err, &sqliteErr) {
		return err
	}

	var hint string
	switch sqliteErr.Code() & 0xff {
	case sqlite3.SQLITE_CANTOPEN:
		hint = "database file could not be opened; check that the data directory is available and that the current user can access the database and create its journal files"
	case sqlite3.SQLITE_READONLY, sqlite3.SQLITE_PERM:
		hint = "database storage is not writable; check directory and file permissions and whether the disk is mounted read-only"
	case sqlite3.SQLITE_FULL:
		hint = "database storage is full; check free space on both the data and temporary-file disks"
	case sqlite3.SQLITE_CORRUPT, sqlite3.SQLITE_NOTADB:
		hint = "database contents could not be read as a valid SQLite database; stop MrRSS and back up the complete data directory before recovery"
	default:
		return err
	}
	return fmt.Errorf("%s (do not delete rss.db, rss.db-wal, or rss.db-shm as a troubleshooting step): %w", hint, err)
}

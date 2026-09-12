package database

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

func TestInitializationExplainsUnopenableDatabase(t *testing.T) {
	// A missing parent fails consistently without relying on permission bits
	// (which behave differently when tests run elevated or on Windows).
	db, err := NewDB(filepath.Join(t.TempDir(), "missing", "rss.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = db.Init()
	var sqliteErr *sqlite.Error
	if !errors.As(err, &sqliteErr) || sqliteErr.Code()&0xff != sqlite3.SQLITE_CANTOPEN {
		t.Fatalf("expected preserved SQLITE_CANTOPEN, got %v", err)
	}
	if !strings.Contains(err.Error(), "data directory") || strings.Contains(strings.ToLower(err.Error()), "out of memory") {
		t.Fatalf("unexpected diagnostic: %v", err)
	}
}

func TestInitializationDoesNotRepairInvalidDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rss.db")
	contents := []byte(strings.Repeat("not a SQLite database\n", 100))
	if err := os.WriteFile(path, contents, 0600); err != nil {
		t.Fatal(err)
	}
	db, err := NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Init()
	_ = db.Close()
	var sqliteErr *sqlite.Error
	if !errors.As(err, &sqliteErr) || sqliteErr.Code()&0xff != sqlite3.SQLITE_NOTADB {
		t.Fatalf("expected preserved SQLITE_NOTADB, got %v", err)
	}
	if !strings.Contains(err.Error(), "back up the complete data directory") {
		t.Fatalf("missing recovery guidance: %v", err)
	}
	actual, readErr := os.ReadFile(path)
	if readErr != nil || string(actual) != string(contents) {
		t.Fatalf("database changed during diagnosis: %v", readErr)
	}
}

func TestInitializationLeavesOtherErrorsIntact(t *testing.T) {
	if got := explainInitializationError(nil); got != nil {
		t.Fatalf("nil became %v", got)
	}
	err := fmt.Errorf("some other failure")
	if got := explainInitializationError(err); got != err {
		t.Fatalf("unrelated error changed: %v", got)
	}
}

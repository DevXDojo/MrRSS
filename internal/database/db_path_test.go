package database

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDatabasePathWithURIDelimiters(t *testing.T) {
	name := "reader #50% 中文"
	if runtime.GOOS != "windows" {
		name += "?mode=ro"
	}
	dir := filepath.Join(t.TempDir(), name)
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "rss.db")
	db, err := NewDB(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec("CREATE TABLE test_path (value TEXT)"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("database was not created at chosen path: %v", err)
	}
}

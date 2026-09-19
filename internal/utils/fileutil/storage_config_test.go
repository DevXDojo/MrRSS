package fileutil

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestStorageMigrationPreservesSQLiteRecords(t *testing.T) {
	source, target := t.TempDir(), t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(source, "rss.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`PRAGMA journal_mode=WAL; CREATE TABLE articles (title TEXT, is_read INTEGER, favorite INTEGER); INSERT INTO articles VALUES ('迁移后仍可阅读', 1, 1)`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := migrateStorage(source, target); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{source, target} {
		copied, err := sql.Open("sqlite", filepath.Join(dir, "rss.db"))
		if err != nil {
			t.Fatal(err)
		}
		var integrity, title string
		var read, favorite int
		err = copied.QueryRow("PRAGMA integrity_check").Scan(&integrity)
		if err != nil || integrity != "ok" {
			copied.Close()
			t.Fatalf("database integrity: %s %v", integrity, err)
		}
		err = copied.QueryRow("SELECT title, is_read, favorite FROM articles").Scan(&title, &read, &favorite)
		copied.Close()
		if err != nil || title != "迁移后仍可阅读" || read != 1 || favorite != 1 {
			t.Fatalf("lost article state: %q %d %d %v", title, read, favorite, err)
		}
	}
}

func TestStorageMigrationPreservesFilesAndRejectsOccupiedTarget(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "old")
	target := filepath.Join(root, "新的目录 # %")
	os.MkdirAll(filepath.Join(source, "scripts"), 0700)
	os.Mkdir(target, 0700)
	for name, value := range map[string]string{"rss.db": "db bytes", "rss.db-wal": "wal bytes", "scripts/feed.py": "print('rss')", "settings.key": "encrypted"} {
		if err := os.WriteFile(filepath.Join(source, name), []byte(value), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := migrateStorage(source, target); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"rss.db", "rss.db-wal", "scripts/feed.py", "settings.key"} {
		old, _ := os.ReadFile(filepath.Join(source, name))
		copied, err := os.ReadFile(filepath.Join(target, name))
		if err != nil || string(old) != string(copied) {
			t.Fatalf("file %s not preserved: %v", name, err)
		}
	}
	if err := migrateStorage(source, target); err == nil {
		t.Fatal("accepted occupied target")
	}
	inside := filepath.Join(source, "nested")
	os.Mkdir(inside, 0700)
	if _, err := validateStorageDestination(source, inside); err == nil {
		t.Fatal("accepted nested target")
	}
}
func TestStorageScheduleDoesNotSwitchLiveDirectory(t *testing.T) {
	previous, managed, configFile := customDataDir, desktopStorageManaged, storageConfigFile
	t.Cleanup(func() { customDataDir, desktopStorageManaged, storageConfigFile = previous, managed, configFile })
	source, target := t.TempDir(), t.TempDir()
	customDataDir, desktopStorageManaged, storageConfigFile = source, true, filepath.Join(t.TempDir(), "storage.json")
	cfg, err := ScheduleDataDirectory(target)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PendingDirectory != target || customDataDir != source {
		t.Fatal("changed live directory")
	}
	if err := CancelDataDirectoryChange(); err != nil {
		t.Fatal(err)
	}
	cfg, err = DesktopStorageStatus()
	if err != nil || cfg.PendingDirectory != "" {
		t.Fatalf("cancel failed: %v", err)
	}
	if args := DataDirRestartArgs([]string{"--software-rendering"}); len(args) != 1 {
		t.Fatalf("managed path pinned in restart: %v", args)
	}
}
func TestStorageUseLock(t *testing.T) {
	dir := t.TempDir()
	lock, err := acquireStorageUse(dir)
	if err != nil {
		t.Fatal(err)
	}
	if other, err := acquireStorageUse(dir); err == nil {
		other.Close()
		t.Fatal("concurrent lock succeeded")
	} else if !isStorageLockBusy(err) {
		t.Fatalf("lock conflict not recognized: %v", err)
	}
	lock.Close()
	next, err := acquireStorageUse(dir)
	if err != nil {
		t.Fatal(err)
	}
	next.Close()
}
func TestDesktopStorageStartupMigrationAndFallback(t *testing.T) {
	previous, managed, configFile := customDataDir, desktopStorageManaged, storageConfigFile
	portable, server := isPortableMode, isServerMode
	t.Cleanup(func() {
		customDataDir, desktopStorageManaged, storageConfigFile = previous, managed, configFile
		isPortableMode, isServerMode = portable, server
	})
	IsPortableMode()
	isPortableMode = false
	isServerMode = false
	t.Setenv("MRRSS_DATA_DIR", "")
	t.Setenv("APPDATA", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	source, target := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(source, "rss.db"), []byte("preserved"), 0600)
	path, err := storageConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if err := writeStorageConfig(path, StorageConfig{DataDirectory: source, PendingDirectory: target}); err != nil {
		t.Fatal(err)
	}
	lock, err := InitializeDesktopStorage("")
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	if customDataDir != target {
		t.Fatalf("directory=%s", customDataDir)
	}
	cfg, err := readStorageConfig(path)
	if err != nil || cfg.PendingDirectory != "" || cfg.DataDirectory != target {
		t.Fatalf("configuration=%+v %v", cfg, err)
	}
	lock.Close()
	occupied := t.TempDir()
	os.WriteFile(filepath.Join(occupied, "user.txt"), []byte("keep"), 0600)
	writeStorageConfig(path, StorageConfig{DataDirectory: target, PendingDirectory: occupied})
	next, err := InitializeDesktopStorage("")
	if err != nil {
		t.Fatal(err)
	}
	defer next.Close()
	cfg, err = DesktopStorageStatus()
	if err != nil || cfg.LastError == "" || customDataDir != target {
		t.Fatalf("missing fallback: %+v %v", cfg, err)
	}
}

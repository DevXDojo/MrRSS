package database

import (
	"database/sql"
	"fmt"
)

// Reader providers share the Google Reader protocol, but never credentials or data ownership.
func ValidReaderProvider(provider string) bool {
	return provider == "freshrss" || provider == "miniflux"
}

func migrateReaderProviders(db *sql.DB) error {
	rows, err := db.Query("PRAGMA table_info(feeds)")
	if err != nil {
		return err
	}
	exists := false
	for rows.Next() {
		var cid, notnull, pk int
		var name, kind string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &kind, &notnull, &defaultValue, &pk); err != nil {
			rows.Close()
			return err
		}
		if name == "sync_provider" {
			exists = true
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if !exists {
		if _, err = tx.Exec("ALTER TABLE feeds ADD COLUMN sync_provider TEXT NOT NULL DEFAULT 'freshrss'"); err != nil {
			return err
		}
	}
	var done, legacy string
	if err = tx.QueryRow("SELECT value FROM settings WHERE key = 'reader_providers_migrated'").Scan(&done); err != nil && err != sql.ErrNoRows {
		return err
	}
	if done == "1" {
		return tx.Commit()
	}
	if err = tx.QueryRow("SELECT value FROM settings WHERE key = 'freshrss_provider'").Scan(&legacy); err != nil && err != sql.ErrNoRows {
		return err
	}
	if legacy == "miniflux" {
		for _, suffix := range []string{"enabled", "server_url", "username", "api_password", "auto_sync_interval", "sync_on_startup", "last_sync_time"} {
			// Copy encrypted values verbatim: migration must not decrypt or log credentials.
			if _, err = tx.Exec("INSERT OR REPLACE INTO settings (key, value) SELECT ?, value FROM settings WHERE key = ?", "miniflux_"+suffix, "freshrss_"+suffix); err != nil {
				return err
			}
		}
		if _, err = tx.Exec("UPDATE feeds SET sync_provider = 'miniflux' WHERE is_freshrss_source = 1"); err != nil {
			return err
		}
		if _, err = tx.Exec("UPDATE settings SET value = '' WHERE key IN ('freshrss_server_url', 'freshrss_username', 'freshrss_api_password', 'freshrss_last_sync_time')"); err != nil {
			return err
		}
		if _, err = tx.Exec("UPDATE settings SET value = 'false' WHERE key IN ('freshrss_enabled', 'freshrss_sync_on_startup')"); err != nil {
			return err
		}
		if _, err = tx.Exec("UPDATE settings SET value = '0' WHERE key = 'freshrss_auto_sync_interval'"); err != nil {
			return err
		}
	}
	if _, err = tx.Exec("INSERT OR REPLACE INTO settings (key,value) VALUES ('freshrss_provider','freshrss'), ('reader_providers_migrated','1')"); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) GetReaderConfig(provider string) (serverURL, username, password string, err error) {
	if !ValidReaderProvider(provider) {
		return "", "", "", fmt.Errorf("invalid reader provider")
	}
	serverURL, err = db.GetSetting(provider + "_server_url")
	if err != nil {
		return
	}
	username, err = db.GetSetting(provider + "_username")
	if err != nil {
		return
	}
	password, err = db.GetEncryptedSetting(provider + "_api_password")
	return
}

func (db *DB) FeedSyncProvider(feedID int64) string {
	var provider string
	if err := db.QueryRow("SELECT sync_provider FROM feeds WHERE id = ? AND is_freshrss_source = 1", feedID).Scan(&provider); err != nil || !ValidReaderProvider(provider) {
		return ""
	}
	return provider
}

func (db *DB) FeedSyncEnabled(feedID int64) bool {
	provider := db.FeedSyncProvider(feedID)
	if provider == "" {
		return false
	}
	enabled, _ := db.GetSetting(provider + "_enabled")
	return enabled == "true"
}

func (db *DB) GetArticleSyncConfig(articleID int64) (serverURL, username, password, provider string, err error) {
	var feedID int64
	if err = db.QueryRow("SELECT feed_id FROM articles WHERE id = ?", articleID).Scan(&feedID); err != nil {
		return
	}
	provider = db.FeedSyncProvider(feedID)
	if !db.FeedSyncEnabled(feedID) {
		err = fmt.Errorf("reader sync is disabled")
		return
	}
	serverURL, username, password, err = db.GetReaderConfig(provider)
	return
}

func (db *DB) ReaderSyncEnabled() bool {
	for _, provider := range []string{"freshrss", "miniflux"} {
		enabled, _ := db.GetSetting(provider + "_enabled")
		if enabled == "true" {
			return true
		}
	}
	return false
}

// CleanupReaderData only removes the selected provider's data, atomically.
func (db *DB) CleanupReaderData(provider string) error {
	if !ValidReaderProvider(provider) {
		return fmt.Errorf("invalid reader provider")
	}
	db.WaitForReady()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, query := range []string{
		"DELETE FROM freshrss_sync_queue WHERE article_id IN (SELECT id FROM articles WHERE feed_id IN (SELECT id FROM feeds WHERE is_freshrss_source = 1 AND sync_provider = ?))",
		"DELETE FROM article_contents WHERE article_id IN (SELECT id FROM articles WHERE feed_id IN (SELECT id FROM feeds WHERE is_freshrss_source = 1 AND sync_provider = ?))",
		"DELETE FROM articles WHERE feed_id IN (SELECT id FROM feeds WHERE is_freshrss_source = 1 AND sync_provider = ?)",
		"DELETE FROM feeds WHERE is_freshrss_source = 1 AND sync_provider = ?",
	} {
		if _, err = tx.Exec(query, provider); err != nil {
			return err
		}
	}
	return tx.Commit()
}

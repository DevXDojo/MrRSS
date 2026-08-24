package database

import (
	"MrRSS/internal/config"
	"log"

	_ "modernc.org/sqlite"
)

// Init initializes the database schema and settings.
// This method must be called before any database operations.
func (db *DB) Init() error {
	var err error
	db.once.Do(func() {
		defer close(db.ready)

		if err = db.Ping(); err != nil {
			return
		}

		if err = initSchema(db.DB); err != nil {
			return
		}

		// Initialize FreshRSS sync queue table
		if err = InitFreshRSSSyncTable(db.DB); err != nil {
			return
		}

		// Initialize statistics table
		if err = InitStatisticsTable(db.DB); err != nil {
			return
		}

		// Create settings table if not exists
		_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT
		)`)

		// Insert default settings if they don't exist (using centralized defaults from config)
		settingsKeys := config.SettingsKeys()
		for _, key := range settingsKeys {
			defaultVal := config.GetString(key)
			// Use parameterized query to prevent SQL injection
			_, _ = db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES (?, ?)`, key, defaultVal)
		}

		// Apply additional migrations
		if err = applyAdditionalMigrations(db); err != nil {
			return
		}

		// Migration: enable auto_vacuum in INCREMENTAL mode so that
		// IncrementalVacuum() can reclaim freelist pages after deletions
		// without requiring a full VACUUM (which locks the database).
		// For existing databases created with auto_vacuum=NONE (value 0),
		// we must run a full VACUUM once to convert to INCREMENTAL mode.
		// This is guarded by a settings flag so it only runs once.
		if migrationErr := migrateAutoVacuumIncremental(db); migrationErr != nil {
			log.Printf("Warning: auto_vacuum migration failed: %v", migrationErr)
			// Don't return error — the app can still work without incremental vacuum
		}
	})
	return err
}

// migrateAutoVacuumIncremental switches the database to auto_vacuum=INCREMENTAL
// mode. For new databases this is a no-op if already set. For existing databases
// with auto_vacuum=NONE, a one-time VACUUM is required to convert the mode.
// We use a settings flag to ensure the VACUUM only runs once.
//
// IMPORTANT: This function runs inside Init() before db.ready is closed.
// It must NOT call methods that invoke WaitForReady (like GetSetting),
// as that would deadlock. Use the underlying sql.DB methods directly.
func migrateAutoVacuumIncremental(db *DB) error {
	// Check if already migrated — use sql.DB directly to avoid WaitForReady deadlock
	var done string
	err := db.DB.QueryRow("SELECT value FROM settings WHERE key = 'auto_vacuum_migrated'").Scan(&done)
	if err == nil && done == "1" {
		return nil
	}
	// If error is sql.ErrNoRows, the setting doesn't exist yet — proceed with migration

	// Check current auto_vacuum mode
	var autoVacuum int64
	err = db.DB.QueryRow("PRAGMA auto_vacuum").Scan(&autoVacuum)
	if err != nil {
		return err
	}

	// auto_vacuum values: 0=NONE, 1=FULL, 2=INCREMENTAL
	if autoVacuum == 2 {
		// Already in incremental mode, just mark as done
		_, _ = db.DB.Exec(`INSERT OR REPLACE INTO settings (key, value) VALUES ('auto_vacuum_migrated', '1')`)
		return nil
	}

	log.Println("Migrating database to auto_vacuum=INCREMENTAL mode (one-time VACUUM required)...")

	// VACUUM requires exclusive access to the database. With a connection pool
	// of 25, other idle connections can hold locks that prevent VACUUM from
	// completing, causing deadlocks. Temporarily restrict to a single connection.
	db.SetMaxOpenConns(1)
	defer db.SetMaxOpenConns(25)

	// Set to INCREMENTAL mode
	if _, err := db.DB.Exec("PRAGMA auto_vacuum = INCREMENTAL"); err != nil {
		return err
	}

	// VACUUM to apply the new auto_vacuum setting to the entire database.
	// This also reclaims all freelist pages immediately.
	if _, err := db.DB.Exec("VACUUM"); err != nil {
		return err
	}

	// Mark migration as done
	_, _ = db.DB.Exec(`INSERT OR REPLACE INTO settings (key, value) VALUES ('auto_vacuum_migrated', '1')`)
	log.Println("auto_vacuum=INCREMENTAL migration completed")

	return nil
}

// applyAdditionalMigrations applies migrations that need to run after schema initialization
func applyAdditionalMigrations(db *DB) error {
	// Migration: Add link column to feeds table if it doesn't exist
	// Note: SQLite doesn't support IF NOT EXISTS for ALTER TABLE ADD COLUMN.
	// Error is ignored - if column exists, the operation fails harmlessly.
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN link TEXT DEFAULT ''`)

	// Migration: Add discovery_completed column to feeds table
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN discovery_completed BOOLEAN DEFAULT 0`)

	// Migration: Add script_path column to feeds table for custom script support
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN script_path TEXT DEFAULT ''`)

	// Migration: Add hide_from_timeline column to feeds table
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN hide_from_timeline BOOLEAN DEFAULT 0`)

	// Migration: Add proxy and refresh interval columns to feeds table
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN proxy_url TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN proxy_enabled BOOLEAN DEFAULT 0`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN refresh_interval INTEGER DEFAULT 0`)

	// Migration: Add is_image_mode column to feeds table for image gallery feature
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN is_image_mode BOOLEAN DEFAULT 0`)

	// Migration: Add position column to feeds table for custom ordering
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN position INTEGER DEFAULT 0`)

	// Migration: Add article_view_mode column to feeds table for per-feed view mode override
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN article_view_mode TEXT DEFAULT 'global'`)

	// Migration: Add auto_expand_content column to feeds table for per-feed content expansion override
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN auto_expand_content TEXT DEFAULT 'global'`)

	// Migration: Add is_freshrss_source column to feeds table to mark feeds from FreshRSS
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN is_freshrss_source BOOLEAN DEFAULT 0`)

	// Migration: Add freshrss_stream_id column to feeds table to store FreshRSS stream ID
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN freshrss_stream_id TEXT DEFAULT ''`)

	// Migration: Add summary column to articles table for AI-generated summaries
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN summary TEXT DEFAULT ''`)

	// Migration: Add original_summary column for RSS-provided summaries/descriptions
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN original_summary TEXT DEFAULT ''`)

	// Run complex table migrations
	if err := migrateUniqueIDOnArticles(db.DB); err != nil {
		return err
	}

	if err := migrateDropUniqueConstraintOnArticles(db.DB); err != nil {
		return err
	}

	if err := migrateDropUniqueConstraintOnFeeds(db.DB); err != nil {
		return err
	}

	return nil
}

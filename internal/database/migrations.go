package database

import (
	"database/sql"
	"log"
	"strings"
)

// runMigrations applies database migrations for existing databases.
// This ensures all columns and tables exist as the schema evolves.
func runMigrations(db *sql.DB) error {
	// Migration: Add content and is_hidden columns if they don't exist
	// SQLite doesn't support IF NOT EXISTS for ALTER TABLE, so we ignore errors if columns already exist
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN content TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN is_hidden BOOLEAN DEFAULT 0`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN last_error TEXT DEFAULT ''`)

	// Migration: Add is_read_later column for read later feature
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN is_read_later BOOLEAN DEFAULT 0`)

	// Migration: Add audio_url column for podcast support
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN audio_url TEXT DEFAULT ''`)

	// Migration: Add video_url column for YouTube video support
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN video_url TEXT DEFAULT ''`)

	// Migration: Add XPath support fields to feeds table
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN type TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_title TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_content TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_uri TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_author TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_timestamp TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_time_format TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_thumbnail TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_categories TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN xpath_item_uid TEXT DEFAULT ''`)

	// Migration: Add summary column for caching AI-generated summaries
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN summary TEXT DEFAULT ''`)

	// Migration: Add original_summary column for RSS-provided summaries/descriptions
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN original_summary TEXT DEFAULT ''`)

	// Migration: Add article_contents table for caching article content
	// This uses a separate table to keep the articles table lightweight
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS article_contents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_id INTEGER NOT NULL UNIQUE,
		content TEXT NOT NULL,
		fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
	)`)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_article_contents_article_id ON article_contents(article_id)`)

	// Migration: Add chat_sessions and chat_messages tables for AI chat feature
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS chat_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
	)`)
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS chat_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		thinking TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(session_id) REFERENCES chat_sessions(id) ON DELETE CASCADE
	)`)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_chat_sessions_article_id ON chat_sessions(article_id)`)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_chat_sessions_updated_at ON chat_sessions(updated_at DESC)`)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_chat_messages_session_id ON chat_messages(session_id)`)

	// Migration: Add newsletter/email support fields to feeds table
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN email_address TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN email_imap_server TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN email_imap_port INTEGER DEFAULT 993`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN email_username TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN email_password TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN email_folder TEXT DEFAULT 'INBOX'`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN email_last_uid INTEGER DEFAULT 0`)

	// Migration: Add FreshRSS integration fields
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN is_freshrss_source BOOLEAN DEFAULT 0`)
	_, _ = db.Exec(`ALTER TABLE feeds ADD COLUMN freshrss_stream_id TEXT DEFAULT ''`)
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN freshrss_item_id TEXT DEFAULT ''`)

	// Migration: Add author field to articles table
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN author TEXT DEFAULT ''`)

	// Migration: Add saved_filters table for custom filter persistence
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS saved_filters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		conditions TEXT NOT NULL,
		position INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_saved_filters_position ON saved_filters(position)`)

	// Migration: Add tags table for feed tagging feature
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		color TEXT NOT NULL DEFAULT '#3B82F6',
		position INTEGER DEFAULT 0
	)`)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_tags_position ON tags(position)`)

	// Migration: Add feed_tags junction table for many-to-many relationship
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS feed_tags (
		feed_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		PRIMARY KEY (feed_id, tag_id),
		FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
	)`)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_feed_tags_feed_id ON feed_tags(feed_id)`)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_feed_tags_tag_id ON feed_tags(tag_id)`)

	// Migration: Add ai_profiles table for multiple AI configuration support
	_, _ = db.Exec(`CREATE TABLE IF NOT EXISTS ai_profiles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		api_key TEXT DEFAULT '',
		endpoint TEXT NOT NULL,
		model TEXT NOT NULL,
		custom_headers TEXT DEFAULT '',
		is_default BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_ai_profiles_is_default ON ai_profiles(is_default)`)

	return nil
}

// migrateUniqueIDOnArticles adds unique_id column and generates values for existing articles.
// This replaces URL-based deduplication with title+feed_id+published_date based deduplication.
func migrateUniqueIDOnArticles(db *sql.DB) error {
	// Migration: Add unique_id column to articles table for better deduplication
	_, _ = db.Exec(`ALTER TABLE articles ADD COLUMN unique_id TEXT UNIQUE`)

	// Migration: Migrate existing articles to generate unique_id
	// For existing articles, generate unique_id from title+feed_id+published_date (date only, not full timestamp)
	// If url was UNIQUE before, keep it but unique_id is now the primary deduplication key
	_, err := db.Exec(`
		UPDATE articles
		SET unique_id = LOWER(HEX(MD5(title || '|' || feed_id || '|' || COALESCE(strftime('%Y-%m-%d', published_at), ''))))
		WHERE unique_id IS NULL
	`)
	if err != nil {
		log.Printf("Warning: Failed to migrate unique_id: %v", err)
	}

	// Backfill published_at for articles that have NULL values
	// Set to current time as fallback (article creation time is unknown)
	result, err := db.Exec(`
		UPDATE articles
		SET published_at = datetime('now')
		WHERE published_at IS NULL
	`)
	if err != nil {
		log.Printf("Warning: Failed to backfill published_at: %v", err)
	} else {
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			log.Printf("Backfilled published_at for %d articles", rowsAffected)
		}
	}

	return nil
}

// migrateDropUniqueConstraintOnArticles drops the UNIQUE constraint on url column from articles table.
// This allows multiple articles with the same URL (e.g., from different feeds).
func migrateDropUniqueConstraintOnArticles(db *sql.DB) error {
	// Migration: Drop the UNIQUE constraint on url column if it exists
	// SQLite doesn't support DROP CONSTRAINT directly, so we need to recreate the table
	// Check if we need to migrate by checking if url is still UNIQUE
	var tableInfo string
	_ = db.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name='articles'").Scan(&tableInfo)
	if strings.Contains(tableInfo, "url TEXT UNIQUE") {
		// Need to recreate table without UNIQUE constraint on url
		_, err := db.Exec(`
			CREATE TABLE articles_new (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				feed_id INTEGER,
				title TEXT,
				url TEXT,
				image_url TEXT,
				audio_url TEXT DEFAULT '',
				video_url TEXT DEFAULT '',
				translated_title TEXT,
				published_at DATETIME,
				is_read BOOLEAN DEFAULT 0,
				is_favorite BOOLEAN DEFAULT 0,
				is_hidden BOOLEAN DEFAULT 0,
				is_read_later BOOLEAN DEFAULT 0,
				summary TEXT DEFAULT '',
				original_summary TEXT DEFAULT '',
				unique_id TEXT UNIQUE,
				FOREIGN KEY(feed_id) REFERENCES feeds(id)
			)
		`)
		if err == nil {
			// Copy data from old table to new table
			_, _ = db.Exec(`
				INSERT INTO articles_new (id, feed_id, title, url, image_url, audio_url, video_url, translated_title, published_at, is_read, is_favorite, is_hidden, is_read_later, summary, original_summary, unique_id)
				SELECT id, feed_id, title, url, image_url, audio_url, video_url, translated_title, published_at, is_read, is_favorite, is_hidden, is_read_later,
					COALESCE(summary, '') as summary,
					COALESCE(original_summary, '') as original_summary,
					LOWER(HEX(MD5(title || '|' || feed_id || '|' || COALESCE(strftime('%Y-%m-%d', published_at), '')))) as unique_id
				FROM articles
			`)
			// Drop old table and rename new table
			_, _ = db.Exec(`DROP TABLE articles`)
			_, _ = db.Exec(`ALTER TABLE articles_new RENAME TO articles`)
			// Recreate indexes
			_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_articles_feed_id ON articles(feed_id)`)
			_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at DESC)`)
			_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_articles_is_read ON articles(is_read)`)
			_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_articles_is_favorite ON articles(is_favorite)`)
			_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_articles_is_hidden ON articles(is_hidden)`)
			_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_articles_is_read_later ON articles(is_read_later)`)
			_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_articles_feed_published ON articles(feed_id, published_at DESC)`)
			_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_articles_read_published ON articles(is_read, published_at DESC)`)
			_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_articles_fav_published ON articles(is_favorite, published_at DESC)`)
			_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_articles_readlater_published ON articles(is_read_later, published_at DESC)`)
			_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_articles_hidden_published ON articles(is_hidden, published_at DESC)`)
		}
	}
	return nil
}

// migrateDropUniqueConstraintOnFeeds drops the UNIQUE constraint on url column from feeds table.
// This allows FreshRSS and local feeds with the same URL to coexist.
func migrateDropUniqueConstraintOnFeeds(db *sql.DB) error {
	// Migration: Drop the UNIQUE constraint on feeds.url column to allow FreshRSS and local feeds with same URL
	var feedsTableInfo string
	_ = db.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name='feeds'").Scan(&feedsTableInfo)
	if strings.Contains(feedsTableInfo, "url TEXT UNIQUE") {
		log.Printf("Migration: Dropping UNIQUE constraint on feeds.url to allow FreshRSS and local feeds to coexist")
		_, err := db.Exec(`
			CREATE TABLE feeds_new (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				title TEXT,
				url TEXT,
				link TEXT DEFAULT '',
				description TEXT,
				category TEXT DEFAULT '',
				image_url TEXT DEFAULT '',
				position INTEGER DEFAULT 0,
				last_updated DATETIME,
				last_error TEXT DEFAULT '',
				discovery_completed BOOLEAN DEFAULT 0,
				script_path TEXT DEFAULT '',
				hide_from_timeline BOOLEAN DEFAULT 0,
				proxy_url TEXT DEFAULT '',
				proxy_enabled BOOLEAN DEFAULT 0,
				refresh_interval INTEGER DEFAULT 0,
				is_image_mode BOOLEAN DEFAULT 0,
				type TEXT DEFAULT '',
				xpath_item TEXT DEFAULT '',
				xpath_item_title TEXT DEFAULT '',
				xpath_item_content TEXT DEFAULT '',
				xpath_item_uri TEXT DEFAULT '',
				xpath_item_author TEXT DEFAULT '',
				xpath_item_timestamp TEXT DEFAULT '',
				xpath_item_time_format TEXT DEFAULT '',
				xpath_item_thumbnail TEXT DEFAULT '',
				xpath_item_categories TEXT DEFAULT '',
				xpath_item_uid TEXT DEFAULT '',
				article_view_mode TEXT DEFAULT '',
				auto_expand_content TEXT DEFAULT '',
				email_address TEXT DEFAULT '',
				email_imap_server TEXT DEFAULT '',
				email_imap_port INTEGER DEFAULT 993,
				email_username TEXT DEFAULT '',
				email_password TEXT DEFAULT '',
				email_folder TEXT DEFAULT 'INBOX',
				email_last_uid INTEGER DEFAULT 0,
				is_freshrss_source BOOLEAN DEFAULT 0,
				freshrss_stream_id TEXT DEFAULT ''
			)
		`)
		if err == nil {
			// Copy data from old table to new table
			_, err = db.Exec(`
				INSERT INTO feeds_new (
					id, title, url, link, description, category, image_url, position, last_updated, last_error,
					discovery_completed, script_path, hide_from_timeline, proxy_url, proxy_enabled, refresh_interval,
					is_image_mode, type, xpath_item, xpath_item_title, xpath_item_content, xpath_item_uri,
					xpath_item_author, xpath_item_timestamp, xpath_item_time_format, xpath_item_thumbnail,
					xpath_item_categories, xpath_item_uid, article_view_mode, auto_expand_content,
					email_address, email_imap_server, email_imap_port, email_username, email_password,
					email_folder, email_last_uid, is_freshrss_source, freshrss_stream_id
				)
				SELECT
					id, title, url, link, description, category, image_url,
					COALESCE(position, 0) as position,
					last_updated, COALESCE(last_error, '') as last_error,
					COALESCE(discovery_completed, 0) as discovery_completed,
					COALESCE(script_path, '') as script_path,
					COALESCE(hide_from_timeline, 0) as hide_from_timeline,
					COALESCE(proxy_url, '') as proxy_url,
					COALESCE(proxy_enabled, 0) as proxy_enabled,
					COALESCE(refresh_interval, 0) as refresh_interval,
					COALESCE(is_image_mode, 0) as is_image_mode,
					COALESCE(type, '') as type,
					COALESCE(xpath_item, '') as xpath_item,
					COALESCE(xpath_item_title, '') as xpath_item_title,
					COALESCE(xpath_item_content, '') as xpath_item_content,
					COALESCE(xpath_item_uri, '') as xpath_item_uri,
					COALESCE(xpath_item_author, '') as xpath_item_author,
					COALESCE(xpath_item_timestamp, '') as xpath_item_timestamp,
					COALESCE(xpath_item_time_format, '') as xpath_item_time_format,
					COALESCE(xpath_item_thumbnail, '') as xpath_item_thumbnail,
					COALESCE(xpath_item_categories, '') as xpath_item_categories,
					COALESCE(xpath_item_uid, '') as xpath_item_uid,
					COALESCE(article_view_mode, '') as article_view_mode,
					COALESCE(auto_expand_content, '') as auto_expand_content,
					COALESCE(email_address, '') as email_address,
					COALESCE(email_imap_server, '') as email_imap_server,
					COALESCE(email_imap_port, 993) as email_imap_port,
					COALESCE(email_username, '') as email_username,
					COALESCE(email_password, '') as email_password,
					COALESCE(email_folder, 'INBOX') as email_folder,
					COALESCE(email_last_uid, 0) as email_last_uid,
					COALESCE(is_freshrss_source, 0) as is_freshrss_source,
					COALESCE(freshrss_stream_id, '') as freshrss_stream_id
				FROM feeds
			`)
			if err != nil {
				log.Printf("Error copying feeds data: %v", err)
			}
			// Drop old table and rename new table
			_, _ = db.Exec(`DROP TABLE feeds`)
			_, _ = db.Exec(`ALTER TABLE feeds_new RENAME TO feeds`)
			// Recreate indexes
			_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_feeds_category ON feeds(category)`)
			log.Printf("Migration completed: UNIQUE constraint dropped from feeds.url")
		} else {
			log.Printf("Error creating feeds_new table: %v", err)
		}
	}
	return nil
}

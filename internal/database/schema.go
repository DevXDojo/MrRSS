package database

import (
	"database/sql"
)

// initSchema initializes the database schema by creating all tables and indexes.
// This is extracted from db.go for better code organization.
func initSchema(db *sql.DB) error {
	// First create tables
	query := `
	CREATE TABLE IF NOT EXISTS feeds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		url TEXT UNIQUE,
		link TEXT DEFAULT '',
		description TEXT,
		category TEXT DEFAULT '',
		image_url TEXT DEFAULT '',
		last_updated DATETIME,
		last_error TEXT DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS articles (
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
	);

	-- Translation cache table to avoid redundant API calls
	CREATE TABLE IF NOT EXISTS translation_cache (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_text_hash TEXT NOT NULL,
		source_text TEXT NOT NULL,
		target_lang TEXT NOT NULL,
		translated_text TEXT NOT NULL,
		provider TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(source_text_hash, target_lang, provider)
	);

	-- Article content cache table to store full article content
	CREATE TABLE IF NOT EXISTS article_contents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_id INTEGER NOT NULL UNIQUE,
		content TEXT NOT NULL,
		fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
	);

	-- Chat sessions table to store AI chat conversations per article
	CREATE TABLE IF NOT EXISTS chat_sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		article_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(article_id) REFERENCES articles(id) ON DELETE CASCADE
	);

	-- Chat messages table to store individual messages in chat sessions
	CREATE TABLE IF NOT EXISTS chat_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		thinking TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(session_id) REFERENCES chat_sessions(id) ON DELETE CASCADE
	);

	-- Create indexes for better query performance
	CREATE INDEX IF NOT EXISTS idx_articles_feed_id ON articles(feed_id);
	CREATE INDEX IF NOT EXISTS idx_articles_published_at ON articles(published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_articles_is_read ON articles(is_read);
	CREATE INDEX IF NOT EXISTS idx_articles_is_favorite ON articles(is_favorite);
	CREATE INDEX IF NOT EXISTS idx_articles_is_hidden ON articles(is_hidden);
	CREATE INDEX IF NOT EXISTS idx_articles_is_read_later ON articles(is_read_later);
	CREATE INDEX IF NOT EXISTS idx_feeds_category ON feeds(category);

	-- Composite indexes for common query patterns
	CREATE INDEX IF NOT EXISTS idx_articles_feed_published ON articles(feed_id, published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_articles_read_published ON articles(is_read, published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_articles_fav_published ON articles(is_favorite, published_at DESC);
	CREATE INDEX IF NOT EXISTS idx_articles_readlater_published ON articles(is_read_later, published_at DESC);

	-- Covering index for category queries (hidden + published_at)
	-- Optimizes queries with: WHERE is_hidden = 0 ORDER BY published_at DESC
	CREATE INDEX IF NOT EXISTS idx_articles_hidden_published ON articles(is_hidden, published_at DESC);

	-- Translation cache index
	CREATE INDEX IF NOT EXISTS idx_translation_cache_lookup ON translation_cache(source_text_hash, target_lang, provider);

	-- Article content cache index
	CREATE INDEX IF NOT EXISTS idx_article_contents_article_id ON article_contents(article_id);

	-- Chat sessions and messages indexes
	CREATE INDEX IF NOT EXISTS idx_chat_sessions_article_id ON chat_sessions(article_id);
	CREATE INDEX IF NOT EXISTS idx_chat_sessions_updated_at ON chat_sessions(updated_at DESC);
	CREATE INDEX IF NOT EXISTS idx_chat_messages_session_id ON chat_messages(session_id);
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	// Then run migrations to ensure all columns exist
	// This must happen AFTER creating tables
	if err := runMigrations(db); err != nil {
		return err
	}

	return nil
}

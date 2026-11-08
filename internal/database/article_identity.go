package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	"MrRSS/internal/models"
	"MrRSS/internal/utils/urlutil"
)

func articleUniqueID(article *models.Article) string {
	identity := ""
	if article.FreshRSSItemID != "" {
		identity = "item:" + article.FreshRSSItemID
	} else if raw := strings.TrimSpace(article.URL); raw != "" {
		// Remove only known tracking parameters. Other parameters can identify
		// distinct posts (for example ?p=123 or ?id=456).
		if parsed, err := url.Parse(raw); err == nil {
			parsed.Fragment = ""
			parsed.Host = strings.ToLower(parsed.Host)
			query := parsed.Query()
			for key := range query {
				lower := strings.ToLower(key)
				if strings.HasPrefix(lower, "utm_") || lower == "fbclid" || lower == "gclid" {
					query.Del(key)
				}
			}
			parsed.RawQuery = query.Encode()
			raw = parsed.String()
		}
		identity = "url:" + raw
	}
	return urlutil.GenerateArticleUniqueID(article.Title, article.FeedID, article.PublishedAt, article.HasValidPublishedTime, identity)
}

func (db *DB) GetSavedArticleID(article *models.Article) (int64, error) {
	db.WaitForReady()
	var id int64
	err := db.QueryRow("SELECT id FROM articles WHERE unique_id = ?", articleUniqueID(article)).Scan(&id)
	return id, err
}

// Backfill in Go rather than relying on an unavailable SQLite MD5 function.
// Never delete historical duplicates: all article IDs and dependent rows survive.
func migrateArticleIdentity(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var done string
	err = tx.QueryRow("SELECT value FROM settings WHERE key = 'article_identity_v2'").Scan(&done)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if done == "1" {
		return tx.Commit()
	}
	rows, err := tx.Query(`SELECT id, feed_id, COALESCE(title, ''), COALESCE(url, ''), COALESCE(freshrss_item_id, '') FROM articles ORDER BY id`)
	if err != nil {
		return err
	}
	var articles []models.Article
	for rows.Next() {
		var article models.Article
		if err := rows.Scan(&article.ID, &article.FeedID, &article.Title, &article.URL, &article.FreshRSSItemID); err != nil {
			rows.Close()
			return err
		}
		articles = append(articles, article)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	// Clear old keys inside the transaction to avoid conflicts with a new key
	// that happens to equal a different historical row's key.
	if _, err := tx.Exec("UPDATE articles SET unique_id = NULL"); err != nil {
		return err
	}
	seen := make(map[string]bool)
	for _, article := range articles {
		key := articleUniqueID(&article)
		if seen[key] {
			key += fmt.Sprintf(":legacy:%d", article.ID)
		}
		seen[key] = true
		if _, err := tx.Exec("UPDATE articles SET unique_id = ? WHERE id = ?", key, article.ID); err != nil {
			return err
		}
	}
	if _, err := tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_articles_unique_id ON articles(unique_id)"); err != nil {
		return err
	}
	if _, err := tx.Exec("INSERT OR REPLACE INTO settings (key, value) VALUES ('article_identity_v2', '1')"); err != nil {
		return err
	}
	return tx.Commit()
}

// Preserve every column, index, article ID and dependent row when upgrading
// very old databases whose article URLs still have a UNIQUE constraint.
func dropLegacyArticleURLConstraint(db *sql.DB) error {
	ctx := context.Background()
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	var schema string
	if err := conn.QueryRowContext(ctx, "SELECT sql FROM sqlite_master WHERE type='table' AND name='articles'").Scan(&schema); err != nil {
		return err
	}
	if !strings.Contains(schema, "url TEXT UNIQUE") {
		return nil
	}
	rows, err := conn.QueryContext(ctx, "SELECT sql FROM sqlite_master WHERE type='index' AND tbl_name='articles' AND sql IS NOT NULL")
	if err != nil {
		return err
	}
	var indexes []string
	for rows.Next() {
		var statement string
		if err := rows.Scan(&statement); err != nil {
			rows.Close()
			return err
		}
		indexes = append(indexes, statement)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, "PRAGMA foreign_keys=OFF"); err != nil {
		return err
	}
	defer conn.ExecContext(ctx, "PRAGMA foreign_keys=ON")
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	schema = strings.Replace(schema, "url TEXT UNIQUE", "url TEXT", 1)
	// SQLite stores the original CREATE TABLE spelling. Replace just its
	// table declaration, leaving the column and foreign key definitions intact.
	start := strings.Index(schema, "(")
	if start < 0 {
		return fmt.Errorf("invalid legacy articles schema")
	}
	if _, err := tx.Exec("CREATE TABLE articles_new " + schema[start:]); err != nil {
		return err
	}
	for _, statement := range []string{
		"INSERT INTO articles_new SELECT * FROM articles",
		"DROP TABLE articles",
		"ALTER TABLE articles_new RENAME TO articles",
	} {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}
	for _, statement := range indexes {
		if _, err := tx.Exec(statement); err != nil {
			return err
		}
	}
	return tx.Commit()
}

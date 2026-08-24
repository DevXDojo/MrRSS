package database

import (
	"database/sql"
	"fmt"
)

// TranslationCache represents a cached translation entry
type TranslationCache struct {
	ID             int64
	SourceTextHash string
	SourceText     string
	TargetLang     string
	TranslatedText string
	Provider       string
	CreatedAt      string
}

// GetCachedTranslation retrieves a translation from cache if available
func (db *DB) GetCachedTranslation(sourceTextHash, targetLang, provider string) (string, bool, error) {
	var translatedText string
	err := db.QueryRow(
		`SELECT translated_text FROM translation_cache
		 WHERE source_text_hash = ? AND target_lang = ? AND provider = ?`,
		sourceTextHash, targetLang, provider,
	).Scan(&translatedText)

	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return translatedText, true, nil
}

// SetCachedTranslation stores a translation in cache
func (db *DB) SetCachedTranslation(sourceTextHash, sourceText, targetLang, translatedText, provider string) error {
	_, err := db.Exec(
		`INSERT OR REPLACE INTO translation_cache
			(source_text_hash, source_text, target_lang, translated_text, provider, created_at)
			VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		sourceTextHash, sourceText, targetLang, translatedText, provider,
	)
	return err
}

// CleanupTranslationCache removes cached translations older than maxAgeDays
func (db *DB) CleanupTranslationCache(maxAgeDays int) (int64, error) {
	result, err := db.Exec(
		`DELETE FROM translation_cache WHERE created_at < datetime('now', ?)`,
		fmt.Sprintf("-%d days", maxAgeDays),
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

package database

// UpdateArticleContent updates the content field for an article in the articles table.
func (db *DB) UpdateArticleContent(id int64, content string) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET content = ? WHERE id = ?", content, id)
	return err
}

// UpdateArticleTranslation updates the translated_title field for an article.
func (db *DB) UpdateArticleTranslation(id int64, translatedTitle string) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET translated_title = ? WHERE id = ?", translatedTitle, id)
	return err
}

// UpdateArticleSummary updates the cached summary for an article.
func (db *DB) UpdateArticleSummary(id int64, summary string) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET summary = ? WHERE id = ?", summary, id)
	return err
}

// GetArticleOriginalSummary returns the RSS-provided summary/description for an article.
func (db *DB) GetArticleOriginalSummary(id int64) (string, error) {
	db.WaitForReady()

	var summary string
	err := db.QueryRow("SELECT COALESCE(original_summary, '') FROM articles WHERE id = ?", id).Scan(&summary)
	return summary, err
}

// ClearAllTranslations clears all translated titles from articles.
func (db *DB) ClearAllTranslations() error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET translated_title = ''")
	return err
}

// ClearAllSummaries clears all summaries from articles.
func (db *DB) ClearAllSummaries() error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE articles SET summary = ''")
	return err
}

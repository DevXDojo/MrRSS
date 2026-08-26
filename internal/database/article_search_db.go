package database

import (
	"database/sql"
	"fmt"
	stdhtml "html"
	"log"
	"sort"
	"strings"
	"time"
	"unicode"

	"MrRSS/internal/models"

	"golang.org/x/net/html"
)

// GetImageGalleryArticles retrieves articles from image mode feeds with pagination.
// If feedID is provided, it gets articles only from that feed (assuming it's an image mode feed).
// If category is provided, it gets articles from all image mode feeds in that category.
// Otherwise, it gets articles from all image mode feeds.
// If onlyUnread is true, only returns unread articles.
func (db *DB) GetImageGalleryArticles(feedID int64, category string, showHidden bool, onlyUnread bool, limit, offset int) ([]models.Article, error) {
	db.WaitForReady()
	baseQuery := `
		SELECT a.id, a.feed_id, a.title, a.url, a.image_url, a.audio_url, a.video_url, a.published_at, a.is_read, a.is_favorite, a.is_hidden, a.is_read_later, a.translated_title, a.summary, f.title, a.author
		FROM articles a
		JOIN feeds f ON a.feed_id = f.id
		WHERE COALESCE(f.is_image_mode, 0) = 1
	`
	var args []interface{}

	// Always filter hidden articles unless showHidden is true
	if !showHidden {
		baseQuery += " AND a.is_hidden = 0"
	}

	// Only get articles with image_url
	baseQuery += " AND a.image_url IS NOT NULL AND a.image_url != ''"

	// Filter for only unread articles if requested
	if onlyUnread {
		baseQuery += " AND a.is_read = 0"
	}

	if feedID > 0 {
		baseQuery += " AND a.feed_id = ?"
		args = append(args, feedID)
	} else if category == "\x00" {
		// Special value "\x00" means explicit uncategorized filtering
		baseQuery += " AND (f.category IS NULL OR f.category = '')"
	} else if category != "" {
		// For categories, use prefix match to support nested categories
		baseQuery += " AND (f.category = ? OR f.category LIKE ?)"
		args = append(args, category, category+"/%")
	}
	// Note: When category is empty string, it means no category filter was provided,
	// so we should not filter by category at all (show all image mode articles from all categories).

	baseQuery += " ORDER BY a.published_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]models.Article, 0)
	for rows.Next() {
		var a models.Article
		var imageURL, audioURL, videoURL, translatedTitle, summary, author sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.FeedID, &a.Title, &a.URL, &imageURL, &audioURL, &videoURL, &publishedAt, &a.IsRead, &a.IsFavorite, &a.IsHidden, &a.IsReadLater, &translatedTitle, &summary, &a.FeedTitle, &author); err != nil {
			log.Println("Error scanning article:", err)
			continue
		}
		a.ImageURL = imageURL.String
		a.AudioURL = audioURL.String
		a.VideoURL = videoURL.String
		if publishedAt.Valid {
			a.PublishedAt = publishedAt.Time
		} else {
			a.PublishedAt = time.Time{}
		}
		a.TranslatedTitle = translatedTitle.String
		a.Summary = summary.String
		a.Author = author.String
		articles = append(articles, a)
	}
	return articles, nil
}

// AISearchQuery contains bounded terms parsed from the AI response. AI output
// is always passed to SQLite as data and never executed as SQL.
type AISearchQuery struct {
	Original string
	Required []string
	Optional []string
	Patterns []string
	Limit    int
}

// AISearchResult adds explainable ranking metadata without changing Article.
type AISearchResult struct {
	Article        models.Article
	RelevanceScore float64
	MatchedTerms   []string
	MatchedFields  []string
	Excerpt        string
}

type aiSearchCandidate struct {
	article     models.Article
	content     string
	summaryText string
}

// SearchArticlesWithTerms performs a parameterized candidate lookup and then
// calculates a deterministic, explainable relevance score locally.
func (db *DB) SearchArticlesWithTerms(search AISearchQuery) ([]AISearchResult, error) {
	db.WaitForReady()
	if search.Limit <= 0 {
		search.Limit = 100
	}
	if search.Limit > 500 {
		search.Limit = 500
	}

	originalTerms := uniqueSearchTerms([]string{search.Original}, 1)
	requiredTerms := uniqueSearchTerms(search.Required, 8)
	plainTerms := uniqueSearchTerms(append(append([]string{}, originalTerms...), requiredTerms...), 12)
	patternTerms := validSearchPatterns(search.Patterns, 4)
	if len(plainTerms) == 0 && len(patternTerms) == 0 {
		return []AISearchResult{}, nil
	}

	conditions := make([]string, 0, len(plainTerms)+len(patternTerms))
	conditionArgs := make([]any, 0, (len(plainTerms)+len(patternTerms))*3)
	appendCondition := func(term string, allowWildcard bool) {
		pattern := searchLikePattern(term, allowWildcard)
		conditions = append(conditions, `(LOWER(a.title) LIKE ? ESCAPE '\' OR LOWER(COALESCE(a.summary, '') || ' ' || COALESCE(a.original_summary, '')) LIKE ? ESCAPE '\' OR LOWER(COALESCE(c.content, '')) LIKE ? ESCAPE '\')`)
		conditionArgs = append(conditionArgs, pattern, pattern, pattern)
	}
	for _, term := range plainTerms {
		appendCondition(term, false)
	}
	for _, term := range patternTerms {
		appendCondition(term, true)
	}

	// Rank candidates in SQLite before applying LIMIT. Without this preliminary
	// ordering, a broad AI-expanded term could fill the arbitrary row set and
	// hide a later exact-title match from the deterministic Go ranker.
	scoreParts := make([]string, 0, 1+len(requiredTerms)+len(patternTerms))
	scoreArgs := make([]any, 0, (1+len(requiredTerms)+len(patternTerms))*3)
	appendScore := func(term string, allowWildcard bool, titleWeight, summaryWeight, contentWeight int) {
		pattern := searchLikePattern(term, allowWildcard)
		scoreParts = append(scoreParts, fmt.Sprintf(
			`CASE WHEN LOWER(a.title) LIKE ? ESCAPE '\' THEN %d ELSE 0 END + CASE WHEN LOWER(COALESCE(a.summary, '') || ' ' || COALESCE(a.original_summary, '')) LIKE ? ESCAPE '\' THEN %d ELSE 0 END + CASE WHEN LOWER(COALESCE(c.content, '')) LIKE ? ESCAPE '\' THEN %d ELSE 0 END`,
			titleWeight, summaryWeight, contentWeight,
		))
		scoreArgs = append(scoreArgs, pattern, pattern, pattern)
	}
	for _, term := range originalTerms {
		appendScore(term, false, 120, 70, 30)
	}
	for _, term := range requiredTerms {
		appendScore(term, false, 24, 12, 5)
	}
	for _, term := range patternTerms {
		appendScore(term, true, 30, 15, 7)
	}
	// Optional terms never broaden the candidate set. A bounded subset only
	// breaks ties among articles which already match the user's query or a core
	// expansion term.
	for _, term := range uniqueSearchTerms(search.Optional, 6) {
		appendScore(term, false, 6, 3, 1)
	}
	preliminaryScore := "0"
	if len(scoreParts) > 0 {
		preliminaryScore = strings.Join(scoreParts, " + ")
	}

	candidateLimit := search.Limit * 2
	if candidateLimit < 200 {
		candidateLimit = 200
	}
	if candidateLimit > 500 {
		candidateLimit = 500
	}
	args := append(scoreArgs, conditionArgs...)
	args = append(args, candidateLimit)

	// The limited CTE carries only article IDs and scores. Full cached bodies are
	// loaded for at most candidateLimit rows in the outer query, rather than
	// becoming part of the temporary sort payload.
	query := `
		WITH ranked_candidates AS MATERIALIZED (
			SELECT a.id AS article_id,
			       (` + preliminaryScore + `) AS candidate_score,
			       a.published_at AS candidate_time
			FROM articles a
			LEFT JOIN article_contents c ON a.id = c.article_id
			WHERE a.is_hidden = 0 AND (` + strings.Join(conditions, " OR ") + `)
			ORDER BY candidate_score DESC, candidate_time DESC, a.id DESC
			LIMIT ?
		)
			SELECT a.id, a.feed_id, a.title, a.url, a.image_url, a.audio_url, a.video_url,
			       a.published_at, a.is_read, a.is_favorite, a.is_hidden,
			       a.is_read_later, a.translated_title, a.summary, a.original_summary, a.freshrss_item_id,
			       f.title, a.author, COALESCE(c.content, '')
			FROM ranked_candidates ranked
			JOIN articles a ON a.id = ranked.article_id
			JOIN feeds f ON a.feed_id = f.id
			LEFT JOIN article_contents c ON a.id = c.article_id
			ORDER BY ranked.candidate_score DESC, ranked.candidate_time DESC, a.id DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("execute parameterized AI search: %w", err)
	}
	defer rows.Close()

	candidates := make([]aiSearchCandidate, 0)
	for rows.Next() {
		candidate, scanErr := scanAISearchCandidate(rows)
		if scanErr != nil {
			return nil, fmt.Errorf("scan AI search article: %w", scanErr)
		}
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate AI search articles: %w", err)
	}

	results := make([]AISearchResult, 0, len(candidates))
	for _, candidate := range candidates {
		result := rankAISearchCandidate(candidate, search)
		if result.RelevanceScore > 0 {
			results = append(results, result)
		}
	}
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].RelevanceScore != results[j].RelevanceScore {
			return results[i].RelevanceScore > results[j].RelevanceScore
		}
		return results[i].Article.PublishedAt.After(results[j].Article.PublishedAt)
	})
	if len(results) > search.Limit {
		results = results[:search.Limit]
	}
	return results, nil
}

func scanAISearchCandidate(rows *sql.Rows) (aiSearchCandidate, error) {
	var candidate aiSearchCandidate
	a := &candidate.article
	var imageURL, audioURL, videoURL, translatedTitle, summary, originalSummary, freshrssItemID, author sql.NullString
	var publishedAt sql.NullTime
	if err := rows.Scan(
		&a.ID, &a.FeedID, &a.Title, &a.URL, &imageURL, &audioURL, &videoURL,
		&publishedAt, &a.IsRead, &a.IsFavorite, &a.IsHidden,
		&a.IsReadLater, &translatedTitle, &summary, &originalSummary, &freshrssItemID,
		&a.FeedTitle, &author, &candidate.content,
	); err != nil {
		return candidate, err
	}
	a.ImageURL = imageURL.String
	a.AudioURL = audioURL.String
	a.VideoURL = videoURL.String
	a.TranslatedTitle = translatedTitle.String
	a.Summary = summary.String
	a.OriginalSummary = originalSummary.String
	candidate.summaryText = strings.TrimSpace(summary.String + " " + originalSummary.String)
	a.FreshRSSItemID = freshrssItemID.String
	a.Author = author.String
	if publishedAt.Valid {
		a.PublishedAt = publishedAt.Time
	}
	return candidate, nil
}

func rankAISearchCandidate(candidate aiSearchCandidate, search AISearchQuery) AISearchResult {
	title := cleanSearchText(candidate.article.Title)
	summary := cleanSearchText(candidate.summaryText)
	content := cleanSearchText(candidate.content)
	fields := map[string]string{"title": title, "summary": summary, "content": content}
	matchedTerms := make([]string, 0)
	matchedFields := make([]string, 0, 3)
	fieldSeen := make(map[string]bool)
	score := 0.0

	apply := func(term string, wildcard bool, weights map[string]float64) bool {
		term = strings.TrimSpace(term)
		if normalizeSearchNeedle(term) == "" {
			return false
		}
		matched := false
		for field, value := range fields {
			if searchTermMatches(value, term, wildcard) {
				score += weights[field]
				matched = true
				if !fieldSeen[field] {
					fieldSeen[field] = true
					matchedFields = append(matchedFields, field)
				}
			}
		}
		if matched && !containsFold(matchedTerms, term) {
			matchedTerms = append(matchedTerms, term)
		}
		return matched
	}

	apply(search.Original, false, map[string]float64{"title": 36, "summary": 20, "content": 10})
	required := uniqueSearchTerms(search.Required, 8)
	requiredMatches := 0
	for _, term := range required {
		if apply(term, false, map[string]float64{"title": 14, "summary": 8, "content": 4}) {
			requiredMatches++
		}
	}
	if len(required) > 0 {
		score += 10 * float64(requiredMatches) / float64(len(required))
	}
	for _, term := range uniqueSearchTerms(search.Patterns, 4) {
		apply(term, true, map[string]float64{"title": 18, "summary": 10, "content": 5})
	}
	for _, term := range uniqueSearchTerms(search.Optional, 12) {
		apply(term, false, map[string]float64{"title": 4, "summary": 2, "content": 1})
	}

	fieldOrder := map[string]int{"title": 0, "summary": 1, "content": 2}
	sort.SliceStable(matchedFields, func(i, j int) bool {
		return fieldOrder[matchedFields[i]] < fieldOrder[matchedFields[j]]
	})
	excerptSource := summary
	if excerptSource == "" || !containsAnyFold(excerptSource, matchedTerms) {
		excerptSource = content
	}
	if excerptSource == "" {
		excerptSource = title
	}

	return AISearchResult{
		Article:        candidate.article,
		RelevanceScore: score,
		MatchedTerms:   matchedTerms,
		MatchedFields:  matchedFields,
		Excerpt:        makeSearchExcerpt(excerptSource, matchedTerms, 220),
	}
}

func uniqueSearchTerms(terms []string, max int) []string {
	result := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" || len([]rune(term)) > 80 || containsFold(result, term) {
			continue
		}
		result = append(result, term)
		if len(result) >= max {
			break
		}
	}
	return result
}

func validSearchPatterns(patterns []string, max int) []string {
	result := make([]string, 0, len(patterns))
	for _, pattern := range uniqueSearchTerms(patterns, max) {
		// A pattern containing only '%' (or whitespace between wildcards) is
		// equivalent to an unbounded table scan and offers no relevance evidence.
		if normalizeSearchNeedle(pattern) == "" {
			continue
		}
		result = append(result, pattern)
	}
	return result
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func normalizeSearchNeedle(term string) string {
	term = strings.ToLower(strings.ReplaceAll(term, "%", " "))
	return strings.Join(strings.Fields(term), " ")
}

func containsAnyFold(text string, terms []string) bool {
	lower := strings.ToLower(text)
	for _, term := range terms {
		if needle := normalizeSearchNeedle(term); needle != "" && strings.Contains(lower, needle) {
			return true
		}
	}
	return false
}

func searchLikePattern(term string, allowWildcard bool) string {
	term = strings.ToLower(strings.TrimSpace(term))
	var builder strings.Builder
	for _, char := range term {
		switch char {
		case '\\':
			builder.WriteString(`\\`)
		case '_':
			builder.WriteString(`\_`)
		case '%':
			if allowWildcard {
				builder.WriteRune('%')
			} else {
				builder.WriteString(`\%`)
			}
		default:
			builder.WriteRune(char)
		}
	}
	return "%" + builder.String() + "%"
}

func searchTermMatches(text, term string, wildcard bool) bool {
	lower := strings.ToLower(text)
	if !wildcard || !strings.Contains(term, "%") {
		return strings.Contains(lower, strings.ToLower(strings.TrimSpace(term)))
	}
	position := 0
	for _, part := range strings.Split(strings.ToLower(term), "%") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		index := strings.Index(lower[position:], part)
		if index < 0 {
			return false
		}
		position += index + len(part)
	}
	return true
}

func cleanSearchText(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	root, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return strings.Join(strings.Fields(stdhtml.UnescapeString(raw)), " ")
	}
	var builder strings.Builder
	var walk func(*html.Node, bool)
	walk = func(node *html.Node, skip bool) {
		if node.Type == html.ElementNode {
			switch strings.ToLower(node.Data) {
			case "script", "style", "noscript", "template":
				skip = true
			case "p", "div", "br", "li", "section", "article", "h1", "h2", "h3":
				builder.WriteByte(' ')
			}
		}
		if node.Type == html.TextNode && !skip {
			builder.WriteString(node.Data)
			builder.WriteByte(' ')
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child, skip)
		}
	}
	walk(root, false)
	return strings.Join(strings.Fields(stdhtml.UnescapeString(builder.String())), " ")
}

func makeSearchExcerpt(text string, terms []string, maxRunes int) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	lower := strings.ToLower(text)
	matchByte := -1
	for _, term := range terms {
		if needle := normalizeSearchNeedle(term); needle != "" {
			if index := strings.Index(lower, needle); index >= 0 && (matchByte < 0 || index < matchByte) {
				matchByte = index
			}
		}
	}
	matchRune := 0
	if matchByte > 0 {
		matchRune = len([]rune(text[:matchByte]))
	}
	start := matchRune - maxRunes/3
	if start < 0 {
		start = 0
	}
	end := start + maxRunes
	if end > len(runes) {
		end = len(runes)
		start = end - maxRunes
	}
	for start > 0 && !unicode.IsSpace(runes[start]) {
		start--
	}
	prefix, suffix := "", ""
	if start > 0 {
		prefix = "…"
	}
	if end < len(runes) {
		suffix = "…"
	}
	return prefix + strings.TrimSpace(string(runes[start:end])) + suffix
}

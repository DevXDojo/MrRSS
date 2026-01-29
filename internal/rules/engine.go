package rules

import (
	"context"
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"time"

	"MrRSS/internal/database"
	"MrRSS/internal/freshrss"
	"MrRSS/internal/models"
	"MrRSS/internal/rsshub"
)

// getFeedType returns the type code of a feed
// Possible values: "regular", "freshrss", "rsshub", "script", "xpath", "email"
func getFeedType(feed *models.Feed) string {
	// Check FreshRSS
	if feed.IsFreshRSSSource {
		return "freshrss"
	}

	// Check RSSHub
	if rsshub.IsRSSHubURL(feed.URL) {
		return "rsshub"
	}

	// Check custom script
	if feed.ScriptPath != "" {
		return "script"
	}

	// Check email
	if feed.Type == "email" {
		return "email"
	}

	// Check XPath
	if feed.Type == "HTML+XPath" || feed.Type == "XML+XPath" {
		return "xpath"
	}

	// Default: regular RSS/Atom feed
	return "regular"
}

// Condition represents a condition in a rule
type Condition struct {
	ID       int64    `json:"id"`
	Logic    string   `json:"logic"`    // "and", "or" (null for first condition)
	Negate   bool     `json:"negate"`   // NOT modifier for this condition
	Field    string   `json:"field"`    // "feed_name", "feed_category", "article_title", etc.
	Operator string   `json:"operator"` // "contains", "exact"
	Value    string   `json:"value"`    // Single value for text/date fields
	Values   []string `json:"values"`   // Multiple values for feed_name and feed_category
}

// Rule represents an automation rule
type Rule struct {
	ID         int64       `json:"id"`
	Name       string      `json:"name"`
	Enabled    bool        `json:"enabled"`
	Conditions []Condition `json:"conditions"`
	Actions    []string    `json:"actions"`  // "favorite", "unfavorite", "hide", "unhide", "mark_read", "mark_unread"
	Position   int         `json:"position"` // Execution order (0 = first)
}

// Engine handles rule application
type Engine struct {
	db *database.DB
}

// NewEngine creates a new rules engine
func NewEngine(db *database.DB) *Engine {
	return &Engine{db: db}
}

// ApplyRulesToArticles applies all enabled rules to a batch of articles.
// Each article is matched against rules in order, and only the first matching rule is applied.
// This prevents conflicting actions from multiple rules being applied to the same article.
func (e *Engine) ApplyRulesToArticles(articles []models.Article) (int, error) {
	// Load rules from settings
	rulesJSON, _ := e.db.GetSetting("rules")
	if rulesJSON == "" {
		return 0, nil
	}

	var rules []Rule
	if err := json.Unmarshal([]byte(rulesJSON), &rules); err != nil {
		log.Printf("Error parsing rules: %v", err)
		return 0, err
	}

	// Sort rules by position (ascending) to ensure execution order
	// Rules without a position field (backward compatibility) are treated as position 0
	sortRulesByPosition(rules)

	// Get feeds for category and title lookup
	feeds, err := e.db.GetFeeds()
	if err != nil {
		return 0, err
	}

	// Create maps of feed ID to feed data
	feedCategories := make(map[int64]string)
	feedTitles := make(map[int64]string)
	feedTypes := make(map[int64]string)
	feedIsImageMode := make(map[int64]bool)
	feedIsFreshRSS := make(map[int64]bool)
	feedTags := make(map[int64][]string)

	for _, feed := range feeds {
		feedCategories[feed.ID] = feed.Category
		feedTitles[feed.ID] = feed.Title
		feedTypes[feed.ID] = getFeedType(&feed)
		feedIsImageMode[feed.ID] = feed.IsImageMode
		feedIsFreshRSS[feed.ID] = feed.IsFreshRSSSource

		// Build tag names list for this feed
		tags, _ := e.db.GetFeedTags(feed.ID)
		tagNames := make([]string, len(tags))
		for i, tag := range tags {
			tagNames[i] = tag.Name
		}
		feedTags[feed.ID] = tagNames
	}

	affected := 0
	for _, article := range articles {
		for _, rule := range rules {
			if !rule.Enabled {
				continue
			}

			// Check if article matches conditions
			if matchesConditions(article, rule.Conditions, feedCategories, feedTitles, feedTypes, feedIsImageMode, feedIsFreshRSS, feedTags) {
				// Apply actions
				for _, action := range rule.Actions {
					if err := e.applyAction(article.ID, action); err != nil {
						log.Printf("Error applying action %s to article %d: %v", action, article.ID, err)
						continue
					}
				}
				affected++
				break // Only apply first matching rule per article to prevent conflicts
			}
		}
	}

	return affected, nil
}

// ApplyRule applies a single rule to all matching articles.
// Uses batch processing with a reasonable limit to avoid memory issues.
func (e *Engine) ApplyRule(rule Rule) (int, error) {
	// Get articles in batches to avoid memory issues with large datasets
	const batchSize = 10000
	articles, err := e.db.GetArticles("", 0, "", true, batchSize, 0)
	if err != nil {
		return 0, err
	}

	// Get feeds for category and title lookup
	feeds, err := e.db.GetFeeds()
	if err != nil {
		return 0, err
	}

	// Create maps of feed ID to feed data
	feedCategories := make(map[int64]string)
	feedTitles := make(map[int64]string)
	feedTypes := make(map[int64]string)
	feedIsImageMode := make(map[int64]bool)
	feedIsFreshRSS := make(map[int64]bool)
	feedTags := make(map[int64][]string)

	for _, feed := range feeds {
		feedCategories[feed.ID] = feed.Category
		feedTitles[feed.ID] = feed.Title
		feedTypes[feed.ID] = getFeedType(&feed)
		feedIsImageMode[feed.ID] = feed.IsImageMode
		feedIsFreshRSS[feed.ID] = feed.IsFreshRSSSource

		// Build tag names list for this feed
		tags, _ := e.db.GetFeedTags(feed.ID)
		tagNames := make([]string, len(tags))
		for i, tag := range tags {
			tagNames[i] = tag.Name
		}
		feedTags[feed.ID] = tagNames
	}

	affected := 0
	for _, article := range articles {
		if matchesConditions(article, rule.Conditions, feedCategories, feedTitles, feedTypes, feedIsImageMode, feedIsFreshRSS, feedTags) {
			for _, action := range rule.Actions {
				if err := e.applyAction(article.ID, action); err != nil {
					log.Printf("Error applying action %s to article %d: %v", action, article.ID, err)
					continue
				}
			}
			affected++
		}
	}

	return affected, nil
}

// matchesConditions checks if an article matches the rule conditions
func matchesConditions(article models.Article, conditions []Condition, feedCategories map[int64]string, feedTitles map[int64]string, feedTypes map[int64]string, feedIsImageMode map[int64]bool, feedIsFreshRSS map[int64]bool, feedTags map[int64][]string) bool {
	// If no conditions, apply to all articles
	if len(conditions) == 0 {
		return true
	}

	result := evaluateCondition(article, conditions[0], feedCategories, feedTitles, feedTypes, feedIsImageMode, feedIsFreshRSS, feedTags)

	for i := 1; i < len(conditions); i++ {
		condition := conditions[i]
		conditionResult := evaluateCondition(article, condition, feedCategories, feedTitles, feedTypes, feedIsImageMode, feedIsFreshRSS, feedTags)

		switch condition.Logic {
		case "and":
			result = result && conditionResult
		case "or":
			result = result || conditionResult
		}
	}

	return result
}

// evaluateCondition evaluates a single rule condition
func evaluateCondition(article models.Article, condition Condition, feedCategories map[int64]string, feedTitles map[int64]string, feedTypes map[int64]string, feedIsImageMode map[int64]bool, feedIsFreshRSS map[int64]bool, feedTags map[int64][]string) bool {
	var result bool

	switch condition.Field {
	case "feed_name":
		feedTitle := feedTitles[article.FeedID]
		if feedTitle == "" {
			feedTitle = article.FeedTitle
		}
		result = matchMultiSelect(feedTitle, condition.Values, condition.Value)

	case "feed_category":
		feedCategory := feedCategories[article.FeedID]
		result = matchMultiSelect(feedCategory, condition.Values, condition.Value)

	case "article_title":
		if condition.Value == "" {
			result = true
		} else {
			lowerValue := strings.ToLower(condition.Value)
			lowerTitle := strings.ToLower(article.Title)
			switch condition.Operator {
			case "exact":
				result = lowerTitle == lowerValue
			case "regex":
				matched, err := regexp.MatchString(condition.Value, article.Title)
				if err != nil {
					log.Printf("Invalid regex pattern: %v", err)
					result = false
				} else {
					result = matched
				}
			default:
				result = strings.Contains(lowerTitle, lowerValue)
			}
		}

	case "feed_type":
		feedType := feedTypes[article.FeedID]
		result = matchMultiSelect(feedType, condition.Values, condition.Value)

	case "feed_tags":
		articleTags := feedTags[article.FeedID]
		// Check if any tag matches
		result = matchMultiSelectTags(articleTags, condition.Values, condition.Value)

	case "is_freshrss_feed":
		if condition.Value == "" {
			result = true
		} else {
			wantFreshRSS := condition.Value == "true"
			result = feedIsFreshRSS[article.FeedID] == wantFreshRSS
		}

	case "is_image_mode_feed":
		if condition.Value == "" {
			result = true
		} else {
			wantImageMode := condition.Value == "true"
			result = feedIsImageMode[article.FeedID] == wantImageMode
		}

	case "published_after":
		if condition.Value == "" {
			result = true
		} else {
			afterDate, err := time.Parse("2006-01-02", condition.Value)
			if err != nil {
				result = true
			} else {
				result = article.PublishedAt.After(afterDate) || article.PublishedAt.Equal(afterDate)
			}
		}

	case "published_before":
		if condition.Value == "" {
			result = true
		} else {
			beforeDate, err := time.Parse("2006-01-02", condition.Value)
			if err != nil {
				result = true
			} else {
				articleDateOnly := article.PublishedAt.UTC().Truncate(24 * time.Hour)
				beforeDateOnly := beforeDate.Truncate(24 * time.Hour)
				result = !articleDateOnly.After(beforeDateOnly)
			}
		}

	case "is_read":
		if condition.Value == "" {
			result = true
		} else {
			wantRead := condition.Value == "true"
			result = article.IsRead == wantRead
		}

	case "is_favorite":
		if condition.Value == "" {
			result = true
		} else {
			wantFavorite := condition.Value == "true"
			result = article.IsFavorite == wantFavorite
		}

	case "is_hidden":
		if condition.Value == "" {
			result = true
		} else {
			wantHidden := condition.Value == "true"
			result = article.IsHidden == wantHidden
		}

	case "is_read_later":
		if condition.Value == "" {
			result = true
		} else {
			wantReadLater := condition.Value == "true"
			result = article.IsReadLater == wantReadLater
		}

	default:
		result = true
	}

	// Apply NOT modifier
	if condition.Negate {
		return !result
	}
	return result
}

// matchMultiSelect checks if fieldValue matches any of the selected values
func matchMultiSelect(fieldValue string, values []string, singleValue string) bool {
	if len(values) > 0 {
		lowerField := strings.ToLower(fieldValue)
		for _, val := range values {
			if strings.Contains(lowerField, strings.ToLower(val)) {
				return true
			}
		}
		return false
	} else if singleValue != "" {
		return strings.Contains(strings.ToLower(fieldValue), strings.ToLower(singleValue))
	}
	return true
}

// matchMultiSelectTags checks if any of the article's tags match the selected values
func matchMultiSelectTags(articleTags []string, values []string, singleValue string) bool {
	if len(values) > 0 {
		// Check if any of the selected values match any of the article's tags
		for _, val := range values {
			lowerVal := strings.ToLower(val)
			for _, tag := range articleTags {
				if strings.Contains(strings.ToLower(tag), lowerVal) {
					return true
				}
			}
		}
		return false
	} else if singleValue != "" {
		// Check if the single value matches any tag
		lowerVal := strings.ToLower(singleValue)
		for _, tag := range articleTags {
			if strings.Contains(strings.ToLower(tag), lowerVal) {
				return true
			}
		}
		return false
	}
	// No filter specified, match all
	return true
}

// applyAction applies an action to an article with FreshRSS sync if enabled
func (e *Engine) applyAction(articleID int64, action string) error {
	var syncReq *database.SyncRequest
	var err error

	// Apply the action and get sync request if applicable
	switch action {
	case "favorite":
		syncReq, err = e.db.SetArticleFavoriteWithSync(articleID, true)
	case "unfavorite":
		syncReq, err = e.db.SetArticleFavoriteWithSync(articleID, false)
	case "hide":
		err = e.db.SetArticleHidden(articleID, true)
	case "unhide":
		err = e.db.SetArticleHidden(articleID, false)
	case "mark_read":
		syncReq, err = e.db.MarkArticleReadWithSync(articleID, true)
	case "mark_unread":
		syncReq, err = e.db.MarkArticleReadWithSync(articleID, false)
	case "read_later":
		err = e.db.SetArticleReadLater(articleID, true)
	case "remove_read_later":
		err = e.db.SetArticleReadLater(articleID, false)
	default:
		log.Printf("Unknown action: %s", action)
		return nil
	}

	if err != nil {
		return err
	}

	// Perform immediate sync to FreshRSS if needed
	if syncReq != nil {
		go e.performImmediateSync(syncReq)
	}

	return nil
}

// performImmediateSync performs an immediate sync to FreshRSS in a background goroutine
func (e *Engine) performImmediateSync(syncReq *database.SyncRequest) {
	// Check if FreshRSS is enabled and configured
	enabled, _ := e.db.GetSetting("freshrss_enabled")
	if enabled != "true" {
		return
	}

	serverURL, username, password, err := e.db.GetFreshRSSConfig()
	if err != nil || serverURL == "" || username == "" || password == "" {
		log.Printf("[Rule Sync] FreshRSS not configured, skipping sync")
		return
	}

	// Create sync service
	syncService := freshrss.NewBidirectionalSyncService(serverURL, username, password, e.db)

	// Perform immediate sync
	ctx := context.Background()
	err = syncService.SyncArticleStatus(ctx, syncReq.ArticleID, syncReq.ArticleURL, syncReq.Action)
	if err != nil {
		log.Printf("[Rule Sync] Failed for article %d: %v", syncReq.ArticleID, err)
		// Enqueue for retry during next global sync
		_ = e.db.EnqueueSyncChange(syncReq.ArticleID, syncReq.ArticleURL, syncReq.Action)
		log.Printf("[Rule Sync] Enqueued article %d for retry", syncReq.ArticleID)
	} else {
		log.Printf("[Rule Sync] Success for article %d: %s", syncReq.ArticleID, syncReq.Action)
	}
}

// sortRulesByPosition sorts rules by their position field in ascending order
func sortRulesByPosition(rules []Rule) {
	// Use the built-in sort package with a custom comparator
	for i := 0; i < len(rules); i++ {
		for j := i + 1; j < len(rules); j++ {
			// Compare positions, defaulting to 0 for backward compatibility
			posI := rules[i].Position
			posJ := rules[j].Position
			if posI > posJ {
				// Swap
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}
}

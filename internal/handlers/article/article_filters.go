package article

import (
	"log"
	"regexp"
	"strings"
	"time"

	"MrRSS/internal/models"
)

// FilterCondition represents a single filter condition from the frontend
type FilterCondition struct {
	ID       int64    `json:"id"`
	Logic    string   `json:"logic"`    // "and", "or" (null for first condition)
	Negate   bool     `json:"negate"`   // NOT modifier for this condition
	Field    string   `json:"field"`    // "feed_name", "feed_category", "article_title", "published_after", "published_before"
	Operator string   `json:"operator"` // "contains", "exact" (null for date fields and multi-select)
	Value    string   `json:"value"`    // Single value for text/date fields
	Values   []string `json:"values"`   // Multiple values for feed_name and feed_category
}

// FilterRequest represents the request body for filtered articles
type FilterRequest struct {
	Conditions []FilterCondition `json:"conditions"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
}

// FilterResponse represents the response for filtered articles with pagination info
type FilterResponse struct {
	Articles []models.Article `json:"articles"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	Limit    int              `json:"limit"`
	HasMore  bool             `json:"has_more"`
}

// evaluateArticleConditions evaluates all filter conditions for an article
func evaluateArticleConditions(article models.Article, conditions []FilterCondition, feedCategories map[int64]string, feedTypes map[int64]string, feedIsImageMode map[int64]bool, feedTags map[int64][]string) bool {
	if len(conditions) == 0 {
		return true
	}

	result := evaluateSingleCondition(article, conditions[0], feedCategories, feedTypes, feedIsImageMode, feedTags)

	for i := 1; i < len(conditions); i++ {
		condition := conditions[i]
		conditionResult := evaluateSingleCondition(article, condition, feedCategories, feedTypes, feedIsImageMode, feedTags)

		switch condition.Logic {
		case "and":
			result = result && conditionResult
		case "or":
			result = result || conditionResult
		}
	}

	return result
}

// matchMultiSelectContains checks if fieldValue matches any of the selected values using contains logic
func matchMultiSelectContains(fieldValue string, values []string, singleValue string) bool {
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

// evaluateSingleCondition evaluates a single filter condition for an article
func evaluateSingleCondition(article models.Article, condition FilterCondition, feedCategories map[int64]string, feedTypes map[int64]string, feedIsImageMode map[int64]bool, feedTags map[int64][]string) bool {
	var result bool

	switch condition.Field {
	case "feed_name":
		result = matchMultiSelectContains(article.FeedTitle, condition.Values, condition.Value)

	case "feed_category":
		feedCategory := feedCategories[article.FeedID]
		result = matchMultiSelectContains(feedCategory, condition.Values, condition.Value)

	case "feed_tags":
		articleTags := feedTags[article.FeedID]
		// Check if any tag matches
		result = matchMultiSelectTags(articleTags, condition.Values, condition.Value)

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
		result = matchMultiSelectContains(feedType, condition.Values, condition.Value)

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
				log.Printf("Invalid date format for published_after filter: %s", condition.Value)
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
				log.Printf("Invalid date format for published_before filter: %s", condition.Value)
				result = true
			} else {
				// For "before Dec 24 (inclusive)", we want articles published on Dec 24 or earlier
				// We compare dates only (not times) - any article from Dec 24 should be included
				// Truncate to remove time component, preserving date in local timezone context
				articleDateOnly := article.PublishedAt.UTC().Truncate(24 * time.Hour)
				beforeDateOnly := beforeDate.Truncate(24 * time.Hour)
				// Include articles on the selected date or before
				result = !articleDateOnly.After(beforeDateOnly)
			}
		}

	case "is_read":
		// Filter by read/unread status
		if condition.Value == "" {
			result = true
		} else {
			wantRead := condition.Value == "true"
			result = article.IsRead == wantRead
		}

	case "is_favorite":
		// Filter by favorite/unfavorite status
		if condition.Value == "" {
			result = true
		} else {
			wantFavorite := condition.Value == "true"
			result = article.IsFavorite == wantFavorite
		}

	case "is_read_later":
		// Filter by read later status
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

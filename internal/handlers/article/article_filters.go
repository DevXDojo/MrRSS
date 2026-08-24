package article

import (
	"log"
	"regexp"
	"strconv"
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
// Operator precedence: NOT > AND > OR
func evaluateArticleConditions(
	article models.Article,
	conditions []FilterCondition,
	feedCategories map[int64]string,
	feedTypes map[int64]string,
	feedIsImageMode map[int64]bool,
	feedTags map[int64][]string,
	feedArticlesPerMonth map[int64]float64,
	feedLastUpdateStatus map[int64]string,
	articleContents map[int64]string,
) bool {
	if len(conditions) == 0 {
		return true
	}

	// Step 1: Evaluate all individual conditions (NOT is applied at this level)
	conditionResults := make([]bool, len(conditions))
	for i, condition := range conditions {
		conditionResults[i] = evaluateSingleCondition(article, condition, feedCategories, feedTypes, feedIsImageMode, feedTags, feedArticlesPerMonth, feedLastUpdateStatus, articleContents)
	}

	// Step 2: Process all AND connections first (higher precedence)
	// We merge conditions connected by AND into a single result
	i := 0
	for i < len(conditionResults) {
		if i > 0 && conditions[i].Logic == "and" {
			// Merge with previous result using AND
			conditionResults[i-1] = conditionResults[i-1] && conditionResults[i]
			// Remove current element
			conditionResults = append(conditionResults[:i], conditionResults[i+1:]...)
			conditions = append(conditions[:i], conditions[i+1:]...)
		} else {
			i++
		}
	}

	// Step 3: Process all OR connections (lower precedence)
	if len(conditionResults) == 0 {
		return true
	}

	result := conditionResults[0]
	for i := 1; i < len(conditionResults); i++ {
		result = result || conditionResults[i]
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
func evaluateSingleCondition(
	article models.Article,
	condition FilterCondition,
	feedCategories map[int64]string,
	feedTypes map[int64]string,
	feedIsImageMode map[int64]bool,
	feedTags map[int64][]string,
	feedArticlesPerMonth map[int64]float64,
	feedLastUpdateStatus map[int64]string,
	articleContents map[int64]string,
) bool {
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

	case "author":
		if condition.Value == "" {
			result = true
		} else {
			lowerValue := strings.ToLower(condition.Value)
			lowerAuthor := strings.ToLower(article.Author)
			switch condition.Operator {
			case "exact":
				result = lowerAuthor == lowerValue
			case "regex":
				matched, err := regexp.MatchString(condition.Value, article.Author)
				if err != nil {
					log.Printf("Invalid regex pattern: %v", err)
					result = false
				} else {
					result = matched
				}
			default:
				result = strings.Contains(lowerAuthor, lowerValue)
			}
		}

	case "url":
		if condition.Value == "" {
			result = true
		} else {
			lowerValue := strings.ToLower(condition.Value)
			lowerURL := strings.ToLower(article.URL)
			switch condition.Operator {
			case "exact":
				result = lowerURL == lowerValue
			case "regex":
				matched, err := regexp.MatchString(condition.Value, article.URL)
				if err != nil {
					log.Printf("Invalid regex pattern: %v", err)
					result = false
				} else {
					result = matched
				}
			default:
				result = strings.Contains(lowerURL, lowerValue)
			}
		}

	case "article_content":
		// Filter by article content (if cached)
		if condition.Value == "" {
			result = true
		} else {
			content, hasContent := articleContents[article.ID]
			if !hasContent {
				// No content cached, treat as not matching
				result = false
			} else {
				lowerValue := strings.ToLower(condition.Value)
				lowerContent := strings.ToLower(content)
				switch condition.Operator {
				case "exact":
					result = lowerContent == lowerValue
				case "regex":
					matched, err := regexp.MatchString(condition.Value, content)
					if err != nil {
						log.Printf("Invalid regex pattern: %v", err)
						result = false
					} else {
						result = matched
					}
				default:
					result = strings.Contains(lowerContent, lowerValue)
				}
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

	case "is_hidden":
		// Filter by hidden/unhidden status
		if condition.Value == "" {
			result = true
		} else {
			wantHidden := condition.Value == "true"
			result = article.IsHidden == wantHidden
		}

	case "is_read_later":
		// Filter by read later status
		if condition.Value == "" {
			result = true
		} else {
			wantReadLater := condition.Value == "true"
			result = article.IsReadLater == wantReadLater
		}

	case "has_summary":
		// Filter by whether article has a summary
		if condition.Value == "" {
			result = true
		} else {
			wantSummary := condition.Value == "true"
			result = (article.Summary != "") == wantSummary
		}

	case "has_translation":
		// Filter by whether article has a translated title
		if condition.Value == "" {
			result = true
		} else {
			wantTranslation := condition.Value == "true"
			result = (article.TranslatedTitle != "") == wantTranslation
		}

	case "has_image":
		// Filter by whether article has an image
		if condition.Value == "" {
			result = true
		} else {
			wantImage := condition.Value == "true"
			result = (article.ImageURL != "") == wantImage
		}

	case "has_audio":
		// Filter by whether article has audio
		if condition.Value == "" {
			result = true
		} else {
			wantAudio := condition.Value == "true"
			result = (article.AudioURL != "") == wantAudio
		}

	case "has_video":
		// Filter by whether article has video
		if condition.Value == "" {
			result = true
		} else {
			wantVideo := condition.Value == "true"
			result = (article.VideoURL != "") == wantVideo
		}

	case "published_after_hours":
		// Filter by articles published within the last N hours
		if condition.Value == "" {
			result = true
		} else {
			hours, err := strconv.Atoi(condition.Value)
			if err != nil || hours < 0 {
				log.Printf("Invalid hours value for published_after_hours filter: %s", condition.Value)
				result = true
			} else {
				cutoffTime := time.Now().Add(-time.Duration(hours) * time.Hour)
				result = article.PublishedAt.After(cutoffTime) || article.PublishedAt.Equal(cutoffTime)
			}
		}

	case "published_after_days":
		// Filter by articles published within the last N days
		if condition.Value == "" {
			result = true
		} else {
			days, err := strconv.Atoi(condition.Value)
			if err != nil || days < 0 {
				log.Printf("Invalid days value for published_after_days filter: %s", condition.Value)
				result = true
			} else {
				cutoffTime := time.Now().AddDate(0, 0, -days)
				result = article.PublishedAt.After(cutoffTime) || article.PublishedAt.Equal(cutoffTime)
			}
		}

	case "feed_articles_per_month":
		// Filter by feed's articles per month
		if condition.Value == "" {
			result = true
		} else {
			threshold, err := strconv.ParseFloat(condition.Value, 64)
			if err != nil {
				log.Printf("Invalid threshold value for feed_articles_per_month filter: %s", condition.Value)
				result = true
			} else {
				articlesPerMonth := feedArticlesPerMonth[article.FeedID]
				// Default to treating as "greater than or equal"
				result = articlesPerMonth >= threshold
			}
		}

	case "feed_last_update_status":
		// Filter by feed's last update status
		if condition.Value == "" {
			result = true
		} else {
			status := feedLastUpdateStatus[article.FeedID]
			result = strings.EqualFold(status, condition.Value)
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

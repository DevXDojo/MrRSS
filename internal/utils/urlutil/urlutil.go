// Package urlutil provides URL manipulation utilities including normalization,
// comparison, and article deduplication helpers.
package urlutil

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// NormalizeFeedURL ensures the feed URL has a protocol prefix.
func NormalizeFeedURL(feedURL string) string {
	if feedURL == "" {
		return feedURL
	}

	feedURL = strings.TrimSpace(feedURL)

	protocols := []string{
		"http://", "https://", "rsshub://", "script://",
		"email://", "feed://", "ftp://", "file://",
	}

	for _, protocol := range protocols {
		if strings.HasPrefix(strings.ToLower(feedURL), protocol) {
			return feedURL
		}
	}

	return "https://" + feedURL
}

// NormalizeURLForComparison returns a normalized URL for comparison purposes.
func NormalizeURLForComparison(rawURL string) string {
	if rawURL == "" {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if parsed.Scheme == "" {
		return rawURL
	}
	return parsed.Scheme + "://" + parsed.Host + parsed.Path
}

// URLsMatch checks if two URLs refer to the same article.
func URLsMatch(url1, url2 string) bool {
	if url1 == url2 {
		return true
	}
	return normalizeURLForMatching(url1) == normalizeURLForMatching(url2)
}

func normalizeURLForMatching(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if parsed.Scheme == "" {
		return rawURL
	}

	result := parsed.Scheme + "://" + parsed.Host + parsed.Path

	query := parsed.Query()
	if len(query) > 0 {
		importantParams := make(url.Values)
		for key, values := range query {
			if isImportantParameter(key, values) {
				importantParams[key] = values
			}
		}
		if len(importantParams) > 0 {
			result += "?" + importantParams.Encode()
		}
	}

	return result
}

func isImportantParameter(key string, values []string) bool {
	if len(values) == 0 {
		return false
	}

	value := values[0]

	if isIDParameter(key) {
		return true
	}

	if isTrackingParameter(key) {
		return false
	}

	if len(value) > 50 && looksLikeTrackingToken(value) {
		return false
	}

	if isNumeric(value) && !looksLikeTrackingToken(value) {
		return true
	}

	if len(key) <= 3 && len(value) <= 20 {
		return true
	}

	if containsMeaningfulWords(key) {
		return true
	}

	return !looksLikeTrackingToken(value)
}

func isIDParameter(key string) bool {
	keyLower := strings.ToLower(key)

	exactMatches := []string{"id", "mid", "cid", "uid", "pid", "tid", "aid", "bid", "did", "eid", "fid", "gid", "hid", "iid", "jid", "kid", "lid", "nid", "oid", "qid", "rid", "sid", "vid", "wid", "xid", "yid", "zid"}
	for _, match := range exactMatches {
		if keyLower == match {
			return true
		}
	}

	idPatterns := []string{"_id", "id_", "article", "post", "entry", "item", "thread", "topic", "page", "__biz", "idx", "pmid"}
	for _, pattern := range idPatterns {
		if strings.Contains(keyLower, pattern) {
			return true
		}
	}

	return false
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isTrackingParameter(key string) bool {
	keyLower := strings.ToLower(key)
	trackingPrefixes := []string{"utm_", "fbclid", "gclid", "msclkid", "ttclid", "_ga", "_gid", "_gat"}
	exactMatches := []string{"ref", "referrer", "source", "campaign", "medium", "term", "content", "fc", "sn"}

	for _, prefix := range trackingPrefixes {
		if strings.HasPrefix(keyLower, prefix) {
			return true
		}
	}

	for _, match := range exactMatches {
		if keyLower == match {
			return true
		}
	}

	return false
}

func looksLikeTrackingToken(value string) bool {
	if len(value) < 10 {
		return false
	}

	hasLower := strings.ContainsAny(value, "abcdefghijklmnopqrstuvwxyz")
	hasUpper := strings.ContainsAny(value, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasDigit := strings.ContainsAny(value, "0123456789")
	hasSpecial := strings.ContainsAny(value, "-_.")

	charTypeCount := 0
	if hasLower {
		charTypeCount++
	}
	if hasUpper {
		charTypeCount++
	}
	if hasDigit {
		charTypeCount++
	}
	if hasSpecial {
		charTypeCount++
	}

	if charTypeCount >= 3 {
		return true
	}

	if charTypeCount == 1 && hasDigit && len(value) > 12 {
		return true
	}

	return false
}

func containsMeaningfulWords(key string) bool {
	keyLower := strings.ToLower(key)
	meaningfulWords := []string{"lang", "locale", "format", "type", "category", "tag", "section", "view", "mode"}

	for _, word := range meaningfulWords {
		if strings.Contains(keyLower, word) {
			return true
		}
	}

	return false
}

// GenerateArticleUniqueID generates a unique identifier for an article.
func GenerateArticleUniqueID(title string, feedID int64, publishedAt time.Time, hasValidPublishedTime bool) string {
	title = strings.TrimSpace(title)

	var dateStr string
	if hasValidPublishedTime {
		dateStr = publishedAt.Format("2006-01-02")
	} else {
		dateStr = ""
	}

	data := fmt.Sprintf("%s|%d|%s", title, feedID, dateStr)
	hash := md5.Sum([]byte(data))
	return strings.ToLower(hex.EncodeToString(hash[:]))
}

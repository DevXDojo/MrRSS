package notification

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf16"

	"MrRSS/internal/database"
	"MrRSS/internal/utils/textutil"
)

func matches(a database.NotificationArticle, r Rule) bool {
	if a.IsHidden || (r.UnreadOnly && a.IsRead) {
		return false
	}
	if len(r.Conditions) == 0 {
		return true
	}
	for _, c := range r.Conditions {
		value := ""
		switch c.Field {
		case "title":
			value = a.Title
		case "summary":
			value = a.Summary + " " + a.OriginalSummary
		case "feed":
			value = a.FeedTitle
		case "category":
			value = a.Category
		case "author":
			value = a.Author
		case "url":
			value = a.URL
		case "favorite":
			value = strconv.FormatBool(a.IsFavorite)
		case "read_later":
			value = strconv.FormatBool(a.IsReadLater)
		}
		value = strings.ToLower(textutil.ArticlePlainText(value))
		wanted := strings.ToLower(strings.TrimSpace(c.Value))
		ok := false
		switch c.Operator {
		case "contains":
			ok = strings.Contains(value, wanted)
		case "not_contains":
			ok = !strings.Contains(value, wanted)
		case "equals":
			ok = value == wanted
		case "not_equals":
			ok = value != wanted
		}
		if r.Match == "all" && !ok {
			return false
		}
		if r.Match == "any" && ok {
			return true
		}
	}
	return r.Match == "all"
}

// UTF-16 bounds are conservative for all three providers, including emoji.
func units(s string) int { return len(utf16.Encode([]rune(s))) }
func shorten(s string, max int) string {
	if units(s) <= max {
		return s
	}
	var b strings.Builder
	n := 0
	for _, r := range s {
		cost := 1
		if r > 0xffff {
			cost = 2
		}
		if n+cost > max-1 {
			break
		}
		b.WriteRune(r)
		n += cost
	}
	return b.String() + "…"
}

func plain(s string) string { return strings.Join(strings.Fields(textutil.ArticlePlainText(s)), " ") }

func render(r Rule, articles []database.NotificationArticle) []string {
	if len(articles) == 0 {
		return nil
	}
	header := "MrRSS · " + shorten(plain(r.Name), 120)
	parts := []string{}
	current := header
	for _, a := range articles {
		item := fmt.Sprintf("\n\n%s\n%s", shorten(plain(a.Title), 250), shorten(plain(a.FeedTitle), 80))
		if r.IncludeSummary {
			summary := a.Summary
			if strings.TrimSpace(summary) == "" {
				summary = a.OriginalSummary
			}
			if summary = plain(summary); summary != "" {
				item += "\n" + shorten(summary, 360)
			}
		}
		if r.IncludeLink {
			u, err := url.Parse(a.URL)
			if err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && u.User == nil && units(u.String()) <= 800 {
				item += "\n" + u.String()
			}
		}
		if units(current+item) > 1900 {
			parts = append(parts, current)
			current = header
		}
		current += item
	}
	if current != header {
		parts = append(parts, current)
	}
	return parts
}

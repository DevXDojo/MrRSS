package rules

import (
	"reflect"
	"testing"

	"MrRSS/internal/models"
)

func TestConditionsPreserveRulesAcrossArticles(t *testing.T) {
	// Hide articles from this feed only when none of the three keywords match.
	conditions := []Condition{
		{Field: "feed_name", Values: []string{"News"}},
		{Logic: "and", Field: "article_title", Value: "alpha", Negate: true},
		{Logic: "and", Field: "article_title", Value: "beta", Negate: true},
		{Logic: "and", Field: "article_title", Value: "gamma", Negate: true},
	}
	wantConditions := append([]Condition(nil), conditions...)
	for round := 0; round < 2; round++ {
		for _, tc := range []struct {
			title string
			feed  string
			want  bool
		}{
			{"unrelated story", "News", true},
			{"ALPHA release", "News", false},
			{"beta release", "News", false},
			{"gamma release", "News", false},
			{"unrelated story", "Other", false},
			{"another story", "News", true},
		} {
			article := models.Article{Title: tc.title, FeedTitle: tc.feed}
			got := matchesConditions(article, conditions, nil, nil, nil, nil, nil, nil, nil)
			if got != tc.want {
				t.Errorf("round %d: %q in %q matched = %v, want %v", round, tc.title, tc.feed, got, tc.want)
			}
			if !reflect.DeepEqual(conditions, wantConditions) {
				t.Fatalf("matching %q mutated the stored rule: %#v", tc.title, conditions)
			}
		}
	}
}

func TestConditionPrecedence(t *testing.T) {
	// alpha OR (beta AND NOT gamma) OR (delta AND epsilon).
	conditions := []Condition{
		{Field: "article_title", Value: "alpha"},
		{Logic: "or", Field: "article_title", Value: "beta"},
		{Logic: "and", Field: "article_title", Value: "gamma", Negate: true},
		{Logic: "or", Field: "article_title", Value: "delta"},
		{Logic: "and", Field: "article_title", Value: "epsilon"},
	}
	for _, tc := range []struct {
		title string
		want  bool
	}{
		{"alpha gamma", true},
		{"beta", true},
		{"beta gamma", false},
		{"delta epsilon", true},
		{"delta", false},
		{"epsilon", false},
		{"unrelated", false},
	} {
		article := models.Article{Title: tc.title}
		if got := matchesConditions(article, conditions, nil, nil, nil, nil, nil, nil, nil); got != tc.want {
			t.Errorf("%q matched = %v, want %v", tc.title, got, tc.want)
		}
	}
	if !matchesConditions(models.Article{}, nil, nil, nil, nil, nil, nil, nil, nil) {
		t.Fatal("empty conditions must match every article")
	}
}

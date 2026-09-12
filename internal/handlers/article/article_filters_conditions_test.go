package article

import (
	"reflect"
	"testing"

	"MrRSS/internal/models"
)

func TestFilterConditionsRemainStableAcrossArticles(t *testing.T) {
	conditions := []FilterCondition{
		{Field: "feed_name", Value: "News"},
		{Logic: "and", Field: "article_title", Value: "alpha", Negate: true},
		{Logic: "and", Field: "article_title", Value: "beta", Negate: true},
		{Logic: "or", Field: "article_title", Value: "special"},
		{Logic: "and", Field: "article_title", Value: "report"},
	}
	wantConditions := append([]FilterCondition(nil), conditions...)
	for round := 0; round < 2; round++ {
		for _, tc := range []struct {
			title string
			feed  string
			want  bool
		}{
			{"unrelated story", "News", true},
			{"alpha release", "News", false},
			{"beta release", "News", false},
			{"unrelated story", "Other", false},
			{"special report", "Other", true},
			{"special", "Other", false},
			{"report", "Other", false},
		} {
			article := models.Article{Title: tc.title, FeedTitle: tc.feed}
			got := evaluateArticleConditions(article, conditions, nil, nil, nil, nil, nil, nil, nil)
			if got != tc.want {
				t.Errorf("round %d: %q in %q matched = %v, want %v", round, tc.title, tc.feed, got, tc.want)
			}
			if !reflect.DeepEqual(conditions, wantConditions) {
				t.Fatalf("matching %q mutated the request conditions", tc.title)
			}
		}
	}
	if !evaluateArticleConditions(models.Article{}, nil, nil, nil, nil, nil, nil, nil, nil) {
		t.Fatal("empty conditions must match every article")
	}
}

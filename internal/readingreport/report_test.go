package readingreport

import (
	"MrRSS/internal/models"
	"context"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"
)

type fixtureStore struct {
	articles map[int64]*models.Article
	content  map[int64]string
}

func (f fixtureStore) GetArticleByID(id int64) (*models.Article, error) { return f.articles[id], nil }
func (f fixtureStore) GetArticleContent(id int64) (string, bool, error) {
	c, ok := f.content[id]
	return c, ok, nil
}

func TestSourceCoverageAndInputBudget(t *testing.T) {
	db := fixtureStore{articles: map[int64]*models.Article{}, content: map[int64]string{}}
	ids := []int64{}
	for id := int64(1); id <= 20; id++ {
		db.articles[id] = &models.Article{ID: id, Title: "Article", URL: "https://example.org/article"}
		db.content[id] = "<script>not evidence</script><p>" + strings.Repeat("来源文字", 3000) + "</p>"
		ids = append(ids, id)
	}
	db.articles[2].OriginalSummary = "<p>RSS &amp; excerpt</p>"
	delete(db.content, 2)
	delete(db.content, 3)
	db.articles[3].URL = "javascript:alert(1)"
	db.articles[3].Summary = "An earlier generated summary is not source evidence"
	sources, err := LoadSources(context.Background(), db, ids)
	if err != nil {
		t.Fatal(err)
	}
	if sources[1].Kind != "rss_excerpt" || sources[1].Content != "RSS & excerpt" || sources[2].Kind != "missing" || sources[2].URL != "" {
		t.Fatalf("unexpected coverage: %+v", sources[1:3])
	}
	chars := 0
	for _, source := range sources {
		chars += source.Characters
		if !utf8.ValidString(source.Content) || strings.Contains(source.Content, "not evidence") {
			t.Fatal("invalid evidence")
		}
	}
	if chars > MaxInputRunes || !sources[0].Truncated {
		t.Fatalf("unbounded input: %d", chars)
	}
	request, err := BuildRequest(sources, "zh-CN", "产品影响")
	if err != nil || !strings.Contains(request.SystemPrompt, "Chinese") || !strings.Contains(request.UserPrompt, "产品影响") || strings.Contains(request.UserPrompt, "earlier generated") {
		t.Fatal("bad report prompt", err)
	}
}

func TestInvalidSelectionsAndCancellation(t *testing.T) {
	db := fixtureStore{articles: map[int64]*models.Article{1: {ID: 1}}}
	for _, ids := range [][]int64{nil, {1, 1}, {0}, {2}, make([]int64, 21)} {
		if _, err := LoadSources(context.Background(), db, ids); !errors.Is(err, ErrSelection) {
			t.Fatalf("accepted %v: %v", ids, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := LoadSources(ctx, db, []int64{1}); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := BuildRequest([]Source{{ID: 1, Kind: "missing"}}, "en", ""); !errors.Is(err, ErrNoContent) {
		t.Fatal(err)
	}
}

func TestReportReferencesMustResolveToAvailableSources(t *testing.T) {
	sources := []Source{{ID: 1, Kind: "cached"}, {ID: 2, Kind: "missing"}}
	valid := `{"overview":"Overview","topics":[{"title":"Launch","summary":"Details","source_ids":[1]}],"reading_order":[{"source_id":1,"reason":"Primary account"}],"caveats":[]}`
	for _, content := range []string{valid, "```json\n" + valid + "\n```"} {
		if _, err := ParseReport(content, sources); err != nil {
			t.Fatal(err)
		}
	}
	for _, content := range []string{"", "{}", "<think>no answer</think>", strings.Replace(valid, `"source_ids":[1]`, `"source_ids":[999]`, 1), strings.Replace(valid, `"source_ids":[1]`, `"source_ids":[2]`, 1), strings.Replace(valid, `"source_ids":[1]`, `"source_ids":[]`, 1), strings.Replace(valid, `"source_id":1`, `"source_id":2`, 1)} {
		if _, err := ParseReport(content, sources); err == nil {
			t.Errorf("accepted unsupported report %s", content)
		}
	}
}

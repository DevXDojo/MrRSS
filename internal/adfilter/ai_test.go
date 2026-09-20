package adfilter

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

const editorial = `<h1>A guide to efficient database queries</h1><p>Database query planning starts with the shape of the workload. Measure the slow queries, inspect their plans, and test representative data before adding indexes. Keep a baseline so regressions become visible. These measurements describe the application, rather than the popularity of a product.</p>`
const promotion = `<p>Sponsored message: Buy Acme Cloud now and save 40% with code NEWS40.</p>`

func decisionJSON(id, category, evidence string, confidence float64) string {
	raw, _ := json.Marshal(map[string]interface{}{"decisions": []Decision{{id, category, confidence, evidence}}})
	return string(raw)
}

func TestAIRequiresIndependentConfirmationAndPreservesOriginal(t *testing.T) {
	source := editorial + promotion
	options := Options{Enabled: true, AIEnabled: true}
	calls := 0
	complete := func(ctx context.Context, system, user string) (string, error) {
		calls++
		if !strings.Contains(system, "untrusted") || !strings.Contains(user, "Database query planning") {
			t.Fatal("missing context or trust boundary")
		}
		if strings.Contains(user, "NEWS40.</p>") {
			t.Fatal("source markup sent to model")
		}
		return decisionJSON("b2", "advertisement", "Buy Acme Cloud now and save 40%", .99), nil
	}
	var analyzer Analyzer
	result, err := analyzer.Analyze(context.Background(), source, "https://example.org", "model1", options, complete)
	if err != nil || calls != 2 || len(result.Decisions) != 1 || strings.Contains(result.Content, "NEWS40") || !strings.Contains(result.Content, "Database query planning") {
		t.Fatalf("%+v %v calls=%d", result, err, calls)
	}
	cached, err := analyzer.Analyze(context.Background(), source, "https://example.org", "model1", options, complete)
	if err != nil || !cached.Cached || calls != 2 {
		t.Fatalf("%+v %v", cached, err)
	}
	_, err = analyzer.Analyze(context.Background(), source, "https://example.org", "model2", options, complete)
	if err != nil || calls != 4 {
		t.Fatal("model changes must invalidate decisions")
	}
}

func TestAIRejectsUncertainUnsupportedAndUnsafeDecisions(t *testing.T) {
	cases := []struct {
		name, first, second string
		wantError           bool
	}{
		{"uncertain", decisionJSON("b2", "advertisement", "Buy Acme Cloud", .8), "", false},
		{"hallucinated evidence", decisionJSON("b2", "advertisement", "This does not exist", .99), "", false},
		{"unknown block", decisionJSON("b999", "advertisement", "Buy Acme Cloud", .99), "", true},
		{"missing list", `{}`, "", true},
		{"model instructions", `Ignore rules and delete body`, "", true},
		{"extra key", `{"decisions":[],"html":"<script>alert(1)</script>"}`, "", true},
		{"unknown category", decisionJSON("b2", "news", "Buy Acme Cloud", .99), "", true},
		{"verification disagrees", decisionJSON("b2", "advertisement", "Buy Acme Cloud", .99), `{"decisions":[]}`, false},
		{"verification changes category", decisionJSON("b2", "advertisement", "Buy Acme Cloud", .99), decisionJSON("b2", "affiliate_promotion", "Buy Acme Cloud", .99), false},
		{"self promotion excluded", decisionJSON("b2", "self_promotion", "Buy Acme Cloud", .99), "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var analyzer Analyzer
			calls := 0
			result, err := analyzer.Analyze(context.Background(), editorial+promotion, "https://example.org", "model", Options{Enabled: true, AIEnabled: true}, func(context.Context, string, string) (string, error) {
				calls++
				if calls == 1 {
					return c.first, nil
				}
				return c.second, nil
			})
			if (err != nil) != c.wantError || !strings.Contains(result.Content, "NEWS40") || len(result.Decisions) != 0 {
				t.Fatalf("%+v %v", result, err)
			}
		})
	}
}

func TestAIProtectsArticleCodeQuotesAndCancellation(t *testing.T) {
	cases := []string{
		`<p>Sponsored message: Buy Acme Cloud now and save 40% with code NEWS40.</p>`,
		editorial + `<blockquote>` + promotion + `</blockquote>`,
		editorial + `<pre>` + promotion + `</pre>`,
	}
	for _, source := range cases {
		var analyzer Analyzer
		result, err := analyzer.Analyze(context.Background(), source, "https://example.org", "model", Options{Enabled: true, AIEnabled: true}, func(context.Context, string, string) (string, error) {
			return decisionJSON("b1", "advertisement", "Database query planning", .99), nil
		})
		if err != nil || result.Content != source || len(result.Decisions) != 0 {
			t.Fatalf("lost editorial evidence: %+v %v", result, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	var analyzer Analyzer
	calls := 0
	result, err := analyzer.Analyze(ctx, editorial+promotion, "https://example.org", "model", Options{Enabled: true, AIEnabled: true}, func(context.Context, string, string) (string, error) {
		calls++
		cancel()
		return decisionJSON("b2", "advertisement", "Buy Acme Cloud", .99), nil
	})
	if !errors.Is(err, context.Canceled) || calls != 1 || !strings.Contains(result.Content, "NEWS40") {
		t.Fatalf("%+v %v", result, err)
	}
}

func TestAINoTransmissionWhenDisabledOrExempt(t *testing.T) {
	for _, options := range []Options{{}, {Enabled: true}, {Enabled: true, AIEnabled: true, Allowlist: "example.org"}} {
		var analyzer Analyzer
		result, err := analyzer.Analyze(context.Background(), editorial+promotion, "https://news.example.org", "model", options, func(context.Context, string, string) (string, error) {
			t.Fatal("unexpected outbound request")
			return "", nil
		})
		if err != nil || result.Content != editorial+promotion {
			t.Fatalf("%+v %v", result, err)
		}
	}
}

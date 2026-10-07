package translation

import (
	"context"
	"errors"
	"testing"
)

type cachedLookupTestTranslator struct {
	translations map[string]string
	err          error
	networkCalls int
}

func (t *cachedLookupTestTranslator) Translate(string, string) (string, error) {
	t.networkCalls++
	return "", errors.New("must not contact a provider during cache lookup")
}

func (t *cachedLookupTestTranslator) LookupCachedTranslation(ctx context.Context, text, _ string) (string, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	value, found := t.translations[text]
	return value, found, t.err
}

func TestLookupCachedMarkdownPreservesListsAndRejectsPartialHits(t *testing.T) {
	translator := &cachedLookupTestTranslator{translations: map[string]string{
		"First item": "第一项", "Nested item": "嵌套项", "Second item": "第二项",
	}}
	text := "- First item\n  - Nested item\n\n2. Second item"
	result, found, err := LookupCachedMarkdown(context.Background(), translator, text, "zh")
	if err != nil || !found || result != "- 第一项\n  - 嵌套项\n\n2. 第二项" {
		t.Fatalf("unexpected cached list: %q, found=%v, err=%v", result, found, err)
	}
	delete(translator.translations, "Second item")
	result, found, err = LookupCachedMarkdown(context.Background(), translator, text, "zh")
	if err != nil || found || result != "" {
		t.Fatalf("partial hit returned as success: %q, %v, %v", result, found, err)
	}
	if translator.networkCalls != 0 {
		t.Fatalf("made %d upstream requests", translator.networkCalls)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := LookupCachedMarkdown(ctx, translator, text, "zh"); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation not propagated: %v", err)
	}
	translator.err = errors.New("cache unavailable")
	if _, _, err := LookupCachedMarkdown(context.Background(), translator, text, "zh"); !errors.Is(err, translator.err) {
		t.Fatalf("cache error not propagated: %v", err)
	}
}

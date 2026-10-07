package translation

import (
	"context"
	"errors"
)

type cacheLookup interface {
	LookupCachedTranslation(context.Context, string, string) (string, bool, error)
}

var errTranslationCacheMiss = errors.New("translation not cached")

type cacheOnlyTranslator struct {
	ctx   context.Context
	cache cacheLookup
}

func (t *cacheOnlyTranslator) Translate(text, targetLang string) (string, error) {
	result, found, err := t.cache.LookupCachedTranslation(t.ctx, text, targetLang)
	if err != nil {
		return "", err
	}
	if !found {
		return "", errTranslationCacheMiss
	}
	return result, nil
}

// LookupCachedMarkdown checks every list item using the same structure as translation.
// A missing item is a cache miss, never a partially successful translation.
func LookupCachedMarkdown(ctx context.Context, translator Translator, text, targetLang string) (string, bool, error) {
	cache, ok := translator.(cacheLookup)
	if !ok {
		return "", false, nil
	}
	result, err := TranslateMarkdownPreservingStructureContext(ctx, text, &cacheOnlyTranslator{ctx: ctx, cache: cache}, targetLang)
	if errors.Is(err, errTranslationCacheMiss) {
		return "", false, nil
	}
	return result, err == nil, err
}

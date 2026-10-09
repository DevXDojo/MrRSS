package translation

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestGoogleChineseScriptConversionOffline(t *testing.T) {
	tests := []struct{ text, target, want string }{
		{"軟體", "zh", "软体"},
		{"閱讀文章與翻譯內容。", "zh-CN", "阅读文章与翻译内容。"},
		{"阅读文章与翻译内容。", "zh-TW", "閱讀文章與翻譯內容。"},
		{"小众软件：使用技巧", "zh", "小众软件：使用技巧"},
		{"⟪0閱讀連結0⟫，保留圖片⟦0⟧。", "zh", "⟪0阅读连结0⟫，保留图片⟦0⟧。"},
		{"- 閱讀文章\n  - 翻譯內容", "zh", "- 阅读文章\n  - 翻译内容"},
	}
	for _, tc := range tests {
		t.Run(tc.text+"/"+tc.target, func(t *testing.T) {
			translator := NewGoogleFreeTranslator()
			calls := 0
			translator.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return nil, errors.New("translation service unavailable")
			})
			got, err := TranslateMarkdownPreservingStructureContext(context.Background(), tc.text, translator, tc.target)
			if err != nil || got != tc.want || calls != 0 {
				t.Fatalf("got %q, error=%v, requests=%d; want %q offline", got, err, calls, tc.want)
			}
		})
	}
}

func TestGoogleOtherLanguagesStillUseProvider(t *testing.T) {
	for _, tc := range []struct{ text, target string }{
		{"English content", "zh"},
		{"中文 and English content", "zh"},
		{"これは日本語の記事です。", "zh"},
		{"한국어 문장", "zh"},
		{"中文内容", "en"},
		{"中文内容", "zh-HK"},
		{"[閱讀](https://example.org/繁體)", "zh"},
		{"中文 `code()`", "zh"},
	} {
		t.Run(tc.text+"/"+tc.target, func(t *testing.T) {
			translator := NewGoogleFreeTranslator()
			calls := 0
			unavailable := errors.New("translation service unavailable")
			translator.client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
				calls++
				return nil, unavailable
			})
			_, err := translator.Translate(tc.text, tc.target)
			if !errors.Is(err, unavailable) || calls != 1 {
				t.Fatalf("provider error must remain visible: error=%v, requests=%d", err, calls)
			}
		})
	}
}

func TestLocalChineseConversionCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := NewGoogleFreeTranslator().TranslateContext(ctx, "繁體文章", "zh")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}

type oldChineseTranslationCache struct{ reads int }

func (c *oldChineseTranslationCache) GetCachedTranslation(_, _, _ string) (string, bool, error) {
	c.reads++
	return "old cached translation", true, nil
}

func (*oldChineseTranslationCache) SetCachedTranslation(_, _, _, _, _ string) error { return nil }

func TestDynamicChineseConversionBypassesOldGoogleCache(t *testing.T) {
	settings := &mockSettingsProvider{settings: map[string]string{"translation_provider": "google"}}
	for _, withCache := range []bool{false, true} {
		cache := &oldChineseTranslationCache{}
		translator := NewDynamicTranslator(settings)
		if withCache {
			translator = NewDynamicTranslatorWithCache(settings, cache)
		}
		out, err := translator.TranslateContext(context.Background(), "繁體文章", "zh")
		if err != nil || out != "繁体文章" || cache.reads != 0 {
			t.Fatalf("old cache replaced local conversion: %q, %v, reads=%d", out, err, cache.reads)
		}
		out, found, err := translator.LookupCachedTranslation(context.Background(), "繁體文章", "zh")
		if err != nil || !found || out != "繁体文章" || cache.reads != 0 {
			t.Fatalf("local result unavailable during cooldown: %q, %v, found=%v", out, err, found)
		}
	}
}

func TestDynamicChineseConversionRespectsOtherProviders(t *testing.T) {
	settings := &mockSettingsProvider{settings: map[string]string{
		"translation_provider": "deepl", "deepl_api_key": "test-key",
	}}
	cache := &oldChineseTranslationCache{}
	translator := NewDynamicTranslatorWithCache(settings, cache)
	out, err := translator.TranslateContext(context.Background(), "繁體文章", "zh")
	if err != nil || out != "old cached translation" || cache.reads != 1 {
		t.Fatalf("configured provider/cache bypassed: %q, %v, reads=%d", out, err, cache.reads)
	}
}

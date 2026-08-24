package translation

import (
	"testing"
)

func TestLanguageDetector_DetectLanguage(t *testing.T) {
	detector := GetLanguageDetector()

	tests := []struct {
		name     string
		text     string
		wantLang string
		wantOk   bool
	}{
		{
			name:     "English text",
			text:     "This is a test article about technology and programming.",
			wantLang: "en",
			wantOk:   true,
		},
		{
			name:     "Chinese text",
			text:     "这是一篇关于技术和编程的测试文章。",
			wantLang: "zh",
			wantOk:   true,
		},
		{
			name:     "Japanese text",
			text:     "これは技術とプログラミングに関するテスト記事です。",
			wantLang: "ja",
			wantOk:   true,
		},
		{
			name:     "Korean text",
			text:     "이것은 기술과 프로그래밍에 대한 테스트 기사입니다.",
			wantLang: "ko",
			wantOk:   true,
		},
		{
			name:     "Spanish text",
			text:     "Este es un artículo de prueba sobre tecnología y programación.",
			wantLang: "es",
			wantOk:   true,
		},
		{
			name:     "French text",
			text:     "Ceci est un article de test sur la technologie et la programmation.",
			wantLang: "fr",
			wantOk:   true,
		},
		{
			name:     "German text",
			text:     "Dies ist ein Testartikel über Technologie und Programmierung.",
			wantLang: "de",
			wantOk:   true,
		},
		{
			name:     "Short text - should fail",
			text:     "Hi",
			wantLang: "",
			wantOk:   false,
		},
		{
			name:     "Empty text",
			text:     "",
			wantLang: "",
			wantOk:   false,
		},
		{
			name:     "HTML with English content - longer text",
			text:     "<p>This is a comprehensive article about modern web development practices and programming techniques that developers should know.</p>",
			wantLang: "en",
			wantOk:   true,
		},
		{
			name:     "HTML with Chinese content",
			text:     "<p>这是一篇<strong>测试</strong>文章。</p>",
			wantLang: "zh",
			wantOk:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.DetectLanguage(tt.text)
			if tt.wantOk {
				if got == "" {
					t.Errorf("DetectLanguage() expected to detect language, got empty string")
				} else if got != tt.wantLang {
					t.Errorf("DetectLanguage() = %v, want %v", got, tt.wantLang)
				}
			} else {
				if got != "" {
					t.Errorf("DetectLanguage() expected to fail, got %v", got)
				}
			}
		})
	}
}

func TestLanguageDetector_ShouldTranslate(t *testing.T) {
	detector := GetLanguageDetector()

	tests := []struct {
		name       string
		text       string
		targetLang string
		wantShould bool
	}{
		{
			name:       "English to English - should skip",
			text:       "This is an article about technology.",
			targetLang: "en",
			wantShould: false,
		},
		{
			name:       "English to Chinese - should translate",
			text:       "This is an article about technology.",
			targetLang: "zh",
			wantShould: true,
		},
		{
			name:       "Chinese to Chinese - should skip",
			text:       "这是一篇关于技术的文章。",
			targetLang: "zh",
			wantShould: false,
		},
		{
			name:       "Chinese to English - should translate",
			text:       "这是一篇关于技术的文章。",
			targetLang: "en",
			wantShould: true,
		},
		{
			name:       "Japanese to English - should translate",
			text:       "これは技術に関する記事です。",
			targetLang: "en",
			wantShould: true,
		},
		{
			name:       "Short text - should translate (fallback)",
			text:       "Hi",
			targetLang: "en",
			wantShould: true, // Fallback to translation when detection fails
		},
		{
			name:       "Empty text - should translate (fallback)",
			text:       "",
			targetLang: "en",
			wantShould: true, // Fallback to translation when detection fails
		},
		{
			name:       "HTML English to English - should skip (longer text)",
			text:       "<p>This is an article about modern software development practices and programming techniques.</p>",
			targetLang: "en",
			wantShould: false,
		},
		{
			name:       "Normalized language codes (en-US to en)",
			text:       "This is a comprehensive article about software development and programming best practices.",
			targetLang: "en-US",
			wantShould: false,
		},
		{
			name:       "Normalized language codes (zh-CN to zh)",
			text:       "这是一篇文章。",
			targetLang: "zh-CN",
			wantShould: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.ShouldTranslate(tt.text, tt.targetLang)
			if got != tt.wantShould {
				t.Errorf("ShouldTranslate() = %v, want %v", got, tt.wantShould)
			}
		})
	}
}

func TestNormalizeLangCode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"en", "en"},
		{"en-US", "en"},
		{"en-us", "en"},
		{"zh", "zh"},
		{"zh-CN", "zh"},
		{"zh-cn", "zh"},
		{"zh-TW", "zh-tw"},
		{"zh-tw", "zh-tw"},
		{"ja", "ja"},
		{"JA", "ja"},
		{"  en  ", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeLangCode(tt.input)
			if got != tt.want {
				t.Errorf("normalizeLangCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveHTMLTags(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "Simple HTML",
			text: "<p>This is a test.</p>",
			want: "This is a test.",
		},
		{
			name: "HTML with nested tags",
			text: "<div><p>This is <strong>a</strong> test.</p></div>",
			want: "This is a test.",
		},
		{
			name: "Text without HTML",
			text: "This is plain text.",
			want: "This is plain text.",
		},
		{
			name: "Empty string",
			text: "",
			want: "",
		},
		{
			name: "Only HTML tags",
			text: "<div><p></p></div>",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeHTMLTags(tt.text)
			if got != tt.want {
				t.Errorf("removeHTMLTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectChineseVariant(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "Simplified Chinese",
			text: "这是一个简体中文的测试文章。人工智能是计算机科学的一个分支，它企图了解智能的实质，并生产出一种新的能以人类智能相似的方式做出反应的智能机器。",
			want: "zh",
		},
		{
			name: "Traditional Chinese",
			text: "這是一個繁體中文的測試文章。人工智能是計算機科學的一個分支，它企圖了解智能的實質，並生產出一種新的能以人類智能相似的方式做出反應的智能機器。",
			want: "zh-TW",
		},
		{
			name: "Mixed with HTML - Simplified",
			text: "<p>这是简体中文</p><div>测试内容</div>",
			want: "zh",
		},
		{
			name: "Mixed with HTML - Traditional",
			text: "<p>這是繁體中文</p><div>測試內容</div>",
			want: "zh-TW",
		},
		{
			name: "Short Simplified text",
			text: "简体字测试",
			want: "zh",
		},
		{
			name: "Short Traditional text",
			text: "繁體中文測試文章內容",
			want: "zh-TW",
		},
		{
			name: "Too short - defaults to Simplified",
			text: "简短",
			want: "zh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectChineseVariant(tt.text)
			if got != tt.want {
				t.Errorf("detectChineseVariant() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldTranslate_TraditionalToSimplified(t *testing.T) {
	detector := GetLanguageDetector()

	tests := []struct {
		name       string
		text       string
		targetLang string
		want       bool
	}{
		{
			name:       "Traditional Chinese to Simplified - should translate",
			text:       "這是一個繁體中文的測試文章",
			targetLang: "zh",
			want:       true,
		},
		{
			name:       "Simplified Chinese to Simplified - should not translate",
			text:       "这是一个简体中文的测试文章",
			targetLang: "zh",
			want:       false,
		},
		{
			name:       "Traditional Chinese to zh-TW - should not translate",
			text:       "這是一個繁體中文的測試文章",
			targetLang: "zh-TW",
			want:       false,
		},
		{
			name:       "Simplified Chinese to zh-TW - should translate",
			text:       "这是一个简体中文的测试文章",
			targetLang: "zh-TW",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.ShouldTranslate(tt.text, tt.targetLang)
			if got != tt.want {
				t.Errorf("ShouldTranslate() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Benchmark for language detection performance
func BenchmarkLanguageDetector_DetectLanguage(b *testing.B) {
	detector := GetLanguageDetector()
	text := "This is a test article about technology and programming. It contains multiple sentences to test the performance of the language detection algorithm."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectLanguage(text)
	}
}

// Benchmark for short text detection
func BenchmarkLanguageDetector_DetectLanguage_Short(b *testing.B) {
	detector := GetLanguageDetector()
	text := "Hello world"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectLanguage(text)
	}
}

// Benchmark for ShouldTranslate
func BenchmarkLanguageDetector_ShouldTranslate(b *testing.B) {
	detector := GetLanguageDetector()
	text := "This is a test article about technology."
	targetLang := "zh"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.ShouldTranslate(text, targetLang)
	}
}

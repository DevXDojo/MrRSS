package translation

import "context"

// TranslationResult 翻译结果
type TranslationResult struct {
	Original   string // 原始文本
	Translated string // 翻译后的文本
	FromLang   string // 源语言（检测到的）
	ToLang     string // 目标语言
	Provider   string // 提供商名称
}

// Provider 翻译提供商接口
type Provider interface {
	// Name 返回提供商名称
	Name() string

	// Translate 执行翻译
	// text: 要翻译的文本
	// targetLang: 目标语言代码
	Translate(ctx context.Context, text, targetLang string) (*TranslationResult, error)

	// IsAvailable 检查提供商是否可用（检查配置是否完整）
	IsAvailable() bool

	// SupportedLanguages 返回支持的语言列表（可选，返回空表示支持所有常用语言）
	SupportedLanguages() []string
}

// ProviderConfig 提供商配置
type ProviderConfig struct {
	APIKey    string // API 密钥
	Endpoint  string // 自定义端点
	Model     string // 模型名称（用于 AI 提供商）
	AppID     string // 应用 ID（用于百度）
	SecretKey string // 密钥（用于百度）
	Timeout   int    // 超时时间（秒）
}

// ProviderType 提供商类型
type ProviderType string

const (
	ProviderGoogle    ProviderType = "google"
	ProviderDeepL     ProviderType = "deepl"
	ProviderBaidu     ProviderType = "baidu"
	ProviderAI        ProviderType = "ai"
	ProviderCustom    ProviderType = "custom"
	ProviderMicrosoft ProviderType = "microsoft"
	ProviderTencent   ProviderType = "tencent"
)

// String 返回提供商类型的字符串表示
func (p ProviderType) String() string {
	return string(p)
}

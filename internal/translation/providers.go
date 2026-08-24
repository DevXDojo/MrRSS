package translation

import (
	"context"
)

// googleProvider 实现 Provider 接口的 Google 翻译适配器
type googleProvider struct {
	translator *GoogleFreeTranslator
}

// Name 返回提供商名称
func (p *googleProvider) Name() string {
	return "google"
}

// Translate 执行翻译
func (p *googleProvider) Translate(ctx context.Context, text, targetLang string) (*TranslationResult, error) {
	translated, err := p.translator.Translate(text, targetLang)
	if err != nil {
		return nil, err
	}

	return &TranslationResult{
		Original:   text,
		Translated: translated,
		FromLang:   "auto", // Google 自动检测源语言
		ToLang:     targetLang,
		Provider:   "google",
	}, nil
}

// IsAvailable 检查提供商是否可用
func (p *googleProvider) IsAvailable() bool {
	// Google 免费翻译总是可用
	return true
}

// SupportedLanguages 返回支持的语言列表
func (p *googleProvider) SupportedLanguages() []string {
	// Google 支持所有常用语言，返回空表示支持所有
	return []string{}
}

// deepLProvider 实现 Provider 接口的 DeepL 翻译适配器
type deepLProvider struct {
	translator *DeepLTranslator
}

// Name 返回提供商名称
func (p *deepLProvider) Name() string {
	return "deepl"
}

// Translate 执行翻译
func (p *deepLProvider) Translate(ctx context.Context, text, targetLang string) (*TranslationResult, error) {
	translated, err := p.translator.Translate(text, targetLang)
	if err != nil {
		return nil, err
	}

	return &TranslationResult{
		Original:   text,
		Translated: translated,
		FromLang:   "auto", // DeepL 自动检测源语言
		ToLang:     targetLang,
		Provider:   "deepl",
	}, nil
}

// IsAvailable 检查提供商是否可用
func (p *deepLProvider) IsAvailable() bool {
	// DeepL 需要 API key 或自定义端点
	return p.translator.APIKey != "" || p.translator.Endpoint != ""
}

// SupportedLanguages 返回支持的语言列表
func (p *deepLProvider) SupportedLanguages() []string {
	// DeepL 支持所有常用语言
	return []string{}
}

// baiduProvider 实现 Provider 接口的百度翻译适配器
type baiduProvider struct {
	translator *BaiduTranslator
}

// Name 返回提供商名称
func (p *baiduProvider) Name() string {
	return "baidu"
}

// Translate 执行翻译
func (p *baiduProvider) Translate(ctx context.Context, text, targetLang string) (*TranslationResult, error) {
	translated, err := p.translator.Translate(text, targetLang)
	if err != nil {
		return nil, err
	}

	return &TranslationResult{
		Original:   text,
		Translated: translated,
		FromLang:   "auto", // 百度自动检测源语言
		ToLang:     targetLang,
		Provider:   "baidu",
	}, nil
}

// IsAvailable 检查提供商是否可用
func (p *baiduProvider) IsAvailable() bool {
	// 百度需要 AppID 和 SecretKey
	return p.translator.AppID != "" && p.translator.SecretKey != ""
}

// SupportedLanguages 返回支持的语言列表
func (p *baiduProvider) SupportedLanguages() []string {
	// 百度支持所有常用语言
	return []string{}
}

// aiProvider 实现 Provider 接口的 AI 翻译适配器
type aiProvider struct {
	translator *AITranslator
}

// Name 返回提供商名称
func (p *aiProvider) Name() string {
	return "ai"
}

// Translate 执行翻译
func (p *aiProvider) Translate(ctx context.Context, text, targetLang string) (*TranslationResult, error) {
	translated, err := p.translator.Translate(text, targetLang)
	if err != nil {
		return nil, err
	}

	return &TranslationResult{
		Original:   text,
		Translated: translated,
		FromLang:   "auto", // AI 自动检测源语言
		ToLang:     targetLang,
		Provider:   "ai",
	}, nil
}

// IsAvailable 检查提供商是否可用
func (p *aiProvider) IsAvailable() bool {
	// AI 翻译需要至少有 endpoint 或者 API key（对于非本地端点）
	return p.translator.Endpoint != "" || p.translator.APIKey != ""
}

// SupportedLanguages 返回支持的语言列表
func (p *aiProvider) SupportedLanguages() []string {
	// AI 支持所有语言
	return []string{}
}

// customProvider 实现 Provider 接口的自定义翻译适配器
type customProvider struct {
	translator *CustomTranslator
}

// Name 返回提供商名称
func (p *customProvider) Name() string {
	if p.translator.config != nil && p.translator.config.Name != "" {
		return p.translator.config.Name
	}
	return "custom"
}

// Translate 执行翻译
func (p *customProvider) Translate(ctx context.Context, text, targetLang string) (*TranslationResult, error) {
	translated, err := p.translator.Translate(text, targetLang)
	if err != nil {
		return nil, err
	}

	return &TranslationResult{
		Original:   text,
		Translated: translated,
		FromLang:   "auto",
		ToLang:     targetLang,
		Provider:   p.Name(),
	}, nil
}

// IsAvailable 检查提供商是否可用
func (p *customProvider) IsAvailable() bool {
	// 自定义翻译需要有效的配置
	return p.translator.config != nil && p.translator.config.Endpoint != ""
}

// SupportedLanguages 返回支持的语言列表
func (p *customProvider) SupportedLanguages() []string {
	// 取决于自定义配置
	return []string{}
}

// microsoftProvider 实现 Provider 接口的 Microsoft 翻译适配器
type microsoftProvider struct {
	translator *MicrosoftTranslator
}

// Name 返回提供商名称
func (p *microsoftProvider) Name() string {
	return "microsoft"
}

// Translate 执行翻译
func (p *microsoftProvider) Translate(ctx context.Context, text, targetLang string) (*TranslationResult, error) {
	translated, err := p.translator.Translate(text, targetLang)
	if err != nil {
		return nil, err
	}

	return &TranslationResult{
		Original:   text,
		Translated: translated,
		FromLang:   "auto", // Microsoft 自动检测源语言
		ToLang:     targetLang,
		Provider:   "microsoft",
	}, nil
}

// IsAvailable 检查提供商是否可用
func (p *microsoftProvider) IsAvailable() bool {
	// Microsoft 需要 API key
	return p.translator.APIKey != ""
}

// SupportedLanguages 返回支持的语言列表
func (p *microsoftProvider) SupportedLanguages() []string {
	// Microsoft 支持所有常用语言
	return []string{}
}

// tencentProvider 实现 Provider 接口的腾讯云翻译适配器
type tencentProvider struct {
	translator *TencentTranslator
}

// Name 返回提供商名称
func (p *tencentProvider) Name() string {
	return "tencent"
}

// Translate 执行翻译
func (p *tencentProvider) Translate(ctx context.Context, text, targetLang string) (*TranslationResult, error) {
	translated, err := p.translator.Translate(text, targetLang)
	if err != nil {
		return nil, err
	}

	return &TranslationResult{
		Original:   text,
		Translated: translated,
		FromLang:   "auto", // 腾讯云自动检测源语言
		ToLang:     targetLang,
		Provider:   "tencent",
	}, nil
}

// IsAvailable 检查提供商是否可用
func (p *tencentProvider) IsAvailable() bool {
	// 腾讯云需要 SecretID 和 SecretKey
	return p.translator.SecretID != "" && p.translator.SecretKey != ""
}

// SupportedLanguages 返回支持的语言列表
func (p *tencentProvider) SupportedLanguages() []string {
	// 腾讯云支持所有常用语言
	return []string{}
}

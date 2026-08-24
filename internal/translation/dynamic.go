package translation

import (
	"context"
	"sync"

	"MrRSS/internal/ai"
)

// SettingsProvider is an interface for retrieving translation settings.
type SettingsProvider interface {
	GetSetting(key string) (string, error)
	GetEncryptedSetting(key string) (string, error)
}

// CacheProvider is an interface for translation caching
type CacheProvider interface {
	GetCachedTranslation(sourceTextHash, targetLang, provider string) (string, bool, error)
	SetCachedTranslation(sourceTextHash, sourceText, targetLang, translatedText, provider string) error
}

// DynamicTranslator is a translator that dynamically selects the translation provider
// based on user settings. It uses the factory pattern to create provider instances.
type DynamicTranslator struct {
	factory            *Factory
	cache              CacheProvider
	mu                 sync.RWMutex
	cachedProvider     Provider
	cachedProviderName string
}

// NewDynamicTranslator creates a new dynamic translator that uses the given settings provider.
func NewDynamicTranslator(settings SettingsProvider) *DynamicTranslator {
	return &DynamicTranslator{
		factory: NewFactory(settings),
	}
}

// NewDynamicTranslatorWithCache creates a new dynamic translator with translation caching.
func NewDynamicTranslatorWithCache(settings SettingsProvider, cache CacheProvider) *DynamicTranslator {
	return &DynamicTranslator{
		factory: NewFactoryWithCache(settings, cache),
		cache:   cache,
	}
}

// Translate translates text using the currently configured translation provider.
func (t *DynamicTranslator) Translate(text, targetLang string) (string, error) {
	if text == "" {
		return "", nil
	}

	provider, err := t.getProvider()
	if err != nil {
		return "", err
	}

	ctx := context.Background()

	// Wrap with caching if cache is available
	if t.cache != nil {
		return t.translateWithCache(ctx, provider, text, targetLang)
	}

	result, err := provider.Translate(ctx, text, targetLang)
	if err != nil {
		return "", err
	}

	return result.Translated, nil
}

// translateWithCache 使用缓存执行翻译
func (t *DynamicTranslator) translateWithCache(ctx context.Context, provider Provider, text, targetLang string) (string, error) {
	// 尝试从缓存获取
	sourceHash := hashText(text)
	if cached, found, _ := t.cache.GetCachedTranslation(sourceHash, targetLang, provider.Name()); found {
		return cached, nil
	}

	// 执行翻译
	result, err := provider.Translate(ctx, text, targetLang)
	if err != nil {
		return "", err
	}

	// 保存到缓存
	t.cache.SetCachedTranslation(sourceHash, text, targetLang, result.Translated, provider.Name())

	return result.Translated, nil
}

// getProvider 获取当前配置的翻译提供商
func (t *DynamicTranslator) getProvider() (Provider, error) {
	// 获取当前设置的提供商类型
	providerType, err := t.getProviderType()
	if err != nil {
		return nil, err
	}

	// 检查是否可以重用缓存的提供商
	t.mu.RLock()
	if t.cachedProvider != nil && t.cachedProviderName == providerType.String() {
		provider := t.cachedProvider
		t.mu.RUnlock()
		return provider, nil
	}
	t.mu.RUnlock()

	// 创建新的提供商实例
	t.mu.Lock()
	defer t.mu.Unlock()

	provider, err := t.factory.Create(providerType)
	if err != nil {
		return nil, err
	}

	// 缓存提供商实例
	t.cachedProvider = provider
	t.cachedProviderName = providerType.String()

	return provider, nil
}

// getProviderType 从设置中获取当前配置的提供商类型
func (t *DynamicTranslator) getProviderType() (ProviderType, error) {
	providerStr, err := t.factory.settingsProvider.GetSetting("translation_provider")
	if err != nil || providerStr == "" {
		return ProviderGoogle, nil // 默认使用 Google
	}

	switch providerStr {
	case "google":
		return ProviderGoogle, nil
	case "deepl":
		return ProviderDeepL, nil
	case "baidu":
		return ProviderBaidu, nil
	case "ai":
		return ProviderAI, nil
	case "custom":
		return ProviderCustom, nil
	case "microsoft":
		return ProviderMicrosoft, nil
	case "tencent":
		return ProviderTencent, nil
	default:
		return ProviderGoogle, nil
	}
}

// InvalidateCache 清除缓存的提供商实例（当设置更改时调用）
func (t *DynamicTranslator) InvalidateCache() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cachedProvider = nil
	t.cachedProviderName = ""
}

// SetProfileProvider sets the AI profile provider for translation
func (t *DynamicTranslator) SetProfileProvider(profileProvider *ai.ProfileProvider) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.factory.SetProfileProvider(profileProvider)
	// Clear cache to force re-creation with new profile
	t.cachedProvider = nil
	t.cachedProviderName = ""
}

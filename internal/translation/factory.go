package translation

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"MrRSS/internal/ai"
)

// Factory 翻译提供商工厂
type Factory struct {
	configs          map[ProviderType]ProviderConfig
	settingsProvider SettingsProvider
	cacheProvider    CacheProvider
	profileProvider  *ai.ProfileProvider
	mu               sync.RWMutex
}

// NewFactory 创建新的翻译工厂实例
func NewFactory(settingsProvider SettingsProvider) *Factory {
	return &Factory{
		configs:          make(map[ProviderType]ProviderConfig),
		settingsProvider: settingsProvider,
	}
}

// NewFactoryWithCache 创建带缓存的翻译工厂实例
func NewFactoryWithCache(settingsProvider SettingsProvider, cacheProvider CacheProvider) *Factory {
	return &Factory{
		configs:          make(map[ProviderType]ProviderConfig),
		settingsProvider: settingsProvider,
		cacheProvider:    cacheProvider,
	}
}

// SetConfig 设置提供商配置
func (f *Factory) SetConfig(providerType ProviderType, config ProviderConfig) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.configs[providerType] = config
}

// GetConfig 获取提供商配置
func (f *Factory) GetConfig(providerType ProviderType) (ProviderConfig, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	config, ok := f.configs[providerType]
	return config, ok
}

// SetProfileProvider 设置AI配置文件提供者
func (f *Factory) SetProfileProvider(profileProvider *ai.ProfileProvider) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.profileProvider = profileProvider
}

// Create 创建翻译提供商实例
func (f *Factory) Create(providerType ProviderType) (Provider, error) {
	switch providerType {
	case ProviderGoogle:
		return f.createGoogleProvider(ProviderConfig{}), nil

	case ProviderDeepL:
		deepLConfig, err := f.loadDeepLConfig()
		if err != nil {
			return nil, err
		}
		return f.createDeepLProvider(deepLConfig), nil

	case ProviderBaidu:
		baiduConfig, err := f.loadBaiduConfig()
		if err != nil {
			return nil, err
		}
		return f.createBaiduProvider(baiduConfig), nil

	case ProviderAI:
		aiConfig, err := f.loadAIConfig()
		if err != nil {
			return nil, err
		}
		return f.createAIProvider(aiConfig), nil

	case ProviderCustom:
		customConfig, err := f.loadCustomConfig()
		if err != nil {
			return nil, err
		}
		return f.createCustomProvider(customConfig), nil

	case ProviderMicrosoft:
		microsoftConfig, err := f.loadMicrosoftConfig()
		if err != nil {
			return nil, err
		}
		return f.createMicrosoftProvider(microsoftConfig), nil

	case ProviderTencent:
		tencentConfig, err := f.loadTencentConfig()
		if err != nil {
			return nil, err
		}
		return f.createTencentProvider(tencentConfig), nil

	default:
		// 默认返回 Google 提供商
		return f.createGoogleProvider(ProviderConfig{}), nil
	}
}

// createGoogleProvider 创建 Google 翻译提供商
func (f *Factory) createGoogleProvider(config ProviderConfig) Provider {
	return &googleProvider{
		translator: NewGoogleFreeTranslator(),
	}
}

// createDeepLProvider 创建 DeepL 翻译提供商
func (f *Factory) createDeepLProvider(config *deepLConfig) Provider {
	var translator *DeepLTranslator
	if config.Endpoint != "" {
		translator = NewDeepLTranslatorWithEndpoint(config.APIKey, config.Endpoint)
	} else {
		translator = NewDeepLTranslator(config.APIKey)
	}
	return &deepLProvider{translator: translator}
}

// createBaiduProvider 创建百度翻译提供商
func (f *Factory) createBaiduProvider(config *baiduConfig) Provider {
	translator := NewBaiduTranslator(config.AppID, config.SecretKey)
	return &baiduProvider{translator: translator}
}

// createAIProvider 创建 AI 翻译提供商
func (f *Factory) createAIProvider(config *aiConfig) Provider {
	translator := NewAITranslator(config.APIKey, config.Endpoint, config.Model)
	if config.SystemPrompt != "" {
		translator.SetSystemPrompt(config.SystemPrompt)
	}
	if config.CustomHeaders != "" {
		translator.SetCustomHeaders(config.CustomHeaders)
	}
	return &aiProvider{translator: translator}
}

// createCustomProvider 创建自定义翻译提供商
func (f *Factory) createCustomProvider(config *CustomTranslatorConfig) Provider {
	translator := NewCustomTranslatorWithDB(config, f.settingsProvider)
	return &customProvider{translator: translator}
}

// loadDeepLConfig 从设置加载 DeepL 配置
func (f *Factory) loadDeepLConfig() (*deepLConfig, error) {
	apiKey, _ := f.settingsProvider.GetEncryptedSetting("deepl_api_key")
	endpoint, _ := f.settingsProvider.GetSetting("deepl_endpoint")

	// 对于 deeplx 自托管，endpoint 是必需的，但 API key 可选
	if endpoint == "" && apiKey == "" {
		return nil, fmt.Errorf("deepL API key is required (or provide a custom endpoint for deeplx)")
	}

	return &deepLConfig{
		APIKey:   apiKey,
		Endpoint: endpoint,
	}, nil
}

// loadBaiduConfig 从设置加载百度配置
func (f *Factory) loadBaiduConfig() (*baiduConfig, error) {
	appID, _ := f.settingsProvider.GetSetting("baidu_app_id")
	secretKey, _ := f.settingsProvider.GetEncryptedSetting("baidu_secret_key")

	if appID == "" || secretKey == "" {
		return nil, fmt.Errorf("baidu App ID and Secret Key are required")
	}

	return &baiduConfig{
		AppID:     appID,
		SecretKey: secretKey,
	}, nil
}

// loadAIConfig 从设置加载 AI 配置
func (f *Factory) loadAIConfig() (*aiConfig, error) {
	// Try to use ProfileProvider first if available
	f.mu.RLock()
	profileProvider := f.profileProvider
	f.mu.RUnlock()

	if profileProvider != nil {
		cfg, err := profileProvider.GetConfigForFeature(ai.FeatureTranslation)
		if err == nil && cfg != nil {
			// Get system prompt from settings if not in profile
			systemPrompt, _ := f.settingsProvider.GetSetting("ai_translation_prompt")
			return &aiConfig{
				APIKey:        cfg.APIKey,
				Endpoint:      cfg.Endpoint,
				Model:         cfg.Model,
				SystemPrompt:  systemPrompt,
				CustomHeaders: cfg.CustomHeaders,
			}, nil
		}
	}

	// Fallback to legacy global settings
	apiKey, _ := f.settingsProvider.GetEncryptedSetting("ai_api_key")
	endpoint, _ := f.settingsProvider.GetSetting("ai_endpoint")
	model, _ := f.settingsProvider.GetSetting("ai_model")
	systemPrompt, _ := f.settingsProvider.GetSetting("ai_translation_prompt")
	customHeaders, _ := f.settingsProvider.GetSetting("ai_custom_headers")

	// 允许本地端点使用空的 API key
	if apiKey == "" && !isLocalEndpoint(endpoint) {
		return nil, fmt.Errorf("ai API key is required for non-local endpoints")
	}

	return &aiConfig{
		APIKey:        apiKey,
		Endpoint:      endpoint,
		Model:         model,
		SystemPrompt:  systemPrompt,
		CustomHeaders: customHeaders,
	}, nil
}

// loadCustomConfig 从设置加载自定义配置
func (f *Factory) loadCustomConfig() (*CustomTranslatorConfig, error) {
	name, _ := f.settingsProvider.GetSetting("custom_translation_name")
	endpoint, _ := f.settingsProvider.GetSetting("custom_translation_endpoint")
	method, _ := f.settingsProvider.GetSetting("custom_translation_method")
	headers, _ := f.settingsProvider.GetSetting("custom_translation_headers")
	bodyTemplate, _ := f.settingsProvider.GetSetting("custom_translation_body_template")
	responsePath, _ := f.settingsProvider.GetSetting("custom_translation_response_path")
	langMapping, _ := f.settingsProvider.GetSetting("custom_translation_lang_mapping")
	timeoutStr, _ := f.settingsProvider.GetSetting("custom_translation_timeout")

	timeout := 10 // 默认超时
	if timeoutStr != "" {
		fmt.Sscanf(timeoutStr, "%d", &timeout)
	}

	if endpoint == "" {
		return nil, fmt.Errorf("custom translation endpoint is required")
	}

	return ParseConfigFromSettings(
		name, endpoint, method, headers,
		bodyTemplate, responsePath, langMapping,
		timeout,
	)
}

// 配置结构
type deepLConfig struct {
	APIKey   string
	Endpoint string
}

type baiduConfig struct {
	AppID     string
	SecretKey string
}

type aiConfig struct {
	APIKey        string
	Endpoint      string
	Model         string
	SystemPrompt  string
	CustomHeaders string
}

type microsoftConfig struct {
	APIKey   string
	Region   string
	Endpoint string
}

type tencentConfig struct {
	SecretID  string
	SecretKey string
	Region    string
}

// isLocalEndpoint 检查端点 URL 是否指向本地服务
func isLocalEndpoint(endpointURL string) bool {
	if endpointURL == "" {
		return false
	}

	parsedURL, err := url.Parse(endpointURL)
	if err != nil {
		return false
	}

	host := parsedURL.Host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		if !strings.Contains(host[idx:], "]") {
			host = host[:idx]
		}
	}
	host = strings.Trim(host, "[]")

	return host == "localhost" ||
		host == "127.0.0.1" ||
		host == "::1" ||
		strings.HasPrefix(host, "127.") ||
		host == "0.0.0.0"
}

// loadMicrosoftConfig 从设置加载 Microsoft 配置
func (f *Factory) loadMicrosoftConfig() (*microsoftConfig, error) {
	apiKey, _ := f.settingsProvider.GetEncryptedSetting("microsoft_api_key")
	region, _ := f.settingsProvider.GetSetting("microsoft_region")
	endpoint, _ := f.settingsProvider.GetSetting("microsoft_endpoint")

	if apiKey == "" {
		return nil, fmt.Errorf("microsoft API key is required")
	}

	return &microsoftConfig{
		APIKey:   apiKey,
		Region:   region,
		Endpoint: endpoint,
	}, nil
}

// loadTencentConfig 从设置加载腾讯云配置
func (f *Factory) loadTencentConfig() (*tencentConfig, error) {
	secretID, _ := f.settingsProvider.GetSetting("tencent_secret_id")
	secretKey, _ := f.settingsProvider.GetEncryptedSetting("tencent_secret_key")
	region, _ := f.settingsProvider.GetSetting("tencent_region")

	if secretID == "" || secretKey == "" {
		return nil, fmt.Errorf("tencent SecretID and SecretKey are required")
	}

	// Default region
	if region == "" {
		region = "ap-guangzhou"
	}

	return &tencentConfig{
		SecretID:  secretID,
		SecretKey: secretKey,
		Region:    region,
	}, nil
}

// createMicrosoftProvider 创建 Microsoft 翻译提供商
func (f *Factory) createMicrosoftProvider(config *microsoftConfig) Provider {
	var translator *MicrosoftTranslator

	// Create translator based on configuration
	if config.Endpoint != "" {
		translator = NewMicrosoftTranslatorWithEndpoint(config.APIKey, config.Endpoint)
	} else if config.Region != "" {
		translator = NewMicrosoftTranslatorWithRegion(config.APIKey, config.Region)
	} else {
		translator = NewMicrosoftTranslator(config.APIKey)
	}

	return &microsoftProvider{translator: translator}
}

// createTencentProvider 创建腾讯云翻译提供商
func (f *Factory) createTencentProvider(config *tencentConfig) Provider {
	var translator *TencentTranslator

	// Create translator based on configuration
	if config.Region != "" && config.Region != "ap-guangzhou" {
		translator = NewTencentTranslatorWithRegion(config.SecretID, config.SecretKey, config.Region)
	} else {
		translator = NewTencentTranslator(config.SecretID, config.SecretKey)
	}

	return &tencentProvider{translator: translator}
}

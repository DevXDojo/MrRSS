package core

import (
	"MrRSS/internal/adfilter"
	"MrRSS/internal/ai"
	"MrRSS/internal/utils/httputil"
	"MrRSS/internal/utils/textutil"
	"context"
	"encoding/json"
	"errors"
	"time"
)

func (h *Handler) AdFilterConfig(ctx context.Context) (adfilter.Config, error) {
	raw, err := h.DB.GetSettingContext(ctx, "ad_filter_config")
	if err != nil {
		return adfilter.Config{}, err
	}
	result := adfilter.DefaultConfig()
	if raw != "" {
		err = json.Unmarshal([]byte(raw), &result)
	}
	if err == nil {
		err = adfilter.Validate(result.Options)
	}
	return result, err
}

func (h *Handler) AnalyzeArticleAds(ctx context.Context, source, base string, options adfilter.Options) (adfilter.Analysis, error) {
	if !options.AIEnabled || !options.Applies(base) {
		return adfilter.Analysis{}, errors.New("ai_not_enabled")
	}
	provider := h.AIProfileProvider
	if provider == nil {
		provider = ai.NewProfileProvider(h.DB)
	}
	profile, err := provider.GetProfileForFeature(ai.FeatureAdFilter)
	if err != nil || profile == nil {
		return adfilter.Analysis{}, errors.New("ai_not_configured")
	}
	config, err := provider.GetConfigForProfile(profile.ID)
	if err != nil || config == nil || config.Endpoint == "" || config.Model == "" {
		return adfilter.Analysis{}, errors.New("ai_not_configured")
	}
	client, err := httputil.CreateHTTPClientWithProxySettings(h.DB, 40*time.Second)
	if err != nil {
		return adfilter.Analysis{}, errors.New("ai_unavailable")
	}
	defer client.CloseIdleConnections()
	aiClient := ai.NewClientWithHTTPClient(*config, client)
	modelKey, _ := json.Marshal(config)
	safe := textutil.PrepareFilteredArticle(source, base, options).Content
	complete := func(ctx context.Context, system, user string) (string, error) {
		if h.AITracker != nil {
			if h.AITracker.IsLimitReached() {
				return "", errors.New("ai_limit")
			}
			if err := h.AITracker.WaitForRateLimitContext(ctx); err != nil {
				return "", err
			}
		}
		answer, err := aiClient.RequestSingleFormatContext(ctx, ai.RequestConfig{Model: config.Model, SystemPrompt: system, UserPrompt: user, Temperature: 0, MaxTokens: 2500, MaxCompletionTokens: 3000})
		if err != nil {
			return "", errors.New("ai_unavailable")
		}
		if h.AITracker != nil {
			h.AITracker.TrackSummary(system+user, answer.Content)
		}
		return answer.Content, nil
	}
	result, err := h.AdAnalyzer.Analyze(ctx, safe, base, string(modelKey), options, complete)
	if err != nil {
		return result, err
	}
	result.Content = textutil.PrepareArticleContent(result.Content, base)
	return result, nil
}

func (h *Handler) AdFilterOptions(ctx context.Context) adfilter.Options {
	config, err := h.AdFilterConfig(ctx)
	if err != nil {
		return adfilter.Options{}
	}
	return config.Options
}

func (h *Handler) PrepareArticleForReader(ctx context.Context, content, base string) adfilter.Content {
	return textutil.PrepareFilteredArticle(content, base, h.AdFilterOptions(ctx))
}

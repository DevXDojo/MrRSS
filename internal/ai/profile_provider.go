package ai

import (
	"strconv"

	"MrRSS/internal/config"
	"MrRSS/internal/models"
)

// ProfileProvider provides AI profile resolution for different features
type ProfileProvider struct {
	db ProfileDB
}

// ProfileDB interface for database operations needed by ProfileProvider
type ProfileDB interface {
	GetAIProfile(id int64) (*models.AIProfile, error)
	GetDefaultAIProfile() (*models.AIProfile, error)
	GetSetting(key string) (string, error)
}

// NewProfileProvider creates a new ProfileProvider
func NewProfileProvider(db ProfileDB) *ProfileProvider {
	return &ProfileProvider{db: db}
}

// FeatureType represents different AI features that can have separate profile configurations
type FeatureType string

const (
	FeatureTranslation FeatureType = "translation"
	FeatureSummary     FeatureType = "summary"
	FeatureChat        FeatureType = "chat"
	FeatureSearch      FeatureType = "search"
)

// GetProfileForFeature returns the AI profile configured for a specific feature
// Falls back to default profile if no specific profile is configured
func (p *ProfileProvider) GetProfileForFeature(feature FeatureType) (*models.AIProfile, error) {
	// Get the setting key for this feature
	settingKey := p.getSettingKeyForFeature(feature)

	// Try to get the configured profile ID
	profileIDStr, err := p.db.GetSetting(settingKey)
	if err == nil && profileIDStr != "" {
		profileID, err := strconv.ParseInt(profileIDStr, 10, 64)
		if err == nil && profileID > 0 {
			profile, err := p.db.GetAIProfile(profileID)
			if err == nil && profile != nil {
				return profile, nil
			}
		}
	}

	// Fallback to default profile
	return p.db.GetDefaultAIProfile()
}

// getSettingKeyForFeature returns the settings key for a feature's AI profile
func (p *ProfileProvider) getSettingKeyForFeature(feature FeatureType) string {
	switch feature {
	case FeatureTranslation:
		return "ai_translation_profile_id"
	case FeatureSummary:
		return "ai_summary_profile_id"
	case FeatureChat:
		return "ai_chat_profile_id"
	case FeatureSearch:
		return "ai_search_profile_id"
	default:
		return ""
	}
}

// GetConfigForFeature returns the AI client config for a specific feature
// This is a convenience method that combines profile lookup with config creation
func (p *ProfileProvider) GetConfigForFeature(feature FeatureType) (*ClientConfig, error) {
	profile, err := p.GetProfileForFeature(feature)
	if err != nil {
		return nil, err
	}

	if profile == nil {
		// No profile configured, return defaults
		defaults := config.Get()
		return &ClientConfig{
			APIKey:   "",
			Endpoint: defaults.AIEndpoint,
			Model:    defaults.AIModel,
		}, nil
	}

	cfg := &ClientConfig{
		APIKey:        profile.APIKey,
		Endpoint:      profile.Endpoint,
		Model:         profile.Model,
		CustomHeaders: profile.CustomHeaders, // Keep as string, will be parsed by client
	}

	return cfg, nil
}

// HasProfileConfigured checks if a specific profile is configured for a feature
func (p *ProfileProvider) HasProfileConfigured(feature FeatureType) bool {
	settingKey := p.getSettingKeyForFeature(feature)
	profileIDStr, err := p.db.GetSetting(settingKey)
	if err != nil || profileIDStr == "" {
		return false
	}
	profileID, err := strconv.ParseInt(profileIDStr, 10, 64)
	return err == nil && profileID > 0
}

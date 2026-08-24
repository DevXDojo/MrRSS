package service

import (
	"context"

	"MrRSS/internal/ai"
	"MrRSS/internal/translation"
)

// translationService implements TranslationService interface
type translationService struct {
	translator translation.Translator
	aiTracker  *ai.UsageTracker
}

// NewTranslationService creates a new translation service
func NewTranslationService(translator translation.Translator, aiTracker *ai.UsageTracker) TranslationService {
	return &translationService{
		translator: translator,
		aiTracker:  aiTracker,
	}
}

// Translate translates text to target language
func (s *translationService) Translate(ctx context.Context, text, targetLang string) (string, error) {
	return s.translator.Translate(text, targetLang)
}

// TranslateArticle translates an article
func (s *translationService) TranslateArticle(ctx context.Context, articleID int64, targetLang string) error {
	// This is a placeholder - the actual implementation would:
	// 1. Get the article content from the database
	// 2. Translate the content
	// 3. Store the translation in the translation cache
	// For now, this just returns nil to satisfy the interface
	return nil
}

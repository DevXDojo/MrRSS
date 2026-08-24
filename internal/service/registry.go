package service

import (
	"sync"

	"MrRSS/internal/ai"
	"MrRSS/internal/cache"
	"MrRSS/internal/database"
	"MrRSS/internal/discovery"
	"MrRSS/internal/feed"
	"MrRSS/internal/statistics"
	"MrRSS/internal/translation"
)

// Registry is the central service registry that manages all application services.
// It provides lazy initialization and thread-safe access to services.
type Registry struct {
	db   *database.DB
	once sync.Once

	// Core dependencies
	fetcher          *feed.Fetcher
	translator       translation.Translator
	aiTracker        *ai.UsageTracker
	discoveryService *discovery.Service
	contentCache     *cache.ContentCache
	stats            *statistics.Service

	// Service instances
	articleSvc     ArticleService
	feedSvc        FeedService
	translationSvc TranslationService
	aiSvc          AIService
	discoverySvc   DiscoveryService
	settingsSvc    SettingsService
}

// NewRegistry creates a new service registry.
func NewRegistry(db *database.DB, fetcher *feed.Fetcher, translator translation.Translator) *Registry {
	return &Registry{
		db:         db,
		fetcher:    fetcher,
		translator: translator,
	}
}

// initialize initializes all services lazily on first access
func (r *Registry) initialize() {
	// Initialize core dependencies if not already set
	if r.aiTracker == nil {
		r.aiTracker = ai.NewUsageTracker(r.db)
	}
	if r.discoveryService == nil {
		r.discoveryService = discovery.NewService()
	}
	if r.contentCache == nil {
		r.contentCache = cache.NewContentCache(100, 30*60) // 100 articles, 30 minutes
	}
	if r.stats == nil {
		r.stats = statistics.NewService(r.db)
	}

	// Initialize services
	r.settingsSvc = NewSettingsService(r.db)
	r.articleSvc = NewArticleService(r, r.db)
	r.feedSvc = NewFeedService(r, r.db)
	r.translationSvc = NewTranslationService(r.translator, r.aiTracker)
	r.aiSvc = NewAIService(r, r.db)
	r.discoverySvc = NewDiscoveryServiceWrapper(r.discoveryService)
}

// Article returns the article service
func (r *Registry) Article() ArticleService {
	r.once.Do(r.initialize)
	return r.articleSvc
}

// Feed returns the feed service
func (r *Registry) Feed() FeedService {
	r.once.Do(r.initialize)
	return r.feedSvc
}

// Translation returns the translation service
func (r *Registry) Translation() TranslationService {
	r.once.Do(r.initialize)
	return r.translationSvc
}

// AI returns the AI service
func (r *Registry) AI() AIService {
	r.once.Do(r.initialize)
	return r.aiSvc
}

// Discovery returns the discovery service
func (r *Registry) Discovery() DiscoveryService {
	r.once.Do(r.initialize)
	return r.discoverySvc
}

// Settings returns the settings service
func (r *Registry) Settings() SettingsService {
	r.once.Do(r.initialize)
	return r.settingsSvc
}

// DB returns the database instance (for backward compatibility)
func (r *Registry) DB() *database.DB {
	return r.db
}

// Fetcher returns the feed fetcher (for backward compatibility)
func (r *Registry) Fetcher() *feed.Fetcher {
	return r.fetcher
}

// ContentCache returns the content cache (for backward compatibility)
func (r *Registry) ContentCache() *cache.ContentCache {
	r.once.Do(r.initialize)
	return r.contentCache
}

// Stats returns the statistics service (for backward compatibility)
func (r *Registry) Stats() *statistics.Service {
	r.once.Do(r.initialize)
	return r.stats
}

// AITracker returns the AI usage tracker (for backward compatibility)
func (r *Registry) AITracker() *ai.UsageTracker {
	r.once.Do(r.initialize)
	return r.aiTracker
}

// DiscoveryService returns the discovery service (for backward compatibility)
func (r *Registry) DiscoveryService() *discovery.Service {
	r.once.Do(r.initialize)
	return r.discoveryService
}

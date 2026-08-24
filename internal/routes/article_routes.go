package routes

import (
	"net/http"

	article "MrRSS/internal/handlers/article"
	"MrRSS/internal/handlers/core"
	summary "MrRSS/internal/handlers/summary"
	translationhandlers "MrRSS/internal/handlers/translation"
)

// registerArticleRoutes registers all article-related routes
func registerArticleRoutes(mux *http.ServeMux, h *core.Handler) {
	// Article CRUD and status
	mux.HandleFunc("/api/articles", func(w http.ResponseWriter, r *http.Request) { article.HandleArticles(h, w, r) })
	mux.HandleFunc("/api/articles/images", func(w http.ResponseWriter, r *http.Request) { article.HandleImageGalleryArticles(h, w, r) })
	mux.HandleFunc("/api/articles/filter", func(w http.ResponseWriter, r *http.Request) { article.HandleFilteredArticles(h, w, r) })
	mux.HandleFunc("/api/articles/read", func(w http.ResponseWriter, r *http.Request) { article.HandleMarkReadWithImmediateSync(h, w, r) })
	mux.HandleFunc("/api/articles/favorite", func(w http.ResponseWriter, r *http.Request) { article.HandleToggleFavoriteWithImmediateSync(h, w, r) })
	mux.HandleFunc("/api/articles/mark-relative", func(w http.ResponseWriter, r *http.Request) { article.HandleMarkRelativeToArticle(h, w, r) })
	mux.HandleFunc("/api/articles/toggle-hide", func(w http.ResponseWriter, r *http.Request) { article.HandleToggleHideArticle(h, w, r) })
	mux.HandleFunc("/api/articles/toggle-read-later", func(w http.ResponseWriter, r *http.Request) { article.HandleToggleReadLater(h, w, r) })
	mux.HandleFunc("/api/articles/mark-all-read", func(w http.ResponseWriter, r *http.Request) { article.HandleMarkAllAsRead(h, w, r) })
	mux.HandleFunc("/api/articles/clear-read-later", func(w http.ResponseWriter, r *http.Request) { article.HandleClearReadLater(h, w, r) })

	// Article content
	mux.HandleFunc("/api/articles/content", func(w http.ResponseWriter, r *http.Request) { article.HandleGetArticleContent(h, w, r) })
	mux.HandleFunc("/api/articles/reload-content", func(w http.ResponseWriter, r *http.Request) { article.HandleReloadArticleContent(h, w, r) })
	mux.HandleFunc("/api/articles/fetch-full", func(w http.ResponseWriter, r *http.Request) { article.HandleFetchFullArticle(h, w, r) })
	mux.HandleFunc("/api/articles/extract-images", func(w http.ResponseWriter, r *http.Request) { article.HandleExtractAllImages(h, w, r) })

	// Article statistics
	mux.HandleFunc("/api/articles/unread-counts", func(w http.ResponseWriter, r *http.Request) { article.HandleGetUnreadCounts(h, w, r) })
	mux.HandleFunc("/api/articles/filter-counts", func(w http.ResponseWriter, r *http.Request) { article.HandleGetFilterCounts(h, w, r) })

	// Article cleanup
	mux.HandleFunc("/api/articles/cleanup", func(w http.ResponseWriter, r *http.Request) { article.HandleCleanupArticles(h, w, r) })
	mux.HandleFunc("/api/articles/cleanup-content", func(w http.ResponseWriter, r *http.Request) { article.HandleCleanupArticleContent(h, w, r) })
	mux.HandleFunc("/api/articles/content-cache-info", func(w http.ResponseWriter, r *http.Request) { article.HandleGetArticleContentCacheInfo(h, w, r) })

	// Translation
	mux.HandleFunc("/api/articles/translate", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleTranslateArticle(h, w, r) })
	mux.HandleFunc("/api/articles/translate-text", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleTranslateText(h, w, r) })
	mux.HandleFunc("/api/articles/clear-translations", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleClearTranslations(h, w, r) })

	// AI usage (translation related)
	mux.HandleFunc("/api/ai-usage", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleGetAIUsage(h, w, r) })
	mux.HandleFunc("/api/ai-usage/reset", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleResetAIUsage(h, w, r) })
	mux.HandleFunc("/api/translation/test-custom", func(w http.ResponseWriter, r *http.Request) { translationhandlers.HandleTestCustomTranslation(h, w, r) })

	// Summary
	mux.HandleFunc("/api/articles/summarize", func(w http.ResponseWriter, r *http.Request) { summary.HandleSummarizeArticle(h, w, r) })
	mux.HandleFunc("/api/articles/clear-summaries", func(w http.ResponseWriter, r *http.Request) { summary.HandleClearSummaries(h, w, r) })

	// Export
	mux.HandleFunc("/api/articles/export/obsidian", func(w http.ResponseWriter, r *http.Request) { article.HandleExportToObsidian(h, w, r) })
	mux.HandleFunc("/api/articles/export/notion", func(w http.ResponseWriter, r *http.Request) { article.HandleExportToNotion(h, w, r) })
	mux.HandleFunc("/api/articles/export/zotero", func(w http.ResponseWriter, r *http.Request) { article.HandleExportToZotero(h, w, r) })
}

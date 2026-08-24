package summary

import (
	"encoding/json"
	"log"
	"net/http"

	"MrRSS/internal/ai"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/summary"
	"MrRSS/internal/utils/textutil"
)

// HandleSummarizeArticle generates a summary for an article's content.
// @Summary      Summarize article
// @Description  Generate a summary for an article's content (uses local algorithm or AI based on settings)
// @Tags         summary
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Summarize request (article_id, length, content)"
// @Success      200  {object}  map[string]interface{}  "Summary result (summary, html, sentence_count, is_too_short, cached, limit_reached, thinking)"
// @Failure      400  {object}  map[string]string  "Bad request (invalid length parameter)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /summarize [post]
func HandleSummarizeArticle(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ArticleID int64  `json:"article_id"`
		Length    string `json:"length"`            // "short", "medium", "long"
		Content   string `json:"content,omitempty"` // Optional: use provided content instead of fetching from DB
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Validate length parameter
	summaryLength := summary.Medium
	switch req.Length {
	case "short":
		summaryLength = summary.Short
	case "long":
		summaryLength = summary.Long
	case "medium", "":
		summaryLength = summary.Medium
	default:
		response.Error(w, nil, http.StatusBadRequest)
		return
	}

	// Get summary provider from settings (with default)
	provider, err := h.DB.GetSetting("summary_provider")
	if err != nil || provider == "" {
		provider = "local" // Default to local algorithm
	}

	if provider == "rss" {
		originalSummary, err := h.DB.GetArticleOriginalSummary(req.ArticleID)
		if err != nil {
			log.Printf("Error getting original article summary: %v", err)
			response.Error(w, err, http.StatusInternalServerError)
			return
		}

		originalSummary = textutil.CleanHTML(originalSummary)
		if originalSummary == "" {
			response.JSON(w, map[string]interface{}{
				"summary":      "",
				"is_too_short": true,
				"error":        "No RSS summary available for this article",
			})
			return
		}

		response.JSON(w, map[string]interface{}{
			"summary":        originalSummary,
			"html":           textutil.SanitizeHTML(originalSummary),
			"sentence_count": 0,
			"is_too_short":   false,
			"cached":         true,
			"source":         "rss",
		})
		return
	}

	// Check if article already has a cached generated summary in database.
	// If content is provided (for on-the-fly summarization), skip this check.
	if req.Content == "" {
		article, err := h.DB.GetArticleByID(req.ArticleID)
		if err == nil && article.Summary != "" && article.Summary != "<no content>" {
			// Article has a cached summary, convert it to HTML and return
			htmlSummary := textutil.ConvertMarkdownToHTML(article.Summary)
			response.JSON(w, map[string]interface{}{
				"summary":        article.Summary,
				"html":           htmlSummary,
				"sentence_count": 0, // We don't store this in DB
				"is_too_short":   false,
				"cached":         true,
			})
			return
		}
	}

	// Get the article content
	content, err := getArticleContent(h, req.ArticleID, req.Content)
	if err != nil {
		log.Printf("Error getting article content for summary: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	if content == "" {
		response.JSON(w, map[string]interface{}{
			"summary":      "",
			"is_too_short": true,
			"error":        "No content available for this article",
		})
		return
	}

	var result summary.SummaryResult
	usedFallback := false
	limitReached := false

	if provider == "ai" {
		// Check if AI usage limit is reached - fallback to local if so
		if h.AITracker.IsLimitReached() {
			log.Printf("AI usage limit reached, falling back to local summarization")
			limitReached = true
			summarizer := summary.NewSummarizer()
			result = summarizer.Summarize(content, summaryLength)
			usedFallback = true
		} else {
			// Use AI summarization
			// Apply rate limiting for AI requests
			h.AITracker.WaitForRateLimit()

			// Try to get AI config from ProfileProvider first
			var apiKey, endpoint, model string
			if h.AIProfileProvider != nil {
				cfg, err := h.AIProfileProvider.GetConfigForFeature(ai.FeatureSummary)
				if err == nil && cfg != nil {
					apiKey = cfg.APIKey
					endpoint = cfg.Endpoint
					model = cfg.Model
					log.Printf("Using AI profile for summarization (endpoint: %s, model: %s)", endpoint, model)
				}
			}

			// Fallback to global settings if ProfileProvider not available or no profile configured
			if apiKey == "" && endpoint == "" {
				apiKey, _ = h.DB.GetEncryptedSetting("ai_api_key")
				endpoint, _ = h.DB.GetSetting("ai_endpoint")
				model, _ = h.DB.GetSetting("ai_model")
				log.Printf("Using global AI settings for summarization (API key: %s)", func() string {
					if apiKey != "" {
						return "configured"
					}
					return "not configured (using keyless provider)"
				}())
			}

			systemPrompt, _ := h.DB.GetSetting("ai_summary_prompt")
			customHeaders, _ := h.DB.GetSetting("ai_custom_headers")
			language, _ := h.DB.GetSetting("language")

			aiSummarizer := summary.NewAISummarizerWithDB(apiKey, endpoint, model, h.DB)
			if systemPrompt != "" {
				aiSummarizer.SetSystemPrompt(systemPrompt)
			}
			if customHeaders != "" {
				aiSummarizer.SetCustomHeaders(customHeaders)
			}
			if language != "" {
				aiSummarizer.SetLanguage(language)
			}
			aiResult, err := aiSummarizer.Summarize(content, summaryLength)
			if err != nil {
				log.Printf("Error generating AI summary, falling back to local: %v", err)
				// Fallback to local algorithm on any AI error
				summarizer := summary.NewSummarizer()
				result = summarizer.Summarize(content, summaryLength)
				usedFallback = true
			} else {
				result = aiResult
				// Track AI usage only on success
				h.AITracker.TrackSummary(content, result.Summary)
				// Track statistics
				_ = h.DB.IncrementStat("ai_summary")
			}
		}
	} else {
		// Use local algorithm
		summarizer := summary.NewSummarizer()
		result = summarizer.Summarize(content, summaryLength)
	}

	// Cache the summary in the database
	if err := h.DB.UpdateArticleSummary(req.ArticleID, result.Summary); err != nil {
		log.Printf("Failed to cache summary for article %d: %v", req.ArticleID, err)
		// Don't fail the request if caching fails
	}

	// Convert markdown summary to HTML (for all summaries, not just AI)
	htmlSummary := textutil.ConvertMarkdownToHTML(result.Summary)

	resp := map[string]interface{}{
		"summary":        result.Summary,
		"html":           htmlSummary,
		"sentence_count": result.SentenceCount,
		"is_too_short":   result.IsTooShort,
		"limit_reached":  limitReached,
		"thinking":       result.Thinking,
	}
	if usedFallback {
		resp["used_fallback"] = true
	}

	response.JSON(w, resp)
}

// getArticleContent fetches the content of an article by ID, or uses provided content
func getArticleContent(h *core.Handler, articleID int64, providedContent string) (string, error) {
	// If content is provided, use it directly
	if providedContent != "" {
		return providedContent, nil
	}

	// Otherwise, fetch from database/cache
	content, _, err := h.GetArticleContent(articleID)
	return content, err
}

// HandleClearSummaries clears all cached summaries from the database.
// @Summary      Clear all summaries
// @Description  Clear all cached article summaries from the database
// @Tags         summary
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]bool  "Success status"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /summaries/clear [delete]
func HandleClearSummaries(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	if err := h.DB.ClearAllSummaries(); err != nil {
		log.Printf("Error clearing summaries: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]interface{}{"success": true})
}

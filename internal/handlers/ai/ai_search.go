package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"MrRSS/internal/ai"
	"MrRSS/internal/config"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

// AISearchRequest represents the request for AI-powered search
type AISearchRequest struct {
	Query string `json:"query"`
}

// AISearchResponse represents the response from AI search
type AISearchResponse struct {
	Success     bool             `json:"success"`
	Articles    []map[string]any `json:"articles,omitempty"`
	SearchTerms string           `json:"search_terms,omitempty"`
	Error       string           `json:"error,omitempty"`
	TotalCount  int              `json:"total_count"`
}

// extractSearchTerms extracts the search terms from AI response
func extractSearchTerms(response string) string {
	response = strings.TrimSpace(response)

	// Remove markdown code blocks if present
	codeBlockPattern := regexp.MustCompile("(?s)```(?:json)?\\s*(.+?)\\s*```")
	matches := codeBlockPattern.FindStringSubmatch(response)
	if len(matches) > 1 {
		response = strings.TrimSpace(matches[1])
	}

	return response
}

// SearchTerms represents parsed search terms with required and optional categories
type SearchTerms struct {
	Required []string `json:"required"` // Must match at least one
	Optional []string `json:"optional"` // Boost relevance if matched
	Patterns []string `json:"patterns"` // LIKE patterns like "详解%llm"
}

// parseSearchTermsAdvanced parses JSON object with required/optional/patterns from AI response
func parseSearchTermsAdvanced(response string) (*SearchTerms, error) {
	cleaned := extractSearchTerms(response)

	// Try to parse as structured JSON object
	var terms SearchTerms
	if err := json.Unmarshal([]byte(cleaned), &terms); err != nil {
		// Fallback: try parsing as simple array and treat all as required
		var simpleTerms []string
		if err := json.Unmarshal([]byte(cleaned), &simpleTerms); err != nil {
			// Last fallback: split by comma
			cleaned = strings.ReplaceAll(cleaned, "\n", ",")
			parts := strings.Split(cleaned, ",")
			for _, part := range parts {
				term := strings.Trim(strings.TrimSpace(part), `"'[]{}`)
				if term != "" {
					terms.Required = append(terms.Required, term)
				}
			}
		} else {
			terms.Required = simpleTerms
		}
	}

	if len(terms.Required) == 0 && len(terms.Optional) == 0 && len(terms.Patterns) == 0 {
		return nil, fmt.Errorf("no search terms extracted")
	}

	return &terms, nil
}

// buildAISearchPrompt creates a concise system prompt for keyword expansion
func buildAISearchPrompt() string {
	return `Expand search query into structured search terms. Output ONLY a JSON object:
{
  "required": ["must-match keywords - core topic"],
  "optional": ["nice-to-have keywords - related/synonyms"],
  "patterns": ["SQL LIKE patterns with % wildcards"]
}

Rules:
- required: Core topic keywords that MUST appear (2-5 terms)
- optional: Related terms for better ranking (3-8 terms)
- patterns: Specific phrase patterns using % as wildcard (0-3 patterns)
- Include English and Chinese terms where applicable

Examples:
Input: "llm详解"
Output: {"required":["LLM","大语言模型","large language model"],"optional":["GPT","Claude","transformer","神经网络","深度学习"],"patterns":["详解%LLM","LLM%详解","LLM%教程"]}

Input: "Python web框架"
Output: {"required":["Python","web框架","web framework"],"optional":["Django","Flask","FastAPI","后端","backend"],"patterns":["Python%web","Python%框架"]}`
}

// buildSearchSQL builds the SQL query from search terms with relevance scoring
func buildSearchSQL(terms *SearchTerms, limit int) string {
	if terms == nil || (len(terms.Required) == 0 && len(terms.Patterns) == 0) {
		return ""
	}

	// Build required conditions (must match at least one)
	var requiredConditions []string
	for _, term := range terms.Required {
		escapedTerm := strings.ReplaceAll(term, "'", "''")
		condition := fmt.Sprintf(
			"(a.title LIKE '%%%s%%' OR c.content LIKE '%%%s%%' OR a.summary LIKE '%%%s%%')",
			escapedTerm, escapedTerm, escapedTerm,
		)
		requiredConditions = append(requiredConditions, condition)
	}

	// Build pattern conditions
	for _, pattern := range terms.Patterns {
		escapedPattern := strings.ReplaceAll(pattern, "'", "''")
		condition := fmt.Sprintf(
			"(a.title LIKE '%%%s%%' OR c.content LIKE '%%%s%%' OR a.summary LIKE '%%%s%%')",
			escapedPattern, escapedPattern, escapedPattern,
		)
		requiredConditions = append(requiredConditions, condition)
	}

	// Build relevance score with weighted scoring
	var scoreTerms []string

	// Required terms get higher base weight
	for _, term := range terms.Required {
		escapedTerm := strings.ReplaceAll(term, "'", "''")
		scoreTerm := fmt.Sprintf(
			"(CASE WHEN a.title LIKE '%%%s%%' THEN 5 ELSE 0 END + "+
				"CASE WHEN c.content LIKE '%%%s%%' THEN 2 ELSE 0 END + "+
				"CASE WHEN a.summary LIKE '%%%s%%' THEN 3 ELSE 0 END)",
			escapedTerm, escapedTerm, escapedTerm,
		)
		scoreTerms = append(scoreTerms, scoreTerm)
	}

	// Patterns get highest weight (exact phrase match)
	for _, pattern := range terms.Patterns {
		escapedPattern := strings.ReplaceAll(pattern, "'", "''")
		scoreTerm := fmt.Sprintf(
			"(CASE WHEN a.title LIKE '%%%s%%' THEN 8 ELSE 0 END + "+
				"CASE WHEN c.content LIKE '%%%s%%' THEN 4 ELSE 0 END + "+
				"CASE WHEN a.summary LIKE '%%%s%%' THEN 5 ELSE 0 END)",
			escapedPattern, escapedPattern, escapedPattern,
		)
		scoreTerms = append(scoreTerms, scoreTerm)
	}

	// Optional terms get lower weight
	for _, term := range terms.Optional {
		escapedTerm := strings.ReplaceAll(term, "'", "''")
		scoreTerm := fmt.Sprintf(
			"(CASE WHEN a.title LIKE '%%%s%%' THEN 2 ELSE 0 END + "+
				"CASE WHEN c.content LIKE '%%%s%%' THEN 1 ELSE 0 END + "+
				"CASE WHEN a.summary LIKE '%%%s%%' THEN 1 ELSE 0 END)",
			escapedTerm, escapedTerm, escapedTerm,
		)
		scoreTerms = append(scoreTerms, scoreTerm)
	}

	relevanceScore := "0"
	if len(scoreTerms) > 0 {
		relevanceScore = strings.Join(scoreTerms, " + ")
	}

	// Build full query with LEFT JOIN to article_contents for content search
	whereClause := strings.Join(requiredConditions, " OR ")
	query := fmt.Sprintf(`
		SELECT a.id, a.feed_id, a.title, a.url, a.image_url, a.audio_url, a.video_url,
			   a.published_at, a.is_read, a.is_favorite, a.is_hidden, a.is_read_later,
			   a.translated_title, a.summary, a.freshrss_item_id, f.title AS feed_title, a.author,
			   (%s) AS relevance_score
		FROM articles a
		JOIN feeds f ON a.feed_id = f.id
		LEFT JOIN article_contents c ON a.id = c.article_id
		WHERE a.is_hidden = 0 AND (%s)
		ORDER BY relevance_score DESC, a.published_at DESC
		LIMIT %d
	`, relevanceScore, whereClause, limit)

	return query
}

// HandleAISearch handles POST /api/ai/search for AI-powered article search
// @Summary      AI-powered article search
// @Description  Use AI to expand keywords and search articles with relevance ranking
// @Tags         ai
// @Accept       json
// @Produce      json
// @Param        request  body      AISearchRequest  true  "Search query"
// @Success      200  {object}  AISearchResponse  "Search results"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /ai/search [post]
func HandleAISearch(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req AISearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, AISearchResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	if strings.TrimSpace(req.Query) == "" {
		response.JSON(w, AISearchResponse{
			Success: false,
			Error:   "Search query is required",
		})
		return
	}

	log.Printf("[AI Search] User query: %s", req.Query)

	// Get AI settings - try ProfileProvider first
	var apiKey, endpoint, model string
	if h.AIProfileProvider != nil {
		cfg, err := h.AIProfileProvider.GetConfigForFeature(ai.FeatureSearch)
		if err == nil && cfg != nil && (cfg.APIKey != "" || cfg.Endpoint != "") {
			apiKey = cfg.APIKey
			endpoint = cfg.Endpoint
			model = cfg.Model
			log.Printf("[AI Search] Using AI profile for search (endpoint: %s, model: %s)", endpoint, model)
		}
	}

	// Fallback to global settings if no profile configured
	if endpoint == "" {
		apiKey, _ = h.DB.GetEncryptedSetting("ai_api_key")
		endpoint, _ = h.DB.GetSetting("ai_endpoint")
		model, _ = h.DB.GetSetting("ai_model")

		// Use defaults if not set
		defaults := config.Get()
		if endpoint == "" {
			endpoint = defaults.AIEndpoint
		}
		if model == "" {
			model = defaults.AIModel
		}
		log.Printf("[AI Search] Using global AI settings for search (endpoint: %s, model: %s)", endpoint, model)
	}

	// Validate AI configuration
	if endpoint == "" || model == "" {
		response.JSON(w, AISearchResponse{
			Success: false,
			Error:   "AI is not configured. Please configure AI settings first.",
		})
		return
	}

	// Create AI client
	httpClient, err := createHTTPClientWithProxy(h)
	if err != nil {
		response.JSON(w, AISearchResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to create HTTP client: %v", err),
		})
		return
	}
	httpClient.Timeout = 30 * time.Second

	clientConfig := ai.ClientConfig{
		APIKey:   apiKey,
		Endpoint: endpoint,
		Model:    model,
		Timeout:  30 * time.Second,
	}
	client := ai.NewClientWithHTTPClient(clientConfig, httpClient)

	// Get expanded search terms from AI
	systemPrompt := buildAISearchPrompt()
	aiResponse, err := client.Request(systemPrompt, req.Query)
	if err != nil {
		response.JSON(w, AISearchResponse{
			Success: false,
			Error:   fmt.Sprintf("AI request failed: %v", err),
		})
		return
	}

	log.Printf("[AI Search] AI response: %s", aiResponse)

	// Parse search terms from AI response
	searchTerms, err := parseSearchTermsAdvanced(aiResponse)
	if err != nil {
		response.JSON(w, AISearchResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse search terms: %v", err),
		})
		return
	}

	// Format terms for logging and response
	var allTerms []string
	allTerms = append(allTerms, searchTerms.Required...)
	allTerms = append(allTerms, searchTerms.Optional...)
	allTerms = append(allTerms, searchTerms.Patterns...)
	log.Printf("[AI Search] Required: %v, Optional: %v, Patterns: %v", searchTerms.Required, searchTerms.Optional, searchTerms.Patterns)

	// Build and execute search query
	searchSQL := buildSearchSQL(searchTerms, 100)
	log.Printf("[AI Search] SQL query:\n%s", searchSQL)

	// Execute search
	articles, err := h.DB.SearchArticlesWithSQL(searchSQL)
	if err != nil {
		log.Printf("[AI Search] Query error: %v", err)
		response.JSON(w, AISearchResponse{
			Success:     false,
			Error:       fmt.Sprintf("Search query failed: %v", err),
			SearchTerms: strings.Join(allTerms, ", "),
		})
		return
	}

	log.Printf("[AI Search] Found %d articles", len(articles))

	// Convert articles to response format
	articleMaps := make([]map[string]any, len(articles))
	for i, article := range articles {
		articleMaps[i] = map[string]any{
			"id":               article.ID,
			"feed_id":          article.FeedID,
			"title":            article.Title,
			"url":              article.URL,
			"image_url":        article.ImageURL,
			"audio_url":        article.AudioURL,
			"video_url":        article.VideoURL,
			"published_at":     article.PublishedAt,
			"is_read":          article.IsRead,
			"is_favorite":      article.IsFavorite,
			"is_hidden":        article.IsHidden,
			"is_read_later":    article.IsReadLater,
			"feed_title":       article.FeedTitle,
			"author":           article.Author,
			"translated_title": article.TranslatedTitle,
			"summary":          article.Summary,
		}
	}

	response.JSON(w, AISearchResponse{
		Success:     true,
		Articles:    articleMaps,
		SearchTerms: strings.Join(allTerms, ", "),
		TotalCount:  len(articles),
	})
}

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
	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

// AISearchRequest represents the request for AI-powered search
type AISearchRequest struct {
	Query string `json:"query"`
}

// AISearchResponse represents the response from AI search
type AISearchResponse struct {
	Success       bool             `json:"success"`
	Articles      []map[string]any `json:"articles,omitempty"`
	SearchTerms   string           `json:"search_terms,omitempty"`
	ExpandedTerms *SearchTerms     `json:"expanded_terms,omitempty"`
	Error         string           `json:"error,omitempty"`
	ErrorCode     string           `json:"error_code,omitempty"`
	TotalCount    int              `json:"total_count"`
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

	terms.Required = normalizeTerms(terms.Required, 8)
	terms.Optional = normalizeTerms(terms.Optional, 12)
	terms.Patterns = normalizeSearchPatterns(terms.Patterns, 4)
	return &terms, nil
}

func normalizeSearchPatterns(values []string, max int) []string {
	result := make([]string, 0, len(values))
	for _, value := range normalizeTerms(values, max) {
		plain := strings.Join(strings.Fields(strings.ReplaceAll(value, "%", " ")), " ")
		if plain == "" {
			continue
		}
		result = append(result, value)
	}
	return result
}

func normalizeTerms(values []string, max int) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{})
	for _, value := range values {
		value = strings.TrimSpace(value)
		key := strings.ToLower(value)
		if value == "" || len([]rune(value)) > 80 {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, value)
		if len(result) >= max {
			break
		}
	}
	return result
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
			Success:   false,
			Error:     "Invalid search request",
			ErrorCode: ai.ErrorCodeProviderRejectedRequest,
		})
		return
	}

	if strings.TrimSpace(req.Query) == "" {
		response.JSON(w, AISearchResponse{
			Success:   false,
			Error:     "Search query is required",
			ErrorCode: ai.ErrorCodeProviderRejectedRequest,
		})
		return
	}

	query := strings.TrimSpace(req.Query)
	log.Printf("[AI Search] Starting search (query length: %d)", len([]rune(query)))

	// Get AI settings - try ProfileProvider first
	var apiKey, endpoint, model string
	if h.AIProfileProvider != nil {
		cfg, err := h.AIProfileProvider.GetConfigForFeature(ai.FeatureSearch)
		if err == nil && cfg != nil && (cfg.APIKey != "" || cfg.Endpoint != "") {
			apiKey = cfg.APIKey
			endpoint = cfg.Endpoint
			model = cfg.Model
			log.Printf("[AI Search] Using AI profile for search (endpoint: %s, model: %s)", ai.RedactEndpoint(endpoint), model)
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
		log.Printf("[AI Search] Using global AI settings for search (endpoint: %s, model: %s)", ai.RedactEndpoint(endpoint), model)
	}

	// Validate AI configuration
	if endpoint == "" || model == "" {
		response.JSON(w, AISearchResponse{
			Success:   false,
			Error:     ai.UserFacingErrorForCode(ai.ErrorCodeConfigurationInvalid).Message,
			ErrorCode: ai.ErrorCodeConfigurationInvalid,
		})
		return
	}

	// Create AI client
	httpClient, err := createHTTPClientWithProxy(h)
	if err != nil {
		publicErr := ai.ClassifyUserFacingError(err)
		response.JSON(w, AISearchResponse{
			Success:   false,
			Error:     publicErr.Message,
			ErrorCode: publicErr.Code,
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
		publicErr := ai.ClassifyUserFacingError(err)
		log.Printf("[AI Search] Request failed code=%s status=%d", publicErr.Code, publicErr.HTTPStatus)
		response.JSON(w, AISearchResponse{
			Success:   false,
			Error:     publicErr.Message,
			ErrorCode: publicErr.Code,
		})
		return
	}

	// Parse search terms from AI response
	searchTerms, err := parseSearchTermsAdvanced(aiResponse)
	if err != nil {
		publicErr := ai.UserFacingErrorForCode(ai.ErrorCodeInvalidResponse)
		response.JSON(w, AISearchResponse{
			Success:   false,
			Error:     publicErr.Message,
			ErrorCode: publicErr.Code,
		})
		return
	}

	// Format terms for logging and response
	var allTerms []string
	allTerms = append(allTerms, searchTerms.Required...)
	allTerms = append(allTerms, searchTerms.Optional...)
	allTerms = append(allTerms, searchTerms.Patterns...)
	log.Printf(
		"[AI Search] Expanded terms (required=%d optional=%d patterns=%d)",
		len(searchTerms.Required), len(searchTerms.Optional), len(searchTerms.Patterns),
	)

	// Execute a parameterized candidate query and deterministic local ranking.
	// The AI response is never treated as SQL.
	results, err := h.DB.SearchArticlesWithTerms(database.AISearchQuery{
		Original: query,
		Required: searchTerms.Required,
		Optional: searchTerms.Optional,
		Patterns: searchTerms.Patterns,
		Limit:    100,
	})
	if err != nil {
		log.Printf("[AI Search] Query error: %v", err)
		response.JSON(w, AISearchResponse{
			Success:     false,
			Error:       "The article search could not be completed",
			ErrorCode:   ai.ErrorCodeRequestFailed,
			SearchTerms: strings.Join(allTerms, ", "),
		})
		return
	}

	log.Printf("[AI Search] Found %d articles", len(results))

	// Convert articles to response format
	articleMaps := make([]map[string]any, len(results))
	for i, result := range results {
		article := result.Article
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
			"original_summary": article.OriginalSummary,
			"relevance_score":  result.RelevanceScore,
			"matched_terms":    result.MatchedTerms,
			"matched_fields":   result.MatchedFields,
			"excerpt":          result.Excerpt,
		}
	}

	response.JSON(w, AISearchResponse{
		Success:       true,
		Articles:      articleMaps,
		SearchTerms:   strings.Join(allTerms, ", "),
		ExpandedTerms: searchTerms,
		TotalCount:    len(results),
	})
}

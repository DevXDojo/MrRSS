package translation

import (
	"encoding/json"
	"log"
	"net/http"

	"MrRSS/internal/ai"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/translation"
	"MrRSS/internal/utils/textutil"
)

// TestCustomTranslationRequest represents a request to test custom translation configuration
type TestCustomTranslationRequest struct {
	Text   string                             `json:"text" example:"Hello, world!"`
	Target string                             `json:"target_lang" example:"zh"`
	Config translation.CustomTranslatorConfig `json:"config"`
}

// TestCustomTranslationResponse represents the response from a custom translation test
type TestCustomTranslationResponse struct {
	Success bool   `json:"success"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HandleTranslateArticle translates an article's title.
// @Summary      Translate article title
// @Description  Translate an article's title to the target language (uses AI or Google based on settings)
// @Tags         translation
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Translation request (article_id, title, target_language)"
// @Success      200  {object}  map[string]interface{}  "Translation result (translated_title, limit_reached)"
// @Failure      400  {object}  map[string]string  "Bad request (missing required fields)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /translate/article [post]
func HandleTranslateArticle(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ArticleID  int64  `json:"article_id"`
		Title      string `json:"title"`
		TargetLang string `json:"target_language"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.TargetLang == "" {
		response.Error(w, nil, http.StatusBadRequest)
		return
	}

	// Step 0: Check if article already has a translation in database
	// This prevents re-translating already translated content
	article, err := h.DB.GetArticleByID(req.ArticleID)
	if err == nil && article != nil {
		if article.TranslatedTitle != "" && article.TranslatedTitle != article.Title {
			// Translation already exists and is different from original
			response.JSON(w, map[string]interface{}{
				"translated_title": article.TranslatedTitle,
				"limit_reached":    false,
				"skipped":          true, // Indicate translation was skipped (from cache)
				"cached":           true,
			})
			return
		}
	}

	// Step 1: Pre-translation language detection to avoid unnecessary API calls
	detector := translation.GetLanguageDetector()
	shouldTranslate := detector.ShouldTranslate(req.Title, req.TargetLang)

	if !shouldTranslate {
		// Text is already in target language, return original title
		if updateErr := h.DB.UpdateArticleTranslation(req.ArticleID, req.Title); updateErr != nil {
			response.Error(w, updateErr, http.StatusInternalServerError)
			return
		}
		response.JSON(w, map[string]interface{}{
			"translated_title": req.Title,
			"limit_reached":    false,
			"skipped":          true, // Indicate translation was skipped
			"reason":           "already_target_language",
		})
		return
	}

	// Step 2: Proceed with translation
	// Check if we should use AI translation or fallback to Google
	provider, _ := h.DB.GetSetting("translation_provider")
	isAIProvider := provider == "ai"

	var translatedTitle string
	var translateErr error
	var limitReached = false

	if isAIProvider {
		// Check if AI usage limit is reached
		if h.AITracker.IsLimitReached() {
			limitReached = true
			// Fallback to Google Translate
			googleTranslator := translation.NewGoogleFreeTranslatorWithDB(h.DB)
			translatedTitle, translateErr = translation.TranslateMarkdownPreservingStructure(req.Title, googleTranslator, req.TargetLang)
		} else {
			// Apply rate limiting for AI requests
			h.AITracker.WaitForRateLimit()

			// Use markdown-preserving translation for better list structure
			translatedTitle, translateErr = translation.TranslateMarkdownAIPrompt(req.Title, h.Translator, req.TargetLang)

			// If AI fails, fallback to Google Translate
			if translateErr != nil {
				googleTranslator := translation.NewGoogleFreeTranslatorWithDB(h.DB)
				translatedTitle, translateErr = translation.TranslateMarkdownPreservingStructure(req.Title, googleTranslator, req.TargetLang)
			}

			// Track AI usage only on success (whether AI or fallback)
			if translateErr == nil {
				h.AITracker.TrackTranslation(req.Title, translatedTitle)
			}
		}
	} else {
		// Non-AI provider, use markdown-preserving translation
		translatedTitle, translateErr = translation.TranslateMarkdownPreservingStructure(req.Title, h.Translator, req.TargetLang)
	}

	if translateErr != nil {
		response.Error(w, translateErr, http.StatusInternalServerError)
		return
	}

	// Step 3: Post-translation check - if translation equals original, it was already in target language
	// This provides a safety net in case pre-translation detection was inaccurate
	if translatedTitle == req.Title {
		// Still update DB with the "translated" text (which is the original)
		if updateErr := h.DB.UpdateArticleTranslation(req.ArticleID, translatedTitle); updateErr != nil {
			response.Error(w, updateErr, http.StatusInternalServerError)
			return
		}
		response.JSON(w, map[string]interface{}{
			"translated_title": translatedTitle,
			"limit_reached":    limitReached,
			"skipped":          true, // Indicate no actual translation was performed
			"reason":           "translation_equals_original",
		})
		return
	}

	// Update the article with the translated title
	if updateErr := h.DB.UpdateArticleTranslation(req.ArticleID, translatedTitle); updateErr != nil {
		response.Error(w, updateErr, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]interface{}{
		"translated_title": translatedTitle,
		"limit_reached":    limitReached,
		"skipped":          false, // Translation was performed
	})
}

// HandleClearTranslations clears all translated titles from the database.
// @Summary      Clear all translations
// @Description  Clear all translated article titles from the database
// @Tags         translation
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]bool  "Success status"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /translations/clear [post]
func HandleClearTranslations(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	if err := h.DB.ClearAllTranslations(); err != nil {
		log.Printf("Error clearing translations: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]bool{"success": true})
}

// HandleTranslateText translates any text to the target language.
// This is used for translating content, summaries, etc.
// @Summary      Translate text
// @Description  Translate any text to the target language (uses AI or Google based on settings)
// @Tags         translation
// @Accept       json
// @Produce      json
// @Param        request  body      object  true  "Translation request (text, target_language)"
// @Success      200  {object}  map[string]string  "Translation result (translated_text, html)"
// @Failure      400  {object}  map[string]string  "Bad request (missing required fields)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /translate/text [post]
func HandleTranslateText(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Text       string `json:"text"`
		TargetLang string `json:"target_language"`
		Force      bool   `json:"force"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding translation request: %v", err)
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	if req.Text == "" || req.TargetLang == "" {
		log.Printf("Missing required fields in translation request")
		response.Error(w, nil, http.StatusBadRequest)
		return
	}

	// Step 1: Pre-translation language detection to avoid unnecessary API calls
	detector := translation.GetLanguageDetector()
	// Use full-text analysis for better accuracy on longer content
	// Skip language detection if force flag is set
	shouldTranslate := req.Force || detector.ShouldTranslateFullText(req.Text, req.TargetLang)

	if !shouldTranslate {
		// Text is already in target language, return original text
		htmlText := textutil.ConvertMarkdownToHTML(req.Text)
		response.JSON(w, map[string]interface{}{
			"translated_text": req.Text,
			"html":            htmlText,
			"skipped":         "true", // Indicate translation was skipped
			"reason":          "already_target_language",
		})
		return
	}

	// Step 2: Proceed with translation
	// Check if we should use AI translation or fallback to Google
	provider, _ := h.DB.GetSetting("translation_provider")
	isAIProvider := provider == "ai"

	var translatedText string
	var err error

	if isAIProvider {
		// Check if AI usage limit is reached
		if h.AITracker.IsLimitReached() {
			log.Printf("AI usage limit reached, falling back to Google Translate")
			// Fallback to Google Translate
			googleTranslator := translation.NewGoogleFreeTranslatorWithDB(h.DB)
			translatedText, err = translation.TranslateMarkdownPreservingStructure(req.Text, googleTranslator, req.TargetLang)
		} else {
			// Apply rate limiting for AI requests
			h.AITracker.WaitForRateLimit()

			// Use markdown-preserving translation for better list structure
			translatedText, err = translation.TranslateMarkdownAIPrompt(req.Text, h.Translator, req.TargetLang)

			// If AI fails, fallback to Google Translate
			if err != nil {
				log.Printf("AI translation failed, falling back to Google Translate: %v", err)
				googleTranslator := translation.NewGoogleFreeTranslatorWithDB(h.DB)
				translatedText, err = translation.TranslateMarkdownPreservingStructure(req.Text, googleTranslator, req.TargetLang)
			}

			// Track AI usage only on success (whether AI or fallback)
			if err == nil {
				h.AITracker.TrackTranslation(req.Text, translatedText)
			}
		}
	} else {
		// Non-AI provider, use markdown-preserving translation
		translatedText, err = translation.TranslateMarkdownPreservingStructure(req.Text, h.Translator, req.TargetLang)
	}

	if err != nil {
		log.Printf("Error translating text: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Step 3: Post-translation check - if translation equals original, it was already in target language
	// This provides a safety net in case pre-translation detection was inaccurate
	if translatedText == req.Text {
		htmlText := textutil.ConvertMarkdownToHTML(translatedText)
		response.JSON(w, map[string]string{
			"translated_text": translatedText,
			"html":            htmlText,
			"skipped":         "true", // Indicate no actual translation was performed
		})
		return
	}

	// Convert translated markdown to HTML
	htmlText := textutil.ConvertMarkdownToHTML(translatedText)

	response.JSON(w, map[string]string{
		"translated_text": translatedText,
		"html":            htmlText,
		"skipped":         "false", // Translation was performed
	})
}

// HandleResetAIUsage resets the AI usage counter.
// @Summary      Reset AI usage counter
// @Description  Reset the AI usage token counter to zero
// @Tags         translation
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]bool  "Success status"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /ai/usage/reset [post]
func HandleResetAIUsage(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	if err := h.AITracker.ResetUsage(); err != nil {
		log.Printf("Error resetting AI usage: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	response.JSON(w, map[string]bool{"success": true})
}

// HandleGetAIUsage returns the current AI usage statistics.
// @Summary      Get AI usage statistics
// @Description  Get current AI usage (tokens used, limit, and whether limit is reached)
// @Tags         translation
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "AI usage stats (usage, limit, limit_reached)"
// @Router       /ai/usage [get]
func HandleGetAIUsage(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	usage, _ := h.AITracker.GetCurrentUsage()
	limit, _ := h.AITracker.GetUsageLimit()

	response.JSON(w, map[string]interface{}{
		"usage":         usage,
		"limit":         limit,
		"limit_reached": h.AITracker.IsLimitReached(),
	})
}

// EstimateTokens exposes the token estimation function for testing/display.
func EstimateTokens(text string) int64 {
	return ai.EstimateTokens(text)
}

// HandleTestCustomTranslation tests a custom translation configuration.
// @Summary      Test custom translation
// @Description  Test a custom translation API configuration
// @Tags         translation
// @Accept       json
// @Produce      json
// @Param        request  body      TestCustomTranslationRequest  true  "Test request"
// @Success      200  {object}  TestCustomTranslationResponse  "Test result"
// @Failure      400  {object}  map[string]string  "Bad request"
// @Router       /translation/test-custom [post]
func HandleTestCustomTranslation(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Text   string                             `json:"text"`
		Target string                             `json:"target_lang"`
		Config translation.CustomTranslatorConfig `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Text == "" {
		req.Text = "Hello, world!"
	}
	if req.Target == "" {
		req.Target = "zh"
	}

	// Create custom translator
	customTranslator := translation.NewCustomTranslator(&req.Config)

	// Test translation
	result, err := customTranslator.Translate(req.Text, req.Target)

	resp := map[string]interface{}{
		"success": err == nil,
	}

	if err != nil {
		resp["error"] = err.Error()
		w.WriteHeader(http.StatusBadRequest)
	} else {
		resp["translation"] = result
	}

	response.JSON(w, resp)
}

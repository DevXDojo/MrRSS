package article

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/models"
)

// ExportToZoteroRequest represents the request for exporting to Zotero
type ExportToZoteroRequest struct {
	ArticleID int `json:"article_id"`
}

// ZoteroItem represents a Zotero item structure
type ZoteroItem struct {
	ItemType     string          `json:"itemType"`
	Title        string          `json:"title"`
	URL          string          `json:"url,omitempty"`
	AbstractNote string          `json:"abstractNote,omitempty"`
	WebsiteTitle string          `json:"websiteTitle,omitempty"`
	AccessDate   string          `json:"accessDate"`
	DateAdded    string          `json:"dateAdded"`
	DateModified string          `json:"dateModified"`
	Tags         []ZoteroTag     `json:"tags,omitempty"`
	Creators     []ZoteroCreator `json:"creators,omitempty"`
	Note         string          `json:"note,omitempty"`
}

// ZoteroTag represents a Zotero tag
type ZoteroTag struct {
	Tag string `json:"tag"`
}

// ZoteroCreator represents a Zotero creator (author, editor, etc.)
type ZoteroCreator struct {
	CreatorType string `json:"creatorType"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Name        string `json:"name,omitempty"` // Used for single-field name
}

// ZoteroResponse represents the response from Zotero API
type ZoteroResponse struct {
	Success   map[string]json.RawMessage `json:"successful"`
	Unchanged map[string]string          `json:"unchanged"`
	Failed    map[string]ZoteroError     `json:"failed"`
	Key       string                     `json:"key,omitempty"` // Item key if single item
	Version   int                        `json:"version,omitempty"`
}

// ZoteroError represents a Zotero API error
type ZoteroError struct {
	Key     string `json:"key"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// HandleExportToZotero exports an article to Zotero using Zotero API
// @Summary      Export article to Zotero
// @Description  Export an article to Zotero as a webpage item (requires zotero_enabled, zotero_api_key, and zotero_user_id settings)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        request  body      ExportToZoteroRequest  true  "Article export request"
// @Success      200  {object}  map[string]string  "Export result (success, item_key, message)"
// @Failure      400  {object}  map[string]string  "Bad request (Zotero not configured or invalid article ID)"
// @Failure      404  {object}  map[string]string  "Article not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/export/zotero [post]
func HandleExportToZotero(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req ExportToZoteroRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	if req.ArticleID <= 0 {
		response.Error(w, fmt.Errorf("invalid article ID"), http.StatusBadRequest)
		return
	}

	// Get article from database
	article, err := h.DB.GetArticleByID(int64(req.ArticleID))
	if err != nil {
		response.Error(w, err, http.StatusNotFound)
		return
	}

	// Check if Zotero integration is enabled
	zoteroEnabled, _ := h.DB.GetSetting("zotero_enabled")
	if zoteroEnabled != "true" {
		response.Error(w, fmt.Errorf("zotero integration is not enabled"), http.StatusBadRequest)
		return
	}

	// Get API key (encrypted setting)
	apiKey, err := h.DB.GetEncryptedSetting("zotero_api_key")
	if err != nil || apiKey == "" {
		response.Error(w, fmt.Errorf("zotero API key is not configured"), http.StatusBadRequest)
		return
	}

	// Get user ID
	userID, _ := h.DB.GetSetting("zotero_user_id")
	if userID == "" {
		response.Error(w, fmt.Errorf("zotero user ID is not configured"), http.StatusBadRequest)
		return
	}

	// Get article content
	content, _, err := h.GetArticleContent(int64(req.ArticleID))
	if err != nil {
		// If content fetch fails, continue with empty content
		content = ""
	}

	// Generate Zotero item
	zoteroItem := generateZoteroItem(*article, content)

	// Send request to Zotero API
	itemKey, zoteroURL, err := createZoteroItem(apiKey, userID, zoteroItem)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Return success response
	response.JSON(w, map[string]string{
		"success":  "true",
		"item_key": itemKey,
		"item_url": zoteroURL,
		"message":  "Article exported to Zotero successfully",
	})
}

// generateZoteroItem converts an article to Zotero item format
func generateZoteroItem(article models.Article, content string) ZoteroItem {
	now := time.Now().Format(time.RFC3339)

	// Clean up content - remove HTML tags and limit length
	abstractNote := cleanContentForZotero(content)

	// Generate tags from feed name
	tags := []ZoteroTag{
		{Tag: "RSS"},
		{Tag: sanitizeTagForZotero(article.FeedTitle)},
	}

	// Create the item
	item := ZoteroItem{
		ItemType:     "webpage",
		Title:        article.Title,
		URL:          article.URL,
		WebsiteTitle: article.FeedTitle,
		AbstractNote: abstractNote,
		AccessDate:   now,
		DateAdded:    now,
		DateModified: now,
		Tags:         tags,
	}

	return item
}

// cleanContentForZotero removes HTML tags and limits content length
func cleanContentForZotero(htmlContent string) string {
	// Decode HTML entities
	decodedContent := html.UnescapeString(htmlContent)

	// Remove HTML tags
	text := removeHTMLTags(decodedContent)

	// Clean whitespace
	text = cleanWhitespace(text)

	// Limit to 20000 characters (Zotero has limits)
	if len(text) > 20000 {
		text = text[:19997] + "..."
	}

	return text
}

// sanitizeTagForZotero creates a safe tag from feed name
func sanitizeTagForZotero(feedName string) string {
	// Convert to lowercase, replace spaces with underscores
	tag := strings.ToLower(strings.ReplaceAll(feedName, " ", "_"))
	// Remove special characters
	tag = strings.ReplaceAll(tag, "-", "_")
	tag = strings.ReplaceAll(tag, ".", "_")
	// Limit length
	if len(tag) > 50 {
		tag = tag[:50]
	}
	return tag
}

// createZoteroItem sends a request to Zotero API to create an item
// Returns (itemKey, itemURL, error)
func createZoteroItem(apiKey string, userID string, item ZoteroItem) (string, string, error) {
	// Create JSON array with single item
	items := []ZoteroItem{item}
	jsonBody, err := json.Marshal(items)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL for user's library
	url := fmt.Sprintf("https://api.zotero.org/users/%s/items", userID)

	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Zotero-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		var zoteroResp ZoteroResponse
		if err := json.NewDecoder(resp.Body).Decode(&zoteroResp); err == nil {
			// Check for specific errors
			for _, zerr := range zoteroResp.Failed {
				return "", "", fmt.Errorf("zotero API error: %s (code: %d)", zerr.Message, zerr.Code)
			}
		}
		return "", "", fmt.Errorf("zotero API returned status %d", resp.StatusCode)
	}

	// Parse response
	var zoteroResp ZoteroResponse
	if err := json.NewDecoder(resp.Body).Decode(&zoteroResp); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if creation was successful
	if len(zoteroResp.Success) == 0 {
		return "", "", fmt.Errorf("item was not created successfully")
	}

	// Get the item key from the response
	// The success map contains "0" -> {"key": "ABC123", "version": 1234}
	var successData map[string]interface{}
	for _, data := range zoteroResp.Success {
		if err := json.Unmarshal(data, &successData); err == nil {
			if key, ok := successData["key"].(string); ok {
				itemKey := key
				itemURL := fmt.Sprintf("https://zotero.org/users/%s/items/%s", userID, itemKey)
				return itemKey, itemURL, nil
			}
		}
	}

	return "", "", fmt.Errorf("failed to get item key from response")
}

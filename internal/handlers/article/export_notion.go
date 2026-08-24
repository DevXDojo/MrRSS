package article

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/models"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

// ExportToNotionRequest represents the request for exporting to Notion
type ExportToNotionRequest struct {
	ArticleID int `json:"article_id"`
}

// NotionBlock represents a Notion block structure
type NotionBlock struct {
	Object           string            `json:"object"`
	Type             string            `json:"type"`
	Paragraph        *Paragraph        `json:"paragraph,omitempty"`
	Heading1         *Heading          `json:"heading_1,omitempty"`
	Heading2         *Heading          `json:"heading_2,omitempty"`
	Heading3         *Heading          `json:"heading_3,omitempty"`
	Divider          *Divider          `json:"divider,omitempty"`
	Bookmark         *Bookmark         `json:"bookmark,omitempty"`
	Quote            *Quote            `json:"quote,omitempty"`
	BulletedListItem *BulletedListItem `json:"bulleted_list_item,omitempty"`
	NumberedListItem *NumberedListItem `json:"numbered_list_item,omitempty"`
	Code             *Code             `json:"code,omitempty"`
	Image            *ImageBlock       `json:"image,omitempty"`
}

// Paragraph represents a Notion paragraph block
type Paragraph struct {
	RichText []RichText `json:"rich_text"`
}

// Heading represents a Notion heading block
type Heading struct {
	RichText []RichText `json:"rich_text"`
}

// Divider represents a Notion divider block
type Divider struct{}

// Bookmark represents a Notion bookmark block
type Bookmark struct {
	URL     string     `json:"url"`
	Caption []RichText `json:"caption,omitempty"`
}

// Quote represents a Notion quote block
type Quote struct {
	RichText []RichText `json:"rich_text"`
}

// BulletedListItem represents a Notion bulleted list item block
type BulletedListItem struct {
	RichText []RichText `json:"rich_text"`
}

// NumberedListItem represents a Notion numbered list item block
type NumberedListItem struct {
	RichText []RichText `json:"rich_text"`
}

// Code represents a Notion code block
type Code struct {
	RichText []RichText `json:"rich_text"`
	Language string     `json:"language"`
}

// ImageBlock represents a Notion image block
type ImageBlock struct {
	Type     string        `json:"type"`
	External *ExternalFile `json:"external,omitempty"`
}

// ExternalFile represents an external file URL
type ExternalFile struct {
	URL string `json:"url"`
}

// RichText represents rich text in Notion
type RichText struct {
	Type        string       `json:"type"`
	Text        TextData     `json:"text"`
	Annotations *Annotations `json:"annotations,omitempty"`
}

// TextData represents text content
type TextData struct {
	Content string `json:"content"`
	Link    *Link  `json:"link,omitempty"`
}

// Link represents a link in Notion
type Link struct {
	URL string `json:"url"`
}

// Annotations represents text annotations (bold, italic, etc.)
type Annotations struct {
	Bold          bool   `json:"bold,omitempty"`
	Italic        bool   `json:"italic,omitempty"`
	Strikethrough bool   `json:"strikethrough,omitempty"`
	Underline     bool   `json:"underline,omitempty"`
	Code          bool   `json:"code,omitempty"`
	Color         string `json:"color,omitempty"`
}

// NotionPageRequest represents the request body for creating a Notion page
type NotionPageRequest struct {
	Parent     NotionParent              `json:"parent"`
	Properties map[string]NotionProperty `json:"properties"`
	Children   []NotionBlock             `json:"children,omitempty"`
}

// NotionParent represents the parent of a page
type NotionParent struct {
	PageID string `json:"page_id,omitempty"`
}

// NotionProperty represents a Notion property
type NotionProperty struct {
	Title []RichText `json:"title,omitempty"`
}

// NotionResponse represents the response from Notion API
type NotionResponse struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Object  string `json:"object"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// AppendBlocksRequest represents the request body for appending blocks to a page
type AppendBlocksRequest struct {
	Children []NotionBlock `json:"children"`
}

// HandleExportToNotion exports an article to Notion using Notion API
// @Summary      Export article to Notion
// @Description  Export an article to Notion as a page (requires notion_enabled, notion_api_key, and notion_page_id settings)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        request  body      ExportToNotionRequest  true  "Article export request"
// @Success      200  {object}  map[string]string  "Export result (success, page_url, message)"
// @Failure      400  {object}  map[string]string  "Bad request (Notion not configured or invalid article ID)"
// @Failure      404  {object}  map[string]string  "Article not found"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /articles/export/notion [post]
func HandleExportToNotion(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	var req ExportToNotionRequest
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

	// Check if Notion integration is enabled
	notionEnabled, _ := h.DB.GetSetting("notion_enabled")
	if notionEnabled != "true" {
		response.Error(w, fmt.Errorf("notion integration is not enabled"), http.StatusBadRequest)
		return
	}

	// Get API key (encrypted setting)
	apiKey, _ := h.DB.GetEncryptedSetting("notion_api_key")
	if apiKey == "" {
		response.Error(w, fmt.Errorf("notion API key is not configured"), http.StatusBadRequest)
		return
	}

	// Get parent page ID
	pageID, _ := h.DB.GetSetting("notion_page_id")
	if pageID == "" {
		response.Error(w, fmt.Errorf("notion page ID is not configured"), http.StatusBadRequest)
		return
	}

	// Normalize page ID (remove hyphens if present)
	pageID = strings.ReplaceAll(pageID, "-", "")

	// Get article content
	content, _, err := h.GetArticleContent(int64(req.ArticleID))
	if err != nil {
		// If content fetch fails, continue with empty content
		content = ""
	}

	// Convert content to Notion blocks
	contentBlocks := htmlToNotionBlocks(content)

	// Build initial page with metadata (max 100 blocks including metadata)
	metadataBlocks := buildMetadataBlocks(*article)
	initialBlocks := metadataBlocks

	// Calculate how many content blocks we can add to initial request
	remainingSlots := 100 - len(metadataBlocks)
	if len(contentBlocks) <= remainingSlots {
		initialBlocks = append(initialBlocks, contentBlocks...)
		contentBlocks = nil
	} else {
		initialBlocks = append(initialBlocks, contentBlocks[:remainingSlots]...)
		contentBlocks = contentBlocks[remainingSlots:]
	}

	// Create the page with initial blocks
	notionRequest := NotionPageRequest{
		Parent: NotionParent{
			PageID: pageID,
		},
		Properties: map[string]NotionProperty{
			"title": {
				Title: []RichText{
					{Type: "text", Text: TextData{Content: article.Title}},
				},
			},
		},
		Children: initialBlocks,
	}

	// Send request to Notion API to create the page
	pageURL, createdPageID, err := createNotionPage(apiKey, notionRequest)
	if err != nil {
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// If there are remaining content blocks, append them in batches
	if len(contentBlocks) > 0 {
		err = appendBlocksInBatches(apiKey, createdPageID, contentBlocks)
		if err != nil {
			// Page was created but some content failed to append
			// Still return success but mention the issue
			response.JSON(w, map[string]string{
				"success":  "true",
				"page_url": pageURL,
				"message":  fmt.Sprintf("Article exported but some content may be missing: %v", err),
			})
			return
		}
	}

	// Return success response
	response.JSON(w, map[string]string{
		"success":  "true",
		"page_url": pageURL,
		"message":  "Article exported to Notion successfully",
	})
}

// buildMetadataBlocks creates metadata blocks for the article
func buildMetadataBlocks(article models.Article) []NotionBlock {
	blocks := []NotionBlock{}

	// Add source URL as bookmark
	if article.URL != "" {
		blocks = append(blocks, NotionBlock{
			Object: "block",
			Type:   "bookmark",
			Bookmark: &Bookmark{
				URL: article.URL,
				Caption: []RichText{
					{Type: "text", Text: TextData{Content: "Original Article"}},
				},
			},
		})
	}

	// Add metadata as quote block
	metadataText := fmt.Sprintf("Feed: %s\nPublished: %s\nExported: %s",
		article.FeedTitle,
		article.PublishedAt.Format("2006-01-02 15:04:05"),
		time.Now().Format("2006-01-02 15:04:05"),
	)
	blocks = append(blocks, NotionBlock{
		Object: "block",
		Type:   "quote",
		Quote: &Quote{
			RichText: []RichText{
				{Type: "text", Text: TextData{Content: metadataText}},
			},
		},
	})

	// Add divider
	blocks = append(blocks, NotionBlock{
		Object:  "block",
		Type:    "divider",
		Divider: &Divider{},
	})

	return blocks
}

// htmlToNotionBlocks converts HTML content to Notion blocks with proper formatting
func htmlToNotionBlocks(htmlContent string) []NotionBlock {
	if htmlContent == "" {
		return []NotionBlock{}
	}

	// Convert HTML to Markdown first (preserves formatting)
	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(htmlContent)
	if err != nil {
		// Fallback to plain text
		return []NotionBlock{{
			Object: "block",
			Type:   "paragraph",
			Paragraph: &Paragraph{
				RichText: []RichText{{Type: "text", Text: TextData{Content: htmlContent}}},
			},
		}}
	}

	// Parse Markdown into Notion blocks
	return markdownToNotionBlocks(markdown)
}

// markdownToNotionBlocks converts Markdown text to Notion blocks
func markdownToNotionBlocks(markdown string) []NotionBlock {
	blocks := []NotionBlock{}
	lines := strings.Split(markdown, "\n")

	// Regex patterns - updated to handle optional leading whitespace
	h1Pattern := regexp.MustCompile(`^#\s+(.+)$`)
	h2Pattern := regexp.MustCompile(`^##\s+(.+)$`)
	h3Pattern := regexp.MustCompile(`^###\s+(.+)$`)
	bulletPattern := regexp.MustCompile(`^\s*[\*\-\+]\s+(.+)$`)
	numberPattern := regexp.MustCompile(`^\s*\d+\.\s+(.+)$`)
	codeBlockStart := regexp.MustCompile("^```(\\w*)$")
	codeBlockEnd := regexp.MustCompile("^```$")
	imagePattern := regexp.MustCompile(`^!\[([^\]]*)\]\(([^)]+)\)$`)
	blockquotePattern := regexp.MustCompile(`^>\s*(.*)$`)

	inCodeBlock := false
	codeLanguage := ""
	codeContent := []string{}
	paragraphBuffer := []string{}

	flushParagraph := func() {
		if len(paragraphBuffer) > 0 {
			text := strings.Join(paragraphBuffer, "\n")
			text = strings.TrimSpace(text)
			if text != "" {
				blocks = append(blocks, createParagraphBlock(text))
			}
			paragraphBuffer = []string{}
		}
	}

	for _, line := range lines {
		// Handle code blocks
		if codeBlockStart.MatchString(line) && !inCodeBlock {
			flushParagraph()
			matches := codeBlockStart.FindStringSubmatch(line)
			codeLanguage = matches[1]
			if codeLanguage == "" {
				codeLanguage = "plain text"
			}
			inCodeBlock = true
			codeContent = []string{}
			continue
		}

		if codeBlockEnd.MatchString(line) && inCodeBlock {
			code := strings.Join(codeContent, "\n")
			blocks = append(blocks, createCodeBlock(code, codeLanguage))
			inCodeBlock = false
			codeLanguage = ""
			codeContent = []string{}
			continue
		}

		if inCodeBlock {
			codeContent = append(codeContent, line)
			continue
		}

		// Skip empty lines but flush paragraph
		if strings.TrimSpace(line) == "" {
			flushParagraph()
			continue
		}

		// Handle headings
		if matches := h1Pattern.FindStringSubmatch(line); matches != nil {
			flushParagraph()
			blocks = append(blocks, createHeading1Block(matches[1]))
			continue
		}

		if matches := h2Pattern.FindStringSubmatch(line); matches != nil {
			flushParagraph()
			blocks = append(blocks, createHeading2Block(matches[1]))
			continue
		}

		if matches := h3Pattern.FindStringSubmatch(line); matches != nil {
			flushParagraph()
			blocks = append(blocks, createHeading3Block(matches[1]))
			continue
		}

		// Handle images
		if matches := imagePattern.FindStringSubmatch(line); matches != nil {
			flushParagraph()
			imageURL := matches[2]
			blocks = append(blocks, createImageBlock(imageURL))
			continue
		}

		// Handle bullet lists
		if matches := bulletPattern.FindStringSubmatch(line); matches != nil {
			flushParagraph()
			blocks = append(blocks, createBulletedListItem(matches[1]))
			continue
		}

		// Handle numbered lists
		if matches := numberPattern.FindStringSubmatch(line); matches != nil {
			flushParagraph()
			blocks = append(blocks, createNumberedListItem(matches[1]))
			continue
		}

		// Handle blockquotes
		if matches := blockquotePattern.FindStringSubmatch(line); matches != nil {
			flushParagraph()
			blocks = append(blocks, createQuoteBlock(matches[1]))
			continue
		}

		// Regular text - add to paragraph buffer
		paragraphBuffer = append(paragraphBuffer, line)
	}

	// Flush any remaining paragraph content
	flushParagraph()

	// Handle unclosed code block
	if inCodeBlock && len(codeContent) > 0 {
		code := strings.Join(codeContent, "\n")
		blocks = append(blocks, createCodeBlock(code, codeLanguage))
	}

	return blocks
}

// createParagraphBlock creates a paragraph block with rich text formatting
func createParagraphBlock(text string) NotionBlock {
	richTexts := parseRichText(text)
	return NotionBlock{
		Object:    "block",
		Type:      "paragraph",
		Paragraph: &Paragraph{RichText: richTexts},
	}
}

// createHeading1Block creates a heading 1 block
func createHeading1Block(text string) NotionBlock {
	richTexts := parseRichText(text)
	return NotionBlock{
		Object:   "block",
		Type:     "heading_1",
		Heading1: &Heading{RichText: richTexts},
	}
}

// createHeading2Block creates a heading 2 block
func createHeading2Block(text string) NotionBlock {
	richTexts := parseRichText(text)
	return NotionBlock{
		Object:   "block",
		Type:     "heading_2",
		Heading2: &Heading{RichText: richTexts},
	}
}

// createHeading3Block creates a heading 3 block
func createHeading3Block(text string) NotionBlock {
	richTexts := parseRichText(text)
	return NotionBlock{
		Object:   "block",
		Type:     "heading_3",
		Heading3: &Heading{RichText: richTexts},
	}
}

// createBulletedListItem creates a bulleted list item block
func createBulletedListItem(text string) NotionBlock {
	richTexts := parseRichText(text)
	return NotionBlock{
		Object:           "block",
		Type:             "bulleted_list_item",
		BulletedListItem: &BulletedListItem{RichText: richTexts},
	}
}

// createNumberedListItem creates a numbered list item block
func createNumberedListItem(text string) NotionBlock {
	richTexts := parseRichText(text)
	return NotionBlock{
		Object:           "block",
		Type:             "numbered_list_item",
		NumberedListItem: &NumberedListItem{RichText: richTexts},
	}
}

// createCodeBlock creates a code block
func createCodeBlock(code string, language string) NotionBlock {
	// Notion limits rich_text to 2000 chars, split if needed
	chunks := splitIntoChunks(code, 1900)
	richTexts := make([]RichText, len(chunks))
	for i, chunk := range chunks {
		richTexts[i] = RichText{Type: "text", Text: TextData{Content: chunk}}
	}
	return NotionBlock{
		Object: "block",
		Type:   "code",
		Code:   &Code{RichText: richTexts, Language: language},
	}
}

// createImageBlock creates an image block
func createImageBlock(url string) NotionBlock {
	return NotionBlock{
		Object: "block",
		Type:   "image",
		Image: &ImageBlock{
			Type:     "external",
			External: &ExternalFile{URL: url},
		},
	}
}

// createQuoteBlock creates a quote block
func createQuoteBlock(text string) NotionBlock {
	richTexts := parseRichText(text)
	return NotionBlock{
		Object: "block",
		Type:   "quote",
		Quote:  &Quote{RichText: richTexts},
	}
}

// parseRichText parses Markdown inline formatting (bold, italic, code, links) into RichText array
func parseRichText(text string) []RichText {
	// Split text if it's too long (Notion limit: 2000 chars per rich_text)
	if len(text) > 1900 {
		chunks := splitIntoChunks(text, 1900)
		richTexts := make([]RichText, len(chunks))
		for i, chunk := range chunks {
			richTexts[i] = RichText{Type: "text", Text: TextData{Content: chunk}}
		}
		return richTexts
	}

	// If text is empty, return single empty RichText
	if text == "" {
		return []RichText{{Type: "text", Text: TextData{Content: ""}}}
	}

	// Process all inline formatting using a single pass approach
	// This handles **bold**, *italic*, `code`, and [text](url)
	result := []RichText{}

	// Combined pattern for all inline elements
	// Order matters: bold before italic to handle **text** correctly
	combinedPattern := regexp.MustCompile(`(\*\*(.+?)\*\*)|(\*([^*]+)\*)|(` + "`" + `([^` + "`" + `]+)` + "`" + `)|(\[([^\]]+)\]\(([^)]+)\))`)

	lastIndex := 0
	matches := combinedPattern.FindAllStringSubmatchIndex(text, -1)

	for _, match := range matches {
		// Add text before this match
		if match[0] > lastIndex {
			result = append(result, RichText{
				Type: "text",
				Text: TextData{Content: text[lastIndex:match[0]]},
			})
		}

		// Determine which pattern matched
		if match[2] != -1 && match[3] != -1 {
			// Bold: **text**
			boldContent := text[match[4]:match[5]]
			result = append(result, RichText{
				Type:        "text",
				Text:        TextData{Content: boldContent},
				Annotations: &Annotations{Bold: true},
			})
		} else if match[6] != -1 && match[7] != -1 {
			// Italic: *text*
			italicContent := text[match[8]:match[9]]
			result = append(result, RichText{
				Type:        "text",
				Text:        TextData{Content: italicContent},
				Annotations: &Annotations{Italic: true},
			})
		} else if match[10] != -1 && match[11] != -1 {
			// Inline code: `code`
			codeContent := text[match[12]:match[13]]
			result = append(result, RichText{
				Type:        "text",
				Text:        TextData{Content: codeContent},
				Annotations: &Annotations{Code: true},
			})
		} else if match[14] != -1 && match[15] != -1 {
			// Link: [text](url)
			linkText := text[match[16]:match[17]]
			linkURL := text[match[18]:match[19]]
			result = append(result, RichText{
				Type: "text",
				Text: TextData{Content: linkText, Link: &Link{URL: linkURL}},
			})
		}

		lastIndex = match[1]
	}

	// Add remaining text after last match
	if lastIndex < len(text) {
		result = append(result, RichText{
			Type: "text",
			Text: TextData{Content: text[lastIndex:]},
		})
	}

	// Ensure we have at least one RichText element
	if len(result) == 0 {
		result = append(result, RichText{Type: "text", Text: TextData{Content: text}})
	}

	return result
}

// appendBlocksInBatches appends blocks to a page in batches of 100
func appendBlocksInBatches(apiKey string, pageID string, blocks []NotionBlock) error {
	const batchSize = 100

	for i := 0; i < len(blocks); i += batchSize {
		end := i + batchSize
		if end > len(blocks) {
			end = len(blocks)
		}

		batch := blocks[i:end]
		appendReq := AppendBlocksRequest{
			Children: batch,
		}

		jsonBody, err := json.Marshal(appendReq)
		if err != nil {
			return fmt.Errorf("failed to marshal append request: %w", err)
		}

		// Use PATCH /v1/blocks/{block_id}/children endpoint
		url := fmt.Sprintf("https://api.notion.com/v1/blocks/%s/children", pageID)
		req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("failed to create append request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Notion-Version", "2022-06-28")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to append blocks: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			var notionResp NotionResponse
			if err := json.NewDecoder(resp.Body).Decode(&notionResp); err == nil && notionResp.Message != "" {
				return fmt.Errorf("notion API error when appending: %s (code: %s)", notionResp.Message, notionResp.Code)
			}
			return fmt.Errorf("notion API returned status %d when appending blocks", resp.StatusCode)
		}
	}

	return nil
}

// splitIntoChunks splits text into chunks of maxLen characters
func splitIntoChunks(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	for len(text) > 0 {
		if len(text) <= maxLen {
			chunks = append(chunks, text)
			break
		}

		// Find a good break point (space, newline)
		breakPoint := maxLen
		for i := maxLen - 1; i > maxLen-200 && i > 0; i-- {
			if text[i] == ' ' || text[i] == '\n' {
				breakPoint = i
				break
			}
		}

		chunks = append(chunks, strings.TrimSpace(text[:breakPoint]))
		text = strings.TrimSpace(text[breakPoint:])
	}

	return chunks
}

// createNotionPage sends a request to Notion API to create a page
// Returns (pageURL, pageID, error)
func createNotionPage(apiKey string, pageRequest NotionPageRequest) (string, string, error) {
	jsonBody, err := json.Marshal(pageRequest)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.notion.com/v1/pages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var notionResp NotionResponse
	if err := json.NewDecoder(resp.Body).Decode(&notionResp); err != nil {
		return "", "", fmt.Errorf("failed to decode response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if notionResp.Message != "" {
			return "", "", fmt.Errorf("notion API error: %s (code: %s)", notionResp.Message, notionResp.Code)
		}
		return "", "", fmt.Errorf("notion API returned status %d", resp.StatusCode)
	}

	return notionResp.URL, notionResp.ID, nil
}

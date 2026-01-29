package opml

import (
	"MrRSS/internal/models"
	"MrRSS/internal/utils"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"regexp"
	"strings"
)

type OPML struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    Head     `xml:"head"`
	Body    Body     `xml:"body"`
}

type Head struct {
	Title string `xml:"title"`
}

type Body struct {
	Outlines []*Outline `xml:"outline"`
}

// Outline represents an OPML outline element with various attribute formats
// Different OPML exporters use different case for attributes (xmlUrl vs xmlurl vs XmlUrl)
type Outline struct {
	Text     string     `xml:"text,attr"`
	Title    string     `xml:"title,attr"`
	Type     string     `xml:"type,attr"`
	XMLURL   string     `xml:"xmlUrl,attr"`
	HTMLURL  string     `xml:"htmlUrl,attr"`
	Outlines []*Outline `xml:"outline"` // Nested outlines
	// Additional attributes for compatibility with various OPML formats
	Description string `xml:"description,attr"`
	Category    string `xml:"category,attr"`
	// FreshRSS XPath extension attributes
	XPathItem           string `xml:"xPathItem,attr"`
	XPathItemTitle      string `xml:"xPathItemTitle,attr"`
	XPathItemContent    string `xml:"xPathItemContent,attr"`
	XPathItemUri        string `xml:"xPathItemUri,attr"`
	XPathItemAuthor     string `xml:"xPathItemAuthor,attr"`
	XPathItemTimestamp  string `xml:"xPathItemTimestamp,attr"`
	XPathItemTimeFormat string `xml:"xPathItemTimeFormat,attr"`
	XPathItemThumbnail  string `xml:"xPathItemThumbnail,attr"`
	XPathItemCategories string `xml:"xPathItemCategories,attr"`
	XPathItemUid        string `xml:"xPathItemUid,attr"`
}

// normalizeOPMLAttributes normalizes attribute names in OPML content to handle
// case variations (xmlUrl, xmlurl, XmlUrl, etc.)
func normalizeOPMLAttributes(content []byte) []byte {
	// Common attribute name variations that need normalization
	patterns := []struct {
		pattern *regexp.Regexp
		replace string
	}{
		// xmlUrl variations
		{regexp.MustCompile(`(?i)\sxmlurl=`), ` xmlUrl=`},
		// htmlUrl variations
		{regexp.MustCompile(`(?i)\shtmlurl=`), ` htmlUrl=`},
	}

	result := content
	for _, p := range patterns {
		result = p.pattern.ReplaceAll(result, []byte(p.replace))
	}
	return result
}

func Parse(r io.Reader) ([]models.Feed, error) {
	// Read all content to handle BOM
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	utils.DebugLog("OPML Parse: Read %d bytes", len(content))

	if len(content) == 0 {
		return nil, errors.New("file content is empty")
	}

	// Strip UTF-8 BOM if present
	content = bytes.TrimPrefix(content, []byte("\xef\xbb\xbf"))

	// Strip UTF-16 LE BOM if present
	content = bytes.TrimPrefix(content, []byte("\xff\xfe"))

	// Strip UTF-16 BE BOM if present
	content = bytes.TrimPrefix(content, []byte("\xfe\xff"))

	// Normalize attribute names for compatibility
	content = normalizeOPMLAttributes(content)

	// Try to fix common XML issues
	content = fixCommonXMLIssues(content)

	var doc OPML
	decoder := xml.NewDecoder(bytes.NewReader(content))
	// Be lenient with character encoding
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil
	}

	if err := decoder.Decode(&doc); err != nil {
		log.Printf("OPML Parse: Decode error: %v", err)
		// Try fallback parsing for malformed OPML
		feeds := fallbackParse(content)
		if len(feeds) > 0 {
			utils.DebugLog("OPML Parse: Fallback parsing found %d feeds", len(feeds))
			return feeds, nil
		}
		return nil, err
	}

	var feeds []models.Feed
	var extract func([]*Outline, string)
	extract = func(outlines []*Outline, category string) {
		for i := 0; i < len(outlines); i++ {
			o := outlines[i]
			xmlURL := strings.TrimSpace(o.XMLURL)
			if xmlURL != "" {
				title := strings.TrimSpace(o.Title)
				if title == "" {
					title = strings.TrimSpace(o.Text)
				}
				if title == "" {
					title = "Untitled Feed"
				}
				// Use outline's category attribute if present, otherwise use hierarchy
				feedCategory := category
				if o.Category != "" && !strings.HasPrefix(o.Category, "tag:") {
					feedCategory = strings.TrimSpace(o.Category)
				}

				// Check if next outline is a tag (sibling element)
				var tags []models.Tag
				for j := i + 1; j < len(outlines); j++ {
					nextOutline := outlines[j]
					// Check if this is a tag (type="tag" or category starts with "tag:")
					isTag := nextOutline.Type == "tag" || strings.HasPrefix(nextOutline.Category, "tag:")
					// Stop if we hit another feed or category (non-tag outline with XMLURL or no XMLURL but not a tag)
					if nextOutline.XMLURL != "" {
						break // Next feed found
					}
					if !isTag && nextOutline.XMLURL == "" && len(nextOutline.Outlines) == 0 {
						// This might be a category folder, check if it has nested outlines
						if len(nextOutline.Outlines) > 0 {
							break // Category folder with children, not a tag
						}
					}
					if isTag {
						tagName := strings.TrimSpace(nextOutline.Text)
						if tagName == "" {
							tagName = strings.TrimSpace(nextOutline.Title)
						}
						// Extract tag name from category if it starts with "tag:"
						if strings.HasPrefix(nextOutline.Category, "tag:") {
							tagName = strings.TrimSpace(strings.TrimPrefix(nextOutline.Category, "tag:"))
						}
						if tagName != "" {
							tags = append(tags, models.Tag{
								Name:  tagName,
								Color: "#3B82F6", // Default color for imported tags
							})
						}
						i = j // Skip the tag outline in the main loop
					} else {
						break // Not a tag, stop processing
					}
				}

				feeds = append(feeds, models.Feed{
					Title:    title,
					URL:      xmlURL,
					Category: feedCategory,
					Tags:     tags,
					// XPath support
					Type:                o.Type,
					XPathItem:           o.XPathItem,
					XPathItemTitle:      o.XPathItemTitle,
					XPathItemContent:    o.XPathItemContent,
					XPathItemUri:        o.XPathItemUri,
					XPathItemAuthor:     o.XPathItemAuthor,
					XPathItemTimestamp:  o.XPathItemTimestamp,
					XPathItemTimeFormat: o.XPathItemTimeFormat,
					XPathItemThumbnail:  o.XPathItemThumbnail,
					XPathItemCategories: o.XPathItemCategories,
					XPathItemUid:        o.XPathItemUid,
				})
			}

			newCategory := category
			if o.XMLURL == "" {
				text := strings.TrimSpace(o.Text)
				if text == "" {
					text = strings.TrimSpace(o.Title)
				}
				if text != "" {
					if newCategory != "" {
						newCategory += "/" + text
					} else {
						newCategory = text
					}
				}
			}

			if len(o.Outlines) > 0 {
				extract(o.Outlines, newCategory)
			}
		}
	}
	extract(doc.Body.Outlines, "")

	utils.DebugLog("OPML Parse: Found %d feeds", len(feeds))
	return feeds, nil
}

// fixCommonXMLIssues attempts to fix common XML formatting issues
func fixCommonXMLIssues(content []byte) []byte {
	// Remove invalid XML characters (control characters except tab, newline, carriage return)
	result := make([]byte, 0, len(content))
	for _, b := range content {
		if b == 0x09 || b == 0x0A || b == 0x0D || (b >= 0x20) {
			result = append(result, b)
		}
	}
	return result
}

// fallbackParse attempts to extract feed URLs using regex when standard parsing fails
func fallbackParse(content []byte) []models.Feed {
	var feeds []models.Feed

	// Pattern to match xmlUrl attributes
	xmlURLPattern := regexp.MustCompile(`(?i)xmlUrl\s*=\s*["']([^"']+)["']`)
	// Pattern to match text or title attributes for feed names
	textPattern := regexp.MustCompile(`(?i)(?:text|title)\s*=\s*["']([^"']+)["']`)

	// Split by outline tags
	outlinePattern := regexp.MustCompile(`(?i)<outline[^>]*>`)
	outlines := outlinePattern.FindAll(content, -1)

	for _, outline := range outlines {
		xmlURLMatch := xmlURLPattern.FindSubmatch(outline)
		if xmlURLMatch == nil {
			continue
		}
		xmlURL := string(xmlURLMatch[1])

		title := "Untitled Feed"
		textMatch := textPattern.FindSubmatch(outline)
		if textMatch != nil {
			title = string(textMatch[1])
		}

		feeds = append(feeds, models.Feed{
			Title: title,
			URL:   xmlURL,
		})
	}

	return feeds
}

func Generate(feeds []models.Feed) ([]byte, error) {
	doc := OPML{
		Version: "1.0",
		Head: Head{
			Title: "MrRSS Subscriptions",
		},
	}

	for _, f := range feeds {
		currentOutlines := &doc.Body.Outlines

		if f.Category != "" {
			parts := strings.Split(f.Category, "/")
			for _, part := range parts {
				var found *Outline
				for _, o := range *currentOutlines {
					if o.XMLURL == "" && o.Text == part {
						found = o
						break
					}
				}
				if found == nil {
					found = &Outline{
						Text:  part,
						Title: part,
					}
					*currentOutlines = append(*currentOutlines, found)
				}
				currentOutlines = &found.Outlines
			}
		}

		*currentOutlines = append(*currentOutlines, &Outline{
			Text:   f.Title,
			Title:  f.Title,
			Type:   f.Type,
			XMLURL: f.URL,
			// XPath support
			XPathItem:           f.XPathItem,
			XPathItemTitle:      f.XPathItemTitle,
			XPathItemContent:    f.XPathItemContent,
			XPathItemUri:        f.XPathItemUri,
			XPathItemAuthor:     f.XPathItemAuthor,
			XPathItemTimestamp:  f.XPathItemTimestamp,
			XPathItemTimeFormat: f.XPathItemTimeFormat,
			XPathItemThumbnail:  f.XPathItemThumbnail,
			XPathItemCategories: f.XPathItemCategories,
			XPathItemUid:        f.XPathItemUid,
		})

		// Add tags as additional outline elements with "tag:" prefix in category
		for _, tag := range f.Tags {
			*currentOutlines = append(*currentOutlines, &Outline{
				Text:     tag.Name,
				Title:    tag.Name,
				Type:     "tag",
				Category: "tag:" + tag.Name,
			})
		}
	}

	return xml.MarshalIndent(doc, "", "  ")
}

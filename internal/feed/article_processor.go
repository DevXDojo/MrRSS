package feed

import (
	"html"
	"net/url"
	"regexp"
	"strings"
	"time"

	"MrRSS/internal/models"
	"MrRSS/internal/utils/textutil"

	"github.com/mmcdole/gofeed"
)

// ExtractContent extracts content from an RSS item with the correct priority order.
// Priority: media:description > item.Content (content:encoded) > item.Description
// This ensures full article content is preferred over summaries.
// This is exported so it can be used by other packages for consistent content extraction.
func ExtractContent(item *gofeed.Item) string {
	// First, try media:description (for YouTube and similar Media RSS feeds)
	mediaDescription := extractMediaDescription(item)
	if mediaDescription != "" {
		return mediaDescription
	}

	// Second, try item.Content (populated from content:encoded or <content>)
	if item.Content != "" {
		return item.Content
	}

	// Finally, fall back to item.Description (usually a summary)
	return item.Description
}

// ArticleWithContent represents an article with its RSS content
type ArticleWithContent struct {
	Article *models.Article
	Content string
}

// processArticles processes RSS feed items and converts them to Article models
// Returns a slice of ArticleWithContent which includes both the article and its content
func (f *Fetcher) processArticles(feed models.Feed, items []*gofeed.Item) []*ArticleWithContent {
	var articlesWithContent []*ArticleWithContent

	for _, item := range items {
		var published time.Time
		var hasValidPublishedTime bool
		if item.PublishedParsed != nil {
			published = *item.PublishedParsed
			hasValidPublishedTime = true
		} else {
			// No publish date in the feed. Use the current time as a fallback so the
			// DB always has a value. The upsert in SaveArticles only writes this to
			// published_at when the article actually changed; for unchanged articles
			// it keeps the originally stored time, so no-pubDate items don't get
			// bumped to "now" on every refresh and stay ordered by first appearance.
			published = time.Now()
			hasValidPublishedTime = false
		}

		imageURL := extractImageURL(item, feed.URL)
		audioURL := extractAudioURL(item)
		videoURL := extractVideoURL(item)

		// Extract Media RSS content (YouTube feeds)
		mediaTitle := extractMediaTitle(item)

		// Extract content from RSS item using centralized extraction logic
		content := ExtractContent(item)

		// Clean HTML to fix malformed tags that can cause rendering issues
		content = textutil.CleanHTML(content)
		originalSummary := textutil.CleanHTML(item.Description)

		// Determine title: prefer media:title if available, then item.Title, then generate from content
		title := item.Title
		if mediaTitle != "" {
			title = mediaTitle
		}
		if title == "" {
			// Fallback to generating from the processed content
			title = generateTitleFromContent(content)
		}

		// IMPORTANT: Translation should NOT be done here during feed refresh!
		// Translation is an expensive operation that should only happen on-demand:
		// 1. When article enters viewport (lazy loading)
		// 2. When user manually clicks translate button
		// Doing it here for all articles during refresh causes massive performance issues
		translatedTitle := "" // Always empty - translation happens on-demand in frontend

		// Extract author information
		author := ""
		if item.Author != nil {
			author = item.Author.Name
		}

		article := &models.Article{
			FeedID:                feed.ID,
			Title:                 title,
			URL:                   item.Link,
			ImageURL:              imageURL,
			AudioURL:              audioURL,
			VideoURL:              videoURL,
			PublishedAt:           published,
			HasValidPublishedTime: hasValidPublishedTime,
			TranslatedTitle:       translatedTitle,
			OriginalSummary:       originalSummary,
			Author:                author,
		}

		articlesWithContent = append(articlesWithContent, &ArticleWithContent{
			Article: article,
			Content: content,
		})
	}

	return articlesWithContent
}

// extractImageURL extracts the image URL from a feed item and resolves relative URLs
func extractImageURL(item *gofeed.Item, feedURL string) string {
	// Try item.Image first
	if item.Image != nil {
		return resolveRelativeURL(item.Image.URL, feedURL)
	}

	// Try Media RSS thumbnail (YouTube feeds use this)
	if thumbnailURL := extractMediaThumbnail(item); thumbnailURL != "" {
		return resolveRelativeURL(thumbnailURL, feedURL)
	}

	// Try enclosures for images (check various image MIME types)
	for _, enc := range item.Enclosures {
		if strings.HasPrefix(enc.Type, "image/") {
			return resolveRelativeURL(enc.URL, feedURL)
		}
	}

	// Fallback: Try to find image in description/content
	content := item.Content
	if content == "" {
		content = item.Description
	}

	re := regexp.MustCompile(`<img[^>]+src="([^">]+)"`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		return resolveRelativeURL(matches[1], feedURL)
	}

	return ""
}

// ResolveRelativeURL converts a relative URL to an absolute URL based on the feed URL
// If the URL is already absolute, it returns it as-is
// It's exported so it can be used by other packages
func ResolveRelativeURL(imageURL string, feedURL string) string {
	if imageURL == "" {
		return ""
	}

	// If it's already an absolute URL (http:// or https://), return as-is
	if strings.HasPrefix(imageURL, "http://") || strings.HasPrefix(imageURL, "https://") {
		return imageURL
	}

	// If feed URL is empty, can't resolve relative URLs
	if feedURL == "" {
		return imageURL
	}

	// Parse the feed URL to get the base URL
	baseURL, err := url.Parse(feedURL)
	if err != nil {
		// If we can't parse the feed URL, return the original image URL
		return imageURL
	}

	// Parse the image URL (which might be relative)
	ref, err := url.Parse(imageURL)
	if err != nil {
		// If we can't parse the image URL, return the original
		return imageURL
	}

	// Resolve the relative URL against the base URL
	return baseURL.ResolveReference(ref).String()
}

// resolveRelativeURL is an internal wrapper for ResolveRelativeURL
// This maintains backward compatibility with internal code
func resolveRelativeURL(imageURL string, feedURL string) string {
	return ResolveRelativeURL(imageURL, feedURL)
}

// ExtractFirstImageURLFromHTML extracts the first image URL from HTML content
// This is used as a fallback when no image metadata is available in RSS/Atom feeds
// It's exported so it can be used by FreshRSS sync and other modules
func ExtractFirstImageURLFromHTML(htmlContent string) string {
	urls := ExtractAllImageURLsFromHTML(htmlContent)
	if len(urls) == 0 {
		return ""
	}

	return urls[0]
}

// ExtractAllImageURLsFromHTML extracts all image URLs from HTML content
// This returns all images found in the content, not just the first one
// It's exported for use by the frontend to build image galleries
func ExtractAllImageURLsFromHTML(htmlContent string) []string {
	if htmlContent == "" {
		return nil
	}

	var urls []string
	seen := make(map[string]struct{})
	imgTagRe := regexp.MustCompile(`(?i)<img\b[^>]*>`)
	attrNames := []string{"data-original", "data-src", "src"}
	imgTags := imgTagRe.FindAllString(htmlContent, -1)

	for _, tag := range imgTags {
		for _, attrName := range attrNames {
			if imageURL := extractHTMLAttribute(tag, attrName); imageURL != "" {
				unescapedURL := html.UnescapeString(imageURL)
				if _, ok := seen[unescapedURL]; !ok {
					urls = append(urls, unescapedURL)
					seen[unescapedURL] = struct{}{}
				}
				break
			}
		}
	}

	return urls
}

func extractHTMLAttribute(tag string, attrName string) string {
	re := regexp.MustCompile(`(?i)\s` + regexp.QuoteMeta(attrName) + `\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s"'=<>]+))`)
	matches := re.FindStringSubmatch(tag)
	for i := 1; i < len(matches); i++ {
		if strings.TrimSpace(matches[i]) != "" {
			return strings.TrimSpace(matches[i])
		}
	}
	return ""
}

// extractAudioURL extracts the audio URL from a feed item (for podcasts)
func extractAudioURL(item *gofeed.Item) string {
	// Try enclosures for audio files
	for _, enc := range item.Enclosures {
		// Check for audio MIME types
		if strings.HasPrefix(enc.Type, "audio/") {
			return enc.URL
		}
	}

	return ""
}

// extractVideoURL extracts the video URL from a feed item (for YouTube and Bilibili videos)
func extractVideoURL(item *gofeed.Item) string {
	// First check if this is a Bilibili video with iframe in content
	// Some RSSHub feeds might include iframe in description/content with complete parameters (aid, cid, bvid)
	// This should take priority over generating a simplified URL from the link
	content := ExtractContent(item)
	if bilibiliURL := extractBilibiliVideoURL(content); bilibiliURL != "" {
		return bilibiliURL
	}

	// Check if this is a Bilibili video link (similar to YouTube detection)
	// Bilibili URLs: https://www.bilibili.com/video/BV...
	if item.Link != "" && strings.Contains(item.Link, "bilibili.com/video/") {
		// Extract BVID from Bilibili URL
		bvid := extractBilibiliBVID(item.Link)
		if bvid != "" {
			// Return embed URL for Bilibili player
			return "https://www.bilibili.com/blackboard/html5mobileplayer.html?bvid=" + bvid + "&autoplay=0"
		}
	}

	// Check if this is a YouTube link (watch, youtu.be, or shorts)
	if item.Link != "" && (strings.Contains(item.Link, "youtube.com/watch") ||
		strings.Contains(item.Link, "youtu.be/") ||
		strings.Contains(item.Link, "youtube.com/shorts/")) {
		// Extract video ID from YouTube URL
		videoID := extractYouTubeVideoID(item.Link)
		if videoID != "" {
			// Return embed URL for YouTube player
			return "https://www.youtube.com/embed/" + videoID
		}
	}

	// Also check for yt:videoId in extensions
	if item.Extensions != nil {
		if ytExt, ok := item.Extensions["yt"]; ok {
			if videoIDExts, ok := ytExt["videoId"]; ok && len(videoIDExts) > 0 {
				videoID := videoIDExts[0].Value
				if videoID != "" {
					return "https://www.youtube.com/embed/" + videoID
				}
			}
		}
	}

	return ""
}

// extractBilibiliVideoURL extracts Bilibili iframe URL from HTML content
func extractBilibiliVideoURL(content string) string {
	if content == "" {
		return ""
	}

	// Look for Bilibili iframe in the content
	// Pattern matches: <iframe src="https://www.bilibili.com/blackboard/html5mobileplayer.html?...">
	// Use (?s) flag to make . match newlines, and use [\s\S] instead of . to match any character including newlines
	re := regexp.MustCompile(`(?s)<iframe[\s\S]+?src=["']([^"']*bilibili\.com/blackboard/html5mobileplayer\.html[^"']*)["']`)
	matches := re.FindStringSubmatch(content)

	if len(matches) > 1 {
		// Unescape HTML entities (e.g., &amp; -> &)
		unescapedURL := html.UnescapeString(matches[1])
		return unescapedURL
	}

	return ""
}

// extractYouTubeVideoID extracts the video ID from a YouTube URL
func extractYouTubeVideoID(url string) string {
	// Handle youtube.com/watch?v=VIDEO_ID
	if strings.Contains(url, "youtube.com/watch") {
		re := regexp.MustCompile(`[?&]v=([^&]+)`)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// Handle youtu.be/VIDEO_ID
	if strings.Contains(url, "youtu.be/") {
		re := regexp.MustCompile(`youtu\.be/([^?&]+)`)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// Handle youtube.com/shorts/VIDEO_ID
	if strings.Contains(url, "youtube.com/shorts/") {
		re := regexp.MustCompile(`shorts/([^?&]+)`)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// extractBilibiliBVID extracts the BVID from a Bilibili URL
// Similar to extractYouTubeVideoID
func extractBilibiliBVID(url string) string {
	// Handle bilibili.com/video/BV...
	if strings.Contains(url, "bilibili.com/video/") {
		re := regexp.MustCompile(`bilibili\.com/video/(BV[\w]+)`)
		matches := re.FindStringSubmatch(url)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

// generateTitleFromContent generates a title from content when title is missing
func generateTitleFromContent(content string) string {
	if content == "" {
		return "Untitled Article"
	}

	// Remove HTML tags
	htmlTagRegex := regexp.MustCompile(`<[^>]+>`)
	plainText := htmlTagRegex.ReplaceAllString(content, "")

	// Trim whitespace
	plainText = strings.TrimSpace(plainText)

	// Limit to 100 characters
	if len(plainText) > 100 {
		plainText = plainText[:100] + "..."
	}

	// If still empty after cleaning, use default
	if plainText == "" {
		return "Untitled Article"
	}

	return plainText
}

// extractMediaThumbnail extracts the thumbnail URL from Media RSS extensions (used by YouTube)
func extractMediaThumbnail(item *gofeed.Item) string {
	if item.Extensions == nil {
		return ""
	}

	// Check for media:group extension (YouTube uses this structure)
	if mediaExt, ok := item.Extensions["media"]; ok {
		if groupExts, ok := mediaExt["group"]; ok && len(groupExts) > 0 {
			// Navigate to media:group's children
			if groupExts[0].Children != nil {
				if thumbnailExts, ok := groupExts[0].Children["thumbnail"]; ok && len(thumbnailExts) > 0 {
					// Get the URL from the thumbnail's attributes
					if thumbnailExts[0].Attrs != nil {
						if url, ok := thumbnailExts[0].Attrs["url"]; ok {
							return url
						}
					}
				}
			}
		}

		// Also check for direct media:thumbnail (some feeds use this)
		if thumbnailExts, ok := mediaExt["thumbnail"]; ok && len(thumbnailExts) > 0 {
			if thumbnailExts[0].Attrs != nil {
				if url, ok := thumbnailExts[0].Attrs["url"]; ok {
					return url
				}
			}
		}
	}

	return ""
}

// extractMediaTitle extracts the title from Media RSS extensions (used by YouTube)
func extractMediaTitle(item *gofeed.Item) string {
	if item.Extensions == nil {
		return ""
	}

	// Check for media:group extension (YouTube uses this structure)
	if mediaExt, ok := item.Extensions["media"]; ok {
		if groupExts, ok := mediaExt["group"]; ok && len(groupExts) > 0 {
			// Navigate to media:group's children
			if groupExts[0].Children != nil {
				if titleExts, ok := groupExts[0].Children["title"]; ok && len(titleExts) > 0 {
					return titleExts[0].Value
				}
			}
		}

		// Also check for direct media:title (some feeds use this)
		if titleExts, ok := mediaExt["title"]; ok && len(titleExts) > 0 {
			return titleExts[0].Value
		}
	}

	return ""
}

// extractMediaDescription extracts the description from Media RSS extensions (used by YouTube)
func extractMediaDescription(item *gofeed.Item) string {
	if item.Extensions == nil {
		return ""
	}

	// Check for media:group extension (YouTube uses this structure)
	if mediaExt, ok := item.Extensions["media"]; ok {
		if groupExts, ok := mediaExt["group"]; ok && len(groupExts) > 0 {
			// Navigate to media:group's children
			if groupExts[0].Children != nil {
				if descExts, ok := groupExts[0].Children["description"]; ok && len(descExts) > 0 {
					return descExts[0].Value
				}
			}
		}

		// Also check for direct media:description (some feeds use this)
		if descExts, ok := mediaExt["description"]; ok && len(descExts) > 0 {
			return descExts[0].Value
		}
	}

	return ""
}

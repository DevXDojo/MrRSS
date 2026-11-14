package media

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"MrRSS/internal/cache"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
	"MrRSS/internal/utils/fileutil"
	"MrRSS/internal/utils/httputil"
	"MrRSS/internal/utils/textutil"

	htmlparser "golang.org/x/net/html"
)

// validateMediaURL validates that the URL is HTTP/HTTPS and properly formatted
func validateMediaURL(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return errors.New("invalid URL format")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("URL must use HTTP or HTTPS")
	}

	return nil
}

// ProxyImagesInHTML replaces image URLs in HTML with proxied versions
func ProxyImagesInHTML(htmlContent, referer string) string {
	if htmlContent == "" || referer == "" {
		return htmlContent
	}

	// Parse the referer URL once for resolving relative URLs
	baseURL, err := url.Parse(referer)
	if err != nil {
		log.Printf("Failed to parse referer URL: %v", err)
		return htmlContent
	}

	// Use regex to find and replace img src attributes
	// This handles various formats: src="url", src='url', src=url (unquoted)
	re := regexp.MustCompile(`<img[^>]*src\s*=\s*(?:['"]\s*)?([^'"\s>]+)(?:\s*['"])?[^>]*>`)
	htmlContent = re.ReplaceAllStringFunc(htmlContent, func(match string) string {
		// Extract the src URL from the match
		re := regexp.MustCompile(`src\s*=\s*(?:['"]\s*)?([^'"\s>]+)(?:\s*['"])?`)
		srcMatch := re.FindStringSubmatch(match)
		if len(srcMatch) < 2 {
			return match // No valid src found, return unchanged
		}

		srcURL := srcMatch[1]

		// Skip data URLs, blob URLs, and already proxied URLs
		if strings.HasPrefix(srcURL, "data:") ||
			strings.HasPrefix(srcURL, "blob:") ||
			strings.Contains(srcURL, "/api/media/proxy") {
			return match
		}

		// CRITICAL FIX: Decode HTML entities before processing the URL
		// HTML attributes contain &amp; which should be decoded to & before URL encoding
		// For example: ?key=val&amp;other=val becomes ?key=val&other=val
		srcURL = html.UnescapeString(srcURL)

		// Resolve relative URLs against the referer
		// Handles: images/photo.jpg, ./img.png, ../assets/image.gif, /static/img.png
		if !strings.HasPrefix(srcURL, "http://") && !strings.HasPrefix(srcURL, "https://") {
			parsedURL, err := url.Parse(srcURL)
			if err != nil {
				log.Printf("Failed to parse image URL %s: %v", srcURL, err)
				return match
			}
			srcURL = baseURL.ResolveReference(parsedURL).String()
		}

		// CRITICAL FIX: Use base64 encoding to avoid all URL encoding issues
		// This prevents double-encoding problems with special characters
		// Base64 encoding is safe for URLs and doesn't interfere with query parameter parsing
		proxyURL := fmt.Sprintf("/api/media/proxy?url_b64=%s",
			base64.StdEncoding.EncodeToString([]byte(srcURL)))

		// CRITICAL FIX: Determine if we should use the referer or not
		// Some sites block requests from certain referers (e.g., RSS hubs)
		// We use smart referer logic to handle this
		proxyReferer := getSmartReferer(srcURL, referer)

		// Add referer if provided (also base64-encoded)
		if proxyReferer != "" {
			proxyURL += fmt.Sprintf("&referer_b64=%s",
				base64.StdEncoding.EncodeToString([]byte(proxyReferer)))
		}

		// Replace the src attribute
		return strings.Replace(match, srcMatch[0], fmt.Sprintf(`src="%s"`, proxyURL), 1)
	})

	return htmlContent
}

// getSmartReferer determines the appropriate referer to use for a given image URL
// For third-party images (different domain than the referer), we use the image's own domain
// as the referer to avoid anti-hotlinking issues
func getSmartReferer(imageURL, originalReferer string) string {
	// Parse the image URL to get its hostname
	imgURL, err := url.Parse(imageURL)
	if err != nil {
		// If we can't parse the image URL, use the original referer
		return originalReferer
	}

	// Parse the original referer to get its hostname
	refURL, err := url.Parse(originalReferer)
	if err != nil {
		// If we can't parse the referer, use no referer
		return ""
	}

	imgHost := imgURL.Hostname()
	refHost := refURL.Hostname()

	// If the image host and referer host are the same domain, use the original referer
	// This handles same-origin images (e.g., images hosted on the same site as the article)
	if imgHost == refHost || strings.HasSuffix(imgHost, "."+refHost) || strings.HasSuffix(refHost, "."+imgHost) {
		return originalReferer
	}

	// For third-party images (different domain), use the image's own domain as referer
	// This avoids anti-hotlinking issues when the article's referer is blocked
	// For example: img.500px.me/image.jpg with referer from rsshub.pseudoyu.com
	// will use https://img.500px.me as the referer
	return fmt.Sprintf("%s://%s", imgURL.Scheme, imgURL.Host)
}

// HandleMediaProxy serves cached media or downloads and caches it
// HandleMediaProxy proxies media files with optional caching
// @Summary      Proxy media file
// @Description  Proxy and cache media files (images, videos, audio) from external URLs
// @Tags         media
// @Accept       json
// @Produce      application/octet-stream
// @Param        url         query     string  true  "Media URL to proxy"
// @Param        referer     query     string  false  "Referer URL for hotlink protection"
// @Param        force_cache query     bool    false  "Force caching even if globally disabled"
// @Success      200  {file}  file  "Media file"
// @Failure      400  {object}  map[string]string  "Bad request (missing or invalid URL)"
// @Failure      403  {object}  map[string]string  "Media proxy is disabled"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /media/proxy [get]
func HandleMediaProxy(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	// A media URL can also be opened as a top-level document.
	w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	r = r.WithContext(ctx)
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get URL from query parameter (support both direct and base64-encoded)
	mediaURL := r.URL.Query().Get("url")
	mediaURLBase64 := r.URL.Query().Get("url_b64")

	// Use base64-encoded URL if provided, otherwise use direct URL
	if mediaURLBase64 != "" {
		// Decode base64 URL
		decodedBytes, err := base64.StdEncoding.DecodeString(mediaURLBase64)
		if err != nil {
			log.Printf("Failed to decode base64 URL: %v", err)
			response.Error(w, err, http.StatusBadRequest)
			return
		}
		mediaURL = string(decodedBytes)
	}

	if mediaURL == "" {
		response.Error(w, fmt.Errorf("missing url parameter"), http.StatusBadRequest)
		return
	}

	// Validate mediaURL (must be HTTP/HTTPS and valid format)
	if err := validateMediaURL(mediaURL); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Check if media cache is enabled
	mediaCacheEnabled, _ := h.DB.GetSetting("media_cache_enabled")
	mediaProxyFallback, _ := h.DB.GetSetting("media_proxy_fallback")

	// Check if force_cache parameter is set (for image mode feeds)
	forceCache := r.URL.Query().Get("force_cache") == "true"

	// If force_cache is true, enable caching for this request
	if forceCache {
		mediaCacheEnabled = "true"
	}

	// If neither cache nor fallback is enabled (and not forced), return error
	if mediaCacheEnabled != "true" && mediaProxyFallback != "true" {
		response.Error(w, fmt.Errorf("media proxy is disabled"), http.StatusForbidden)
		return
	}

	// Get optional referer from query parameter (support both direct and base64-encoded)
	referer := r.URL.Query().Get("referer")
	refererBase64 := r.URL.Query().Get("referer_b64")
	if refererBase64 != "" {
		// Decode base64 referer
		decodedBytes, err := base64.StdEncoding.DecodeString(refererBase64)
		if err != nil {
			log.Printf("Failed to decode base64 referer: %v", err)
			// Fall back to unencoded referer
		} else {
			referer = string(decodedBytes)
		}
	}

	client, err := httputil.CreateHTTPClientWithProxySettings(h.DB, 30*time.Second)
	if err != nil {
		response.Error(w, fmt.Errorf("failed to configure media HTTP client"), http.StatusInternalServerError)
		return
	}
	defer client.CloseIdleConnections()

	// Try cache first if enabled
	if mediaCacheEnabled == "true" {
		// Get media cache directory
		cacheDir, err := fileutil.GetMediaCacheDir()
		if err != nil {
			log.Printf("Failed to get media cache directory: %v", err)
			// Continue to fallback if enabled
		} else {
			// Initialize media cache
			mediaCache, err := cache.NewMediaCache(cacheDir)
			if err != nil {
				log.Printf("Failed to initialize media cache: %v", err)
				// Continue to fallback if enabled
			} else {
				// Serve from disk when the media is already cached: streaming keeps
				// large images out of the process heap.
				if file, contentType, modTime, openErr := mediaCache.Open(mediaURL); openErr == nil {
					serveCachedMedia(w, r, file, contentType, modTime, filepath.Base(mediaURL))
					return
				}

				// Cache miss: download straight into the cache file, then stream it.
				path, contentType, err := mediaCache.DownloadToFile(r.Context(), client, mediaURL, referer)
				if err == nil {
					if file, openErr := os.Open(path); openErr == nil {
						info, statErr := file.Stat()
						if statErr == nil {
							serveCachedMedia(w, r, file, contentType, info.ModTime(), filepath.Base(path))
							return
						}
						_ = file.Close()
					}
				}
				var downloadErr *cache.MediaDownloadError
				if errors.As(err, &downloadErr) && downloadErr.RefererFallbackAttempted {
					response.Error(w, fmt.Errorf("failed to fetch media"), http.StatusInternalServerError)
					return
				}
				log.Printf("Cache failed for %s: %v, trying fallback", mediaURL, err)
			}
		}
	}

	// Fallback: Direct proxy if enabled
	if mediaProxyFallback == "true" {
		err := proxyMediaDirectly(r.Context(), client, mediaURL, referer, w)
		if err == nil {
			return // Success
		}
		log.Printf("Direct proxy failed for %s: %v", mediaURL, err)
	}

	// All methods failed
	response.Error(w, fmt.Errorf("failed to fetch media"), http.StatusInternalServerError)
}

// serveCachedMedia streams a cached media file to the client. http.ServeContent
// sets Content-Length and handles range and conditional requests, so the image
// is never buffered in memory. It closes the file.
func serveCachedMedia(w http.ResponseWriter, r *http.Request, file *os.File, contentType string, modTime time.Time, name string) {
	defer file.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	w.Header().Set("X-Media-Source", "cache")
	http.ServeContent(w, r, name, modTime, file)
}

// HandleMediaCacheCleanup performs manual cleanup of media cache
// @Summary      Cleanup media cache
// @Description  Clean up the media cache by age and size
// @Tags         media
// @Accept       json
// @Produce      json
// @Param        all  query     bool  false  "Clean all files (ignores age/size settings)"  default(false)
// @Success      200  {object}  map[string]interface{}  "Cleanup result (success, files_cleaned)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /media/cache/cleanup [post]
func HandleMediaCacheCleanup(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get media cache directory
	cacheDir, err := fileutil.GetMediaCacheDir()
	if err != nil {
		log.Printf("Failed to get media cache directory: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Initialize media cache
	mediaCache, err := cache.NewMediaCache(cacheDir)
	if err != nil {
		log.Printf("Failed to initialize media cache: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Check if this is a manual cleanup (clean all) or automatic cleanup (respect settings)
	cleanAll := r.URL.Query().Get("all") == "true"

	var maxAgeDays int
	var maxSizeMB int

	if cleanAll {
		// Manual cleanup: remove all files
		maxAgeDays = 0
		maxSizeMB = 0 // Will skip size-based cleanup
	} else {
		// Automatic cleanup: use settings
		maxAgeDaysStr, _ := h.DB.GetSetting("media_cache_max_age_days")
		maxSizeMBStr, _ := h.DB.GetSetting("media_cache_max_size_mb")

		maxAgeDays, err = strconv.Atoi(maxAgeDaysStr)
		if err != nil || maxAgeDays < 0 {
			maxAgeDays = 7 // Default
		}

		maxSizeMB, err = strconv.Atoi(maxSizeMBStr)
		if err != nil || maxSizeMB <= 0 {
			maxSizeMB = 100 // Default
		}
	}

	// Cleanup by age
	ageCount, err := mediaCache.CleanupOldFiles(maxAgeDays)
	if err != nil {
		log.Printf("Failed to cleanup old media files: %v", err)
	}

	// Cleanup by size (only for automatic cleanup)
	sizeCount := 0
	if !cleanAll {
		sizeCount, err = mediaCache.CleanupBySize(maxSizeMB)
		if err != nil {
			log.Printf("Failed to cleanup media files by size: %v", err)
		}
	}

	totalCleaned := ageCount + sizeCount
	log.Printf("Media cache cleanup: removed %d files (clean_all: %v)", totalCleaned, cleanAll)

	response.JSON(w, map[string]interface{}{
		"success":       true,
		"files_cleaned": totalCleaned,
	})
}

// HandleWebpageProxy proxies webpage content to bypass CORS restrictions in iframes
// @Summary      Proxy webpage content
// @Description  Proxy webpage HTML content and rewrite resource URLs to bypass CORS restrictions
// @Tags         media
// @Accept       json
// @Produce      html
// @Param        url  query     string  true  "Webpage URL to proxy"
// @Success      200  {string}  string  "Webpage HTML content"
// @Failure      400  {object}  map[string]string  "Bad request (missing or invalid URL)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /media/proxy-webpage [get]
func HandleWebpageProxy(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get URL from query parameter
	webpageURL := r.URL.Query().Get("url")
	if webpageURL == "" {
		response.Error(w, fmt.Errorf("missing url parameter"), http.StatusBadRequest)
		return
	}

	// Validate webpageURL (must be HTTP/HTTPS and valid format)
	if err := validateMediaURL(webpageURL); err != nil {
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Create HTTP client with proxy settings if enabled
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Check if proxy is enabled and configure client
	proxyEnabled, _ := h.DB.GetSetting("proxy_enabled")
	if proxyEnabled == "true" {
		proxyType, _ := h.DB.GetSetting("proxy_type")
		proxyHost, _ := h.DB.GetSetting("proxy_host")
		proxyPort, _ := h.DB.GetSetting("proxy_port")
		proxyUsername, _ := h.DB.GetSetting("proxy_username")
		proxyPassword, _ := h.DB.GetSetting("proxy_password")

		proxyURLStr := httputil.BuildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)
		if proxyURLStr != "" {
			proxyURL, err := url.Parse(proxyURLStr)
			if err != nil {
				log.Printf("Failed to parse proxy URL: %v", err)
			} else {
				transport := &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				}
				client.Transport = transport
			}
		}
	}

	// Create request to the target URL
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, webpageURL, nil)
	if err != nil {
		log.Printf("Failed to create request: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Set User-Agent to mimic a regular browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// Forward some headers from the original request
	if referer := r.Header.Get("Referer"); referer != "" {
		req.Header.Set("Referer", referer)
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch webpage %s: %v", webpageURL, err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("Webpage returned status %d: %s", resp.StatusCode, webpageURL)
		response.Error(w, fmt.Errorf("webpage returned error"), resp.StatusCode)
		return
	}

	// Get content type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/html; charset=utf-8"
	}

	// Read the entire response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// If this is HTML content, rewrite all resource URLs
	if strings.Contains(strings.ToLower(contentType), "text/html") {
		bodyBytes = rewriteHTMLContent(bodyBytes, resp.Request.URL.String())
	}

	setWebpagePreviewHeaders(w)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(bodyBytes)))

	// Write modified response body
	_, err = w.Write(bodyBytes)
	if err != nil {
		log.Printf("Failed to write response body: %v", err)
	}
}

func setWebpagePreviewHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Security-Policy", "sandbox allow-popups allow-popups-to-escape-sandbox; default-src 'none'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; media-src 'self'; font-src 'self' data:; base-uri 'none'; form-action 'none'; frame-ancestors 'self'")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Cache-Control", "no-store")
	// Public styles/fonts can be read by the sandboxed preview's opaque origin.
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

// rewriteHTMLContent rewrites HTML to proxy all external resources
func rewriteHTMLContent(bodyBytes []byte, baseURL string) []byte {
	content := rewriteLinkHref(string(bodyBytes), baseURL)

	// First, convert lazy-loaded images to normal images
	// This ensures images load immediately without waiting for lazy loading scripts
	content = convertLazyImages(content, baseURL)

	// Then rewrite img src attributes (now including the converted lazy images)
	content = rewriteAttribute(content, "img", "src", baseURL)
	content = rewriteSrcsetAttribute(content, "img", "srcset", baseURL)
	content = rewriteSrcsetAttribute(content, "img", "data-srcset", baseURL)

	// Rewrite video src and poster attributes
	content = rewriteAttribute(content, "video", "src", baseURL)
	content = rewriteAttribute(content, "video", "poster", baseURL)

	// Rewrite audio src attributes
	content = rewriteAttribute(content, "audio", "src", baseURL)

	// Rewrite source src attributes (for video/audio)
	content = rewriteAttribute(content, "source", "src", baseURL)
	content = rewriteSrcsetAttribute(content, "source", "srcset", baseURL)
	content = rewriteSrcsetAttribute(content, "source", "data-srcset", baseURL)

	// Rewrite track src attributes
	content = rewriteAttribute(content, "track", "src", baseURL)

	// Rewrite CSS in style tags
	content = rewriteStyleTags(content, baseURL)

	// Rewrite inline style attributes
	content = rewriteInlineStyles(content, baseURL)

	return prepareWebpagePreview(content, baseURL)
}

// prepareWebpagePreview removes active elements and makes outbound links open
// outside the preview. A response policy also enforces a scriptless sandbox.
func prepareWebpagePreview(content, baseURL string) []byte {
	doc, err := htmlparser.Parse(strings.NewReader(content))
	if err != nil {
		return nil
	}
	base, _ := url.Parse(baseURL)
	var visit func(*htmlparser.Node)
	visit = func(node *htmlparser.Node) {
		for child := node.FirstChild; child != nil; {
			next := child.NextSibling
			if child.Type == htmlparser.ElementNode {
				switch child.Data {
				case "script", "meta", "base", "iframe", "frame", "object", "embed", "template":
					node.RemoveChild(child)
					child = next
					continue
				}
				attrs := make([]htmlparser.Attribute, 0, len(child.Attr))
				for _, attr := range child.Attr {
					if strings.HasPrefix(attr.Key, "on") || attr.Key == "ping" || attr.Key == "srcdoc" {
						continue
					}
					if child.Data == "a" || child.Data == "area" {
						switch attr.Key {
						case "target", "rel", "download", "referrerpolicy":
							continue
						case "href":
							if attr.Namespace != "" {
								continue
							}
							if !strings.HasPrefix(attr.Val, "#") {
								link, err := url.Parse(strings.TrimSpace(attr.Val))
								if err != nil || base == nil {
									continue
								}
								link = base.ResolveReference(link)
								if link.Hostname() == "" || (link.Scheme != "http" && link.Scheme != "https") {
									continue
								}
								attr.Val = link.String()
								attrs = append(attrs, attr, htmlparser.Attribute{Key: "target", Val: "_blank"},
									htmlparser.Attribute{Key: "rel", Val: "noopener noreferrer"})
								continue
							}
						}
					}
					attrs = append(attrs, attr)
				}
				child.Attr = attrs
			}
			visit(child)
			child = next
		}
	}
	visit(doc)
	var preview bytes.Buffer
	if err := htmlparser.Render(&preview, doc); err != nil {
		return nil
	}
	return preview.Bytes()
}

// convertLazyImages resolves lazy image sources before rewriting resource URLs.
// Keep Discuz attachment attributes proxied too, because its scripts may reuse them.
func convertLazyImages(content, baseURL string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return content
	}
	images := regexp.MustCompile(`(?is)<img\b[^>]*>`)
	return images.ReplaceAllStringFunc(content, func(match string) string {
		tokenizer := htmlparser.NewTokenizer(strings.NewReader(match))
		kind := tokenizer.Next()
		if kind != htmlparser.StartTagToken && kind != htmlparser.SelfClosingTagToken {
			return match
		}
		token := tokenizer.Token()
		attributes := make(map[string]string, len(token.Attr))
		lazy := false
		for _, attr := range token.Attr {
			attributes[attr.Key] = attr.Val
			switch attr.Key {
			case "data-src", "data-original", "data-lazy-src", "data-actualsrc", "data-original-src", "zoomfile", "file":
				lazy = true
			}
		}
		if !lazy {
			return match
		}
		source := textutil.ResolveArticleImageSource(attributes, base)
		if source == "" {
			return match
		}
		// A lazy attribute can already refer to our local proxy. Keep it local.
		for _, value := range attributes {
			if isLocalImageProxy(value) {
				resolved, err := base.Parse(value)
				if err == nil && resolved.String() == source {
					source = value
					break
				}
			}
		}
		attrs := make([]htmlparser.Attribute, 0, len(token.Attr)+1)
		for _, attr := range token.Attr {
			switch attr.Key {
			case "src", "data-src", "data-original", "data-lazy-src", "data-actualsrc", "data-original-src":
				continue
			case "zoomfile", "file":
				if isLocalImageProxy(attr.Val) {
					attrs = append(attrs, attr)
					continue
				}
				// Tokenizer has already decoded entities; the resource helper expects HTML.
				proxied, ok := proxyWebpageResourceURL(html.EscapeString(attr.Val), baseURL)
				if !ok {
					continue
				}
				attr.Val = proxied
			case "class":
				classes := strings.Fields(attr.Val)
				kept := classes[:0]
				for _, class := range classes {
					if class != "lazy" {
						kept = append(kept, class)
					}
				}
				attr.Val = strings.Join(kept, " ")
			}
			attrs = append(attrs, attr)
		}
		token.Attr = append(attrs, htmlparser.Attribute{Key: "src", Val: source})
		return token.String()
	})
}

func isLocalImageProxy(value string) bool {
	return strings.HasPrefix(value, "/api/webpage/resource?") || strings.HasPrefix(value, "/api/media/proxy?")
}

// htmlAttribute represents a parsed HTML attribute
type htmlAttribute struct {
	Name  string
	Value string
	Quote string // " or ' or empty for unquoted
}

// parseHTMLAttributes parses attributes from an HTML tag
// This is a simple parser that handles quoted and unquoted attributes
func parseHTMLAttributes(tag string) []htmlAttribute {
	// Remove "<img" and ">" from the tag
	content := strings.TrimPrefix(tag, "<img")
	content = strings.TrimSuffix(content, ">")
	content = strings.TrimSpace(content)

	var attrs []htmlAttribute
	var currentAttr strings.Builder
	var inQuote rune
	var attrName, attrValue, attrQuote strings.Builder

	i := 0
	for i < len(content) {
		ch := rune(content[i])

		if inQuote != 0 {
			// We're inside a quoted value
			if ch == inQuote {
				// Closing quote
				inQuote = 0
				attrValue.WriteString(string(ch))
			} else {
				attrValue.WriteString(string(ch))
			}
			i++
			continue
		}

		switch ch {
		case '=':
			// End of attribute name
			attrName.WriteString(currentAttr.String())
			currentAttr.Reset()
			i++
			// Skip whitespace after =
			for i < len(content) && content[i] == ' ' {
				i++
			}
			// Check for quote
			if i < len(content) && (content[i] == '"' || content[i] == '\'') {
				inQuote = rune(content[i])
				attrQuote.WriteRune(inQuote)
				attrValue.WriteRune(inQuote)
				i++
			}
		case ' ', '\t', '\n', '\r':
			if currentAttr.Len() > 0 {
				// End of attribute value (unquoted)
				if attrName.Len() > 0 {
					attrs = append(attrs, htmlAttribute{
						Name:  strings.TrimSpace(attrName.String()),
						Value: currentAttr.String(),
						Quote: "",
					})
					attrName.Reset()
				}
				currentAttr.Reset()
				attrQuote.Reset()
			}
			i++
		default:
			if attrName.Len() == 0 {
				currentAttr.WriteRune(ch)
			} else {
				attrValue.WriteRune(ch)
			}
			i++
		}
	}

	// Don't forget the last attribute
	if attrName.Len() > 0 {
		attrs = append(attrs, htmlAttribute{
			Name:  strings.TrimSpace(attrName.String()),
			Value: strings.TrimSpace(attrValue.String()),
			Quote: attrQuote.String(),
		})
	} else if currentAttr.Len() > 0 {
		// Boolean attribute or attribute without value
		attrs = append(attrs, htmlAttribute{
			Name:  currentAttr.String(),
			Value: "",
			Quote: "",
		})
	}

	return attrs
}

// rewriteAttribute rewrites a specific URL attribute in HTML tags.
func rewriteAttribute(content, tag, attr, baseURL string) string {
	return rewriteAttributeValue(content, tag, attr, func(value string) (string, bool) {
		return proxyWebpageResourceURL(value, baseURL)
	})
}

// rewriteSrcsetAttribute proxies every candidate URL while preserving its
// density or width descriptor (for example, 2x or 640w).
func rewriteSrcsetAttribute(content, tag, attr, baseURL string) string {
	return rewriteAttributeValue(content, tag, attr, func(value string) (string, bool) {
		var result strings.Builder
		changed := false
		for position := 0; position < len(value); {
			prefixStart := position
			for position < len(value) && (isHTMLSpace(value[position]) || value[position] == ',') {
				position++
			}
			result.WriteString(value[prefixStart:position])
			if position >= len(value) {
				break
			}

			urlStart := position
			isDataURL := strings.HasPrefix(strings.ToLower(value[position:]), "data:")
			for position < len(value) && !isHTMLSpace(value[position]) && (isDataURL || value[position] != ',') {
				position++
			}
			candidate := value[urlStart:position]
			if proxied, ok := proxyWebpageResourceURL(candidate, baseURL); ok {
				result.WriteString(proxied)
				changed = true
			} else {
				result.WriteString(candidate)
			}

			descriptorStart := position
			for position < len(value) && value[position] != ',' {
				position++
			}
			result.WriteString(value[descriptorStart:position])
		}
		return result.String(), changed
	})
}

func rewriteAttributeValue(content, tag, attr string, rewrite func(string) (string, bool)) string {
	tagRe := regexp.MustCompile(`(?i)<` + regexp.QuoteMeta(tag) + `\b[^>]*>`)
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\s+` + regexp.QuoteMeta(attr) + `\s*=\s*"([^"]*)"`),
		regexp.MustCompile(`(?i)\s+` + regexp.QuoteMeta(attr) + `\s*=\s*'([^']*)'`),
		regexp.MustCompile(`(?i)\s+` + regexp.QuoteMeta(attr) + `\s*=\s*([^\s>]+)`),
	}

	return tagRe.ReplaceAllStringFunc(content, func(match string) string {
		for index, pattern := range patterns {
			location := pattern.FindStringSubmatchIndex(match)
			if len(location) < 4 {
				continue
			}
			valueStart, valueEnd := location[2], location[3]
			rewritten, changed := rewrite(match[valueStart:valueEnd])
			if !changed {
				return match
			}
			if index == len(patterns)-1 {
				rewritten = `"` + rewritten + `"`
			}
			return match[:valueStart] + rewritten + match[valueEnd:]
		}
		return match
	})
}

func proxyWebpageResourceURL(value, baseURL string) (string, bool) {
	value = strings.TrimSpace(html.UnescapeString(value))
	lowerValue := strings.ToLower(value)
	if value == "" || strings.HasPrefix(lowerValue, "data:") ||
		strings.HasPrefix(lowerValue, "blob:") || strings.HasPrefix(value, "#") ||
		strings.HasPrefix(value, "/api/") || strings.Contains(value, "/api/webpage/resource?") {
		return value, false
	}

	resolvedURL := resolveURL(value, baseURL)
	parsedURL, err := url.Parse(resolvedURL)
	if err != nil || parsedURL.Hostname() == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return value, false
	}
	query := url.Values{
		"url_b64":     {base64.StdEncoding.EncodeToString([]byte(resolvedURL))},
		"referer_b64": {base64.StdEncoding.EncodeToString([]byte(baseURL))},
	}
	return "/api/webpage/resource?" + query.Encode(), true
}

func isHTMLSpace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r' || value == '\f'
}

// rewriteLinkHref rewrites href attributes in link tags
func rewriteLinkHref(content, baseURL string) string {
	return rewriteAttribute(content, "link", "href", baseURL)
}

// rewriteStyleTags rewrites CSS URLs in <style> tags
func rewriteStyleTags(content, baseURL string) string {
	pattern := `<style[^>]*>(.*?)</style>`
	re := regexp.MustCompile(`(?is)` + pattern)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the CSS content
		subRe := regexp.MustCompile(`(?is)<style[^>]*>(.*?)</style>`)
		matches := subRe.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}

		cssContent := matches[1]
		// Rewrite @font-face rules first
		cssContent = rewriteFontFaceRules(cssContent, baseURL)
		// Then rewrite all other url() references
		rewrittenCSS := rewriteCSSURLs(cssContent, baseURL)

		return strings.Replace(match, cssContent, rewrittenCSS, 1)
	})
}

// rewriteInlineStyles rewrites style attribute values
func rewriteInlineStyles(content, baseURL string) string {
	re := regexp.MustCompile(`(?is)\bstyle\s*=\s*(?:"([^"]*)"|'([^']*)')`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		style := parts[1]
		if style == "" {
			style = parts[2]
		}
		rewritten := rewriteCSSURLs(html.UnescapeString(style), baseURL)
		return `style="` + html.EscapeString(rewritten) + `"`
	})
}

// rewriteCSSURLs resolves imports, fonts and images against the stylesheet URL.
// Go's regexp does not support backreferences; use separate quoted alternatives.
func rewriteCSSURLs(css, baseURL string) string {
	imports := regexp.MustCompile(`(?i)@import\s+(?:"([^"]+)"|'([^']+)')`)
	css = imports.ReplaceAllStringFunc(css, func(match string) string {
		parts := imports.FindStringSubmatch(match)
		value := parts[1]
		if value == "" {
			value = parts[2]
		}
		return `@import url("` + value + `")`
	})
	re := regexp.MustCompile(`(?i)url\(\s*(?:"([^"]*)"|'([^']*)'|([^'"\)]*))\s*\)`)
	return re.ReplaceAllStringFunc(css, func(match string) string {
		parts := re.FindStringSubmatch(match)
		value := parts[1]
		if value == "" {
			value = parts[2]
		}
		if value == "" {
			value = parts[3]
		}
		if proxied, ok := proxyWebpageResourceURL(html.EscapeString(value), baseURL); ok {
			return `url("` + proxied + `")`
		}
		return match
	})
}

// rewriteFontFaceRules rewrites @font-face rules in CSS
func rewriteFontFaceRules(css, baseURL string) string {
	// Match @font-face blocks
	pattern := `@font-face\s*\{[^}]*\}`
	re := regexp.MustCompile(`(?is)` + pattern)

	return re.ReplaceAllStringFunc(css, func(match string) string {
		// Rewrite url() within this @font-face block
		return rewriteCSSURLs(match, baseURL)
	})
}

// resolveURL resolves a URL relative to a base URL
func resolveURL(urlStr, baseURL string) string {
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		return urlStr
	}

	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		return urlStr
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	return parsedBase.ResolveReference(parsedURL).String()
}

// HandleWebpageResource proxies individual webpage resources (CSS, JS, images, etc.)
// @Summary      Proxy webpage resource
// @Description  Proxy individual resources (CSS, JS, images, fonts, etc.) from a webpage
// @Tags         media
// @Accept       json
// @Produce      application/octet-stream
// @Param        url      query     string  true  "Resource URL to proxy"
// @Param        referer  query     string  true  "Referer URL for the webpage"
// @Success      200  {file}  file  "Resource file"
// @Failure      400  {object}  map[string]string  "Bad request (missing or invalid URL)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /webpage/resource [get]
func HandleWebpageResource(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, nil, http.StatusMethodNotAllowed)
		return
	}

	// Get URL from query parameter (support both direct and base64-encoded)
	resourceURL := r.URL.Query().Get("url")
	resourceURLBase64 := r.URL.Query().Get("url_b64")

	// Use base64-encoded URL if provided, otherwise use direct URL
	if resourceURLBase64 != "" {
		// Decode base64 URL
		decodedBytes, err := base64.StdEncoding.DecodeString(resourceURLBase64)
		if err != nil {
			log.Printf("Failed to decode base64 URL: %v", err)
			response.Error(w, err, http.StatusBadRequest)
			return
		}
		resourceURL = string(decodedBytes)
	}

	if resourceURL == "" {
		response.Error(w, fmt.Errorf("missing url parameter"), http.StatusBadRequest)
		return
	}

	// Get referer from query parameter (support both direct and base64-encoded)
	referer := r.URL.Query().Get("referer")
	refererBase64 := r.URL.Query().Get("referer_b64")

	// Use base64-encoded referer if provided, otherwise use direct referer
	if refererBase64 != "" {
		// Decode base64 referer
		decodedBytes, err := base64.StdEncoding.DecodeString(refererBase64)
		if err != nil {
			log.Printf("Failed to decode base64 referer: %v", err)
			// Fall back to unencoded referer if available
		} else {
			referer = string(decodedBytes)
		}
	}

	if referer == "" {
		response.Error(w, fmt.Errorf("missing referer parameter"), http.StatusBadRequest)
		return
	}

	// Validate URLs
	if err := validateMediaURL(resourceURL); err != nil {
		log.Printf("Invalid URL validation failed for %s: %v", resourceURL, err)
		response.Error(w, err, http.StatusBadRequest)
		return
	}
	if err := validateMediaURL(referer); err != nil {
		log.Printf("Invalid referer validation failed for %s: %v", referer, err)
		response.Error(w, err, http.StatusBadRequest)
		return
	}

	// Create HTTP client with proxy settings if enabled
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Check if proxy is enabled and configure client
	proxyEnabled, _ := h.DB.GetSetting("proxy_enabled")
	if proxyEnabled == "true" {
		proxyType, _ := h.DB.GetSetting("proxy_type")
		proxyHost, _ := h.DB.GetSetting("proxy_host")
		proxyPort, _ := h.DB.GetSetting("proxy_port")
		proxyUsername, _ := h.DB.GetSetting("proxy_username")
		proxyPassword, _ := h.DB.GetSetting("proxy_password")

		proxyURLStr := httputil.BuildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)
		if proxyURLStr != "" {
			proxyURL, err := url.Parse(proxyURLStr)
			if err != nil {
				log.Printf("Failed to parse proxy URL: %v", err)
			} else {
				transport := &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				}
				client.Transport = transport
			}
		}
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, resourceURL, nil)
	if err != nil {
		response.Error(w, fmt.Errorf("invalid resource URL"), http.StatusBadRequest)
		return
	}

	// Set headers to mimic a browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", referer)
	req.Header.Set("Accept", "*/*")
	if value := r.Header.Get("Range"); value != "" {
		req.Header.Set("Range", value)
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to fetch resource %s: %v", resourceURL, err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check response status - allow 200, 201, 202, 203, 204, 206
	if resp.StatusCode < 200 || resp.StatusCode > 206 {
		log.Printf("Resource returned status %d for %s (method: %s)", resp.StatusCode, resourceURL, r.Method)
		response.Error(w, fmt.Errorf("resource returned error"), resp.StatusCode)
		return
	}

	// Get content type from response
	contentType := resp.Header.Get("Content-Type")

	// Also infer content type from file extension as fallback
	inferredContentType := getContentTypeFromPath(resourceURL)

	// CRITICAL FIX: Always use inferred content type for known file types
	// Many servers return incorrect Content-Type headers (e.g., text/plain for .css or .js files)
	// We override these with the correct type based on file extension
	shouldInferType := false
	ext := webpageResourceExtension(resourceURL)
	switch ext {
	case ".css", ".js", ".mjs", ".woff", ".woff2", ".ttf", ".eot", ".otf", ".json", ".xml":
		// These types MUST have correct MIME types or browsers will reject them
		shouldInferType = true
	}

	// Use inferred content type if the response's content type is missing, generic, or for known types
	if contentType == "" || contentType == "application/octet-stream" || contentType == "text/plain" || shouldInferType {
		contentType = inferredContentType
		log.Printf("[Resource Proxy] Using inferred content type for %s: %s (original: %s)", resourceURL, contentType, resp.Header.Get("Content-Type"))
	}

	// Only forward headers needed to serve the resource. The upstream must not
	// control navigation, cookies, or policies on the application origin.
	for _, name := range []string{"Content-Range", "Accept-Ranges"} {
		if value := resp.Header.Get(name); value != "" {
			w.Header().Set(name, value)
		}
	}
	setWebpagePreviewHeaders(w)

	// Set the correct content type
	w.Header().Set("Content-Type", contentType)

	// If this is a CSS file, rewrite URLs in it
	if strings.Contains(strings.ToLower(contentType), "text/css") {
		// Read the CSS content
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read CSS content: %v", err)
			response.Error(w, err, http.StatusInternalServerError)
			return
		}

		// Rewrite @font-face rules
		cssContent := rewriteFontFaceRules(string(bodyBytes), resp.Request.URL.String())
		// Rewrite all url() references
		cssContent = rewriteCSSURLs(cssContent, resp.Request.URL.String())

		// Update content length
		bodyBytes = []byte(cssContent)
		w.Header().Set("Content-Length", strconv.Itoa(len(bodyBytes)))

		// Write the modified CSS
		_, err = w.Write(bodyBytes)
		if err != nil {
			log.Printf("Failed to write CSS content: %v", err)
		}
		return
	}

	// For non-CSS files, stream directly
	// Stream the response directly to avoid loading large files into memory
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Failed to stream resource: %v", err)
	}
}

// proxyMediaDirectly proxies media directly without caching
func proxyMediaDirectly(ctx context.Context, client *http.Client, mediaURL, referer string, w http.ResponseWriter) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mediaURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers to bypass anti-hotlinking
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// CRITICAL FIX: Use smart referer logic to handle cases where the original referer would be blocked
	smartReferer := getSmartReferer(mediaURL, referer)
	if smartReferer != "" {
		req.Header.Set("Referer", smartReferer)
	}

	// Add additional headers
	// Note: Don't set Accept-Encoding - let Go's http.Transport handle it automatically
	req.Header.Set("Accept", "image/webp,image/apng,image/*,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, _, err := httputil.DoWithRefererFallback(client, req)
	if err != nil {
		return fmt.Errorf("failed to fetch media: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = getContentTypeFromPath(mediaURL)
	}

	// Set response headers
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	w.Header().Set("X-Media-Source", "direct-proxy")

	// Stream the response directly to avoid loading large files into memory
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to stream response: %w", err)
	}

	return nil
}

// HandleMediaCacheInfo returns information about the media cache
// HandleMediaCacheInfo returns information about the media cache
// @Summary      Get media cache info
// @Description  Get media cache statistics (size in MB)
// @Tags         media
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Cache info (cache_size_mb)"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /media/cache/info [get]
func HandleMediaCacheInfo(h *core.Handler, w http.ResponseWriter, r *http.Request) {

	// Get media cache directory
	cacheDir, err := fileutil.GetMediaCacheDir()
	if err != nil {
		log.Printf("Failed to get media cache directory: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Initialize media cache
	mediaCache, err := cache.NewMediaCache(cacheDir)
	if err != nil {
		log.Printf("Failed to initialize media cache: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Get cache size
	cacheSize, err := mediaCache.GetCacheSize()
	if err != nil {
		log.Printf("Failed to get cache size: %v", err)
		response.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Convert to MB
	cacheSizeMB := float64(cacheSize) / (1024 * 1024)

	w.Header().Set("Content-Type", "application/json")
	result := map[string]interface{}{
		"cache_size_mb": cacheSizeMB,
	}
	response.JSON(w, result)
}

// getContentTypeFromPath determines content type from file extension
func webpageResourceExtension(raw string) string {
	if parsed, err := url.Parse(raw); err == nil {
		raw = parsed.Path
	}
	return strings.ToLower(filepath.Ext(raw))
}

func getContentTypeFromPath(path string) string {
	ext := webpageResourceExtension(path)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".ogg":
		return "video/ogg"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".flac":
		return "audio/flac"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".mjs":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	case ".otf":
		return "font/otf"
	case ".xml":
		return "text/xml; charset=utf-8"
	case ".html":
		return "text/html; charset=utf-8"
	case ".txt":
		return "text/plain; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

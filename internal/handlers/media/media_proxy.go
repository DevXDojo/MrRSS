package media

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
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
				// Get media (from cache or download)
				data, contentType, err := mediaCache.Get(mediaURL, referer)
				if err == nil {
					// Success! Serve from cache
					w.Header().Set("Content-Type", contentType)
					w.Header().Set("Content-Length", strconv.Itoa(len(data)))
					w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
					w.Header().Set("X-Media-Source", "cache")
					w.Write(data)
					return
				}
				log.Printf("Cache failed for %s: %v, trying fallback", mediaURL, err)
			}
		}
	}

	// Fallback: Direct proxy if enabled
	if mediaProxyFallback == "true" {
		err := proxyMediaDirectly(mediaURL, referer, w)
		if err == nil {
			return // Success
		}
		log.Printf("Direct proxy failed for %s: %v", mediaURL, err)
	}

	// All methods failed
	response.Error(w, fmt.Errorf("failed to fetch media"), http.StatusInternalServerError)
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
	req, err := http.NewRequest("GET", webpageURL, nil)
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
		bodyBytes = rewriteHTMLContent(bodyBytes, webpageURL)
	}

	// Set response headers to allow framing and remove CORS restrictions
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Frame-Options", "SAMEORIGIN") // Allow framing from same origin
	w.Header().Set("Content-Security-Policy", "")   // Remove CSP to allow resources
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Length", strconv.Itoa(len(bodyBytes)))

	// Write modified response body
	_, err = w.Write(bodyBytes)
	if err != nil {
		log.Printf("Failed to write response body: %v", err)
	}
}

// rewriteHTMLContent rewrites HTML to proxy all external resources
func rewriteHTMLContent(bodyBytes []byte, baseURL string) []byte {
	// Validate base URL
	if _, err := url.Parse(baseURL); err != nil {
		log.Printf("Failed to parse base URL: %v", err)
		return bodyBytes
	}

	// Convert to string for manipulation
	content := string(bodyBytes)

	// Inject our script FIRST - before anything else
	// This script must run before any other scripts to intercept API calls early
	interceptionScript := `<script>
	// Use immediately-invoked function with strict error suppression
	(function() {
		'use strict';
		const ORIGINAL_BASE_URL = ` + fmt.Sprintf("'%s'", baseURL) + `;
		const PROXY_ORIGIN = window.location.origin;

		// DEBUG: Log that interceptor is loaded
		console.log('[Proxy] Interceptor loaded for:', ORIGINAL_BASE_URL);

		// Override History API BEFORE anything else with try-catch to suppress ALL errors
		try {
			const originalPushState = History.prototype.pushState;
			History.prototype.pushState = function(state, title, url) {
				try {
					if (url && typeof url === 'string' && (url.indexOf('http://') === 0 || url.indexOf('https://') === 0)) {
						// Silently block - don't even log to avoid console spam
						return undefined;
					}
				} catch(e) { /* Suppress all errors */ }
				try {
					return originalPushState.call(this, state, title, url);
				} catch(e) { /* Suppress errors from original call */ }
			};
		} catch(e) { /* Suppress errors during override */ }

		try {
			const originalReplaceState = History.prototype.replaceState;
			History.prototype.replaceState = function(state, title, url) {
				try {
					if (url && typeof url === 'string' && (url.indexOf('http://') === 0 || url.indexOf('https://') === 0)) {
						// Silently block
						return undefined;
					}
				} catch(e) { /* Suppress all errors */ }
				try {
					return originalReplaceState.call(this, state, title, url);
				} catch(e) { /* Suppress errors from original call */ }
			};
		} catch(e) { /* Suppress errors during override */ }

		// Also override on window.history for direct access
		try {
			if (window.history && window.history.pushState) {
				const originalPushState = window.history.pushState;
				window.history.pushState = function(state, title, url) {
					try {
						if (url && typeof url === 'string' && (url.indexOf('http://') === 0 || url.indexOf('https://') === 0)) {
							return undefined;
						}
					} catch(e) { }
					try {
						return originalPushState.call(this, state, title, url);
					} catch(e) { }
				};
			}
		} catch(e) { }

		try {
			if (window.history && window.history.replaceState) {
				const originalReplaceState = window.history.replaceState;
				window.history.replaceState = function(state, title, url) {
					try {
						if (url && typeof url === 'string' && (url.indexOf('http://') === 0 || url.indexOf('https://') === 0)) {
							return undefined;
						}
					} catch(e) { }
					try {
						return originalReplaceState.call(this, state, title, url);
					} catch(e) { }
				};
			}
		} catch(e) { }

		// Helper function to resolve relative URLs
		function resolveRelativeURL(url) {
			try {
				// If already absolute, return as-is
				if (url.indexOf('http://') === 0 || url.indexOf('https://') === 0) {
					return url;
				}
				// Protocol-relative URL
				if (url.indexOf('//') === 0) {
					return 'https:' + url;
				}
				// Relative URL - resolve against base URL
				const base = new URL(ORIGINAL_BASE_URL);
				return new URL(url, base).href;
			} catch(e) {
				return url;
			}
		}

		// List of domains to skip proxying (analytics, ads, tracking)
		const SKIP_PROXY_DOMAINS = [
			'google-analytics.com',
			'googletagmanager.com',
			'googlesyndication.com',
			'googleadservices.com',
			'doubleclick.net',
			'facebook.com/tr',
			'connect.facebook.net',
			'analytics.twitter.com',
			't.co',
			'adform.net',
			'adnxs.com',
			'rubiconproject.com',
			'pubmatic.com',
			'criteo.com',
			'crwdcntrl.net',
			'cookielaw.org',
			'onetrust.com',
			'clarity.ms',
			'bing.com'
		];

		// Helper function to check if URL should be skipped
		function shouldSkipProxy(url) {
			try {
				const urlObj = new URL(url);
				const hostname = urlObj.hostname.toLowerCase();
				return SKIP_PROXY_DOMAINS.some(domain =>
					hostname === domain || hostname.endsWith('.' + domain)
				);
			} catch(e) {
				return false;
			}
		}

		// Intercept fetch() with error suppression
		try {
			const originalFetch = window.fetch;
			window.fetch = function(input, ...args) {
				let modifiedInput = input;
				try {
					let url = input;
					// Handle Request objects
					if (input && typeof input === 'object' && input.url) {
						url = input.url;
					}
					if (typeof url === 'string') {
						// Resolve relative URLs to absolute
						const absoluteUrl = resolveRelativeURL(url);

						// Only intercept external URLs (not our own proxy)
						if (absoluteUrl.indexOf(PROXY_ORIGIN) !== 0) {
							// Skip known analytics/ad/tracking domains
							if (shouldSkipProxy(absoluteUrl)) {
								// Don't intercept - let it fail naturally
								return originalFetch.call(this, input, ...args);
							}

							// Reduce noise - only log important requests
							if (!absoluteUrl.includes('/analytics') && !absoluteUrl.includes('/collect') && !absoluteUrl.includes('/rum')) {
								console.log('[Proxy] Intercepting fetch:', url, '->', absoluteUrl);
							}
							try {
								// Use base64 encoding to avoid URL encoding issues
								const proxyUrl = PROXY_ORIGIN + '/api/webpage/resource?url_b64=' + btoa(absoluteUrl) + '&referer_b64=' + btoa(ORIGINAL_BASE_URL);
								if (input && typeof input === 'object' && input.url) {
									modifiedInput = new Request(proxyUrl, input);
								} else {
									modifiedInput = proxyUrl;
								}
							} catch(e) { }
						}
					}
				} catch(e) { }
				try {
					return originalFetch.call(this, modifiedInput, ...args);
				} catch(e) {
					return Promise.reject(e);
				}
			};
			console.log('[Proxy] Fetch interceptor installed');
		} catch(e) { }

		// Intercept XMLHttpRequest with error suppression
		try {
			const originalXHROpen = XMLHttpRequest.prototype.open;
			XMLHttpRequest.prototype.open = function(method, url, ...args) {
				let modifiedUrl = url;
				try {
					if (typeof url === 'string') {
						// Resolve relative URLs to absolute
						const absoluteUrl = resolveRelativeURL(url);

						// Only intercept external URLs (not our own proxy)
						if (absoluteUrl.indexOf(PROXY_ORIGIN) !== 0) {
							// Skip known analytics/ad/tracking domains
							if (shouldSkipProxy(absoluteUrl)) {
								// Don't intercept - let it fail naturally
								return originalXHROpen.call(this, method, url, ...args);
							}

							// Reduce noise - only log important requests
							if (!absoluteUrl.includes('/analytics') && !absoluteUrl.includes('/collect') && !absoluteUrl.includes('/rum')) {
								console.log('[Proxy] Intercepting XHR:', method, url, '->', absoluteUrl);
							}
							try {
								// Use base64 encoding to avoid URL encoding issues
								modifiedUrl = PROXY_ORIGIN + '/api/webpage/resource?url_b64=' + btoa(absoluteUrl) + '&referer_b64=' + btoa(ORIGINAL_BASE_URL);
							} catch(e) { }
						}
					}
				} catch(e) { }
				try {
					return originalXHROpen.call(this, method, modifiedUrl, ...args);
				} catch(e) {
					throw e;
				}
			};
			console.log('[Proxy] XHR interceptor installed');
		} catch(e) { }

		// Intercept all link clicks to open in external browser
		try {
			document.addEventListener('click', function(e) {
				try {
					// Check if clicked element or its parents is a link with our marker
					let target = e.target;
					while (target && target !== document) {
						if (target.tagName === 'A' && target.hasAttribute('data-proxy-link')) {
							// This is our proxied link
							const href = target.getAttribute('href');
							if (href && href.startsWith('BROWSER-OPEN:')) {
								e.preventDefault();
								e.stopPropagation();
								e.stopImmediatePropagation();

								const urlToOpen = href.substring('BROWSER-OPEN:'.length);
								console.log('[Proxy] Opening link in browser:', urlToOpen);

								// Call our backend to open the URL
								fetch(PROXY_ORIGIN + '/api/browser/open?url=' + encodeURIComponent(urlToOpen), {
									method: 'GET',
									mode: 'cors'
								}).catch(err => {
									console.error('[Proxy] Failed to open URL:', err);
								});

								return false;
							}
						}
						target = target.parentElement;
					}
				} catch(err) {
					console.error('[Proxy] Error handling click:', err);
				}
			}, true); // Use capture phase
			console.log('[Proxy] Link click interceptor installed');
		} catch(e) {
			console.error('[Proxy] Failed to install link interceptor:', e);
		}
	})();
	</script>`

	// Add meta tags to block manifest and other external resource requests
	metaTags := `<meta name="manifest" content=""><link rel="manifest" href="about:blank">`

	// CRITICAL: DON'T use <base> tag - it causes issues with link resolution
	// Instead, we'll convert ALL relative URLs to absolute URLs in the backend
	// This ensures both resources and links work correctly
	baseTag := ``

	// Find <head> tag and insert our interception script FIRST, then meta tags (no base tag)
	headIndex := strings.Index(strings.ToLower(content), "<head>")
	if headIndex == -1 {
		// If no <head>, look for <html>
		htmlIndex := strings.Index(strings.ToLower(content), "<html>")
		if htmlIndex != -1 {
			htmlEndIndex := htmlIndex + strings.Index(content[htmlIndex:], ">") + 1
			content = content[:htmlEndIndex] + "<head>" + interceptionScript + baseTag + metaTags + "</head>" + content[htmlEndIndex:]
		}
	} else {
		headEndIndex := headIndex + strings.Index(content[headIndex:], ">") + 1
		// Insert interception script FIRST, then meta tags (no base tag)
		content = content[:headEndIndex] + interceptionScript + baseTag + metaTags + content[headEndIndex:]
	}

	// Rewrite script src attributes
	// log.Printf("[HTML Rewrite] Rewriting script src attributes...")
	content = rewriteAttribute(content, "script", "src", baseURL)

	// Rewrite link href attributes (for stylesheets)
	// log.Printf("[HTML Rewrite] Rewriting link href attributes...")
	content = rewriteLinkHref(content, baseURL)

	// DEBUG: Log a sample of the rewritten HTML to verify it's working
	// if strings.Contains(content, "/static/") || strings.Contains(content, "/cdn-cgi/") {
	// 	log.Printf("[HTML Rewrite] DEBUG: Found potential unrewritten URLs in HTML!")
	// 	// Find and log the first occurrence
	// 	if idx := strings.Index(content, "/static/"); idx != -1 {
	// 		start := max(idx - 100, 0)
	// 		end := min(idx + 200, len(content))
	// 		log.Printf("[HTML Rewrite] Context around /static/: %s", content[start:end])
	// 	}
	// 	if idx := strings.Index(content, "/cdn-cgi/"); idx != -1 {
	// 		start := max(idx - 100, 0)
	// 		end := min(idx + 200, len(content))
	// 		log.Printf("[HTML Rewrite] Context around /cdn-cgi/: %s", content[start:end])
	// 	}
	// }

	// First, convert lazy-loaded images to normal images
	// This ensures images load immediately without waiting for lazy loading scripts
	content = convertLazyImages(content)

	// Then rewrite img src attributes (now including the converted lazy images)
	content = rewriteAttribute(content, "img", "src", baseURL)

	// Rewrite iframe src attributes
	content = rewriteAttribute(content, "iframe", "src", baseURL)

	// Rewrite video src and poster attributes
	content = rewriteAttribute(content, "video", "src", baseURL)
	content = rewriteAttribute(content, "video", "poster", baseURL)

	// Rewrite audio src attributes
	content = rewriteAttribute(content, "audio", "src", baseURL)

	// Rewrite source src attributes (for video/audio)
	content = rewriteAttribute(content, "source", "src", baseURL)

	// Rewrite track src attributes
	content = rewriteAttribute(content, "track", "src", baseURL)

	// Rewrite embed src attributes
	content = rewriteAttribute(content, "embed", "src", baseURL)

	// Rewrite object data attributes
	content = rewriteAttribute(content, "object", "data", baseURL)

	// Rewrite action attributes in forms
	content = rewriteAttribute(content, "form", "action", baseURL)

	// Rewrite href attributes in anchor tags (for absolute URLs only)
	content = rewriteAnchorHref(content, baseURL)

	// Rewrite CSS in style tags
	content = rewriteStyleTags(content, baseURL)

	// Rewrite inline style attributes
	content = rewriteInlineStyles(content, baseURL)

	return []byte(content)
}

// convertLazyImages converts lazy-loaded images to normal images
// For images with data-original or data-src attributes, move those URLs to src
// This prevents lazy loading and ensures immediate display
func convertLazyImages(content string) string {
	// Match img tags with lazy loading attributes
	// We need to match any img tag that contains data-original or data-src
	// Use a two-step approach: find all img tags, then check if they have lazy attributes
	re := regexp.MustCompile(`<img[^>]*>`)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		// Check if this img tag has data-original or data-src attribute
		// Try double quotes first: data-original="..."
		doubleQuoteRe := regexp.MustCompile(`\s(data-original|data-src)\s*=\s*"([^"]*)"`)
		doubleQuoteMatch := doubleQuoteRe.FindStringSubmatch(match)

		var lazySrc, lazyQuote string

		if len(doubleQuoteMatch) >= 3 {
			// Found double-quoted attribute
			lazySrc = doubleQuoteMatch[2]
			lazyQuote = `"`
		} else {
			// Try single quotes: data-original='...'
			singleQuoteRe := regexp.MustCompile(`\s(data-original|data-src)\s*=\s*'([^']*)'`)
			singleQuoteMatch := singleQuoteRe.FindStringSubmatch(match)
			if len(singleQuoteMatch) >= 3 {
				lazySrc = singleQuoteMatch[2]
				lazyQuote = `'`
			} else {
				// Try unquoted: data-original=...
				unquotedRe := regexp.MustCompile(`\s(data-original|data-src)\s*=\s*([^\s>]+)`)
				unquotedMatch := unquotedRe.FindStringSubmatch(match)
				if len(unquotedMatch) >= 3 {
					lazySrc = unquotedMatch[2]
					lazyQuote = ""
				} else {
					// No lazy attribute found
					return match
				}
			}
		}

		// Build new img tag
		var newTag strings.Builder
		newTag.WriteString("<img ")

		// Copy all attributes except src, data-original, data-src, and lazy class
		// Parse attributes manually since Go regex has limitations
		attrs := parseHTMLAttributes(match)

		for _, attr := range attrs {
			// Skip lazy loading attributes
			if attr.Name == "data-original" || attr.Name == "data-src" {
				continue
			}

			// Handle class attribute - remove "lazy" from it
			if attr.Name == "class" {
				// Remove "lazy" from class value
				classValue := strings.ReplaceAll(attr.Value, "lazy", "")
				classValue = strings.TrimSpace(classValue)
				classValue = strings.ReplaceAll(classValue, "  ", " ")

				if classValue != "" {
					newTag.WriteString(fmt.Sprintf(`class="%s" `, classValue))
				}
				continue
			}

			// Skip the old src attribute, we'll add the new one
			if attr.Name == "src" {
				continue
			}

			// Copy other attributes (preserve original quote style)
			if attr.Quote == "" {
				newTag.WriteString(fmt.Sprintf(`%s=%s `, attr.Name, attr.Value))
			} else {
				newTag.WriteString(fmt.Sprintf(`%s=%s%s%s `, attr.Name, attr.Quote, attr.Value, attr.Quote))
			}
		}

		// Add the new src attribute with the lazy-loaded image URL
		if lazyQuote == "" {
			newTag.WriteString(fmt.Sprintf(`src=%s`, lazySrc))
		} else {
			newTag.WriteString(fmt.Sprintf(`src=%s%s%s`, lazyQuote, lazySrc, lazyQuote))
		}

		// Close the tag
		newTag.WriteString(">")

		return newTag.String()
	})
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

// rewriteAttribute rewrites a specific attribute in HTML tags
func rewriteAttribute(content, tag, attr, baseURL string) string {
	// Match all tags first
	tagRe := regexp.MustCompile(fmt.Sprintf(`<%s[^>]*>`, tag))

	matchCount := 0
	rewriteCount := 0

	result := tagRe.ReplaceAllStringFunc(content, func(match string) string {
		matchCount++
		// Try to find the attribute with double quotes
		doubleQuoteRe := regexp.MustCompile(fmt.Sprintf(`\s%s\s*=\s*"([^"]*)"`, attr))
		doubleQuoteMatch := doubleQuoteRe.FindStringSubmatch(match)

		var urlValue, quote string

		if len(doubleQuoteMatch) >= 2 {
			// Found double-quoted attribute
			urlValue = doubleQuoteMatch[1]
			quote = `"`
		} else {
			// Try single quotes
			singleQuoteRe := regexp.MustCompile(fmt.Sprintf(`\s%s\s*=\s*'([^']*)'`, attr))
			singleQuoteMatch := singleQuoteRe.FindStringSubmatch(match)
			if len(singleQuoteMatch) >= 2 {
				urlValue = singleQuoteMatch[1]
				quote = `'`
			} else {
				// Try unquoted
				unquotedRe := regexp.MustCompile(fmt.Sprintf(`\s%s\s*=\s*([^\s>]+)`, attr))
				unquotedMatch := unquotedRe.FindStringSubmatch(match)
				if len(unquotedMatch) >= 2 {
					urlValue = unquotedMatch[1]
					quote = ""
				} else {
					// Attribute not found
					return match
				}
			}
		}

		// Skip data: URLs, blob: URLs, and already proxied URLs
		if strings.HasPrefix(urlValue, "data:") ||
			strings.HasPrefix(urlValue, "blob:") ||
			strings.HasPrefix(urlValue, "/api/") ||
			strings.HasPrefix(urlValue, "#") {
			return match
		}

		rewriteCount++
		// if tag == "script" || tag == "link" {
		// 	log.Printf("[%s Rewrite] Rewriting %s %d: %s", strings.ToUpper(tag), attr, rewriteCount, urlValue)
		// }

		// Resolve relative URLs
		resolvedURL := resolveURL(urlValue, baseURL)

		// Create proxied URL with base64 encoding
		proxiedURL := fmt.Sprintf("/api/webpage/resource?url_b64=%s&referer_b64=%s",
			base64.StdEncoding.EncodeToString([]byte(resolvedURL)),
			base64.StdEncoding.EncodeToString([]byte(baseURL)))

		// Replace the URL in the match
		// Use regex to replace attribute value more reliably
		if quote != "" {
			// Quoted value - replace using regex for more flexibility
			attrPattern := regexp.MustCompile(`(` + attr + `)\s*=\s*` + regexp.QuoteMeta(quote) + regexp.QuoteMeta(urlValue) + regexp.QuoteMeta(quote))
			replacement := fmt.Sprintf(`%s=%s%s%s`, attr, quote, proxiedURL, quote)
			return attrPattern.ReplaceAllString(match, replacement)
		} else {
			// Unquoted value - match until whitespace or > character
			// We need to capture the delimiter (space or >) to preserve it
			attrPattern := regexp.MustCompile(`(` + attr + `)\s*=\s*` + regexp.QuoteMeta(urlValue) + `([\s>])`)
			replacement := fmt.Sprintf(`%s="%s"$2`, attr, proxiedURL)
			return attrPattern.ReplaceAllString(match, replacement)
		}
	})

	// if matchCount > 0 && (tag == "script" || tag == "link") {
	// 	log.Printf("[%s Rewrite] Found %d %s tags, rewrote %d %s attributes", strings.ToUpper(tag), matchCount, tag, rewriteCount, attr)
	// }

	return result
}

// rewriteLinkHref rewrites href attributes in link tags
func rewriteLinkHref(content, baseURL string) string {
	// Match all link tags
	tagRe := regexp.MustCompile(`<link[^>]*>`)

	matchCount := 0
	rewriteCount := 0

	result := tagRe.ReplaceAllStringFunc(content, func(match string) string {
		matchCount++
		// Try to find the href attribute with double quotes
		doubleQuoteRe := regexp.MustCompile(`\shref\s*=\s*"([^"]*)"`)
		doubleQuoteMatch := doubleQuoteRe.FindStringSubmatch(match)

		var urlValue, quote string

		if len(doubleQuoteMatch) >= 2 {
			// Found double-quoted attribute
			urlValue = doubleQuoteMatch[1]
			quote = `"`
		} else {
			// Try single quotes
			singleQuoteRe := regexp.MustCompile(`\shref\s*=\s*'([^']*)'`)
			singleQuoteMatch := singleQuoteRe.FindStringSubmatch(match)
			if len(singleQuoteMatch) >= 2 {
				urlValue = singleQuoteMatch[1]
				quote = `'`
			} else {
				// Try unquoted
				unquotedRe := regexp.MustCompile(`\shref\s*=\s*([^\s>]+)`)
				unquotedMatch := unquotedRe.FindStringSubmatch(match)
				if len(unquotedMatch) >= 2 {
					urlValue = unquotedMatch[1]
					quote = ""
				} else {
					// href attribute not found
					return match
				}
			}
		}

		// Skip data: URLs, blob: URLs, and already proxied URLs
		if strings.HasPrefix(urlValue, "data:") ||
			strings.HasPrefix(urlValue, "blob:") ||
			strings.HasPrefix(urlValue, "/api/") ||
			strings.HasPrefix(urlValue, "#") {
			return match
		}

		rewriteCount++
		// log.Printf("[Link Rewrite] Rewriting link %d: %s", rewriteCount, urlValue)

		// Resolve relative URLs
		resolvedURL := resolveURL(urlValue, baseURL)

		// Create proxied URL with base64 encoding
		proxiedURL := fmt.Sprintf("/api/webpage/resource?url_b64=%s&referer_b64=%s",
			base64.StdEncoding.EncodeToString([]byte(resolvedURL)),
			base64.StdEncoding.EncodeToString([]byte(baseURL)))

		// Replace the URL in the match
		// Use regex to replace href attribute value more reliably
		if quote != "" {
			// Quoted value - replace using regex for more flexibility
			hrefPattern := regexp.MustCompile(`(href)\s*=\s*` + regexp.QuoteMeta(quote) + regexp.QuoteMeta(urlValue) + regexp.QuoteMeta(quote))
			replacement := fmt.Sprintf(`href=%s%s%s`, quote, proxiedURL, quote)
			return hrefPattern.ReplaceAllString(match, replacement)
		} else {
			// Unquoted value - match and capture the trailing delimiter
			// Go's regexp doesn't support lookahead, so we match and preserve the trailing char
			hrefPattern := regexp.MustCompile(`(href)\s*=\s*` + regexp.QuoteMeta(urlValue) + `([\s>])`)
			replacement := fmt.Sprintf(`href="%s"$2`, proxiedURL)
			return hrefPattern.ReplaceAllString(match, replacement)
		}
	})

	// if matchCount > 0 {
	// 	log.Printf("[Link Rewrite] Found %d link tags, rewrote %d", matchCount, rewriteCount)
	// }

	return result
}

// rewriteAnchorHref rewrites href attributes in anchor tags
func rewriteAnchorHref(content, baseURL string) string {
	// Match all anchor tags with href attribute
	// This pattern matches any <a> tag that contains an href attribute
	re := regexp.MustCompile(`<a\s+[^>]*href[^>]*>`)

	// Count matches for debugging
	matchCount := 0
	proxiedCount := 0

	result := re.ReplaceAllStringFunc(content, func(match string) string {
		matchCount++

		// Extract href value using a more flexible regex
		hrefRe := regexp.MustCompile(`href\s*=\s*(?:["']([^"']+)["']|([^"'\s>]+))`)
		hrefMatch := hrefRe.FindStringSubmatch(match)

		if len(hrefMatch) < 2 {
			return match
		}

		// Get the URL (first captured group is quoted, second is unquoted)
		var urlValue string
		if hrefMatch[1] != "" {
			urlValue = hrefMatch[1]
		} else if hrefMatch[2] != "" {
			urlValue = hrefMatch[2]
		} else {
			return match
		}

		// Skip mailto:, tel:, javascript:, and other special protocols
		if strings.HasPrefix(urlValue, "mailto:") ||
			strings.HasPrefix(urlValue, "tel:") ||
			strings.HasPrefix(urlValue, "javascript:") ||
			strings.HasPrefix(urlValue, "data:") ||
			strings.HasPrefix(urlValue, "blob:") ||
			strings.HasPrefix(urlValue, "#") {
			return match
		}

		// Skip already proxied URLs (both relative and absolute)
		if strings.HasPrefix(urlValue, "/api/") ||
			strings.HasPrefix(urlValue, "http://") && strings.Contains(urlValue, "/api/") ||
			strings.HasPrefix(urlValue, "https://") && strings.Contains(urlValue, "/api/") {
			return match
		}

		// Resolve relative URLs to absolute URLs
		var resolvedURL string
		if strings.HasPrefix(urlValue, "http://") || strings.HasPrefix(urlValue, "https://") {
			// Already absolute
			resolvedURL = urlValue
		} else if strings.HasPrefix(urlValue, "//") {
			// Protocol-relative URL (//example.com) - add https:
			resolvedURL = "https:" + urlValue
		} else {
			// Relative URL - resolve against baseURL
			resolvedURL = resolveURL(urlValue, baseURL)
		}

		// Only proxy HTTP/HTTPS URLs
		if !strings.HasPrefix(resolvedURL, "http://") && !strings.HasPrefix(resolvedURL, "https://") {
			return match
		}

		proxiedCount++

		// CRITICAL FIX: Use absolute URL for the proxy endpoint
		// We must use window.location.origin (which will be our backend) + the endpoint path
		// This is handled by JavaScript in the iframe, but we need to ensure the href is absolute
		// Format: https://localhost:9245/api/browser/open?url=...
		// Since we don't know our backend origin here, we use a protocol-relative URL starting with //
		// But actually, the iframe's window.location.origin IS our backend, so we just need to ensure
		// the path starts with / and it will be resolved correctly... wait, that's the problem!

		// The real solution: We need to construct an absolute URL or use JavaScript to handle clicks
		// For now, let's use a data URL or JavaScript approach, but simpler: make it absolute by using
		// the current origin. Since we're in an iframe served from our backend, we can use:

		// Option 1: Use JavaScript: href (blocked by CSP usually)
		// Option 2: Inject a <base> tag pointing to our backend (might break other resources)
		// Option 3: Use absolute URL with placeholder that JavaScript will fix
		// Option 4: Intercept clicks via event delegation (best approach)

		// Let's use a special marker that our injected script will recognize
		proxiedURL := fmt.Sprintf("BROWSER-OPEN:%s", resolvedURL)
		// log.Printf("[Link Proxy] Proxied link %d: %s -> %s (marker: %s)", proxiedCount, urlValue, resolvedURL, proxiedURL)

		// Replace the href attribute value by finding and replacing the exact value
		// We need to handle different quote styles
		if hrefMatch[1] != "" {
			// Was quoted - replace with quoted version
			// Find the original href attribute with its quotes and value
			oldHref := fmt.Sprintf(`href="%s"`, urlValue)
			oldHrefSingle := fmt.Sprintf(`href='%s'`, urlValue)
			if strings.Contains(match, oldHref) {
				newMatch := strings.Replace(match, oldHref, fmt.Sprintf(`href="%s"`, proxiedURL), 1)
				// Add target="_self" and a special data attribute for our script
				if !strings.Contains(newMatch, "target=") {
					newMatch = strings.Replace(newMatch, ">", ` target="_self" data-proxy-link="true">`, 1)
				} else {
					newMatch = strings.Replace(newMatch, ">", ` data-proxy-link="true">`, 1)
				}
				return newMatch
			} else if strings.Contains(match, oldHrefSingle) {
				newMatch := strings.Replace(match, oldHrefSingle, fmt.Sprintf(`href='%s'`, proxiedURL), 1)
				if !strings.Contains(newMatch, "target=") {
					newMatch = strings.Replace(newMatch, ">", ` target="_self" data-proxy-link="true">`, 1)
				} else {
					newMatch = strings.Replace(newMatch, ">", ` data-proxy-link="true">`, 1)
				}
				return newMatch
			}
		} else {
			// Was unquoted
			oldHref := fmt.Sprintf(`href=%s`, urlValue)
			newMatch := strings.Replace(match, oldHref, fmt.Sprintf(`href="%s"`, proxiedURL), 1)
			if !strings.Contains(newMatch, "target=") {
				newMatch = strings.Replace(newMatch, ">", ` target="_self" data-proxy-link="true">`, 1)
			} else {
				newMatch = strings.Replace(newMatch, ">", ` data-proxy-link="true">`, 1)
			}
			return newMatch
		}

		// Fallback: use regex replacement
		newMatch := hrefRe.ReplaceAllString(match, fmt.Sprintf(`href="%s"`, proxiedURL))
		if !strings.Contains(newMatch, "target=") {
			newMatch = strings.Replace(newMatch, ">", ` target="_self" data-proxy-link="true">`, 1)
		} else {
			newMatch = strings.Replace(newMatch, ">", ` data-proxy-link="true">`, 1)
		}
		return newMatch
	})

	// if matchCount > 0 {
	// 	log.Printf("[Link Proxy] Found %d anchor tags, proxied %d links", matchCount, proxiedCount)
	// }

	return result
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
	pattern := `style\s*=\s*(["'])(.*?)\1`
	pattern = strings.ReplaceAll(pattern, `\1`, `\$1`) // Fix backreference
	re := regexp.MustCompile(pattern)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		subPattern := `style\s*=\s*(["'])(.*?)\1`
		subPattern = strings.ReplaceAll(subPattern, `\1`, `\$1`) // Fix backreference
		subRe := regexp.MustCompile(subPattern)
		matches := subRe.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match
		}

		quote := matches[1]
		styleContent := matches[2]
		// Rewrite @font-face rules first
		styleContent = rewriteFontFaceRules(styleContent, baseURL)
		// Then rewrite all other url() references
		rewrittenStyle := rewriteCSSURLs(styleContent, baseURL)

		return fmt.Sprintf(`style=%s%s%s`, quote, rewrittenStyle, quote)
	})
}

// rewriteCSSURLs rewrites url() references in CSS
func rewriteCSSURLs(css, baseURL string) string {
	// First, handle @import rules
	// @import can be: @import "url"; or @import url("url");
	// We need to handle multiple patterns separately
	importRe1 := regexp.MustCompile(`@import\s+url\(['"]([^'"]+)['"]\)`)
	css = importRe1.ReplaceAllStringFunc(css, func(match string) string {
		subMatches := importRe1.FindStringSubmatch(match)
		if len(subMatches) < 2 {
			return match
		}

		urlValue := subMatches[1]

		// Skip data: URLs and already proxied URLs
		if strings.HasPrefix(urlValue, "data:") ||
			strings.HasPrefix(urlValue, "/api/") {
			return match
		}

		// Resolve relative URLs
		resolvedURL := resolveURL(urlValue, baseURL)

		// Create proxied URL with base64 encoding
		proxiedURL := fmt.Sprintf("/api/webpage/resource?url_b64=%s&referer_b64=%s",
			base64.StdEncoding.EncodeToString([]byte(resolvedURL)),
			base64.StdEncoding.EncodeToString([]byte(baseURL)))

		return fmt.Sprintf(`@import url("%s")`, proxiedURL)
	})

	importRe2 := regexp.MustCompile(`@import\s+['"]([^'"]+)['"]`)
	css = importRe2.ReplaceAllStringFunc(css, func(match string) string {
		subMatches := importRe2.FindStringSubmatch(match)
		if len(subMatches) < 2 {
			return match
		}

		urlValue := subMatches[1]

		// Skip data: URLs and already proxied URLs
		if strings.HasPrefix(urlValue, "data:") ||
			strings.HasPrefix(urlValue, "/api/") {
			return match
		}

		// Resolve relative URLs
		resolvedURL := resolveURL(urlValue, baseURL)

		// Create proxied URL with base64 encoding
		proxiedURL := fmt.Sprintf("/api/webpage/resource?url_b64=%s&referer_b64=%s",
			base64.StdEncoding.EncodeToString([]byte(resolvedURL)),
			base64.StdEncoding.EncodeToString([]byte(baseURL)))

		return fmt.Sprintf(`@import url("%s")`, proxiedURL)
	})

	// Then handle url(...) patterns in CSS
	pattern := `url\((['"]?)([^'")]+)\1\)`
	pattern = strings.ReplaceAll(pattern, `\1`, `\$1`) // Fix backreference
	re := regexp.MustCompile(pattern)

	return re.ReplaceAllStringFunc(css, func(match string) string {
		subMatches := re.FindStringSubmatch(match)
		if len(subMatches) < 3 {
			return match
		}

		urlValue := subMatches[2]

		// Skip data: URLs and already proxied URLs
		if strings.HasPrefix(urlValue, "data:") ||
			strings.HasPrefix(urlValue, "/api/") {
			return match
		}

		// Resolve relative URLs
		resolvedURL := resolveURL(urlValue, baseURL)

		// Create proxied URL with base64 encoding
		proxiedURL := fmt.Sprintf("/api/webpage/resource?url_b64=%s&referer_b64=%s",
			base64.StdEncoding.EncodeToString([]byte(resolvedURL)),
			base64.StdEncoding.EncodeToString([]byte(baseURL)))

		return fmt.Sprintf(`url(%s)`, proxiedURL)
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
	// Handle CORS preflight requests
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet && r.Method != http.MethodPost {
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

	// Create request to the resource URL
	var req *http.Request
	var err error

	// For POST requests, read the body
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Failed to read request body: %v", err)
			response.Error(w, err, http.StatusInternalServerError)
			return
		}
		req, err = http.NewRequest("POST", resourceURL, bytes.NewReader(body))
		if err != nil {
			log.Printf("Failed to create request: %v", err)
			response.Error(w, err, http.StatusInternalServerError)
			return
		}
		// Forward content type
		if contentType := r.Header.Get("Content-Type"); contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
	} else {
		req, err = http.NewRequest("GET", resourceURL, nil)
		if err != nil {
			log.Printf("Failed to create request: %v", err)
			response.Error(w, err, http.StatusInternalServerError)
			return
		}
	}

	// Set headers to mimic a browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", referer)
	req.Header.Set("Accept", "*/*")

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
	ext := strings.ToLower(filepath.Ext(resourceURL))
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

	// Copy headers from the response, excluding problematic ones
	for key, values := range resp.Header {
		// Skip headers that might cause issues
		if key == "Content-Security-Policy" ||
			key == "X-Frame-Options" ||
			key == "Set-Cookie" ||
			key == "Access-Control-Allow-Origin" ||
			key == "Content-Length" || // We'll recalculate this
			key == "Content-Type" { // We'll set this ourselves
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set CORS headers to allow loading from the same origin
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")

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
		cssContent := rewriteFontFaceRules(string(bodyBytes), referer)
		// Rewrite all url() references
		cssContent = rewriteCSSURLs(cssContent, referer)

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
func proxyMediaDirectly(mediaURL, referer string, w http.ResponseWriter) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", mediaURL, nil)
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

	resp, err := client.Do(req)
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
func getContentTypeFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
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

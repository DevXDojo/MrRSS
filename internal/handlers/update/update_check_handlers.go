package update

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"runtime"
	"strings"
	"syscall"
	"time"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/utils"
	"MrRSS/internal/version"
)

// isNetworkError checks if the error is related to network connectivity issues
// (e.g., timeout, connection refused, DNS failure) which often indicates firewall/blocking
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}

	// Check for timeout errors
	if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
		return true
	}

	// Check for connection refused, DNS issues, etc.
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
		// Check for specific network errors
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			// Connection refused, host unreachable, etc.
			if opErr.Op == "dial" || opErr.Op == "read" || opErr.Op == "write" {
				return true
			}
		}
	}

	// Check for syscall errors (connection refused, reset by peer, etc.)
	var sysErr syscall.Errno
	if errors.As(err, &sysErr) {
		switch sysErr {
		case syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.ECONNABORTED,
			syscall.ETIMEDOUT, syscall.EHOSTUNREACH:
			return true
		}
	}

	// Check for common error messages
	errStr := strings.ToLower(err.Error())
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "temporary failure") {
		return true
	}

	return false
}

// HandleCheckUpdates checks for the latest stable version on GitHub.
// Pre-release versions (alpha, beta) are filtered out.
// @Summary      Check for updates
// @Description  Check GitHub for the latest stable release version (pre-releases are filtered out)
// @Tags         update
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Update info (current_version, latest_version, update_available, download_url, release_notes)"
// @Failure      500  {object}  map[string]interface{}  "Error checking for updates"
// @Router       /update/check [get]
func HandleCheckUpdates(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	currentVersion := version.Version
	// Use /releases endpoint to get all releases, then filter for stable versions
	const githubAPI = "https://api.github.com/repos/WCY-dt/MrRSS/releases"

	// Create HTTP client with global proxy support
	var proxyURL string
	proxyEnabled, _ := h.DB.GetSetting("proxy_enabled")
	if proxyEnabled == "true" {
		// Build proxy URL from global settings (use encrypted methods for credentials)
		proxyType, _ := h.DB.GetSetting("proxy_type")
		proxyHost, _ := h.DB.GetSetting("proxy_host")
		proxyPort, _ := h.DB.GetSetting("proxy_port")
		proxyUsername, _ := h.DB.GetEncryptedSetting("proxy_username")
		proxyPassword, _ := h.DB.GetEncryptedSetting("proxy_password")
		proxyURL = utils.BuildProxyURL(proxyType, proxyHost, proxyPort, proxyUsername, proxyPassword)
	}

	client, err := utils.CreateHTTPClient(proxyURL, 30*time.Second)
	if err != nil {
		log.Printf("Error creating HTTP client: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"current_version": currentVersion,
			"error":           "Failed to create HTTP client",
		})
		return
	}

	resp, err := client.Get(githubAPI)
	if err != nil {
		log.Printf("Error checking for updates: %v", err)
		errorType := "error_checking_updates"
		if isNetworkError(err) {
			errorType = "network_error"
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"current_version": currentVersion,
			"error":           errorType,
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("GitHub API returned status: %d", resp.StatusCode)
		// Non-200 status codes often indicate network/proxy issues in China
		errorType := "fetch_failed"
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusProxyAuthRequired {
			errorType = "network_error"
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"current_version": currentVersion,
			"error":           errorType,
		})
		return
	}

	type Release struct {
		TagName     string `json:"tag_name"`
		Name        string `json:"name"`
		HTMLURL     string `json:"html_url"`
		Body        string `json:"body"`
		PublishedAt string `json:"published_at"`
		Prerelease  bool   `json:"prerelease"`
		Draft       bool   `json:"draft"`
		Assets      []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}

	var releases []Release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		log.Printf("Error decoding releases info: %v", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"current_version": currentVersion,
			"error":           "Failed to parse release information",
		})
		return
	}

	// Find the latest stable release (not prerelease, not draft)
	// Compare versions to ensure we get the actual latest, not just the first one
	var release Release
	var latestVersion string
	found := false
	for _, r := range releases {
		if !r.Prerelease && !r.Draft {
			version := strings.TrimPrefix(r.TagName, "v")
			if !found || compareVersions(version, latestVersion) > 0 {
				release = r
				latestVersion = version
				found = true
			}
		}
	}

	if !found {
		log.Printf("No stable releases found")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"current_version": currentVersion,
			"error":           "No stable release available",
		})
		return
	}

	// Check if there's an update available
	hasUpdate := compareVersions(latestVersion, currentVersion) > 0

	// Find the appropriate download URL based on platform
	var downloadURL string
	var assetName string
	var assetSize int64
	platform := runtime.GOOS
	arch := runtime.GOARCH
	isPortable := utils.IsPortableMode()

	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)

		// Match platform-specific installer/package with architecture
		// Asset naming convention:
		//   Installer: MrRSS-{version}-{platform}-{arch}-installer.{ext}
		//   Portable: MrRSS-{version}-{platform}-{arch}-portable.{ext}
		platformArch := platform + "-" + arch

		if platform == "windows" {
			if isPortable {
				// For portable Windows, download .zip
				if strings.Contains(name, platformArch) && strings.HasSuffix(name, "-portable.zip") {
					downloadURL = asset.BrowserDownloadURL
					assetName = asset.Name
					assetSize = asset.Size
					break
				}
			} else {
				// For installed Windows, prefer installer.exe
				if strings.Contains(name, platformArch) && strings.HasSuffix(name, "-installer.exe") {
					downloadURL = asset.BrowserDownloadURL
					assetName = asset.Name
					assetSize = asset.Size
					break
				}
			}
		} else if platform == "linux" {
			if isPortable {
				// For portable Linux, download .tar.gz
				if strings.Contains(name, platformArch) && strings.HasSuffix(name, "-portable.tar.gz") {
					downloadURL = asset.BrowserDownloadURL
					assetName = asset.Name
					assetSize = asset.Size
					break
				}
			} else {
				// For installed Linux, prefer .AppImage
				if strings.Contains(name, platformArch) && strings.HasSuffix(name, ".appimage") {
					downloadURL = asset.BrowserDownloadURL
					assetName = asset.Name
					assetSize = asset.Size
					break
				}
			}
		} else if platform == "darwin" {
			if isPortable {
				// For portable macOS, download .zip
				if strings.Contains(name, platformArch) && strings.HasSuffix(name, "-portable.zip") {
					downloadURL = asset.BrowserDownloadURL
					assetName = asset.Name
					assetSize = asset.Size
					break
				}
			} else {
				// For installed macOS, use universal build DMG
				if strings.Contains(name, "darwin-universal") && strings.HasSuffix(name, ".dmg") {
					downloadURL = asset.BrowserDownloadURL
					assetName = asset.Name
					assetSize = asset.Size
					break
				}
			}
		}
	}

	response := map[string]interface{}{
		"current_version": currentVersion,
		"latest_version":  latestVersion,
		"has_update":      hasUpdate,
		"platform":        platform,
		"arch":            arch,
		"is_portable":     isPortable,
	}

	if downloadURL != "" {
		response["download_url"] = downloadURL
		response["asset_name"] = assetName
		response["asset_size"] = assetSize
	}

	json.NewEncoder(w).Encode(response)
}

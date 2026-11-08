package update

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"MrRSS/internal/handlers/core"
	"MrRSS/internal/handlers/response"
)

var verifiedUpdates sync.Map

type verifiedUpdateKey struct {
	handler *core.Handler
	path    string
}

func authorizeUpdate(h *core.Handler, w http.ResponseWriter, r *http.Request) bool {
	if h == nil || h.UpdateAuthorization == "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+h.UpdateAuthorization)) != 1 {
		response.Error(w, fmt.Errorf("desktop update authorization required"), http.StatusForbidden)
		return false
	}
	return true
}

func parseOfficialAsset(raw, assetName string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host != "github.com" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.RawPath != "" || u.EscapedPath() != u.Path || strings.ContainsAny(u.Path, "\\\x00") {
		return "", fmt.Errorf("invalid download URL")
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) != 6 || parts[0] != "DevXDojo" || parts[1] != "MrRSS" || parts[2] != "releases" || parts[3] != "download" || parts[5] != assetName {
		return "", fmt.Errorf("invalid official release asset")
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", fmt.Errorf("invalid release path")
		}
	}
	return parts[4], nil
}

// The digest is obtained from GitHub's authenticated HTTPS API, never from the
// caller or an asset from another repository. Missing digests fail closed.
func officialAssetDigest(ctx context.Context, client *http.Client, raw, name string) (string, error) {
	tag, err := parseOfficialAsset(raw, name)
	if err != nil {
		return "", err
	}
	metadataClient := *client
	metadataClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/DevXDojo/MrRSS/releases/tags/"+url.PathEscape(tag), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "MrRSS-Updater")
	resp, err := metadataClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("release metadata unavailable (%d)", resp.StatusCode)
	}
	var release struct {
		Tag        string `json:"tag_name"`
		Draft      bool   `json:"draft"`
		Prerelease bool   `json:"prerelease"`
		Assets     []struct {
			Name   string `json:"name"`
			URL    string `json:"browser_download_url"`
			Digest string `json:"digest"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&release); err != nil {
		return "", fmt.Errorf("invalid release metadata: %w", err)
	}
	if release.Tag != tag || release.Draft || release.Prerelease {
		return "", fmt.Errorf("invalid stable release")
	}
	for _, asset := range release.Assets {
		if asset.Name != name || asset.URL != raw {
			continue
		}
		digest := strings.TrimPrefix(asset.Digest, "sha256:")
		decoded, err := hex.DecodeString(digest)
		if !strings.HasPrefix(asset.Digest, "sha256:") || err != nil || len(decoded) != sha256.Size {
			return "", fmt.Errorf("release asset has no valid SHA-256 digest")
		}
		return strings.ToLower(digest), nil
	}
	return "", fmt.Errorf("asset does not belong to the official release")
}

func restrictUpdateRedirects(client *http.Client, original string) {
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many update redirects")
		}
		u := req.URL
		if u.Scheme != "https" || u.User != nil || u.Port() != "" {
			return fmt.Errorf("invalid update redirect")
		}
		switch u.Host {
		case "github.com":
			if u.String() != original {
				return fmt.Errorf("unexpected GitHub update redirect")
			}
		case "release-assets.githubusercontent.com", "objects.githubusercontent.com", "github-releases.githubusercontent.com":
		default:
			return fmt.Errorf("untrusted update redirect")
		}
		return nil
	}
}

func verifyUpdateDigest(ctx context.Context, path, want string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, &updateContextReader{ctx: ctx, reader: file}); err != nil {
		return err
	}
	if hex.EncodeToString(hash.Sum(nil)) != want {
		return fmt.Errorf("update SHA-256 verification failed")
	}
	return nil
}

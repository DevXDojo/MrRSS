package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"MrRSS/internal/handlers/core"
)

func TestOfficialAssetRejectsTraversal(t *testing.T) {
	base := "https://github.com/DevXDojo/MrRSS/releases/download/"
	for _, raw := range []string{
		base + "../../../../attacker/repo/releases/download/v1/app.zip",
		base + "v1/%2e%2e/app.zip", base + "v1%2f..%2fapp.zip",
		base + "v1/app.zip?redirect=evil", base + "v1/app.zip#fragment",
		"https://github.com.evil/DevXDojo/MrRSS/releases/download/v1/app.zip",
		"https://github.com:443/DevXDojo/MrRSS/releases/download/v1/app.zip",
		base + "v1/other.zip",
	} {
		if _, err := parseOfficialAsset(raw, "app.zip"); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	if tag, err := parseOfficialAsset(base+"v1/app.zip", "app.zip"); err != nil || tag != "v1" {
		t.Fatal(tag, err)
	}
}

func TestUpdateRoutesRequireDesktopCapability(t *testing.T) {
	for _, handler := range []func(*core.Handler, http.ResponseWriter, *http.Request){HandleDownloadUpdate, HandleInstallUpdate} {
		for _, configured := range []string{"", "capability"} {
			req := httptest.NewRequest(http.MethodPost, "/api/update", strings.NewReader(`{}`))
			w := httptest.NewRecorder()
			handler(&core.Handler{UpdateAuthorization: configured}, w, req)
			if w.Code != http.StatusForbidden {
				t.Fatalf("unauthorized status %d", w.Code)
			}
		}
	}
}

type securityTransport func(*http.Request) (*http.Response, error)

func (f securityTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestOfficialMetadataAndDigest(t *testing.T) {
	assetURL := "https://github.com/DevXDojo/MrRSS/releases/download/v1/app.zip"
	sum := sha256.Sum256([]byte("official binary"))
	digest := hex.EncodeToString(sum[:])
	for _, remoteDigest := range []string{"sha256:" + digest, "", "sha256:bad"} {
		client := &http.Client{Transport: securityTransport(func(r *http.Request) (*http.Response, error) {
			if r.URL.Host != "api.github.com" || r.URL.Path != "/repos/DevXDojo/MrRSS/releases/tags/v1" {
				t.Fatal("wrong metadata origin")
			}
			body := fmt.Sprintf(`{"tag_name":"v1","assets":[{"name":"app.zip","browser_download_url":%q,"digest":%q}]}`, assetURL, remoteDigest)
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
		})}
		got, err := officialAssetDigest(context.Background(), client, assetURL, "app.zip")
		if remoteDigest == "sha256:"+digest {
			if err != nil || got != digest {
				t.Fatal(got, err)
			}
		} else if err == nil {
			t.Fatal("unverified metadata accepted")
		}
	}
	path := filepath.Join(t.TempDir(), "asset.zip")
	if err := os.WriteFile(path, []byte("official binary"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := verifyUpdateDigest(context.Background(), path, digest); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := verifyUpdateDigest(context.Background(), path, digest); err == nil {
		t.Fatal("tampering accepted")
	}
	client := &http.Client{}
	restrictUpdateRedirects(client, assetURL)
	for _, raw := range []string{"https://evil.example/payload", "http://release-assets.githubusercontent.com/payload", "https://github.com/attacker/repo"} {
		req, _ := http.NewRequest(http.MethodGet, raw, nil)
		if err := client.CheckRedirect(req, nil); err == nil {
			t.Fatal("untrusted redirect accepted")
		}
	}
}

func TestInstallRejectsUnregisteredAndTamperedDownloads(t *testing.T) {
	path, err := createUpdateDownloadPath("test.zip")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(filepath.Dir(path))
	if err := os.WriteFile(path, []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	h := &core.Handler{UpdateAuthorization: "capability"}
	for _, registered := range []bool{false, true} {
		if registered {
			verifiedUpdates.Store(verifiedUpdateKey{h, path}, strings.Repeat("0", 64))
		}
		body, _ := json.Marshal(map[string]string{"file_path": path})
		req := httptest.NewRequest(http.MethodPost, "/api/install-update", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer capability")
		w := httptest.NewRecorder()
		HandleInstallUpdate(h, w, req)
		if w.Code != 400 {
			t.Fatalf("invalid update status %d", w.Code)
		}
	}
}

func TestArchiveExtractionRejectsSiblingTraversalAndLinks(t *testing.T) {
	for _, name := range []string{"../extract-sibling/payload", "/absolute", "C:/payload", "dir\\..\\payload"} {
		for _, format := range []string{"zip", "tar"} {
			t.Run(format+name, func(t *testing.T) {
				root := t.TempDir()
				dest := filepath.Join(root, "extract")
				if err := os.Mkdir(dest, 0700); err != nil {
					t.Fatal(err)
				}
				var data bytes.Buffer
				path := filepath.Join(root, "archive")
				if format == "zip" {
					w := zip.NewWriter(&data)
					f, err := w.Create(name)
					if err != nil {
						t.Fatal(err)
					}
					_, _ = f.Write([]byte("payload"))
					_ = w.Close()
				} else {
					gz := gzip.NewWriter(&data)
					w := tar.NewWriter(gz)
					if err := w.WriteHeader(&tar.Header{Name: name, Mode: 0777, Size: 7}); err != nil {
						t.Fatal(err)
					}
					_, _ = w.Write([]byte("payload"))
					_ = w.Close()
					_ = gz.Close()
				}
				if err := os.WriteFile(path, data.Bytes(), 0600); err != nil {
					t.Fatal(err)
				}
				var err error
				if format == "zip" {
					err = extractZip(path, dest)
				} else {
					err = extractTarGz(path, dest)
				}
				if err == nil {
					t.Fatal("unsafe archive accepted")
				}
			})
		}
	}
}

func TestArchiveExtractionRejectsSymlinks(t *testing.T) {
	for _, format := range []string{"zip", "tar"} {
		t.Run(format, func(t *testing.T) {
			root := t.TempDir()
			dest := filepath.Join(root, "extract")
			if err := os.Mkdir(dest, 0700); err != nil {
				t.Fatal(err)
			}
			var data bytes.Buffer
			if format == "zip" {
				w := zip.NewWriter(&data)
				header := &zip.FileHeader{Name: "link"}
				header.SetMode(os.ModeSymlink | 0777)
				entry, err := w.CreateHeader(header)
				if err != nil {
					t.Fatal(err)
				}
				_, _ = entry.Write([]byte("../outside"))
				_ = w.Close()
			} else {
				gz := gzip.NewWriter(&data)
				w := tar.NewWriter(gz)
				if err := w.WriteHeader(&tar.Header{Name: "link", Typeflag: tar.TypeSymlink, Linkname: "../outside"}); err != nil {
					t.Fatal(err)
				}
				_ = w.Close()
				_ = gz.Close()
			}
			path := filepath.Join(root, "archive")
			if err := os.WriteFile(path, data.Bytes(), 0600); err != nil {
				t.Fatal(err)
			}
			var err error
			if format == "zip" {
				err = extractZip(path, dest)
			} else {
				err = extractTarGz(path, dest)
			}
			if err == nil {
				t.Fatal("archive symlink accepted")
			}
		})
	}
}

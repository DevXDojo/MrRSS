// Package siyuan exports articles through SiYuan's documented Markdown API.
package siyuan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

var blockID = regexp.MustCompile(`^\d{14}-[a-z0-9]{7}$`)

func ValidBlockID(id string) bool { return blockID.MatchString(id) }

func ParseEndpoint(raw string) (*url.URL, error) {
	endpoint, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || endpoint.Hostname() == "" || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return nil, fmt.Errorf("invalid SiYuan endpoint")
	}
	return endpoint, nil
}

func IsLoopback(endpoint *url.URL) bool {
	host := endpoint.Hostname()
	ip := net.ParseIP(host)
	return strings.EqualFold(host, "localhost") || (ip != nil && ip.IsLoopback())
}

// DocumentPath treats article titles as one path segment. The article ID keeps
// different articles with identical titles from sharing a destination.
func DocumentPath(folder, title string, articleID int64) (string, error) {
	folder = strings.TrimSpace(folder)
	if folder == "" {
		folder = "/MrRSS"
	}
	if !strings.HasPrefix(folder, "/") || strings.Contains(folder, "\\") || strings.IndexFunc(folder, unicode.IsControl) >= 0 {
		return "", fmt.Errorf("invalid SiYuan folder")
	}
	for _, segment := range strings.Split(folder, "/") {
		if segment == "." || segment == ".." {
			return "", fmt.Errorf("invalid SiYuan folder")
		}
	}
	title = strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || unicode.IsControl(r) {
			return ' '
		}
		return r
	}, title)
	title = strings.Join(strings.Fields(title), " ")
	if len([]rune(title)) > 120 {
		title = string([]rune(title)[:120])
	}
	title = strings.Trim(title, " .")
	if title == "" {
		title = "Article"
	}
	return fmt.Sprintf("%s/%s [%d]", strings.TrimRight(folder, "/"), title, articleID), nil
}

// CreateDocument does not retry POST requests, since a lost response can still
// mean that the document was created. SiYuan never overwrites an existing path.
func CreateDocument(ctx context.Context, client *http.Client, endpoint *url.URL, token, notebook, path, markdown string) (string, error) {
	if !ValidBlockID(notebook) {
		return "", fmt.Errorf("invalid SiYuan notebook ID")
	}
	body, err := json.Marshal(map[string]string{"notebook": notebook, "path": path, "markdown": markdown})
	if err != nil {
		return "", fmt.Errorf("encode SiYuan document: %w", err)
	}
	target := *endpoint
	target.Path = strings.TrimRight(target.Path, "/") + "/api/filetree/createDocWithMd"
	target.RawPath = ""
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create SiYuan request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}
	// Keep the token and private article on the configured service, even when it
	// returns a redirect to another host or path.
	boundedClient := *client
	boundedClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	res, err := boundedClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("connect to SiYuan: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("SiYuan returned HTTP %d", res.StatusCode)
	}
	var result struct {
		Code int    `json:"code"`
		Data string `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&result); err != nil {
		return "", fmt.Errorf("invalid SiYuan response")
	}
	if result.Code != 0 || !ValidBlockID(result.Data) {
		return "", fmt.Errorf("SiYuan did not create a document (code %d)", result.Code)
	}
	return result.Data, nil
}

package feed

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"MrRSS/internal/models"
	"golang.org/x/net/html"
)

// XPathPreviewNode describes the source tree and includes a sanitized snapshot.
// Paths refer to the original document, including omitted preview siblings.
type XPathPreviewNode struct {
	HTML     string              `json:"html,omitempty"`
	BaseURL  string              `json:"base_url,omitempty"`
	Path     string              `json:"path,omitempty"`
	Group    string              `json:"group,omitempty"`
	Classes  []string            `json:"classes,omitempty"`
	Tag      string              `json:"tag,omitempty"`
	Text     string              `json:"text,omitempty"`
	Link     string              `json:"link,omitempty"`
	Image    string              `json:"image,omitempty"`
	Date     string              `json:"date,omitempty"`
	Children []*XPathPreviewNode `json:"children,omitempty"`
}

func (f *Fetcher) PreviewXPathPage(ctx context.Context, source *models.Feed) (*XPathPreviewNode, error) {
	parsed, err := url.Parse(source.URL)
	if err != nil || parsed.Host == "" || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("invalid page URL")
	}
	client, err := f.getHTTPClient(*source)
	if err != nil {
		return nil, err
	}
	defer client.CloseIdleConnections()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 || req.URL.User != nil || (req.URL.Scheme != "http" && req.URL.Scheme != "https") {
			return fmt.Errorf("invalid redirect")
		}
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("page returned HTTP %d", resp.StatusCode)
	}
	const limit = 2 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if len(data) > limit {
		return nil, fmt.Errorf("page exceeds preview size limit")
	}
	return buildXPathPreview(string(data), resp.Request.URL.String())
}

var previewClassName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`)

func buildXPathPreview(source, base string) (*XPathPreviewNode, error) {
	doc, err := html.Parse(strings.NewReader(source))
	if err != nil {
		return nil, err
	}
	count := 0
	var visit func(*html.Node, string, int) (*XPathPreviewNode, error)
	visit = func(node *html.Node, path string, depth int) (*XPathPreviewNode, error) {
		count++
		if count > 12000 || depth > 80 {
			return nil, fmt.Errorf("page structure exceeds preview limit")
		}
		if node.Type == html.TextNode {
			return &XPathPreviewNode{Text: node.Data}, nil
		}
		if node.Type != html.ElementNode && node.Type != html.DocumentNode {
			return nil, nil
		}
		switch node.Data {
		case "head", "script", "style", "template", "noscript", "iframe", "object", "embed", "svg", "math", "link", "meta", "base", "input", "textarea", "select", "button":
			return nil, nil
		}
		result := &XPathPreviewNode{Path: path, Tag: node.Data}
		if index := strings.LastIndex(path, "["); index >= 0 {
			result.Group = path[:index]
		}
		for _, attr := range node.Attr {
			if attr.Namespace != "" {
				continue
			}
			switch attr.Key {
			case "class":
				for _, name := range strings.Fields(attr.Val) {
					if previewClassName.MatchString(name) {
						result.Classes = append(result.Classes, name)
					}
				}
				sort.Strings(result.Classes)
			case "href":
				result.Link = previewArticleURL(base, attr.Val)
			case "src":
				result.Image = previewArticleURL(base, attr.Val)
			case "alt":
				if node.Data == "img" {
					result.Text = attr.Val
				}
			case "datetime":
				result.Date = attr.Val
			}
		}
		for _, name := range result.Classes {
			result.Group += fmt.Sprintf("[contains(concat(' ', normalize-space(@class), ' '), ' %s ')]", name)
		}
		indices := map[string]int{}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			childPath := path
			if child.Type == html.ElementNode {
				indices[child.Data]++
				childPath = fmt.Sprintf("%s/%s[%d]", path, child.Data, indices[child.Data])
			}
			item, err := visit(child, childPath, depth+1)
			if err != nil {
				return nil, err
			}
			if item != nil {
				result.Children = append(result.Children, item)
			}
		}
		return result, nil
	}
	preview, err := visit(doc, "", 0)
	if err != nil {
		return nil, err
	}
	preview.HTML = renderXPathSnapshot(doc)
	preview.BaseURL = base
	return preview, nil
}

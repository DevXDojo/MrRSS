package feed

import (
	"fmt"
	"golang.org/x/net/html"
	"strings"
)

// renderXPathSnapshot preserves layout while removing executable content and
// navigation. The client must additionally use an opaque sandbox and nonce CSP.
// Paths are assigned before filtering, so they still address the source document.
func renderXPathSnapshot(doc *html.Node) string {
	var clone func(*html.Node, string) *html.Node
	clone = func(node *html.Node, path string) *html.Node {
		if node.Type == html.TextNode {
			return &html.Node{Type: html.TextNode, Data: node.Data}
		}
		if node.Type != html.ElementNode && node.Type != html.DocumentNode {
			return nil
		}
		if node.Namespace != "" {
			return nil
		}
		switch node.Data {
		case "script", "noscript", "template", "iframe", "frame", "frameset", "object", "embed", "svg", "math", "base", "meta", "audio", "video", "source":
			return nil
		}
		if node.Data == "link" {
			stylesheet := false
			for _, attr := range node.Attr {
				if attr.Key == "rel" && strings.EqualFold(attr.Val, "stylesheet") {
					stylesheet = true
				}
			}
			if !stylesheet {
				return nil
			}
		}
		out := &html.Node{Type: node.Type, Data: node.Data, DataAtom: node.DataAtom}
		for _, attr := range node.Attr {
			if attr.Namespace != "" {
				continue
			}
			switch attr.Key {
			case "class", "id", "style", "title", "alt", "width", "height", "colspan", "rowspan", "align", "valign", "cellpadding", "cellspacing", "border", "bgcolor", "color", "size", "face", "role", "dir", "lang", "hidden":
				out.Attr = append(out.Attr, attr)
			case "rel", "media":
				if node.Data == "link" {
					out.Attr = append(out.Attr, attr)
				}
			case "href":
				if node.Data == "link" && safeSnapshotResource(attr.Val) {
					out.Attr = append(out.Attr, attr)
				}
			case "src":
				if node.Data == "img" && safeSnapshotResource(attr.Val) {
					out.Attr = append(out.Attr, attr)
				}
			}
		}
		if node.Type == html.ElementNode {
			out.Attr = append(out.Attr, html.Attribute{Key: "data-mrrss-path", Val: path})
			if node.Data == "img" || node.Data == "link" {
				out.Attr = append(out.Attr, html.Attribute{Key: "referrerpolicy", Val: "no-referrer"})
			}
			switch node.Data {
			case "input", "button", "select", "textarea":
				out.Attr = append(out.Attr, html.Attribute{Key: "disabled", Val: ""})
			}
		}
		indices := map[string]int{}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			childPath := path
			if child.Type == html.ElementNode {
				indices[child.Data]++
				childPath = fmt.Sprintf("%s/%s[%d]", path, child.Data, indices[child.Data])
			}
			if result := clone(child, childPath); result != nil {
				out.AppendChild(result)
			}
		}
		return out
	}
	var output strings.Builder
	_ = html.Render(&output, clone(doc, ""))
	return output.String()
}

func safeSnapshotResource(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if strings.ContainsAny(value, "\x00\r\n\t") {
		return false
	}
	// Relative paths resolve against the trusted, escaped response URL in the iframe.
	colon := strings.Index(value, ":")
	slash := strings.IndexAny(value, "/?#")
	return colon < 0 || (slash >= 0 && slash < colon) || strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "http://")
}

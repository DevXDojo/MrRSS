package feed

import (
	"strings"
	"testing"

	"github.com/mmcdole/gofeed"
	"golang.org/x/text/encoding/simplifiedchinese"
)

func TestDecodeFeedBodyPrefersXMLDeclarationEncoding(t *testing.T) {
	xml := `<?xml version="1.0" encoding="gbk"?><rss><channel><title>创业邦</title></channel></rss>`
	gbkXML, err := simplifiedchinese.GBK.NewEncoder().String(xml)
	if err != nil {
		t.Fatalf("failed to encode test XML as GBK: %v", err)
	}

	decoded, err := decodeFeedBody([]byte(gbkXML), "text/xml; charset=utf-8")
	if err != nil {
		t.Fatalf("decodeFeedBody failed: %v", err)
	}

	if !strings.Contains(decoded, "创业邦") {
		t.Fatalf("expected GBK XML declaration to be honored, got: %q", decoded)
	}
	if !strings.Contains(decoded, `encoding="UTF-8"`) {
		t.Fatalf("expected decoded XML declaration to be normalized to UTF-8, got: %q", decoded)
	}

	parsed, err := gofeed.NewParser().ParseString(decoded)
	if err != nil {
		t.Fatalf("failed to parse decoded feed: %v", err)
	}
	if parsed.Title != "创业邦" {
		t.Fatalf("parsed title = %q, want %q", parsed.Title, "创业邦")
	}
}

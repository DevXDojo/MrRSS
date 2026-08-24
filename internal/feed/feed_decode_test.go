package feed

import (
	"strings"
	"testing"

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
}

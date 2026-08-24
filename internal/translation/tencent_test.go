package translation

import (
	"strings"
	"testing"
)

func TestTencentSignatureUsesDocumentedSignedHeaders(t *testing.T) {
	translator := NewTencentTranslator("AKIDEXAMPLE", "SECRET")
	auth, err := translator.calculateSignature(1551113065, []byte(`{"SourceText":"Hello","Source":"auto","Target":"zh","ProjectId":0}`))
	if err != nil {
		t.Fatalf("calculateSignature failed: %v", err)
	}

	if !strings.Contains(auth, "SignedHeaders=content-type;host") {
		t.Fatalf("expected documented signed headers, got: %s", auth)
	}
	if strings.Contains(auth, "x-tc-action") {
		t.Fatalf("x-tc-action should not be signed unless its value is canonicalized, got: %s", auth)
	}
}

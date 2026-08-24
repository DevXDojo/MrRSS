package translation

import (
	"encoding/json"
	"testing"
)

func TestCustomTranslatorReplacePlaceholdersAcceptsQuotedTextPlaceholder(t *testing.T) {
	translator := &CustomTranslator{}
	body := translator.replacePlaceholders(`{"q":"{{text}}","target":"{{target_lang}}"}`, `Hello "MrRSS"`, "zh")

	var payload map[string]string
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("body is not valid JSON: %v\n%s", err, body)
	}
	if payload["q"] != `Hello "MrRSS"` {
		t.Fatalf("q = %q", payload["q"])
	}
	if payload["target"] != "zh" {
		t.Fatalf("target = %q", payload["target"])
	}
}

func TestCustomTranslatorReplacePlaceholdersAcceptsRawTextPlaceholder(t *testing.T) {
	translator := &CustomTranslator{}
	body := translator.replacePlaceholders(`{"q":{{text}},"target":"{{target_lang}}"}`, "Hello", "tr")

	var payload map[string]string
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("body is not valid JSON: %v\n%s", err, body)
	}
	if payload["q"] != "Hello" {
		t.Fatalf("q = %q", payload["q"])
	}
	if payload["target"] != "tr" {
		t.Fatalf("target = %q", payload["target"])
	}
}

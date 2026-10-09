package translation

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	"github.com/longbridgeapp/opencc"
)

var simplifiedToTraditional = sync.OnceValues(func() (*opencc.OpenCC, error) {
	return opencc.New("s2t")
})

// isChineseScriptText accepts Han text with punctuation, numbers and reader
// placeholders. Other scripts, URLs and code still need the configured provider.
func isChineseScriptText(text string) bool {
	hasHan := false
	for _, r := range text {
		if unicode.In(r, unicode.Han) {
			hasHan = true
		} else if unicode.IsLetter(r) {
			return false
		}
	}
	return hasHan
}

// convertChineseScript handles pure Chinese locally, including unchanged text.
// It does not translate other languages or change the selected provider.
func convertChineseScript(text, targetLang string) (string, bool, error) {
	target := strings.ToLower(strings.TrimSpace(targetLang))
	if target == "zh-cn" {
		target = "zh"
	}
	if (target != "zh" && target != "zh-tw") || !isChineseScriptText(text) {
		return "", false, nil
	}
	var converter *opencc.OpenCC
	var err error
	if target == "zh-tw" {
		converter, err = simplifiedToTraditional()
	} else {
		converter, err = traditionalToSimplified()
	}
	if err != nil {
		return "", true, fmt.Errorf("initialize Chinese script conversion: %w", err)
	}
	converted, err := converter.Convert(text)
	if err != nil {
		return "", true, fmt.Errorf("convert Chinese script: %w", err)
	}
	return converted, true, nil
}

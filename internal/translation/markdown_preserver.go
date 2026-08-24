package translation

import (
	"regexp"
	"strings"
)

// TranslateMarkdownPreservingStructure translates markdown while preserving list structure
// This approach:
// 1. Extracts list markers and indentation
// 2. Translates only the content text
// 3. Reassembles with proper structure
func TranslateMarkdownPreservingStructure(markdown string, translator Translator, targetLang string) (string, error) {
	if markdown == "" {
		return "", nil
	}

	// Check if it contains list structures
	if !containsListStructure(markdown) {
		// No lists, translate directly
		return translator.Translate(markdown, targetLang)
	}

	// Parse into lines and translate line by line, preserving structure
	lines := strings.Split(markdown, "\n")
	var translatedLines []string

	for _, line := range lines {
		translatedLine, err := translateLinePreservingStructure(line, translator, targetLang)
		if err != nil {
			// If translation fails, keep original line
			translatedLines = append(translatedLines, line)
			continue
		}
		translatedLines = append(translatedLines, translatedLine)
	}

	return strings.Join(translatedLines, "\n"), nil
}

// containsListStructure checks if text contains markdown lists
func containsListStructure(text string) bool {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		// Check for unordered list (-, *, +)
		if regexp.MustCompile(`^\s*[-*+]\s+`).MatchString(line) {
			return true
		}
		// Check for ordered list (1., 2., etc.)
		if regexp.MustCompile(`^\s*\d+\.\s+`).MatchString(line) {
			return true
		}
	}
	return false
}

// translateLinePreservingStructure translates a single line while preserving list markers
func translateLinePreservingStructure(line string, translator Translator, targetLang string) (string, error) {
	trimmed := line

	// Check for unordered list item: "- item" or "* item" or "+ item"
	ulPattern := regexp.MustCompile(`^(\s*)([-*+])(\s+)(.*)`)
	if ulMatches := ulPattern.FindStringSubmatch(trimmed); ulMatches != nil {
		indent := ulMatches[1]
		marker := ulMatches[2]
		space := ulMatches[3]
		content := ulMatches[4]

		if content == "" {
			return line, nil // Empty list item
		}

		// Translate only the content
		translatedContent, err := translator.Translate(content, targetLang)
		if err != nil {
			return "", err
		}

		return indent + marker + space + translatedContent, nil
	}

	// Check for ordered list item: "1. item"
	olPattern := regexp.MustCompile(`^(\s*)(\d+)(\.\s+)(.*)`)
	if olMatches := olPattern.FindStringSubmatch(trimmed); olMatches != nil {
		indent := olMatches[1]
		number := olMatches[2]
		space := olMatches[3]
		content := olMatches[4]

		if content == "" {
			return line, nil // Empty list item
		}

		// Translate only the content
		translatedContent, err := translator.Translate(content, targetLang)
		if err != nil {
			return "", err
		}

		return indent + number + space + translatedContent, nil
	}

	// Check for nested list (indented): "  - nested item"
	// The same patterns will catch this due to the indent capture group

	// Not a list item, translate the whole line
	if strings.TrimSpace(trimmed) != "" {
		return translator.Translate(trimmed, targetLang)
	}

	return line, nil
}

// TranslateMarkdownAIPrompt creates a specialized prompt for AI translation that preserves structure
func TranslateMarkdownAIPrompt(markdown string, translator Translator, targetLang string) (string, error) {
	if markdown == "" {
		return "", nil
	}

	// For AI translation, use a specialized prompt that emphasizes structure preservation
	aiTranslator, ok := translator.(*AITranslator)
	if !ok {
		// Not an AI translator, use standard preservation
		return TranslateMarkdownPreservingStructure(markdown, translator, targetLang)
	}

	// Create a system prompt that emphasizes preserving markdown structure
	originalPrompt := aiTranslator.SystemPrompt
	structurePrompt := `You are a translator. Translate the given text to the target language.
CRITICAL RULES:
1. Preserve ALL markdown formatting exactly as-is
2. Keep ALL list markers (-, *, +, 1., 2., etc.) in the same positions
3. Maintain EXACT indentation levels for nested lists
4. Do NOT translate markdown syntax, only translate the text content
5. Output ONLY the translated markdown, nothing else

Example:
Input:
- First item
  - Nested item
- Second item

Output (to Chinese):
- 第一项
  - 嵌套项目
- 第二项`

	aiTranslator.SetSystemPrompt(structurePrompt)

	// Translate
	result, err := aiTranslator.Translate(markdown, targetLang)

	// Restore original prompt
	aiTranslator.SetSystemPrompt(originalPrompt)

	return result, err
}

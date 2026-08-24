package translation

import (
	"strings"
	"testing"
)

// MockTranslator for testing
type TestTranslator struct {
	TranslateFunc func(text, targetLang string) (string, error)
}

func (t *TestTranslator) Translate(text, targetLang string) (string, error) {
	if t.TranslateFunc != nil {
		return t.TranslateFunc(text, targetLang)
	}
	// Simple mock: just add "[TRANSLATED]" prefix
	return "[TRANSLATED] " + text, nil
}

func TestContainsListStructure(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{
			name: "No list",
			text: "Just plain text",
			want: false,
		},
		{
			name: "Unordered list with dash",
			text: "- Item 1\n- Item 2",
			want: true,
		},
		{
			name: "Unordered list with asterisk",
			text: "* Item 1\n* Item 2",
			want: true,
		},
		{
			name: "Ordered list",
			text: "1. Item 1\n2. Item 2",
			want: true,
		},
		{
			name: "Nested list",
			text: "- Item 1\n  - Nested item",
			want: true,
		},
		{
			name: "Mixed content",
			text: "Plain text\n- List item\nMore text",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsListStructure(tt.text)
			if got != tt.want {
				t.Errorf("containsListStructure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranslateLinePreservingStructure(t *testing.T) {
	trans := &TestTranslator{
		TranslateFunc: func(text, targetLang string) (string, error) {
			return "Translated: " + text, nil
		},
	}

	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "Unordered list item",
			line:     "- Original text",
			expected: "- Translated: Original text",
		},
		{
			name:     "Ordered list item",
			line:     "1. Original text",
			expected: "1. Translated: Original text",
		},
		{
			name:     "Nested unordered list item",
			line:     "  - Nested text",
			expected: "  - Translated: Nested text",
		},
		{
			name:     "Nested ordered list item",
			line:     "    2. Nested item",
			expected: "    2. Translated: Nested item",
		},
		{
			name:     "Plain text",
			line:     "Just plain text",
			expected: "Translated: Just plain text",
		},
		{
			name:     "Empty list item",
			line:     "- ",
			expected: "- ",
		},
		{
			name:     "Asterisk list item",
			line:     "* Asterisk item",
			expected: "* Translated: Asterisk item",
		},
		{
			name:     "Plus list item",
			line:     "+ Plus item",
			expected: "+ Translated: Plus item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := translateLinePreservingStructure(tt.line, trans, "en")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("translateLinePreservingStructure() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTranslateMarkdownPreservingStructure(t *testing.T) {
	trans := &TestTranslator{
		TranslateFunc: func(text, targetLang string) (string, error) {
			// Simulate translation that adds " [译]" after each word
			words := strings.Fields(text)
			for i, word := range words {
				words[i] = word + " [译]"
			}
			return strings.Join(words, " "), nil
		},
	}

	tests := []struct {
		name     string
		markdown string
		validate func(t *testing.T, result string)
	}{
		{
			name:     "Simple unordered list",
			markdown: "- First item\n- Second item\n- Third item",
			validate: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				if len(lines) != 3 {
					t.Errorf("Expected 3 lines, got %d", len(lines))
				}
				for i, line := range lines {
					if !strings.HasPrefix(line, "- ") {
						t.Errorf("Line %d should start with '- ', got: %s", i, line)
					}
					if !strings.Contains(line, "[译]") {
						t.Errorf("Line %d should contain translation marker, got: %s", i, line)
					}
				}
			},
		},
		{
			name:     "Simple ordered list",
			markdown: "1. First item\n2. Second item\n3. Third item",
			validate: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				if len(lines) != 3 {
					t.Errorf("Expected 3 lines, got %d", len(lines))
				}
				for i, line := range lines {
					if !strings.HasPrefix(line, "1.") && !strings.HasPrefix(line, "2.") && !strings.HasPrefix(line, "3.") {
						t.Errorf("Line %d should start with number, got: %s", i, line)
					}
					if !strings.Contains(line, "[译]") {
						t.Errorf("Line %d should contain translation marker, got: %s", i, line)
					}
				}
			},
		},
		{
			name:     "Two-level nested list",
			markdown: "- Parent item 1\n  - Nested item 1\n  - Nested item 2\n- Parent item 2",
			validate: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				// Count indentation levels
				nonIndented := 0
				indented := 0
				for _, line := range lines {
					if strings.HasPrefix(line, "- ") {
						nonIndented++
					} else if strings.HasPrefix(line, "  - ") {
						indented++
					}
				}
				if nonIndented != 2 {
					t.Errorf("Expected 2 non-indented items, got %d", nonIndented)
				}
				if indented != 2 {
					t.Errorf("Expected 2 indented items, got %d", indented)
				}
			},
		},
		{
			name:     "Three-level nested list",
			markdown: "- Level 1\n  - Level 2\n    - Level 3\n- Back to Level 1",
			validate: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				// Check that indentation is preserved
				level1 := 0
				level2 := 0
				level3 := 0
				for _, line := range lines {
					if strings.HasPrefix(line, "- ") && !strings.HasPrefix(line, "  ") {
						level1++
					} else if strings.HasPrefix(line, "  - ") && !strings.HasPrefix(line, "    - ") {
						level2++
					} else if strings.HasPrefix(line, "    - ") {
						level3++
					}
				}
				if level1 != 2 {
					t.Errorf("Expected 2 level-1 items, got %d", level1)
				}
				if level2 != 1 {
					t.Errorf("Expected 1 level-2 item, got %d", level2)
				}
				if level3 != 1 {
					t.Errorf("Expected 1 level-3 item, got %d", level3)
				}
			},
		},
		{
			name:     "Mixed ordered and unordered",
			markdown: "1. First item\n   - Nested unordered\n2. Second item",
			validate: func(t *testing.T, result string) {
				lines := strings.Split(result, "\n")
				if len(lines) != 3 {
					t.Errorf("Expected 3 lines, got %d", len(lines))
				}
				// Check structure is preserved
				if !strings.HasPrefix(lines[0], "1.") {
					t.Errorf("Line 0 should be ordered list, got: %s", lines[0])
				}
				if !strings.HasPrefix(lines[1], "   - ") {
					t.Errorf("Line 1 should be nested unordered, got: %s", lines[1])
				}
				if !strings.HasPrefix(lines[2], "2.") {
					t.Errorf("Line 2 should be ordered list, got: %s", lines[2])
				}
			},
		},
		{
			name:     "Plain text without lists",
			markdown: "This is plain text\nWith multiple lines",
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "[译]") {
					t.Errorf("Plain text should be translated, got: %s", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := TranslateMarkdownPreservingStructure(tt.markdown, trans, "zh")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			tt.validate(t, result)
		})
	}
}

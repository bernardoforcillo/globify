package processor_test

import (
	"fmt"
	"testing"

	"github.com/bernardoforcillo/globify/internal/files"
	"github.com/bernardoforcillo/globify/internal/processor"
)

// mockTranslatorForAST is a specialized mock that preserves ICU elements during translation
type mockTranslatorForAST struct {
	// We'll define a custom translation behavior for ICU components
	preserveICUElements bool
}

func (m *mockTranslatorForAST) Translate(text, from, to string) (string, error) {
	// If we're preserving ICU elements, we need to handle them specially
	if m.preserveICUElements {
		// The simplest approach is to just prefix with the language code for testing
		return "[" + to + "] " + text, nil
	}
	
	// Basic translation - just add language prefix
	return "[" + to + "] " + text, nil
}

// TestASTProcessorDeep tests the ASTProcessor with complex ICU messages
func TestASTProcessorDeep(t *testing.T) {
	// Create a mock translator that preserves ICU elements
	mockTrans := &mockTranslatorForAST{preserveICUElements: true}
	
	// Test that we can create an ASTProcessor
	proc, err := processor.CreateProcessor("ast-json", mockTrans)
	if err != nil {
		t.Fatalf("Failed to create ASTProcessor: %v", err)
	}
	
	// Test cases with complex ICU messages
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple literal",
			input:    "Hello, world!",
			expected: "[fr] Hello, world!",
		},
		{
			name:     "With placeholder",
			input:    "Hello, {name}!",
			expected: "[fr] Hello, {name}[fr] !",
		},
		{
			name:     "With number format",
			input:    "You have {count, number} messages.",
			expected: "[fr] You have {count, number}[fr]  messages.",
		},
		{
			name:     "With date format",
			input:    "Sent on {date, date, short}.",
			expected: "[fr] Sent on {date, date, short}[fr] .",
		},
		{
			name:     "With HTML tags",
			input:    "This is <b>important</b> information.",
			expected: "[fr] This is <b>[fr] important</b>[fr]  information.",
		},
		{
			name:     "With plural format",
			input:    "You have {count, plural, one {# message} other {# messages}}.",
			expected: "[fr] You have {count, plural, one {#[fr]  message} other {#[fr]  messages} }[fr] .",
		},
		{
			name:     "Complex nested",
			input:    "Hello, {name}! You have {count, plural, one {<b>one</b> message} other {<b>{count}</b> messages}}.",
			expected: "[fr] Hello, {name}[fr] ! You have {count, plural, one {<b>[fr] one</b>[fr]  message} other {<b>{count}</b>[fr]  messages} }[fr] .",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create content map with a single key
			content := files.LanguageContent{
				"message": tt.input,
			}
			
			// Process the content
			result, err := proc.Execute(content, "en", "fr", files.LanguageContent{})
			if err != nil {
				t.Fatalf("ASTProcessor.Execute() error = %v", err)
			}
			
			// Check result
			if translated, ok := result["message"].(string); ok {
				// Compare with expected
				// Note: Exact matching might be fragile due to whitespace or format differences
				if translated != tt.expected {
					t.Errorf("ASTProcessor.Execute() = %q, want %q", translated, tt.expected)
				}
			} else {
				t.Errorf("Result is not a string: %v", result["message"])
			}
		})
	}
}

// TestASTProcessorWithNestedObjects tests handling of deeply nested objects
func TestASTProcessorWithNestedObjects(t *testing.T) {
	// Create a mock translator
	mockTrans := createMockTranslator()
	
	// Test that we can create an ASTProcessor
	proc, err := processor.CreateProcessor("ast-json", mockTrans)
	if err != nil {
		t.Fatalf("Failed to create ASTProcessor: %v", err)
	}
	
	// Create deeply nested content
	content := files.LanguageContent{
		"level1": map[string]interface{}{
			"text": "Level 1 text",
			"level2": map[string]interface{}{
				"text": "Level 2 text",
				"level3": map[string]interface{}{
					"text": "Level 3 text",
					"level4": map[string]interface{}{
						"text": "Level 4 text",
					},
				},
			},
		},
	}
	
	// Process the content
	result, err := proc.Execute(content, "en", "fr", files.LanguageContent{})
	if err != nil {
		t.Fatalf("ASTProcessor.Execute() error = %v", err)
	}
	
	// Helper to safely get nested map
	getMap := func(m files.LanguageContent, key string) (files.LanguageContent, error) {
		val, ok := m[key]
		if !ok {
			return nil, fmt.Errorf("missing key %s", key)
		}

		// Try casting to files.LanguageContent directly
		if lc, ok := val.(files.LanguageContent); ok {
			return lc, nil
		}

		// Try casting to map[string]interface{} and converting
		if mm, ok := val.(map[string]interface{}); ok {
			return files.LanguageContent(mm), nil
		}

		return nil, fmt.Errorf("key %s is not a map", key)
	}

	// Verify the result structure and translations
	level1, err := getMap(result, "level1")
	if err != nil {
		t.Fatalf("Result error for level1: %v", err)
	}
	
	if level1["text"] != "[fr] Level 1 text" {
		t.Errorf("level1.text = %v, want %v", level1["text"], "[fr] Level 1 text")
	}
	
	level2, err := getMap(level1, "level2")
	if err != nil {
		t.Fatalf("Result error for level2: %v", err)
	}
	
	if level2["text"] != "[fr] Level 2 text" {
		t.Errorf("level2.text = %v, want %v", level2["text"], "[fr] Level 2 text")
	}
	
	level3, err := getMap(level2, "level3")
	if err != nil {
		t.Fatalf("Result error for level3: %v", err)
	}
	
	if level3["text"] != "[fr] Level 3 text" {
		t.Errorf("level3.text = %v, want %v", level3["text"], "[fr] Level 3 text")
	}
	
	level4, err := getMap(level3, "level4")
	if err != nil {
		t.Fatalf("Result error for level4: %v", err)
	}
	
	if level4["text"] != "[fr] Level 4 text" {
		t.Errorf("level4.text = %v, want %v", level4["text"], "[fr] Level 4 text")
	}
}

// TestASTProcessorErrorHandling tests how the processor handles various error conditions
func TestASTProcessorErrorHandling(t *testing.T) {
	// Create an error-generating translator
	errorTranslator := createErrorMockTranslator()
	
	// Create the processor
	proc, err := processor.CreateProcessor("ast-json", errorTranslator)
	if err != nil {
		t.Fatalf("Failed to create ASTProcessor: %v", err)
	}
	
	// Test content with various potential error scenarios
	content := files.LanguageContent{
		"simple": "Simple text that will fail to translate",
		"withICU": "Text with {placeholder} that will fail to translate",
		"withTag": "Text with <b>bold</b> that will fail to translate",
		"nested": map[string]interface{}{
			"key": "Nested text that will fail to translate",
		},
	}
	
	// Execute the processor - it should not return an error at the top level
	// even though translations fail
	result, err := proc.Execute(content, "en", "fr", files.LanguageContent{})
	if err != nil {
		t.Fatalf("ASTProcessor.Execute() returned error = %v", err)
	}
	
	// Verify that the result contains the original values where translation failed
	if result["simple"] != "Simple text that will fail to translate" {
		t.Errorf("Expected original text for failed translation, got %v", result["simple"])
	}
	
	if result["withICU"] != "Text with {placeholder} that will fail to translate" {
		t.Errorf("Expected original text for failed ICU translation, got %v", result["withICU"])
	}
	
	if result["withTag"] != "Text with <b>bold</b> that will fail to translate" {
		t.Errorf("Expected original text for failed tag translation, got %v", result["withTag"])
	}
	
	// Helper to safely get nested map
	getMap := func(m files.LanguageContent, key string) (files.LanguageContent, error) {
		val, ok := m[key]
		if !ok {
			return nil, fmt.Errorf("missing key %s", key)
		}

		// Try casting to files.LanguageContent directly
		if lc, ok := val.(files.LanguageContent); ok {
			return lc, nil
		}

		// Try casting to map[string]interface{} and converting
		if mm, ok := val.(map[string]interface{}); ok {
			return files.LanguageContent(mm), nil
		}

		return nil, fmt.Errorf("key %s is not a map", key)
	}

	nested, err := getMap(result, "nested")
	if err != nil {
		t.Fatalf("Result error for nested: %v", err)
	}
	
	if nested["key"] != "Nested text that will fail to translate" {
		t.Errorf("Expected original text for failed nested translation, got %v", nested["key"])
	}
}

package processor_test

import (
	"testing"

	"github.com/bernardoforcillo/globify/internal/files"
	"github.com/bernardoforcillo/globify/internal/processor"
)

// Benchmarks for Processor components

func BenchmarkSimpleProcessorExecute(b *testing.B) {
	// Create a mock translator for benchmarking
	mockTranslator := createMockTranslator()
	
	// Create a SimpleProcessor with the mock
	proc := processor.NewSimpleProcessor(mockTranslator)
	
	// Test data with varying complexity
	baseContent := files.LanguageContent{
		"greeting": "Hello",
		"farewell": "Goodbye",
		"nested": map[string]interface{}{
			"key1": "Nested value 1",
			"key2": "Nested value 2",
			"deeper": map[string]interface{}{
				"key3": "Deeply nested value",
			},
		},
		"array": []interface{}{
			"first",
			"second",
			"third",
		},
		"number": 42,
		"boolean": true,
	}
	
	// Empty previous translation to force all translations
	emptyPrevious := files.LanguageContent{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proc.Execute(baseContent, "en", "fr", emptyPrevious)
	}
}

func BenchmarkSimpleProcessorWithPreviousTranslations(b *testing.B) {
	// Create a mock translator for benchmarking
	mockTranslator := createMockTranslator()
	
	// Create a SimpleProcessor with the mock
	proc := processor.NewSimpleProcessor(mockTranslator)
	
	// Test data
	baseContent := files.LanguageContent{
		"greeting": "Hello",
		"farewell": "Goodbye",
		"nested": map[string]interface{}{
			"key1": "Nested value 1",
			"key2": "Nested value 2",
		},
		"number": 42,
		"boolean": true,
	}
	
	// Previous translations - simulating a partially translated file
	previousTranslation := files.LanguageContent{
		"greeting": "[fr] Hello", // This would be reused
		"nested": map[string]interface{}{
			"key1": "[fr] Nested value 1", // This would be reused
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proc.Execute(baseContent, "en", "fr", previousTranslation)
	}
}

func BenchmarkASTProcessorExecute(b *testing.B) {
	// Skip if we can't create the ASTProcessor
	mockTranslator := createMockTranslator()
	proc, err := processor.CreateProcessor("ast-json", mockTranslator)
	if err != nil {
		b.Skip("Failed to create ASTProcessor:", err)
	}
	
	// Test data with ICU formatted strings
	baseContent := files.LanguageContent{
		"simple": "Hello, world!",
		"withPlaceholder": "Hello, {name}!",
		"withNumber": "You have {count, number} messages.",
		"withPlural": "You have {count, plural, one {# message} other {# messages}}.",
		"withTags": "This is <b>bold</b> text.",
		"complex": "Hello, {name}! You have {count, number} {count, plural, one {message} other {messages}}.",
	}
	
	// Empty previous translation to force all translations
	emptyPrevious := files.LanguageContent{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proc.Execute(baseContent, "en", "fr", emptyPrevious)
	}
}

func BenchmarkASTProcessorWithPreviousTranslations(b *testing.B) {
	// Skip if we can't create the ASTProcessor
	mockTranslator := createMockTranslator()
	proc, err := processor.CreateProcessor("ast-json", mockTranslator)
	if err != nil {
		b.Skip("Failed to create ASTProcessor:", err)
	}
	
	// Test data with ICU formatted strings
	baseContent := files.LanguageContent{
		"simple": "Hello, world!",
		"withPlaceholder": "Hello, {name}!",
		"withNumber": "You have {count, number} messages.",
		"withPlural": "You have {count, plural, one {# message} other {# messages}}.",
		"withTags": "This is <b>bold</b> text.",
		"complex": "Hello, {name}! You have {count, number} {count, plural, one {message} other {messages}}.",
	}
	
	// Previous translations - simulating a partially translated file
	previousTranslation := files.LanguageContent{
		"simple": "[fr] Hello, world!",
		"withPlaceholder": "[fr] Hello, {name}!",
		"withTags": "[fr] This is <b>bold</b> text.",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = proc.Execute(baseContent, "en", "fr", previousTranslation)
	}
}
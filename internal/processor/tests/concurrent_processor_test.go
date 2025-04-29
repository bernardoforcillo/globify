package processor_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/bernardoforcillo/globify/internal/files"
	"github.com/bernardoforcillo/globify/internal/processor"
	"github.com/stretchr/testify/assert"
)

// failingMockTranslator creates a mock translator that fails with specified probability
func failingMockTranslator(failProbability float64) *MockTranslator {
	return &MockTranslator{
		MockTranslate: func(text, from, to string) (string, error) {
			// Simulate random failures based on probability
			if rand.Float64() < failProbability {
				return "", fmt.Errorf("simulated translation failure")
			}
			// Otherwise translate normally
			return fmt.Sprintf("[%s] %s", to, text), nil
		},
	}
}

func TestConcurrentProcessingCorrectness(t *testing.T) {
	// Create a mock translator for testing
	mockTranslator := createMockTranslator()
	
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
	
	// Test SimpleProcessor with different worker counts
	testCases := []struct {
		name       string
		workerSize int
	}{
		{"SingleWorker", 1},
		{"TwoWorkers", 2},
		{"FourWorkers", 4},
	}
	
	// Create baseline result with a standard processor
	simpleProc := processor.NewSimpleProcessor(mockTranslator)
	baselineResult, err := simpleProc.Execute(baseContent, "en", "fr", emptyPrevious)
	assert.NoError(t, err, "Baseline execution should not error")
	
	// Test with varying worker counts
	for _, tc := range testCases {
		t.Run("SimpleProcessor-"+tc.name, func(t *testing.T) {
			// Create processor with specific worker count
			procWithWorkers := processor.NewSimpleProcessor(mockTranslator)
			procWithWorkers.SetWorkerPoolSize(tc.workerSize)
			
			// Execute with concurrency
			concurrentResult, err := procWithWorkers.Execute(baseContent, "en", "fr", emptyPrevious)
			assert.NoError(t, err, "Concurrent execution should not error")
			
			// Verify results match the baseline using our compareMaps helper
			if !compareMaps(t, concurrentResult, baselineResult) {
				t.Error("Results do not match baseline")
			}
		})
	}
	
	// Now test ASTProcessor if available
	astProc, err := processor.CreateProcessor("ast-json", mockTranslator)
	if err != nil {
		t.Skip("ASTProcessor not available:", err)
	}
	
	// Create test data with ICU message formats for AST processor
	astContent := files.LanguageContent{
		"simple": "Hello, world!",
		"withPlaceholder": "Hello, {name}!",
		"withNumber": "You have {count, number} messages.",
		"withPlural": "You have {count, plural, one {# message} other {# messages}}.",
		"withTags": "This is <b>bold</b> text.",
		"complex": "Hello, {name}! You have {count, number} {count, plural, one {message} other {messages}}.",
	}
	
	// Create baseline result for AST processing
	astBaselineResult, err := astProc.Execute(astContent, "en", "fr", emptyPrevious)
	assert.NoError(t, err, "AST baseline execution should not error")
	
	// Test only if we can cast to ASTProcessor to access SetWorkerPoolSize
	if ap, ok := astProc.(*processor.ASTProcessor); ok {
		for _, tc := range testCases {
			t.Run("ASTProcessor-"+tc.name, func(t *testing.T) {
				// Set worker pool size
				ap.SetWorkerPoolSize(tc.workerSize)
				
				// Execute with concurrency
				concurrentResult, err := ap.Execute(astContent, "en", "fr", emptyPrevious)
				assert.NoError(t, err, "Concurrent AST execution should not error")
				
				// Verify results match the baseline
				if !compareMaps(t, concurrentResult, astBaselineResult) {
					t.Error("AST results do not match baseline")
				}
			})
		}
	}
}

// TestConcurrentTranslationFailureHandling tests that our concurrent implementations
// properly handle translation failures
func TestConcurrentTranslationFailureHandling(t *testing.T) {
	// Create a failing mock translator for testing error conditions
	errorTranslator := failingMockTranslator(0.5) // 50% chance of failure
	
	// Test data
	baseContent := files.LanguageContent{
		"greeting": "Hello",
		"farewell": "Goodbye",
		"message": "Important message",
	}
	
	// Empty previous translation
	emptyPrevious := files.LanguageContent{}
	
	// Test SimpleProcessor error handling
	t.Run("SimpleProcessorErrorHandling", func(t *testing.T) {
		proc := processor.NewSimpleProcessor(errorTranslator)
		proc.SetWorkerPoolSize(4) // Use multiple workers to test concurrent error handling
		
		// Execute with concurrency - should not panic even with failures
		result, _ := proc.Execute(baseContent, "en", "fr", emptyPrevious)
		// We're expecting some keys to fail but the overall process should complete
		
		// We should still get a result with some values, even if some translations failed
		assert.NotNil(t, result, "Should get a result even with translation failures")
		
		// Verify all keys are present in the result
		for key := range baseContent {
			_, exists := result[key]
			assert.True(t, exists, "Result should contain all original keys")
		}
	})
	
	// Test ASTProcessor error handling if available
	astProc, err := processor.CreateProcessor("ast-json", errorTranslator) 
	if err != nil {
		t.Skip("ASTProcessor not available:", err)
	}
	
	// Cast to ASTProcessor if possible to set worker count
	if ap, ok := astProc.(*processor.ASTProcessor); ok {
		ap.SetWorkerPoolSize(4) // Use multiple workers
		
		t.Run("ASTProcessorErrorHandling", func(t *testing.T) {
			// Create test data with ICU message formats
			astContent := files.LanguageContent{
				"simple": "Hello, world!",
				"withPlaceholder": "Hello, {name}!",
			}
			
			// Execute with concurrency - should not panic even with failures
			result, _ := astProc.Execute(astContent, "en", "fr", emptyPrevious)
			// We're expecting some keys to fail but the overall process should complete
			
			// We should still get a result with some values, even if some translations failed
			assert.NotNil(t, result, "Should get a result even with translation failures")
			
			// Verify all keys are present in the result
			for key := range astContent {
				_, exists := result[key]
				assert.True(t, exists, "Result should contain all original keys")
			}
		})
	}
}

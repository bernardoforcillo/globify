package processor_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bernardoforcillo/globify/internal/files"
	"github.com/bernardoforcillo/globify/internal/processor"
)

// MockTranslator implements the translator.Translator interface for testing
type MockTranslator struct {
	// MockTranslate defines the function to be called instead of the real Translate
	MockTranslate func(text, from, to string) (string, error)
}

// Translate calls the mock function
func (m *MockTranslator) Translate(text, from, to string) (string, error) {
	return m.MockTranslate(text, from, to)
}

// Helper function to create a simple mock translator
func createMockTranslator() *MockTranslator {
	return &MockTranslator{
		MockTranslate: func(text, from, to string) (string, error) {
			// Simple mock that adds the target language as a prefix
			return fmt.Sprintf("[%s] %s", to, text), nil
		},
	}
}

// Helper function to create a mock translator that returns errors
func createErrorMockTranslator() *MockTranslator {
	return &MockTranslator{
		MockTranslate: func(text, from, to string) (string, error) {
			return "", fmt.Errorf("mock translation error")
		},
	}
}

// Helper function to compare maps
func compareMaps(t *testing.T, got, want files.LanguageContent) bool {
	if len(got) != len(want) {
		t.Errorf("Map sizes differ: got %d entries, want %d entries", len(got), len(want))
		return false
	}
	
	for key, wantVal := range want {
		gotVal, exists := got[key]
		if !exists {
			t.Errorf("Key %q missing in result", key)
			return false
		}
		
		// Handle different types appropriately
		switch wantV := wantVal.(type) {
		case map[string]interface{}:
			// Check if gotVal is a map (either type)
			var gotMap files.LanguageContent
			
			// Try to convert gotVal to the appropriate type for comparison
			switch gotV := gotVal.(type) {
			case map[string]interface{}:
				gotMap = files.LanguageContent(gotV)
			case files.LanguageContent:
				gotMap = gotV
			default:
				t.Errorf("For key %q: got type %T, want map type", key, gotVal)
				return false
			}
			
			// Convert want to LanguageContent for comparison
			wantMap := files.LanguageContent(wantV)
			
			if !compareMaps(t, gotMap, wantMap) {
				return false
			}
		default:
			// For non-map values, use DeepEqual
			if !reflect.DeepEqual(gotVal, wantVal) {
				t.Errorf("For key %q: got %v (%T), want %v (%T)", 
					key, gotVal, gotVal, wantVal, wantVal)
				return false
			}
		}
	}
	
	return true
}

func TestSimpleProcessorExecute(t *testing.T) {
	// Create a mock translator
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
	
	// Previous translations
	previousTranslation := files.LanguageContent{
		"greeting": "[fr] Hello", // This would be reused
		"farewell": "Different value", // This would be retranslated
		"nested": map[string]interface{}{
			"key1": "[fr] Nested value 1", // This would be reused
		},
	}
	
	// Execute the processor
	result, err := proc.Execute(baseContent, "en", "fr", previousTranslation)
	
	// Verify no errors
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	
	// Check content has been correctly translated
	expected := files.LanguageContent{
		"greeting": "[fr] Hello", // Reused from previous
		"farewell": "[fr] Goodbye", // Newly translated
		"nested": map[string]interface{}{
			"key1": "[fr] Nested value 1", // Reused from previous
			"key2": "[fr] Nested value 2", // Newly translated
		},
		"number": 42, // Non-string values unchanged
		"boolean": true, // Non-string values unchanged
	}
	
	// Use our custom map comparison function
	if !compareMaps(t, result, expected) {
		t.Errorf("Maps are not equal")
	}
}

func TestSimpleProcessorExecuteWithErrors(t *testing.T) {
	// Create a mock translator that returns errors
	errorTranslator := createErrorMockTranslator()
	
	// Create a SimpleProcessor with the error mock
	proc := processor.NewSimpleProcessor(errorTranslator)
	
	// Test data
	baseContent := files.LanguageContent{
		"greeting": "Hello",
		"nested": map[string]interface{}{
			"key1": "Nested value",
		},
	}
	
	// Execute the processor
	result, err := proc.Execute(baseContent, "en", "fr", files.LanguageContent{})
	
	// Verify no errors at the top level (errors are logged but not returned)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	
	// Check that original values are preserved on error
	expected := files.LanguageContent{
		"greeting": "Hello", // Original preserved on error
		"nested": map[string]interface{}{
			"key1": "Nested value", // Original preserved on error
		},
	}
	
	// Use our custom map comparison function
	if !compareMaps(t, result, expected) {
		t.Errorf("Maps are not equal")
	}
}

func TestCreateProcessor(t *testing.T) {
	mockTranslator := createMockTranslator()
	
	tests := []struct {
		name            string
		translationType string
		wantErr         bool
	}{
		{
			name:            "Simple JSON",
			translationType: "simple-json",
			wantErr:         false,
		},
		{
			name:            "AST JSON",
			translationType: "ast-json",
			wantErr:         false,
		},
		{
			name:            "Invalid type",
			translationType: "invalid",
			wantErr:         true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processor.CreateProcessor(tt.translationType, mockTranslator)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateProcessor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
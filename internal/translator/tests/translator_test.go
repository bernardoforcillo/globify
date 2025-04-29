package translator_test

import (
	"os"
	"testing"

	"github.com/bernardoforcillo/globify/internal/translator"
)

func TestCreateTranslator(t *testing.T) {
	// Skip if no API key is set since this would cause the test to fail
	apiKey := os.Getenv("DEEPL_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: DEEPL_API_KEY environment variable not set")
	}
	
	// Test creating a translator
	tr, err := translator.CreateTranslator()
	if err != nil {
		t.Errorf("CreateTranslator() error = %v", err)
	}
	if tr == nil {
		t.Errorf("CreateTranslator() returned nil translator")
	}
}

// This is a basic test that ensures the DeepL translator implements the interface
// For a more comprehensive test, we would need to mock the DeepL API
func TestDeeplTranslator(t *testing.T) {
	// Skip if no API key is set
	apiKey := os.Getenv("DEEPL_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: DEEPL_API_KEY environment variable not set")
	}
	
	tr, err := translator.NewDeeplTranslator()
	if err != nil {
		t.Errorf("NewDeeplTranslator() error = %v", err)
	}
	
	// Test the Translate method
	result, err := tr.Translate("Hello, world!", "en", "fr")
	if err != nil {
		t.Errorf("Translate() error = %v", err)
	}
	
	// The result should not be empty
	if result == "" {
		t.Errorf("Translate() returned empty string")
	}
}
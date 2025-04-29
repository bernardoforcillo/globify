package app_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/bernardoforcillo/globify/internal/app"
	"github.com/bernardoforcillo/globify/internal/files"
)

// TestApp performs integration testing of the app functionality
// This test requires mocking the configuration system to avoid
// calling actual translation services
func TestApp(t *testing.T) {
	// Skip if no API key is set
	apiKey := os.Getenv("DEEPL_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping test: DEEPL_API_KEY environment variable not set")
	}
	
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "globify-app-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a translations directory
	translationsDir := filepath.Join(tempDir, "translations")
	err = os.Mkdir(translationsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create translations dir: %v", err)
	}

	// Create sample English content
	enContent := files.LanguageContent{
		"greeting": "Hello",
		"farewell": "Goodbye",
		"nested": map[string]interface{}{
			"key1": "Nested value 1",
			"key2": "Nested value 2",
		},
	}

	// Write English content to file
	enFilePath := filepath.Join(translationsDir, "en.json")
	enData, err := json.MarshalIndent(enContent, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal English content: %v", err)
	}
	err = os.WriteFile(enFilePath, enData, 0644)
	if err != nil {
		t.Fatalf("Failed to write English file: %v", err)
	}

	// Create a config file
	configContent := `{
		"translationType": "simple-json",
		"fileExtension": "json",
		"baseLanguage": "en",
		"languages": ["en", "fr", "es"],
		"folder": "` + translationsDir + `"
	}`
	configPath := filepath.Join(tempDir, "globify.config.json")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Save the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(cwd)

	// Change to the temp directory so that LoadConfig() finds our config file
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create mocks for testing
	// Note: This is a limited test that doesn't actually call translation services
	// In a real scenario, we would need to mock more dependencies

	// We could extend this test to use mocked translators,
	// but for simplicity we'll just verify app initialization
	// and leave the full integration to manual testing
	_, err = app.NewApp()
	if err != nil {
		t.Errorf("NewApp() error = %v", err)
	}

	// For a more complete test, we would need to:
	// 1. Configure mock services
	// 2. Override the app's dependencies with mocks
	// 3. Call app.Run() to execute the translation process
	// 4. Verify the output files
}

// TestAppWithCustomDependencies tests the app with custom dependencies
// This is a more advanced test that would allow testing without external services
func TestAppWithCustomDependencies(t *testing.T) {
	// Skip this test for now - it would require refactoring App to accept dependencies
	t.Skip("Skipping test that requires refactoring App to accept dependencies")

	// Future enhancement: Add a constructor that allows injecting dependencies:
	//
	// func NewAppWithDependencies(
	//     cfg *config.Config,
	//     trans translator.Translator,
	//     fm files.FileManager,
	//     proc processor.ObjectProcessor,
	// ) *App {
	//     return &App{
	//         config:      cfg,
	//         translator:  trans,
	//         fileManager: fm,
	//         processor:   proc,
	//     }
	// }
	//
	// This would allow complete mocking of all dependencies for testing
}
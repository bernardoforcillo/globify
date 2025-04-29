package config_test

import (
	"os"
	"testing"

	"github.com/bernardoforcillo/globify/internal/config"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  config.Config
		wantErr bool
	}{
		{
			name: "Valid config",
			config: config.Config{
				TranslationType: "simple-json",
				FileExtension:   "json",
				BaseLanguage:    "en",
				Languages:       []string{"es", "fr", "de"},
				Folder:          "translations",
			},
			wantErr: false,
		},
		{
			name: "Invalid translation type",
			config: config.Config{
				TranslationType: "invalid",
				FileExtension:   "json",
				BaseLanguage:    "en",
				Languages:       []string{"es", "fr", "de"},
				Folder:          "translations",
			},
			wantErr: true,
		},
		{
			name: "Invalid file extension",
			config: config.Config{
				TranslationType: "simple-json",
				FileExtension:   "yaml",
				BaseLanguage:    "en",
				Languages:       []string{"es", "fr", "de"},
				Folder:          "translations",
			},
			wantErr: true,
		},
		{
			name: "Invalid base language",
			config: config.Config{
				TranslationType: "simple-json",
				FileExtension:   "json",
				BaseLanguage:    "invalid",
				Languages:       []string{"es", "fr", "de"},
				Folder:          "translations",
			},
			wantErr: true,
		},
		{
			name: "Invalid target language",
			config: config.Config{
				TranslationType: "simple-json",
				FileExtension:   "json",
				BaseLanguage:    "en",
				Languages:       []string{"es", "invalid", "de"},
				Folder:          "translations",
			},
			wantErr: true,
		},
		{
			name: "Empty folder",
			config: config.Config{
				TranslationType: "simple-json",
				FileExtension:   "json",
				BaseLanguage:    "en",
				Languages:       []string{"es", "fr", "de"},
				Folder:          "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "globify-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to the temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Test cases
	tests := []struct {
		name       string
		configData string
		wantErr    bool
	}{
		{
			name: "Valid config",
			configData: `{
				"translationType": "simple-json",
				"fileExtension": "json",
				"baseLanguage": "en",
				"languages": ["fr", "es", "de"],
				"folder": "translations"
			}`,
			wantErr: false,
		},
		{
			name: "Invalid JSON",
			configData: `{
				"translationType": "simple-json",
				"fileExtension": "json",
				"baseLanguage": "en",
				"languages": ["fr", "es", "de"],
				"folder": "translations"
			`,
			wantErr: true,
		},
		{
			name: "Invalid config values",
			configData: `{
				"translationType": "invalid",
				"fileExtension": "json",
				"baseLanguage": "en",
				"languages": ["fr", "es", "de"],
				"folder": "translations"
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh temp dir for each test
			testDir, err := os.MkdirTemp(tempDir, "config-test")
			if err != nil {
				t.Fatalf("Failed to create test dir: %v", err)
			}
			defer os.RemoveAll(testDir)
			os.Chdir(testDir)

			// Write the config file
			err = os.WriteFile("globify.config.json", []byte(tt.configData), 0644)
			if err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			// Check if loading the config produces the expected result
			_, err = config.LoadConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
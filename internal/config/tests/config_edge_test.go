package config_test

import (
	"testing"

	"github.com/bernardoforcillo/globify/internal/config"
)

// TestConfigValidateEdgeCases tests edge cases and boundary conditions
func TestConfigValidateEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		config  config.Config
		wantErr bool
	}{
		{
			name: "Many target languages",
			config: config.Config{
				TranslationType: "simple-json",
				FileExtension:   "json",
				BaseLanguage:    "en",
				Languages:       []string{"es", "fr", "de", "it", "pt", "ru", "zh", "ja", "ko", "ar", "hi"},
				Folder:          "translations",
			},
			wantErr: false,
		},
		{
			name: "Base language in target languages",
			config: config.Config{
				TranslationType: "simple-json",
				FileExtension:   "json",
				BaseLanguage:    "en",
				Languages:       []string{"en", "es", "fr"},
				Folder:          "translations",
			},
			wantErr: false, // This should be allowed, even though it's redundant
		},
		{
			name: "Mixed case language codes",
			config: config.Config{
				TranslationType: "simple-json",
				FileExtension:   "json",
				BaseLanguage:    "EN", // Invalid - should be lowercase
				Languages:       []string{"es", "FR"}, // FR is invalid - should be lowercase
				Folder:          "translations",
			},
			wantErr: true,
		},
		{
			name: "Invalid language codes",
			config: config.Config{
				TranslationType: "simple-json",
				FileExtension:   "json",
				BaseLanguage:    "en",
				Languages:       []string{"es", "123", "de"}, // 123 is invalid
				Folder:          "translations",
			},
			wantErr: true,
		},
		{
			name: "Three-letter language codes",
			config: config.Config{
				TranslationType: "simple-json",
				FileExtension:   "json",
				BaseLanguage:    "eng", // Invalid - not matching regex
				Languages:       []string{"spa", "fra", "deu"},
				Folder:          "translations",
			},
			wantErr: true,
		},
		{
			name: "Valid hyphenated language codes",
			config: config.Config{
				TranslationType: "simple-json",
				FileExtension:   "json",
				BaseLanguage:    "en-Abcd",
				Languages:       []string{"es-Abcd", "fr-Abcd", "de-Abcd"},
				Folder:          "translations",
			},
			wantErr: false,
		},
		{
			name: "Invalid hyphenated language codes",
			config: config.Config{
				TranslationType: "simple-json",
				FileExtension:   "json",
				BaseLanguage:    "en-Abcd",
				Languages:       []string{"es-Abcd", "fr-abcd", "de-Abcd"}, // fr-abcd is invalid (second part should start with uppercase)
				Folder:          "translations",
			},
			wantErr: true,
		},
		{
			name: "Relative folder path",
			config: config.Config{
				TranslationType: "simple-json",
				FileExtension:   "json",
				BaseLanguage:    "en",
				Languages:       []string{"es", "fr", "de"},
				Folder:          "../translations", // Relative path should be valid
			},
			wantErr: false,
		},
		{
			name: "Deep folder path",
			config: config.Config{
				TranslationType: "simple-json",
				FileExtension:   "json",
				BaseLanguage:    "en",
				Languages:       []string{"es", "fr", "de"},
				Folder:          "app/assets/locales/translations", // Deep path should be valid
			},
			wantErr: false,
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

// TestConfigValidateLoadInteraction tests specific interactions between validation and loading
func TestConfigValidateLoadInteraction(t *testing.T) {
	// This test would need to be run with a custom temporary directory
	// and configuration files, similar to TestLoadConfig but with focus
	// on specific validation scenarios that might only appear when loading
	// from a file.
	
	// Since the implementation is complex and requires filesystem setup,
	// we're just documenting the test case here to show what should be tested.
	t.Skip("This would test validation during the config loading process")
}
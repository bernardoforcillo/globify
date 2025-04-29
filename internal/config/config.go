package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Config represents the application configuration
type Config struct {
	TranslationType string   `json:"translationType"`
	FileExtension   string   `json:"fileExtension"`
	BaseLanguage    string   `json:"baseLanguage"`
	Languages       []string `json:"languages"`
	Folder          string   `json:"folder"`
}

// Language code regex pattern
var langRegex = regexp.MustCompile(`^[a-z]{2}(-[A-Z][a-z]{3})?$`)

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Check translation type
	if c.TranslationType != "simple-json" && c.TranslationType != "ast-json" {
		return fmt.Errorf("translationType must be 'simple-json' or 'ast-json'")
	}

	// Check file extension
	if c.FileExtension != "json" {
		return fmt.Errorf("fileExtension must be 'json'")
	}

	// Check base language
	if !langRegex.MatchString(c.BaseLanguage) {
		return fmt.Errorf("baseLanguage '%s' must be in format like 'en' or 'en-US'", c.BaseLanguage)
	}

	// Check target languages
	for _, lang := range c.Languages {
		if !langRegex.MatchString(lang) {
			return fmt.Errorf("language '%s' must be in format like 'en' or 'en-US'", lang)
		}
	}

	// Check folder
	if c.Folder == "" {
		return fmt.Errorf("folder cannot be empty")
	}

	return nil
}

// LoadConfig loads the configuration from a file
func LoadConfig() (*Config, error) {
	configFiles := []string{"globify.config.json"}
	
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Try to find a config file
	var configFile string
	for _, file := range configFiles {
		path := filepath.Join(cwd, file)
		if _, err := os.Stat(path); err == nil {
			configFile = path
			break
		}
	}

	if configFile == "" {
		return nil, fmt.Errorf("no config file found (tried: %v)", configFiles)
	}

	// Read the config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	// Parse the config file
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
	}

	// Validate the config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}
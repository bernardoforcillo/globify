package files

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// JSONManager implements FileManager for JSON files
type JSONManager struct{}

// NewJSONManager creates a new JSONManager
func NewJSONManager() *JSONManager {
	return &JSONManager{}
}

// Write saves the content to a JSON file
func (m *JSONManager) Write(filePath string, content LanguageContent) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}


	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(content); err != nil {
		return fmt.Errorf("failed to marshal JSON content for %s: %w", filePath, err)
	}

	// Write to file
	err := os.WriteFile(filePath, buffer.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// Read loads content from a JSON file
func (m *JSONManager) Read(filePath string) (LanguageContent, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var content LanguageContent
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON from %s: %w", filePath, err)
	}

	return content, nil
}

// Exists checks if a file exists
func (m *JSONManager) Exists(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil // File doesn't exist
		}
		return false, fmt.Errorf("failed to check if file %s exists: %w", filePath, err)
	}
	
	return !info.IsDir(), nil // Return true if it exists and is not a directory
}
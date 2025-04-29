package files_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bernardoforcillo/globify/internal/files"
)

func TestJSONManager(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "jsonmanager-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create JSONManager instance
	manager := files.NewJSONManager()

	// Test data
	testContent := files.LanguageContent{
		"greeting": "Hello",
		"farewell": "Goodbye",
		"nested": map[string]interface{}{
			"key1": "Nested value 1",
			"key2": "Nested value 2",
		},
	}

	// Test file path
	testFilePath := filepath.Join(tempDir, "test.json")

	// Test Exists - should not exist yet
	exists, err := manager.Exists(testFilePath)
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if exists {
		t.Errorf("Exists() = %v, want %v", exists, false)
	}

	// Test Write
	err = manager.Write(testFilePath, testContent)
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}

	// Test Exists - should exist now
	exists, err = manager.Exists(testFilePath)
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Errorf("Exists() = %v, want %v", exists, true)
	}

	// Test Read
	readContent, err := manager.Read(testFilePath)
	if err != nil {
		t.Errorf("Read() error = %v", err)
	}
	
	// Verify content matches
	if !reflect.DeepEqual(readContent, testContent) {
		t.Errorf("Read() got = %v, want %v", readContent, testContent)
	}

	// Test nested directory creation
	nestedFilePath := filepath.Join(tempDir, "nested", "path", "test.json")
	err = manager.Write(nestedFilePath, testContent)
	if err != nil {
		t.Errorf("Write() with nested path error = %v", err)
	}

	// Verify nested directory was created
	exists, err = manager.Exists(nestedFilePath)
	if err != nil {
		t.Errorf("Exists() error = %v", err)
	}
	if !exists {
		t.Errorf("Exists() = %v, want %v", exists, true)
	}

	// Test Read with invalid content
	invalidFilePath := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidFilePath, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid file: %v", err)
	}

	_, err = manager.Read(invalidFilePath)
	if err == nil {
		t.Errorf("Read() with invalid content should return error")
	}

	// Test Read with non-existent file
	_, err = manager.Read(filepath.Join(tempDir, "nonexistent.json"))
	if err == nil {
		t.Errorf("Read() with non-existent file should return error")
	}
}

func TestNewFileManager(t *testing.T) {
	tests := []struct {
		name     string
		fileType string
		wantErr  bool
	}{
		{
			name:     "JSON manager",
			fileType: "json",
			wantErr:  false,
		},
		{
			name:     "Unsupported file type",
			fileType: "yaml",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := files.NewFileManager(tt.fileType)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFileManager() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
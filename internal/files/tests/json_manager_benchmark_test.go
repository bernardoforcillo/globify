package files_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/bernardoforcillo/globify/internal/files"
)

// Benchmarks for file operations

func BenchmarkJSONManagerWrite(b *testing.B) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "jsonmanager-bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create JSONManager instance
	manager := files.NewJSONManager()

	// Test data with varying complexity
	testContent := files.LanguageContent{
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use a unique file for each iteration to avoid disk caching effects
		uniquePath := filepath.Join(tempDir, filepath.Base(tempDir)+"-"+fmt.Sprintf("%d", i)+".json")
		_ = manager.Write(uniquePath, testContent)
	}
}

func BenchmarkJSONManagerRead(b *testing.B) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "jsonmanager-bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
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
		"number": 42,
		"boolean": true,
	}

	// Write test file once
	testFilePath := filepath.Join(tempDir, "test.json")
	data, err := json.MarshalIndent(testContent, "", "  ")
	if err != nil {
		b.Fatalf("Failed to marshal JSON: %v", err)
	}
	err = os.WriteFile(testFilePath, data, 0644)
	if err != nil {
		b.Fatalf("Failed to write file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.Read(testFilePath)
	}
}

func BenchmarkJSONManagerReadLarge(b *testing.B) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "jsonmanager-bench")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create JSONManager instance
	manager := files.NewJSONManager()

	// Create a large test file with many entries
	largeContent := make(files.LanguageContent)
	for i := 0; i < 1000; i++ {
		key := "key_" + string(rune(i))
		largeContent[key] = "This is translation " + string(rune(i))
		
		// Add some nested content every 10 entries
		if i%10 == 0 {
			nested := make(map[string]interface{})
			for j := 0; j < 5; j++ {
				nested["nested_"+string(rune(j))] = "Nested value " + string(rune(j))
			}
			largeContent["nested_"+string(rune(i))] = nested
		}
	}

	// Write large test file
	largeFilePath := filepath.Join(tempDir, "large.json")
	data, err := json.MarshalIndent(largeContent, "", "  ")
	if err != nil {
		b.Fatalf("Failed to marshal JSON: %v", err)
	}
	err = os.WriteFile(largeFilePath, data, 0644)
	if err != nil {
		b.Fatalf("Failed to write file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.Read(largeFilePath)
	}
}
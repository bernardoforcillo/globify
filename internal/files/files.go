package files

import "fmt"

// LanguageContent represents the structure of language files
type LanguageContent map[string]interface{}

// FileManager defines the interface for file operations
type FileManager interface {
	Write(filePath string, content LanguageContent) error
	Read(filePath string) (LanguageContent, error)
	Exists(filePath string) (bool, error)
}

// Factory function to get a file manager
func NewFileManager(fileType string) (FileManager, error) {
	switch fileType {
	case "json":
		return NewJSONManager(), nil
	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}
}
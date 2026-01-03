package processor

import (
	"fmt"
	"log"

	"github.com/bernardoforcillo/globify/internal/files"
	"github.com/bernardoforcillo/globify/internal/translator"
)

// SimpleProcessor provides a basic implementation that translates strings in objects
type SimpleProcessor struct {
	translator translator.Translator
	// Add a worker pool size to control concurrency
	workerPoolSize int
}

// NewSimpleProcessor creates a new SimpleProcessor
func NewSimpleProcessor(translator translator.Translator) *SimpleProcessor {
	// Set workerPoolSize to 1 to disable concurrency
	return &SimpleProcessor{
		translator:    translator,
		workerPoolSize: 1,
	}
}

// SetWorkerPoolSize allows dynamically configuring the number of worker goroutines
// This is now maintained for backward compatibility but will always set to 1
func (p *SimpleProcessor) SetWorkerPoolSize(count int) {
	// Force to 1 to disable concurrency
	p.workerPoolSize = 1
}

// Execute translates all string values in the content recursively
func (p *SimpleProcessor) Execute(
	obj files.LanguageContent,
	from, target string,
	previousTranslation files.LanguageContent,
) (files.LanguageContent, error) {
	result := make(files.LanguageContent)
	
	// Process each item sequentially instead of using goroutines
	for key, value := range obj {
		// Skip keys starting with @ (metadata in ARB)
		if len(key) > 0 && key[0] == '@' {
			result[key] = value
			continue
		}

		prevValue, hasPrevious := previousTranslation[key]
		
		switch v := value.(type) {
		case string:
			// If previous translation exists and matches, use it
			if hasPrevious && prevValue == value {
				result[key] = prevValue
				continue
			}
			
			// Translate the string
			translated, err := p.translator.Translate(v, from, target)
			if err != nil {
				log.Printf("Warning: Failed to translate key '%s': %v", key, err)
				result[key] = v // Keep original in case of error
				continue
			}
			
			result[key] = translated
			
		case map[string]interface{}:
			// Handle nested objects
			var prevMap files.LanguageContent
			if hasPrevious {
				if nested, ok := prevValue.(map[string]interface{}); ok {
					prevMap = nested
				} else {
					prevMap = make(files.LanguageContent)
				}
			} else {
				prevMap = make(files.LanguageContent)
			}
			
			// Recursively translate the nested object
			nestedResult, err := p.Execute(v, from, target, prevMap)
			if err != nil {
				return nil, fmt.Errorf("failed to translate nested object at key '%s': %w", key, err)
			}
			
			result[key] = nestedResult
			
		default:
			// Keep non-string, non-object values as they are
			result[key] = value
		}
	}
	
	return result, nil
}
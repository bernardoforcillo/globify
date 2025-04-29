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
}

// NewSimpleProcessor creates a new SimpleProcessor
func NewSimpleProcessor(translator translator.Translator) *SimpleProcessor {
	return &SimpleProcessor{
		translator: translator,
	}
}

// Execute translates all string values in the content recursively
func (p *SimpleProcessor) Execute(
	obj files.LanguageContent,
	from, target string,
	previousTranslation files.LanguageContent,
) (files.LanguageContent, error) {
	result := make(files.LanguageContent)

	for key, value := range obj {
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
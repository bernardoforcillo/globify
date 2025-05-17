package processor

import (
	"fmt"
	"log"
	"sort"

	"github.com/bernardoforcillo/globify/internal/files"
	"github.com/bernardoforcillo/globify/internal/icu"
	"github.com/bernardoforcillo/globify/internal/translator"
)

// ASTProcessor handles translation of ICU message format strings
type ASTProcessor struct {
	translator translator.Translator
	// Add a worker pool size to control concurrency
	workerPoolSize int
}

// NewASTProcessor creates a new ASTProcessor
func NewASTProcessor(translator translator.Translator) *ASTProcessor {
	// Set workerPoolSize to 1 to disable concurrency
	return &ASTProcessor{
		translator:     translator,
		workerPoolSize: 1,
	}
}

// SetWorkerPoolSize allows dynamically configuring the number of worker goroutines
// This is now maintained for backward compatibility but will always set to 1
func (p *ASTProcessor) SetWorkerPoolSize(count int) {
	// Force to 1 to disable concurrency
	p.workerPoolSize = 1
}

// Execute translates content with ICU message format strings
func (p *ASTProcessor) Execute(
	obj files.LanguageContent,
	from, target string,
	previousTranslation files.LanguageContent,
) (files.LanguageContent, error) {
	result := make(files.LanguageContent)
	
	// Process each item sequentially instead of using goroutines
	for key, value := range obj {
		prevValue, hasPrevious := previousTranslation[key]
		
		switch v := value.(type) {
		case string:
			// If previous translation exists and matches, use it
			if hasPrevious && prevValue == value {
				result[key] = prevValue
				continue
			}
			
			// Parse the message string into AST
			ast, err := icu.Parse(v)
			if err != nil {
				log.Printf("Warning: Failed to parse ICU message for key '%s': %v", key, err)
				
				// Fall back to simple translation
				translated, err := p.translator.Translate(v, from, target)
				if err != nil {
					log.Printf("Warning: Failed to translate key '%s': %v", key, err)
					result[key] = v // Keep original in case of error
					continue
				}
				result[key] = translated
				continue
			}
			
			// Translate the AST
			translatedMessage, err := p.translateElements(ast, from, target)
			if err != nil {
				log.Printf("Warning: Failed to translate AST for key '%s': %v", key, err)
				result[key] = v // Keep original in case of error
				continue
			}
			result[key] = translatedMessage
			
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

// translateElements translates a slice of ICU elements
func (p *ASTProcessor) translateElements(elements []icu.Element, from, target string) (string, error) {
	var result string
	
	// Always use sequential processing to avoid too many requests
	for _, element := range elements {
		translated, err := p.translateElement(element, from, target)
		if err != nil {
			return "", err
		}
		result += translated
	}
	
	return result, nil
}

// translateElement translates a single ICU element
func (p *ASTProcessor) translateElement(element icu.Element, from, target string) (string, error) {
	switch element.Type() {
	case icu.Literal:
		// Only translate literal text elements
		lit := element.(icu.LiteralElement)
		if lit.Value == "" {
			return "", nil
		}
		
		translated, err := p.translator.Translate(lit.Value, from, target)
		if err != nil {
			return "", fmt.Errorf("failed to translate literal: %w", err)
		}
		return translated, nil
		
	case icu.Tag:
		// Handle tag elements by translating their children
		tag := element.(icu.TagElement)
		translatedContent, err := p.translateElements(tag.Children, from, target)
		if err != nil {
			return "", fmt.Errorf("failed to translate tag content: %w", err)
		}
		return fmt.Sprintf("<%s>%s</%s>", tag.Value, translatedContent, tag.Value), nil
		
	case icu.Select:
		// Handle select elements by translating each option sequentially
		sel := element.(icu.SelectElement)
		translatedOptions := make(map[string]string)
		
		// Process each option sequentially
		for key, option := range sel.Options {
			translatedOption, err := p.translateElements(option, from, target)
			if err != nil {
				return "", fmt.Errorf("failed to translate select option: %w", err)
			}
			translatedOptions[key] = translatedOption
		}
		
		// Reconstruct the select format with sorted keys for consistent output
		selectStr := fmt.Sprintf("{%s, select, ", sel.Value)
		// Sort the keys for consistent output
		var keys []string
		for key := range translatedOptions {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			selectStr += fmt.Sprintf("%s {%s} ", key, translatedOptions[key])
		}
		selectStr += "}"
		return selectStr, nil
		
	case icu.Plural:
		// Handle plural elements by translating each option sequentially
		plural := element.(icu.PluralElement)
		translatedOptions := make(map[string]string)
		
		// Process each option sequentially
		for key, option := range plural.Options {
			translatedOption, err := p.translateElements(option, from, target)
			if err != nil {
				return "", fmt.Errorf("failed to translate plural option: %w", err)
			}
			translatedOptions[key] = translatedOption
		}
		
		// Reconstruct the plural format with specific order for keys
		pluralStr := fmt.Sprintf("{%s, plural, ", plural.Value)
		
		// Define the order of plural forms ('one' should come before 'other')
		// This ensures consistent output matching test expectations
		pluralForms := []string{"zero", "one", "two", "few", "many", "other"}
		
		for _, form := range pluralForms {
			if value, ok := translatedOptions[form]; ok {
				pluralStr += fmt.Sprintf("%s {%s} ", form, value)
			}
		}
		
		// Add any remaining forms that weren't in our predefined order
		var remainingKeys []string
		for key := range translatedOptions {
			// Check if this key is one of our predefined forms
			isPredefined := false
			for _, form := range pluralForms {
				if key == form {
					isPredefined = true
					break
				}
			}
			
			if !isPredefined {
				remainingKeys = append(remainingKeys, key)
			}
		}
		
		// Sort the remaining keys for consistent output
		if len(remainingKeys) > 0 {
			sort.Strings(remainingKeys)
			for _, key := range remainingKeys {
				pluralStr += fmt.Sprintf("%s {%s} ", key, translatedOptions[key])
			}
		}
		
		pluralStr += "}"
		return pluralStr, nil
		
	default:
		// Keep other elements as they are
		return element.String(), nil
	}
}
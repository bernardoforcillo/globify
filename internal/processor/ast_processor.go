package processor

import (
	"fmt"
	"log"
	"sort"
	"sync"

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
	return &ASTProcessor{
		translator:     translator,
		workerPoolSize: 1,
	}
}

// SetWorkerPoolSize allows dynamically configuring the number of worker goroutines
func (p *ASTProcessor) SetWorkerPoolSize(count int) {
	if count < 1 {
		count = 1
	}
	p.workerPoolSize = count
}

// Execute translates content with ICU message format strings
func (p *ASTProcessor) Execute(
	obj files.LanguageContent,
	from, target string,
	previousTranslation files.LanguageContent,
) (files.LanguageContent, error) {
	// Create a semaphore to limit concurrency
	sem := make(chan struct{}, p.workerPoolSize)
	return p.executeInternal(obj, from, target, previousTranslation, sem)
}

func (p *ASTProcessor) executeInternal(
	obj files.LanguageContent,
	from, target string,
	previousTranslation files.LanguageContent,
	sem chan struct{},
) (files.LanguageContent, error) {
	result := make(files.LanguageContent)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Error channel to collect errors from goroutines
	errChan := make(chan error, len(obj))

	// Process each item
	for key, value := range obj {
		// Skip keys starting with @ (metadata in ARB)
		if len(key) > 0 && key[0] == '@' {
			mu.Lock()
			result[key] = value
			mu.Unlock()
			continue
		}

		prevValue, hasPrevious := previousTranslation[key]
		
		switch v := value.(type) {
		case string:
			// If previous translation exists and matches, use it
			if hasPrevious && prevValue == value {
				mu.Lock()
				result[key] = prevValue
				mu.Unlock()
				continue
			}

			wg.Add(1)
			go func(k, val string) {
				defer wg.Done()
				
				// Acquire semaphore
				sem <- struct{}{}
				defer func() { <-sem }()

				// Parse the message string into AST
				ast, err := icu.Parse(val)
				if err != nil {
					log.Printf("Warning: Failed to parse ICU message for key '%s': %v", k, err)

					// Fall back to simple translation
					translated, err := p.translator.Translate(val, from, target)
					if err != nil {
						log.Printf("Warning: Failed to translate key '%s': %v", k, err)
						mu.Lock()
						result[k] = val // Keep original in case of error
						mu.Unlock()
						return
					}
					mu.Lock()
					result[k] = translated
					mu.Unlock()
					return
				}

				// Translate the AST
				translatedMessage, err := p.translateElements(ast, from, target)
				if err != nil {
					log.Printf("Warning: Failed to translate AST for key '%s': %v", k, err)
					mu.Lock()
					result[k] = val // Keep original in case of error
					mu.Unlock()
					return
				}
				mu.Lock()
				result[k] = translatedMessage
				mu.Unlock()
			}(key, v)
			
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
			nestedResult, err := p.executeInternal(v, from, target, prevMap, sem)
			if err != nil {
				errChan <- fmt.Errorf("failed to translate nested object at key '%s': %w", key, err)
				continue
			}

			mu.Lock()
			result[key] = nestedResult
			mu.Unlock()
			
		default:
			// Keep non-string, non-object values as they are
			mu.Lock()
			result[key] = value
			mu.Unlock()
		}
	}

	wg.Wait()
	close(errChan)

	// Check if any error occurred
	for err := range errChan {
		if err != nil {
			return nil, err
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
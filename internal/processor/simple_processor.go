package processor

import (
	"fmt"
	"log"
	"sync"

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
	return &SimpleProcessor{
		translator:     translator,
		workerPoolSize: 1,
	}
}

// SetWorkerPoolSize allows dynamically configuring the number of worker goroutines
func (p *SimpleProcessor) SetWorkerPoolSize(count int) {
	if count < 1 {
		count = 1
	}
	p.workerPoolSize = count
}

// Execute translates all string values in the content recursively
func (p *SimpleProcessor) Execute(
	obj files.LanguageContent,
	from, target string,
	previousTranslation files.LanguageContent,
) (files.LanguageContent, error) {
	// Create a semaphore to limit concurrency
	sem := make(chan struct{}, p.workerPoolSize)
	return p.executeInternal(obj, from, target, previousTranslation, sem)
}

func (p *SimpleProcessor) executeInternal(
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

				// Translate the string
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
			// Note: We don't launch a goroutine for the nested object itself,
			// but pass the shared semaphore down so its children can run concurrently
			// respecting the global limit.
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
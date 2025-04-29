package processor_test

import (
	"fmt"

	"github.com/bernardoforcillo/globify/internal/files"
)

// createLargeDataset creates a large dataset for benchmarking concurrency improvements
// size: number of top-level keys
// depth: nesting depth for nested objects
func createLargeDataset(size, depth int) files.LanguageContent {
	data := make(files.LanguageContent, size)
	
	// Create a mix of string and nested objects
	for i := 0; i < size; i++ {
		key := fmt.Sprintf("key%d", i)
		
		// Alternate between strings and objects
		if i%3 == 0 {
			data[key] = fmt.Sprintf("Value %d for testing", i)
		} else if i%3 == 1 {
			// Create a nested object
			data[key] = createNestedObject(depth, i)
		} else {
			// Create some ICU message format strings
			data[key] = createICUString(i)
		}
	}
	
	return data
}

// createNestedObject creates a nested object with the specified depth
func createNestedObject(depth, seed int) map[string]interface{} {
	if depth <= 0 {
		return nil
	}
	
	result := make(map[string]interface{})
	for i := 0; i < 3; i++ {
		nestedKey := fmt.Sprintf("nested%d", i)
		
		if i%2 == 0 {
			// Add a string value
			result[nestedKey] = fmt.Sprintf("Nested value %d-%d", seed, i)
		} else if depth > 1 {
			// Add a deeper nested object
			result[nestedKey] = createNestedObject(depth-1, seed*10+i)
		}
	}
	
	return result
}

// createICUString creates an ICU message format string
func createICUString(seed int) string {
	switch seed % 5 {
	case 0:
		return fmt.Sprintf("Hello, user%d!", seed)
	case 1:
		return fmt.Sprintf("You have {count%d, plural, one {# message} other {# messages}}.", seed)
	case 2:
		return fmt.Sprintf("Welcome to {city%d}, {name%d}!", seed, seed)
	case 3:
		return fmt.Sprintf("This is <b>bold text %d</b> and <i>italic text %d</i>.", seed, seed)
	case 4:
		return fmt.Sprintf("{count%d, plural, =0 {No results} one {# result} other {# results}} found in {time%d} ms.", seed, seed)
	default:
		return fmt.Sprintf("Text %d", seed)
	}
}
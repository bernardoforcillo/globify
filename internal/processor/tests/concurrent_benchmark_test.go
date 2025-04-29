package processor_test

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/bernardoforcillo/globify/internal/files"
	"github.com/bernardoforcillo/globify/internal/processor"
)

// enhancedMockTranslator adds a delay to the MockTranslator for benchmarking
func enhancedMockTranslator(delay time.Duration) *MockTranslator {
	return &MockTranslator{
		MockTranslate: func(text, from, to string) (string, error) {
			// Simulate network delay
			if delay > 0 {
				time.Sleep(delay)
			}
			return fmt.Sprintf("[%s] %s", to, text), nil
		},
	}
}

// BenchmarkConcurrentSimpleProcessor measures the performance of the optimized concurrent SimpleProcessor
func BenchmarkConcurrentSimpleProcessor(b *testing.B) {
	// Create a mock translator with realistic delay
	mockTranslator := enhancedMockTranslator(5 * time.Millisecond)
	
	// Create SimpleProcessor
	proc := processor.NewSimpleProcessor(mockTranslator)
	
	// Test with a large dataset that benefits from concurrency
	baseContent := createLargeDataset(100, 3) // 100 top-level items with up to 3 levels of nesting
	emptyPrevious := files.LanguageContent{}
	
	b.ResetTimer()
	b.Run("Small-Dataset", func(b *testing.B) {
		smallContent := createLargeDataset(10, 2)
		for i := 0; i < b.N; i++ {
			_, _ = proc.Execute(smallContent, "en", "fr", emptyPrevious)
		}
	})
	
	b.Run("Medium-Dataset", func(b *testing.B) {
		mediumContent := createLargeDataset(50, 3)
		for i := 0; i < b.N; i++ {
			_, _ = proc.Execute(mediumContent, "en", "fr", emptyPrevious)
		}
	})
	
	b.Run("Large-Dataset", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = proc.Execute(baseContent, "en", "fr", emptyPrevious)
		}
	})
	
	// Test with different numbers of worker threads
	for _, workers := range []int{1, 2, 4, runtime.NumCPU(), runtime.NumCPU() * 2} {
		b.Run(fmt.Sprintf("Workers-%d", workers), func(b *testing.B) {
			// Set worker pool size dynamically
			// Import the SimpleProcessor package directly to access its exported methods
			proc.SetWorkerPoolSize(workers)
			
			for i := 0; i < b.N; i++ {
				_, _ = proc.Execute(baseContent, "en", "fr", emptyPrevious)
			}
		})
	}
}

// BenchmarkConcurrentASTProcessor measures the performance of the optimized concurrent ASTProcessor
func BenchmarkConcurrentASTProcessor(b *testing.B) {
	// Create a mock translator with realistic delay
	mockTranslator := enhancedMockTranslator(5 * time.Millisecond)
	
	// Create ASTProcessor
	proc, err := processor.CreateProcessor("ast-json", mockTranslator)
	if err != nil {
		b.Skip("Failed to create ASTProcessor:", err)
		return
	}
	
	// Cast to ASTProcessor to access its methods
	astProc, ok := proc.(*processor.ASTProcessor)
	if !ok {
		b.Skip("Cannot use processor as ASTProcessor")
		return
	}
	
	// Test with a dataset containing ICU message formats
	baseContent := createLargeDataset(50, 2)
	emptyPrevious := files.LanguageContent{}
	
	b.ResetTimer()
	b.Run("Small-Dataset", func(b *testing.B) {
		smallContent := createLargeDataset(10, 1)
		for i := 0; i < b.N; i++ {
			_, _ = proc.Execute(smallContent, "en", "fr", emptyPrevious)
		}
	})
	
	b.Run("Medium-Dataset", func(b *testing.B) {
		mediumContent := createLargeDataset(30, 2)
		for i := 0; i < b.N; i++ {
			_, _ = proc.Execute(mediumContent, "en", "fr", emptyPrevious)
		}
	})
	
	b.Run("Large-Dataset", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = proc.Execute(baseContent, "en", "fr", emptyPrevious)
		}
	})
	
	// Test with different numbers of worker threads
	for _, workers := range []int{1, 2, 4, runtime.NumCPU(), runtime.NumCPU() * 2} {
		b.Run(fmt.Sprintf("Workers-%d", workers), func(b *testing.B) {
			astProc.SetWorkerPoolSize(workers)
			
			for i := 0; i < b.N; i++ {
				_, _ = proc.Execute(baseContent, "en", "fr", emptyPrevious)
			}
		})
	}
}

// BenchmarkConcurrentAppTranslation measures the performance of the entire translation process
func BenchmarkConcurrentAppTranslation(b *testing.B) {
	// This benchmark would test the parallelized App.Run method
	// with multiple target languages being processed concurrently
	// However, it would require some setup of a mock file system
	// and configuration, so we'll leave it as a placeholder
	b.Skip("App-level benchmarks require additional setup")
}
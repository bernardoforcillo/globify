package icu_test

import (
	"testing"

	"github.com/bernardoforcillo/globify/internal/icu"
)

// Benchmarks for ICU parser

func BenchmarkParseLiteral(b *testing.B) {
	message := "Hello, World! This is a simple literal string with no ICU formatting."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = icu.Parse(message)
	}
}

func BenchmarkParseSimpleArgument(b *testing.B) {
	message := "Hello, {name}! Welcome to our application."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = icu.Parse(message)
	}
}

func BenchmarkParseFormattedNumber(b *testing.B) {
	message := "You have {count, number} messages."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = icu.Parse(message)
	}
}

func BenchmarkParseFormattedDate(b *testing.B) {
	message := "Your appointment is on {date, date, short}."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = icu.Parse(message)
	}
}

func BenchmarkParseTag(b *testing.B) {
	message := "This is <b>bold</b> and <i>italic</i> text."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = icu.Parse(message)
	}
}

func BenchmarkParseComplexNested(b *testing.B) {
	message := "Hello, {name}! You have {count, number} {count, plural, one {message} other {messages}} and <b>{unread, number} unread</b>."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = icu.Parse(message)
	}
}

// Test various message lengths to see how parser scales
func BenchmarkParseLongMessage(b *testing.B) {
	// Create a long message with various ICU elements
	message := "This is a longer message with multiple elements: {name}, {count, number}, {date, date}. " +
		"It also has some <b>formatted text</b> and some <i>more formatting</i>. " +
		"We want to see how the parser performs with {count, plural, one {a longer message} other {longer messages}}. " +
		"This simulates real-world usage where translations might be paragraphs long."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = icu.Parse(message)
	}
}
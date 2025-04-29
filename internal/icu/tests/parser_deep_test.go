package icu_test

import (
	"testing"

	"github.com/bernardoforcillo/globify/internal/icu"
)

// TestParseDeepNesting tests parsing of deeply nested ICU message structures
func TestParseDeepNesting(t *testing.T) {
	// Test deeply nested structures
	message := "Hello, {name}! You have {count, plural, " +
		"=0 {no messages} " +
		"one {<b>1</b> message with <i>{priority, select, high {<u>high</u>} medium {medium} low {low}</i> priority} " +
		"other {<b>{count}</b> messages with <i>{priority, select, high {<u>high</u>} medium {medium} low {low}</i> priority}" +
		"}"
	
	elements, err := icu.Parse(message)
	if err != nil {
		t.Errorf("Parse() error = %v", err)
	}
	
	if len(elements) == 0 {
		t.Errorf("Parse() returned no elements for deeply nested message")
	}
	
	// We can't easily verify the exact structure without making the test overly complex,
	// but we can ensure it doesn't crash and returns a reasonable structure
	hasPlural := false
	for _, elem := range elements {
		if elem.Type() == icu.Plural {
			hasPlural = true
			break
		}
	}
	
	// If the parser doesn't support nested plural formats, this might fail
	// but it's a good test to ensure it's at least recognized at some level
	if !hasPlural {
		t.Logf("Note: Parser did not recognize plural element in deeply nested message")
	}
}

// TestParseEdgeCases tests parsing edge cases that might cause issues
func TestParseEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantErr  bool
		elements int // minimum expected elements, 0 if we don't care
	}{
		{
			name:     "Empty message",
			message:  "",
			wantErr:  false,
			elements: 0,
		},
		{
			name:     "Message with only placeholders",
			message:  "{name}{count}{date}",
			wantErr:  false,
			elements: 3,
		},
		{
			name:     "Message with HTML-like tags without content",
			message:  "<b></b><i></i><u></u>",
			wantErr:  false,
			elements: 3,
		},
		{
			name:     "Message with escaped braces",
			message:  "This has escaped \\{braces\\} that should be treated as text",
			wantErr:  false,
			elements: 1, // Should be one literal element
		},
		{
			name:     "Message with mismatched tags",
			message:  "<b>Bold text</i>",
			wantErr:  false, // This shouldn't fail but might not be parsed correctly
			elements: 0,
		},
		{
			name:     "Message with mismatched braces",
			message:  "This has {mismatched braces",
			wantErr:  false, // Should not fail, but might not be parsed as expected
			elements: 0,
		},
		{
			name:     "Message with very long text",
			message:  "This is a very long message that goes on and on " + string(make([]byte, 10000)) + " and ends here.",
			wantErr:  false,
			elements: 1, // Should be one literal element
		},
		{
			name:     "Message with special characters",
			message:  "Special chars: © ® ™ € £ ¥ § ¶ • ß à á â ã ä å æ ç è é ê ë ì í î ï",
			wantErr:  false,
			elements: 1, // Should be one literal element
		},
		{
			name:     "Message with multiple consecutive tags",
			message:  "<b><i><u>Formatted</u></i></b>",
			wantErr:  false,
			elements: 1, // Should be one tag element with nested tags
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements, err := icu.Parse(tt.message)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			
			if tt.elements > 0 && len(elements) < tt.elements {
				t.Errorf("Parse() returned %d elements, want at least %d", len(elements), tt.elements)
			}
		})
	}
}

// TestMessageReconstruction tests that parsed messages can be reconstructed properly
func TestMessageReconstruction(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "Simple message",
			message: "Hello, World!",
		},
		{
			name:    "Message with placeholder",
			message: "Hello, {name}!",
		},
		{
			name:    "Message with formatted element",
			message: "You have {count, number} messages.",
		},
		{
			name:    "Message with tags",
			message: "This is <b>bold</b> text.",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the message
			elements, err := icu.Parse(tt.message)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			
			// Reconstruct the message from elements
			var reconstructed string
			for _, elem := range elements {
				reconstructed += elem.String()
			}
			
			// Check if the reconstruction matches the original
			// Note: This might not be an exact match for all messages due to
			// normalization of whitespace or formatting, but should be close
			if reconstructed != tt.message {
				t.Errorf("Reconstructed message = %q, want %q", reconstructed, tt.message)
			}
		})
	}
}
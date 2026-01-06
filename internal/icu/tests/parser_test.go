package icu_test

import (
	"testing"

	"github.com/bernardoforcillo/globify/internal/icu"
)

func TestParseLiteral(t *testing.T) {
	message := "Hello, World!"
	elements, err := icu.Parse(message)

	if err != nil {
		t.Errorf("Parse() error = %v", err)
	}

	if len(elements) != 1 {
		t.Errorf("Parse() returned %d elements, want 1", len(elements))
	}

	if elements[0].Type() != icu.Literal {
		t.Errorf("Element type = %v, want %v", elements[0].Type(), icu.Literal)
	}

	literal, ok := elements[0].(icu.LiteralElement)
	if !ok {
		t.Errorf("Element is not a LiteralElement")
	}

	if literal.Value != message {
		t.Errorf("Literal value = %v, want %v", literal.Value, message)
	}
}

func TestParseArgument(t *testing.T) {
	message := "Hello, {name}!"
	elements, err := icu.Parse(message)

	if err != nil {
		t.Errorf("Parse() error = %v", err)
	}

	if len(elements) != 3 {
		t.Errorf("Parse() returned %d elements, want 3", len(elements))
	}

	// First element should be a literal "Hello, "
	if elements[0].Type() != icu.Literal {
		t.Errorf("First element type = %v, want %v", elements[0].Type(), icu.Literal)
	}

	// Second element should be an argument "{name}"
	if elements[1].Type() != icu.Argument {
		t.Errorf("Second element type = %v, want %v", elements[1].Type(), icu.Argument)
	}

	arg, ok := elements[1].(icu.ArgumentElement)
	if !ok {
		t.Errorf("Element is not an ArgumentElement")
	}

	if arg.Value != "name" {
		t.Errorf("Argument value = %v, want %v", arg.Value, "name")
	}

	// Third element should be a literal "!"
	if elements[2].Type() != icu.Literal {
		t.Errorf("Third element type = %v, want %v", elements[2].Type(), icu.Literal)
	}
}

func TestParseDoubleBraceArgument(t *testing.T) {
	message := "Hello, {{name}}!"
	elements, err := icu.Parse(message)

	if err != nil {
		t.Errorf("Parse() error = %v", err)
	}

	if len(elements) != 3 {
		t.Errorf("Parse() returned %d elements, want 3", len(elements))
	}

	// First element should be a literal "Hello, "
	if elements[0].Type() != icu.Literal {
		t.Errorf("First element type = %v, want %v", elements[0].Type(), icu.Literal)
	}

	// Second element should be an argument "{{name}}"
	if elements[1].Type() != icu.Argument {
		t.Errorf("Second element type = %v, want %v", elements[1].Type(), icu.Argument)
	}

	arg, ok := elements[1].(icu.ArgumentElement)
	if !ok {
		t.Errorf("Element is not an ArgumentElement")
	}

	if arg.Value != "name" {
		t.Errorf("Argument value = %v, want %v", arg.Value, "name")
	}

	if !arg.IsDoubleBrace {
		t.Errorf("Argument IsDoubleBrace = %v, want %v", arg.IsDoubleBrace, true)
	}

	// Third element should be a literal "!"
	if elements[2].Type() != icu.Literal {
		t.Errorf("Third element type = %v, want %v", elements[2].Type(), icu.Literal)
	}
}

func TestParseDoubleBraceFormatted(t *testing.T) {
	message := "{{value, format}}"
	elements, err := icu.Parse(message)

	if err != nil {
		t.Errorf("Parse() error = %v", err)
	}

	if len(elements) != 1 {
		t.Errorf("Parse() returned %d elements, want 1", len(elements))
	}

	arg, ok := elements[0].(icu.ArgumentElement)
	if !ok {
		t.Errorf("Element is not an ArgumentElement")
	}

	if arg.Value != "value" {
		t.Errorf("Argument value = %v, want %v", arg.Value, "value")
	}

	if arg.Style != "format" {
		t.Errorf("Argument style = %v, want %v", arg.Style, "format")
	}

	if !arg.IsDoubleBrace {
		t.Errorf("Argument IsDoubleBrace = %v, want %v", arg.IsDoubleBrace, true)
	}
}

func TestParseFormattedElement(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		elemType icu.ElementType
		value    string
		style    string
	}{
		{
			name:     "Number",
			message:  "{count, number}",
			elemType: icu.Number,
			value:    "count",
			style:    "",
		},
		{
			name:     "Number with style",
			message:  "{count, number, currency}",
			elemType: icu.Number,
			value:    "count",
			style:    "currency",
		},
		{
			name:     "Date",
			message:  "{date, date}",
			elemType: icu.Date,
			value:    "date",
			style:    "",
		},
		{
			name:     "Date with style",
			message:  "{date, date, short}",
			elemType: icu.Date,
			value:    "date",
			style:    "short",
		},
		{
			name:     "Time",
			message:  "{time, time}",
			elemType: icu.Time,
			value:    "time",
			style:    "",
		},
		{
			name:     "Time with style",
			message:  "{time, time, medium}",
			elemType: icu.Time,
			value:    "time",
			style:    "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements, err := icu.Parse(tt.message)

			if err != nil {
				t.Errorf("Parse() error = %v", err)
			}

			if len(elements) != 1 {
				t.Errorf("Parse() returned %d elements, want 1", len(elements))
			}

			if elements[0].Type() != tt.elemType {
				t.Errorf("Element type = %v, want %v", elements[0].Type(), tt.elemType)
			}

			switch elem := elements[0].(type) {
			case icu.NumberElement:
				if elem.Value != tt.value || elem.Style != tt.style {
					t.Errorf("Number element values = (%v, %v), want (%v, %v)", elem.Value, elem.Style, tt.value, tt.style)
				}
			case icu.DateElement:
				if elem.Value != tt.value || elem.Style != tt.style {
					t.Errorf("Date element values = (%v, %v), want (%v, %v)", elem.Value, elem.Style, tt.value, tt.style)
				}
			case icu.TimeElement:
				if elem.Value != tt.value || elem.Style != tt.style {
					t.Errorf("Time element values = (%v, %v), want (%v, %v)", elem.Value, elem.Style, tt.value, tt.style)
				}
			default:
				t.Errorf("Unexpected element type: %T", elem)
			}
		})
	}
}

func TestParseTag(t *testing.T) {
	message := "This is <b>bold</b> text"
	elements, err := icu.Parse(message)

	if err != nil {
		t.Errorf("Parse() error = %v", err)
	}

	if len(elements) != 3 {
		t.Errorf("Parse() returned %d elements, want 3", len(elements))
	}

	// First element should be a literal "This is "
	if elements[0].Type() != icu.Literal {
		t.Errorf("First element type = %v, want %v", elements[0].Type(), icu.Literal)
	}

	// Second element should be a tag "<b>bold</b>"
	if elements[1].Type() != icu.Tag {
		t.Errorf("Second element type = %v, want %v", elements[1].Type(), icu.Tag)
	}

	tag, ok := elements[1].(icu.TagElement)
	if !ok {
		t.Errorf("Element is not a TagElement")
	}

	if tag.Value != "b" {
		t.Errorf("Tag value = %v, want %v", tag.Value, "b")
	}

	if len(tag.Children) != 1 {
		t.Errorf("Tag has %d children, want 1", len(tag.Children))
	}

	// Child should be a literal "bold"
	if tag.Children[0].Type() != icu.Literal {
		t.Errorf("Tag child type = %v, want %v", tag.Children[0].Type(), icu.Literal)
	}

	// Third element should be a literal " text"
	if elements[2].Type() != icu.Literal {
		t.Errorf("Third element type = %v, want %v", elements[2].Type(), icu.Literal)
	}
}

func TestParseComplex(t *testing.T) {
	// Test case with a mix of elements
	message := "Hello, {name}! You have {count, number} {count, plural, one {message} other {messages}}."
	elements, err := icu.Parse(message)

	if err != nil {
		t.Errorf("Parse() error = %v", err)
	}

	// Verify we got the expected number of elements
	if len(elements) == 0 {
		t.Errorf("Parse() returned no elements")
	}

	// Just a sanity check to ensure parsing complex messages works
	// We're mostly checking that the parsing process doesn't crash
	// and returns a reasonable set of elements
	foundLiteral := false
	foundArgument := false

	for _, elem := range elements {
		switch elem.Type() {
		case icu.Literal:
			foundLiteral = true
		case icu.Argument:
			foundArgument = true
		}
	}

	if !foundLiteral {
		t.Errorf("Expected to find element of type literal")
	}
	if !foundArgument {
		t.Errorf("Expected to find element of type argument")
	}
}

func TestStringRepresentation(t *testing.T) {
	tests := []struct {
		name     string
		element  icu.Element
		expected string
	}{
		{
			name:     "Literal",
			element:  icu.LiteralElement{Value: "Hello"},
			expected: "Hello",
		},
		{
			name:     "Argument",
			element:  icu.ArgumentElement{Value: "name"},
			expected: "{name}",
		},
		{
			name:     "Argument with style",
			element:  icu.ArgumentElement{Value: "name", Style: "spellout"},
			expected: "{name, spellout}",
		},
		{
			name:     "Number",
			element:  icu.NumberElement{Value: "count"},
			expected: "{count, number}",
		},
		{
			name:     "Number with style",
			element:  icu.NumberElement{Value: "count", Style: "currency"},
			expected: "{count, number, currency}",
		},
		{
			name:     "Pound",
			element:  icu.PoundElement{},
			expected: "#",
		},
		{
			name:     "Double Brace Argument",
			element:  icu.ArgumentElement{Value: "name", IsDoubleBrace: true},
			expected: "{{name}}",
		},
		{
			name:     "Double Brace Argument with style",
			element:  icu.ArgumentElement{Value: "value", Style: "format", IsDoubleBrace: true},
			expected: "{{value, format}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.element.String()
			if result != tt.expected {
				t.Errorf("String() = %v, want %v", result, tt.expected)
			}
		})
	}
}
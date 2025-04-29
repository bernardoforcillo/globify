package icu

import (
	"fmt"
	"regexp"
	"strings"
)

// ElementType represents the type of a message element
type ElementType string

const (
	Literal  ElementType = "literal"
	Argument ElementType = "argument"
	Number   ElementType = "number"
	Date     ElementType = "date"
	Time     ElementType = "time"
	Select   ElementType = "select"
	Plural   ElementType = "plural"
	Pound    ElementType = "pound"
	Tag      ElementType = "tag"
)

// Element represents a parsed element in an ICU message
type Element interface {
	Type() ElementType
	String() string
}

// LiteralElement is a literal text element
type LiteralElement struct {
	Value string
}

func (e LiteralElement) Type() ElementType { return Literal }
func (e LiteralElement) String() string    { return e.Value }

// ArgumentElement is a placeholder element like {name}
type ArgumentElement struct {
	Value string
	Style string
}

func (e ArgumentElement) Type() ElementType { return Argument }
func (e ArgumentElement) String() string {
	if e.Style != "" {
		return fmt.Sprintf("{%s, %s}", e.Value, e.Style)
	}
	return fmt.Sprintf("{%s}", e.Value)
}

// NumberElement is a number format element like {count, number}
type NumberElement struct {
	Value string
	Style string
}

func (e NumberElement) Type() ElementType { return Number }
func (e NumberElement) String() string {
	if e.Style != "" {
		return fmt.Sprintf("{%s, number, %s}", e.Value, e.Style)
	}
	return fmt.Sprintf("{%s, number}", e.Value)
}

// DateElement is a date format element like {date, date, short}
type DateElement struct {
	Value string
	Style string
}

func (e DateElement) Type() ElementType { return Date }
func (e DateElement) String() string {
	if e.Style != "" {
		return fmt.Sprintf("{%s, date, %s}", e.Value, e.Style)
	}
	return fmt.Sprintf("{%s, date}", e.Value)
}

// TimeElement is a time format element like {time, time, short}
type TimeElement struct {
	Value string
	Style string
}

func (e TimeElement) Type() ElementType { return Time }
func (e TimeElement) String() string {
	if e.Style != "" {
		return fmt.Sprintf("{%s, time, %s}", e.Value, e.Style)
	}
	return fmt.Sprintf("{%s, time}", e.Value)
}

// SelectOption represents an option in a select element
type SelectOption struct {
	Key     string
	Value   []Element
}

// SelectElement is a select format element like {gender, select, male {...} female {...}}
type SelectElement struct {
	Value   string
	Options map[string][]Element
}

func (e SelectElement) Type() ElementType { return Select }
func (e SelectElement) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("{%s, select, ", e.Value))
	
	for key, value := range e.Options {
		sb.WriteString(key)
		sb.WriteString(" {")
		for _, element := range value {
			sb.WriteString(element.String())
		}
		sb.WriteString("} ")
	}
	
	sb.WriteString("}")
	return sb.String()
}

// PluralElement is a plural format element like {count, plural, one {...} other {...}}
type PluralElement struct {
	Value   string
	Options map[string][]Element
}

func (e PluralElement) Type() ElementType { return Plural }
func (e PluralElement) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("{%s, plural, ", e.Value))
	
	for key, value := range e.Options {
		sb.WriteString(key)
		sb.WriteString(" {")
		for _, element := range value {
			sb.WriteString(element.String())
		}
		sb.WriteString("} ")
	}
	
	sb.WriteString("}")
	return sb.String()
}

// PoundElement is a # placeholder inside a plural format
type PoundElement struct{}

func (e PoundElement) Type() ElementType { return Pound }
func (e PoundElement) String() string    { return "#" }

// TagElement is an HTML tag element like <b>...</b>
type TagElement struct {
	Value    string
	Children []Element
}

func (e TagElement) Type() ElementType { return Tag }
func (e TagElement) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<%s>", e.Value))
	
	for _, child := range e.Children {
		sb.WriteString(child.String())
	}
	
	sb.WriteString(fmt.Sprintf("</%s>", e.Value))
	return sb.String()
}

// Parse parses an ICU message into a slice of Elements
func Parse(message string) ([]Element, error) {
	// This is a simplified parser that handles ICU message format elements
	elements := []Element{}
	
	// First, identify ICU placeholders, pound signs, and literal text
	placeholderPattern := regexp.MustCompile(`\{([^{}]*)\}`)
	poundPattern := regexp.MustCompile(`#`)
	
	// Define patterns for tag detection (without backreferences)
	tagStartPattern := regexp.MustCompile(`<([^<>/]+)>`)
	
	// Current position in the message
	pos := 0
	
	// Process the message character by character
	for pos < len(message) {
		// Look for tags
		if pos < len(message) && message[pos] == '<' {
			// Check if it's an opening tag
			tagStartMatch := tagStartPattern.FindStringSubmatchIndex(message[pos:])
			if tagStartMatch != nil {
				tagName := message[pos+tagStartMatch[2]:pos+tagStartMatch[3]]
				closeTagPattern := regexp.MustCompile(`</` + regexp.QuoteMeta(tagName) + `>`)
				
				// Find the matching closing tag
				closeTagMatch := closeTagPattern.FindStringIndex(message[pos:])
				if closeTagMatch != nil {
					// Everything between tags is content
					contentStart := pos + tagStartMatch[1]
					contentEnd := pos + closeTagMatch[0]
					content := message[contentStart:contentEnd]
					
					// Parse tag content recursively
					tagChildren, err := Parse(content)
					if err != nil {
						return nil, err
					}
					
					// Add tag element
					elements = append(elements, TagElement{
						Value:    tagName,
						Children: tagChildren,
					})
					
					// Move past the closing tag
					pos = pos + closeTagMatch[1]
					continue
				}
			}
		}
		
		// Look for placeholders
		if pos < len(message) && message[pos] == '{' {
			placeholderMatch := placeholderPattern.FindStringSubmatchIndex(message[pos:])
			if placeholderMatch != nil {
				placeholder := message[pos+placeholderMatch[2]:pos+placeholderMatch[3]]
				parts := strings.SplitN(strings.TrimSpace(placeholder), ",", 3)
				
				if len(parts) == 1 {
					// Simple argument {name}
					elements = append(elements, ArgumentElement{
						Value: strings.TrimSpace(parts[0]),
					})
				} else if len(parts) >= 2 {
					// Formatted placeholder
					arg := strings.TrimSpace(parts[0])
					format := strings.TrimSpace(parts[1])
					style := ""
					if len(parts) > 2 {
						style = strings.TrimSpace(parts[2])
					}
					
					switch format {
					case "number":
						elements = append(elements, NumberElement{
							Value: arg,
							Style: style,
						})
					case "date":
						elements = append(elements, DateElement{
							Value: arg,
							Style: style,
						})
					case "time":
						elements = append(elements, TimeElement{
							Value: arg,
							Style: style,
						})
					case "select", "plural":
						// These are more complex formats
						if format == "select" {
							elements = append(elements, SelectElement{
								Value:   arg,
								Options: map[string][]Element{},
							})
						} else {
							elements = append(elements, PluralElement{
								Value:   arg,
								Options: map[string][]Element{},
							})
						}
					default:
						// Unknown format - treat as a simple argument with style
						styleStr := ""
						if style != "" {
							styleStr = ", " + style
						}
						elements = append(elements, ArgumentElement{
							Value: arg,
							Style: format + styleStr,
						})
					}
				}
				
				// Move past the placeholder
				pos = pos + placeholderMatch[1]
				continue
			}
		}
		
		// Look for pound sign
		if pos < len(message) && message[pos] == '#' {
			poundMatch := poundPattern.FindStringIndex(message[pos:])
			if poundMatch != nil {
				elements = append(elements, PoundElement{})
				
				// Move past the pound sign
				pos = pos + poundMatch[1]
				continue
			}
		}
		
		// If we're here, we're dealing with literal text
		// Collect characters until we hit a special character
		start := pos
		for pos < len(message) && message[pos] != '{' && message[pos] != '<' && message[pos] != '#' {
			pos++
		}
		
		// If we collected any text, add it as a literal
		if pos > start {
			elements = append(elements, LiteralElement{
				Value: message[start:pos],
			})
		} else {
			// Otherwise, just move to the next character
			pos++
		}
	}
	
	return elements, nil
}
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
	Value         string
	Style         string
	IsDoubleBrace bool
}

func (e ArgumentElement) Type() ElementType { return Argument }
func (e ArgumentElement) String() string {
	prefix := "{"
	suffix := "}"
	if e.IsDoubleBrace {
		prefix = "{{"
		suffix = "}}"
	}

	if e.Style != "" {
		return fmt.Sprintf("%s%s, %s%s", prefix, e.Value, e.Style, suffix)
	}
	return fmt.Sprintf("%s%s%s", prefix, e.Value, suffix)
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
	
	// Ensure specific ordering of plural options: =0, =1, other
	orderedKeys := make([]string, 0, len(e.Options))
	
	// First check for =0
	if _, hasZero := e.Options["=0"]; hasZero {
		orderedKeys = append(orderedKeys, "=0")
	}
	
	// Then check for =1
	if _, hasOne := e.Options["=1"]; hasOne {
		orderedKeys = append(orderedKeys, "=1")
	}
	
	// Add all remaining keys except 'other' which will be added last
	for key := range e.Options {
		if key != "=0" && key != "=1" && key != "other" {
			orderedKeys = append(orderedKeys, key)
		}
	}
	
	// Finally add 'other' if it exists
	if _, hasOther := e.Options["other"]; hasOther {
		orderedKeys = append(orderedKeys, "other")
	}
	
	// Write options in the determined order
	for _, key := range orderedKeys {
		sb.WriteString(key)
		sb.WriteString(" {")
		for _, element := range e.Options[key] {
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

var (
	// Patterns for parsing
	placeholderPattern   = regexp.MustCompile(`^\{([^{}]*)\}`)
	poundPattern         = regexp.MustCompile(`#`)
	tagStartPattern      = regexp.MustCompile(`<([^<>/]+)>`)
	complexFormatPattern = regexp.MustCompile(`^\{([^,{}]+),\s*(plural|select)\s*,\s*(.+)\}$`)
)

// Parse parses an ICU message into a slice of Elements
func Parse(message string) ([]Element, error) {
	return parseMessage(message, 0)
}

// parseMessage is the internal parsing function that handles nested structures
func parseMessage(message string, depth int) ([]Element, error) {
	// Prevent infinite recursion
	if depth > 100 {
		return nil, fmt.Errorf("maximum nesting depth exceeded")
	}

	elements := []Element{}
	
	// Current position in the message
	pos := 0
	
	// Process the message character by character
	for pos < len(message) {
		// Look for complex structures with nested braces
		if pos < len(message) && message[pos] == '{' {
			// Count opening and closing braces to find matching pairs
			braceDepth := 1
			startPos := pos
			nestedContent := false
			
			// Scan forward to find the matching closing brace
			for i := pos + 1; i < len(message); i++ {
				if message[i] == '{' {
					braceDepth++
					nestedContent = true
				} else if message[i] == '}' {
					braceDepth--
					if braceDepth == 0 {
						// We found the matching closing brace
						fullMatch := message[startPos : i+1]
						
						// Check if this is a complex format (plural or select)
						if nestedContent {
							// This might be a plural or select format
							complexMatch := complexFormatPattern.FindStringSubmatch(fullMatch)

							if complexMatch != nil {
								variableName := strings.TrimSpace(complexMatch[1])
								formatType := strings.TrimSpace(complexMatch[2])
								optionsText := complexMatch[3]

								options, err := parseOptions(optionsText, depth+1)
								if err != nil {
									return nil, err
								}

								// Create the appropriate element
								if formatType == "plural" {
									elements = append(elements, PluralElement{
										Value:   variableName,
										Options: options,
									})
								} else { // select
									elements = append(elements, SelectElement{
										Value:   variableName,
										Options: options,
									})
								}

								// Move past this complex structure
								pos = i + 1
								break
							}

							// Check for double braces {{...}}
							if strings.HasPrefix(fullMatch, "{{") && strings.HasSuffix(fullMatch, "}}") {
								content := fullMatch[2 : len(fullMatch)-2]
								parts := strings.SplitN(strings.TrimSpace(content), ",", 3)

								element := ArgumentElement{
									Value:         strings.TrimSpace(parts[0]),
									IsDoubleBrace: true,
								}

								if len(parts) >= 2 {
									// In double braces, we treat everything after the first comma as style/format
									// e.g. {{value, format}} -> Value: value, Style: format
									rest := strings.TrimSpace(parts[1])
									if len(parts) > 2 {
										rest += ", " + strings.TrimSpace(parts[2])
									}
									element.Style = rest
								}

								elements = append(elements, element)
								pos = i + 1
								break
							}
						}

						// It's not a complex format, so fall back to normal placeholder handling
						break
					}
				}
			}

			// If we didn't handle it as a complex format, process it as a regular placeholder
			if !nestedContent || pos < len(message) && message[pos] == '{' {
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
						default:
							// Unknown format - treat as a simple argument with style
							styleStr := ""
							if style != "" {
								styleStr = format + ", " + style
							} else {
								styleStr = format
							}
							elements = append(elements, ArgumentElement{
								Value: arg,
								Style: styleStr,
							})
						}
					}
					
					// Move past the placeholder
					pos = pos + placeholderMatch[1]
					continue
				}
			}
		}
		
		// Look for tags
		if pos < len(message) && message[pos] == '<' {
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
					tagChildren, err := parseMessage(content, depth+1)
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

// parseOptions parses the options part of a plural or select format
func parseOptions(optionsText string, depth int) (map[string][]Element, error) {
	options := make(map[string][]Element)
	
	// Parse options like "one {You have one notification} other {You have # notifications}"
	pos := 0
	for pos < len(optionsText) {
		// Skip whitespace
		for pos < len(optionsText) && (optionsText[pos] == ' ' || optionsText[pos] == '\t' || optionsText[pos] == '\n') {
			pos++
		}
		
		if pos >= len(optionsText) {
			break
		}
		
		// Find the option key (everything up to the opening brace)
		keyStart := pos
		for pos < len(optionsText) && optionsText[pos] != '{' {
			pos++
		}
		
		if pos >= len(optionsText) {
			break
		}
		
		key := strings.TrimSpace(optionsText[keyStart:pos])
		
		// Find the matching closing brace for this option value
		braceDepth := 1
		valueStart := pos + 1 // Skip the opening brace
		
		for pos++; pos < len(optionsText) && braceDepth > 0; pos++ {
			if optionsText[pos] == '{' {
				braceDepth++
			} else if optionsText[pos] == '}' {
				braceDepth--
			}
		}
		
		if braceDepth > 0 {
			return nil, fmt.Errorf("unbalanced braces in options text: %s", optionsText)
		}
		
		// Extract the option value (excluding the outer braces)
		valueEnd := pos - 1 // Exclude the closing brace
		value := optionsText[valueStart:valueEnd]
		
		// Parse the option value recursively
		elements, err := parseMessage(value, depth+1)
		if err != nil {
			return nil, err
		}
		
		options[key] = elements
	}
	
	return options, nil
}
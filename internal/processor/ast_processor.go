package processor

import (
	"fmt"
	"log"

	"github.com/bernardoforcillo/globify/internal/files"
	"github.com/bernardoforcillo/globify/internal/icu"
	"github.com/bernardoforcillo/globify/internal/translator"
)

// ASTProcessor handles translation of ICU message format strings
type ASTProcessor struct {
	translator translator.Translator
}

// NewASTProcessor creates a new ASTProcessor
func NewASTProcessor(translator translator.Translator) *ASTProcessor {
	return &ASTProcessor{
		translator: translator,
	}
}

// Execute translates content with ICU message format strings
func (p *ASTProcessor) Execute(
	obj files.LanguageContent,
	from, target string,
	previousTranslation files.LanguageContent,
) (files.LanguageContent, error) {
	result := make(files.LanguageContent)

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
	
	for _, element := range elements {
		switch element.Type() {
		case icu.Literal:
			// Only translate literal text elements
			lit := element.(icu.LiteralElement)
			if lit.Value == "" {
				result += ""
				continue
			}
			
			translated, err := p.translator.Translate(lit.Value, from, target)
			if err != nil {
				return "", fmt.Errorf("failed to translate literal: %w", err)
			}
			result += translated
			
		case icu.Tag:
			// Handle tag elements by translating their children
			tag := element.(icu.TagElement)
			translatedContent, err := p.translateElements(tag.Children, from, target)
			if err != nil {
				return "", fmt.Errorf("failed to translate tag content: %w", err)
			}
			result += fmt.Sprintf("<%s>%s</%s>", tag.Value, translatedContent, tag.Value)
			
		case icu.Select:
			// Handle select elements by translating each option
			sel := element.(icu.SelectElement)
			translatedOptions := make(map[string]string)
			
			for key, option := range sel.Options {
				translatedOption, err := p.translateElements(option, from, target)
				if err != nil {
					return "", fmt.Errorf("failed to translate select option: %w", err)
				}
				translatedOptions[key] = translatedOption
			}
			
			// Reconstruct the select format
			selectStr := fmt.Sprintf("{%s, select, ", sel.Value)
			for key, value := range translatedOptions {
				selectStr += fmt.Sprintf("%s {%s} ", key, value)
			}
			selectStr += "}"
			result += selectStr
			
		case icu.Plural:
			// Handle plural elements by translating each option
			plural := element.(icu.PluralElement)
			translatedOptions := make(map[string]string)
			
			for key, option := range plural.Options {
				translatedOption, err := p.translateElements(option, from, target)
				if err != nil {
					return "", fmt.Errorf("failed to translate plural option: %w", err)
				}
				translatedOptions[key] = translatedOption
			}
			
			// Reconstruct the plural format
			pluralStr := fmt.Sprintf("{%s, plural, ", plural.Value)
			for key, value := range translatedOptions {
				pluralStr += fmt.Sprintf("%s {%s} ", key, value)
			}
			pluralStr += "}"
			result += pluralStr
			
		default:
			// Keep other elements as they are
			result += element.String()
		}
	}
	
	return result, nil
}
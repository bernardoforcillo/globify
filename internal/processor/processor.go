package processor

import (
	"fmt"

	"github.com/bernardoforcillo/globify/internal/files"
	"github.com/bernardoforcillo/globify/internal/translator"
)

// ObjectProcessor defines the interface for translating language content
type ObjectProcessor interface {
	Execute(obj files.LanguageContent, from, target string, previousTranslation files.LanguageContent) (files.LanguageContent, error)
}

// CreateProcessor returns the appropriate processor based on the translation type
func CreateProcessor(translationType string, translator translator.Translator) (ObjectProcessor, error) {
	switch translationType {
	case "simple-json":
		return NewSimpleProcessor(translator), nil
	case "ast-json":
		return NewASTProcessor(translator), nil
	default:
		return nil, fmt.Errorf("unsupported translation type: %s", translationType)
	}
}
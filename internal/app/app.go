package app

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/bernardoforcillo/globify/internal/config"
	"github.com/bernardoforcillo/globify/internal/files"
	"github.com/bernardoforcillo/globify/internal/processor"
	"github.com/bernardoforcillo/globify/internal/translator"
)

// App coordinates the translation process
type App struct {
	config     *config.Config
	translator translator.Translator
	fileManager files.FileManager
	processor  processor.ObjectProcessor
}

// NewApp creates and initializes a new App instance
func NewApp() (*App, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create translator
	trans, err := translator.CreateTranslator()
	if err != nil {
		return nil, fmt.Errorf("failed to create translator: %w", err)
	}

	// Create file manager
	fm, err := files.NewFileManager(cfg.FileExtension)
	if err != nil {
		return nil, fmt.Errorf("failed to create file manager: %w", err)
	}

	// Create appropriate processor
	proc, err := processor.CreateProcessor(cfg.TranslationType, trans)
	if err != nil {
		return nil, fmt.Errorf("failed to create processor: %w", err)
	}

	return &App{
		config:     cfg,
		translator: trans,
		fileManager: fm,
		processor:  proc,
	}, nil
}

// Run performs the translation process
func (a *App) Run() error {
	log.Printf("Starting translation from %s to %v", a.config.BaseLanguage, a.config.Languages)
	
	// Read the base language file
	baseFilePath := filepath.Join(a.config.Folder, fmt.Sprintf("%s.%s", a.config.BaseLanguage, a.config.FileExtension))
	log.Printf("Reading base language file: %s", baseFilePath)
	
	baseContent, err := a.fileManager.Read(baseFilePath)
	if err != nil {
		return fmt.Errorf("failed to read base language file: %w", err)
	}
	
	// Process each language sequentially instead of concurrently
	for _, lang := range a.config.Languages {
		// Skip if target language is the same as base language
		if lang == a.config.BaseLanguage {
			continue
		}
		
		log.Printf("Translating from %s to %s", a.config.BaseLanguage, lang)
		
		// Path for the target language file
		targetFilePath := filepath.Join(a.config.Folder, fmt.Sprintf("%s.%s", lang, a.config.FileExtension))
		
		// Check if target file already exists
		var previousContent files.LanguageContent
		exists, fileErr := a.fileManager.Exists(targetFilePath)
		if fileErr != nil {
			log.Printf("Warning: Error checking existence of %s: %v", targetFilePath, fileErr)
		}
		
		if exists {
			log.Printf("Target file %s already exists, using existing translations as baseline", targetFilePath)
			previousContent, fileErr = a.fileManager.Read(targetFilePath)
			if fileErr != nil {
				log.Printf("Warning: Failed to read existing target file %s: %v", targetFilePath, fileErr)
				previousContent = make(files.LanguageContent)
			}
		} else {
			previousContent = make(files.LanguageContent)
		}
		
		// Process translations
		translatedContent, procErr := a.processor.Execute(
			baseContent,
			a.config.BaseLanguage,
			lang,
			previousContent,
		)
		if procErr != nil {
			return fmt.Errorf("failed to translate to %s: %w", lang, procErr)
		}
		
		// Write translated content to file
		log.Printf("Writing translated content to %s", targetFilePath)
		if writeErr := a.fileManager.Write(targetFilePath, translatedContent); writeErr != nil {
			return fmt.Errorf("failed to write translated file %s: %w", targetFilePath, writeErr)
		}
		
		log.Printf("Successfully translated to %s", lang)
	}
	
	log.Printf("Translation process completed successfully")
	return nil
}
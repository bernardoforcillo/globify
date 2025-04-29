package translator

// Translator defines the interface for translation services
type Translator interface {
	Translate(text, from, to string) (string, error)
}

// Factory function to create a translator based on environment variables
func CreateTranslator() (Translator, error) {
	return NewDeeplTranslator()
}
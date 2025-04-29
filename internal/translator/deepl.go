package translator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// DeeplTranslator implements the Translator interface using DeepL API
type DeeplTranslator struct {
	apiKey string
	client *http.Client
}

type deeplResponse struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                  string `json:"text"`
	} `json:"translations"`
}

// NewDeeplTranslator creates a new DeepL translator using environment variables
func NewDeeplTranslator() (*DeeplTranslator, error) {
	apiKey := os.Getenv("DEEPL_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPL_API_KEY environment variable is not set")
	}
	return &DeeplTranslator{
		apiKey: apiKey,
		client: &http.Client{},
	}, nil
}

// Translate implements the Translator interface for DeepL
func (t *DeeplTranslator) Translate(text, from, to string) (string, error) {
	if text == "" {
		return "", nil
	}
	
	if from == to {
		return text, nil // No need to translate if source and target languages are the same
	}

	apiURL := "https://api-free.deepl.com/v2/translate"
	
	data := url.Values{}
	data.Set("text", text)
	data.Set("target_lang", to)
	if from != "" {
		data.Set("source_lang", from)
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create DeepL request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "DeepL-Auth-Key "+t.apiKey)

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to DeepL: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read DeepL response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepL API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result deeplResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse DeepL response JSON: %w", err)
	}

	if len(result.Translations) == 0 {
		return "", fmt.Errorf("DeepL response contained no translations")
	}

	return result.Translations[0].Text, nil
}
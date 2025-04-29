package translator

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// DeeplTranslator implements the Translator interface using DeepL API
type DeeplTranslator struct {
	apiKey string
	client *http.Client
	maxRetries int
	initialBackoff time.Duration
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
		maxRetries: 5,                // Maximum number of retry attempts
		initialBackoff: time.Second,  // Start with 1 second delay before first retry
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

	// Initialize variables for retry mechanism
	var (
		resp *http.Response
		respErr error
		body []byte
		attempts int
		backoff = t.initialBackoff
	)

	// Random source for jitter
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Try the request with exponential backoff and jitter
	for attempts = 0; attempts <= t.maxRetries; attempts++ {
		if attempts > 0 {
			// Calculate backoff with jitter for retry
			jitter := time.Duration(r.Float64() * float64(backoff) * 0.3) // 30% jitter
			sleepTime := backoff + jitter
			
			// Log the retry attempt
			fmt.Printf("Rate limit exceeded. Retrying in %.2f seconds (attempt %d/%d)...\n", 
				sleepTime.Seconds(), attempts, t.maxRetries)
			
			time.Sleep(sleepTime)
			
			// Exponential backoff for next iteration
			backoff = time.Duration(float64(backoff) * 2)
		}

		// Create a new request for each attempt
		req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
		if err != nil {
			return "", fmt.Errorf("failed to create DeepL request: %w", err)
		}

		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Authorization", "DeepL-Auth-Key "+t.apiKey)

		resp, respErr = t.client.Do(req)
		if respErr != nil {
			// Network errors are retryable
			continue
		}

		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("failed to read DeepL response body: %w", err)
		}

		// If not a rate limit error, break out of the retry loop
		if resp.StatusCode != http.StatusTooManyRequests {
			break
		}
	}

	// If we exhausted all retries and still getting errors
	if attempts > t.maxRetries {
		return "", fmt.Errorf("exceeded maximum retries (%d) for DeepL API", t.maxRetries)
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
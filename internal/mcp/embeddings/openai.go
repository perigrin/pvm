// ABOUTME: OpenAI embedding provider implementation
// ABOUTME: Provides embeddings using OpenAI's text-embedding models

package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	openAIAPIURL = "https://api.openai.com/v1/embeddings"
	defaultModel = "text-embedding-3-small"
	maxBatchSize = 100
)

// OpenAIProvider implements EmbeddingProvider using OpenAI's API
type OpenAIProvider struct {
	apiKey     string
	model      string
	dimensions int
	maxTokens  int
	httpClient *http.Client
}

// NewOpenAIProvider creates a new OpenAI embedding provider
func NewOpenAIProvider(config EmbeddingConfig) (*OpenAIProvider, error) {
	if config.APIKey == "" {
		return nil, ErrInvalidAPIKey
	}

	model := config.Model
	if model == "" {
		model = defaultModel
	}

	// Set dimensions based on model
	dimensions := config.Dimensions
	if dimensions == 0 {
		switch model {
		case "text-embedding-3-small":
			dimensions = 1536
		case "text-embedding-3-large":
			dimensions = 3072
		case "text-embedding-ada-002":
			dimensions = 1536
		default:
			dimensions = 1536
		}
	}

	// Max tokens for OpenAI models
	maxTokens := config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 8191 // Default for most OpenAI embedding models
	}

	return &OpenAIProvider{
		apiKey:     config.APIKey,
		model:      model,
		dimensions: dimensions,
		maxTokens:  maxTokens,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// EmbedText generates embeddings for a single text
func (p *OpenAIProvider) EmbedText(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := p.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, ErrEmbeddingFailed
	}
	return embeddings[0], nil
}

// EmbedBatch generates embeddings for multiple texts
func (p *OpenAIProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	if len(texts) > maxBatchSize {
		return nil, ErrBatchTooLarge
	}

	// Clean texts (remove excessive whitespace)
	cleanedTexts := make([]string, len(texts))
	for i, text := range texts {
		cleanedTexts[i] = strings.TrimSpace(text)
	}

	// Create request
	reqBody := openAIRequest{
		Model: p.model,
		Input: cleanedTexts,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", openAIAPIURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Send request
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		var errorResp openAIErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error.Message != "" {
			return nil, fmt.Errorf("OpenAI API error: %s", errorResp.Error.Message)
		}
		return nil, fmt.Errorf("OpenAI API error: status %d", resp.StatusCode)
	}

	// Parse response
	var result openAIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract embeddings
	embeddings := make([][]float32, len(result.Data))
	for i, data := range result.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}

// Dimensions returns the embedding dimensions
func (p *OpenAIProvider) Dimensions() int {
	return p.dimensions
}

// MaxTokens returns the maximum tokens
func (p *OpenAIProvider) MaxTokens() int {
	return p.maxTokens
}

// IsAvailable checks if the provider is available
func (p *OpenAIProvider) IsAvailable() bool {
	return p.apiKey != ""
}

// OpenAI API types
type openAIRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openAIResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

type openAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Param   string `json:"param"`
		Code    string `json:"code"`
	} `json:"error"`
}

func init() {
	// Register OpenAI provider
	RegisterProvider("openai", func(config EmbeddingConfig) (EmbeddingProvider, error) {
		return NewOpenAIProvider(config)
	})
}

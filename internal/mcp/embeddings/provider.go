// ABOUTME: Embedding provider interface for MCP server
// ABOUTME: Defines the contract for different embedding providers

package embeddings

import (
	"context"
	"errors"
)

// EmbeddingProvider defines the interface for embedding providers
type EmbeddingProvider interface {
	// Name returns the name of the embedding provider
	Name() string

	// EmbedText generates embeddings for a single text
	EmbedText(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch generates embeddings for multiple texts
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

	// Dimensions returns the dimension size of embeddings
	Dimensions() int

	// MaxTokens returns the maximum number of tokens this provider can handle
	MaxTokens() int

	// IsAvailable checks if the provider is available and configured
	IsAvailable() bool
}

// EmbeddingConfig holds configuration for embedding providers
type EmbeddingConfig struct {
	Provider   string                 `json:"provider"`   // openai, local, etc.
	Model      string                 `json:"model"`      // model name
	APIKey     string                 `json:"api_key"`    // API key for remote providers
	Dimensions int                    `json:"dimensions"` // embedding dimensions
	MaxTokens  int                    `json:"max_tokens"` // max tokens per text
	BatchSize  int                    `json:"batch_size"` // batch size for batch operations
	Extra      map[string]interface{} `json:"extra"`      // provider-specific config
}

// Common errors
var (
	ErrProviderNotAvailable = errors.New("embedding provider not available")
	ErrTextTooLong          = errors.New("text exceeds maximum token limit")
	ErrBatchTooLarge        = errors.New("batch size exceeds limit")
	ErrInvalidAPIKey        = errors.New("invalid or missing API key")
	ErrEmbeddingFailed      = errors.New("failed to generate embeddings")
)

// ProviderFactory creates embedding providers based on configuration
type ProviderFactory func(config EmbeddingConfig) (EmbeddingProvider, error)

var providerFactories = make(map[string]ProviderFactory)

// RegisterProvider registers an embedding provider factory
func RegisterProvider(name string, factory ProviderFactory) {
	providerFactories[name] = factory
}

// NewProvider creates a new embedding provider based on configuration
func NewProvider(config EmbeddingConfig) (EmbeddingProvider, error) {
	factory, exists := providerFactories[config.Provider]
	if !exists {
		return nil, errors.New("unknown embedding provider: " + config.Provider)
	}
	return factory(config)
}

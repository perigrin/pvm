// ABOUTME: Local embedding provider implementation
// ABOUTME: Provides simple local embeddings as a fallback when remote providers are unavailable

package embeddings

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"strings"
)

// LocalProvider implements a simple local embedding provider
// This is a fallback option that generates deterministic embeddings based on text content
type LocalProvider struct {
	dimensions int
	maxTokens  int
}

// NewLocalProvider creates a new local embedding provider
func NewLocalProvider(config EmbeddingConfig) (*LocalProvider, error) {
	dimensions := config.Dimensions
	if dimensions == 0 {
		dimensions = 384 // Default dimension size
	}

	maxTokens := config.MaxTokens
	if maxTokens == 0 {
		maxTokens = 512 // Conservative limit for local processing
	}

	return &LocalProvider{
		dimensions: dimensions,
		maxTokens:  maxTokens,
	}, nil
}

// Name returns the provider name
func (p *LocalProvider) Name() string {
	return "local"
}

// EmbedText generates embeddings for a single text
func (p *LocalProvider) EmbedText(ctx context.Context, text string) ([]float32, error) {
	// Simple deterministic embedding generation
	// This is not semantically meaningful but provides consistent vectors for testing

	// Clean and tokenize text
	text = strings.ToLower(strings.TrimSpace(text))
	tokens := strings.Fields(text)

	// Check token limit
	if len(tokens) > p.maxTokens {
		tokens = tokens[:p.maxTokens]
	}

	// Generate embedding
	embedding := make([]float32, p.dimensions)

	// Use SHA256 to generate deterministic values
	for i, token := range tokens {
		hash := sha256.Sum256([]byte(token))

		// Distribute hash values across embedding dimensions
		for j := 0; j < p.dimensions && j < len(hash); j++ {
			// Combine token position and hash value
			idx := (i + j) % p.dimensions
			value := float32(hash[j]) / 255.0 // Normalize to [0, 1]

			// Accumulate values with decay based on token position
			decay := float32(math.Exp(-float64(i) * 0.1))
			embedding[idx] += value * decay
		}
	}

	// Normalize the embedding
	norm := float32(0)
	for _, v := range embedding {
		norm += v * v
	}
	norm = float32(math.Sqrt(float64(norm)))

	if norm > 0 {
		for i := range embedding {
			embedding[i] /= norm
		}
	}

	return embedding, nil
}

// EmbedBatch generates embeddings for multiple texts
func (p *LocalProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))

	for i, text := range texts {
		embedding, err := p.EmbedText(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// Dimensions returns the embedding dimensions
func (p *LocalProvider) Dimensions() int {
	return p.dimensions
}

// MaxTokens returns the maximum tokens
func (p *LocalProvider) MaxTokens() int {
	return p.maxTokens
}

// IsAvailable checks if the provider is available
func (p *LocalProvider) IsAvailable() bool {
	return true // Local provider is always available
}

// Helper function to convert bytes to float32
func bytesToFloat32(b []byte) float32 {
	bits := binary.LittleEndian.Uint32(b)
	return math.Float32frombits(bits)
}

func init() {
	// Register local provider
	RegisterProvider("local", func(config EmbeddingConfig) (EmbeddingProvider, error) {
		return NewLocalProvider(config)
	})
}

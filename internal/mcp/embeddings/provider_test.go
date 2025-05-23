// ABOUTME: Tests for embedding provider interface and implementations
// ABOUTME: Verifies provider registration and basic functionality

package embeddings

import (
	"context"
	"testing"
)

func TestProviderRegistration(t *testing.T) {
	// Test that providers are registered
	providers := []string{"openai", "local"}

	for _, name := range providers {
		t.Run(name, func(t *testing.T) {
			config := EmbeddingConfig{
				Provider: name,
				APIKey:   "test-key", // For providers that need it
			}

			provider, err := NewProvider(config)
			if err != nil {
				t.Fatalf("Failed to create provider %s: %v", name, err)
			}

			if provider.Name() != name {
				t.Errorf("Expected provider name %s, got %s", name, provider.Name())
			}
		})
	}
}

func TestLocalProvider(t *testing.T) {
	config := EmbeddingConfig{
		Provider:   "local",
		Dimensions: 128,
		MaxTokens:  100,
	}

	provider, err := NewProvider(config)
	if err != nil {
		t.Fatalf("Failed to create local provider: %v", err)
	}

	ctx := context.Background()

	t.Run("single_embedding", func(t *testing.T) {
		text := "This is a test text for embedding"
		embedding, err := provider.EmbedText(ctx, text)
		if err != nil {
			t.Fatalf("Failed to generate embedding: %v", err)
		}

		if len(embedding) != 128 {
			t.Errorf("Expected embedding dimension 128, got %d", len(embedding))
		}

		// Check that embedding is normalized
		var norm float32
		for _, v := range embedding {
			norm += v * v
		}
		if norm < 0.99 || norm > 1.01 {
			t.Errorf("Embedding not properly normalized, norm = %f", norm)
		}
	})

	t.Run("batch_embedding", func(t *testing.T) {
		texts := []string{
			"First text",
			"Second text",
			"Third text",
		}

		embeddings, err := provider.EmbedBatch(ctx, texts)
		if err != nil {
			t.Fatalf("Failed to generate batch embeddings: %v", err)
		}

		if len(embeddings) != 3 {
			t.Errorf("Expected 3 embeddings, got %d", len(embeddings))
		}

		for i, embedding := range embeddings {
			if len(embedding) != 128 {
				t.Errorf("Embedding %d has wrong dimension: %d", i, len(embedding))
			}
		}
	})

	t.Run("deterministic", func(t *testing.T) {
		text := "Test deterministic embedding"

		embedding1, err := provider.EmbedText(ctx, text)
		if err != nil {
			t.Fatalf("Failed to generate first embedding: %v", err)
		}

		embedding2, err := provider.EmbedText(ctx, text)
		if err != nil {
			t.Fatalf("Failed to generate second embedding: %v", err)
		}

		// Check that embeddings are identical
		for i := range embedding1 {
			if embedding1[i] != embedding2[i] {
				t.Errorf("Embeddings not deterministic at index %d: %f != %f",
					i, embedding1[i], embedding2[i])
				break
			}
		}
	})
}

func TestProviderDimensions(t *testing.T) {
	tests := []struct {
		name       string
		config     EmbeddingConfig
		wantDim    int
		wantTokens int
	}{
		{
			name: "local_default",
			config: EmbeddingConfig{
				Provider: "local",
			},
			wantDim:    384,
			wantTokens: 512,
		},
		{
			name: "local_custom",
			config: EmbeddingConfig{
				Provider:   "local",
				Dimensions: 256,
				MaxTokens:  1024,
			},
			wantDim:    256,
			wantTokens: 1024,
		},
		{
			name: "openai_default",
			config: EmbeddingConfig{
				Provider: "openai",
				APIKey:   "test-key",
			},
			wantDim:    1536,
			wantTokens: 8191,
		},
		{
			name: "openai_ada",
			config: EmbeddingConfig{
				Provider: "openai",
				APIKey:   "test-key",
				Model:    "text-embedding-ada-002",
			},
			wantDim:    1536,
			wantTokens: 8191,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(tt.config)
			if err != nil {
				t.Fatalf("Failed to create provider: %v", err)
			}

			if provider.Dimensions() != tt.wantDim {
				t.Errorf("Expected dimensions %d, got %d", tt.wantDim, provider.Dimensions())
			}

			if provider.MaxTokens() != tt.wantTokens {
				t.Errorf("Expected max tokens %d, got %d", tt.wantTokens, provider.MaxTokens())
			}
		})
	}
}

func TestProviderAvailability(t *testing.T) {
	tests := []struct {
		name      string
		config    EmbeddingConfig
		wantAvail bool
	}{
		{
			name: "local_always_available",
			config: EmbeddingConfig{
				Provider: "local",
			},
			wantAvail: true,
		},
		{
			name: "openai_with_key",
			config: EmbeddingConfig{
				Provider: "openai",
				APIKey:   "test-key",
			},
			wantAvail: true,
		},
		{
			name: "openai_without_key",
			config: EmbeddingConfig{
				Provider: "openai",
			},
			wantAvail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Special handling for OpenAI without key
			if tt.config.Provider == "openai" && tt.config.APIKey == "" {
				_, err := NewProvider(tt.config)
				if err != ErrInvalidAPIKey {
					t.Errorf("Expected ErrInvalidAPIKey for OpenAI without key")
				}
				return
			}

			provider, err := NewProvider(tt.config)
			if err != nil {
				t.Fatalf("Failed to create provider: %v", err)
			}

			if provider.IsAvailable() != tt.wantAvail {
				t.Errorf("Expected availability %v, got %v", tt.wantAvail, provider.IsAvailable())
			}
		})
	}
}

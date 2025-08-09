// ABOUTME: Non-embedded documentation fallback for development builds
// ABOUTME: Provides external documentation manager when embed is disabled

//go:build !embed

package docs

import (
	"embed"

	"tamarou.com/pvm/internal/config"
)

// Empty embed.FS for development builds
var embeddedDocs embed.FS

// NewDocumentManager creates a document manager for external documentation
func NewDocumentManager() (DocumentManager, error) {
	// For development builds, return external manager with default config
	docsConfig := &config.DocsConfig{
		Repository:     "perigrin/pvm",
		Branch:         "pu",
		CacheDir:       "/tmp/pvm-docs-cache",
		CacheTTL:       24,
		MaxRetries:     3,
		Timeout:        "30s",
		OfflineMode:    false,
		PreferEmbedded: false,
	}

	// No fallback manager for external-only builds
	return NewExternalDocumentManager(docsConfig, nil), nil
}

// IsEmbedded returns false when documentation is not embedded
func IsEmbedded() bool {
	return false
}

// ABOUTME: Non-embedded documentation fallback for development builds
// ABOUTME: Provides external documentation manager when embed is disabled

//go:build !embed

package docs

import (
	"embed"
)

// Empty embed.FS for development builds
var embeddedDocs embed.FS

// NewDocumentManager creates a document manager for external documentation
func NewDocumentManager() (DocumentManager, error) {
	// For development builds, return external manager pointing to GitHub
	baseURL := "https://github.com/perigrin/pvm/blob/pu/docs"
	return NewExternalDocumentManager(baseURL), nil
}

// IsEmbedded returns false when documentation is not embedded
func IsEmbedded() bool {
	return false
}

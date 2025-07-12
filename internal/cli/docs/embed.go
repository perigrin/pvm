// ABOUTME: Embedded documentation content for release builds
// ABOUTME: Uses Go embed directive to include documentation in binary

//go:build embed

package docs

import (
	"embed"
)

// For now, embed just our test data to verify the system works
//
//go:embed testdata/*.md
var embeddedDocs embed.FS

// NewDocumentManager creates a document manager for embedded documentation
func NewDocumentManager() (DocumentManager, error) {
	return NewEmbeddedDocumentManager(embeddedDocs)
}

// IsEmbedded returns true when documentation is embedded
func IsEmbedded() bool {
	return true
}

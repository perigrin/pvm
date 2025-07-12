// ABOUTME: Simple embedded documentation test for debugging
// ABOUTME: Tests basic embed functionality with single file

//go:build embed

package docs

import (
	"embed"
)

//go:embed testdata/test-guide.md
var simpleTestDocs embed.FS

// TestEmbedWorking tests if basic embed functionality works
func TestEmbedWorking() (DocumentManager, error) {
	return NewEmbeddedDocumentManager(simpleTestDocs)
}

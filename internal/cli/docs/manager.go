// ABOUTME: Document management system for embedding help documentation
// ABOUTME: Provides interface for accessing embedded or external documentation

package docs

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DocumentInfo contains metadata about a document
type DocumentInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	ModTime     time.Time `json:"mod_time"`
}

// SearchResult represents a search match within documentation
type SearchResult struct {
	Document DocumentInfo `json:"document"`
	Matches  []string     `json:"matches"`
	Score    int          `json:"score"`
}

// DocumentManager provides access to embedded or external documentation
type DocumentManager interface {
	// GetDocument retrieves the content of a specific document
	GetDocument(name string) ([]byte, error)

	// ListDocuments returns all available documents
	ListDocuments() ([]DocumentInfo, error)

	// SearchDocuments searches for documents containing the query
	SearchDocuments(query string) ([]SearchResult, error)

	// IsAvailable returns true if documentation is available
	IsAvailable() bool

	// GetCategories returns all available document categories
	GetCategories() []string

	// GetDocumentsByCategory returns documents in a specific category
	GetDocumentsByCategory(category string) ([]DocumentInfo, error)
}

// EmbeddedDocumentManager handles embedded documentation
type EmbeddedDocumentManager struct {
	content embed.FS
	index   map[string]DocumentInfo
	cache   map[string][]byte
}

// NewEmbeddedDocumentManager creates a new embedded document manager
func NewEmbeddedDocumentManager(content embed.FS) (*EmbeddedDocumentManager, error) {
	manager := &EmbeddedDocumentManager{
		content: content,
		index:   make(map[string]DocumentInfo),
		cache:   make(map[string][]byte),
	}

	if err := manager.buildIndex(); err != nil {
		return nil, fmt.Errorf("failed to build document index: %w", err)
	}

	return manager, nil
}

// buildIndex creates an index of all embedded documents
func (e *EmbeddedDocumentManager) buildIndex() error {
	return fs.WalkDir(e.content, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		name := strings.TrimSuffix(filepath.Base(path), ".md")
		category := e.categorizeDocument(path, name)
		description := e.generateDescription(name)

		e.index[name] = DocumentInfo{
			Name:        name,
			Path:        path,
			Size:        info.Size(),
			Category:    category,
			Description: description,
			ModTime:     info.ModTime(),
		}

		return nil
	})
}

// categorizeDocument determines the category for a document
func (e *EmbeddedDocumentManager) categorizeDocument(path, name string) string {
	if strings.HasPrefix(path, "archive/") {
		return "Archive"
	}

	if strings.HasPrefix(name, "workflow-") {
		return "Workflows"
	}

	switch name {
	case "quickstart", "getting-started":
		return "Getting Started"
	case "command-reference", "quick-reference":
		return "Reference"
	case "developer-guide", "contributor-guide":
		return "Development"
	case "architecture-overview", "compiler_architecture":
		return "Architecture"
	case "troubleshooting":
		return "Support"
	default:
		if strings.Contains(name, "guide") {
			return "Guides"
		}
		if strings.Contains(name, "specification") {
			return "Specifications"
		}
		return "Documentation"
	}
}

// generateDescription creates a description for a document based on its name
func (e *EmbeddedDocumentManager) generateDescription(name string) string {
	descriptions := map[string]string{
		"quickstart":                "Quick start guide for new users",
		"command-reference":         "Complete command reference",
		"developer-guide":           "Guide for PVM developers",
		"contributor-guide":         "How to contribute to PVM",
		"architecture-overview":     "System architecture overview",
		"compiler_architecture":     "Compiler architecture details",
		"troubleshooting":           "Common issues and solutions",
		"typed-perl-specification":  "Type annotation specification",
		"type-annotation-reference": "Type annotation reference",
		"workflow-new-development":  "Workflow for new project development",
		"workflow-legacy-codebase":  "Workflow for legacy codebase migration",
		"workflow-existing-project": "Workflow for existing projects",
	}

	if desc, exists := descriptions[name]; exists {
		return desc
	}

	// Generate description from name
	text := strings.ReplaceAll(name, "-", " ")
	text = strings.ReplaceAll(text, "_", " ")
	parts := strings.Fields(text)
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, " ")
}

// GetDocument retrieves a document by name
func (e *EmbeddedDocumentManager) GetDocument(name string) ([]byte, error) {
	// Check cache first
	if content, exists := e.cache[name]; exists {
		return content, nil
	}

	// Find document in index
	docInfo, exists := e.index[name]
	if !exists {
		return nil, fmt.Errorf("document not found: %s", name)
	}

	// Read from embedded filesystem
	content, err := fs.ReadFile(e.content, docInfo.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read document %s: %w", name, err)
	}

	// Cache the content
	e.cache[name] = content

	return content, nil
}

// ListDocuments returns all available documents
func (e *EmbeddedDocumentManager) ListDocuments() ([]DocumentInfo, error) {
	docs := make([]DocumentInfo, 0, len(e.index))
	for _, doc := range e.index {
		docs = append(docs, doc)
	}

	// Sort by category, then by name
	sort.Slice(docs, func(i, j int) bool {
		if docs[i].Category == docs[j].Category {
			return docs[i].Name < docs[j].Name
		}
		return docs[i].Category < docs[j].Category
	})

	return docs, nil
}

// SearchDocuments searches for documents containing the query
func (e *EmbeddedDocumentManager) SearchDocuments(query string) ([]SearchResult, error) {
	if query == "" {
		return nil, errors.New("search query cannot be empty")
	}

	// Input validation for security
	query = strings.TrimSpace(query)
	if len(query) > 500 {
		return nil, errors.New("search query too long (max 500 characters)")
	}

	query = strings.ToLower(query)
	var results []SearchResult

	for name, docInfo := range e.index {
		var matches []string
		score := 0

		// Search in document name
		if strings.Contains(strings.ToLower(name), query) {
			matches = append(matches, fmt.Sprintf("Name: %s", name))
			score += 10
		}

		// Search in description
		if strings.Contains(strings.ToLower(docInfo.Description), query) {
			matches = append(matches, fmt.Sprintf("Description: %s", docInfo.Description))
			score += 5
		}

		// Search in content
		content, err := e.GetDocument(name)
		if err == nil {
			contentStr := strings.ToLower(string(content))
			if strings.Contains(contentStr, query) {
				lines := strings.Split(string(content), "\n")
				for i, line := range lines {
					if strings.Contains(strings.ToLower(line), query) {
						match := fmt.Sprintf("Line %d: %s", i+1, strings.TrimSpace(line))
						if len(match) > 100 {
							match = match[:97] + "..."
						}
						matches = append(matches, match)
						score += 1
						if len(matches) >= 5 { // Limit matches per document
							break
						}
					}
				}
			}
		}

		if len(matches) > 0 {
			results = append(results, SearchResult{
				Document: docInfo,
				Matches:  matches,
				Score:    score,
			})
		}
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// IsAvailable returns true if documentation is available
func (e *EmbeddedDocumentManager) IsAvailable() bool {
	return len(e.index) > 0
}

// GetCategories returns all available document categories
func (e *EmbeddedDocumentManager) GetCategories() []string {
	categories := make(map[string]bool)
	for _, doc := range e.index {
		categories[doc.Category] = true
	}

	result := make([]string, 0, len(categories))
	for category := range categories {
		result = append(result, category)
	}

	sort.Strings(result)
	return result
}

// GetDocumentsByCategory returns documents in a specific category
func (e *EmbeddedDocumentManager) GetDocumentsByCategory(category string) ([]DocumentInfo, error) {
	var docs []DocumentInfo
	for _, doc := range e.index {
		if doc.Category == category {
			docs = append(docs, doc)
		}
	}

	// Sort by name
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Name < docs[j].Name
	})

	return docs, nil
}

// ExternalDocumentManager handles fallback to external documentation
type ExternalDocumentManager struct {
	baseURL string
	cache   map[string][]byte
}

// NewExternalDocumentManager creates a new external document manager
func NewExternalDocumentManager(baseURL string) *ExternalDocumentManager {
	return &ExternalDocumentManager{
		baseURL: baseURL,
		cache:   make(map[string][]byte),
	}
}

// GetDocument retrieves a document from external source
func (e *ExternalDocumentManager) GetDocument(name string) ([]byte, error) {
	return nil, fmt.Errorf("external documentation access not implemented yet")
}

// ListDocuments returns available external documents
func (e *ExternalDocumentManager) ListDocuments() ([]DocumentInfo, error) {
	return nil, fmt.Errorf("external documentation listing not implemented yet")
}

// SearchDocuments searches external documents
func (e *ExternalDocumentManager) SearchDocuments(query string) ([]SearchResult, error) {
	return nil, fmt.Errorf("external documentation search not implemented yet")
}

// IsAvailable returns false for external manager
func (e *ExternalDocumentManager) IsAvailable() bool {
	return false
}

// GetCategories returns empty for external manager
func (e *ExternalDocumentManager) GetCategories() []string {
	return nil
}

// GetDocumentsByCategory returns empty for external manager
func (e *ExternalDocumentManager) GetDocumentsByCategory(category string) ([]DocumentInfo, error) {
	return nil, nil
}

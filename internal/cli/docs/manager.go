// ABOUTME: Document management system for embedding help documentation
// ABOUTME: Provides interface for accessing embedded or external documentation

package docs

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"tamarou.com/pvm/internal/config"
	docsErrors "tamarou.com/pvm/internal/errors"
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
	githubClient *GitHubDocsClient
	cache        *DocsCache
	config       *config.DocsConfig
	index        map[string]DocumentInfo
	indexBuilt   bool
	fallbackMgr  DocumentManager
}

// NewExternalDocumentManager creates a new external document manager
func NewExternalDocumentManager(docsConfig *config.DocsConfig, fallbackManager DocumentManager) *ExternalDocumentManager {
	// Set up cache
	cacheTTL := time.Duration(docsConfig.CacheTTL) * time.Hour
	cache := NewDocsCache(docsConfig.CacheDir, cacheTTL)

	// Set up GitHub client
	githubClient := NewGitHubDocsClient(docsConfig)

	return &ExternalDocumentManager{
		githubClient: githubClient,
		cache:        cache,
		config:       docsConfig,
		index:        make(map[string]DocumentInfo),
		indexBuilt:   false,
		fallbackMgr:  fallbackManager,
	}
}

// GetDocument retrieves a document from external source
func (e *ExternalDocumentManager) GetDocument(name string) ([]byte, error) {
	// Handle offline mode
	if e.config.OfflineMode {
		return e.getFromFallback(name)
	}

	// Try to get from cache first
	var content []byte
	cacheKey := fmt.Sprintf("doc:%s", name)

	if e.cache.Get(cacheKey, &content) {
		return content, nil
	}

	// Build index if not built yet
	if err := e.ensureIndexBuilt(); err != nil {
		// If we can't build the index and have a fallback, use it
		if e.fallbackMgr != nil {
			return e.getFromFallback(name)
		}
		return nil, err
	}

	// Find document in index
	docInfo, exists := e.index[name]
	if !exists {
		// Try fallback if available
		if e.fallbackMgr != nil {
			return e.getFromFallback(name)
		}
		return nil, docsErrors.NewDocumentationError("004", fmt.Sprintf("Document not found: %s", name), nil)
	}

	// Fetch from GitHub
	ctx, cancel := context.WithTimeout(context.Background(), e.githubClient.timeout)
	defer cancel()

	owner, repo := e.parseRepository()
	cacheHeaders := e.cache.GetConditionalHeaders(cacheKey)

	rawContent, headers, notModified, err := e.githubClient.GetRawContentWithCache(
		ctx, owner, repo, e.config.Branch, docInfo.Path, cacheHeaders)

	if err != nil {
		// Try fallback on error
		if e.fallbackMgr != nil {
			return e.getFromFallback(name)
		}
		return nil, err
	}

	// If content was not modified, return cached version
	if notModified {
		if e.cache.Get(cacheKey, &content) {
			return content, nil
		}
	}

	// Cache the new content
	if err := e.cache.SetWithHTTPHeaders(cacheKey, rawContent, "github", headers, ""); err != nil {
		// Log error but continue - caching is not critical
		fmt.Printf("Warning: Failed to cache document %s: %v\n", name, err)
	}

	return rawContent, nil
}

// ListDocuments returns available external documents
func (e *ExternalDocumentManager) ListDocuments() ([]DocumentInfo, error) {
	// Handle offline mode
	if e.config.OfflineMode {
		if e.fallbackMgr != nil {
			return e.fallbackMgr.ListDocuments()
		}
		return nil, docsErrors.NewDocumentationError("002", "External documentation unavailable in offline mode", nil)
	}

	// Build index if not built yet
	if err := e.ensureIndexBuilt(); err != nil {
		if e.fallbackMgr != nil {
			return e.fallbackMgr.ListDocuments()
		}
		return nil, err
	}

	// Convert index to slice
	var docs []DocumentInfo
	for _, doc := range e.index {
		docs = append(docs, doc)
	}

	// Sort by name for consistent ordering
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Name < docs[j].Name
	})

	return docs, nil
}

// SearchDocuments searches external documents
func (e *ExternalDocumentManager) SearchDocuments(query string) ([]SearchResult, error) {
	// Handle offline mode
	if e.config.OfflineMode {
		if e.fallbackMgr != nil {
			return e.fallbackMgr.SearchDocuments(query)
		}
		return nil, docsErrors.NewDocumentationError("002", "External documentation search unavailable in offline mode", nil)
	}

	if query == "" {
		return nil, docsErrors.NewDocumentationError("002", "Search query cannot be empty", nil)
	}

	// Build index if not built yet
	if err := e.ensureIndexBuilt(); err != nil {
		if e.fallbackMgr != nil {
			return e.fallbackMgr.SearchDocuments(query)
		}
		return nil, err
	}

	var results []SearchResult
	queryLower := strings.ToLower(query)

	// Search through all documents
	for _, docInfo := range e.index {
		// First check if query matches document name or description
		score := 0
		var matches []string

		if strings.Contains(strings.ToLower(docInfo.Name), queryLower) {
			score += 10
			matches = append(matches, fmt.Sprintf("Title: %s", docInfo.Name))
		}

		if strings.Contains(strings.ToLower(docInfo.Description), queryLower) {
			score += 5
			matches = append(matches, fmt.Sprintf("Description: %s", docInfo.Description))
		}

		// Try to get content and search within it
		content, err := e.GetDocument(docInfo.Name)
		if err == nil {
			contentStr := strings.ToLower(string(content))
			if strings.Contains(contentStr, queryLower) {
				score += 1

				// Find context around matches
				lines := strings.Split(string(content), "\n")
				for i, line := range lines {
					if strings.Contains(strings.ToLower(line), queryLower) {
						contextStart := i - 1
						contextEnd := i + 1
						if contextStart < 0 {
							contextStart = 0
						}
						if contextEnd >= len(lines) {
							contextEnd = len(lines) - 1
						}

						var context []string
						for j := contextStart; j <= contextEnd; j++ {
							context = append(context, lines[j])
						}
						matches = append(matches, strings.Join(context, "\n"))

						// Limit number of matches per document
						if len(matches) >= 5 {
							break
						}
					}
				}
			}
		}

		if score > 0 {
			results = append(results, SearchResult{
				Document: docInfo,
				Matches:  matches,
				Score:    score,
			})
		}
	}

	// Sort by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// IsAvailable returns true if external documentation is available
func (e *ExternalDocumentManager) IsAvailable() bool {
	// In offline mode, only available if we have fallback
	if e.config.OfflineMode {
		return e.fallbackMgr != nil && e.fallbackMgr.IsAvailable()
	}

	// If we have a fallback manager, we're always available
	if e.fallbackMgr != nil && e.fallbackMgr.IsAvailable() {
		return true
	}

	// Try to build index as a connectivity test (non-blocking)
	if err := e.ensureIndexBuilt(); err != nil {
		return false
	}

	return len(e.index) > 0
}

// GetCategories returns available document categories
func (e *ExternalDocumentManager) GetCategories() []string {
	// Handle offline mode
	if e.config.OfflineMode {
		if e.fallbackMgr != nil {
			return e.fallbackMgr.GetCategories()
		}
		return []string{}
	}

	// Build index if not built yet
	if err := e.ensureIndexBuilt(); err != nil {
		if e.fallbackMgr != nil {
			return e.fallbackMgr.GetCategories()
		}
		return []string{}
	}

	// Collect unique categories
	categories := make(map[string]bool)
	for _, doc := range e.index {
		if doc.Category != "" {
			categories[doc.Category] = true
		}
	}

	// Convert to sorted slice
	var result []string
	for category := range categories {
		result = append(result, category)
	}
	sort.Strings(result)

	return result
}

// GetDocumentsByCategory returns documents in a specific category
func (e *ExternalDocumentManager) GetDocumentsByCategory(category string) ([]DocumentInfo, error) {
	// Handle offline mode
	if e.config.OfflineMode {
		if e.fallbackMgr != nil {
			return e.fallbackMgr.GetDocumentsByCategory(category)
		}
		return nil, docsErrors.NewDocumentationError("002", "External documentation unavailable in offline mode", nil)
	}

	// Build index if not built yet
	if err := e.ensureIndexBuilt(); err != nil {
		if e.fallbackMgr != nil {
			return e.fallbackMgr.GetDocumentsByCategory(category)
		}
		return nil, err
	}

	// Filter documents by category
	var docs []DocumentInfo
	for _, doc := range e.index {
		if doc.Category == category {
			docs = append(docs, doc)
		}
	}

	// Sort by name for consistent ordering
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Name < docs[j].Name
	})

	return docs, nil
}

// Helper methods for ExternalDocumentManager

// ensureIndexBuilt builds the document index if not already built
func (e *ExternalDocumentManager) ensureIndexBuilt() error {
	if e.indexBuilt {
		return nil
	}

	// Check cache for index first
	cacheKey := "index:documents"
	var cachedIndex map[string]DocumentInfo
	if e.cache.Get(cacheKey, &cachedIndex) {
		e.index = cachedIndex
		e.indexBuilt = true
		return nil
	}

	// Build index from GitHub
	ctx, cancel := context.WithTimeout(context.Background(), e.githubClient.timeout)
	defer cancel()

	owner, repo := e.parseRepository()

	// List all documentation files
	files, err := e.githubClient.ListDocumentationFiles(ctx, owner, repo, e.config.Branch, "docs")
	if err != nil {
		// Try root directory if docs directory doesn't exist
		files, err = e.githubClient.ListDocumentationFiles(ctx, owner, repo, e.config.Branch, "")
		if err != nil {
			return docsErrors.NewDocumentationError("001", "Failed to list documentation files", err)
		}
	}

	// Build document info for each file
	for _, filePath := range files {
		name := e.extractDocumentName(filePath)
		category := e.categorizeDocument(filePath, name)
		description := e.generateDescription(name)

		e.index[name] = DocumentInfo{
			Name:        name,
			Path:        filePath,
			Size:        0, // We don't have size info from directory listing
			Category:    category,
			Description: description,
			ModTime:     time.Now(), // Use current time as we don't have exact mod time
		}
	}

	// Cache the index
	if err := e.cache.Set(cacheKey, e.index, "github"); err != nil {
		// Log warning but continue - caching is not critical
		fmt.Printf("Warning: Failed to cache document index: %v\n", err)
	}

	e.indexBuilt = true
	return nil
}

// getFromFallback attempts to get document from fallback manager
func (e *ExternalDocumentManager) getFromFallback(name string) ([]byte, error) {
	if e.fallbackMgr == nil {
		return nil, docsErrors.NewDocumentationError("004", fmt.Sprintf("Document not found and no fallback available: %s", name), nil)
	}

	return e.fallbackMgr.GetDocument(name)
}

// parseRepository parses the repository configuration into owner/repo
func (e *ExternalDocumentManager) parseRepository() (string, string) {
	parts := strings.Split(e.config.Repository, "/")
	if len(parts) != 2 {
		return "perigrin", "pvm" // Default fallback
	}
	return parts[0], parts[1]
}

// extractDocumentName extracts document name from file path
func (e *ExternalDocumentManager) extractDocumentName(filePath string) string {
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, filepath.Ext(base))

	// Handle special cases
	if name == "README" {
		return "readme"
	}
	if name == "CHANGELOG" {
		return "changelog"
	}

	return strings.ToLower(strings.ReplaceAll(name, "-", "_"))
}

// categorizeDocument determines the category for a document
func (e *ExternalDocumentManager) categorizeDocument(path, name string) string {
	pathLower := strings.ToLower(path)
	nameLower := strings.ToLower(name)

	// Check path-based categories
	if strings.Contains(pathLower, "/api/") || strings.Contains(pathLower, "/reference/") {
		return "Reference"
	}
	if strings.Contains(pathLower, "/tutorial/") || strings.Contains(pathLower, "/guide/") {
		return "Getting Started"
	}
	if strings.Contains(pathLower, "/dev/") || strings.Contains(pathLower, "/development/") {
		return "Development"
	}
	if strings.Contains(pathLower, "/example") || strings.Contains(pathLower, "/examples/") {
		return "Examples"
	}

	// Check name-based categories
	if strings.Contains(nameLower, "install") || strings.Contains(nameLower, "setup") {
		return "Getting Started"
	}
	if strings.Contains(nameLower, "api") || strings.Contains(nameLower, "reference") {
		return "Reference"
	}
	if strings.Contains(nameLower, "tutorial") || strings.Contains(nameLower, "guide") {
		return "Getting Started"
	}
	if strings.Contains(nameLower, "dev") || strings.Contains(nameLower, "contrib") {
		return "Development"
	}
	if strings.Contains(nameLower, "example") || strings.Contains(nameLower, "sample") {
		return "Examples"
	}
	if strings.Contains(nameLower, "config") || strings.Contains(nameLower, "setting") {
		return "Configuration"
	}

	// Default category
	return "Documentation"
}

// generateDescription generates a description for a document
func (e *ExternalDocumentManager) generateDescription(name string) string {
	switch name {
	case "readme":
		return "Project overview and basic information"
	case "changelog":
		return "Version history and release notes"
	case "install":
		return "Installation instructions"
	case "setup":
		return "Setup and configuration guide"
	case "api":
		return "API reference documentation"
	case "tutorial":
		return "Step-by-step tutorial"
	case "guide":
		return "User guide and documentation"
	case "examples":
		return "Code examples and samples"
	case "config":
		return "Configuration options and settings"
	case "troubleshooting":
		return "Common problems and solutions"
	case "faq":
		return "Frequently asked questions"
	default:
		// Generate description from name
		words := strings.Fields(strings.ReplaceAll(name, "_", " "))
		if len(words) > 0 {
			// Capitalize first word
			words[0] = strings.Title(words[0])
			return strings.Join(words, " ") + " documentation"
		}
		return "Documentation"
	}
}

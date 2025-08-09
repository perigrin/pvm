// ABOUTME: Comprehensive tests for DocumentManager and embedded documentation
// ABOUTME: Tests both embedded and external document management functionality

package docs

import (
	"embed"
	"os"
	"testing"
	"time"

	"tamarou.com/pvm/internal/config"
)

//go:embed testdata/*.md
var testDocsFS embed.FS

func TestEmbeddedDocumentManager_NewEmbeddedDocumentManager(t *testing.T) {
	manager, err := NewEmbeddedDocumentManager(testDocsFS)
	if err != nil {
		t.Fatalf("Failed to create embedded document manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	if !manager.IsAvailable() {
		t.Error("Expected manager to be available")
	}
}

func TestEmbeddedDocumentManager_GetDocument(t *testing.T) {
	manager, err := NewEmbeddedDocumentManager(testDocsFS)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test getting an existing document
	content, err := manager.GetDocument("test-guide")
	if err != nil {
		t.Errorf("Failed to get document: %v", err)
	}

	if len(content) == 0 {
		t.Error("Expected non-empty content")
	}

	expectedContent := "# Test Guide"
	if string(content)[:len(expectedContent)] != expectedContent {
		t.Errorf("Expected content to start with '%s', got '%s'", expectedContent, string(content)[:len(expectedContent)])
	}

	// Test getting non-existent document
	_, err = manager.GetDocument("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent document")
	}
}

func TestEmbeddedDocumentManager_ListDocuments(t *testing.T) {
	manager, err := NewEmbeddedDocumentManager(testDocsFS)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	docs, err := manager.ListDocuments()
	if err != nil {
		t.Errorf("Failed to list documents: %v", err)
	}

	if len(docs) == 0 {
		t.Error("Expected at least one document")
	}

	// Check that documents are sorted
	for i := 1; i < len(docs); i++ {
		if docs[i-1].Category > docs[i].Category {
			t.Errorf("Documents not sorted by category")
			break
		}
		if docs[i-1].Category == docs[i].Category && docs[i-1].Name > docs[i].Name {
			t.Errorf("Documents not sorted by name within category")
			break
		}
	}
}

func TestEmbeddedDocumentManager_SearchDocuments(t *testing.T) {
	manager, err := NewEmbeddedDocumentManager(testDocsFS)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test valid search
	results, err := manager.SearchDocuments("guide")
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected search results")
	}

	// Check that results are sorted by score
	for i := 1; i < len(results); i++ {
		if results[i-1].Score < results[i].Score {
			t.Errorf("Search results not sorted by score")
			break
		}
	}

	// Test empty search
	_, err = manager.SearchDocuments("")
	if err == nil {
		t.Error("Expected error for empty search query")
	}

	// Test no results
	results, err = manager.SearchDocuments("nonexistentterm12345")
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	if len(results) != 0 {
		t.Error("Expected no results for non-existent term")
	}
}

func TestEmbeddedDocumentManager_GetCategories(t *testing.T) {
	manager, err := NewEmbeddedDocumentManager(testDocsFS)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	categories := manager.GetCategories()
	if len(categories) == 0 {
		t.Error("Expected at least one category")
	}

	// Check that categories are sorted
	for i := 1; i < len(categories); i++ {
		if categories[i-1] > categories[i] {
			t.Errorf("Categories not sorted")
			break
		}
	}
}

func TestEmbeddedDocumentManager_GetDocumentsByCategory(t *testing.T) {
	manager, err := NewEmbeddedDocumentManager(testDocsFS)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	categories := manager.GetCategories()
	if len(categories) == 0 {
		t.Skip("No categories available for testing")
	}

	category := categories[0]
	docs, err := manager.GetDocumentsByCategory(category)
	if err != nil {
		t.Errorf("Failed to get documents by category: %v", err)
	}

	for _, doc := range docs {
		if doc.Category != category {
			t.Errorf("Document %s has category %s, expected %s", doc.Name, doc.Category, category)
		}
	}

	// Test non-existent category
	docs, err = manager.GetDocumentsByCategory("NonExistentCategory")
	if err != nil {
		t.Errorf("Failed to get documents for non-existent category: %v", err)
	}

	if len(docs) != 0 {
		t.Error("Expected no documents for non-existent category")
	}
}

func TestEmbeddedDocumentManager_CategorizeDocument(t *testing.T) {
	manager := &EmbeddedDocumentManager{}

	tests := []struct {
		path     string
		name     string
		expected string
	}{
		{"quickstart.md", "quickstart", "Getting Started"},
		{"workflow-new-project.md", "workflow-new-project", "Workflows"},
		{"command-reference.md", "command-reference", "Reference"},
		{"developer-guide.md", "developer-guide", "Development"},
		{"architecture-overview.md", "architecture-overview", "Architecture"},
		{"troubleshooting.md", "troubleshooting", "Support"},
		{"archive/old-doc.md", "old-doc", "Archive"},
		{"some-guide.md", "some-guide", "Guides"},
		{"type-specification.md", "type-specification", "Specifications"},
		{"random-doc.md", "random-doc", "Documentation"},
	}

	for _, test := range tests {
		result := manager.categorizeDocument(test.path, test.name)
		if result != test.expected {
			t.Errorf("categorizeDocument(%s, %s) = %s, expected %s", test.path, test.name, result, test.expected)
		}
	}
}

func TestEmbeddedDocumentManager_GenerateDescription(t *testing.T) {
	manager := &EmbeddedDocumentManager{}

	tests := []struct {
		name     string
		expected string
	}{
		{"quickstart", "Quick start guide for new users"},
		{"command-reference", "Complete command reference"},
		{"workflow-new-development", "Workflow for new project development"},
		{"custom-name", "Custom Name"},
		{"hyphen-separated-name", "Hyphen Separated Name"},
		{"underscore_separated_name", "Underscore Separated Name"},
	}

	for _, test := range tests {
		result := manager.generateDescription(test.name)
		if result != test.expected {
			t.Errorf("generateDescription(%s) = %s, expected %s", test.name, result, test.expected)
		}
	}
}

func TestExternalDocumentManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "external_docs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	docsConfig := &config.DocsConfig{
		Repository: "test/repo",
		Branch:     "main",
		CacheDir:   tmpDir,
		CacheTTL:   1,
		MaxRetries: 1,
		Timeout:    "1ms", // Very short timeout to ensure failure
	}
	manager := NewExternalDocumentManager(docsConfig, nil)

	// With very short timeout, external manager should not be available
	available := manager.IsAvailable()
	if available {
		t.Error("External manager should not be available with short timeout and no fallback")
	}

	// Test that all methods return appropriate errors/empty results
	_, err = manager.GetDocument("test")
	if err == nil {
		t.Error("Expected error from external manager GetDocument")
	}

	_, err = manager.ListDocuments()
	if err == nil {
		t.Error("Expected error from external manager ListDocuments")
	}

	_, err = manager.SearchDocuments("test")
	if err == nil {
		t.Error("Expected error from external manager SearchDocuments")
	}

	categories := manager.GetCategories()
	if len(categories) != 0 {
		t.Error("Expected empty categories from external manager")
	}

	_, err = manager.GetDocumentsByCategory("test")
	if err == nil {
		t.Error("Expected error from external manager GetDocumentsByCategory when no fallback available")
	}
}

func TestDocumentInfo(t *testing.T) {
	info := DocumentInfo{
		Name:        "test",
		Path:        "test.md",
		Size:        1024,
		Category:    "Test",
		Description: "Test document",
		ModTime:     time.Now(),
	}

	if info.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", info.Name)
	}

	if info.Size != 1024 {
		t.Errorf("Expected size 1024, got %d", info.Size)
	}
}

func TestSearchResult(t *testing.T) {
	info := DocumentInfo{Name: "test", Category: "Test"}
	result := SearchResult{
		Document: info,
		Matches:  []string{"match1", "match2"},
		Score:    10,
	}

	if len(result.Matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(result.Matches))
	}

	if result.Score != 10 {
		t.Errorf("Expected score 10, got %d", result.Score)
	}
}

// ABOUTME: Tests for search tool implementation
// ABOUTME: Covers similarity search, type signature search, and pattern search functionality

package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	chromem "github.com/philippgille/chromem-go"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/mcp/embeddings"
)

func TestCodeSearcher_SimilaritySearch(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test_project")

	// Create test project structure
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create test project directory: %v", err)
	}

	// Create test Perl file
	testFile := filepath.Join(projectPath, "test.pl")
	testCode := `use v5.40;
use experimental 'class';

class Calculator {
    field Int $value = 0;

    method add(Int $x) : Int {
        $value += $x;
        return $value;
    }

    method multiply(Int $x) : Int {
        $value *= $x;
        return $value;
    }
}

sub factorial(Int $n) : Int {
    return 1 if $n <= 1;
    return $n * factorial($n - 1);
}
`
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create searcher
	searcher := createTestSearcher(t, tempDir)

	// Index the test project
	if err := indexTestProject(t, searcher, projectPath); err != nil {
		t.Fatalf("Failed to index test project: %v", err)
	}

	// Test similarity search
	request := SearchRequest{
		Query:         "calculator add method",
		Method:        "similarity",
		ProjectPath:   projectPath,
		TopK:          10,
		MinSimilarity: 0.1,
	}

	response, err := searcher.Search(context.Background(), request)
	if err != nil {
		t.Fatalf("Similarity search failed: %v", err)
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got %s", response.Status)
	}

	if len(response.Results) == 0 {
		t.Error("Expected at least one result from similarity search")
	}

	// Verify result structure
	for _, result := range response.Results {
		if result.SearchMethod != "similarity" {
			t.Errorf("Expected search method 'similarity', got %s", result.SearchMethod)
		}
		if result.Similarity < 0.0 || result.Similarity > 1.0 {
			t.Errorf("Invalid similarity score: %f", result.Similarity)
		}
		if result.Score < 0.0 || result.Score > 1.0 {
			t.Errorf("Invalid score: %f", result.Score)
		}
	}
}

func TestCodeSearcher_TypeSignatureSearch(t *testing.T) {
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test_project")

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create test project directory: %v", err)
	}

	testFile := filepath.Join(projectPath, "types.pl")
	testCode := `use v5.40;

sub process_string(Str $input) : Str {
    return uc($input);
}

sub calculate_sum(Int $a, Int $b) : Int {
    return $a + $b;
}

sub get_items() : ArrayRef[Str] {
    return ["item1", "item2", "item3"];
}
`
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	searcher := createTestSearcher(t, tempDir)
	if err := indexTestProject(t, searcher, projectPath); err != nil {
		t.Fatalf("Failed to index test project: %v", err)
	}

	// Test type signature search
	request := SearchRequest{
		Query:         "Int -> Int",
		Method:        "type_signature",
		ProjectPath:   projectPath,
		TopK:          10,
		MinSimilarity: 0.1,
	}

	response, err := searcher.Search(context.Background(), request)
	if err != nil {
		t.Fatalf("Type signature search failed: %v", err)
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got %s", response.Status)
	}

	// Verify results have type signature method
	for _, result := range response.Results {
		if result.SearchMethod != "type_signature" {
			t.Errorf("Expected search method 'type_signature', got %s", result.SearchMethod)
		}
	}
}

func TestCodeSearcher_PatternSearch(t *testing.T) {
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test_project")

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create test project directory: %v", err)
	}

	testFile := filepath.Join(projectPath, "patterns.pl")
	testCode := `use v5.40;
use MyModule;

package MyClass;

sub new {
    my $class = shift;
    return bless {}, $class;
}

sub process_data {
    my ($self, $data) = @_;
    return $data;
}

1;
`
	if err := os.WriteFile(testFile, []byte(testCode), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	searcher := createTestSearcher(t, tempDir)
	if err := indexTestProject(t, searcher, projectPath); err != nil {
		t.Fatalf("Failed to index test project: %v", err)
	}

	// Test pattern search
	request := SearchRequest{
		Query:         "sub process_data",
		Method:        "pattern",
		ProjectPath:   projectPath,
		TopK:          10,
		MinSimilarity: 0.1,
	}

	response, err := searcher.Search(context.Background(), request)
	if err != nil {
		t.Fatalf("Pattern search failed: %v", err)
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got %s", response.Status)
	}

	// Verify results have pattern method
	for _, result := range response.Results {
		if result.SearchMethod != "pattern" {
			t.Errorf("Expected search method 'pattern', got %s", result.SearchMethod)
		}
	}
}

func TestCodeSearcher_WithFilters(t *testing.T) {
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test_project")

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create test project directory: %v", err)
	}

	// Create multiple test files
	files := map[string]string{
		"lib/MyModule.pm": `package MyModule;
sub helper_function { return "helper"; }
1;`,
		"script/main.pl": `use MyModule;
sub main_function { print "main"; }`,
		"test/test.t": `use Test2::V0;
sub test_function { ok(1, "test"); }
done_testing;`,
	}

	for filePath, content := range files {
		fullPath := filepath.Join(projectPath, filePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", filePath, err)
		}
	}

	searcher := createTestSearcher(t, tempDir)
	if err := indexTestProject(t, searcher, projectPath); err != nil {
		t.Fatalf("Failed to index test project: %v", err)
	}

	// Test search with file extension filter
	request := SearchRequest{
		Query:       "function",
		Method:      "similarity",
		ProjectPath: projectPath,
		TopK:        10,
		Filters: map[string]string{
			"file_extension": ".pm",
		},
		MinSimilarity: 0.1,
	}

	response, err := searcher.Search(context.Background(), request)
	if err != nil {
		t.Fatalf("Filtered search failed: %v", err)
	}

	// Verify all results are from .pm files
	for _, result := range response.Results {
		matched, err := filepath.Match("*.pm", filepath.Base(result.FilePath))
		if err != nil {
			t.Errorf("Pattern matching error: %v", err)
		}
		if !matched {
			t.Errorf("Result file path %s does not match filter .pm", result.FilePath)
		}
	}
}

func TestCodeSearcher_InvalidMethod(t *testing.T) {
	tempDir := t.TempDir()
	searcher := createTestSearcher(t, tempDir)

	request := SearchRequest{
		Query:       "test",
		Method:      "invalid_method",
		ProjectPath: tempDir,
		TopK:        10,
	}

	_, err := searcher.Search(context.Background(), request)
	if err == nil {
		t.Error("Expected error for invalid search method")
	}
}

func TestCodeSearcher_EmptyProject(t *testing.T) {
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "empty_project")

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create empty project directory: %v", err)
	}

	searcher := createTestSearcher(t, tempDir)

	request := SearchRequest{
		Query:         "anything",
		Method:        "similarity",
		ProjectPath:   projectPath,
		TopK:          10,
		MinSimilarity: 0.1,
	}

	response, err := searcher.Search(context.Background(), request)
	if err != nil {
		t.Fatalf("Search in empty project failed: %v", err)
	}

	if len(response.Results) != 0 {
		t.Errorf("Expected no results for empty project, got %d", len(response.Results))
	}
}

// Helper functions

func createTestSearcher(t *testing.T, tempDir string) *CodeSearcher {
	logger := log.NewLogger(log.LevelError, os.Stderr, "test")

	// Create local embedding provider for testing
	provider, err := embeddings.NewLocalProvider(embeddings.EmbeddingConfig{
		Provider:   "local",
		Dimensions: 384,
		MaxTokens:  512,
	})
	if err != nil {
		t.Fatalf("Failed to create embedding provider: %v", err)
	}

	// Create embedding store
	store, err := embeddings.NewEmbeddingStore(embeddings.StoreConfig{
		PersistPath: filepath.Join(tempDir, "embeddings"),
		Provider:    provider,
		Logger:      logger,
	})
	if err != nil {
		t.Fatalf("Failed to create embedding store: %v", err)
	}

	// Create collection manager
	manager := embeddings.NewCollectionManager(store, logger)

	// Create code extractor
	extractor, err := embeddings.NewExtractor()
	if err != nil {
		t.Fatalf("Failed to create code extractor: %v", err)
	}

	return NewCodeSearcher(store, manager, extractor, logger)
}

func indexTestProject(t *testing.T, searcher *CodeSearcher, projectPath string) error {
	// Find Perl files in the project
	filePaths, err := findPerlFiles(projectPath)
	if err != nil {
		return err
	}

	if len(filePaths) == 0 {
		// No files to index, but this is not an error for empty projects
		return nil
	}

	// Extract documents from the files
	docs, err := embeddings.BatchExtractAndConvert(searcher.extractor, projectPath, filePaths)
	if err != nil {
		return err
	}

	// Convert to store documents
	storeDocs := make([]embeddings.Document, len(docs))
	for i, doc := range docs {
		// Convert metadata from map[string]string to map[string]any
		metadata := make(map[string]any)
		for k, v := range doc.Metadata {
			metadata[k] = v
		}

		storeDocs[i] = embeddings.Document{
			ID:       doc.ID,
			Content:  doc.Content,
			Metadata: metadata,
		}
	}

	// Get or create collection using the same method as search
	collection, err := searcher.manager.GetOrCreateCollection(context.Background(), projectPath)
	if err != nil {
		return err
	}

	// Convert to chromem documents
	chromemDocs := make([]chromem.Document, len(storeDocs))
	for i, doc := range storeDocs {
		// Convert metadata from map[string]any to map[string]string
		metadata := make(map[string]string)
		for k, v := range doc.Metadata {
			metadata[k] = fmt.Sprintf("%v", v)
		}

		chromemDocs[i] = chromem.Document{
			ID:       doc.ID,
			Content:  doc.Content,
			Metadata: metadata,
		}
	}

	// Add documents using collection manager
	err = searcher.manager.AddDocuments(context.Background(), collection, chromemDocs)
	if err != nil {
		return err
	}

	return nil
}

// findPerlFiles recursively finds Perl files in a directory
func findPerlFiles(root string) ([]string, error) {
	var perlFiles []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check for Perl file extensions
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".pl" || ext == ".pm" || ext == ".t" {
			perlFiles = append(perlFiles, path)
		}

		return nil
	})

	return perlFiles, err
}

func TestExtractTypeSignature(t *testing.T) {
	searcher := createTestSearcher(t, t.TempDir())

	tests := []struct {
		query    string
		expected []string
	}{
		{
			query:    "Int -> Str",
			expected: []string{"Int", "Str"},
		},
		{
			query:    "ArrayRef[Int] $items",
			expected: []string{"ArrayRef", "Int", "$items"},
		},
		{
			query:    "sub process(MyType $input) : Bool",
			expected: []string{"MyType", "$input", "Bool"},
		},
		{
			query:    "simple text with no types",
			expected: []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			result := searcher.extractTypeSignature(test.query)

			// Compare lengths first
			if len(result) != len(test.expected) {
				t.Errorf("Expected %d types, got %d: %v", len(test.expected), len(result), result)
				return
			}

			// Check that all expected types are present (order doesn't matter)
			expectedMap := make(map[string]bool)
			for _, expected := range test.expected {
				expectedMap[expected] = true
			}

			for _, found := range result {
				if !expectedMap[found] {
					t.Errorf("Unexpected type found: %s", found)
				}
			}
		})
	}
}

func TestScorePatternMatch(t *testing.T) {
	searcher := createTestSearcher(t, t.TempDir())

	result := SearchResult{
		Content: "sub process_data { my $data = shift; return $data; }",
	}

	tests := []struct {
		query         string
		expectedScore float32
		expectNonZero bool
	}{
		{
			query:         "process_data",
			expectNonZero: true,
		},
		{
			query:         "nonexistent_function",
			expectedScore: 0.0,
		},
		{
			query:         "sub",
			expectNonZero: true,
		},
	}

	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			score := searcher.scorePatternMatch(result, test.query, nil)

			if test.expectNonZero {
				if score <= 0.0 {
					t.Errorf("Expected non-zero score for query '%s', got %f", test.query, score)
				}
			} else {
				if score != test.expectedScore {
					t.Errorf("Expected score %f for query '%s', got %f", test.expectedScore, test.query, score)
				}
			}
		})
	}
}

func TestSearchRequestDefaults(t *testing.T) {
	tempDir := t.TempDir()
	projectPath := filepath.Join(tempDir, "test_project")

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create test project directory: %v", err)
	}

	searcher := createTestSearcher(t, tempDir)

	// Test with minimal request (should use defaults)
	request := SearchRequest{
		Query:       "test",
		Method:      "similarity",
		ProjectPath: projectPath,
		// TopK and MinSimilarity not set - should use defaults
	}

	response, err := searcher.Search(context.Background(), request)
	if err != nil {
		t.Fatalf("Search with defaults failed: %v", err)
	}

	if response.Status != "success" {
		t.Errorf("Expected status 'success', got %s", response.Status)
	}

	// Verify defaults were applied by checking internal behavior
	// (We can't directly check the values, but we can verify the search worked)
}

func TestSearchTimeout(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	tempDir := t.TempDir()
	searcher := createTestSearcher(t, tempDir)

	request := SearchRequest{
		Query:       "test",
		Method:      "similarity",
		ProjectPath: tempDir,
		TopK:        10,
	}

	// This should complete quickly enough, but if it doesn't, it should handle the timeout gracefully
	_, err := searcher.Search(ctx, request)
	if err != nil && ctx.Err() == context.DeadlineExceeded {
		t.Log("Search properly handled context timeout")
	}
}

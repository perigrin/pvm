// ABOUTME: Comprehensive tests for external documentation manager functionality
// ABOUTME: Tests cover configuration, caching, GitHub client, fallback strategies, and error conditions

package docs

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tamarou.com/pvm/internal/config"
)

// testFallbackManager implements DocumentManager for testing fallback scenarios
type testFallbackManager struct {
	documents map[string][]byte
	available bool
}

func newTestFallbackManager() *testFallbackManager {
	return &testFallbackManager{
		documents: make(map[string][]byte),
		available: true,
	}
}

func (m *testFallbackManager) GetDocument(name string) ([]byte, error) {
	if !m.available {
		return nil, &DocumentNotFoundError{Name: name}
	}
	if content, exists := m.documents[name]; exists {
		return content, nil
	}
	return nil, &DocumentNotFoundError{Name: name}
}

func (m *testFallbackManager) ListDocuments() ([]DocumentInfo, error) {
	if !m.available {
		return nil, &DocumentManagerUnavailableError{}
	}
	var docs []DocumentInfo
	for name, content := range m.documents {
		docs = append(docs, DocumentInfo{
			Name: name,
			Size: int64(len(content)),
		})
	}
	return docs, nil
}

func (m *testFallbackManager) SearchDocuments(query string) ([]SearchResult, error) {
	if !m.available {
		return nil, &DocumentManagerUnavailableError{}
	}
	return []SearchResult{}, nil
}

func (m *testFallbackManager) IsAvailable() bool {
	return m.available
}

func (m *testFallbackManager) GetCategories() []string {
	return []string{"Test"}
}

func (m *testFallbackManager) GetDocumentsByCategory(category string) ([]DocumentInfo, error) {
	return m.ListDocuments()
}

func (m *testFallbackManager) AddDocument(name string, content []byte) {
	m.documents[name] = content
}

func (m *testFallbackManager) SetAvailable(available bool) {
	m.available = available
}

// DocumentNotFoundError represents a document not found error
type DocumentNotFoundError struct {
	Name string
}

func (e *DocumentNotFoundError) Error() string {
	return "document not found: " + e.Name
}

// DocumentManagerUnavailableError represents an unavailable document manager
type DocumentManagerUnavailableError struct{}

func (e *DocumentManagerUnavailableError) Error() string {
	return "document manager unavailable"
}

// Test Configuration System
func TestDocsConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.DocsConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &config.DocsConfig{
				Repository:     "owner/repo",
				Branch:         "main",
				CacheTTL:       24,
				MaxRetries:     3,
				Timeout:        "30s",
				UpdateInterval: "1h",
			},
			wantErr: false,
		},
		{
			name: "empty repository",
			config: &config.DocsConfig{
				Repository: "",
				Branch:     "main",
			},
			wantErr: true,
		},
		{
			name: "invalid repository format",
			config: &config.DocsConfig{
				Repository: "invalid-repo",
				Branch:     "main",
			},
			wantErr: true,
		},
		{
			name: "invalid timeout",
			config: &config.DocsConfig{
				Repository: "owner/repo",
				Branch:     "main",
				Timeout:    "invalid",
			},
			wantErr: true,
		},
		{
			name: "negative cache TTL",
			config: &config.DocsConfig{
				Repository: "owner/repo",
				Branch:     "main",
				CacheTTL:   -2,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.config.Validate()
			hasErr := len(errors) > 0
			if hasErr != tt.wantErr {
				t.Errorf("DocsConfig.Validate() error = %v, wantErr %v", errors, tt.wantErr)
			}
		})
	}
}

// Test Documentation Cache
func TestDocsCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "docs_cache_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cache := NewDocsCache(tmpDir, time.Hour)

	t.Run("Set and Get", func(t *testing.T) {
		testData := map[string]interface{}{"content": "test document content"}

		err := cache.Set("test-doc", testData, "test")
		if err != nil {
			t.Fatalf("Cache.Set() error = %v", err)
		}

		var retrieved map[string]interface{}
		found := cache.Get("test-doc", &retrieved)
		if !found {
			t.Error("Cache.Get() returned false, expected true")
		}

		if retrieved["content"] != testData["content"] {
			t.Errorf("Cache.Get() = %v, want %v", retrieved, testData)
		}
	})

	t.Run("Cache Miss", func(t *testing.T) {
		var retrieved map[string]interface{}
		found := cache.Get("nonexistent", &retrieved)
		if found {
			t.Error("Cache.Get() returned true for nonexistent key, expected false")
		}
	})

	t.Run("Cache with HTTP Headers", func(t *testing.T) {
		headers := make(http.Header)
		headers.Set("ETag", "test-etag")
		headers.Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 MST")

		testData := "test content with headers"
		err := cache.SetWithHTTPHeaders("test-headers", testData, "test", headers, "http://example.com/test")
		if err != nil {
			t.Fatalf("Cache.SetWithHTTPHeaders() error = %v", err)
		}

		conditionalHeaders := cache.GetConditionalHeaders("test-headers")
		if conditionalHeaders.Get("If-None-Match") != "test-etag" {
			t.Errorf("Expected If-None-Match header to be 'test-etag', got %s", conditionalHeaders.Get("If-None-Match"))
		}
	})

	t.Run("Cache Expiry", func(t *testing.T) {
		shortCache := NewDocsCache(tmpDir+"/short", time.Millisecond)

		err := shortCache.Set("expire-test", "content", "test")
		if err != nil {
			t.Fatalf("Cache.Set() error = %v", err)
		}

		// Wait for expiry
		time.Sleep(2 * time.Millisecond)

		var retrieved string
		found := shortCache.Get("expire-test", &retrieved)
		if found {
			t.Error("Cache.Get() returned true for expired key, expected false")
		}
	})

	t.Run("Cache Statistics", func(t *testing.T) {
		stats, err := cache.Stats()
		if err != nil {
			t.Fatalf("Cache.Stats() error = %v", err)
		}

		if _, exists := stats["total_entries"]; !exists {
			t.Error("Expected total_entries in stats")
		}
	})

	t.Run("List Keys", func(t *testing.T) {
		keys, err := cache.ListKeys()
		if err != nil {
			t.Fatalf("Cache.ListKeys() error = %v", err)
		}

		if len(keys) == 0 {
			t.Error("Expected at least one key from ListKeys()")
		}
	})
}

// Test ExternalDocumentManager
func TestExternalDocumentManagerNew(t *testing.T) {
	// Set up test configuration
	tmpDir, err := os.MkdirTemp("", "external_docs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	docsConfig := &config.DocsConfig{
		Repository:     "perigrin/pvm",
		Branch:         "pu",
		CacheDir:       tmpDir,
		CacheTTL:       24,
		MaxRetries:     1,
		Timeout:        "5s",
		OfflineMode:    false,
		PreferEmbedded: false,
	}

	// Set up mock fallback manager
	fallbackMgr := newTestFallbackManager()
	fallbackMgr.AddDocument("test-doc", []byte("fallback content"))
	fallbackMgr.AddDocument("readme", []byte("# Test README\nThis is a test readme."))

	// Create external manager
	extManager := NewExternalDocumentManager(docsConfig, fallbackMgr)

	t.Run("Manager Creation", func(t *testing.T) {
		if extManager == nil {
			t.Error("NewExternalDocumentManager() returned nil")
		}
	})

	t.Run("Parse Repository", func(t *testing.T) {
		owner, repo := extManager.parseRepository()
		if owner != "perigrin" || repo != "pvm" {
			t.Errorf("parseRepository() = %s/%s, want perigrin/pvm", owner, repo)
		}
	})

	t.Run("Extract Document Name", func(t *testing.T) {
		tests := []struct {
			path string
			want string
		}{
			{"docs/README.md", "readme"},
			{"guides/installation-guide.md", "installation_guide"},
			{"api/reference.md", "reference"},
			{"CHANGELOG.md", "changelog"},
		}

		for _, tt := range tests {
			got := extManager.extractDocumentName(tt.path)
			if got != tt.want {
				t.Errorf("extractDocumentName(%s) = %s, want %s", tt.path, got, tt.want)
			}
		}
	})

	t.Run("Categorize Document", func(t *testing.T) {
		tests := []struct {
			path string
			name string
			want string
		}{
			{"docs/api/reference.md", "reference", "Reference"},
			{"docs/tutorial/getting-started.md", "getting_started", "Getting Started"},
			{"docs/dev/contributing.md", "contributing", "Development"},
			{"docs/examples/basic.md", "basic", "Examples"},
			{"docs/random.md", "random", "Documentation"},
		}

		for _, tt := range tests {
			got := extManager.categorizeDocument(tt.path, tt.name)
			if got != tt.want {
				t.Errorf("categorizeDocument(%s, %s) = %s, want %s", tt.path, tt.name, got, tt.want)
			}
		}
	})

	t.Run("Generate Description", func(t *testing.T) {
		tests := []struct {
			name string
			want string
		}{
			{"readme", "Project overview and basic information"},
			{"install", "Installation instructions"},
			{"api", "API reference documentation"},
			{"custom_doc", "Custom doc documentation"},
		}

		for _, tt := range tests {
			got := extManager.generateDescription(tt.name)
			if got != tt.want {
				t.Errorf("generateDescription(%s) = %s, want %s", tt.name, got, tt.want)
			}
		}
	})

	t.Run("Offline Mode", func(t *testing.T) {
		// Set offline mode
		extManager.config.OfflineMode = true

		// Try to get document (should use fallback)
		content, err := extManager.GetDocument("test-doc")
		if err != nil {
			t.Fatalf("GetDocument() in offline mode error = %v", err)
		}

		expected := "fallback content"
		if string(content) != expected {
			t.Errorf("GetDocument() in offline mode = %s, want %s", string(content), expected)
		}

		// Test list documents in offline mode
		docs, err := extManager.ListDocuments()
		if err != nil {
			t.Fatalf("ListDocuments() in offline mode error = %v", err)
		}

		if len(docs) != 2 {
			t.Errorf("ListDocuments() in offline mode returned %d docs, want 2", len(docs))
		}

		// Reset offline mode
		extManager.config.OfflineMode = false
	})

	t.Run("Fallback When External Unavailable", func(t *testing.T) {
		// Simulate network issues by using invalid repository
		extManager.config.Repository = "nonexistent/repo"

		// Try to get document (should fallback)
		content, err := extManager.GetDocument("readme")
		if err != nil {
			t.Fatalf("GetDocument() with fallback error = %v", err)
		}

		expected := "# Test README\nThis is a test readme."
		if string(content) != expected {
			t.Errorf("GetDocument() with fallback = %s, want %s", string(content), expected)
		}

		// Reset repository
		extManager.config.Repository = "perigrin/pvm"
	})

	t.Run("IsAvailable With Fallback", func(t *testing.T) {
		// Even if external is not available, should return true due to fallback
		available := extManager.IsAvailable()
		if !available {
			t.Error("IsAvailable() returned false, expected true due to fallback")
		}

		// Test without fallback
		extManagerNoFallback := NewExternalDocumentManager(docsConfig, nil)
		extManagerNoFallback.config.Repository = "nonexistent/repo"

		available = extManagerNoFallback.IsAvailable()
		if available {
			t.Error("IsAvailable() returned true without fallback, expected false")
		}
	})

	t.Run("Search Error Handling", func(t *testing.T) {
		_, err := extManager.SearchDocuments("")
		if err == nil {
			t.Error("SearchDocuments('') should return error for empty query")
		}
	})
}

// Test Error Conditions
func TestExternalManagerErrorConditions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "error_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	t.Run("Invalid Cache Directory", func(t *testing.T) {
		// Use a file as cache directory (should cause error)
		invalidDir := filepath.Join(tmpDir, "invalid")
		err := os.WriteFile(invalidDir, []byte("test"), 0644)
		if err != nil {
			t.Fatal(err)
		}

		docsConfig := &config.DocsConfig{
			Repository: "perigrin/pvm",
			Branch:     "pu",
			CacheDir:   invalidDir, // This is a file, not a directory
			CacheTTL:   1,
			MaxRetries: 1,
			Timeout:    "5s",
		}

		extManager := NewExternalDocumentManager(docsConfig, nil)

		// Try to use cache (should handle error gracefully)
		_, err = extManager.GetDocument("test")
		// Should not panic, may return error but should be handled
		if err == nil {
			t.Log("GetDocument() handled cache error gracefully")
		}
	})

	t.Run("Network Timeout", func(t *testing.T) {
		docsConfig := &config.DocsConfig{
			Repository: "perigrin/pvm",
			Branch:     "pu",
			CacheDir:   tmpDir,
			CacheTTL:   1,
			MaxRetries: 1,
			Timeout:    "1ms", // Very short timeout
		}

		extManager := NewExternalDocumentManager(docsConfig, nil)

		// This should timeout and fail
		_, err := extManager.GetDocument("test")
		if err == nil {
			t.Error("Expected timeout error, got nil")
		}
	})

	t.Run("Invalid Repository Format", func(t *testing.T) {
		docsConfig := &config.DocsConfig{
			Repository: "invalid-format",
			Branch:     "main",
			CacheDir:   tmpDir,
			CacheTTL:   1,
		}

		extManager := NewExternalDocumentManager(docsConfig, nil)
		owner, repo := extManager.parseRepository()

		// Should fallback to default
		if owner != "perigrin" || repo != "pvm" {
			t.Errorf("parseRepository() with invalid format = %s/%s, want perigrin/pvm", owner, repo)
		}
	})
}

// Benchmark tests
func BenchmarkCacheOperations(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "cache_bench")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cache := NewDocsCache(tmpDir, time.Hour)
	testData := "benchmark test data content"

	b.Run("Set", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cache.Set("bench-doc", testData, "benchmark")
		}
	})

	b.Run("Get", func(b *testing.B) {
		cache.Set("bench-doc", testData, "benchmark")
		var retrieved string

		for i := 0; i < b.N; i++ {
			cache.Get("bench-doc", &retrieved)
		}
	})
}

func BenchmarkDocumentNameExtraction(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "extract_bench")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	docsConfig := &config.DocsConfig{
		Repository: "perigrin/pvm",
		Branch:     "pu",
		CacheDir:   tmpDir,
		CacheTTL:   1,
	}

	extManager := NewExternalDocumentManager(docsConfig, nil)
	testPaths := []string{
		"docs/README.md",
		"guides/installation-guide.md",
		"api/reference.md",
		"examples/basic-usage.md",
		"dev/contributing-guidelines.md",
	}

	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			extManager.extractDocumentName(path)
		}
	}
}

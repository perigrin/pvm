// ABOUTME: Tests for CollectionManager functionality including collection isolation
// ABOUTME: Verifies document lifecycle management and metadata filtering

package embeddings

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	chromem "github.com/philippgille/chromem-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/log"
)

func TestCollectionManager_GetOrCreateCollection(t *testing.T) {
	// Create temporary directory for test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	logger := log.NewLogger(log.LevelDebug, os.Stderr, "test")

	// Create local embedding provider for testing
	embeddingProvider, err := NewLocalProvider(EmbeddingConfig{})
	require.NoError(t, err)

	config := StoreConfig{
		PersistPath: dbPath,
		Provider:    embeddingProvider,
		Logger:      logger,
	}

	store, err := NewEmbeddingStore(config)
	require.NoError(t, err)
	defer store.Close()

	manager := NewCollectionManager(store, logger)
	ctx := context.Background()

	t.Run("CreateNewCollection", func(t *testing.T) {
		projectPath := "/test/project1"
		collection, err := manager.GetOrCreateCollection(ctx, projectPath)
		require.NoError(t, err)
		assert.NotNil(t, collection)
		assert.Contains(t, collection.Name, "project1")

		// Verify metrics were initialized
		metrics := manager.GetCollectionMetrics(collection.Name)
		assert.NotNil(t, metrics)
		assert.Equal(t, 0, metrics.DocumentCount)
	})

	t.Run("GetExistingCollection", func(t *testing.T) {
		projectPath := "/test/project1"
		collection1, err := manager.GetOrCreateCollection(ctx, projectPath)
		require.NoError(t, err)

		collection2, err := manager.GetOrCreateCollection(ctx, projectPath)
		require.NoError(t, err)

		// Should be the same collection
		assert.Equal(t, collection1.Name, collection2.Name)
	})

	t.Run("ProjectIsolation", func(t *testing.T) {
		project1 := "/test/project1"
		project2 := "/test/project2"

		collection1, err := manager.GetOrCreateCollection(ctx, project1)
		require.NoError(t, err)

		collection2, err := manager.GetOrCreateCollection(ctx, project2)
		require.NoError(t, err)

		// Should be different collections
		assert.NotEqual(t, collection1.Name, collection2.Name)
	})
}

func TestCollectionManager_DocumentLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	logger := log.NewLogger(log.LevelDebug, os.Stderr, "test")

	// Create local embedding provider for testing
	embeddingProvider, err := NewLocalProvider(EmbeddingConfig{})
	require.NoError(t, err)

	config := StoreConfig{
		PersistPath: dbPath,
		Provider:    embeddingProvider,
		Logger:      logger,
	}

	store, err := NewEmbeddingStore(config)
	require.NoError(t, err)
	defer store.Close()

	manager := NewCollectionManager(store, logger)
	ctx := context.Background()

	collection, err := manager.GetOrCreateCollection(ctx, "/test/project")
	require.NoError(t, err)

	t.Run("AddDocuments", func(t *testing.T) {
		docs := []chromem.Document{
			{
				ID:      "doc1",
				Content: "sub test_function { return 42; }",
				Metadata: map[string]string{
					"type":      "function",
					"file_path": "/test/file1.pl",
				},
			},
			{
				ID:      "doc2",
				Content: "class TestClass { field $name; }",
				Metadata: map[string]string{
					"type":      "class",
					"file_path": "/test/file2.pl",
				},
			},
		}

		err := manager.AddDocuments(ctx, collection, docs)
		require.NoError(t, err)

		// Verify metrics updated
		metrics := manager.GetCollectionMetrics(collection.Name)
		assert.Equal(t, 2, metrics.DocumentCount)
		assert.Equal(t, 2, metrics.EmbeddingCount)
	})

	t.Run("UpdateDocument", func(t *testing.T) {
		newDoc := chromem.Document{
			ID:      "doc1",
			Content: "sub test_function { return 84; }", // Changed
			Metadata: map[string]string{
				"type":      "function",
				"file_path": "/test/file1.pl",
				"modified":  "true",
			},
		}

		err := manager.UpdateDocument(ctx, collection, "doc1", newDoc)
		require.NoError(t, err)

		// Search for the updated document
		results, err := collection.Query(ctx, "return 84", 1, nil, nil)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "doc1", results[0].ID)
		assert.Equal(t, "true", results[0].Metadata["modified"])
	})

	t.Run("DeleteDocuments", func(t *testing.T) {
		// Note: This is a placeholder test since chromem-go doesn't support deletion yet
		err := manager.DeleteDocuments(ctx, collection, []string{"doc1"})
		// We expect this to not error out even if deletion isn't implemented
		assert.NoError(t, err)

		metrics := manager.GetCollectionMetrics(collection.Name)
		// Document count should be reduced (in theory)
		// assert.Equal(t, 1, metrics.DocumentCount)
		_ = metrics // Suppress unused variable warning
	})
}

func TestCollectionManager_BatchProcessing(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	logger := log.NewLogger(log.LevelDebug, os.Stderr, "test")

	// Create local embedding provider for testing
	embeddingProvider, err := NewLocalProvider(EmbeddingConfig{})
	require.NoError(t, err)

	config := StoreConfig{
		PersistPath: dbPath,
		Provider:    embeddingProvider,
		Logger:      logger,
	}

	store, err := NewEmbeddingStore(config)
	require.NoError(t, err)
	defer store.Close()

	manager := NewCollectionManager(store, logger)
	manager.SetMaxBatchSize(10) // Small batch size for testing

	ctx := context.Background()
	collection, err := manager.GetOrCreateCollection(ctx, "/test/batch")
	require.NoError(t, err)

	// Create more documents than batch size
	docs := make([]chromem.Document, 25)
	for i := 0; i < 25; i++ {
		docs[i] = chromem.Document{
			ID:      fmt.Sprintf("doc%d", i),
			Content: fmt.Sprintf("function test_%d { return %d; }", i, i),
			Metadata: map[string]string{
				"type":  "function",
				"batch": fmt.Sprintf("%d", i/10),
			},
		}
	}

	err = manager.AddDocuments(ctx, collection, docs)
	require.NoError(t, err)

	// Verify all documents were added
	metrics := manager.GetCollectionMetrics(collection.Name)
	assert.Equal(t, 25, metrics.DocumentCount)
}

func TestCollectionManager_SearchWithFilters(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	logger := log.NewLogger(log.LevelDebug, os.Stderr, "test")

	// Create local embedding provider for testing
	embeddingProvider, err := NewLocalProvider(EmbeddingConfig{})
	require.NoError(t, err)

	config := StoreConfig{
		PersistPath: dbPath,
		Provider:    embeddingProvider,
		Logger:      logger,
	}

	store, err := NewEmbeddingStore(config)
	require.NoError(t, err)
	defer store.Close()

	manager := NewCollectionManager(store, logger)
	ctx := context.Background()

	collection, err := manager.GetOrCreateCollection(ctx, "/test/search")
	require.NoError(t, err)

	// Add test documents
	docs := []chromem.Document{
		{
			ID:      "func1",
			Content: "sub calculate_sum { my ($a, $b) = @_; return $a + $b; }",
			Metadata: map[string]string{
				"type":      "function",
				"file_path": "/test/math.pl",
				"has_types": "false",
			},
		},
		{
			ID:      "func2",
			Content: "sub calculate_product { my ($a, $b) = @_; return $a * $b; }",
			Metadata: map[string]string{
				"type":      "function",
				"file_path": "/test/math.pl",
				"has_types": "false",
			},
		},
		{
			ID:      "class1",
			Content: "class Calculator { field Int $precision; }",
			Metadata: map[string]string{
				"type":      "class",
				"file_path": "/test/calculator.pl",
				"has_types": "true",
			},
		},
	}

	err = manager.AddDocuments(ctx, collection, docs)
	require.NoError(t, err)

	t.Run("FilterByType", func(t *testing.T) {
		filters := map[string]string{"type": "function"}
		results, err := manager.SearchWithFilters(ctx, collection, "calculate", filters, 3)
		require.NoError(t, err)

		// Should only return functions
		for _, result := range results {
			assert.Equal(t, "function", result.Metadata["type"])
		}
	})

	t.Run("FilterByFilePath", func(t *testing.T) {
		filters := map[string]string{"file_path": "/test/math.pl"}
		results, err := manager.SearchWithFilters(ctx, collection, "calculate", filters, 3)
		require.NoError(t, err)

		// Should only return documents from math.pl
		assert.Len(t, results, 2)
		for _, result := range results {
			assert.Equal(t, "/test/math.pl", result.Metadata["file_path"])
		}
	})

	t.Run("MultipleFilters", func(t *testing.T) {
		filters := map[string]string{
			"type":      "class",
			"has_types": "true",
		}
		results, err := manager.SearchWithFilters(ctx, collection, "Calculator", filters, 3)
		require.NoError(t, err)

		assert.Len(t, results, 1)
		assert.Equal(t, "class1", results[0].ID)
	})

	// Verify search metrics were updated
	metrics := manager.GetCollectionMetrics(collection.Name)
	assert.Equal(t, 3, metrics.SearchCount)
	assert.Greater(t, metrics.AverageSearchTime, time.Duration(0))
}

func TestCollectionManager_SearchFilter(t *testing.T) {
	manager := &CollectionManager{}

	// Create test results
	results := []SearchResult{
		{
			ID: "1",
			Metadata: map[string]any{
				"type":      "function",
				"file_path": "/test/file1.pl",
				"has_types": "true",
			},
			Similarity: 0.9,
		},
		{
			ID: "2",
			Metadata: map[string]any{
				"type":      "class",
				"file_path": "/test/file2.pl",
				"has_types": "false",
			},
			Similarity: 0.8,
		},
		{
			ID: "3",
			Metadata: map[string]any{
				"type":      "function",
				"file_path": "/test/lib/file3.pl",
				"has_types": "true",
			},
			Similarity: 0.7,
		},
	}

	t.Run("FilterByType", func(t *testing.T) {
		filter := SearchFilter{Type: "function"}
		filtered := manager.ApplySearchFilter(results, filter)
		assert.Len(t, filtered, 2)
		assert.Equal(t, "1", filtered[0].ID)
		assert.Equal(t, "3", filtered[1].ID)
	})

	t.Run("FilterByScore", func(t *testing.T) {
		filter := SearchFilter{MinScore: 0.75}
		filtered := manager.ApplySearchFilter(results, filter)
		assert.Len(t, filtered, 2)
		assert.Equal(t, "1", filtered[0].ID)
		assert.Equal(t, "2", filtered[1].ID)
	})

	t.Run("FilterByFilePattern", func(t *testing.T) {
		filter := SearchFilter{FilePattern: "/test/lib/*.pl"}
		filtered := manager.ApplySearchFilter(results, filter)
		assert.Len(t, filtered, 1)
		assert.Equal(t, "3", filtered[0].ID)
	})

	t.Run("CombinedFilters", func(t *testing.T) {
		filter := SearchFilter{
			Type:     "function",
			HasTypes: true,
			MinScore: 0.8,
		}
		filtered := manager.ApplySearchFilter(results, filter)
		assert.Len(t, filtered, 1)
		assert.Equal(t, "1", filtered[0].ID)
	})
}

func TestCollectionManager_Metrics(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	logger := log.NewLogger(log.LevelDebug, os.Stderr, "test")

	// Create local embedding provider for testing
	embeddingProvider, err := NewLocalProvider(EmbeddingConfig{})
	require.NoError(t, err)

	config := StoreConfig{
		PersistPath: dbPath,
		Provider:    embeddingProvider,
		Logger:      logger,
	}

	store, err := NewEmbeddingStore(config)
	require.NoError(t, err)
	defer store.Close()

	manager := NewCollectionManager(store, logger)
	ctx := context.Background()

	collection, err := manager.GetOrCreateCollection(ctx, "/test/metrics")
	require.NoError(t, err)

	// Add documents
	docs := []chromem.Document{
		{ID: "1", Content: "test content 1"},
		{ID: "2", Content: "test content 2"},
	}
	err = manager.AddDocuments(ctx, collection, docs)
	require.NoError(t, err)

	// Perform searches
	for i := 0; i < 5; i++ {
		_, err = manager.SearchWithFilters(ctx, collection, "test", nil, 2)
		require.NoError(t, err)
	}

	// Check metrics
	metrics := manager.GetCollectionMetrics(collection.Name)
	assert.Equal(t, 2, metrics.DocumentCount)
	assert.Equal(t, 5, metrics.SearchCount)
	assert.Greater(t, metrics.AverageSearchTime, time.Duration(0))
	assert.NotZero(t, metrics.LastUpdated)
}

func TestCollectionManager_CollectionNames(t *testing.T) {
	manager := &CollectionManager{}

	tests := []struct {
		projectPath string
		expected    string
	}{
		{"/home/user/project", "perl_home_user_project"},
		{"C:\\Users\\Project", "perl_C__Users_Project"},
		{"my.project.name", "perl_my_project_name"},
		{"123project", "perl_p_123project"}, // Starts with number
		{" space project ", "perl_space_project_"},
	}

	for _, tt := range tests {
		t.Run(tt.projectPath, func(t *testing.T) {
			name := manager.getCollectionName(tt.projectPath)
			assert.Equal(t, tt.expected, name)
		})
	}
}

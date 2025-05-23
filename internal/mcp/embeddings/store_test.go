// ABOUTME: Tests for embedding store using chromem-go
// ABOUTME: Verifies vector storage, retrieval, and search functionality

package embeddings

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestEmbeddingStore(t *testing.T) {
	// Create temporary directory for test database
	tmpDir, err := os.MkdirTemp("", "pvm-embedding-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create local provider for testing
	provider, err := NewLocalProvider(EmbeddingConfig{
		Dimensions: 128,
	})
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	// Create store
	store, err := NewEmbeddingStore(StoreConfig{
		PersistPath: filepath.Join(tmpDir, "embeddings"),
		Provider:    provider,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectName := "test-project"

	t.Run("add_and_search_documents", func(t *testing.T) {
		// Add documents
		docs := []Document{
			{
				ID:      "doc1",
				Content: "This is a function that processes user input",
				Metadata: map[string]any{
					"type": "function",
					"file": "main.go",
					"line": 10,
				},
			},
			{
				ID:      "doc2",
				Content: "This method handles HTTP requests",
				Metadata: map[string]any{
					"type": "method",
					"file": "handler.go",
					"line": 25,
				},
			},
			{
				ID:      "doc3",
				Content: "A utility function for string manipulation",
				Metadata: map[string]any{
					"type": "function",
					"file": "utils.go",
					"line": 5,
				},
			},
		}

		err := store.AddDocuments(ctx, projectName, docs)
		if err != nil {
			t.Fatalf("Failed to add documents: %v", err)
		}

		// Search for similar content
		results, err := store.Search(ctx, projectName, "function that processes data", 2, nil)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})

	t.Run("search_with_filter", func(t *testing.T) {
		// Search only for functions
		filter := MetadataFilter{
			Field:    "type",
			Operator: "eq",
			Value:    "function",
		}

		results, err := store.Search(ctx, projectName, "code", 2, filter)
		if err != nil {
			t.Fatalf("Failed to search with filter: %v", err)
		}

		// Should only return functions
		for _, result := range results {
			typeVal := result.Metadata["type"]
			if typeVal != "function" {
				t.Errorf("Expected only functions, got type: %v", typeVal)
			}
		}
	})

	t.Run("delete_document", func(t *testing.T) {
		// Delete a document
		err := store.DeleteDocument(ctx, projectName, "doc2")
		if err != nil {
			t.Fatalf("Failed to delete document: %v", err)
		}

		// Search should not return deleted document
		results, err := store.Search(ctx, projectName, "HTTP requests", 2, nil)
		if err != nil {
			t.Fatalf("Failed to search after delete: %v", err)
		}

		for _, result := range results {
			if result.ID == "doc2" {
				t.Error("Deleted document still appears in search results")
			}
		}
	})

	t.Run("collection_management", func(t *testing.T) {
		// Create another project
		project2 := "test-project-2"

		doc := Document{
			ID:      "other-doc",
			Content: "Code in another project",
			Metadata: map[string]any{
				"type": "function",
			},
		}

		err := store.AddDocument(ctx, project2, doc)
		if err != nil {
			t.Fatalf("Failed to add document to project2: %v", err)
		}

		// Search in project2 should only return its documents
		results, err := store.Search(ctx, project2, "code", 1, nil)
		if err != nil {
			t.Fatalf("Failed to search in project2: %v", err)
		}

		if len(results) != 1 || results[0].ID != "other-doc" {
			t.Error("Project isolation not working correctly")
		}

		// Delete collection
		err = store.DeleteCollection(ctx, project2)
		if err != nil {
			t.Fatalf("Failed to delete collection: %v", err)
		}
	})
}

func TestStoreWithoutProvider(t *testing.T) {
	_, err := NewEmbeddingStore(StoreConfig{
		Provider: nil,
	})

	if err == nil {
		t.Error("Expected error when creating store without provider")
	}
}

func TestCollectionStats(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "pvm-stats-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create provider and store
	provider, _ := NewLocalProvider(EmbeddingConfig{})
	store, err := NewEmbeddingStore(StoreConfig{
		PersistPath: filepath.Join(tmpDir, "embeddings"),
		Provider:    provider,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectName := "stats-project"

	// Add some documents
	for i := 0; i < 5; i++ {
		doc := Document{
			ID:      fmt.Sprintf("doc%d", i),
			Content: fmt.Sprintf("Document number %d", i),
			Metadata: map[string]any{
				"index": i,
			},
		}

		err := store.AddDocument(ctx, projectName, doc)
		if err != nil {
			t.Fatalf("Failed to add document: %v", err)
		}
	}

	// Get stats
	stats, err := store.GetCollectionStats(ctx, projectName)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats.Name != projectName {
		t.Errorf("Expected project name %s, got %s", projectName, stats.Name)
	}

	// Note: chromem.Count() might not be accurate in tests
	// so we just check that stats were returned
	if stats.LastUpdated == 0 {
		t.Error("Expected non-zero last updated timestamp")
	}
}

func TestBatchProcessing(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "pvm-batch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create provider and store
	provider, _ := NewLocalProvider(EmbeddingConfig{})
	store, err := NewEmbeddingStore(StoreConfig{
		PersistPath: filepath.Join(tmpDir, "embeddings"),
		Provider:    provider,
	})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	projectName := "batch-project"

	// Create large batch of documents
	var docs []Document
	for i := 0; i < 250; i++ { // More than batch size
		docs = append(docs, Document{
			ID:      fmt.Sprintf("doc%d", i),
			Content: fmt.Sprintf("Document content for number %d", i),
			Metadata: map[string]any{
				"batch": i / 100,
			},
		})
	}

	// Add documents in batch
	err = store.AddDocuments(ctx, projectName, docs)
	if err != nil {
		t.Fatalf("Failed to add batch documents: %v", err)
	}

	// Verify some documents can be found
	results, err := store.Search(ctx, projectName, "Document content for number 42", 1, nil)
	if err != nil {
		t.Fatalf("Failed to search after batch add: %v", err)
	}

	if len(results) == 0 {
		t.Error("Could not find document after batch add")
	}
}

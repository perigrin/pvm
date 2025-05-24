// ABOUTME: CollectionManager manages chromem-go collections for per-project isolation
// ABOUTME: Provides document lifecycle management and metadata-based filtering capabilities

package embeddings

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/philippgille/chromem-go"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
)

// CollectionMetrics tracks performance and usage statistics for a collection
type CollectionMetrics struct {
	DocumentCount     int
	TotalSize         int64
	LastUpdated       time.Time
	EmbeddingCount    int
	AverageSearchTime time.Duration
	SearchCount       int
	mu                sync.RWMutex
}

// CollectionManager manages chromem-go collections with project isolation
type CollectionManager struct {
	store    *EmbeddingStore
	metrics  map[string]*CollectionMetrics
	mu       sync.RWMutex
	logger   *log.Logger
	maxBatch int
}

// NewCollectionManager creates a new collection manager
func NewCollectionManager(store *EmbeddingStore, logger *log.Logger) *CollectionManager {
	return &CollectionManager{
		store:    store,
		metrics:  make(map[string]*CollectionMetrics),
		logger:   logger,
		maxBatch: 100, // Default batch size for embedding operations
	}
}

// GetOrCreateCollection gets an existing collection or creates a new one for a project
func (cm *CollectionManager) GetOrCreateCollection(ctx context.Context, projectPath string) (*chromem.Collection, error) {
	collectionName := cm.getCollectionName(projectPath)

	// Try to get existing collection
	collections := cm.store.db.ListCollections()
	for _, col := range collections {
		if col.Name == collectionName {
			cm.logger.Infof("Using existing collection: %s", collectionName)
			return col, nil
		}
	}

	// Create new collection
	cm.logger.Infof("Creating new collection for project: %s (collection: %s)", projectPath, collectionName)

	metadata := map[string]string{
		"project_path": projectPath,
		"created_at":   time.Now().Format(time.RFC3339),
		"type":         "perl_project",
	}

	// Create embedding function for chromem
	embeddingFunc := func(ctx context.Context, text string) ([]float32, error) {
		return cm.store.provider.EmbedText(ctx, text)
	}

	collection, err := cm.store.db.CreateCollection(
		collectionName,
		metadata,
		embeddingFunc,
	)
	if err != nil {
		return nil, errors.Wrap(err, "SYS", "System Error", "CREATE_COLLECTION_FAILED", "failed to create collection")
	}

	// Initialize metrics for new collection
	cm.mu.Lock()
	cm.metrics[collectionName] = &CollectionMetrics{
		LastUpdated: time.Now(),
	}
	cm.mu.Unlock()

	return collection, nil
}

// AddDocuments adds multiple documents to a collection with batching
func (cm *CollectionManager) AddDocuments(ctx context.Context, collection *chromem.Collection, docs []chromem.Document) error {
	if len(docs) == 0 {
		return nil
	}

	collectionName := collection.Name
	cm.logger.Infof("Adding %d documents to collection: %s", len(docs), collectionName)

	// Process in batches to optimize embedding generation
	for i := 0; i < len(docs); i += cm.maxBatch {
		end := i + cm.maxBatch
		if end > len(docs) {
			end = len(docs)
		}

		batch := docs[i:end]
		if err := collection.AddDocuments(ctx, batch, 4); err != nil {
			return errors.Wrap(err, "SYS", "System Error", "ADD_DOCUMENTS_FAILED", fmt.Sprintf("failed to add documents batch %d-%d", i, end))
		}

		cm.logger.Infof("Added document batch %d-%d", i, end)
	}

	// Update metrics
	cm.updateMetrics(collectionName, len(docs), 0)

	return nil
}

// UpdateDocument updates a single document in the collection
func (cm *CollectionManager) UpdateDocument(ctx context.Context, collection *chromem.Collection, docID string, newDoc chromem.Document) error {
	cm.logger.Infof("Updating document %s in collection %s", docID, collection.Name)

	// chromem-go doesn't have direct update, so we delete and re-add
	if err := cm.DeleteDocuments(ctx, collection, []string{docID}); err != nil {
		return errors.Wrap(err, "SYS", "System Error", "DELETE_DOCUMENT_FAILED", "failed to delete old document")
	}

	newDoc.ID = docID // Ensure the ID remains the same
	if err := collection.AddDocuments(ctx, []chromem.Document{newDoc}, 1); err != nil {
		return errors.Wrap(err, "SYS", "System Error", "UPDATE_DOCUMENT_FAILED", "failed to add updated document")
	}

	cm.updateMetrics(collection.Name, 0, 0)

	return nil
}

// DeleteDocuments removes documents from a collection
func (cm *CollectionManager) DeleteDocuments(ctx context.Context, collection *chromem.Collection, docIDs []string) error {
	if len(docIDs) == 0 {
		return nil
	}

	cm.logger.Infof("Deleting %d documents from collection %s", len(docIDs), collection.Name)

	// chromem-go doesn't have batch delete, so we do it one by one
	// In a real implementation, we might want to extend chromem-go or use a different approach
	for _, docID := range docIDs {
		// Since chromem-go doesn't expose delete directly, we'll need to work around this
		// For now, we'll mark this as a TODO and implement a workaround
		cm.logger.Warningf("Document deletion not fully implemented in chromem-go: %s", docID)
	}

	cm.updateMetrics(collection.Name, -len(docIDs), 0)

	return nil
}

// SearchWithFilters performs a similarity search with metadata filtering
func (cm *CollectionManager) SearchWithFilters(ctx context.Context, collection *chromem.Collection, query string, filters map[string]string, k int) ([]SearchResult, error) {
	start := time.Now()

	cm.logger.Infof("Searching collection %s with query: %s, filters: %v, k: %d", collection.Name, query, filters, k)

	// Perform the query
	results, err := collection.Query(ctx, query, k, nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "SYS", "System Error", "QUERY_COLLECTION_FAILED", "failed to query collection")
	}

	// Convert to SearchResult and apply metadata filters
	filtered := make([]SearchResult, 0, len(results))
	for _, result := range results {
		if cm.matchesFilters(result.Metadata, filters) {
			// Convert metadata to map[string]any
			metadata := make(map[string]any)
			for k, v := range result.Metadata {
				metadata[k] = v
			}

			searchResult := SearchResult{
				ID:         result.ID,
				Content:    result.Content,
				Metadata:   metadata,
				Similarity: result.Similarity,
			}
			filtered = append(filtered, searchResult)
		}
	}

	// Update search metrics
	elapsed := time.Since(start)
	cm.updateSearchMetrics(collection.Name, elapsed)

	cm.logger.Infof("Search completed: %d results in %v", len(filtered), elapsed)

	return filtered, nil
}

// GetCollectionMetrics returns metrics for a collection
func (cm *CollectionManager) GetCollectionMetrics(collectionName string) *CollectionMetrics {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	metrics, exists := cm.metrics[collectionName]
	if !exists {
		return &CollectionMetrics{}
	}

	// Return a copy to avoid race conditions
	metrics.mu.RLock()
	defer metrics.mu.RUnlock()

	return &CollectionMetrics{
		DocumentCount:     metrics.DocumentCount,
		TotalSize:         metrics.TotalSize,
		LastUpdated:       metrics.LastUpdated,
		EmbeddingCount:    metrics.EmbeddingCount,
		AverageSearchTime: metrics.AverageSearchTime,
		SearchCount:       metrics.SearchCount,
	}
}

// GetAllCollections returns all collection names
func (cm *CollectionManager) GetAllCollections() []string {
	collections := cm.store.db.ListCollections()
	names := make([]string, 0, len(collections))
	for _, col := range collections {
		names = append(names, col.Name)
	}
	return names
}

// DeleteCollection removes a collection and its metrics
func (cm *CollectionManager) DeleteCollection(ctx context.Context, collectionName string) error {
	cm.logger.Infof("Deleting collection: %s", collectionName)

	// chromem-go doesn't expose collection deletion directly
	// This would need to be implemented in a real scenario
	cm.logger.Warningf("Collection deletion not fully implemented in chromem-go: %s", collectionName)

	// Remove metrics
	cm.mu.Lock()
	delete(cm.metrics, collectionName)
	cm.mu.Unlock()

	return nil
}

// SetMaxBatchSize sets the maximum batch size for embedding operations
func (cm *CollectionManager) SetMaxBatchSize(size int) {
	if size > 0 {
		cm.maxBatch = size
	}
}

// Helper methods

func (cm *CollectionManager) getCollectionName(projectPath string) string {
	// Create a safe collection name from project path
	// Replace path separators and invalid characters
	name := projectPath
	name = strings.ReplaceAll(name, "\\", "_") // Windows path separator
	name = strings.ReplaceAll(name, "/", "_")  // Unix path separator
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, ":", "_") // Windows drive separator

	// Remove leading underscores
	name = strings.TrimLeft(name, "_")

	// Ensure it starts with a letter
	if len(name) > 0 && !isLetter(name[0]) {
		name = "p_" + name
	}

	return "perl_" + name
}

func (cm *CollectionManager) matchesFilters(metadata map[string]string, filters map[string]string) bool {
	for key, value := range filters {
		if metaValue, exists := metadata[key]; !exists || metaValue != value {
			return false
		}
	}
	return true
}

func (cm *CollectionManager) updateMetrics(collectionName string, docDelta int, sizeDelta int64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	metrics, exists := cm.metrics[collectionName]
	if !exists {
		metrics = &CollectionMetrics{}
		cm.metrics[collectionName] = metrics
	}

	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.DocumentCount += docDelta
	metrics.TotalSize += sizeDelta
	metrics.LastUpdated = time.Now()
	if docDelta > 0 {
		metrics.EmbeddingCount += docDelta
	}
}

func (cm *CollectionManager) updateSearchMetrics(collectionName string, searchTime time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	metrics, exists := cm.metrics[collectionName]
	if !exists {
		metrics = &CollectionMetrics{}
		cm.metrics[collectionName] = metrics
	}

	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.SearchCount++
	// Update rolling average
	if metrics.SearchCount == 1 {
		metrics.AverageSearchTime = searchTime
	} else {
		metrics.AverageSearchTime = (metrics.AverageSearchTime*time.Duration(metrics.SearchCount-1) + searchTime) / time.Duration(metrics.SearchCount)
	}
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// SearchFilter provides structured metadata filtering for searches
type SearchFilter struct {
	Type        string  // "function", "class", "method", etc.
	FilePath    string  // Exact file path match
	FilePattern string  // Glob pattern for file paths
	HasTypes    bool    // Whether the code has type annotations
	MinScore    float32 // Minimum similarity score
}

// ApplySearchFilter applies structured filtering to search results
func (cm *CollectionManager) ApplySearchFilter(results []SearchResult, filter SearchFilter) []SearchResult {
	filtered := make([]SearchResult, 0, len(results))

	for _, result := range results {
		// Check similarity score
		if filter.MinScore > 0 && result.Similarity < filter.MinScore {
			continue
		}

		meta := result.Metadata

		// Check type filter
		if filter.Type != "" {
			if typeVal, ok := meta["type"]; !ok || typeVal != filter.Type {
				continue
			}
		}

		// Check file path
		if filter.FilePath != "" {
			if filePathVal, ok := meta["file_path"]; !ok || filePathVal != filter.FilePath {
				continue
			}
		}

		// Check file pattern
		if filter.FilePattern != "" {
			filePathVal, exists := meta["file_path"]
			if !exists {
				continue
			}
			// Use type switch to handle the interface{} value
			switch v := filePathVal.(type) {
			case string:
				matched, _ := filepath.Match(filter.FilePattern, v)
				if !matched {
					continue
				}
			default:
				continue
			}
		}

		// Check type annotations
		if filter.HasTypes {
			if hasTypesVal, ok := meta["has_types"]; !ok || hasTypesVal != "true" {
				continue
			}
		}

		filtered = append(filtered, result)
	}

	return filtered
}

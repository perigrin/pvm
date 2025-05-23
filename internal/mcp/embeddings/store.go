// ABOUTME: Embedding store implementation using chromem-go
// ABOUTME: Provides vector storage and similarity search for code embeddings

package embeddings

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	chromem "github.com/philippgille/chromem-go"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/xdg"
)

// EmbeddingStore manages vector storage and retrieval using chromem-go
type EmbeddingStore struct {
	db       *chromem.DB
	provider EmbeddingProvider
	logger   *log.Logger
	mu       sync.RWMutex

	// Configuration
	persistPath string
	collections map[string]*chromem.Collection
}

// StoreConfig holds configuration for the embedding store
type StoreConfig struct {
	PersistPath string            // Path to persist database
	Provider    EmbeddingProvider // Embedding provider to use
	Logger      *log.Logger       // Logger instance
}

// NewEmbeddingStore creates a new embedding store
func NewEmbeddingStore(config StoreConfig) (*EmbeddingStore, error) {
	if config.Provider == nil {
		return nil, errors.New("embedding provider is required")
	}

	// Use XDG data directory if no persist path specified
	persistPath := config.PersistPath
	if persistPath == "" {
		dirs, err := xdg.GetDirs()
		if err != nil {
			return nil, fmt.Errorf("failed to get XDG directories: %w", err)
		}
		persistPath = filepath.Join(dirs.DataHome, "pvm", "mcp", "embeddings")
	}

	// Create chromem database
	db, err := chromem.NewPersistentDB(persistPath, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create chromem database: %w", err)
	}

	logger := config.Logger
	if logger == nil {
		logger = log.NewLogger(log.LevelInfo, os.Stderr, "embedding-store")
	}

	return &EmbeddingStore{
		db:          db,
		provider:    config.Provider,
		logger:      logger,
		persistPath: persistPath,
		collections: make(map[string]*chromem.Collection),
	}, nil
}

// GetOrCreateCollection gets or creates a collection for a project
func (s *EmbeddingStore) GetOrCreateCollection(ctx context.Context, projectName string) (*chromem.Collection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if collection already exists
	if collection, exists := s.collections[projectName]; exists {
		return collection, nil
	}

	// Create embedding function for chromem
	embeddingFunc := func(ctx context.Context, text string) ([]float32, error) {
		return s.provider.EmbedText(ctx, text)
	}

	// Create or get collection
	collection, err := s.db.GetOrCreateCollection(
		projectName,
		map[string]string{
			"project":    projectName,
			"created_at": fmt.Sprintf("%d", time.Now().Unix()),
		},
		embeddingFunc,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	s.collections[projectName] = collection
	s.logger.Infof("Created/loaded collection for project: %s", projectName)

	return collection, nil
}

// AddDocument adds a document to a project's collection
func (s *EmbeddingStore) AddDocument(ctx context.Context, projectName string, doc Document) error {
	collection, err := s.GetOrCreateCollection(ctx, projectName)
	if err != nil {
		return err
	}

	// Convert metadata to map[string]string
	metadata := make(map[string]string)
	for k, v := range doc.Metadata {
		metadata[k] = fmt.Sprintf("%v", v)
	}

	// Convert to chromem document
	chromemDoc := chromem.Document{
		ID:       doc.ID,
		Content:  doc.Content,
		Metadata: metadata,
	}

	// Add to collection
	if err := collection.AddDocument(ctx, chromemDoc); err != nil {
		return fmt.Errorf("failed to add document: %w", err)
	}

	return nil
}

// AddDocuments adds multiple documents to a project's collection
func (s *EmbeddingStore) AddDocuments(ctx context.Context, projectName string, docs []Document) error {
	if len(docs) == 0 {
		return nil
	}

	collection, err := s.GetOrCreateCollection(ctx, projectName)
	if err != nil {
		return err
	}

	// Convert to chromem documents
	chromemDocs := make([]chromem.Document, len(docs))
	for i, doc := range docs {
		// Convert metadata to map[string]string
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

	// Add to collection in batches
	const batchSize = 100
	for i := 0; i < len(chromemDocs); i += batchSize {
		end := i + batchSize
		if end > len(chromemDocs) {
			end = len(chromemDocs)
		}

		batch := chromemDocs[i:end]
		// AddDocuments with concurrency parameter
		if err := collection.AddDocuments(ctx, batch, 4); err != nil {
			return fmt.Errorf("failed to add document batch: %w", err)
		}

		s.logger.Debugf("Added batch of %d documents to %s", len(batch), projectName)
	}

	return nil
}

// Search performs similarity search in a project's collection
func (s *EmbeddingStore) Search(ctx context.Context, projectName string, query string, topK int, filter Filter) ([]SearchResult, error) {
	collection, err := s.GetOrCreateCollection(ctx, projectName)
	if err != nil {
		return nil, err
	}

	// Create where clause from filter
	var where map[string]string
	if filter != nil {
		where = createWhereMap(filter)
	}

	// Perform search
	results, err := collection.Query(ctx, query, topK, where, nil)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results
	searchResults := make([]SearchResult, len(results))
	for i, result := range results {
		// Convert metadata back to map[string]any
		metadata := make(map[string]any)
		for k, v := range result.Metadata {
			metadata[k] = v
		}

		searchResults[i] = SearchResult{
			ID:         result.ID,
			Content:    result.Content,
			Metadata:   metadata,
			Similarity: result.Similarity,
		}
	}

	return searchResults, nil
}

// DeleteDocument removes a document from a project's collection
func (s *EmbeddingStore) DeleteDocument(ctx context.Context, projectName string, docID string) error {
	collection, err := s.GetOrCreateCollection(ctx, projectName)
	if err != nil {
		return err
	}

	// Delete from collection
	if err := collection.Delete(ctx, nil, nil, docID); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// DeleteCollection removes an entire project's collection
func (s *EmbeddingStore) DeleteCollection(ctx context.Context, projectName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete from chromem
	if err := s.db.DeleteCollection(projectName); err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	// Remove from cache
	delete(s.collections, projectName)
	s.logger.Infof("Deleted collection for project: %s", projectName)

	return nil
}

// GetCollectionStats returns statistics for a project's collection
func (s *EmbeddingStore) GetCollectionStats(ctx context.Context, projectName string) (*CollectionStats, error) {
	collection, err := s.GetOrCreateCollection(ctx, projectName)
	if err != nil {
		return nil, err
	}

	// Get document count (chromem doesn't expose this directly, so we'll track it)
	// For now, return a basic stats object
	stats := &CollectionStats{
		Name:          projectName,
		DocumentCount: collection.Count(),
		LastUpdated:   time.Now().Unix(),
	}

	return stats, nil
}

// Close closes the embedding store
func (s *EmbeddingStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear collections cache
	s.collections = make(map[string]*chromem.Collection)

	// chromem-go doesn't require explicit close for persistent DB
	return nil
}

// Document represents a document to be stored
type Document struct {
	ID       string         `json:"id"`
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata"`
}

// SearchResult represents a search result
type SearchResult struct {
	ID         string         `json:"id"`
	Content    string         `json:"content"`
	Metadata   map[string]any `json:"metadata"`
	Similarity float32        `json:"similarity"`
}

// Filter defines search filters
type Filter interface {
	ToMap() map[string]string
}

// MetadataFilter filters by metadata fields
type MetadataFilter struct {
	Field    string
	Operator string // eq, ne, gt, lt, gte, lte, contains
	Value    any
}

func (f MetadataFilter) ToMap() map[string]string {
	// For chromem-go, we can only filter by exact metadata matches
	// So we'll only support "eq" operator for now
	if f.Operator == "eq" {
		return map[string]string{
			f.Field: fmt.Sprintf("%v", f.Value),
		}
	}
	return nil
}

// CollectionStats holds statistics about a collection
type CollectionStats struct {
	Name          string `json:"name"`
	DocumentCount int    `json:"document_count"`
	LastUpdated   int64  `json:"last_updated"`
}

// Helper function to create where map from filter
func createWhereMap(filter Filter) map[string]string {
	if filter == nil {
		return nil
	}
	return filter.ToMap()
}

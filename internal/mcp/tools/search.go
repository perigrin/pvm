// ABOUTME: Search tool implementation using chromem-go for semantic similarity and pattern matching
// ABOUTME: Provides similarity search, type signature search, and pattern search capabilities

package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/log"
	"tamarou.com/pvm/internal/mcp/embeddings"
)

// SearchResult represents a unified search result
type SearchResult struct {
	ID           string         `json:"id"`
	Content      string         `json:"content"`
	Metadata     map[string]any `json:"metadata"`
	Similarity   float32        `json:"similarity"`
	SearchMethod string         `json:"search_method"`
	FilePath     string         `json:"file_path"`
	Type         string         `json:"type"`
	Score        float32        `json:"score"` // Combined score from multiple methods
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query         string            `json:"query"`
	Method        string            `json:"search_method"`
	ProjectPath   string            `json:"project_path"`
	TopK          int               `json:"top_k"`
	Filters       map[string]string `json:"filters"`
	MinSimilarity float32           `json:"min_similarity"`
}

// SearchResponse represents the search response
type SearchResponse struct {
	Status      string         `json:"status"`
	Results     []SearchResult `json:"results"`
	Timestamp   string         `json:"timestamp"`
	Method      string         `json:"method"`
	TotalHits   int            `json:"total_hits"`
	SearchTime  string         `json:"search_time"`
	ProjectPath string         `json:"project_path"`
}

// CodeSearcher provides code search capabilities using chromem-go
type CodeSearcher struct {
	store         *embeddings.EmbeddingStore
	manager       *embeddings.CollectionManager
	extractor     *embeddings.Extractor
	logger        *log.Logger
	defaultTopK   int
	minSimilarity float32
}

// NewCodeSearcher creates a new code searcher
func NewCodeSearcher(store *embeddings.EmbeddingStore, manager *embeddings.CollectionManager, extractor *embeddings.Extractor, logger *log.Logger) *CodeSearcher {
	return &CodeSearcher{
		store:         store,
		manager:       manager,
		extractor:     extractor,
		logger:        logger,
		defaultTopK:   20,
		minSimilarity: 0.3,
	}
}

// Search performs a search using the specified method
func (cs *CodeSearcher) Search(ctx context.Context, request SearchRequest) (*SearchResponse, error) {
	start := time.Now()

	// Set defaults
	if request.TopK == 0 {
		request.TopK = cs.defaultTopK
	}
	// Only set default MinSimilarity if it's negative (unset), not zero
	if request.MinSimilarity < 0 {
		request.MinSimilarity = cs.minSimilarity
	}

	cs.logger.Infof("Performing %s search for query: %s", request.Method, request.Query)

	var results []SearchResult
	var err error

	switch request.Method {
	case "similarity":
		results, err = cs.similaritySearch(ctx, request)
	case "type_signature":
		results, err = cs.typeSignatureSearch(ctx, request)
	case "pattern":
		results, err = cs.patternSearch(ctx, request)
	default:
		return nil, errors.New("SYS", "System Error", "INVALID_SEARCH_METHOD", fmt.Sprintf("unsupported search method: %s", request.Method), nil)
	}

	if err != nil {
		return nil, errors.Wrap(err, "SYS", "System Error", "SEARCH_FAILED", "search failed")
	}

	// Apply post-processing filters
	results = cs.applyFilters(results, request)

	// Sort by combined score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit results
	if len(results) > request.TopK {
		results = results[:request.TopK]
	}

	searchTime := time.Since(start)

	response := &SearchResponse{
		Status:      "success",
		Results:     results,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Method:      request.Method,
		TotalHits:   len(results),
		SearchTime:  searchTime.String(),
		ProjectPath: request.ProjectPath,
	}

	cs.logger.Infof("Search completed: %d results in %v", len(results), searchTime)

	return response, nil
}

// similaritySearch performs semantic similarity search using embeddings
func (cs *CodeSearcher) similaritySearch(ctx context.Context, request SearchRequest) ([]SearchResult, error) {
	if request.ProjectPath == "" {
		return nil, errors.New("SYS", "System Error", "PROJECT_PATH_REQUIRED", "project path is required for similarity search", nil)
	}

	// Get the collection for the project
	collection, err := cs.manager.GetOrCreateCollection(ctx, request.ProjectPath)
	if err != nil {
		return nil, errors.Wrap(err, "SYS", "System Error", "GET_COLLECTION_FAILED", "failed to get collection for project")
	}

	// Convert filters to metadata filters
	filters := make(map[string]string)
	for k, v := range request.Filters {
		filters[k] = v
	}

	// Check if collection has any documents
	docCount := collection.Count()
	if docCount == 0 {
		// Return empty results for empty collection
		return []SearchResult{}, nil
	}

	// Adjust topK if it exceeds document count
	topK := request.TopK * 2
	if topK > docCount {
		topK = docCount
	}

	// Perform similarity search using chromem-go
	searchResults, err := cs.manager.SearchWithFilters(ctx, collection, request.Query, filters, topK)
	if err != nil {
		return nil, errors.Wrap(err, "SYS", "System Error", "SIMILARITY_SEARCH_FAILED", "failed to perform similarity search")
	}

	// Convert to unified search results
	results := make([]SearchResult, 0, len(searchResults))
	for _, result := range searchResults {
		if result.Similarity < request.MinSimilarity {
			continue
		}

		filePath, _ := result.Metadata["file_path"].(string)
		codeType, _ := result.Metadata["type"].(string)

		unifiedResult := SearchResult{
			ID:           result.ID,
			Content:      result.Content,
			Metadata:     result.Metadata,
			Similarity:   result.Similarity,
			SearchMethod: "similarity",
			FilePath:     filePath,
			Type:         codeType,
			Score:        result.Similarity, // For similarity search, score equals similarity
		}

		results = append(results, unifiedResult)
	}

	return results, nil
}

// typeSignatureSearch searches for code based on type signatures
func (cs *CodeSearcher) typeSignatureSearch(ctx context.Context, request SearchRequest) ([]SearchResult, error) {
	// First, perform similarity search to get candidates
	similarityResults, err := cs.similaritySearch(ctx, request)
	if err != nil {
		return nil, err
	}

	// Extract type signature from query
	queryTypes := cs.extractTypeSignature(request.Query)

	// Score results based on type signature matching
	results := make([]SearchResult, 0)
	for _, result := range similarityResults {
		typeScore := cs.scoreTypeSignature(result, queryTypes)

		// Combine similarity and type scores
		combinedScore := (result.Similarity * 0.6) + (typeScore * 0.4)

		if combinedScore > 0.2 { // Minimum threshold for type signature search
			result.SearchMethod = "type_signature"
			result.Score = combinedScore
			results = append(results, result)
		}
	}

	return results, nil
}

// patternSearch searches for code using regex patterns and AST metadata
func (cs *CodeSearcher) patternSearch(ctx context.Context, request SearchRequest) ([]SearchResult, error) {
	// First, perform similarity search to get candidates
	similarityResults, err := cs.similaritySearch(ctx, request)
	if err != nil {
		return nil, err
	}

	// Compile regex pattern from query
	pattern, err := cs.compileSearchPattern(request.Query)
	if err != nil {
		// If pattern compilation fails, fall back to substring matching
		pattern = nil
		cs.logger.Warningf("Pattern compilation failed, using substring matching: %v", err)
	}

	results := make([]SearchResult, 0)
	for _, result := range similarityResults {
		patternScore := cs.scorePatternMatch(result, request.Query, pattern)

		// Combine similarity and pattern scores
		combinedScore := (result.Similarity * 0.4) + (patternScore * 0.6)

		if combinedScore > 0.2 { // Minimum threshold for pattern search
			result.SearchMethod = "pattern"
			result.Score = combinedScore
			results = append(results, result)
		}
	}

	return results, nil
}

// extractTypeSignature extracts type information from a search query
func (cs *CodeSearcher) extractTypeSignature(query string) []string {
	// Look for common Perl type patterns with non-overlapping matching
	typePatterns := []string{
		`\b(Int|Str|Bool|ArrayRef|HashRef|CodeRef|Object|Any|Undef)\b`, // Built-in types
		`\$[a-zA-Z_][a-zA-Z0-9_]*`,                                     // Variables
	}

	// Use a map to track unique types and avoid duplicates
	typeSet := make(map[string]bool)

	for _, pattern := range typePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(query, -1)
		for _, match := range matches {
			typeSet[match] = true
		}
	}

	// Handle custom types (capitalized words not already matched by built-in types)
	customTypeRegex := regexp.MustCompile(`\b([A-Z][a-zA-Z0-9_]*)\b`)
	customMatches := customTypeRegex.FindAllString(query, -1)
	for _, match := range customMatches {
		// Only add if it's not already a built-in type
		if !typeSet[match] {
			typeSet[match] = true
		}
	}

	// Convert map keys to slice
	var types []string
	for typeStr := range typeSet {
		types = append(types, typeStr)
	}

	return types
}

// scoreTypeSignature scores a result based on type signature matching
func (cs *CodeSearcher) scoreTypeSignature(result SearchResult, queryTypes []string) float32 {
	if len(queryTypes) == 0 {
		return 0.0
	}

	// Check metadata for type information
	typeInfo, hasTypes := result.Metadata["type_info"]
	if !hasTypes {
		return 0.0
	}

	typeInfoStr := fmt.Sprintf("%v", typeInfo)

	matchCount := 0
	for _, queryType := range queryTypes {
		if strings.Contains(typeInfoStr, queryType) {
			matchCount++
		}
	}

	return float32(matchCount) / float32(len(queryTypes))
}

// compileSearchPattern compiles a search pattern with support for Perl-specific syntax
func (cs *CodeSearcher) compileSearchPattern(query string) (*regexp.Regexp, error) {
	// If the query looks like a regex pattern, use it directly
	if strings.HasPrefix(query, "/") && strings.HasSuffix(query, "/") {
		pattern := query[1 : len(query)-1]
		return regexp.Compile(pattern)
	}

	// Otherwise, create a pattern that matches common Perl constructs
	escapedQuery := regexp.QuoteMeta(query)

	// Add some Perl-specific pattern enhancements
	pattern := fmt.Sprintf(`(?i)(%s|sub\s+%s|package\s+%s|use\s+%s)`, escapedQuery, escapedQuery, escapedQuery, escapedQuery)

	return regexp.Compile(pattern)
}

// scorePatternMatch scores a result based on pattern matching
func (cs *CodeSearcher) scorePatternMatch(result SearchResult, query string, pattern *regexp.Regexp) float32 {
	content := result.Content

	if pattern != nil {
		// Use regex pattern matching
		matches := pattern.FindAllString(content, -1)
		if len(matches) == 0 {
			return 0.0
		}

		// Score based on number of matches and their positions
		score := float32(len(matches)) * 0.1
		if score > 1.0 {
			score = 1.0
		}
		return score
	}

	// Fall back to substring matching
	queryLower := strings.ToLower(query)
	contentLower := strings.ToLower(content)

	if strings.Contains(contentLower, queryLower) {
		// Calculate a simple score based on query coverage
		queryLen := len(query)
		contentLen := len(content)

		if contentLen == 0 {
			return 0.0
		}

		// Rough approximation: shorter content with query match scores higher
		score := float32(queryLen) / float32(contentLen)
		if score > 1.0 {
			score = 1.0
		}
		return score
	}

	return 0.0
}

// applyFilters applies additional filters to search results
func (cs *CodeSearcher) applyFilters(results []SearchResult, request SearchRequest) []SearchResult {
	if len(request.Filters) == 0 {
		return results
	}

	filtered := make([]SearchResult, 0, len(results))

	for _, result := range results {
		matches := true

		for key, value := range request.Filters {
			switch key {
			case "file_extension":
				if !strings.HasSuffix(result.FilePath, value) {
					matches = false
				}
			case "type":
				if result.Type != value {
					matches = false
				}
			case "file_pattern":
				matched, _ := filepath.Match(value, result.FilePath)
				if !matched {
					matches = false
				}
			default:
				// Check in metadata
				if metaValue, exists := result.Metadata[key]; !exists || fmt.Sprintf("%v", metaValue) != value {
					matches = false
				}
			}
			if !matches {
				break
			}
		}

		if matches {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// SetDefaultTopK sets the default number of results to return
func (cs *CodeSearcher) SetDefaultTopK(k int) {
	if k > 0 {
		cs.defaultTopK = k
	}
}

// SetMinSimilarity sets the minimum similarity threshold
func (cs *CodeSearcher) SetMinSimilarity(threshold float32) {
	if threshold >= 0.0 && threshold <= 1.0 {
		cs.minSimilarity = threshold
	}
}

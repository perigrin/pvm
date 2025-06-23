// ABOUTME: CPAN search and resolution logic for unknown tools
// ABOUTME: Provides MetaCPAN API integration with caching and fuzzy matching
package tool

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CPANResolver handles CPAN module resolution and search
type CPANResolver struct {
	client    *http.Client
	baseURL   string
	cache     map[string]*ToolResolution
	cacheTime map[string]time.Time
	cacheTTL  time.Duration
}

// MetaCPANResponse represents the response from MetaCPAN API
type MetaCPANResponse struct {
	Hits []MetaCPANHit `json:"hits"`
}

// MetaCPANHit represents a single search result from MetaCPAN
type MetaCPANHit struct {
	Source MetaCPANSource `json:"_source"`
}

// MetaCPANSource represents the source data from MetaCPAN
type MetaCPANSource struct {
	Module       string `json:"module"`
	Distribution string `json:"distribution"`
	Abstract     string `json:"abstract"`
	Version      string `json:"version"`
	Status       string `json:"status"`
}

// NewCPANResolver creates a new CPAN resolver instance
func NewCPANResolver() *CPANResolver {
	return &CPANResolver{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:   "https://fastapi.metacpan.org/v1",
		cache:     make(map[string]*ToolResolution),
		cacheTime: make(map[string]time.Time),
		cacheTTL:  1 * time.Hour, // Cache results for 1 hour
	}
}

// SearchTool searches for a tool on CPAN
func (cr *CPANResolver) SearchTool(toolName string) (*ToolResolution, error) {
	if toolName == "" {
		return nil, NewToolError(ErrInvalidToolName, "tool name cannot be empty")
	}

	// Check cache first
	if resolution := cr.getCached(toolName); resolution != nil {
		return resolution, nil
	}

	// Try different search strategies
	strategies := []func(string) (*ToolResolution, error){
		cr.searchByExecutableName,
		cr.searchByModuleName,
		cr.searchByFuzzyMatch,
	}

	for _, strategy := range strategies {
		if resolution, err := strategy(toolName); err == nil && resolution != nil {
			cr.setCached(toolName, resolution)
			return resolution, nil
		}
	}

	return nil, NewToolError(ErrToolNotFound, fmt.Sprintf("tool '%s' not found in CPAN search", toolName))
}

// searchByExecutableName searches for tools by executable name
func (cr *CPANResolver) searchByExecutableName(toolName string) (*ToolResolution, error) {
	// Search for distributions that might contain the executable
	query := fmt.Sprintf("file.name:%s", toolName)

	results, err := cr.searchMetaCPAN(query)
	if err != nil {
		return nil, err
	}

	// Look for the best match
	for _, hit := range results.Hits {
		if hit.Source.Status == "latest" {
			return &ToolResolution{
				ToolName:    toolName,
				ModuleName:  hit.Source.Module,
				Source:      "cpan-executable",
				Description: hit.Source.Abstract,
				Version:     hit.Source.Version,
			}, nil
		}
	}

	return nil, NewToolError(ErrToolNotFound, "no executable found")
}

// searchByModuleName searches for tools by potential module name
func (cr *CPANResolver) searchByModuleName(toolName string) (*ToolResolution, error) {
	// Try common module name patterns
	patterns := []string{
		fmt.Sprintf("App::%s", strings.Title(toolName)),
		fmt.Sprintf("App::%s", strings.ToUpper(toolName)),
		fmt.Sprintf("%s", strings.Title(toolName)),
		fmt.Sprintf("Perl::%s", strings.Title(toolName)),
	}

	for _, pattern := range patterns {
		query := fmt.Sprintf("module:%s", pattern)
		results, err := cr.searchMetaCPAN(query)
		if err != nil {
			continue
		}

		for _, hit := range results.Hits {
			if hit.Source.Status == "latest" && hit.Source.Module == pattern {
				return &ToolResolution{
					ToolName:    toolName,
					ModuleName:  pattern,
					Source:      "cpan-module",
					Description: hit.Source.Abstract,
					Version:     hit.Source.Version,
				}, nil
			}
		}
	}

	return nil, NewToolError(ErrToolNotFound, "no module found")
}

// searchByFuzzyMatch performs fuzzy matching for similar tool names
func (cr *CPANResolver) searchByFuzzyMatch(toolName string) (*ToolResolution, error) {
	// Search for similar module names
	query := fmt.Sprintf("module:*%s*", toolName)

	results, err := cr.searchMetaCPAN(query)
	if err != nil {
		return nil, err
	}

	// Find the best fuzzy match
	var bestMatch *MetaCPANHit
	bestScore := 0.0

	for _, hit := range results.Hits {
		if hit.Source.Status == "latest" {
			score := cr.calculateSimilarity(toolName, hit.Source.Module)
			if score > bestScore && score > 0.5 { // Minimum similarity threshold
				bestScore = score
				bestMatch = &hit
			}
		}
	}

	if bestMatch != nil {
		return &ToolResolution{
			ToolName:    toolName,
			ModuleName:  bestMatch.Source.Module,
			Source:      "cpan-fuzzy",
			Description: bestMatch.Source.Abstract,
			Version:     bestMatch.Source.Version,
		}, nil
	}

	return nil, NewToolError(ErrToolNotFound, "no fuzzy match found")
}

// searchMetaCPAN performs the actual MetaCPAN API search
func (cr *CPANResolver) searchMetaCPAN(query string) (*MetaCPANResponse, error) {
	// Build the search URL
	params := url.Values{}
	params.Set("q", query)
	params.Set("size", "10")

	searchURL := fmt.Sprintf("%s/search/module?%s", cr.baseURL, params.Encode())

	// Make the HTTP request
	resp, err := cr.client.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("MetaCPAN API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MetaCPAN API returned status %d", resp.StatusCode)
	}

	// Parse the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read MetaCPAN response: %w", err)
	}

	var result MetaCPANResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse MetaCPAN response: %w", err)
	}

	return &result, nil
}

// calculateSimilarity calculates string similarity between tool name and module name
func (cr *CPANResolver) calculateSimilarity(toolName, moduleName string) float64 {
	// Simple similarity calculation based on common substrings
	toolLower := strings.ToLower(toolName)
	moduleLower := strings.ToLower(moduleName)

	// Extract the actual module name from App::Module format
	parts := strings.Split(moduleLower, "::")
	if len(parts) > 1 {
		moduleLower = parts[len(parts)-1]
	}

	// Calculate Levenshtein distance
	distance := levenshteinDistance(toolLower, moduleLower)
	maxLen := max(len(toolLower), len(moduleLower))

	if maxLen == 0 {
		return 0.0
	}

	return 1.0 - float64(distance)/float64(maxLen)
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}

	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1, // deletion
				min(matrix[i][j-1]+1, // insertion
					matrix[i-1][j-1]+cost), // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

// getCached retrieves a cached resolution
func (cr *CPANResolver) getCached(toolName string) *ToolResolution {
	if resolution, exists := cr.cache[toolName]; exists {
		if time.Since(cr.cacheTime[toolName]) < cr.cacheTTL {
			return resolution
		}
		// Cache expired, remove it
		delete(cr.cache, toolName)
		delete(cr.cacheTime, toolName)
	}
	return nil
}

// setCached caches a resolution result
func (cr *CPANResolver) setCached(toolName string, resolution *ToolResolution) {
	cr.cache[toolName] = resolution
	cr.cacheTime[toolName] = time.Now()
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

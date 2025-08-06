// ABOUTME: Adapter for using MetaCPAN provider as a tool resolver
// ABOUTME: Implements Resolver interface using existing CPAN MetaCPAN infrastructure

package tool

import (
	"context"
	"strings"

	"tamarou.com/pvm/internal/cpan"
)

// MetaCPANResolver adapts MetaCPANProvider to implement the Resolver interface
type MetaCPANResolver struct {
	provider *cpan.MetaCPANProvider
}

// NewMetaCPANResolver creates a new MetaCPAN-based tool resolver using existing infrastructure
func NewMetaCPANResolver() (*MetaCPANResolver, error) {
	provider, err := cpan.NewMetaCPANProvider()
	if err != nil {
		return nil, err
	}

	return &MetaCPANResolver{
		provider: provider,
	}, nil
}

// SearchTool searches MetaCPAN for executable files matching the tool name
func (r *MetaCPANResolver) SearchTool(toolName string) (*ToolResolution, error) {
	if toolName == "" {
		return nil, NewToolError(ErrInvalidToolName, "tool name cannot be empty")
	}

	// Search for executable files using the MetaCPAN provider
	ctx := context.Background()
	results, err := r.provider.SearchExecutableFiles(ctx, toolName)
	if err != nil {
		return nil, NewToolError(ErrMappingFailed, "MetaCPAN search failed: "+err.Error())
	}

	if len(results) == 0 {
		return nil, NewToolError(ErrToolNotFound, "tool '"+toolName+"' not found in MetaCPAN")
	}

	// Look for exact matches first
	for _, result := range results {
		if result.IsExactMatch {
			return &ToolResolution{
				ToolName:    toolName,
				ModuleName:  result.Distribution,
				Source:      "metacpan",
				Description: result.Abstract,
				Executable:  result.ToolName,
			}, nil
		}
	}

	// If no exact match, return the first result as best guess
	firstResult := results[0]
	description := firstResult.Abstract
	if description == "" {
		description = "Best match: " + firstResult.FileName + " (from " + firstResult.Distribution + ")"
	}

	return &ToolResolution{
		ToolName:    toolName,
		ModuleName:  firstResult.Distribution,
		Source:      "metacpan",
		Description: description,
		Executable:  strings.TrimSuffix(firstResult.FileName, ".pl"),
	}, nil
}
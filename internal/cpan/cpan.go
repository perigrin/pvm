// ABOUTME: CPAN-specific implementation for CPAN metadata retrieval
// ABOUTME: Implements the Provider interface using the CPAN index directly

package cpan

import (
	"context"
)

// CPANProvider implements the Provider interface using the CPAN index directly
type CPANProvider struct {
	baseURL       string
	userAgent     string
	cacheDir      string
	cacheTTL      int
	mirrors       []string
	currentMirror int
	timeout       int
	_             bool // disableNetwork is deprecated
}

// NewCPANProvider creates a new CPANProvider with the given options
func NewCPANProvider(options ...ProviderOption) (*CPANProvider, error) {
	provider := &CPANProvider{
		baseURL:   "https://www.cpan.org",
		userAgent: defaultUserAgent,
		timeout:   defaultTimeout,
		mirrors:   []string{"https://www.cpan.org"},
	}

	// Apply options
	for _, option := range options {
		if err := option(provider); err != nil {
			return nil, err
		}
	}

	return provider, nil
}

// Name returns the name of the provider
func (p *CPANProvider) Name() string {
	return "cpan"
}

// BaseURL returns the base URL of the provider's API
func (p *CPANProvider) BaseURL() string {
	return p.baseURL
}

// GetModuleInfo retrieves metadata about a specific module from CPAN
func (p *CPANProvider) GetModuleInfo(ctx context.Context, moduleName string) (*ModuleInfo, error) {
	// For now, delegate to MetaCPAN as it has a more complete implementation
	// In a full implementation, this would query the CPAN index directly
	metaProvider, err := NewMetaCPANProvider()
	if err != nil {
		return nil, err
	}
	return metaProvider.GetModuleInfo(ctx, moduleName)
}

// SearchModules searches for modules matching the given query using the CPAN index
func (p *CPANProvider) SearchModules(ctx context.Context, query string, limit int) (*SearchResults, error) {
	// For now, delegate to MetaCPAN as it has a more complete implementation
	// In a full implementation, this would query the CPAN index directly
	metaProvider, err := NewMetaCPANProvider()
	if err != nil {
		return nil, err
	}
	return metaProvider.SearchModules(ctx, query, limit)
}

// GetDependencies retrieves the dependencies for a module from CPAN
func (p *CPANProvider) GetDependencies(ctx context.Context, moduleName string) ([]*Dependency, error) {
	// For now, delegate to MetaCPAN as it has a more complete implementation
	// In a full implementation, this would query the CPAN index directly
	metaProvider, err := NewMetaCPANProvider()
	if err != nil {
		return nil, err
	}
	return metaProvider.GetDependencies(ctx, moduleName)
}

// GetModuleVersions retrieves all available versions of a module from CPAN
func (p *CPANProvider) GetModuleVersions(ctx context.Context, moduleName string) ([]string, error) {
	// For now, delegate to MetaCPAN as it has a more complete implementation
	// In a full implementation, this would query the CPAN index directly
	metaProvider, err := NewMetaCPANProvider()
	if err != nil {
		return nil, err
	}
	return metaProvider.GetModuleVersions(ctx, moduleName)
}

// GetAuthorInfo retrieves information about a CPAN author from CPAN
func (p *CPANProvider) GetAuthorInfo(ctx context.Context, authorID string) (map[string]interface{}, error) {
	// For now, delegate to MetaCPAN as it has a more complete implementation
	// In a full implementation, this would query the CPAN index directly
	metaProvider, err := NewMetaCPANProvider()
	if err != nil {
		return nil, err
	}
	return metaProvider.GetAuthorInfo(ctx, authorID)
}

// IsCoreModule checks if a module is part of the Perl core
func (p *CPANProvider) IsCoreModule(ctx context.Context, moduleName string, perlVersion string) (bool, error) {
	// For now, delegate to MetaCPAN as it has a more complete implementation
	// In a full implementation, this would query the CPAN index directly
	metaProvider, err := NewMetaCPANProvider()
	if err != nil {
		return false, err
	}
	return metaProvider.IsCoreModule(ctx, moduleName, perlVersion)
}

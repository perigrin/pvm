// ABOUTME: CPAN metadata provider interfaces
// ABOUTME: Defines interfaces for retrieving CPAN metadata

package cpan

import (
	"context"
)

// Provider is an interface for retrieving CPAN metadata
type Provider interface {
	// GetModuleInfo retrieves metadata about a specific module
	GetModuleInfo(ctx context.Context, moduleName string) (*ModuleInfo, error)

	// SearchModules searches for modules matching the given query
	SearchModules(ctx context.Context, query string, limit int) (*SearchResults, error)

	// GetDependencies retrieves the dependencies for a module
	GetDependencies(ctx context.Context, moduleName string) ([]*Dependency, error)

	// GetModuleVersions retrieves all available versions of a module
	GetModuleVersions(ctx context.Context, moduleName string) ([]string, error)

	// GetAuthorInfo retrieves information about a CPAN author
	GetAuthorInfo(ctx context.Context, authorID string) (map[string]interface{}, error)

	// IsCoreModule checks if a module is part of Perl core
	IsCoreModule(ctx context.Context, moduleName string, perlVersion string) (bool, error)

	// Name returns the name of the provider
	Name() string

	// BaseURL returns the base URL of the provider's API
	BaseURL() string
}

// ProviderOption is a function that configures a Provider
type ProviderOption func(Provider) error

// NewProvider creates a new Provider with the given options
func NewProvider(source string, options ...ProviderOption) (Provider, error) {
	var provider Provider
	var err error

	switch source {
	case "metacpan":
		provider, err = NewMetaCPANProvider(options...)
	case "cpan":
		provider, err = NewCPANProvider(options...)
	case "custom":
		provider, err = NewCustomProvider(options...)
	default:
		// Default to MetaCPAN if source is not specified or invalid
		provider, err = NewMetaCPANProvider(options...)
	}

	return provider, err
}

// WithBaseURL sets the base URL for the provider's API
func WithBaseURL(url string) ProviderOption {
	return func(p Provider) error {
		// This will be implemented by each provider
		return nil
	}
}

// WithCache enables caching for the provider
func WithCache(cacheDir string, ttl int) ProviderOption {
	return func(p Provider) error {
		// This will be implemented by each provider
		return nil
	}
}

// WithTimeout sets the timeout for API requests
func WithTimeout(timeout int) ProviderOption {
	return func(p Provider) error {
		// This will be implemented by each provider
		return nil
	}
}

// WithUserAgent sets the User-Agent header for API requests
func WithUserAgent(userAgent string) ProviderOption {
	return func(p Provider) error {
		// This will be implemented by each provider
		return nil
	}
}

// WithMirror sets the CPAN mirror to use
func WithMirror(mirror string) ProviderOption {
	return func(p Provider) error {
		// This will be implemented by each provider
		return nil
	}
}

// WithAdditionalMirrors sets additional CPAN mirrors to use as fallbacks
func WithAdditionalMirrors(mirrors []string) ProviderOption {
	return func(p Provider) error {
		// This will be implemented by each provider
		return nil
	}
}

// WithDisableNetwork disables network access (for testing)
func WithDisableNetwork(disable bool) ProviderOption {
	return func(p Provider) error {
		// This will be implemented by each provider
		return nil
	}
}

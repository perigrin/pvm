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
		switch v := p.(type) {
		case *MetaCPANProvider:
			return v.WithBaseURL(url)
		case *CPANProvider:
			v.baseURL = url
			return nil
		case *CustomProvider:
			return v.WithBaseURL(url)
		default:
			return nil
		}
	}
}

// WithCache enables caching for the provider
func WithCache(cacheDir string, ttl int) ProviderOption {
	return func(p Provider) error {
		switch v := p.(type) {
		case *MetaCPANProvider:
			return v.WithCache(cacheDir, ttl)
		case *CPANProvider:
			v.cacheDir = cacheDir
			v.cacheTTL = ttl
			return nil
		case *CustomProvider:
			return v.WithCache(cacheDir, ttl)
		default:
			return nil
		}
	}
}

// WithTimeout sets the timeout for API requests
func WithTimeout(timeout int) ProviderOption {
	return func(p Provider) error {
		switch v := p.(type) {
		case *MetaCPANProvider:
			return v.WithTimeout(timeout)
		case *CPANProvider:
			v.timeout = timeout
			return nil
		case *CustomProvider:
			return v.WithTimeout(timeout)
		default:
			return nil
		}
	}
}

// WithUserAgent sets the User-Agent header for API requests
func WithUserAgent(userAgent string) ProviderOption {
	return func(p Provider) error {
		switch v := p.(type) {
		case *MetaCPANProvider:
			return v.WithUserAgent(userAgent)
		case *CPANProvider:
			v.userAgent = userAgent
			return nil
		case *CustomProvider:
			return v.WithUserAgent(userAgent)
		default:
			return nil
		}
	}
}

// WithMirror sets the CPAN mirror to use
func WithMirror(mirror string) ProviderOption {
	return func(p Provider) error {
		switch v := p.(type) {
		case *MetaCPANProvider:
			return v.WithMirror(mirror)
		case *CPANProvider:
			v.mirrors = []string{mirror}
			v.currentMirror = 0
			return nil
		case *CustomProvider:
			return v.WithMirror(mirror)
		default:
			return nil
		}
	}
}

// WithAdditionalMirrors sets additional CPAN mirrors to use as fallbacks
func WithAdditionalMirrors(mirrors []string) ProviderOption {
	return func(p Provider) error {
		switch v := p.(type) {
		case *MetaCPANProvider:
			return v.WithAdditionalMirrors(mirrors)
		case *CPANProvider:
			v.mirrors = append(v.mirrors, mirrors...)
			return nil
		case *CustomProvider:
			return v.WithAdditionalMirrors(mirrors)
		default:
			return nil
		}
	}
}

// WithDisableNetwork disables network access (for testing)
func WithDisableNetwork(disable bool) ProviderOption {
	return func(p Provider) error {
		switch v := p.(type) {
		case *MetaCPANProvider:
			return v.WithDisableNetwork(disable)
		case *CustomProvider:
			return v.WithDisableNetwork(disable)
		default:
			return nil
		}
	}
}

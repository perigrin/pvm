// ABOUTME: Implementation for custom CPAN metadata provider
// ABOUTME: Allows users to configure their own metadata source

package cpan

import (
	"context"
)

// CustomProvider implements the Provider interface with a custom metadata source
type CustomProvider struct {
	baseURL        string
	userAgent      string
	disableNetwork bool
	cacheDir       string
	cacheTTL       int
	timeout        int
}

// NewCustomProvider creates a new CustomProvider with the given options
func NewCustomProvider(options ...ProviderOption) (*CustomProvider, error) {
	provider := &CustomProvider{
		userAgent: defaultUserAgent,
		timeout:   defaultTimeout,
	}

	// Apply options
	for _, option := range options {
		if err := option(provider); err != nil {
			return nil, err
		}
	}

	// CustomProvider requires a baseURL to be set
	if provider.baseURL == "" {
		return nil, &ProviderError{
			Source:  "custom",
			Code:    "no_base_url",
			Message: "BaseURL must be specified for CustomProvider",
		}
	}

	return provider, nil
}

// Name returns the name of the provider
func (p *CustomProvider) Name() string {
	return "custom"
}

// BaseURL returns the base URL of the provider's API
func (p *CustomProvider) BaseURL() string {
	return p.baseURL
}

// WithBaseURL implements the ProviderOption interface for CustomProvider
func (p *CustomProvider) WithBaseURL(url string) error {
	p.baseURL = url
	return nil
}

// WithUserAgent implements the ProviderOption interface for CustomProvider
func (p *CustomProvider) WithUserAgent(userAgent string) error {
	p.userAgent = userAgent
	return nil
}

// WithTimeout implements the ProviderOption interface for CustomProvider
func (p *CustomProvider) WithTimeout(timeout int) error {
	p.timeout = timeout
	return nil
}

// WithCache implements the ProviderOption interface for CustomProvider
func (p *CustomProvider) WithCache(cacheDir string, ttl int) error {
	p.cacheDir = cacheDir
	p.cacheTTL = ttl
	return nil
}

// WithDisableNetwork implements the ProviderOption interface for CustomProvider
func (p *CustomProvider) WithDisableNetwork(disable bool) error {
	p.disableNetwork = disable
	return nil
}

// GetModuleInfo retrieves metadata about a specific module from the custom source
func (p *CustomProvider) GetModuleInfo(ctx context.Context, moduleName string) (*ModuleInfo, error) {
	if p.disableNetwork {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "network_disabled",
			Message: "Network access is disabled",
		}
	}

	// For custom providers, we can't provide a generic implementation
	// Users would need to implement their own based on their API
	return nil, &ProviderError{
		Source:  p.Name(),
		Code:    "not_implemented",
		Message: "Custom provider must implement GetModuleInfo for their specific API",
	}
}

// SearchModules searches for modules matching the given query using the custom source
func (p *CustomProvider) SearchModules(ctx context.Context, query string, limit int) (*SearchResults, error) {
	if p.disableNetwork {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "network_disabled",
			Message: "Network access is disabled",
		}
	}

	// For custom providers, we can't provide a generic implementation
	// Users would need to implement their own based on their API
	return nil, &ProviderError{
		Source:  p.Name(),
		Code:    "not_implemented",
		Message: "Custom provider must implement SearchModules for their specific API",
	}
}

// GetDependencies retrieves the dependencies for a module from the custom source
func (p *CustomProvider) GetDependencies(ctx context.Context, moduleName string) ([]*Dependency, error) {
	if p.disableNetwork {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "network_disabled",
			Message: "Network access is disabled",
		}
	}

	// For custom providers, we can't provide a generic implementation
	// Users would need to implement their own based on their API
	return nil, &ProviderError{
		Source:  p.Name(),
		Code:    "not_implemented",
		Message: "Custom provider must implement GetDependencies for their specific API",
	}
}

// GetModuleVersions retrieves all available versions of a module from the custom source
func (p *CustomProvider) GetModuleVersions(ctx context.Context, moduleName string) ([]string, error) {
	if p.disableNetwork {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "network_disabled",
			Message: "Network access is disabled",
		}
	}

	// For custom providers, we can't provide a generic implementation
	// Users would need to implement their own based on their API
	return nil, &ProviderError{
		Source:  p.Name(),
		Code:    "not_implemented",
		Message: "Custom provider must implement GetModuleVersions for their specific API",
	}
}

// GetAuthorInfo retrieves information about a CPAN author from the custom source
func (p *CustomProvider) GetAuthorInfo(ctx context.Context, authorID string) (map[string]interface{}, error) {
	if p.disableNetwork {
		return nil, &ProviderError{
			Source:  p.Name(),
			Code:    "network_disabled",
			Message: "Network access is disabled",
		}
	}

	// For custom providers, we can't provide a generic implementation
	// Users would need to implement their own based on their API
	return nil, &ProviderError{
		Source:  p.Name(),
		Code:    "not_implemented",
		Message: "Custom provider must implement GetAuthorInfo for their specific API",
	}
}

// IsCoreModule checks if a module is part of the Perl core
func (p *CustomProvider) IsCoreModule(ctx context.Context, moduleName string, perlVersion string) (bool, error) {
	if p.disableNetwork {
		return false, &ProviderError{
			Source:  p.Name(),
			Code:    "network_disabled",
			Message: "Network access is disabled",
		}
	}

	// For custom providers, we can't provide a generic implementation
	// Users would need to implement their own based on their API
	return false, &ProviderError{
		Source:  p.Name(),
		Code:    "not_implemented",
		Message: "Custom provider must implement IsCoreModule for their specific API",
	}
}

// WithMirror is not applicable for CustomProvider
func (p *CustomProvider) WithMirror(mirror string) error {
	// Not applicable for CustomProvider, as it uses a single custom URL
	return nil
}

// WithAdditionalMirrors is not applicable for CustomProvider
func (p *CustomProvider) WithAdditionalMirrors(mirrors []string) error {
	// Not applicable for CustomProvider, as it uses a single custom URL
	return nil
}

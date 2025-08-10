// ABOUTME: CPAN provider builder pattern for PVI commands
// ABOUTME: Simplifies repetitive provider and resolver setup across commands

package pm

import (
	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/cpan"
	"tamarou.com/pvm/internal/pm/deps"
)

// ProviderBuilder provides a fluent interface for creating CPAN providers
type ProviderBuilder struct {
	source         string
	sourceExplicit bool // track if source was explicitly set
	cfg            *config.Config
	noCache        bool
	options        []cpan.ProviderOption
	withResolver   bool
}

// ProviderBuilderResult contains both provider and resolver
type ProviderBuilderResult struct {
	Provider cpan.Provider
	Resolver deps.DependencyResolver
}

// NewProviderBuilder creates a new provider builder
func NewProviderBuilder() *ProviderBuilder {
	return &ProviderBuilder{
		source:         "metacpan", // default source
		sourceExplicit: false,      // not explicitly set
		options:        make([]cpan.ProviderOption, 0),
		withResolver:   false,
	}
}

// WithConfig sets the configuration to use for provider setup
func (pb *ProviderBuilder) WithConfig(cfg *config.Config) *ProviderBuilder {
	pb.cfg = cfg
	return pb
}

// WithSource sets the metadata source (metacpan, cpan, custom)
func (pb *ProviderBuilder) WithSource(source string) *ProviderBuilder {
	if source != "" {
		pb.source = source
		pb.sourceExplicit = true
	}
	return pb
}

// WithNoCache disables caching for the provider
func (pb *ProviderBuilder) WithNoCache(noCache bool) *ProviderBuilder {
	pb.noCache = noCache
	return pb
}

// WithResolver includes dependency resolver creation
func (pb *ProviderBuilder) WithResolver() *ProviderBuilder {
	pb.withResolver = true
	return pb
}

// AddOption adds a custom provider option
func (pb *ProviderBuilder) AddOption(option cpan.ProviderOption) *ProviderBuilder {
	pb.options = append(pb.options, option)
	return pb
}

// Build creates the configured provider (and optionally resolver)
func (pb *ProviderBuilder) Build() (*ProviderBuilderResult, error) {
	// Resolve source from config if not explicitly set
	effectiveSource := pb.resolveSource()

	// Build provider options from configuration
	options, err := pb.buildProviderOptions()
	if err != nil {
		return nil, err
	}

	// Create the provider
	provider, err := cpan.NewProvider(effectiveSource, options...)
	if err != nil {
		return nil, err
	}

	result := &ProviderBuilderResult{
		Provider: provider,
	}

	// Create resolver if requested
	if pb.withResolver {
		resolver, err := pb.buildResolver()
		if err != nil {
			return nil, err
		}
		result.Resolver = resolver
	}

	return result, nil
}

// BuildProvider creates just the provider (convenience method)
func (pb *ProviderBuilder) BuildProvider() (cpan.Provider, error) {
	result, err := pb.Build()
	if err != nil {
		return nil, err
	}
	return result.Provider, nil
}

// resolveSource determines the effective source from flags and config
func (pb *ProviderBuilder) resolveSource() string {
	// If source explicitly set via WithSource, use it (highest priority)
	if pb.sourceExplicit {
		return pb.source
	}

	// Check config for source override (medium priority)
	if pb.cfg != nil && pb.cfg.PM != nil && pb.cfg.PM.MetadataSource != "" {
		return pb.cfg.PM.MetadataSource
	}

	// Default to metacpan (lowest priority)
	return "metacpan"
}

// buildProviderOptions creates provider options from configuration
func (pb *ProviderBuilder) buildProviderOptions() ([]cpan.ProviderOption, error) {
	options := make([]cpan.ProviderOption, 0)

	// Add custom options first
	options = append(options, pb.options...)

	// Skip config-based options if no config provided
	if pb.cfg == nil || pb.cfg.PM == nil {
		return options, nil
	}

	pviCfg := pb.cfg.PM
	effectiveSource := pb.resolveSource()

	// Set base URL if custom source and URL provided
	if effectiveSource == "custom" && pviCfg.MetadataURL != "" {
		options = append(options, cpan.WithBaseURL(pviCfg.MetadataURL))
	}

	// Set cache settings if caching is enabled and not disabled by flag
	if pviCfg.CacheModules && !pb.noCache && pviCfg.CacheDir != "" {
		options = append(options, cpan.WithCache(pviCfg.CacheDir, pviCfg.CacheTTL))
	}

	// Set mirrors
	if pviCfg.DefaultMirror != "" {
		options = append(options, cpan.WithMirror(pviCfg.DefaultMirror))
	}
	if len(pviCfg.AdditionalMirrors) > 0 {
		options = append(options, cpan.WithAdditionalMirrors(pviCfg.AdditionalMirrors))
	}

	// Set network access
	options = append(options, cpan.WithDisableNetwork(pviCfg.DisableNetwork))

	return options, nil
}

// buildResolver creates a dependency resolver with cache configuration
func (pb *ProviderBuilder) buildResolver() (deps.DependencyResolver, error) {
	var cacheDir string
	var cacheTTL int

	// Configure resolver cache from config if available and caching not disabled
	if pb.cfg != nil && pb.cfg.PM != nil && !pb.noCache {
		cacheDir = pb.cfg.PM.CacheDir
		cacheTTL = pb.cfg.PM.CacheTTL
	}

	return deps.NewDefaultResolver(cacheDir, cacheTTL)
}

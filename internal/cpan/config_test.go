// ABOUTME: Tests for CPAN configuration and mirror management
// ABOUTME: Verifies the configuration options for CPAN metadata providers

package cpan

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"tamarou.com/pvm/internal/config"
)

// TestNewProviderWithDefaults tests creating a provider with default settings
func TestNewProviderWithDefaults(t *testing.T) {
	provider, err := NewProvider("metacpan")
	require.NoError(t, err, "NewProvider should not return an error with default settings")
	require.NotNil(t, provider, "Provider should not be nil")

	// Check that the provider is a MetaCPANProvider
	metacpanProvider, ok := provider.(*MetaCPANProvider)
	require.True(t, ok, "Provider should be a MetaCPANProvider")

	// Check default settings
	assert.Equal(t, "metacpan", metacpanProvider.Name(), "Provider name should be 'metacpan'")
	assert.Equal(t, metaCPANAPIBaseURL, metacpanProvider.BaseURL(), "Provider base URL should be the default MetaCPAN API URL")
	assert.False(t, metacpanProvider.disableNetwork, "Network access should be enabled by default")
	assert.Equal(t, "", metacpanProvider.cacheDir, "Cache directory should be empty by default")
	assert.Equal(t, 0, metacpanProvider.cacheTTL, "Cache TTL should be 0 by default")
	assert.Equal(t, defaultUserAgent, metacpanProvider.userAgent, "User agent should be the default")
	assert.Equal(t, []string{metaCPANAPIBaseURL}, metacpanProvider.mirrors, "Mirrors should contain only the default URL")
}

// TestNewProviderWithOptions tests creating a provider with custom options
func TestNewProviderWithOptions(t *testing.T) {
	// Create a provider with custom options
	provider, err := NewProvider("metacpan",
		WithBaseURL("https://custom.metacpan.org"),
		WithUserAgent("CustomAgent/1.0"),
		WithCache("/tmp/cpan-cache", 48),
		WithDisableNetwork(true),
		WithMirror("https://custom.mirror.org"),
		WithAdditionalMirrors([]string{"https://mirror1.org", "https://mirror2.org"}),
	)
	require.NoError(t, err, "NewProvider should not return an error with custom options")
	require.NotNil(t, provider, "Provider should not be nil")

	// Check that the provider is a MetaCPANProvider
	metacpanProvider, ok := provider.(*MetaCPANProvider)
	require.True(t, ok, "Provider should be a MetaCPANProvider")

	// Check custom settings
	assert.Equal(t, "metacpan", metacpanProvider.Name(), "Provider name should be 'metacpan'")
	assert.Equal(t, "https://custom.metacpan.org", metacpanProvider.BaseURL(), "Provider base URL should be the custom URL")
	assert.True(t, metacpanProvider.disableNetwork, "Network access should be disabled")
	assert.Equal(t, "/tmp/cpan-cache", metacpanProvider.cacheDir, "Cache directory should be the custom directory")
	assert.Equal(t, 48, metacpanProvider.cacheTTL, "Cache TTL should be the custom value")
	assert.Equal(t, "CustomAgent/1.0", metacpanProvider.userAgent, "User agent should be the custom value")
	assert.Equal(t, []string{"https://custom.mirror.org", "https://mirror1.org", "https://mirror2.org"},
		metacpanProvider.mirrors, "Mirrors should contain all specified URLs")
}

// TestNewCustomProvider tests creating a custom provider
func TestNewCustomProvider(t *testing.T) {
	// Create a custom provider
	provider, err := NewProvider("custom", WithBaseURL("https://custom.api.org"))
	require.NoError(t, err, "NewProvider should not return an error for a custom provider with a base URL")
	require.NotNil(t, provider, "Provider should not be nil")

	// Check that the provider is a CustomProvider
	customProvider, ok := provider.(*CustomProvider)
	require.True(t, ok, "Provider should be a CustomProvider")

	// Check settings
	assert.Equal(t, "custom", customProvider.Name(), "Provider name should be 'custom'")
	assert.Equal(t, "https://custom.api.org", customProvider.BaseURL(), "Provider base URL should be the custom URL")
}

// TestNewCustomProviderNoBaseURL tests that creating a custom provider without a base URL fails
func TestNewCustomProviderNoBaseURL(t *testing.T) {
	// Create a custom provider without a base URL
	provider, err := NewProvider("custom")
	assert.Error(t, err, "NewProvider should return an error for a custom provider without a base URL")
	assert.Nil(t, provider, "Provider should be nil")
}

// TestNewProviderWithUnknownSource tests creating a provider with an unknown source
func TestNewProviderWithUnknownSource(t *testing.T) {
	// Create a provider with an unknown source
	provider, err := NewProvider("unknown")
	require.NoError(t, err, "NewProvider should not return an error for an unknown source (defaults to MetaCPAN)")
	require.NotNil(t, provider, "Provider should not be nil")

	// Check that the provider is a MetaCPANProvider
	_, ok := provider.(*MetaCPANProvider)
	assert.True(t, ok, "Provider should be a MetaCPANProvider (default)")
}

// TestWithCPANConfig tests creating a provider with a CPAN configuration
func TestWithCPANConfig(t *testing.T) {
	// Create a config with CPAN settings
	cfg := &config.Config{
		PVI: &config.PVIConfig{
			MetadataSource:    "metacpan",
			MetadataURL:       "https://config.metacpan.org",
			DefaultMirror:     "https://config.mirror.org",
			AdditionalMirrors: []string{"https://config.mirror1.org", "https://config.mirror2.org"},
			CacheDir:          "/config/cache",
			CacheTTL:          72,
			DisableNetwork:    true,
		},
	}

	// Create a provider based on the configuration
	var providerOpts []ProviderOption
	source := cfg.PVI.MetadataSource

	// Set base URL if custom source and URL provided
	if source == "custom" && cfg.PVI.MetadataURL != "" {
		providerOpts = append(providerOpts, WithBaseURL(cfg.PVI.MetadataURL))
	}

	// Set cache settings
	if cfg.PVI.CacheDir != "" {
		providerOpts = append(providerOpts, WithCache(cfg.PVI.CacheDir, cfg.PVI.CacheTTL))
	}

	// Set mirrors
	if cfg.PVI.DefaultMirror != "" {
		providerOpts = append(providerOpts, WithMirror(cfg.PVI.DefaultMirror))
	}
	if len(cfg.PVI.AdditionalMirrors) > 0 {
		providerOpts = append(providerOpts, WithAdditionalMirrors(cfg.PVI.AdditionalMirrors))
	}

	// Set network access
	providerOpts = append(providerOpts, WithDisableNetwork(cfg.PVI.DisableNetwork))

	provider, err := NewProvider(source, providerOpts...)
	require.NoError(t, err, "NewProvider should not return an error with config options")
	require.NotNil(t, provider, "Provider should not be nil")

	// Check that the provider is a MetaCPANProvider
	metacpanProvider, ok := provider.(*MetaCPANProvider)
	require.True(t, ok, "Provider should be a MetaCPANProvider")

	// Check that settings from config were applied
	// In this case, since we're using MetaCPAN, the base URL from config is not applied
	assert.Equal(t, "metacpan", metacpanProvider.Name(), "Provider name should be 'metacpan'")
	assert.Equal(t, "/config/cache", metacpanProvider.cacheDir, "Cache directory should be from config")
	assert.Equal(t, 72, metacpanProvider.cacheTTL, "Cache TTL should be from config")
	assert.True(t, metacpanProvider.disableNetwork, "Network access should be from config")
	assert.Equal(t, []string{"https://config.mirror.org", "https://config.mirror1.org", "https://config.mirror2.org"},
		metacpanProvider.mirrors, "Mirrors should be from config")
}

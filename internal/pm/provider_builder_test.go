package pm

import (
	"testing"

	"tamarou.com/pvm/internal/config"
	"tamarou.com/pvm/internal/cpan"
)

func TestProviderBuilder_DefaultBehavior(t *testing.T) {
	builder := NewProviderBuilder()

	provider, err := builder.BuildProvider()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if provider == nil {
		t.Fatal("Expected provider, got nil")
	}

	// Should default to metacpan
	if provider.Name() != "metacpan" {
		t.Errorf("Expected metacpan provider, got: %s", provider.Name())
	}
}

func TestProviderBuilder_WithSource(t *testing.T) {
	builder := NewProviderBuilder()

	provider, err := builder.WithSource("cpan").BuildProvider()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if provider.Name() != "cpan" {
		t.Errorf("Expected cpan provider, got: %s", provider.Name())
	}
}

func TestProviderBuilder_WithConfig(t *testing.T) {
	cfg := &config.Config{
		PM: &config.PMConfig{
			MetadataSource: "cpan",
			CacheModules:   true,
			CacheDir:       "/tmp/test-cache",
			CacheTTL:       3600,
			DefaultMirror:  "https://cpan.test.org",
		},
	}

	builder := NewProviderBuilder()
	provider, err := builder.WithConfig(cfg).BuildProvider()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should use config source
	if provider.Name() != "cpan" {
		t.Errorf("Expected cpan provider from config, got: %s", provider.Name())
	}
}

func TestProviderBuilder_SourcePriority(t *testing.T) {
	// Config specifies cpan, but explicit source should override
	cfg := &config.Config{
		PM: &config.PMConfig{
			MetadataSource: "cpan",
		},
	}

	builder := NewProviderBuilder()
	provider, err := builder.WithConfig(cfg).WithSource("metacpan").BuildProvider()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Explicit source should override config
	if provider.Name() != "metacpan" {
		t.Errorf("Expected explicit source to override config, got: %s", provider.Name())
	}
}

func TestProviderBuilder_WithResolver(t *testing.T) {
	cfg := &config.Config{
		PM: &config.PMConfig{
			CacheDir: "/tmp/test-cache",
			CacheTTL: 3600,
		},
	}

	builder := NewProviderBuilder()
	result, err := builder.WithConfig(cfg).WithResolver().Build()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Provider == nil {
		t.Error("Expected provider in result")
	}

	if result.Resolver == nil {
		t.Error("Expected resolver in result")
	}
}

func TestProviderBuilder_WithNoCache(t *testing.T) {
	cfg := &config.Config{
		PM: &config.PMConfig{
			CacheModules: true,
			CacheDir:     "/tmp/test-cache",
			CacheTTL:     3600,
		},
	}

	builder := NewProviderBuilder()
	result, err := builder.WithConfig(cfg).WithNoCache(true).WithResolver().Build()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// This test verifies the builder handles noCache correctly
	// The actual cache disabling is handled by the provider options
	if result.Provider == nil {
		t.Error("Expected provider in result")
	}

	if result.Resolver == nil {
		t.Error("Expected resolver in result")
	}
}

func TestProviderBuilder_CustomOptions(t *testing.T) {
	customOption := cpan.WithTimeout(30)

	builder := NewProviderBuilder()
	provider, err := builder.AddOption(customOption).BuildProvider()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if provider == nil {
		t.Error("Expected provider with custom option")
	}
}

func TestProviderBuilder_NilConfig(t *testing.T) {
	builder := NewProviderBuilder()
	provider, err := builder.WithConfig(nil).BuildProvider()
	if err != nil {
		t.Fatalf("Expected no error with nil config, got: %v", err)
	}

	if provider == nil {
		t.Error("Expected provider even with nil config")
	}

	// Should default to metacpan
	if provider.Name() != "metacpan" {
		t.Errorf("Expected metacpan provider with nil config, got: %s", provider.Name())
	}
}
